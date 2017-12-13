package main

import (
	"context"
	"sync"

	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/satori/uuid"
	"google.golang.org/grpc"
)

var batchSize = 1000
var batch = 1000
var poolSize = 10
var batcherClient []batcherrpc.BatcherRPCClient

func main() {
	batcherClient = make([]batcherrpc.BatcherRPCClient, poolSize)
	for i := 0; i < poolSize; i++ {
		conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
		if err != nil {
			panic("couldn't connect to batcher" + err.Error())
		}
		batcherClient[i] = batcherrpc.NewBatcherRPCClient(conn)
	}
	var wg sync.WaitGroup
	wg.Add(batch * 2)
	for b := 0; b < batch; b++ {
		id1 := make([]string, batchSize)
		id2 := make([]string, batchSize)
		seed := make([]string, batchSize)
		for i := 0; i < batchSize; i++ {
			id1[i] = uuid.NewV4().String()
			id2[i] = uuid.NewV4().String()
			seed[i] = uuid.NewV4().String()
		}
		go func() {
			defer wg.Done()
			records := batcherrpc.RPCRecords{
				Records: make([]*batcherrpc.RPCRecord, batchSize),
			}
			for i := range records.Records {
				records.Records[i] = &batcherrpc.RPCRecord{
					Id:     id1[i],
					Host:   1,
					Tags:   map[string]string{},
					Parent: "",
					Seed:   seed[i],
				}
			}
			_, err := batcherClient[batch%poolSize].ReceiveRecords(context.Background(), &records)
			if err != nil {
				panic("couldn't send to batcher" + err.Error())
			}
		}()
		go func() {
			defer wg.Done()
			records := batcherrpc.RPCRecords{
				Records: make([]*batcherrpc.RPCRecord, batchSize),
			}
			for i := range records.Records {
				records.Records[i] = &batcherrpc.RPCRecord{
					Id:     id2[i],
					Host:   1,
					Tags:   map[string]string{},
					Parent: id1[i],
					Seed:   seed[i],
				}
			}
			_, err := batcherClient[batch%poolSize].ReceiveRecords(context.Background(), &records)
			if err != nil {
				panic("couldn't send to batcher" + err.Error())
			}
		}()
	}
	wg.Wait()
}
