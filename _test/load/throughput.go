package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var host = "localhost:9000"
var batchSize = 1000
var batch = 1000
var poolSize = 10
var speed = 1000 // records per second
var depLevel = 2
var waitTime time.Duration
var batcherClient []batcherrpc.BatcherRPCClient

var (
	fHost      = kingpin.Flag("host", "Batcher hostname.").Short('h').String()
	fBatchSize = kingpin.Flag("batch_size", "Number of records per batch.").Short('n').Int()
	fBatch     = kingpin.Flag("batch_num", "Number of batches.").Short('b').Int()
	fSpeed     = kingpin.Flag("speed", "Number of records sent per second.").Short('s').Int()
	fDepLevel  = kingpin.Flag("dep_level", "The level of dependency.").Short('l').Int()
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
	if *fDepLevel > 0 {
		depLevel = *fDepLevel
	}
	if *fHost != "" {
		host = *fHost
	}
	waitTime = time.Second * time.Duration(batchSize) / time.Duration(speed)
	fmt.Println("Send " + strconv.Itoa(speed) + " records per second.")
	fmt.Println("Send a batch per " + fmt.Sprint(waitTime) + ".")
	fmt.Println("Host: ", host)
	batcherClient = make([]batcherrpc.BatcherRPCClient, poolSize)
	for i := 0; i < poolSize; i++ {
		conn, err := grpc.Dial(host, grpc.WithInsecure())
		if err != nil {
			panic("couldn't connect to batcher" + err.Error())
		}
		batcherClient[i] = batcherrpc.NewBatcherRPCClient(conn)
	}
	var wg sync.WaitGroup
	wg.Add(batch * depLevel)

	lastSent := time.Now()
	for b := 0; b < batch; b++ {
		id := make([][]string, depLevel)
		for i := range id {
			id[i] = make([]string, batchSize)
		}
		seed := make([]string, batchSize)
		for i := 0; i < batchSize; i++ {
			for l := range id {
				id[l][i] = uuid.New().String()
			}
			seed[i] = uuid.New().String()
		}

		for time.Since(lastSent) < waitTime {
			//
		}

		for l := 0; l < depLevel; l++ {
			ll := l
			go func() {
				defer wg.Done()
				records := batcherrpc.RPCRecords{
					Records: make([]*batcherrpc.RPCRecord, batchSize),
				}
				for i := range records.Records {
					records.Records[i] = &batcherrpc.RPCRecord{
						Id:     id[ll][i],
						Host:   1,
						Tags:   map[string]string{},
						Parent: "",
						Seed:   seed[i],
					}
					if ll > 0 {
						records.Records[i].Parent = id[ll-1][i]
					}
				}
				_, err := batcherClient[batch%poolSize].ReceiveRecords(context.Background(), &records)
				if err != nil {
					panic("couldn't send to batcher" + err.Error())
				}
			}()
		}

		lastSent = time.Now()
	}
	wg.Wait()
}
