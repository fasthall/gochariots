#!/bin/sh

kill $(cat *.pid)
rm -rf *.pid *.log flstore/
