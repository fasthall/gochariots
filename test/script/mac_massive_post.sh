#!/bin/sh

function send {
	a=$(gdate +%s%N)
	#for ((i=1;i<=$1;i++));
	while true;
	do
		curl -i -XPOST -H "Content-Type: application/json" -d "@../example.json" http://localhost:8080/record
	done
	b=$(gdate +%s%N)
	echo $((($b - $a) / 1000)) us
}

send $1
