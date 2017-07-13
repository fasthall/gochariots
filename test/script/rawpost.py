import socket
import time
import sys
import uuid
import atexit

batcher = ('localhost', 9000)

def build_payload(key, value, seed):
    return '{"Host":0,"Tags":{"' + key + '":"' + value + '"},"Seed":' + str(seed) + '}'
bs = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
bs.connect(batcher)

def onexit():
    bs.close()

atexit.register(onexit)

payload = build_payload(sys.argv[1], sys.argv[2], 0)
l = len(payload) + 1
header = l.to_bytes(4, byteorder='big')
header += b'r'
bs.send(header + payload.encode())
