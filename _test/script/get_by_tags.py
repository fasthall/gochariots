import sys
import time
import socket

indexer = [('localhost', 9040), ('localhost', 9140)]

def build(tag, value):
    tmp = '{"'+tag+'":"'+value+'"}'
    n = len(tmp) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'g'
    return header + tmp.encode()

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect(indexer[int(sys.argv[1])])
s.send(build(sys.argv[2], sys.argv[3]))
buf = s.recv(1024)
print(buf)
s.close()
