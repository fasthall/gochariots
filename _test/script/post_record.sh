#!/bin/sh
curl -i -XPOST -H "Content-Type: application/json" -d "@$1" http://localhost:8080/record
