import socket
import time
import sys
import uuid
import atexit
import fnv
import threading

rate = int(sys.argv[2])
duration = int(sys.argv[3])
precision = 0.1 # 100ms
batcher = ('localhost', 9000)
indexer = ('localhost', 9040)
client2 = ('localhost', int(sys.argv[1]))
hmap = {}

def subcription():
    n = 1
    header = n.to_bytes(4, byteorder='big')
    header += b's'
    return header

def build_payload(suuid):
    tags = suuid + ':1'
    hmap[fnv.hash(tags.encode(), bits=64)] = suuid 
    return '{"Host":0,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"1"},"Pre":{"Host":0,"TOId":0}}'

indexerSocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
indexerSocket.connect(indexer)
client2Socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
client2Socket.connect(client2)
batcherSocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
batcherSocket.connect(batcher)

def onexit():
    indexerSocket.close()
    client2Socket.close()
    batcherSocket.close()

atexit.register(onexit)

indexerSocket.send(subcription())
class listenThread(threading.Thread):
    def run(self):
        while indexerSocket:
            buf = indexerSocket.recv(4)
            toid = int.from_bytes(buf, byteorder='big')
            buf = indexerSocket.recv(8)
            hash = int.from_bytes(buf, byteorder='big')
            # print('TOId', toid, 'Hash', hash)
            if hash == b'':
                break
            if hash in hmap:
                suuid = hmap[hash]
                client2Socket.send((str(suuid) + ':' + str(toid)).encode() + b'\n')

thread = listenThread()
thread.start()

ttl = rate * duration
last = time.time()
eid = 0
while ttl > 0:
    if time.time() - last > precision:
        last = time.time()
        for i in range(int(rate * precision)):
            suuid = str(eid + i)

            payload = build_payload(suuid)
            l = len(payload) + 1
            header = l.to_bytes(4, byteorder='big')
            header += b'r'

            # send to batcher
            batcherSocket.send(header + payload.encode())
        eid += int(rate * precision)
        ttl -= int(rate * precision)
        print(time.time(), '\tAppended', eid, 'events,', ttl, 'remaining.')

thread.join()
