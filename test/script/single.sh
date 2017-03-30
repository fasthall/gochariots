#!/bin/sh

nohup gochariots-app 8080 1 0 > $0.log &
nohup gochariots-controller 8081 1 0  >> $0.log &

sleep 1

nohup gochariots-batcher 9000 1 0 >> $0.log &
curl -XPOST localhost:8080/batcher?host=localhost:9000
curl -XGET localhost:8080/batcher
curl -XPOST localhost:8081/batcher?host=localhost:9000
curl -XGET localhost:8081/batcher

nohup gochariots-filter 9010 1 0 >> $0.log &
curl -XPOST localhost:8081/filter?host=localhost:9010
curl -XGET localhost:8081/filter

nohup gochariots-queue 9020 1 0 true >> $0.log &

curl -XPOST localhost:8081/queue?host=localhost:9020

nohup gochariots-maintainer 9030 1 0 >> $0.log &

curl -XPOST localhost:8081/maintainer?host=localhost:9030
curl -XGET localhost:8081/maintainer

tail -f $0.log
