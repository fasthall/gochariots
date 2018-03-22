package maintainer

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

var xxx = 0
var mu sync.Mutex

func (s *Server) TOIDReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.TOIDRecord, len(in.GetRecords()))
	propagated := []record.TOIDRecord{}
	for i, ri := range in.GetRecords() {
		records[i] = record.TOIDRecord{
			Id:        ri.GetId(),
			Timestamp: ri.GetTimestamp(),
			Host:      ri.GetHost(),
			TOId:      ri.GetToid(),
			LId:       ri.GetLid(),
			Tags:      ri.GetTags(),
			Pre: record.TOIDCausality{
				Host: ri.GetCausality().GetHost(),
				TOId: ri.GetCausality().GetToid(),
			},
		}
		if records[i].Host == uint32(info.ID) {
			propagated = append(propagated, records[i])
		}
	}
	go func() {
		// mongodb sharding
		err := mongoClient[0].PutTOIDRecords(records)
		if err != nil {
			logrus.WithError(err).Error("couldn't put records to mongodb")
		} else {
			benchmark.Logging(len(records))
		}
	}()
	go TOIDPropagate(propagated)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) TOIDUpdateBatchers(ctx context.Context, in *RPCBatchers) (*RPCReply, error) {
	return s.UpdateBatchers(ctx, in)
}

func (s *Server) TOIDReadByLId(ctx context.Context, in *RPCLId) (*RPCReply, error) {
	return &RPCReply{Message: "not implemented"}, nil
}
