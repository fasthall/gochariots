import sys
import time
import socket

batchers = [('localhost', 9000)]
payload = '{"Host":0,"TOId":0,"LId":0,"Tags":{"from":"id"},"Pre":{"Host":0,"TOId":0,"Tags":{}}},'
b = 1

def build(conn, id):
    tmp = payload * b
    tmp = '[' + tmp[:-1] + ']'
    n = len(tmp) + 1
    if id % 100000 == 0:
        print(id)
    header = n.to_bytes(4, byteorder='big')
    header += b's'
    return header + tmp.encode()

def post(s, payload):
    s.send(payload)

total = int(sys.argv[1])
pool = 10

s = [0] * pool
for i in range(pool):
    s[i] = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s[i].connect(batchers[0])

cnt = 0
diff = 0
start = time.time()
while cnt < total:
    for j in range(pool):
        cnt += b
        post(s[j], build(j, cnt))
print(time.time() - start)

for j in range(pool):
    s[j].close()
