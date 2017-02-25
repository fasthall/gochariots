// Package batcher batches records sent from applications or other datacenters.
// When the buffer is full the records then are sent to corresponding filter.
package batcher

import (
	"fmt"
	"net"
	"time"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
)

const bufferSize int = 32

var buffer [][]log.Record
var filterHost []string
var numFilters int

// InitBatcher allocates n buffers, where n is the number of filters
func InitBatcher(n int) {
	numFilters = n
	buffer = make([][]log.Record, numFilters)
	for i := range buffer {
		buffer[i] = make([]log.Record, 0, bufferSize)
	}
	filterHost = make([]string, numFilters)
	fmt.Printf("%s is initialized with %d filter channels\n", info.Name, n)
}

// SetFilterHost sets the host of filter
func SetFilterHost(id int, host string) {
	filterHost[id] = host
}

// arrival buffers arriving records.
// Upon records arrive, depends on where the record origins it goes to a certain buffer.
// When a buffer is full, all the records in the buffer will be sent to the corresponding filter.
// BUG(fasthall) In Arrival(), the mechanism to match thre record and filter needs to be done. Currently the number of filters needs to be equal to datacenters.
func arrival(record log.Record) {
	dc := record.Host
	buffer[dc] = append(buffer[dc], record)

	// if the buffer is full, send all records to the filter
	if len(buffer[dc]) == cap(buffer[dc]) {
		sendToFilter(dc)
	}
}

func sendToFilter(dc int) {
	if len(buffer[dc]) == 0 {
		return
	}
	b := []byte{'r'}
	jsonBytes, err := log.ToJSONArray(buffer[dc])
	if err != nil {
		fmt.Println("Couldn't convert buffer to records")
	}
	buffer[dc] = buffer[dc][:0]
	conn, err := net.Dial("tcp", filterHost[dc])
	if err != nil {
		fmt.Printf("Couldn't connect to filterHost[%d] %s\n", dc, filterHost[dc])
	}
	defer conn.Close()
	conn.Write(append(b, jsonBytes...))
	fmt.Println(info.Name, "sent to filter", filterHost[dc])
}

func Sweeper() {
	for {
		time.Sleep(1 * time.Second)
		for i := range buffer {
			sendToFilter(i)
		}
	}
}

func HandleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	l, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Couldn't read buffer")
	}
	record, err := log.ToRecord(buf[:l])
	if err != nil {
		fmt.Println("Couldn't convert buffer to record")
	}
	fmt.Println(info.Name, "received:", record)
	arrival(record)
	conn.Close()
}
