#!/bin/sh

rm controller.log
nohup gochariots-controller 8081 >> controller.log &

tail -f controller.log
