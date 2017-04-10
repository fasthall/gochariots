import json
import sys

'''
For checking if there is conflict dependency.
python check.py num_records
'''
n = int(sys.argv[1])
mask = [False] * (n + 1)
mask[0] = True
for i in range(1, n + 1):
	f = open(str(i), 'r')
	j = json.load(f)
	mask[j['TOId']] = True
	pretoid = j['Pre']['TOId']
	if mask[pretoid] == False:
		print('Record ' + str(pretoid) + ' not appended yet')
