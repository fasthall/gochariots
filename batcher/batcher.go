// Package batcher batches records sent from applications or other datacenters.
// When the buffer is full the records then are sent to corresponding filter.
package batcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
)

const bufferSize int = 32

var mutex sync.Mutex
var buffer [][]record.Record
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
	fmt.Printf("%s is initialized with %d filter channels\n", info.GetName(), n)
}

// arrival buffers arriving records.
// Upon records arrive, depends on where the record origins it goes to a certain buffer.
// When a buffer is full, all the records in the buffer will be sent to the corresponding filter.
// BUG(fasthall) In Arrival(), the mechanism to match thre record and filter needs to be done. Currently the number of filters needs to be equal to datacenters.
func arrival(record record.Record) {
	mutex.Lock()
	dc := record.Host
	buffer[dc] = append(buffer[dc], record)

	// if the buffer is full, send all records to the filter
	if len(buffer[dc]) == cap(buffer[dc]) {
		sendToFilter(dc)
	}
	mutex.Unlock()
}

func dialConn(dc int) {
	var err error
	filterConn[dc], err = net.Dial("tcp", filterHost[dc])
	if err != nil {
		fmt.Printf("Couldn't connect to filterHost[%d] %s\n", dc, filterHost[dc])
		fmt.Println(err)
	} else {
		fmt.Printf("%s connected to filterHost[%d] %s\n", info.GetName(), dc, filterHost[dc])
	}
}

func sendToFilter(dc int) {
	if len(buffer[dc]) == 0 {
		return
	}
	mutex.Lock()
	b := []byte{'r'}
	jsonBytes, err := record.ToJSONArray(buffer[dc])
	if err != nil {
		fmt.Println("Couldn't convert buffer to records")
	}
	buffer[dc] = buffer[dc][:0]
	if filterConn[dc] == nil {
		dialConn(dc)
	}
	sent := false
	for sent == false {
		_, err := filterConn[dc].Write(append(b, jsonBytes...))
		if err != nil {
			dialConn(dc)
		} else {
			sent = true
		}
	}
	fmt.Println(info.GetName(), "sent to filter", filterHost[dc])
	mutex.Unlock()
}

func Sweeper() {
	for {
		time.Sleep(100 * time.Millisecond)
		for i := range buffer {
			sendToFilter(i)
		}
	}
}

func HandleRequest(conn net.Conn) {
	for {
		// Read the incoming connection into the buffer.
		buf := make([]byte, 2048)
		l, err := conn.Read(buf)
		if err == io.EOF {
			return
		} else if err != nil {
			fmt.Println("Error during reading buffer")
			panic(err)
		}
		if buf[0] == 'r' { // received records
			record, err := record.ToRecord(buf[1:l])
			if err != nil {
				fmt.Println("Couldn't convert buffer to record")
				panic(err)
			}
			fmt.Println(info.GetName(), "received:", record)
			arrival(record)
		} else if buf[0] == 'f' { //received filter update
			err := json.Unmarshal(buf[1:l], &filterHost)
			if err != nil {
				fmt.Println("Couldn't convert bytes to filter list")
				fmt.Println("Bytes:", string(buf[1:l]))
				panic(err)
			}
			fmt.Println(info.GetName(), "update filter:", filterHost)
		}
	}
}
