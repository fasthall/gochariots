// Package batcher batches records sent from applications or other datacenters.
// When the buffer is full the records then are sent to corresponding filter.
package batcher

import (
	"math/rand"
	"net"
	"sync"
	"time"

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
var queueConn []net.Conn
var queuePool []string
var queuePoolVer int
var numFilters int

type Server struct{}

func (s *Server) ReceiveRecord(ctx context.Context, in *RPCRecord) (*RPCReply, error) {
	r := record.Record{
		Timestamp: in.GetTimestamp(),
		Host:      int(in.GetHost()),
		LId:       int(in.GetLid()),
		Tags:      in.GetTags(),
		Hash:      in.GetHash(),
		Seed:      in.GetSeed(),
	}
	arrival(r)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) ReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	for _, i := range in.GetRecords() {
		r := record.Record{
			Timestamp: i.GetTimestamp(),
			Host:      int(i.GetHost()),
			LId:       int(i.GetLid()),
			Tags:      i.GetTags(),
			Hash:      i.GetHash(),
			Seed:      i.GetSeed(),
		}
		arrival(r)
	}
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) UpdateQueue(ctx context.Context, in *RPCQueues) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > queuePoolVer {
		queuePoolVer = ver
		queuePool = in.GetQueues()
		queueClient = make([]queue.QueueClient, len(queuePool))
		for i := range queuePool {
			conn, err := grpc.Dial(queuePool[i], grpc.WithInsecure())
			if err != nil {
				reply := RPCReply{
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
	return &RPCReply{Message: "ok"}, nil
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

func dialConn(queueID int) error {
	host := queuePool[queueID]
	var err error
	queueConn[queueID], err = net.Dial("tcp", host)
	return err
}

func sendToQueue() {
	if len(buffer) == 0 {
		return
	}
	// logrus.WithField("timestamp", time.Now()).Debug("sendToQueue")
	rpcRecords := queue.RPCRecords{}
	for _, r := range buffer {
		rpcRecords.Records = append(rpcRecords.Records, &queue.RPCRecord{
			Timestamp: r.Timestamp,
			Host:      int32(r.Host),
			Lid:       int32(r.LId),
			Tags:      r.Tags,
			Hash:      r.Hash,
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
