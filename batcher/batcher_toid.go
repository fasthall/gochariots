// Package batcher batches records sent from applications or other datacenters.
// When the buffer is full the records then are sent to corresponding filter.
package batcher

import (
	"math/rand"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/queue"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

var TOIDbuffer []record.TOIDRecord

func (s *Server) TOIDReceiveRecord(ctx context.Context, in *batcherrpc.RPCRecord) (*batcherrpc.RPCReply, error) {
	r := record.TOIDRecord{
		Id:        in.GetId(),
		Timestamp: in.GetTimestamp(),
		Host:      in.GetHost(),
		TOId:      in.GetToid(),
		LId:       in.GetLid(),
		Tags:      in.GetTags(),
		Pre: record.TOIDCausality{
			Host: in.GetCausality().GetHost(),
			TOId: in.GetCausality().GetToid(),
		},
	}
	if r.Host == 0 {
		r.Host = uint32(info.ID)
	}
	TOIDarrival(r)
	return &batcherrpc.RPCReply{Message: "ok"}, nil
}

func (s *Server) TOIDReceiveRecords(ctx context.Context, in *batcherrpc.RPCRecords) (*batcherrpc.RPCReply, error) {
	for _, i := range in.GetRecords() {
		r := record.TOIDRecord{
			Id:        i.GetId(),
			Timestamp: i.GetTimestamp(),
			Host:      i.GetHost(),
			TOId:      i.GetToid(),
			LId:       i.GetLid(),
			Tags:      i.GetTags(),
			Pre: record.TOIDCausality{
				Host: i.GetCausality().GetHost(),
				TOId: i.GetCausality().GetToid(),
			},
		}
		if r.Host == 0 {
			r.Host = uint32(info.ID)
		}
		TOIDarrival(r)
	}
	return &batcherrpc.RPCReply{Message: "ok"}, nil
}

func (s *Server) TOIDUpdateQueue(ctx context.Context, in *batcherrpc.RPCQueues) (*batcherrpc.RPCReply, error) {
	return s.UpdateQueue(ctx, in)
}

// InitBatcher allocates n buffers, where n is the number of filters
func TOIDInitBatcher() {
	TOIDbuffer = make([]record.TOIDRecord, 0, bufferSize)
}

// arrival buffers arriving records.
// Upon records arrive, depends on where the record origins it goes to a certain buffer.
// When a buffer is full, all the records in the buffer will be sent to the corresponding filter.
// BUG(fasthall) In Arrival(), the mechanism to match the record and filter needs to be done. Currently the number of filters needs to be equal to datacenters.
func TOIDarrival(r record.TOIDRecord) {
	// r.Timestamp = time.Now()
	bufMutex.Lock()
	TOIDbuffer = append(TOIDbuffer, r)

	// if the buffer is full, send all records to the filter
	if len(TOIDbuffer) == cap(TOIDbuffer) {
		TOIDsendToQueue()
	}
	bufMutex.Unlock()
}

func TOIDsendToQueue() {
	if len(TOIDbuffer) == 0 {
		return
	}
	// logrus.WithField("timestamp", time.Now()).Debug("sendToQueue")
	rpcRecords := queue.RPCRecords{}
	for _, r := range TOIDbuffer {
		rpcRecords.Records = append(rpcRecords.Records, &queue.RPCRecord{
			Id:        r.Id,
			Timestamp: r.Timestamp,
			Host:      r.Host,
			Toid:      r.TOId,
			Lid:       r.LId,
			Tags:      r.Tags,
			Causality: &queue.RPCCausality{
				Host: r.Pre.Host,
				Toid: r.Pre.TOId,
			},
		})
	}
	TOIDbuffer = TOIDbuffer[:0]

	queueID := rand.Intn(len(queuePool))
	_, err := queueClient[queueID].TOIDReceiveRecords(context.Background(), &rpcRecords)
	if err != nil {
		logrus.WithField("id", queueID).Error("couldn't connect to queue")
	} else {
		logrus.WithField("id", queueID).Debug("sent to queue")
	}
}

// Sweeper periodcally sends the buffer content to filters
func TOIDSweeper() {
	for {
		time.Sleep(10 * time.Millisecond)
		bufMutex.Lock()
		TOIDsendToQueue()
		bufMutex.Unlock()
	}
}
