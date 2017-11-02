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
	wg.Add(batch)
	for b := 0; b < batch; b++ {
		go func() {
			defer wg.Done()
			records := batcherrpc.RPCRecords{
				Records: make([]*batcherrpc.RPCRecord, batchSize),
			}
			for i := range records.Records {
				records.Records[i] = &batcherrpc.RPCRecord{
					Id:   uuid.NewV4().String(),
					Host: 1,
					Tags: map[string]string{},
					Causality: &batcherrpc.RPCCausality{
						Host: 0,
						Toid: 0,
					},
				}
			}
			_, err := batcherClient[batch%poolSize].TOIDReceiveRecords(context.Background(), &records)
			if err != nil {
				panic("couldn't send to batcher" + err.Error())
			}
		}()
	}
	wg.Wait()
}