#!/bin/sh

rm maintainer.log

nohup gochariots-maintainer 9030 >> maintainer.log &
nohup gochariots-maintainer 9031 >> maintainer.log &
nohup gochariots-maintainer 9032 >> maintainer.log &

sleep 1

curl -XPOST localhost:8081/maintainer?host=localhost:9030
curl -XPOST localhost:8081/maintainer?host=localhost:9031
curl -XPOST localhost:8081/maintainer?host=localhost:9032
curl -XGET localhost:8081/maintainer

tail -f maintainer.log
