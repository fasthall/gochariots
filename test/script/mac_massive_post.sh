#!/bin/sh
a=$(gdate +%s%N)
for ((i=1;i<=$1;i++));
do
	curl -i -XPOST -H "Content-Type: application/json" -d "@../example.json" http://localhost:8080/record
done
b=$(gdate +%s%N)
echo $((($b - $a) / 1000)) us
