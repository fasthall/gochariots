package queue

import (
	"container/heap"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/misc/connection"
	"github.com/fasthall/gochariots/record"
)

var sameDCBuffered []record.TOIDRecord
var TOIDbuffered []BufferHeap

var Carry bool

// Token is used by queues to ensure causality of LId assignment
type TOIDToken struct {
	MaxTOId         []int
	LastLId         int
	DeferredRecords []record.TOIDRecord
}

// InitQueue initializes the buffer and hashmap for queued records
func TOIDInitQueue(hasToken, carry bool) {
	Carry = carry
	TOIDbuffered = make([]BufferHeap, info.NumDC)
	for i := range TOIDbuffered {
		TOIDbuffered[i] = BufferHeap{}
		heap.Init(&TOIDbuffered[i])
	}
	if hasToken {
		var token TOIDToken
		token.InitToken(make([]int, info.NumDC), 0)
		if Carry {
			TokenArrivalCarryDeferred(token)
		} else {
			TokenArrivalBufferDeferred(token)
		}
	}
	if hasToken {
		logrus.WithField("token", true).Info("initialized")
	} else {
		logrus.WithField("token", false).Info("initialized")
	}
}

// InitToken intializes a token. The IDs info should be accuired from log maintainers
func (token *TOIDToken) InitToken(maxTOId []int, lastLId int) {
	token.MaxTOId = maxTOId
	token.LastLId = lastLId
	token.DeferredRecords = []record.TOIDRecord{}
}

// recordsArrival deals with the records received from filters
func TOIDrecordsArrival(records []record.TOIDRecord) {
	// info.LogTimestamp("recordsArrival")
	bufMutex.Lock()
	for _, r := range records {
		if r.Host == info.ID {
			sameDCBuffered = append(sameDCBuffered, r)
		} else {
			heap.Push(&TOIDbuffered[r.Host], r)
		}
	}
	bufMutex.Unlock()
}

// TokenArrivalCarryDeferred function deals with token received.
// For each deferred records in the token, check if the current max TOId in shared log satisfies the dependency.
// If so, the deferred records are sent to the log maintainers.
func TokenArrivalCarryDeferred(token TOIDToken) {
	logrus.WithField("deferred", len(token.DeferredRecords)).Info("TokenArrivalCarryDeferred")
	bufMutex.Lock()
	// append buffered records to the token in order
	for host := range TOIDbuffered {
		for TOIDbuffered[host].Len() > 0 {
			r := heap.Pop(&TOIDbuffered[host]).(record.TOIDRecord)
			token.DeferredRecords = append(token.DeferredRecords, r)
		}
	}
	token.DeferredRecords = append(token.DeferredRecords, sameDCBuffered...)
	sameDCBuffered = []record.TOIDRecord{}
	bufMutex.Unlock()
	// put the deffered records with dependency satisfied into dispatch slice

	dispatch := []record.TOIDRecord{}
	head := 0
	for _, r := range token.DeferredRecords {
		if r.Host != info.ID {
			// default value of TOId is 0 so no need to check if the record has dependency or not
			if r.TOId == token.MaxTOId[r.Host]+1 && r.Pre.TOId <= token.MaxTOId[r.Pre.Host] {
				dispatch = append(dispatch, r)
				token.MaxTOId[r.Host] = r.TOId
			} else {
				token.DeferredRecords[head] = r
				head++
			}
		} else {
			// if it's from the same DC, TOId needs to be assigned
			if r.Pre.TOId <= token.MaxTOId[r.Pre.Host] {
				r.TOId = token.MaxTOId[r.Host] + 1
				dispatch = append(dispatch, r)
				token.MaxTOId[r.Host] = r.TOId
			} else {
				token.DeferredRecords[head] = r
				head++
			}
		}
	}
	token.DeferredRecords = token.DeferredRecords[:head]

	if len(dispatch) > 0 {
		// assign LId and send to log maintainers
		lastID := TOIDassignLId(dispatch, token.LastLId)
		token.LastLId = lastID
		toDispatch := make([][]record.TOIDRecord, len(logMaintainerConn))
		for _, r := range dispatch {
			id := maintainer.AssignToMaintainer(r.LId, len(logMaintainerConn))
			toDispatch[id] = append(toDispatch[id], r)
		}
		for id, t := range toDispatch {
			if len(t) > 0 {
				TOIDdispatchRecords(t, id)
			}
		}
	}
	go TOIDpassToken(&token)
}

// TokenArrivalBufferDeferred is similar to TokenArrivalCarryDeferred, except deferred records will be buffered rather than carried with token
func TokenArrivalBufferDeferred(token TOIDToken) {
	logrus.Info("TokenArrivalBufferDeferred")
	dispatch := []record.TOIDRecord{}
	bufMutex.Lock()
	for host := range TOIDbuffered {
		for TOIDbuffered[host].Len() > 0 {
			r := heap.Pop(&TOIDbuffered[host]).(record.TOIDRecord)
			if r.TOId == token.MaxTOId[r.Host]+1 && r.Pre.TOId <= token.MaxTOId[r.Pre.Host] {
				dispatch = append(dispatch, r)
				token.MaxTOId[r.Host] = r.TOId
			} else {
				heap.Push(&TOIDbuffered[host], r)
				break
			}
		}
	}
	head := 0
	for _, r := range sameDCBuffered {
		if r.Pre.TOId <= token.MaxTOId[r.Pre.Host] {
			r.TOId = token.MaxTOId[r.Host] + 1
			dispatch = append(dispatch, r)
			token.MaxTOId[r.Host] = r.TOId
		} else {
			sameDCBuffered[head] = r
			head++
		}
	}
	sameDCBuffered = sameDCBuffered[:head]
	bufMutex.Unlock()
	// put the deffered records with dependency satisfied into dispatch slice
	if len(dispatch) > 0 {
		// assign LId and send to log maintainers
		lastID := TOIDassignLId(dispatch, token.LastLId)
		token.LastLId = lastID
		toDispatch := make([][]record.TOIDRecord, len(logMaintainerConn))
		for _, r := range dispatch {
			id := maintainer.AssignToMaintainer(r.LId, len(logMaintainerConn))
			toDispatch[id] = append(toDispatch[id], r)
		}
		for id, t := range toDispatch {
			if len(t) > 0 {
				TOIDdispatchRecords(t, id)
			}
		}
	}
	go TOIDpassToken(&token)
}

// assignLId assigns LId to all the records to be sent to log maintainers
// return the last LId assigned
func TOIDassignLId(records []record.TOIDRecord, lastLId int) int {
	for i := range records {
		lastLId++
		records[i].LId = lastLId
	}
	return lastLId
}

// passToken sends the token to the next queue in the ring
func TOIDpassToken(token *TOIDToken) {
	time.Sleep(100 * time.Millisecond)
	if nextQueueHost == "" {
		if Carry {
			TokenArrivalCarryDeferred(*token)
		} else {
			TokenArrivalBufferDeferred(*token)
		}
	} else {
		bytes, err := json.Marshal(token)
		if err != nil {
			logrus.WithError(err).Warning("couldn't convert token to bytes")
			panic(err)
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
				logrus.WithField("host", nextQueueHost).Info("sent the token to the next queue")
			}
		}
	}
}

// dispatchRecords sends the ready records to log maintainers
func TOIDdispatchRecords(records []record.TOIDRecord, maintainerID int) {
	// info.LogTimestamp("dispatchRecords")
	bytes, err := record.TOIDToGobArray(records)
	if err != nil {
		logrus.WithField("records", records).Error("couldn't convert records to bytes")
		return
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))
	maintainerConnMutex.Lock()
	if logMaintainerConn[maintainerID] == nil {
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
		if logMaintainerConn[maintainerID] != nil {
			_, err = logMaintainerConn[maintainerID].Write(append(b, bytes...))
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
			logrus.WithFields(logrus.Fields{"length": len(records), "id": maintainerID}).Info("ent the records to maintainer")
		}
	}
	// log.Printf("TIMESTAMP %s:record in queue %s\n", info.GetName(), time.Since(lastTime))
}

// HandleRequest handles incoming connection
func TOIDHandleRequest(conn net.Conn) {
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
			records := []record.TOIDRecord{}
			err := record.TOIDGobToRecordArray(buf[1:totalLength], &records)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to record")
				continue
			}
			logrus.WithField("length", len(records)).Info("received incoming record")
			TOIDrecordsArrival(records)
		} else if buf[0] == 'q' { // received next host update
			nextQueueHost = string(buf[1:totalLength])
			if nextQueueConn != nil {
				nextQueueConn.Close()
				nextQueueConn = nil
			}
			logrus.WithField("host", nextQueueHost).Info("updates next queue")
		} else if buf[0] == 't' { // received token
			var token TOIDToken
			err := json.Unmarshal(buf[1:totalLength], &token)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to token")
				panic(err)
			}
			if Carry {
				TokenArrivalCarryDeferred(token)
			} else {
				TokenArrivalBufferDeferred(token)
			}
			logrus.Info("received token")
		} else if buf[0] == 'm' { // received maintainer update
			err := json.Unmarshal(buf[1:totalLength], &logMaintainerHost)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to maintainer hosts")
			}
			logMaintainerConn = make([]net.Conn, len(logMaintainerHost))
		} else if buf[0] == 'i' { // received indexer update
			err := json.Unmarshal(buf[1:totalLength], &indexerHost)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to indexer hosts")
			}
			indexerConn = make([]net.Conn, len(indexerHost))
		} else {
			logrus.WithField("header", buf[0]).Warning("couldn't understand request")
		}
	}
}
