package queue

import (
	"encoding/binary"
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
	LastLId         int
	DeferredRecords map[uint64]record.Record // the key of the map is the casual dependency of the record
}

// InitQueue initializes the buffer and hashmap for queued records
func InitQueue(hasToken bool) {
	buffered = []record.Record{}
	if hasToken {
		var token Token
		token.InitToken(0)
		TokenArrival(token)
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
	token.DeferredRecords = make(map[uint64]record.Record)
}

// recordsArrival deals with the records received from filters
func recordsArrival(records []record.Record) {
	// info.LogTimestamp("recordsArrival")
	bufMutex.Lock()
	buffered = append(buffered, records...)
	bufMutex.Unlock()
}

// TokenArrival function deals with token received.
// For each deferred records in the token, check if the current max TOId in shared log satisfies the dependency.
// If so, the deferred records are sent to the log maintainers.
func TokenArrival(token Token) {
	dispatch := []record.Record{}
	bufMutex.Lock()
	// append buffered records to the token in order
	for _, r := range buffered {
		if r.Pre.Hash == 0 {
			dispatch = append(dispatch, r)
		} else {
			token.DeferredRecords[r.Pre.Hash] = r
		}
	}
	buffered = []record.Record{}
	bufMutex.Unlock()
	// put the deffered records with dependency satisfied into dispatch slice

	if len(dispatch) > 0 {
		checkAgain := true
		for checkAgain {
			checkAgain = false
			// check if to-be-dispatched record satisify other deferred records
			for _, d := range dispatch {
				for key, value := range d.Tags {
					if deferred, ok := token.DeferredRecords[indexer.TagToHash(key, value)]; ok {
						dispatch = append(dispatch, deferred)
						delete(token.DeferredRecords, indexer.TagToHash(key, value))
						checkAgain = true
					}
				}
			}
		}
	}
	if len(token.DeferredRecords) > 0 {
		indexerQuery := []uint64{}
		for hash := range token.DeferredRecords {
			indexerQuery = append(indexerQuery, hash)
		}
		existed := make([]bool, len(indexerQuery))
		for i := range indexerConn {
			result := AskIndexerByHashes(indexerQuery, i)
			for j := 0; j < len(existed); j++ {
				existed[j] = existed[j] || result[j]
			}
		}
		for i, e := range existed {
			if e {
				dispatch = append(dispatch, token.DeferredRecords[indexerQuery[i]])
				delete(token.DeferredRecords, indexerQuery[i])
			}
		}
	}

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
		TokenArrival(*token)
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

func AskIndexerByTags(tags map[string]string, indexerID int) []int {
	tmp, err := json.Marshal(tags)
	if err != nil {
		log.Println(info.GetName(), "couldn't convert tags to bytes:", tags)
		return nil
	}
	b := make([]byte, 5)
	b[4] = byte('i')
	binary.BigEndian.PutUint32(b, uint32(len(tmp)+1))

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
			_, err = indexerConn[indexerID].Write(append(b, tmp...))
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

	lenbuf := make([]byte, 4)
	_, err = indexerConn[indexerID].Read(lenbuf)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		log.Panicln(err)
		return nil
	}
	totalLength := int(binary.BigEndian.Uint32(lenbuf))
	buf := make([]byte, totalLength)
	_, err = indexerConn[indexerID].Read(buf)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		log.Panicln(err)
		return nil
	}
	var lids []int
	err = json.Unmarshal(buf, &lids)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		log.Panicln(err)
		return nil
	}
	return lids
}

func AskIndexerByHash(hash uint64, indexerID int) []int {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, hash)
	b := make([]byte, 5)
	b[4] = byte('h')
	binary.BigEndian.PutUint32(b, uint32(len(tmp)+1))

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
	var err error
	for sent == false {
		indexerConnMutex.Lock()
		if indexerConn[indexerID] != nil {
			_, err = indexerConn[indexerID].Write(append(b, tmp...))
		} else {
			err = errors.New("indexerConn[hostID] == nil")
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
	lenbuf := make([]byte, 4)
	_, err = indexerConn[indexerID].Read(lenbuf)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		log.Panicln(err)
		return nil
	}
	totalLength := int(binary.BigEndian.Uint32(lenbuf))
	buf := make([]byte, totalLength)
	_, err = indexerConn[indexerID].Read(buf)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		log.Panicln(err)
		return nil
	}
	var lids []int
	err = json.Unmarshal(buf, &lids)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		log.Panicln(err)
		return nil
	}
	return lids
}

func AskIndexerByHashes(hash []uint64, indexerID int) []bool {
	bytes, err := json.Marshal(hash)
	if err != nil {
		log.Println(info.GetName(), "couldn't convert hashes into bytes", hash)
		log.Panicln(err)
	}
	b := make([]byte, 5)
	b[4] = byte('H')
	binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))

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
			_, err = indexerConn[indexerID].Write(append(b, bytes...))
		} else {
			err = errors.New("indexerConn[hostID] == nil")
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
	totalLength, err := connection.Read(indexerConn[indexerID], &indexerBuf)
	if err != nil {
		log.Println(info.GetName(), "couldn't read response from indexer")
		log.Panicln(err)
	}
	var result []bool
	err = json.Unmarshal(indexerBuf[:totalLength], &result)
	if err != nil {
		log.Println(info.GetName(), "couldn't unmarshal response from indexer")
		log.Panicln(err)
		return nil
	}
	return result
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
			TokenArrival(token)
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
