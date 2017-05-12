import socket
import time
import sys
import struct
import atexit

host = (socket.gethostname(), 9999)
batcher = ('169.231.235.71', 9100)

def build_payload(suuid):
    return '{"Host":1,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"2"},"Pre":{"Host":0,"TOId":0}}'

ss = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
ss.bind(host)
bs = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
bs.connect(batcher)

ss.listen(0)
c1socket, addr = ss.accept()

def onexit():
    ss.close()
    c1socket.close()
    bs.close()

atexit.register(onexit)

while True:
    buf = c1socket.recv(1024)
    if len(buf) == 0:
        continue
    while buf[-1] != 10:
        buf += c1socket.recv(1024)
    suuids = buf[:-1].split(b'\n')
    for suuid in suuids:
        end = str(time.time())
        print('Append event 2 at:', end, '@B', suuid.decode())

        # send to batcher
        payload = build_payload(suuid.decode())
        n = len(payload) + 1
        header = n.to_bytes(4, byteorder='big')
        header += b'r'
        bs.send(header + payload.encode())