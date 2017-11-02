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

var TOIDbuffer chan record.TOIDRecord

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
	go TOIDarrival(r)
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
		go TOIDarrival(r)
	}
	return &batcherrpc.RPCReply{Message: "ok"}, nil
}

func (s *Server) TOIDUpdateQueue(ctx context.Context, in *batcherrpc.RPCQueues) (*batcherrpc.RPCReply, error) {
	return s.UpdateQueue(ctx, in)
}

// InitBatcher allocates n buffers, where n is the number of filters
func TOIDInitBatcher(bs int) {
	bufferSize = bs
	TOIDbuffer = make(chan record.TOIDRecord, bufferSize)
	bufferFull = make(chan bool)
	go TOIDSweeper()
}

// arrival buffers arriving records.
// Upon records arrive, depends on where the record origins it goes to a certain buffer.
// When a buffer is full, all the records in the buffer will be sent to the corresponding filter.
// BUG(fasthall) In Arrival(), the mechanism to match the record and filter needs to be done. Currently the number of filters needs to be equal to datacenters.
func TOIDarrival(r record.TOIDRecord) {
	// r.Timestamp = time.Now()
	for {
		select {
		case TOIDbuffer <- r:
			// send record into buffered channel
			return
		default:
			bufferFull <- true
		}
	}
}

func TOIDsendToQueue() {
	if len(TOIDbuffer) == 0 {
		return
	}
	// logrus.WithField("timestamp", time.Now()).Debug("sendToQueue")
	rpcRecords := queue.RPCRecords{}
	done := false
	for !done {
		select {
		case r := <-TOIDbuffer:
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
			if len(rpcRecords.Records) == bufferSize {
				done = true
			}
		default:
			done = true
		}
	}

	go func() {
		queueID := rand.Intn(len(queuePool))
		_, err := queueClient[queueID].TOIDReceiveRecords(context.Background(), &rpcRecords)
		if err != nil {
			logrus.WithField("id", queueID).Error("couldn't connect to queue")
		} else {
			logrus.WithField("id", queueID).Debug("sent to queue")
		}
	}()
}

// Sweeper periodcally sends the buffer content to filters
func TOIDSweeper() {
	cron := time.NewTicker(10 * time.Millisecond)
	for {
		select {
		case <-cron.C:
			TOIDsendToQueue()
		case <-bufferFull:
			TOIDsendToQueue()
		}
	}
}
