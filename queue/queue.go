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
	"github.com/fasthall/gochariots/record"
)

const dispatchSize int = 32

var lastTime time.Time
var bufMutex sync.Mutex
var buffered []map[int]record.Record
var sameDCBuffered []record.Record
var connMutex sync.Mutex
var logMaintainerConn net.Conn
var logMaintainerHost string
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
	bufMutex.Unlock()
	token.DeferredRecords = append(token.DeferredRecords, sameDCBuffered...)
	sameDCBuffered = []record.Record{}

	// put the deffered records with dependency satisfied into dispatch slice
	dispatch := []record.Record{}
	head := 0
	for i := 0; i < len(token.DeferredRecords); i++ {
		v := token.DeferredRecords[i]
		if v.Host != info.ID {
			// default value of TOId is 0 so no need to check if the record has dependency or not
			if v.TOId == token.MaxTOId[v.Host]+1 && v.Pre.TOId <= token.MaxTOId[v.Pre.Host] {
				dispatch = append(dispatch, v)
				token.MaxTOId[v.Host] = v.TOId
			} else {
				token.DeferredRecords[head] = v
				head++
			}
		} else {
			// if it's from the same DC, TOId needs to be assigned
			if v.Pre.TOId <= token.MaxTOId[v.Pre.Host] {
				v.TOId = token.MaxTOId[v.Host] + 1
				dispatch = append(dispatch, v)
				token.MaxTOId[v.Host] = v.TOId
			} else {
				token.DeferredRecords[head] = v
				head++
			}
		}
	}
	token.DeferredRecords = token.DeferredRecords[:head]

	if len(dispatch) > 0 {
		// assign LId and send to log maintainers
		lastID := assignLId(dispatch, token.LastLId)
		token.LastLId = lastID
		for i := 0; i < len(dispatch); i += dispatchSize {
			if i+dispatchSize < len(dispatch) {
				dispatchRecords(dispatch[i : i+dispatchSize])
			} else {
				dispatchRecords(dispatch[i:])
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

func dialLogMaintainer() error {
	var err error
	logMaintainerConn, err = net.Dial("tcp", logMaintainerHost)
	return err
}

// dispatchRecords sends the ready records to log maintainers
func dispatchRecords(records []record.Record) {
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
	if logMaintainerConn == nil {
		err = dialLogMaintainer()
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
		if logMaintainerConn != nil {
			_, err = logMaintainerConn.Write(append(b, jsonBytes...))
		} else {
			err = errors.New("batcherConn[hostID] == nil")
		}
		connMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialLogMaintainer()
				if err != nil {
					log.Printf("%s couldn't connect to log maintainer %s\n", info.GetName(), logMaintainerHost)
				}
			} else {
				log.Printf("%s failed to connect to log maintainer %s after retrying 5 times\n", info.GetName(), logMaintainerHost)
				break
			}
		} else {
			sent = true
			log.Println(info.GetName(), "sent the records to", logMaintainerHost, string(jsonBytes))
		}
	}
	log.Printf("TIMESTAMP %s:record in queue %s\n", info.GetName(), time.Since(lastTime))
}

// HandleRequest handles incoming connection
func HandleRequest(conn net.Conn) {
	for {
		// Read the incoming connection into the buffer.
		lenbuf := make([]byte, 4)
		_, err := conn.Read(lenbuf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(info.GetName(), "couldn't read incoming request")
			log.Println(info.GetName(), err)
			break
		}
		totalLength := int(binary.BigEndian.Uint32(lenbuf))
		head := 0
		buf := make([]byte, totalLength)
		for totalLength > 0 {
			l, err := conn.Read(buf[head:])
			if err == io.EOF {
				break
			} else if err != nil {
				log.Println(info.GetName(), "couldn't read incoming request")
				log.Println(info.GetName(), err)
				break
			} else {
				totalLength -= l
			}
		}
		if totalLength > 0 {
			break
		}
		if buf[0] == 'r' { // received records
			info.LogTimestamp("HandleRequest")
			lastTime = time.Now()
			records, err := record.ToRecordArray(buf[1:])
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to records:", string(buf[1:]))
				continue
			}
			log.Println(info.GetName(), "received records:", records)
			recordsArrival(records)
		} else if buf[0] == 'q' { // received next host update
			nextQueueHost = string(buf[1:])
			if nextQueueConn != nil {
				nextQueueConn.Close()
				nextQueueConn = nil
			}
			log.Println(info.GetName(), "updates next queue host to:", nextQueueHost)
		} else if buf[0] == 't' { // received token
			var token Token
			err := json.Unmarshal(buf[1:], &token)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to token:", string(buf[1:]))
				log.Panicln(err)
			}
			TokenArrival(token)
			if len(token.DeferredRecords) > 0 {
				log.Println(info.GetName(), "received token:", token)
			}
		} else if buf[0] == 'm' { // received maintainer update
			var hosts []string
			err := json.Unmarshal(buf[1:], &hosts)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to maintainer hosts:", string(buf[1:]))
			}
			if len(hosts) > 0 {
				logMaintainerHost = hosts[0]
				if logMaintainerConn != nil {
					logMaintainerConn.Close()
					logMaintainerConn = nil
				}
			}
		}
	}
}
