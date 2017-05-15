import sys
import time
import socket

indexer = ('localhost', 9040)

def build(tag, value):
    tmp = '{"'+tag+'":"'+value+'"}'
    n = len(tmp) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'g'
    return header + tmp.encode()

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect(indexer)
s.send(build(sys.argv[1], sys.argv[2]))
buf = s.recv(1024)
print(buf)
s.close()
