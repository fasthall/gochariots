#!/bin/bash

ip=$1
delay=$2

sudo tc qdisc add dev eth0 root handle 1: prio
sudo tc filter add dev eth0 parent 1:0 protocol ip prio 1 u32 match ip dst $ip flowid 2:1
sudo tc qdisc add dev eth0 parent 1:1 handle 2: netem delay $delay
