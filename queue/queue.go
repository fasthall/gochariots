package queue

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/maintainer/indexer"
	"github.com/fasthall/gochariots/misc/connection"
	"github.com/fasthall/gochariots/record"
)

var lastTime time.Time
var bufMutex sync.Mutex
var buffered []record.Record
var maintainerConnMutex sync.Mutex
var maintainersConn []net.Conn
var maintainersHost []string
var maintainersVer int
var indexerConnMutex sync.Mutex
var indexerConn []net.Conn
var indexerHost []string
var indexersVer int
var queueConnMutex sync.Mutex
var nextQueueConn net.Conn
var nextQueueHost string
var nextQueueVer int

var indexerBuf []byte

type queueHost int

// Token is used by queues to ensure causality of LId assignment
type Token struct {
	LastLId int
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
func (token *Token) InitToken(lastLId int) {
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
		if r.Hash == 0 {
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
		for i := range indexerConn {
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
	toDispatch := make([][]record.Record, len(maintainersConn))
	for _, r := range dispatch {
		id := maintainer.AssignToMaintainer(r.LId, len(maintainersConn))
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
func assignLId(records []record.Record, lastLId int) int {
	for i := range records {
		lastLId++
		records[i].LId = lastLId
	}
	return lastLId
}

func dialNextQueue() error {
	queueConnMutex.Lock()
	var err error
	nextQueueConn, err = net.Dial("tcp", nextQueueHost)
	queueConnMutex.Unlock()
	return err
}

// passToken sends the token to the next queue in the ring
func passToken(token *Token) {
	time.Sleep(100 * time.Millisecond)
	if nextQueueHost == "" {
		tokenArrival(*token)
	} else {
		bytes, err := json.Marshal(token)
		if err != nil {
			logrus.WithError(err).Warning("couldn't convert token to bytes")
		}
		b := make([]byte, 5)
		b[4] = byte('t')
		binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))
		if nextQueueConn == nil {
			err = dialNextQueue()
			if err != nil {
				logrus.WithField("host", nextQueueHost).Error("couldn't connect to the next queue")
			} else {
				logrus.WithField("host", nextQueueHost).Info("connected to the next queue")
			}
		}

		cnt := 5
		sent := false
		for sent == false {
			var err error
			if nextQueueConn != nil {
				_, err = nextQueueConn.Write(append(b, bytes...))
			} else {
				err = errors.New("batcherConn[hostID] == nil")
			}
			if err != nil {
				if cnt >= 0 {
					cnt--
					err = dialNextQueue()
					if err != nil {
						logrus.WithField("attempt", cnt).Warning("connected to the next queue, retrying")
					} else {
						logrus.WithField("host", nextQueueHost).Info("connected to the next queue")
					}
				} else {
					logrus.WithField("host", nextQueueHost).Error("failed to connect to the next queue after retrying 5 times")
					break
				}
			} else {
				sent = true
				// logrus.WithField("host", nextQueueHost).Debug("sent the token to the next queue")
			}
		}
	}
}

func dialLogMaintainer(maintainerID int) error {
	var err error
	maintainersConn[maintainerID], err = net.Dial("tcp", maintainersHost[maintainerID])
	return err
}

func dialIndexer(indexerID int) error {
	var err error
	indexerConn[indexerID], err = net.Dial("tcp", indexerHost[indexerID])
	return err
}

// dispatchRecords sends the ready records to log maintainers
func dispatchRecords(records []record.Record, maintainerID int) {
	// info.LogTimestamp("dispatchRecords")
	bytes, err := record.ToGobArray(records)
	if err != nil {
		logrus.WithField("records", records).Error("couldn't convert records to bytes")
		return
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))
	maintainerConnMutex.Lock()
	if maintainersConn[maintainerID] == nil {
		err = dialLogMaintainer(maintainerID)
		if err != nil {
			logrus.WithField("id", maintainerID).Error("couldn't connect to log maintainer")
		} else {
			logrus.WithField("id", maintainerID).Info("connected to log maintainer")
		}
	}
	maintainerConnMutex.Unlock()

	cnt := 5
	sent := false
	for sent == false {
		maintainerConnMutex.Lock()
		if maintainersConn[maintainerID] != nil {
			_, err = maintainersConn[maintainerID].Write(append(b, bytes...))
		} else {
			err = errors.New("logMaintainerConn[hostID] == nil")
		}
		maintainerConnMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialLogMaintainer(maintainerID)
				if err != nil {
					logrus.WithField("attempt", cnt).Warning("couldn't connect to log maintainer, retrying...")
				} else {
					logrus.WithField("id", maintainerID).Info("connected to log maintainer")
				}
			} else {
				logrus.WithField("id", maintainerID).Error("failed to connect to log maintainer after retrying 5 times")
				break
			}
		} else {
			sent = true
			logrus.WithFields(logrus.Fields{"records": records, "id": maintainerID}).Debug("sent the records to maintainer")
		}
	}
	// log.Printf("TIMESTAMP %s:record in queue %s\n", info.GetName(), time.Since(lastTime))
}

func queryIndexer(query []indexer.Query, indexerID int) ([]bool, error) {
	var tmp bytes.Buffer
	enc := gob.NewEncoder(&tmp)
	err := enc.Encode(query)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 5)
	b[4] = byte('q')
	binary.BigEndian.PutUint32(b, uint32(len(tmp.Bytes())+1))

	indexerConnMutex.Lock()
	if indexerConn[indexerID] == nil {
		err := dialIndexer(indexerID)
		if err != nil {
			logrus.WithFields(logrus.Fields{"id": indexerID, "host": indexerHost[indexerID]}).Error("couldn't connect to indexer")
		} else {
			logrus.WithFields(logrus.Fields{"id": indexerID, "host": indexerHost[indexerID]}).Info("connected to indexer")
		}
	}
	indexerConnMutex.Unlock()

	cnt := 5
	sent := false
	for sent == false {
		indexerConnMutex.Lock()
		if indexerConn[indexerID] != nil {
			_, err = indexerConn[indexerID].Write(append(b, tmp.Bytes()...))
		} else {
			err = errors.New("indexerConn[indexerID] == nil")
		}
		indexerConnMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialIndexer(indexerID)
				if err != nil {
					logrus.WithField("attempt", cnt).Warning("couldn't connect to indexer, retrying...")
				} else {
					logrus.WithField("id", indexerID).Info("connected to indexer")
				}
			} else {
				logrus.WithField("id", indexerID).Error("failed to connect to indexer after retrying 5 times")
				break
			}
		} else {
			sent = true
		}
	}

	buf := make([]byte, len(query)+64)
	totalLength, err := connection.Read(indexerConn[indexerID], &buf)
	result := []bool{}
	dec := gob.NewDecoder(bytes.NewBuffer(buf[:totalLength]))
	err = dec.Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// HandleRequest handles incoming connection
func HandleRequest(conn net.Conn) {
	buf := make([]byte, 1024*1024*32)
	for {
		totalLength, err := connection.Read(conn, &buf)
		if err == io.EOF {
			return
		} else if err != nil {
			logrus.WithError(err).Error("couldn't read incoming request")
			break
		}
		if buf[0] == 'r' { // received records
			// info.LogTimestamp("HandleRequest")
			lastTime = time.Now()
			records := []record.Record{}
			err := record.GobToRecordArray(buf[1:totalLength], &records)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to record")
				continue
			}
			logrus.WithField("records", records).Debug("received incoming record")

			recordsArrival(records)
		} else if buf[0] == 'q' { // received next host update
			ver := int(binary.BigEndian.Uint32(buf[1:5]))
			nextQueueHost = string(buf[5:totalLength])
			if ver > nextQueueVer {
				nextQueueVer = ver
				if nextQueueConn != nil {
					nextQueueConn.Close()
					nextQueueConn = nil
				}
				logrus.WithField("host", nextQueueHost).Info("received next host update")
			} else {
				logrus.WithFields(logrus.Fields{"current": nextQueueVer, "received": ver}).Debug("receiver older version of next queue host")
			}
		} else if buf[0] == 't' { // received token
			var token Token
			err := json.Unmarshal(buf[1:totalLength], &token)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to token")
			}
			tokenArrival(token)
			// logrus.Debug("received token")
		} else if buf[0] == 'm' { // received maintainer update
			ver := int(binary.BigEndian.Uint32(buf[1:5]))
			if ver > maintainersVer {
				maintainersVer = ver
				err := json.Unmarshal(buf[5:totalLength], &maintainersHost)
				if err != nil {
					logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to maintainer hosts")
				} else {
					maintainersConn = make([]net.Conn, len(maintainersHost))
					logrus.WithField("host", maintainersHost).Info("receiver maintainer hosts update")
				}
			} else {
				logrus.WithFields(logrus.Fields{"current": maintainersVer, "received": ver}).Debug("receiver older version of maintainer list")
			}
		} else if buf[0] == 'i' { // received indexer update
			ver := int(binary.BigEndian.Uint32(buf[1:5]))
			if ver > indexersVer {
				indexersVer = ver
				err := json.Unmarshal(buf[5:totalLength], &indexerHost)
				if err != nil {
					logrus.WithField("buffer", string(buf[5:totalLength])).Error("couldn't convert read buffer to indexer hosts")
				} else {
					indexerConn = make([]net.Conn, len(indexerHost))
					logrus.WithField("host", indexerHost).Info("receiver indexer hosts update")
				}
			} else {
				logrus.WithFields(logrus.Fields{"current": indexersVer, "received": ver}).Debug("receiver older version of indexer list")
			}
		} else {
			logrus.WithField("header", buf[0]).Warning("couldn't understand request")
		}
	}
}
