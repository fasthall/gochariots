package queue

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/record"
)

var lastTime time.Time
var bufMutex sync.Mutex
var buffered []map[int]record.Record
var sameDCBuffered []record.Record
var connMutex sync.Mutex
var logMaintainerConn []net.Conn
var logMaintainerHost []string
var nextQueueConn net.Conn
var nextQueueHost string

type queueHost int

// Token is used by queues to ensure causality of LId assignment
type Token struct {
	MaxTOId         []int
	LastLId         int
	DeferredRecords []record.Record
}

// InitQueue initializes the buffer and hashmap for queued records
func InitQueue(hasToken bool) {
	buffered = make([]map[int]record.Record, info.NumDC)
	for i := range buffered {
		buffered[i] = make(map[int]record.Record)
	}
	if hasToken {
		var token Token
		token.InitToken(make([]int, info.NumDC), 0)
		TokenArrival(token)
	}
	if hasToken {
		log.Println(info.GetName(), "initialized with token")
	} else {
		log.Println(info.GetName(), "initialized without token")
	}
}

// InitToken intializes a token. The IDs info should be accuired from log maintainers
func (token *Token) InitToken(maxTOId []int, lastLId int) {
	token.MaxTOId = maxTOId
	token.LastLId = lastLId
}

// recordsArrival deals with the records received from filters
func recordsArrival(records []record.Record) {
	info.LogTimestamp("recordsArrival")
	bufMutex.Lock()
	for _, record := range records {
		if record.Host == info.ID {
			sameDCBuffered = append(sameDCBuffered, record)
		} else {
			buffered[record.Host][record.TOId] = record
		}
	}
	bufMutex.Unlock()
}

// TokenArrival function deals with token received.
// For each deferred records in the token, check if the current max TOId in shared log satisfies the dependency.
// If so, the deferred records are sent to the log maintainers.
func TokenArrival(token Token) {
	bufMutex.Lock()
	// append buffered records to the token in order
	for dc := range buffered {
		keys := []int{}
		for k := range buffered[dc] {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			v := buffered[dc][k]
			token.DeferredRecords = append(token.DeferredRecords, v)
		}
		buffered[dc] = map[int]record.Record{}
	}
	token.DeferredRecords = append(token.DeferredRecords, sameDCBuffered...)
	sameDCBuffered = []record.Record{}
	bufMutex.Unlock()
	log.Println("WTF", token.DeferredRecords)
	// put the deffered records with dependency satisfied into dispatch slice
	dispatch := []record.Record{}
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
			if len(r.Pre.Tags) > 0 {
				lids := []int{}
				for maintainerID := range logMaintainerConn {
					lids = append(lids, AskIndexer(r.Pre.Tags, maintainerID)...)
				}
				log.Println(lids)
				if len(lids) > 0 {
					r.TOId = token.MaxTOId[r.Host] + 1
					dispatch = append(dispatch, r)
					token.MaxTOId[r.Host] = r.TOId
				} else {
					token.DeferredRecords[head] = r
					head++
				}
			} else if r.Pre.TOId <= token.MaxTOId[r.Pre.Host] {
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
		lastID := assignLId(dispatch, token.LastLId)
		token.LastLId = lastID
		toDispatch := make([][]record.Record, len(logMaintainerConn))
		for _, r := range dispatch {
			id := maintainer.AssignToMaintainer(r.LId, len(logMaintainerConn))
			toDispatch[id] = append(toDispatch[id], r)
		}
		for id, t := range toDispatch {
			if len(t) > 0 {
				dispatchRecords(t, id)
			}
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
	connMutex.Lock()
	var err error
	nextQueueConn, err = net.Dial("tcp", nextQueueHost)
	connMutex.Unlock()
	return err
}

// passToken sends the token to the next queue in the ring
func passToken(token *Token) {
	time.Sleep(100 * time.Millisecond)
	if nextQueueHost == "" {
		TokenArrival(*token)
	} else {
		jsonBytes, err := json.Marshal(token)
		if err != nil {
			log.Println(info.GetName(), "couldn't convert token to bytes:", token)
			log.Panicln(err)
		}
		b := make([]byte, 5)
		b[4] = byte('t')
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
		if nextQueueConn == nil {
			err = dialNextQueue()
			if err != nil {
				log.Printf("%s couldn't connect to the next queue %s\n", info.GetName(), nextQueueHost)
			} else {
				log.Printf("%s is connected to the next queue %s\n", info.GetName(), nextQueueHost)
			}
		}

		cnt := 5
		sent := false
		for sent == false {
			var err error
			if nextQueueConn != nil {
				_, err = nextQueueConn.Write(append(b, jsonBytes...))
			} else {
				err = errors.New("batcherConn[hostID] == nil")
			}
			if err != nil {
				if cnt >= 0 {
					cnt--
					err = dialNextQueue()
					if err != nil {
						log.Printf("%s couldn't connect to the next queue %s\n", info.GetName(), nextQueueHost)
					}
				} else {
					log.Printf("%s failed to connect to the next queue %s after retrying 5 times\n", info.GetName(), nextQueueHost)
					break
				}
			} else {
				sent = true
				if len(token.DeferredRecords) > 0 {
					log.Println(info.GetName(), "sent the token to", nextQueueHost)
				}
			}
		}
	}
}

func dialLogMaintainer(maintainerID int) error {
	var err error
	logMaintainerConn[maintainerID], err = net.Dial("tcp", logMaintainerHost[maintainerID])
	return err
}

// dispatchRecords sends the ready records to log maintainers
func dispatchRecords(records []record.Record, maintainerID int) {
	info.LogTimestamp("dispatchRecords")
	jsonBytes, err := record.ToJSONArray(records)
	if err != nil {
		log.Println(info.GetName(), "couldn't convert records to bytes:", records)
		return
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
	connMutex.Lock()
	if logMaintainerConn[maintainerID] == nil {
		err = dialLogMaintainer(maintainerID)
		if err != nil {
			log.Printf("%s couldn't connect to log maintainer %s\n", info.GetName(), logMaintainerHost)
			log.Println(err)
		} else {
			log.Printf("%s is connected to log maintainer %s\n", info.GetName(), logMaintainerHost)
		}
	}
	connMutex.Unlock()

	cnt := 5
	sent := false
	for sent == false {
		connMutex.Lock()
		if logMaintainerConn[maintainerID] != nil {
			_, err = logMaintainerConn[maintainerID].Write(append(b, jsonBytes...))
		} else {
			err = errors.New("logMaintainerConn[hostID] == nil")
		}
		connMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialLogMaintainer(maintainerID)
				if err != nil {
					log.Printf("%s couldn't connect to log maintainer %s\n", info.GetName(), logMaintainerHost)
				}
			} else {
				log.Printf("%s failed to connect to log maintainer %s after retrying 5 times\n", info.GetName(), logMaintainerHost)
				break
			}
		} else {
			sent = true
			log.Println(info.GetName(), "sent the records to", logMaintainerHost[maintainerID], string(jsonBytes))
		}
	}
	log.Printf("TIMESTAMP %s:record in queue %s\n", info.GetName(), time.Since(lastTime))
}

func AskIndexer(tags map[string]string, maintainerID int) []int {
	tmp, err := json.Marshal(tags)
	if err != nil {
		log.Println(info.GetName(), "couldn't convert tags to bytes:", tags)
		return nil
	}
	b := make([]byte, 5)
	b[4] = byte('i')
	binary.BigEndian.PutUint32(b, uint32(len(tmp)+1))

	connMutex.Lock()
	if logMaintainerConn[maintainerID] == nil {
		err := dialLogMaintainer(maintainerID)
		if err != nil {
			log.Printf("%s couldn't connect to log maintainer %s\n", info.GetName(), logMaintainerHost)
			log.Println(err)
		} else {
			log.Printf("%s is connected to log maintainer %s\n", info.GetName(), logMaintainerHost)
		}
	}
	connMutex.Unlock()

	cnt := 5
	sent := false
	for sent == false {
		connMutex.Lock()
		if logMaintainerConn[maintainerID] != nil {
			_, err = logMaintainerConn[maintainerID].Write(append(b, tmp...))
		} else {
			err = errors.New("logMaintainerConn[hostID] == nil")
		}
		connMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialLogMaintainer(maintainerID)
				if err != nil {
					log.Printf("%s couldn't connect to log maintainer %s\n", info.GetName(), logMaintainerHost)
				}
			} else {
				log.Printf("%s failed to connect to log maintainer %s after retrying 5 times\n", info.GetName(), logMaintainerHost)
				break
			}
		} else {
			sent = true
		}
	}
	lenbuf := make([]byte, 4)
	_, err = logMaintainerConn[maintainerID].Read(lenbuf)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		return nil
	}
	totalLength := int(binary.BigEndian.Uint32(lenbuf))
	buf := make([]byte, totalLength)
	_, err = logMaintainerConn[maintainerID].Read(buf)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		return nil
	}
	var lids []int
	err = json.Unmarshal(buf, &lids)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		return nil
	}
	return lids
}

// HandleRequest handles incoming connection
func HandleRequest(conn net.Conn) {
	lenbuf := make([]byte, 4)
	buf := make([]byte, 1024*1024*32)
	for {
		_, err := conn.Read(lenbuf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(info.GetName(), "couldn't read incoming request")
			log.Println(info.GetName(), err)
			break
		}
		totalLength := int(binary.BigEndian.Uint32(lenbuf))
		if totalLength > cap(buf) {
			buf = make([]byte, totalLength)
		}
		remain := totalLength
		head := 0
		for remain > 0 {
			l, err := conn.Read(buf[head : head+remain])
			if err == io.EOF {
				break
			} else if err != nil {
				log.Println(info.GetName(), "couldn't read incoming request")
				log.Println(info.GetName(), err)
				break
			} else {
				remain -= l
				head += l
			}
		}
		if remain != 0 {
			log.Println(info.GetName(), "couldn't read incoming request", remain)
			break
		}
		if buf[0] == 'r' { // received records
			info.LogTimestamp("HandleRequest")
			lastTime = time.Now()
			records, err := record.ToRecordArray(buf[1:totalLength])
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to records:", string(buf[1:totalLength]))
				continue
			}
			log.Println(info.GetName(), "received records:", records)
			recordsArrival(records)
		} else if buf[0] == 'q' { // received next host update
			nextQueueHost = string(buf[1:totalLength])
			if nextQueueConn != nil {
				nextQueueConn.Close()
				nextQueueConn = nil
			}
			log.Println(info.GetName(), "updates next queue host to:", nextQueueHost)
		} else if buf[0] == 't' { // received token
			var token Token
			err := json.Unmarshal(buf[1:totalLength], &token)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to token:", string(buf[1:totalLength]))
				log.Panicln(err)
			}
			TokenArrival(token)
			if len(token.DeferredRecords) > 0 {
				log.Println(info.GetName(), "received token:", token)
			}
		} else if buf[0] == 'm' { // received maintainer update
			err := json.Unmarshal(buf[1:totalLength], &logMaintainerHost)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to maintainer hosts:", string(buf[1:totalLength]))
			}
			logMaintainerConn = make([]net.Conn, len(logMaintainerHost))
		}
	}
}
