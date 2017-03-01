#!/bin/sh

rm batcher.log
nohup gochariots-batcher 9000 >> batcher.log &
nohup gochariots-batcher 9001 >> batcher.log &
nohup gochariots-batcher 9002 >> batcher.log &

curl -XPOST localhost:8080/batcher?host=localhost:9000
curl -XPOST localhost:8080/batcher?host=localhost:9001
curl -XPOST localhost:8080/batcher?host=localhost:9002
curl -XGET localhost:8080/batcher
curl -XPOST localhost:8081/batcher?host=localhost:9000
curl -XPOST localhost:8081/batcher?host=localhost:9001
curl -XPOST localhost:8081/batcher?host=localhost:9002
curl -XGET localhost:8081/batcher

tail -f batcher.log
