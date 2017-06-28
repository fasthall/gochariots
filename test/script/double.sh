#!/bin/sh

sh kill.sh
mkdir logs

# Cluster A
nohup gochariots controller 8081 2 0 -i > /dev/null &
nohup gochariots app 8080 2 0 -i -f config1.yaml > logs/$0.log &
nohup gochariots batcher 9000 2 0 -i -f config1.yaml > /dev/null &
nohup gochariots filter 9010 2 0 -i -f config1.yaml > /dev/null &
nohup gochariots filter 9011 2 0 -i -f config1.yaml > /dev/null &
nohup gochariots queue --hold 9020 2 0 -i -f config1.yaml > /dev/null &
nohup gochariots maintainer 9030 2 0 -i -f config1.yaml > /dev/null &
nohup gochariots indexer 9040 2 0 -i -f config1.yaml > /dev/null &

# Cluster B
nohup gochariots controller 8181 2 1 -i > /dev/null &
nohup gochariots app 8180 2 1 -i -f config2.yaml > logs/$0.log &
nohup gochariots batcher 9100 2 1 -i -f config2.yaml > /dev/null &
nohup gochariots filter 9110 2 1 -i -f config2.yaml > /dev/null &
nohup gochariots filter 9111 2 1 -i -f config2.yaml > /dev/null &
nohup gochariots queue --hold 9120 2 1 -i -f config2.yaml > /dev/null &
nohup gochariots maintainer 9130 2 1 -i -f config2.yaml > /dev/null &
nohup gochariots indexer 9140 2 1 -i -f config2.yaml > /dev/null &

# Report remote batcher
curl -XPOST localhost:8081/remote/batcher?dc=1\&host=localhost:9100
curl -XPOST localhost:8181/remote/batcher?dc=0\&host=localhost:9000

