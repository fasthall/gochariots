import socket
import time
import sys
import uuid
import atexit

rate = int(sys.argv[2])
duration = int(sys.argv[3])
precision = 0.1 # 100ms
batcher = ('localhost', 9000)
client2 = ('localhost', int(sys.argv[1]))

def build_payload(suuid):
    return '{"Host":0,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"1"},"Seed":'+suuid+'}'

c2socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
c2socket.connect(client2)
bs = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
bs.connect(batcher)

def onexit():
    # c2socket.close()
    bs.close()

atexit.register(onexit)

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
            c2socket.send(suuid.encode() + b'\n')
            bs.send(header + payload.encode())
        eid += int(rate * precision)
        ttl -= int(rate * precision)
        print(time.time(), '\tAppended', eid, 'events,', ttl, 'remaining.')
