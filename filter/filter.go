// Package filter ensures uniqueness of records.
// A record may arrive several times due to duplicated transmission or gossip from different datacenters.
// BUG(fasthall) needs a lock on nextTOId (or not, make filter single-threaded)
package filter

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
)

var queuePool []string
var nextTOId []int
var buffer []log.Record

// InitFilter Initializes all the expected TOId as 1
func InitFilter(n int) {
	nextTOId = make([]int, n)
	for i := range nextTOId {
		nextTOId[i] = 1
	}
	buffer = make([]log.Record, 0)
}

// AddQueue adds a queue to the host pool
func AddQueue(host string) {
	queuePool = append(queuePool, host)
}

// arrival deals with the records the filter received.
// If the TOId is the same as expected, the record will be forwared to the queue.
// If the TOId is larger than expected, the record will be buffered.
func arrival(records []log.Record) {
	queued := []log.Record{}
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
}

func sendToQueue(records []log.Record) {
	jsonBytes, err := log.ToJSONArray(records)
	if err != nil {
		panic(err)
	}
	host := queuePool[rand.Intn(len(queuePool))]
	conn, _ := net.Dial("tcp", host)
	defer conn.Close()
	conn.Write(jsonBytes)
	fmt.Println(info.Name, "sent to", host)
}

func HandleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	l, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	records, err := log.ToRecordArray(buf[:l])
	if err != nil {
		panic(err)
	}
	fmt.Println(info.Name, "received:", records)
	arrival(records)
	conn.Close()
}
