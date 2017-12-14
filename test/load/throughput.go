package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/satori/uuid"
	"google.golang.org/grpc"
)

var batchSize = 1000
var batch = 1000
var poolSize = 10
var speed = 1000 // records per second
var waitTime time.Duration
var batcherClient []batcherrpc.BatcherRPCClient

var (
	fBatchSize = kingpin.Flag("batch_size", "Number of records per batch.").Short('n').Int()
	fBatch     = kingpin.Flag("batch_num", "Number of batches.").Short('b').Int()
	fSpeed     = kingpin.Flag("speed", "Number of records sent per second.").Short('s').Int()
)

func main() {
	kingpin.Parse()
	if *fBatchSize > 0 {
		batchSize = *fBatchSize
	}
	if *fBatch > 0 {
		batch = *fBatch
	}
	if *fSpeed > 0 {
		speed = *fSpeed
	}
	waitTime = time.Second * time.Duration(batchSize) / time.Duration(speed)
	fmt.Println("Send " + strconv.Itoa(speed) + " records per second.")
	fmt.Println("Send a batch per " + fmt.Sprint(waitTime) + ".")
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

	lastSent := time.Now()
	for b := 0; b < batch; b++ {
		id1 := make([]string, batchSize)
		id2 := make([]string, batchSize)
		seed := make([]string, batchSize)
		for i := 0; i < batchSize; i++ {
			id1[i] = uuid.NewV4().String()
			id2[i] = uuid.NewV4().String()
			seed[i] = uuid.NewV4().String()
		}

		for time.Since(lastSent) < waitTime {
			//
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

		lastSent = time.Now()
	}
	wg.Wait()
}
