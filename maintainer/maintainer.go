package maintainer

import (
	"hash/crc32"
	"time"

	"github.com/fasthall/gochariots/cache"
	"github.com/fasthall/gochariots/maintainer/adapter/mongodb"
	"github.com/fasthall/gochariots/misc"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

const batchSize int = 10000

var benchmark misc.Benchmark
var cacheHost []string
var cacheVer int
var cacheClient []cache.CacheClient
var mongoHost []string
var mongoVer int
var mongoClient []mongodb.Client

type Server struct{}

func (s *Server) UpdateMongos(ctx context.Context, in *RPCMongos) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > mongoVer {
		mongoVer = ver
		mongoHost = in.GetHosts()
		mongoClient = make([]mongodb.Client, len(mongoHost))
		for i := range mongoHost {
			mongoClient[i] = mongodb.NewClient(mongoHost[i])
		}
		logrus.WithField("host", in.GetHosts()).Info("received mongoDB hosts update")
	} else {
		logrus.WithFields(logrus.Fields{"current": mongoVer, "received": ver}).Debug("received older version of mongoDB hosts")
	}
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) UpdateCaches(ctx context.Context, in *RPCCaches) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > cacheVer {
		cacheVer = ver
		cacheHost = in.GetHosts()
		cacheClient = make([]cache.CacheClient, len(cacheHost))
		for i := range cacheHost {
			conn, err := grpc.Dial(cacheHost[i], grpc.WithInsecure())
			if err != nil {
				reply := RPCReply{
					Message: "couldn't connect to cache",
				}
				return &reply, err
			}
			cacheClient[i] = cache.NewCacheClient(conn)
		}
		logrus.WithField("host", in.GetHosts()).Info("received Redis hosts update")
	} else {
		logrus.WithFields(logrus.Fields{"current": cacheVer, "received": ver}).Debug("received older version of cache hosts")
	}
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) ReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.Record, len(in.GetRecords()))
	propagated := []record.Record{}
	cacheUpdate := make([][]string, len(cacheClient))
	for i, ri := range in.GetRecords() {
		records[i] = record.Record{
			ID:        ri.GetId(),
			LID:       ri.GetLid(),
			Parent:    ri.GetParent(),
			Timestamp: ri.GetTimestamp(),
			Host:      ri.GetHost(),
			Tags:      ri.GetTags(),
			Trace:     ri.GetTrace(),
		}
		if ri.GetHost() == uint32(info.ID) {
			propagated = append(propagated, records[i])
		}
		// notify caches about these IDs
		hashed := crc32.ChecksumIEEE([]byte(ri.GetId())) % uint32(len(cacheClient))
		cacheUpdate[hashed] = append(cacheUpdate[hashed], ri.GetId())
	}
	go func() {
		// TODO mongodb sharding
		err := mongoClient[0].PutRecords(records)
		if err != nil {
			logrus.WithError(err).Error("couldn't put records to mongodb")
		} else {
			benchmark.Logging(len(records))
		}
		err = notifyCache(cacheUpdate)
		if err != nil {
			logrus.WithError(err).Error("failed to update redis cache")
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

// InitLogMaintainer initializes the maintainer and assign the path name to store the records
func InitLogMaintainer(benchmarkAccuracy int) {
	benchmark = misc.NewBenchmark(benchmarkAccuracy)
}

// Tell cache that these IDs are in the storage
func notifyCache(ids [][]string) error {
	start := time.Now()
	cnt := 0
	for i, partial := range ids {
		cnt += len(partial)
		rpcIDs := cache.RPCIDs{
			Ids: partial,
		}
		_, err := cacheClient[i].Put(context.Background(), &rpcIDs)
		if err != nil {
			logrus.WithError(err).Error("failed to update cache")
		}
	}

	logrus.Debug("Notify cache about ", cnt, "IDs took ", time.Since(start))
	return nil
}
