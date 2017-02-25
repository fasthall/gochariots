#!/bin/sh

rm controller.log
nohup go run main.go controller 8081 >> controller.log &

tail -f controller.log
