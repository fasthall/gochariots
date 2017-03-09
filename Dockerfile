FROM golang:1.8-alpine
MAINTAINER Wei-Tsung Lin <fasthall@gmail.com>

RUN apk update
RUN apk add git
RUN go get github.com/fasthall/gochariots/...

