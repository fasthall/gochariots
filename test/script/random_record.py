import sys
import json
import time
import random
import requests

max_window = 100
dependency_prob = 0.8

def build_json(key, value, prehost, pretoid):
    payload = {'tags': {key: value}, 'prehost': prehost, 'pretoid': pretoid}
    return json.dumps(payload)

def send_json(payload):
    url = 'http://localhost:8080/record'
    headers = {'content-type': 'application/json'}
    print(payload)
    return requests.post(url, data=payload, headers=headers)

if __name__ == '__main__':
    n = int(sys.argv[1])
    for i in range(n):
        code = 503
        while code == 503:
            if random.random() < dependency_prob:
                key = 'low'
                pretoid = max(0, i - random.randint(1, max_window))
            else:
                key = 'high'
                pretoid = 0
            payload = build_json(key, str(i + 1), 0, pretoid)
            result = send_json(payload)
            code = result.status_code
            if code == 503:
                time.sleep(3)
