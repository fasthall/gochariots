import sys
import time
import socket

batcher = ('localhost', 9000)
maintainer = ('localhost', 9030)
payload = '{"Host":0,"TOId":0,"LId":0,"Tags":{"from":"id"},"Pre":{"Host":0,"TOId":0,"Tags":{}}}'

def build_record(conn, id):
    tmp = payload.replace('from', str(conn)).replace('id', str(id))
    n = len(tmp) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'r'
    return header + tmp.encode()

def build_query(tag, value):
    tmp = '{"'+tag+'":"'+value+'"}'
    n = len(tmp) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'i'
    return header + tmp.encode()

def post(s, payload):
    s.send(payload)

batcher_conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
batcher_conn.connect(batcher)
maintainer_conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
maintainer_conn.connect(maintainer)

query = build_query('test', 'value')
start = time.time()
post(batcher_conn, build_record('test', 'value'))

buf = []
while len(buf) <= 6:
    maintainer_conn.send(query)
    buf = maintainer_conn.recv(1024)
print(time.time() - start)
print(buf)

batcher_conn.close()
maintainer_conn.close()
