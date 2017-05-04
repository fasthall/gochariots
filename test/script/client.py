import sys
import time
import socket

batchers = [('localhost', 9000)]
payload = '{"Host":0,"TOId":0,"LId":0,"Tags":{"from":"id"},"Pre":{"Host":0,"TOId":0,"Tags":{}}}'

def build(conn, id):
    tmp = payload.replace('from', str(conn)).replace('id', str(id))
    n = len(tmp) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'r'
    return header + tmp.encode()

def post(s, payload):
    s.send(payload)

duration = int(sys.argv[1])
rate = int(sys.argv[2])
total = duration * rate
precision = 100
interval = 1.0 / rate
pool = 10

s = [0] * pool
for i in range(pool):
    s[i] = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s[i].connect(batchers[0])

start = time.time()
last = time.time()

cnt = 0
diff = 0
while cnt < total:
    if time.time() - last > interval * pool * precision - diff:
        a = time.time()
        for i in range(precision):
            for j in range(pool):
                cnt += 1
                post(s[j], build(j, cnt))
        diff = time.time() - a
        last = time.time()
print(time.time() - start)

for j in range(pool):
    s[j].close()
