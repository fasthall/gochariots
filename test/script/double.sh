#!/bin/sh


# Cluster A
nohup gochariots-app 8080 2 0  > $0.log &
nohup gochariots-controller 8081 2 0  >> $0.log &

sleep 1

nohup gochariots-batcher 9000 2 0 >> $0.log &
curl -XPOST localhost:8080/batcher?host=localhost:9000
curl -XGET localhost:8080/batcher
curl -XPOST localhost:8081/batcher?host=localhost:9000
curl -XGET localhost:8081/batcher

sleep 1

nohup gochariots-filter 9010 2 0 >> $0.log &
nohup gochariots-filter 9011 2 0 >> $0.log &
curl -XPOST localhost:8081/filter?host=localhost:9010
curl -XPOST localhost:8081/filter?host=localhost:9011
curl -XGET localhost:8081/filter

nohup gochariots-queue 9020 2 0 true >> $0.log &

sleep 1

curl -XPOST localhost:8081/queue?host=localhost:9020

nohup gochariots-maintainer 9030 2 0 >> $0.log &

sleep 1

curl -XPOST localhost:8081/maintainer?host=localhost:9030
curl -XGET localhost:8081/maintainer

# Cluster B
nohup gochariots-app 8180 2 1 > $0.log &
nohup gochariots-controller 8181 2 1 >> $0.log &

sleep 1

nohup gochariots-batcher 9100 2 1 >> $0.log &
curl -XPOST localhost:8180/batcher?host=localhost:9100
curl -XGET localhost:8180/batcher
curl -XPOST localhost:8181/batcher?host=localhost:9100
curl -XGET localhost:8181/batcher

sleep 1

nohup gochariots-filter 9110 2 1 >> $0.log &
nohup gochariots-filter 9111 2 1 >> $0.log &
curl -XPOST localhost:8181/filter?host=localhost:9110
curl -XPOST localhost:8181/filter?host=localhost:9111
curl -XGET localhost:8181/filter

nohup gochariots-queue 9120 2 1 true >> $0.log &

sleep 1

curl -XPOST localhost:8181/queue?host=localhost:9120

nohup gochariots-maintainer 9130 2 1 >> $0.log &

sleep 1

curl -XPOST localhost:8181/maintainer?host=localhost:9130
curl -XGET localhost:8181/maintainer

# Report remote batcher
curl -XPOST localhost:8081/remote/batcher?dc=1\&host=localhost:9100
curl -XGET localhost:8081/maintainer
curl -XPOST localhost:8181/remote/batcher?dc=0\&host=localhost:9000
 curl -XGET localhost:8181/maintainer

tail -f $0.log
