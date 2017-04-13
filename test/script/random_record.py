import sys
import json
import time
import random
import requests
import _thread

url = ['http://localhost:8080/record', 'http://localhost:8180/record']

def build_json(key, value, prehost, pretoid):
    payload = {'tags': {key: value}, 'prehost': prehost, 'pretoid': pretoid}
    return json.dumps(payload)

def send_json(payload, host_id):
    host = url[host_id]
    headers = {'content-type': 'application/json'}
    return requests.post(host, data=payload, headers=headers)

if __name__ == '__main__':
    if len(sys.argv) < 5:
        print('Usage: python random_record.py num_records dependency_prob max_window margin')
        exit(0)
    n = int(sys.argv[1])
    dependency_prob = float(sys.argv[2])
    max_window = int(sys.argv[3])
    margin = int(sys.argv[4])
    for i in range(n):
        print(i)
        if random.random() < dependency_prob:
            key = 'low'
            pretoid = int(max(0, (i - random.randint(1, max_window) - margin) / len(url)))
        else:
            key = 'high'
            pretoid = 0
        host_id = random.randint(0, len(url) - 1)
        # host_id = i % 2
        payload = build_json(key, str(i + 1), host_id, pretoid)
        # result = send_json(payload, host_id)
        # code = result.status_code
        _thread.start_new_thread(send_json, (payload, host_id))
        