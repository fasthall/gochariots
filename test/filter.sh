#!/bin/sh

rm filter.log
nohup go run main.go filter 9010 >> filter.log &

curl -XPOST localhost:8081/filter?host=localhost:9010
curl -XGET localhost:8081/filter

tail -f filter.log
