import socket
import time
import sys
import uuid
import atexit

n = int(sys.argv[1])
batcher = ('169.231.235.49', 9000)
client2 = ('169.231.235.59', 9999)

def build_payload(suuid):
    return '{"Host":0,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"1"},"Pre":{"Host":0,"TOId":0}}'

c2socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
c2socket.connect(client2)
bs = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
bs.connect(batcher)

def onexit():
    c2socket.close()
    bs.close()

atexit.register(onexit)

for i in range(n):
    # suuid = str(uuid.uuid4())
    suuid = str(i)

    payload = build_payload(suuid)
    n = len(payload) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'r'

    if i % 10000 == 0:
        start = str(time.time())
        print('Append event 1 at:', start, '@A', suuid)

    # send to batcher
    c2socket.send(suuid.encode() + b'\n')
    bs.send(header + payload.encode())

