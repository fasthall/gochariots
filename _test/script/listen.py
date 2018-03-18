import sys
import fnv
import time
import json
import socket

maintainer = ('localhost', 9030)
host = 'localhost'
port = 9090

def build():
    n = 1
    header = n.to_bytes(4, byteorder='big')
    header += b's'
    return header

data = b'key:value'
hash = fnv.hash(data, algorithm=fnv.fnv_1a, bits=64)
print(hash)

client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
client.connect(maintainer)
client.send(build())
while client:
    buf = client.recv(1024)
    while buf[-1] != 10:
        buf += client.recv(1024)
    jsons = buf[:-1].split(b'\n')
    for j in jsons:
        d = json.loads(j.decode())
        print(d)
client.close()