// Package batcher batches records sent from applications or other datacenters.
// When the buffer is full the records then are sent to corresponding filter.
package batcher

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
	"github.com/fasthall/gochariots/record"
)

const bufferSize int = 256

var bufMutex sync.Mutex
var buffer [][]record.Record
var connMutex sync.Mutex
var filterConn []net.Conn
var filterHost []string
var numFilters int

// InitBatcher allocates n buffers, where n is the number of filters
func InitBatcher(n int) {
	numFilters = n
	buffer = make([][]record.Record, numFilters)
	for i := range buffer {
		buffer[i] = make([]record.Record, 0, bufferSize)
	}
	filterHost = make([]string, numFilters)
	filterConn = make([]net.Conn, numFilters)
	log.Printf("%s is initialized with %d filter channels\n", info.GetName(), n)
}

// arrival buffers arriving records.
// Upon records arrive, depends on where the record origins it goes to a certain buffer.
// When a buffer is full, all the records in the buffer will be sent to the corresponding filter.
// BUG(fasthall) In Arrival(), the mechanism to match thre record and filter needs to be done. Currently the number of filters needs to be equal to datacenters.
func arrival(record record.Record) {
	bufMutex.Lock()
	dc := record.Host
	buffer[dc] = append(buffer[dc], record)

	// if the buffer is full, send all records to the filter
	if len(buffer[dc]) == cap(buffer[dc]) {
		sendToFilter(dc)
	}
	bufMutex.Unlock()
}

func dialConn(dc int) error {
	var err error
	filterConn[dc], err = net.Dial("tcp", filterHost[dc])
	return err
}

func sendToFilter(dc int) {
	if len(buffer[dc]) == 0 {
		return
	}
	info.LogTimestamp("sendToFilter")

	bufMutex.Lock()
	jsonBytes, err := record.ToJSONArray(buffer[dc])
	if err != nil {
		log.Panicln(info.GetName(), "couldn't convert buffer to records")
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))

	buffer[dc] = buffer[dc][:0]
	connMutex.Lock()
	if filterConn[dc] == nil {
		err = dialConn(dc)
		if err != nil {
			log.Printf("%s couldn't connect to filterHost[%d] %s\n", info.GetName(), dc, filterHost[dc])
		} else {
			log.Printf("%s is connected to filterHost[%d] %s\n", info.GetName(), dc, filterHost[dc])
		}
	}
	connMutex.Unlock()
	cnt := 5
	sent := false
	for sent == false {
		var err error
		connMutex.Lock()
		if filterConn[dc] != nil {
			_, err = filterConn[dc].Write(append(b, jsonBytes...))
		} else {
			err = errors.New("batcherConn[hostID] == nil")
		}
		connMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialConn(dc)
				if err != nil {
					log.Printf("%s couldn't connect to filterHost[%d] %s, retrying...\n", info.GetName(), dc, filterHost[dc])
				}
			} else {
				log.Printf("%s failed to connect to filterHost[%d] %s after retrying 5 times\n", info.GetName(), dc, filterHost[dc])
				break
			}
		} else {
			sent = true
			log.Printf("%s sent to filterHost[%d] %s\n", info.GetName(), dc, filterHost[dc])
		}
	}

	bufMutex.Unlock()
}

// Sweeper periodcally sends the buffer content to filters
func Sweeper() {
	for {
		time.Sleep(1 * time.Millisecond)
		for i := range buffer {
			sendToFilter(i)
		}
	}
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
		buf := make([]byte, binary.BigEndian.Uint32(lenbuf))
		_, err = conn.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(info.GetName(), "couldn't read incoming request")
			log.Println(info.GetName(), err)
			break
		}
		if buf[0] == 'r' { // received records
			info.LogTimestamp("HandleRequest")
			start := time.Now()
			record, err := record.ToRecord(buf[1:])
			if err != nil {
				log.Println(info.GetName(), "couldn't convert read buffer to record:", string(buf[1:]))
				continue
			}
			log.Println(info.GetName(), "received incoming record:", string(buf[1:]))
			arrival(record)
			elapsed := time.Since(start)
			log.Printf("TIMESTAMP %s:HandleRequest took %s\n", info.GetName(), elapsed)
		} else if buf[0] == 'f' { //received filter update
			err := json.Unmarshal(buf[1:], &filterHost)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert bytes to filter list:", string(buf[1:]))
				continue
			} else {
				log.Println(info.GetName(), "updates filter:", filterHost)
			}
		}
	}
}
