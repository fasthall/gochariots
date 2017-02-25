#!/bin/sh

rm queue.log
nohup go run main.go queue 9020 true >> queue.log &
nohup go run main.go queue 9021 false >> queue.log &
nohup go run main.go queue 9022 false >> queue.log &

sleep 1

curl -XPOST localhost:8081/queue?host=localhost:9020
curl -XPOST localhost:8081/queue?host=localhost:9021
curl -XPOST localhost:8081/queue?host=localhost:9022

tail -f queue.log
