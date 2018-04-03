FROM golang:1.8-alpine
MAINTAINER Wei-Tsung Lin <fasthall@gmail.com>

RUN apk update
RUN apk add git
RUN go get github.com/Sirupsen/logrus
RUN go get go get github.com/google/uuid
RUN go get github.com/gin-gonic/gin
RUN go get gopkg.in/mgo.v2
RUN go get google.golang.org/grpc
RUN go get gopkg.in/alecthomas/kingpin.v2
COPY . /go/src/github.com/fasthall/gochariots/
RUN go install github.com/fasthall/gochariots
