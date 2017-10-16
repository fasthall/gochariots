// Package batcher batches records sent from applications or other datacenters.
// When the buffer is full the records then are sent to corresponding filter.
package batcher

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/fasthall/gochariots/info"
	"github.com/satori/uuid"

	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/fasthall/gochariots/queue"
	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

const bufferSize int = 256

var bufMutex sync.Mutex
var buffer []record.Record
var connMutex sync.Mutex
var queueClient []queue.QueueClient
var queuePool []string
var queuePoolVer int
var numFilters int

type Server struct{}

func (s *Server) ReceiveRecord(ctx context.Context, in *batcherrpc.RPCRecord) (*batcherrpc.RPCReply, error) {
	r := record.Record{
		Id:        in.GetId(),
		Timestamp: in.GetTimestamp(),
		Host:      in.GetHost(),
		LId:       in.GetLid(),
		Tags:      in.GetTags(),
		Parent:    in.GetParent(),
		Seed:      in.GetSeed(),
	}
	if r.Host == 0 {
		r.Host = uint32(info.ID)
	}
	if r.Id == "" {
		r.Id = uuid.NewV4().String()
	}
	arrival(r)
	return &batcherrpc.RPCReply{Message: r.Id}, nil
}

func (s *Server) ReceiveRecords(ctx context.Context, in *batcherrpc.RPCRecords) (*batcherrpc.RPCReply, error) {
	ids := []string{}
	for _, i := range in.GetRecords() {
		r := record.Record{
			Id:        i.GetId(),
			Timestamp: i.GetTimestamp(),
			Host:      i.GetHost(),
			LId:       i.GetLid(),
			Tags:      i.GetTags(),
			Parent:    i.GetParent(),
			Seed:      i.GetSeed(),
		}
		if r.Host == 0 {
			r.Host = uint32(info.ID)
		}
		if r.Id == "" {
			r.Id = uuid.NewV4().String()
		}
		ids = append(ids, r.Id)
		arrival(r)
	}
	b, err := json.Marshal(ids)
	return &batcherrpc.RPCReply{Message: string(b)}, err
}

func (s *Server) UpdateQueue(ctx context.Context, in *batcherrpc.RPCQueues) (*batcherrpc.RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > queuePoolVer {
		queuePoolVer = ver
		queuePool = in.GetQueues()
		queueClient = make([]queue.QueueClient, len(queuePool))
		for i := range queuePool {
			conn, err := grpc.Dial(queuePool[i], grpc.WithInsecure())
			if err != nil {
				reply := batcherrpc.RPCReply{
					Message: "couldn't connect to queue",
				}
				return &reply, err
			}
			queueClient[i] = queue.NewQueueClient(conn)
		}
		logrus.WithField("queues", queuePool).Info("received new queue update")
	} else {
		logrus.WithFields(logrus.Fields{"current": queuePoolVer, "received": ver}).Debug("receiver older version of queue list")
	}
	return &batcherrpc.RPCReply{Message: "ok"}, nil
}

// InitBatcher allocates n buffers, where n is the number of filters
func InitBatcher() {
	buffer = make([]record.Record, 0, bufferSize)
}

// arrival buffers arriving records.
// Upon records arrive, depends on where the record origins it goes to a certain buffer.
// When a buffer is full, all the records in the buffer will be sent to the corresponding filter.
func arrival(r record.Record) {
	// r.Timestamp = time.Now()
	bufMutex.Lock()
	buffer = append(buffer, r)

	// if the buffer is full, send all records to the filter
	if len(buffer) == cap(buffer) {
		sendToQueue()
	}
	bufMutex.Unlock()
}

func sendToQueue() {
	if len(buffer) == 0 {
		return
	}
	// logrus.WithField("timestamp", time.Now()).Debug("sendToQueue")
	rpcRecords := queue.RPCRecords{}
	for _, r := range buffer {
		rpcRecords.Records = append(rpcRecords.Records, &queue.RPCRecord{
			Id:        r.Id,
			Timestamp: r.Timestamp,
			Host:      r.Host,
			Lid:       r.LId,
			Tags:      r.Tags,
			Parent:    r.Parent,
			Seed:      r.Seed,
		})
	}
	buffer = buffer[:0]

	queueID := rand.Intn(len(queuePool))
	_, err := queueClient[queueID].ReceiveRecords(context.Background(), &rpcRecords)
	if err != nil {
		logrus.WithField("id", queueID).Error("couldn't connect to queue")
	} else {
		logrus.WithField("id", queueID).Debug("sent to queue")
	}
}

// Sweeper periodcally sends the buffer content to filters
func Sweeper() {
	for {
		time.Sleep(10 * time.Millisecond)
		bufMutex.Lock()
		sendToQueue()
		bufMutex.Unlock()
	}
}
