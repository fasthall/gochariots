package queue

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/fasthall/gochariots/maintainer/adapter/mongodb"
	"github.com/fasthall/gochariots/misc"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/cache"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

var bufferedMutex sync.Mutex
var bufferedRecord []record.Record
var readyMutex sync.Mutex
var readyRecord []record.Record
var maintainersClient []maintainer.MaintainerClient
var maintainersHost []string
var maintainersVer int
var nextQueueClient QueueClient
var nextQueueHost string
var nextQueueVer int
var cacheHost []string
var cacheVer int
var cacheClient []cache.CacheClient
var mongoHost []string
var mongoVer int
var mongoClient []mongodb.Client

var maintainerPacketSize = 1000
var querySizeLimit = 1000
var benchmark misc.Benchmark

// Token is used by queues to ensure causality of LId assignment
type Token struct {
	LastLID uint32
}

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
	for i, ri := range in.GetRecords() {
		records[i] = record.Record{
			ID:        ri.GetId(),
			Parent:    ri.GetParent(),
			Timestamp: ri.GetTimestamp(),
			Host:      ri.GetHost(),
			Tags:      ri.GetTags(),
			Trace:     ri.GetTrace(),
		}
	}
	recordsArrival(records)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) ReceiveToken(ctx context.Context, in *RPCToken) (*RPCReply, error) {
	token := Token{
		LastLID: in.Lastlid,
	}
	tokenArrival(token)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) UpdateNextQueue(ctx context.Context, in *RPCQueue) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > nextQueueVer {
		nextQueueVer = ver
		nextQueueHost = in.GetQueue()
		conn, err := grpc.Dial(nextQueueHost, grpc.WithInsecure())
		if err != nil {
			reply := RPCReply{
				Message: "couldn't connect to next queue",
			}
			return &reply, err
		}
		nextQueueClient = NewQueueClient(conn)
		logrus.WithField("host", nextQueueHost).Info("received next host update")
	} else {
		logrus.WithFields(logrus.Fields{"current": nextQueueVer, "received": ver}).Debug("receiver older version of next queue host")
	}
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) UpdateMaintainers(ctx context.Context, in *RPCMaintainers) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > maintainersVer {
		maintainersVer = ver
		maintainersHost = in.GetMaintainer()
		maintainersClient = make([]maintainer.MaintainerClient, len(maintainersHost))
		for i := range maintainersHost {
			conn, err := grpc.Dial(maintainersHost[i], grpc.WithInsecure())
			if err != nil {
				reply := RPCReply{
					Message: "couldn't connect to maintainer",
				}
				return &reply, err
			}
			maintainersClient[i] = maintainer.NewMaintainerClient(conn)
		}
		logrus.WithField("host", in.GetMaintainer()).Info("received maintainer hosts update")
	} else {
		logrus.WithFields(logrus.Fields{"current": maintainersVer, "received": ver}).Debug("receiver older version of maintainer hosts")
	}
	return &RPCReply{Message: "ok"}, nil
}

// InitQueue initializes the buffer and hashmap for queued records
func InitQueue(hasToken bool, querySize, benchmarkAccuracy int) {
	querySizeLimit = querySize
	benchmark = misc.NewBenchmark(benchmarkAccuracy)
	bufferedRecord = []record.Record{}
	readyRecord = []record.Record{}
	if hasToken {
		var token Token
		token.LastLID = 0
		tokenArrival(token)
	}
	if hasToken {
		logrus.WithField("token", true).Info("initialized")
	} else {
		logrus.WithField("token", false).Info("initialized")
	}
	go queryLoop()
}

func queryLoop() {
	for {
		bufferedMutex.Lock()
		// ask cache
		exists, nonexists, err := queryCaches(bufferedRecord)
		if err != nil {
			logrus.WithError(err).Error("couldn't connect to DB")
		}
		bufferedRecord = nonexists
		bufferedMutex.Unlock()

		readyMutex.Lock()
		readyRecord = append(readyRecord, exists...)
		readyMutex.Unlock()
	}
}

// recordsArrival deals with the records received from filters
func recordsArrival(records []record.Record) {
	logrus.WithField("timestamp", time.Now()).Debug("recordsArrival")
	bufferedMutex.Lock()
	bufferedRecord = append(bufferedRecord, records...)
	bufferedMutex.Unlock()
}

// tokenArrival function deals with token received.
// For each deferred records in the token, check if the current max TOId in shared log satisfies the dependency.
// If so, the deferred records are sent to the log maintainers.
func tokenArrival(token Token) {
	lastLID := token.LastLID
	readyMutex.Lock()
	if len(readyRecord) > 0 {
		for i := range readyRecord {
			lastLID++
			readyRecord[i].LID = lastLID
			readyRecord[i].Timestamp = time.Now().Unix()
		}
		readyRecord = dispatchRecords(readyRecord)
	}
	readyMutex.Unlock()
	token.LastLID = lastLID
	go passToken(&token)
}

// passToken sends the token to the next queue in the ring
func passToken(token *Token) {
	// time.Sleep(10 * time.Millisecond)
	if nextQueueHost == "" || nextQueueClient == nil {
		tokenArrival(*token)
	} else {
		rpcToken := RPCToken{
			Lastlid: token.LastLID,
		}
		nextQueueClient.ReceiveToken(context.Background(), &rpcToken)
	}
}

// dispatchRecords sends the ready records to log maintainers
func dispatchRecords(records []record.Record) []record.Record {
	for i := 0; i < len(records); i += maintainerPacketSize {
		end := i + maintainerPacketSize
		if len(records) < end {
			end = len(records)
		}
		go sendToMaintainer(records[i:end], rand.Intn(len(maintainersClient)))
	}
	return []record.Record{}
}

func sendToMaintainer(records []record.Record, maintainerID int) {
	// info.LogTimestamp("dispatchRecords")
	rpcRecords := maintainer.RPCRecords{
		Records: make([]*maintainer.RPCRecord, len(records)),
	}
	for i, r := range records {
		tmp := maintainer.RPCRecord{
			Id:        r.ID,
			Lid:       r.LID,
			Parent:    r.Parent,
			Timestamp: r.Timestamp,
			Host:      r.Host,
			Tags:      r.Tags,
			Trace:     r.Trace,
		}
		rpcRecords.Records[i] = &tmp
	}
	_, err := maintainersClient[maintainerID].ReceiveRecords(context.Background(), &rpcRecords)
	if err != nil {
		logrus.WithError(err).Error("failed to connect to maintainer ", len(records))
	} else {
		benchmark.Logging(len(records))
		logrus.WithFields(logrus.Fields{"records": len(records), "id": maintainerID}).Debug("sent the records to maintainer")
	}
}

func queryCache(records []record.Record, cacheID int) ([]bool, error) {
	rpcIDs := cache.RPCIDs{}
	for _, record := range records {
		rpcIDs.Ids = append(rpcIDs.Ids, record.Parent)
	}
	result, err := cacheClient[cacheID].Get(context.Background(), &rpcIDs)
	if err != nil {
		return nil, err
	}
	return result.Exists, nil
}

func queryCaches(records []record.Record) ([]record.Record, []record.Record, error) {
	if len(records) == 0 {
		return []record.Record{}, []record.Record{}, nil
	}

	start := time.Now()

	exists := []record.Record{}
	nonexists := []record.Record{}
	wg := sync.WaitGroup{}
	wg.Add(int(math.Ceil(float64(len(records)) / float64(querySizeLimit))))
	mutex := sync.Mutex{}

	for i := 0; i < len(records); i += querySizeLimit {
		li := i
		ri := i + querySizeLimit
		if len(records) < ri {
			ri = len(records)
		}
		go func(li, ri int) error {
			partialExists, partialNonexists, err := queryPartialCaches(records[li:ri])
			if err != nil {
				logrus.WithError(err).Error("couldn't query partial cache, len=", ri-li)
				wg.Done()
				return err
			}
			mutex.Lock()
			exists = append(exists, partialExists...)
			nonexists = append(nonexists, partialNonexists...)
			mutex.Unlock()
			wg.Done()
			return nil
		}(li, ri)
	}
	wg.Wait()

	logrus.Debug("Query ", len(records), " records took ", time.Since(start))
	return exists, nonexists, nil
}

func queryPartialCaches(records []record.Record) ([]record.Record, []record.Record, error) {
	exists := []record.Record{}
	nonexists := []record.Record{}

	poolSize := len(cacheClient)
	partialRecords := make([][]record.Record, poolSize)
	results := make([][]bool, poolSize)
	for _, record := range records {
		if record.Parent == "" {
			exists = append(exists, record)
		} else {
			clientID := misc.HashID(record.Parent, poolSize)
			partialRecords[clientID] = append(partialRecords[clientID], record)
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		j := i
		go func() {
			var err error
			results[j], err = queryCache(partialRecords[j], j)
			if err != nil {
				logrus.WithError(err).Error("couldn't query partial cache, len=", len(partialRecords[j]))
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for i := 0; i < poolSize; i++ {
		for j := range results[i] {
			if results[i][j] {
				exists = append(exists, partialRecords[i][j])
			} else {
				nonexists = append(nonexists, partialRecords[i][j])
			}
		}
	}
	return exists, nonexists, nil
}
