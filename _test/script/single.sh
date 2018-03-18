#!/bin/sh

sh kill.sh

mkdir logs
nohup gochariots-app 8080 1 0 > logs/$0.log &
nohup gochariots-controller 8081 1 0 > /dev/null &
nohup gochariots-batcher 9000 1 0 > /dev/null &
nohup gochariots-filter 9010 1 0 > /dev/null &
nohup gochariots-queue 9020 1 0 true > /dev/null &
nohup gochariots-maintainer 9030 1 0 > /dev/null &
nohup gochariots-indexer 9040 1 0 > /dev/null &

sleep 2

curl -XPOST localhost:8080/batcher?host=localhost:9000
curl -XPOST localhost:8081/batcher?host=localhost:9000
curl -XPOST localhost:8081/filter?host=localhost:9010
curl -XPOST localhost:8081/queue?host=localhost:9020
curl -XPOST localhost:8081/maintainer?host=localhost:9030\&indexer=localhost:9040

tail logs/$0.log
