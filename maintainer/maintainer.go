package maintainer

import (
	"hash/crc32"

	"github.com/fasthall/gochariots/maintainer/adapter/mongodb"
	"github.com/fasthall/gochariots/misc"
	"github.com/go-redis/redis"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

const batchSize int = 10000

var benchmark misc.Benchmark
var redisCacheHost []string
var redisCacheVer int
var redisClient []*redis.Client

type RecordHead struct {
	Depth uint32
	SeqID int64
}

type Server struct{}

func (s *Server) UpdateRedisCache(ctx context.Context, in *RPCRedisCache) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > redisCacheVer {
		redisCacheVer = ver
		redisCacheHost = in.GetRedisCache()
		redisClient = make([]*redis.Client, len(redisCacheHost))
		for i := range redisCacheHost {
			redisClient[i] = redis.NewClient(&redis.Options{
				Addr:     redisCacheHost[i],
				Password: "",
				DB:       0,
			})
		}
		logrus.WithField("host", in.GetRedisCache()).Info("received Redis hosts update")
	} else {
		logrus.WithFields(logrus.Fields{"current": redisCacheVer, "received": ver}).Debug("received older version of Redis hosts")
	}
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) ReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.Record, len(in.GetRecords()))
	propagated := []record.Record{}
	for i, ri := range in.GetRecords() {
		records[i] = record.Record{
			ID:        ri.GetId(),
			Timestamp: ri.GetTimestamp(),
			Host:      ri.GetHost(),
			SeqID:     ri.GetSeqid(),
			Depth:     ri.GetDepth(),
			Tags:      ri.GetTags(),
			Trace:     ri.GetTrace(),
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

// InitLogMaintainer initializes the maintainer and assign the path name to store the records
func InitLogMaintainer(benchmarkAccuracy int) {
	benchmark = misc.NewBenchmark(benchmarkAccuracy)
}

func NotifyRedisCache(head map[string]RecordHead) {
	pipes := make([]redis.Pipeliner, len(redisClient))
	for i := range redisClient {
		pipes[i] = redisClient[i].Pipeline()
	}
	for k, v := range head {
		key := k + string(v.Depth)
		i := crc32.ChecksumIEEE([]byte(key)) % uint32(len(redisClient))
		err := pipes[i].SetBit(key, v.SeqID, 1).Err()
		if err != nil {
			logrus.WithError(err).Error("couldn't add command to redis pipeline")
		}
	}
	for _, pipe := range pipes {
		_, err := pipe.Exec()
		if err != nil {
			logrus.WithError(err).Error("couldn't execute commands in redis pipeline")
		}
	}
}
