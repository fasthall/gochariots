import sys
import fnv
import time
import json
import socket

indexer = ('localhost', 9040)

def build():
    n = 1
    header = n.to_bytes(4, byteorder='big')
    header += b's'
    return header

client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
client.connect(indexer)
client.send(build())
while client:
    buf = client.recv(4)
    print('TOId', int.from_bytes(buf, byteorder='big'))
    buf = client.recv(8)
    print('Hash', int.from_bytes(buf, byteorder='big'))

client.close()
