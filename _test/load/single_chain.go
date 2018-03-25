package main

import (
	"context"

	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var host = "localhost:9000"
var batcherClient batcherrpc.BatcherRPCClient
var cnt = 4

func main() {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		panic("couldn't connect to batcher" + err.Error())
	}
	batcherClient = batcherrpc.NewBatcherRPCClient(conn)
	trace := uuid.New().String()
	id := make([]string, cnt)
	for i := range id {
		id[i] = uuid.New().String()
	}

	records := batcherrpc.RPCRecords{}
	for i := range id {
		records.Records = append(records.Records, &batcherrpc.RPCRecord{
			Id:     id[i],
			Host:   1,
			Tags:   map[string]string{},
			Parent: "",
			Trace:  trace,
		})
		if i > 0 {
			records.Records[i].Parent = id[i-1]
		}
	}
	_, err = batcherClient.ReceiveRecords(context.Background(), &records)
	if err != nil {
		panic("couldn't send to batcher" + err.Error())
	}
}
