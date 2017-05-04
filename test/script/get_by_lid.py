import sys
import time
import socket

maintainer = ('localhost', 9030)

def build(lid):
    lid = lid.to_bytes(4, byteorder='big')
    n = len(lid) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'l'
    return header + lid

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect(maintainer)
s.send(build(int(sys.argv[1])))
buf = s.recv(1024)
print(buf)
s.close()
