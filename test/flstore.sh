#!/bin/sh

rm flstore.log
nohup go run main.go maintainer 9030 >> flstore.log &
nohup go run main.go maintainer 9031 >> flstore.log &
nohup go run main.go maintainer 9032 >> flstore.log &

sleep 1

curl -XPOST localhost:8081/maintainer?host=localhost:9030
curl -XPOST localhost:8081/maintainer?host=localhost:9031
curl -XPOST localhost:8081/maintainer?host=localhost:9032
curl -XGET localhost:8081/maintainer

tail -f flstore.log
