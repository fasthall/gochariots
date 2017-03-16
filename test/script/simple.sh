#!/bin/sh

nohup gochariots-controller 8081 >> $0.log &

nohup gochariots-app 8080 > $0.log &

sleep 1

nohup gochariots-batcher 9000 1 >> $0.log &
nohup gochariots-batcher 9001 1 >> $0.log &
nohup gochariots-batcher 9002 1 >> $0.log &
curl -XPOST localhost:8080/batcher?host=localhost:9000
curl -XPOST localhost:8080/batcher?host=localhost:9001
curl -XPOST localhost:8080/batcher?host=localhost:9002
curl -XGET localhost:8080/batcher
curl -XPOST localhost:8081/batcher?host=localhost:9000
curl -XPOST localhost:8081/batcher?host=localhost:9001
curl -XPOST localhost:8081/batcher?host=localhost:9002
curl -XGET localhost:8081/batcher

sleep 1

nohup gochariots-filter 9010 >> $0.log &
curl -XPOST localhost:8081/filter?host=localhost:9010
curl -XGET localhost:8081/filter

nohup gochariots-queue 9020 true >> $0.log &
nohup gochariots-queue 9021 false >> $0.log &
nohup gochariots-queue 9022 false >> $0.log &

sleep 1

curl -XPOST localhost:8081/queue?host=localhost:9020
curl -XPOST localhost:8081/queue?host=localhost:9021
curl -XPOST localhost:8081/queue?host=localhost:9022

nohup gochariots-maintainer 9030 >> $0.log &
nohup gochariots-maintainer 9031 >> $0.log &
nohup gochariots-maintainer 9032 >> $0.log &

sleep 1

curl -XPOST localhost:8081/maintainer?host=localhost:9030
curl -XPOST localhost:8081/maintainer?host=localhost:9031
curl -XPOST localhost:8081/maintainer?host=localhost:9032
curl -XGET localhost:8081/maintainer

tail -f $0.log