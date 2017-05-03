import socket
import time
import sys
import uuid

host = (sys.argv[1], int(sys.argv[2]))
delay = float(sys.argv[3])
suuid = str(uuid.uuid4())
payload = '{"Host":0,"TOId":0,"LId":0,"Tags":{"UUID":"' + suuid + '", "event":"1"},"Pre":{"Host":0,"TOId":0}}'
batchers = [('localhost', 9000)]

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect(host)

start = str(time.time())
print('Append event 1 at:', start, '@A')
s.send(suuid.encode())
tm = s.recv(1024)
s.close()

time.sleep(delay / 1000)
# send to batcher
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect(batchers[0])
n = len(payload) + 1
header = n.to_bytes(4, byteorder='big')
header += b'r'
s.send(header + payload.encode())
s.close()