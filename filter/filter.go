// Package filter ensures uniqueness of records.
// A record may arrive several times due to duplicated transmission or gossip from different datacenters.
// BUG(fasthall) needs a lock on nextTOId (or not, make filter single-threaded)
package filter

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
)

var connMutex sync.Mutex
var queueConn []net.Conn
var queuePool []string
var nextTOId []int
var bufMutex sync.Mutex
var buffer []record.Record

// InitFilter Initializes all the expected TOId as 1
func InitFilter(n int) {
	nextTOId = make([]int, n)
	for i := range nextTOId {
		nextTOId[i] = 1
	}
	bufMutex.Lock()
	buffer = make([]record.Record, 0)
	bufMutex.Unlock()
}

// arrival deals with the records the filter received.
// If the TOId is the same as expected, the record will be forwared to the queue.
// If the TOId is larger than expected, the record will be buffered.
func arrival(records []record.Record) {
	// info.LogTimestamp("arrival")
	bufMutex.Lock()
	queued := []record.Record{}
	for _, record := range records {
		if record.Host == info.ID {
			// this record is from the same datacenter, the TOId hasn't been generated yet
			if record.TOId == 0 {
				queued = append(queued, record)
			}
		} else if record.TOId > nextTOId[record.Host] {
			buffer = append(buffer, record)
		} else if record.TOId == nextTOId[record.Host] {
			queued = append(queued, record)
			nextTOId[record.Host]++
			changed := true
			for changed {
				changed = false
				head := 0
				for _, v := range buffer {
					if v.Host == record.Host && v.TOId == nextTOId[record.Host] {
						changed = true
						queued = append(queued, v)
						nextTOId[record.Host]++
					} else {
						buffer[head] = v
						head++
					}
				}
				buffer = buffer[:head]
			}
		}
	}
	sendToQueue(queued)
	bufMutex.Unlock()
}

func dialConn(queueID int) error {
	host := queuePool[queueID]
	var err error
	queueConn[queueID], err = net.Dial("tcp", host)
	return err
}

func sendToQueue(records []record.Record) {
	// info.LogTimestamp("sendToQueue")
	jsonBytes, err := record.ToJSONArray(records)
	if err != nil {
		panic(err)
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
	queueID := rand.Intn(len(queuePool))
	connMutex.Lock()
	if queueConn[queueID] == nil {
		err = dialConn(queueID)
		if err != nil {
			log.Printf("%s couldn't connect to queuePool[%d] %s\n", info.GetName(), queueID, queuePool[queueID])
		} else {
			log.Printf("%s is connected to queuePool[%d] %s\n", info.GetName(), queueID, queuePool[queueID])
		}
	}
	connMutex.Unlock()

	cnt := 5
	sent := false
	for sent == false {
		var err error
		connMutex.Lock()
		if queueConn[queueID] != nil {
			_, err = queueConn[queueID].Write(append(b, jsonBytes...))
		} else {
			err = errors.New("batcherConn[hostID] == nil")
		}
		connMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialConn(queueID)
				if err != nil {
					log.Printf("%s couldn't connect to queuePool[%d] %s, retrying...\n", info.GetName(), queueID, queuePool[queueID])
				}
			} else {
				log.Printf("%s failed to connect to queuePool[%d] %s after retrying 5 times\n", info.GetName(), queueID, queuePool[queueID])
				break
			}
		} else {
			sent = true
			// log.Printf("%s sent to queuePool[%d] %s\n", info.GetName(), queueID, queuePool[queueID])
		}
	}
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
			// info.LogTimestamp("HandleRequest")
			// start := time.Now()
			records, err := record.ToRecordArray(buf[1:totalLength])
			if err != nil {
				log.Println(info.GetName(), "couldn't convert buffer to record:", string(buf[1:totalLength]))
				continue
			}
			// log.Println(info.GetName(), "received incoming records:", records)
			arrival(records)
			// log.Printf("TIMESTAMP %s:HandleRequest took %s\n", info.GetName(), time.Since(start))
		} else if buf[0] == 'q' { // received queue hosts
			queuePool = append(queuePool, string(buf[1:totalLength]))
			queueConn = make([]net.Conn, len(queuePool))
			log.Println(info.GetName(), "received new queue update:", string(buf[1:totalLength]))
		}
	}
}
