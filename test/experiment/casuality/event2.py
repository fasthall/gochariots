import socket
import time
import sys
import struct

host = (socket.gethostname(), int(sys.argv[1]))
delay = float(sys.argv[2])
batchers = [('localhost', 9000)]

ss = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
ss.bind(host)

ss.listen(0)
while True:
    s, addr = ss.accept()
    suuid = s.recv(1024).decode()
    payload = '{"Host":1,"TOId":0,"LId":0,"Tags":{"UUID":"' + suuid + '", "event":"2"},"Pre":{"Host":0,"TOId":1}}'

    time.sleep(delay / 1000)
    end = str(time.time())
    print('Append event 2 at:', end, '@B')
    s.send(end.encode())
    s.close()

    # send to batcher
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.connect(batchers[0])
    n = len(payload) + 1
    header = n.to_bytes(4, byteorder='big')
    header += b'r'
    s.send(header + payload.encode())
    s.close()