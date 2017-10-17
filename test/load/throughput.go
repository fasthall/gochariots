package main

import (
	"context"

	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/satori/uuid"

	"google.golang.org/grpc"
)

var batch = 100
var batcherClient batcherrpc.BatcherRPCClient

func main() {
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		panic("couldn't connect to batcher")
	}
	batcherClient = batcherrpc.NewBatcherRPCClient(conn)

	records := batcherrpc.RPCRecords{
		Records: make([]*batcherrpc.RPCRecord, batch),
	}
	for i := range records.Records {
		records.Records[i] = &batcherrpc.RPCRecord{
			Id:     uuid.NewV4().String(),
			Host:   1,
			Tags:   map[string]string{},
			Parent: "",
			Seed:   uuid.NewV4().String(),
		}
	}

	_, err = batcherClient.ReceiveRecords(context.Background(), &records)
	if err != nil {
		panic("couldn't connect to batcher")
	}
}
