#!/bin/sh

rm -rf *.pid *.log flstore/

kill $(ps -A | grep gochariots | awk '{print $1;}')
