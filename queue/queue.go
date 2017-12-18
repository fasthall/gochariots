package queue

import (
	"math/rand"
	"sync"
	"time"

	"github.com/fasthall/gochariots/misc"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/maintainer/adapter/mongodb"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

var bufMutex sync.Mutex
var bufferedCausality []record.Causality // for two phase append
var bufferedRecord []record.Record
var maintainersClient []maintainer.MaintainerClient
var maintainersHost []string
var maintainersVer int
var nextQueueClient QueueClient
var nextQueueHost string
var nextQueueVer int
var twoPhase bool

var maintainerPacketSize = 1000
var querySizeLimit = 1000
var benchmark misc.Benchmark

// Token is used by queues to ensure causality of LId assignment
type Token struct {
	LastLId uint32
}

type Server struct{}

func (s *Server) ReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.Record, len(in.GetRecords()))
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
	}
	recordsArrival(records)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) ReceiveToken(ctx context.Context, in *RPCToken) (*RPCReply, error) {
	token := Token{
		LastLId: in.Lastlid,
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
func InitQueue(hasToken, twoPhaseAppend bool, querySize, benchmarkAccuracy int) {
	twoPhase = twoPhaseAppend
	querySizeLimit = querySize
	benchmark = misc.NewBenchmark(benchmarkAccuracy)
	if twoPhase {
		bufferedCausality = []record.Causality{}
	} else {
		bufferedRecord = []record.Record{}
	}
	if hasToken {
		var token Token
		token.InitToken(0)
		tokenArrival(token)
	}
	if hasToken {
		logrus.WithField("token", true).Info("initialized")
	} else {
		logrus.WithField("token", false).Info("initialized")
	}
}

// InitToken intializes a token. The IDs info should be accuired from log maintainers
func (token *Token) InitToken(lastLId uint32) {
	token.LastLId = lastLId
}

// recordsArrival deals with the records received from filters
func recordsArrival(records []record.Record) {
	logrus.WithField("timestamp", time.Now()).Debug("recordsArrival")
	if twoPhase {
		dispatchRecords(records)
		bufMutex.Lock()
		cs := make([]record.Causality, len(records))
		for i, r := range records {
			cs[i] = record.Causality{
				Id:     r.Id,
				Parent: r.Parent,
			}
		}
		bufferedCausality = append(bufferedCausality, cs...)
		bufMutex.Unlock()
	} else {
		bufMutex.Lock()
		bufferedRecord = append(bufferedRecord, records...)
		bufMutex.Unlock()
	}
}

// tokenArrival function deals with token received.
// For each deferred records in the token, check if the current max TOId in shared log satisfies the dependency.
// If so, the deferred records are sent to the log maintainers.
func tokenArrival(token Token) {
	lastLId := token.LastLId
	// two phase append
	if twoPhase {
		bufMutex.Lock()
		// build queries
		if len(bufferedCausality) > 0 {
			// ask MongoDB if the prerequisite records exist
			exist, nonexist, err := mongodb.ParellelQueryDBCausality(bufferedCausality, querySizeLimit)
			if err != nil {
				logrus.WithError(err).Error("couldn't connect to DB")
			}
			bufferedCausality = nonexist
			// update LId for those records with existing parent
			ids := []string{}
			lids := []uint32{}
			for _, c := range exist {
				lastLId++
				ids = append(ids, c.Id)
				lids = append(lids, lastLId)
			}
			if len(ids) > 0 {
				insertLIDs(ids, lids)
				benchmark.Logging(len(ids))
			}
		}
		bufMutex.Unlock()
	} else {
		// non two phase append
		bufMutex.Lock()
		// build queries
		if len(bufferedRecord) > 0 {
			exist, nonexist, err := mongodb.ParellelQueryDBRecord(bufferedRecord, querySizeLimit)
			if err != nil {
				logrus.WithError(err).Error("couldn't connect to DB")
			}
			bufferedRecord = nonexist
			// update LId for those records with existing parent
			if len(exist) > 0 {
				for i := range exist {
					lastLId++
					exist[i].LId = lastLId
				}
				dispatchRecords(exist)
				benchmark.Logging(len(exist))
			}
		}
		bufMutex.Unlock()
	}
	token.LastLId = lastLId
	go passToken(&token)
}

// passToken sends the token to the next queue in the ring
func passToken(token *Token) {
	// time.Sleep(10 * time.Millisecond)
	if nextQueueHost == "" || nextQueueClient == nil {
		tokenArrival(*token)
	} else {
		rpcToken := RPCToken{
			Lastlid: token.LastLId,
		}
		nextQueueClient.ReceiveToken(context.Background(), &rpcToken)
	}
}

// dispatchRecords sends the ready records to log maintainers
func insertLIDs(ids []string, lids []uint32) {
	if len(ids) != len(lids) {
		logrus.Error("length doesn't match when inserting LIDs")
		return
	}
	for i := 0; i < len(ids); i += maintainerPacketSize {
		end := i + maintainerPacketSize
		if len(ids) < end {
			end = len(ids)
		}
		go mongodb.InsertLIds(ids[i:end], lids[i:end])
	}
}

// dispatchRecords sends the ready records to log maintainers
func dispatchRecords(records []record.Record) {
	for i := 0; i < len(records); i += maintainerPacketSize {
		end := i + maintainerPacketSize
		if len(records) < end {
			end = len(records)
		}
		go sendToMaintainer(records[i:end], rand.Intn(len(maintainersClient)))
	}
}

func sendToMaintainer(records []record.Record, maintainerID int) {
	// info.LogTimestamp("dispatchRecords")
	rpcRecords := maintainer.RPCRecords{
		Records: make([]*maintainer.RPCRecord, len(records)),
	}
	for i, r := range records {
		tmp := maintainer.RPCRecord{
			Id:        r.Id,
			Timestamp: r.Timestamp,
			Host:      r.Host,
			Lid:       r.LId,
			Tags:      r.Tags,
			Parent:    r.Parent,
			Seed:      r.Seed,
		}
		rpcRecords.Records[i] = &tmp
	}
	_, err := maintainersClient[maintainerID].ReceiveRecords(context.Background(), &rpcRecords)
	if err != nil {
		logrus.WithError(err).Error("failed to connect to maintainer", len(records))
	} else {
		logrus.WithFields(logrus.Fields{"records": len(records), "id": maintainerID}).Debug("sent the records to maintainer")
	}
}
