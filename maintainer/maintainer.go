package maintainer

import (
	"github.com/fasthall/gochariots/maintainer/adapter/mongodb"
	"github.com/fasthall/gochariots/misc"

	"google.golang.org/grpc"

	"github.com/fasthall/gochariots/batcher/batcherrpc"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

const batchSize int = 10000

var benchmark misc.Benchmark

type Server struct{}

func (s *Server) ReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.Record, len(in.GetRecords()))
	propagated := []record.Record{}
	for i, ri := range in.GetRecords() {
		records[i] = record.Record{
			Id:        ri.GetId(),
			Timestamp: ri.GetTimestamp(),
			Host:      ri.GetHost(),
			LId:       ri.GetLid(),
			Tags:      ri.GetTags(),
			Parent:    ri.GetParent(),
			Seed:      ri.GetSeed(),
		}
		if ri.GetHost() == uint32(info.ID) {
			propagated = append(propagated, records[i])
		}
	}
	go func() {
		err := mongodb.PutRecords(records)
		if err != nil {
			logrus.WithError(err).Error("couldn't put records to mongodb")
		} else {
			benchmark.Logging(len(records))
		}
	}()
	go Propagate(propagated)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) UpdateBatchers(ctx context.Context, in *RPCBatchers) (*RPCReply, error) {
	ver := in.GetVersion()
	if ver > remoteBatcherVer {
		remoteBatcherVer = ver
		remoteBatchers = in.GetBatcher()
		remoteBatchersClient = make([]batcherrpc.BatcherRPCClient, len(remoteBatchers))
		for i := range remoteBatchersClient {
			conn, err := grpc.Dial(remoteBatchers[i], grpc.WithInsecure())
			if err != nil {
				return nil, err
			}
			logrus.WithFields(logrus.Fields{"id": i, "host": remoteBatchers[i]}).Debug("remote batcher client connected")
			remoteBatchersClient[i] = batcherrpc.NewBatcherRPCClient(conn)
		}
		logrus.WithField("host", in.GetBatcher()).Info("received remote batcher hosts update")
	} else {
		logrus.WithFields(logrus.Fields{"current": remoteBatcherVer, "received": ver}).Debug("receiver older version of remote batcher hosts")
	}
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) ReadByLId(ctx context.Context, in *RPCLId) (*RPCReply, error) {
	return &RPCReply{Message: "not implemented"}, nil
}

// InitLogMaintainer initializes the maintainer and assign the path name to store the records
func InitLogMaintainer(benchmarkAccuracy int) {
	benchmark = misc.NewBenchmark(benchmarkAccuracy)
}

func AssignToMaintainer(LId uint32, numMaintainers int) int {
	return int((LId - 1) / uint32(batchSize) % uint32(numMaintainers))
}
