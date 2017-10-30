package maintainer

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/fasthall/gochariots/maintainer/adapter/mongodb"

	"google.golang.org/grpc"

	"github.com/fasthall/gochariots/batcher/batcherrpc"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

const batchSize int = 10000

var path = "flstore"
var f *os.File
var maintainerInterface int

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
	lid := int(in.GetLid())
	r, err := ReadByLId(lid)
	if err != nil {
		return nil, err
	}
	j, err := json.Marshal(r)
	return &RPCReply{Message: string(j)}, nil
}

// InitLogMaintainer initializes the maintainer and assign the path name to store the records
func InitLogMaintainer(p string, itf int) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		logrus.WithField("path", path).Error("couldn't access path")
	}
	f, err = os.Create(filepath.Join(path, p))
	if err != nil {
		logrus.WithField("path", p).Error("couldn't create file")
		panic(err)
	}
	maintainerInterface = itf
}

// ReadByLId reads from the maintainer according to LId.
func ReadByLId(lid int) (record.Record, error) {
	lenbuf := make([]byte, 4)
	_, err := f.ReadAt(lenbuf, int64(512*lid))
	if err != nil {
		return record.Record{}, err
	}
	length := int(binary.BigEndian.Uint32(lenbuf))
	buf := make([]byte, length)
	_, err = f.ReadAt(buf, int64(512*lid+4))
	if err != nil {
		return record.Record{}, err
	}
	var r record.Record
	err = record.JSONToRecord(buf, &r)
	return r, err
}

// ReadByLIds reads multiple records
func ReadByLIds(lids []int) ([]record.Record, error) {
	result := []record.Record{}
	for _, lid := range lids {
		r, err := ReadByLId(lid)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func AssignToMaintainer(LId uint32, numMaintainers int) int {
	return int((LId - 1) / uint32(batchSize) % uint32(numMaintainers))
}
