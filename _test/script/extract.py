import sys
import json
import csv
from datetime import datetime

f = open(sys.argv[1], 'rb')
f.seek(512)
o = open(sys.argv[2], 'w')
n = int(sys.argv[3])
w = csv.writer(o, delimiter=',', quotechar='"', quoting=csv.QUOTE_MINIMAL)
w.writerow(['LId 1', 'Timestamp 1', 'LId 2', 'Timestamp 2'])

data = []
for i in range(n):
    data.append([None] * 4)
for i in range(n * 2):
    print(i)
    tmp = f.read(4)
    len = int.from_bytes(tmp, byteorder='big')
    tmp = f.read(len)
    j = json.loads(tmp.decode())
    event = list(j['Tags'].keys())[0]
    id = j['Tags'][event]
    if id == '1':
        data[int(event)][0] = j['LId']
        data[int(event)][1] = j['Timestamp']
    elif id == '2':
        data[int(event)][2] = j['LId']
        data[int(event)][3] = j['Timestamp']
    f.seek(512 - 4 - len, 1)

for i in range(n):
    print(i)
    w.writerow(data[i])

f.close()
o.close()