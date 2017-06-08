package queue

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/maintainer/indexer"
	"github.com/fasthall/gochariots/misc/connection"
	"github.com/fasthall/gochariots/record"
)

var lastTime time.Time
var bufMutex sync.Mutex
var buffered []record.Record
var maintainerConnMutex sync.Mutex
var logMaintainerConn []net.Conn
var logMaintainerHost []string
var indexerConnMutex sync.Mutex
var indexerConn []net.Conn
var indexerHost []string
var queueConnMutex sync.Mutex
var nextQueueConn net.Conn
var nextQueueHost string

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
		log.Println(info.GetName(), "initialized with token")
	} else {
		log.Println(info.GetName(), "initialized without token")
	}
	indexerBuf = make([]byte, 1024*1024*32)
}

// InitToken intializes a token. The IDs info should be accuired from log maintainers
func (token *Token) InitToken(lastLId int) {
	token.LastLId = lastLId
}

// recordsArrival deals with the records received from filters
func recordsArrival(records []record.Record) {
	// info.LogTimestamp("recordsArrival")
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
		if r.Pre.Hash == 0 {
			dispatch = append(dispatch, r)
		} else {
			query = append(query, indexer.Query{
				Hash: r.Pre.Hash,
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
			result, err := queryIndexer(query, i)
			if err != nil {
				log.Println("Indexer", i, "stops responding")
				log.Println(err)
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
			log.Println(info.GetName(), "couldn't convert token to bytes:", token)
			log.Panicln(err)
		}
		b := make([]byte, 5)
		b[4] = byte('t')
		binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))
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
				_, err = nextQueueConn.Write(append(b, bytes...))
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
				// if len(token.DeferredRecords) > 0 {
				// 	log.Println(info.GetName(), "sent the token to", nextQueueHost)
				// }
			}
		}
	}
}

func dialLogMaintainer(maintainerID int) error {
	var err error
	logMaintainerConn[maintainerID], err = net.Dial("tcp", logMaintainerHost[maintainerID])
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
		log.Println(info.GetName(), "couldn't convert records to bytes:", records)
		return
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))
	maintainerConnMutex.Lock()
	if logMaintainerConn[maintainerID] == nil {
		err = dialLogMaintainer(maintainerID)
		if err != nil {
			log.Printf("%s couldn't connect to log maintainer %s\n", info.GetName(), logMaintainerHost)
			log.Println(err)
		} else {
			log.Printf("%s is connected to log maintainer %s\n", info.GetName(), logMaintainerHost)
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
					log.Printf("%s couldn't connect to log maintainer %s\n", info.GetName(), logMaintainerHost)
				}
			} else {
				log.Printf("%s failed to connect to log maintainer %s after retrying 5 times\n", info.GetName(), logMaintainerHost)
				break
			}
		} else {
			sent = true
			// log.Println(info.GetName(), "sent the records to", logMaintainerHost[maintainerID], string(jsonBytes))
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
			log.Printf("%s couldn't connect to indexer %s\n", info.GetName(), indexerHost[indexerID])
			log.Println(err)
		} else {
			log.Printf("%s is connected to indexer %s\n", info.GetName(), indexerHost[indexerID])
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
					log.Printf("%s couldn't connect to indexer %s\n", info.GetName(), indexerHost[indexerID])
				}
			} else {
				log.Printf("%s failed to connect to indexer %s after retrying 5 times\n", info.GetName(), indexerHost[indexerID])
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
			log.Println(info.GetName(), "couldn't read incoming request")
			log.Println(info.GetName(), err)
			break
		}
		if buf[0] == 'r' { // received records
			// info.LogTimestamp("HandleRequest")
			lastTime = time.Now()
			records := []record.Record{}
			err := record.GobToRecordArray(buf[1:totalLength], &records)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to records:", string(buf[1:totalLength]))
				continue
			}
			// log.Println(info.GetName(), "received records:", records)
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
			tokenArrival(token)
			// if len(token.DeferredRecords) > 0 {
			// 	log.Println(info.GetName(), "received token:", token)
			// }
		} else if buf[0] == 'm' { // received maintainer update
			err := json.Unmarshal(buf[1:totalLength], &logMaintainerHost)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to maintainer hosts:", string(buf[1:totalLength]))
			}
			logMaintainerConn = make([]net.Conn, len(logMaintainerHost))
		} else if buf[0] == 'i' { // received indexer update
			err := json.Unmarshal(buf[1:totalLength], &indexerHost)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to indexer hosts:", string(buf[1:totalLength]))
			}
			indexerConn = make([]net.Conn, len(indexerHost))
		}
	}
}
