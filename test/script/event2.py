import socket
import time
import sys
import struct
import atexit
import fnv

host = ('localhost', int(sys.argv[1]))
batcher = ('localhost', 9100)

def build_payload(suuid):
    return '{"Host":1,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"2"},"Pre":{"Host":0,"TOId":0,"Tags":{"' + suuid + '":"1"}}}'

def build_payload2(suuid):
    return '{"Host":1,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"2"},"Pre":{"Host":0,"TOId":' + str(int(suuid)) + '}}'

def build_payload_hash(suuid):
    hash = fnv.hash((suuid + ':1').encode(), bits=64)
    return '{"Host":1,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"2"},"Pre":{"Host":0,"TOId":0,"Hash":' + str(hash) + '}}'

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
    suuids = buf[:-1].split(b'\n')
    for suuid in suuids:
        # end = str(time.time())
        # print('Append event 2 at:', end, '@B', suuid.decode())

        # send to batcher
        payload = build_payload2(suuid.decode())
        n = len(payload) + 1
        header = n.to_bytes(4, byteorder='big')
        header += b'r'
        bs.send(header + payload.encode())
