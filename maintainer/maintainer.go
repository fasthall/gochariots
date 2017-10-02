package maintainer

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"

	"github.com/fasthall/gochariots/maintainer/indexer"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer/adapter"
	"github.com/fasthall/gochariots/maintainer/adapter/cosmos"
	"github.com/fasthall/gochariots/maintainer/adapter/datastore"
	"github.com/fasthall/gochariots/maintainer/adapter/dynamodb"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

const batchSize int = 1000

// LastLId records the last LID maintained by this maintainer
var LastLId int
var path = "flstore"
var f *os.File
var indexerClient indexer.IndexerClient
var indexerHost string
var indexerVer int
var maintainerInterface int

var logFirstTime time.Time
var logRecordNth int

type Server struct{}

func (s *Server) ReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.Record, len(in.GetRecords()))
	for i, ri := range in.GetRecords() {
		records[i] = record.Record{
			Timestamp: ri.GetTimestamp(),
			Host:      int(ri.GetHost()),
			LId:       int(ri.GetLid()),
			Tags:      ri.GetTags(),
			Hash:      ri.GetHash(),
			Seed:      ri.GetSeed(),
		}
	}
	recordsArrival(records)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) UpdateBatchers(ctx context.Context, in *RPCBatchers) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > remoteBatcherVer {
		remoteBatcherVer = ver
		remoteBatchers = in.GetBatcher()
		logrus.WithField("host", in.GetBatcher()).Info("received remote batcher hosts update")
	} else {
		logrus.WithFields(logrus.Fields{"current": remoteBatcherVer, "received": ver}).Debug("receiver older version of remote batcher hosts")
	}
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) UpdateIndexer(ctx context.Context, in *RPCIndexer) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > indexerVer {
		indexerVer = ver
		indexerHost = in.GetIndexer()
		conn, err := grpc.Dial(indexerHost, grpc.WithInsecure())
		if err != nil {
			reply := RPCReply{
				Message: "couldn't connect to indexer",
			}
			return &reply, err
		}
		indexerClient = indexer.NewIndexerClient(conn)
		logrus.WithField("host", in.GetIndexer()).Info("received indexer hosts update")
	} else {
		logrus.WithFields(logrus.Fields{"current": indexerVer, "received": ver}).Debug("receiver older version of indexer hosts")
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
func InitLogMaintainer(p string, n int, itf int) {
	logRecordNth = n
	LastLId = 0
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

// Append appends a new record to the maintainer.
func Append(r record.Record) error {
	// logrus.WithField("timestamp", time.Now()).Debug("Append")
	r.Timestamp = time.Now().UnixNano()
	if logRecordNth > 0 {
		if r.LId == 1 {
			logFirstTime = time.Now()
		} else if r.LId == logRecordNth {
			logrus.WithField("duration", time.Since(logFirstTime)).Info("appended", logRecordNth, "records")
		}
	}

	if maintainerInterface == adapter.DYNAMODB {
		err := dynamodb.PutRecord(r)
		if err != nil {
			return err
		}
	} else if maintainerInterface == adapter.DATASTORE {
		err := datastore.PutRecord(r)
		if err != nil {
			return err
		}
	} else if maintainerInterface == adapter.FLSTORE {
		b, err := record.ToJSON(r)
		if err != nil {
			return err
		}
		lenbuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenbuf, uint32(len(b)))
		lid := r.LId
		_, err = f.WriteAt(append(lenbuf, b...), int64(512*lid))
		if err != nil {
			return err
		}
		// log.Println(info.GetName(), "wrote record ", lid)
	} else if maintainerInterface == adapter.COSMOSDB {
		err := cosmos.PutRecord(r)
		if err != nil {
			return err
		}
	}

	InsertIndexer(r)
	LastLId = r.LId
	if r.Host == info.ID {
		Propagate(r)
	}
	return nil
}

func InsertIndexer(r record.Record) {
	rpcTags := indexer.RPCTags{
		Lid:  int32(r.LId),
		Seed: r.Seed,
		Tags: r.Tags,
	}
	_, err := indexerClient.InsertTags(context.Background(), &rpcTags)
	if err != nil {
		logrus.WithField("host", indexerHost).Error("failed to connect to indexer")
	}
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

func recordsArrival(records []record.Record) {
	// info.LogTimestamp("recordsArrival")
	for _, record := range records {
		err := Append(record)
		if err != nil {
			logrus.WithError(err).Error("couldn't append the record to the store")
		}
	}
}

func AssignToMaintainer(LId, numMaintainers int) int {
	return ((LId - 1) / batchSize) % numMaintainers
}
