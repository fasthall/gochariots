#!/bin/sh

sh kill.sh
mkdir logs

# Cluster A
nohup gochariots controller -p 8081 2 0 -i $1 > /dev/null &
nohup gochariots app -p 8080 2 0 -i $1 -f config1.yaml > logs/$0.log &
nohup gochariots batcher -p 9000 2 0 -i $1 -f config1.yaml > /dev/null &
nohup gochariots filter -p 9010 2 0 -i $1 -f config1.yaml > /dev/null &
nohup gochariots filter -p 9011 2 0 -i $1 -f config1.yaml > /dev/null &
nohup gochariots queue --hold -p 9020 2 0 -i $1 -f config1.yaml > /dev/null &
nohup gochariots maintainer -p 9030 2 0 -i $1 -f config1.yaml > /dev/null &
nohup gochariots indexer -p 9040 2 0 -i $1 -f config1.yaml > /dev/null &

# Cluster B
nohup gochariots controller -p 8181 2 1 -i $1 > /dev/null &
nohup gochariots app -p 8180 2 1 -i $1 -f config2.yaml > logs/$0.log &
nohup gochariots batcher -p 9100 2 1 -i $1 -f config2.yaml > /dev/null &
nohup gochariots filter -p 9110 2 1 -i $1 -f config2.yaml > /dev/null &
nohup gochariots filter -p 9111 2 1 -i $1 -f config2.yaml > /dev/null &
nohup gochariots queue --hold -p 9120 2 1 -i $1 -f config2.yaml > /dev/null &
nohup gochariots maintainer -p 9130 2 1 -i $1 -f config2.yaml > /dev/null &
nohup gochariots indexer -p 9140 2 1 -i $1 -f config2.yaml > /dev/null &

# Report remote batcher
curl -XPOST localhost:8081/remote/batcher?dc=1\&host=localhost:9100
curl -XPOST localhost:8181/remote/batcher?dc=0\&host=localhost:9000

