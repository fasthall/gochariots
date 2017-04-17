import sys
import time
import socket

batchers = [('localhost', 9000)]
payload = '{"Host":0,"TOId":0,"LId":0,"Tags":{"key":"value"},"Pre":{"Host":0,"TOId":0}}'

def post(s):
    s.send(header)

duration = int(sys.argv[1])
rate = int(sys.argv[2])
total = duration * rate
precision = 1000
interval = 1.0 / rate
pool = 10

s = [0] * pool
for i in range(pool):
    s[i] = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s[i].connect(batchers[0])

n = len(payload) + 1
header = n.to_bytes(4, byteorder='big')
header += b'r'
header += payload.encode()

start = time.time()
last = time.time()

cnt = 0
diff = 0
while cnt < total / pool:
    if time.time() - last > interval * pool * precision - diff:
        a = time.time()
        for i in range(precision):
            cnt += 1
            for j in range(pool):
                post(s[j])
        diff = time.time() - a
        last = time.time()
print(time.time() - start)

for j in range(pool):
    s[j].close()