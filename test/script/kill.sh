#!/bin/sh

rm -rf *.pid logs/ flstore/

kill $(ps -A | grep gochariots | awk '{print $1;}')
