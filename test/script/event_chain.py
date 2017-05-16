import socket
import time
import sys
import uuid
import atexit

n = int(sys.argv[1])
batcher = [('localhost', 9000), ('localhost', 9100)]

def build_payload1(suuid):
    return '{"Host":0,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"1"},"Pre":{"Host":0,"TOId":0}}'

def build_payload2(suuid):
    return '{"Host":1,"TOId":0,"LId":0,"Tags":{"' + suuid + '":"2"},"Pre":{"Host":0,"TOId":0,"Tags":{"' + suuid + '":"1"}}}'

bs1 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
bs1.connect(batcher[0])
bs2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
bs2.connect(batcher[1])

def onexit():
    bs1.close()
    bs2.close()

atexit.register(onexit)

for i in range(n):
    suuid = str(i)

    # send to batcher
    payload1 = build_payload1(suuid)
    n1 = len(payload1) + 1
    header1 = n1.to_bytes(4, byteorder='big')
    header1 += b'r'

    payload2 = build_payload2(suuid)
    n2 = len(payload2) + 1
    header2 = n2.to_bytes(4, byteorder='big')
    header2 += b'r'

    bs2.send(header2 + payload2.encode())


    time.sleep(3)
    bs1.send(header1 + payload1.encode())
