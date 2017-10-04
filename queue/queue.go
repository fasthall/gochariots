package queue

import (
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/maintainer/indexer"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

var lastTime time.Time
var bufMutex sync.Mutex
var buffered []record.Record
var maintainersClient []maintainer.MaintainerClient
var maintainersHost []string
var maintainersVer int
var indexerClient []indexer.IndexerClient
var indexerHost []string
var indexersVer int
var nextQueueClient QueueClient
var nextQueueHost string
var nextQueueVer int

var indexerBuf []byte

// Token is used by queues to ensure causality of LId assignment
type Token struct {
	LastLId uint32
}

type Server struct{}

func (s *Server) ReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.Record, len(in.GetRecords()))
	for i, ri := range in.GetRecords() {
		records[i] = record.Record{
			Timestamp: ri.GetTimestamp(),
			Host:      ri.GetHost(),
			LId:       ri.GetLid(),
			Tags:      ri.GetTags(),
			Hash:      ri.GetHash(),
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

func (s *Server) UpdateIndexers(ctx context.Context, in *RPCIndexers) (*RPCReply, error) {
	ver := int(in.GetVersion())
	if ver > indexersVer {
		indexersVer = ver
		indexerHost = in.GetIndexer()
		indexerClient = make([]indexer.IndexerClient, len(indexerHost))
		for i := range indexerHost {
			conn, err := grpc.Dial(indexerHost[i], grpc.WithInsecure())
			if err != nil {
				reply := RPCReply{
					Message: "couldn't connect to indexer",
				}
				return &reply, err
			}
			indexerClient[i] = indexer.NewIndexerClient(conn)
		}
		logrus.WithField("host", in.GetIndexer()).Info("received indexer hosts update")
	} else {
		logrus.WithFields(logrus.Fields{"current": indexersVer, "received": ver}).Debug("receiver older version of indexer hosts")
	}
	return &RPCReply{Message: "ok"}, nil
}

// InitQueue initializes the buffer and hashmap for queued records
func InitQueue(hasToken bool) {
	buffered = []record.Record{}
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
	indexerBuf = make([]byte, 1024*1024*32)
}

// InitToken intializes a token. The IDs info should be accuired from log maintainers
func (token *Token) InitToken(lastLId uint32) {
	token.LastLId = lastLId
}

// recordsArrival deals with the records received from filters
func recordsArrival(records []record.Record) {
	logrus.WithField("timestamp", time.Now()).Debug("recordsArrival")
	bufMutex.Lock()
	buffered = append(buffered, records...)
	bufMutex.Unlock()
}

// tokenArrival function deals with token received.
// For each deferred records in the token, check if the current max TOId in shared log satisfies the dependency.
// If so, the deferred records are sent to the log maintainers.
func tokenArrival(token Token) {
	dispatch := []record.Record{}
	query := []indexer.Query{}
	// append buffered records to the token in order
	bufMutex.Lock()
	head := 0
	for _, r := range buffered {
		if len(r.Hash) == 0 {
			dispatch = append(dispatch, r)
		} else {
			query = append(query, indexer.Query{
				Hash: r.Hash,
				Seed: r.Seed,
			})
			buffered[head] = r
			head++
		}
	}
	buffered = buffered[:head]

	// Ask indexer if the prerequisite records have been indexed already
	if len(query) > 0 {
		existed := make([]bool, len(query))
		for i := range indexerClient {
			logrus.WithField("query", query).Debug("queryIndexer")
			result, err := queryIndexer(query, i)
			if err != nil {
				logrus.WithField("id", i).Warning("indexer stops responding")
			} else {
				for j := range existed {
					existed[j] = existed[j] || result[j]
				}
			}
		}
		head = 0
		for i, r := range buffered {
			if existed[i] {
				dispatch = append(dispatch, r)
			} else {
				buffered[head] = r
				head++
			}
		}
		buffered = buffered[:head]
	}
	bufMutex.Unlock()
	// assign LId and send to log maintainers
	lastID := assignLId(dispatch, token.LastLId)
	token.LastLId = lastID
	toDispatch := make([][]record.Record, len(maintainersClient))
	for _, r := range dispatch {
		id := maintainer.AssignToMaintainer(r.LId, len(maintainersClient))
		toDispatch[id] = append(toDispatch[id], r)
	}
	for id, t := range toDispatch {
		if len(t) > 0 {
			dispatchRecords(t, id)
		}
	}
	go passToken(&token)
}

// assignLId assigns LId to all the records to be sent to log maintainers
// return the last LId assigned
func assignLId(records []record.Record, lastLId uint32) uint32 {
	for i := range records {
		lastLId++
		records[i].LId = lastLId
	}
	return lastLId
}

// passToken sends the token to the next queue in the ring
func passToken(token *Token) {
	time.Sleep(100 * time.Millisecond)
	if nextQueueHost == "" {
		tokenArrival(*token)
	} else {
		rpcToken := RPCToken{
			Lastlid: token.LastLId,
		}
		nextQueueClient.ReceiveToken(context.Background(), &rpcToken)
	}
}

// dispatchRecords sends the ready records to log maintainers
func dispatchRecords(records []record.Record, maintainerID int) {
	// info.LogTimestamp("dispatchRecords")
	rpcRecords := maintainer.RPCRecords{
		Records: make([]*maintainer.RPCRecord, len(records)),
	}
	for i, r := range records {
		tmp := maintainer.RPCRecord{
			Timestamp: r.Timestamp,
			Host:      r.Host,
			Lid:       r.LId,
			Tags:      r.Tags,
			Hash:      r.Hash,
			Seed:      r.Seed,
		}
		rpcRecords.Records[i] = &tmp
	}
	_, err := maintainersClient[maintainerID].ReceiveRecords(context.Background(), &rpcRecords)
	if err != nil {
		logrus.WithField("id", maintainerID).Error("failed to connect to maintainer")
	} else {
		logrus.WithFields(logrus.Fields{"records": records, "id": maintainerID}).Debug("sent the records to maintainer")
	}
	// log.Printf("TIMESTAMP %s:record in queue %s\n", info.GetName(), time.Since(lastTime))
}

func queryIndexer(query []indexer.Query, indexerID int) ([]bool, error) {
	rpcQueries := indexer.RPCQueries{
		Queries: make([]*indexer.RPCQuery, len(query)),
	}
	for i, q := range query {
		tmp := indexer.RPCQuery{
			Hash: q.Hash,
			Seed: q.Seed,
		}
		rpcQueries.Queries[i] = &tmp
	}
	reply, err := indexerClient[indexerID].Query(context.Background(), &rpcQueries)
	if err != nil {
		logrus.WithField("id", indexerID).Error("failed to connect to indexer after retrying 5 times")
		return nil, err
	}
	return reply.GetReply(), nil
}
