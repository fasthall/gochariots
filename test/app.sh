#!/bin/sh

rm app.log
nohup gochariots-app 8080 > app.log &

tail -f app.log
