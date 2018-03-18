#!/bin/sh

rm -rf *.pid logs/ flstore/ *.db

kill $(ps -A | grep gochariots | awk '{print $1;}')
