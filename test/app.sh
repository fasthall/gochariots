#!/bin/sh

rm app.log
nohup go run main.go app 8080 > app.log &

tail -f app.log
