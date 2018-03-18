import sys
import time
import socket

maintainer = [('localhost', 9030), ('localhost', 9130)]

def build(lid):
    lid = lid.to_bytes(4, byteorder='big')
    n = len(lid) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'l'
    return header + lid

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect(maintainer[int(sys.argv[1])])
s.send(build(int(sys.argv[2])))
buf = s.recv(1024)
print(buf)
s.close()
