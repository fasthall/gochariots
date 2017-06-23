import socket
import time
import sys
import struct
import atexit
import fnv

host = ('localhost', int(sys.argv[1]))
batcher = ('localhost', 9100)

def build_payload_toid(suuid, toid):
    return '{"Host":1,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"2"},"Pre":{"Host":0,"TOId":' + toid + '}}'

bs = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
bs.connect(batcher)
ss = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
ss.bind(host)
ss.listen(0)

def onexit():
    bs.close()
    ss.close()
    print('exit')

atexit.register(onexit)

c1socket, addr = ss.accept()

while True:
    buf = c1socket.recv(1024)
    if len(buf) == 0:
        continue
    while buf[-1] != 10:
        buf += c1socket.recv(1024)
    idPairs = buf[:-1].split(b'\n')
    for idPair in idPairs:
        # end = str(time.time())
        # print('Append event 2 at:', end, '@B', suuid.decode())

        # send to batcher
        idPair = idPair.decode().split(':')
        suuid = idPair[0]
        toid = idPair[1]
        payload = build_payload_toid(suuid, toid)
        n = len(payload) + 1
        header = n.to_bytes(4, byteorder='big')
        header += b'r'
        bs.send(header + payload.encode())
