// Package filter ensures uniqueness of records.
// A record may arrive several times due to duplicated transmission or gossip from different datacenters.
// BUG(fasthall) needs a lock on nextTOId (or not, make filter single-threaded)
package filter

import (
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/misc/connection"
	"github.com/fasthall/gochariots/record"
)

var TOIDbuffer []record.TOIDRecord

// InitFilter Initializes all the expected TOId as 1
func TOIDInitFilter(n int) {
	nextTOId = make([]int, n)
	for i := range nextTOId {
		nextTOId[i] = 1
	}
	bufMutex.Lock()
	TOIDbuffer = make([]record.TOIDRecord, 0)
	bufMutex.Unlock()
}

// arrival deals with the records the filter received.
// If the TOId is the same as expected, the record will be forwared to the queue.
// If the TOId is larger than expected, the record will be buffered.
func TOIDarrival(records []record.TOIDRecord) {
	// info.LogTimestamp("arrival")
	TOIDsendToQueue(records)
}

func TOIDsendToQueue(records []record.TOIDRecord) {
	logrus.WithField("timestamp", time.Now()).Info("sendToQueue")
	bytes, err := record.TOIDToGobArray(records)
	if err != nil {
		panic(err)
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))
	queueID := rand.Intn(len(queuePool))
	connMutex.Lock()
	if queueConn[queueID] == nil {
		err = dialConn(queueID)
		if err != nil {
			logrus.WithField("id", queueID).Error("couldn't connect to queue")
		} else {
			logrus.WithField("id", queueID).Info("connected to queue")
		}
	}
	connMutex.Unlock()

	cnt := 5
	sent := false
	for sent == false {
		var err error
		connMutex.Lock()
		if queueConn[queueID] != nil {
			_, err = queueConn[queueID].Write(append(b, bytes...))
		} else {
			err = errors.New("batcherConn[hostID] == nil")
		}
		connMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialConn(queueID)
				if err != nil {
					logrus.WithField("attempt", cnt).Warning("couldn't connect to queue, retrying...")
				} else {
					logrus.WithField("id", queueID).Info("connected to queue")
				}
			} else {
				logrus.WithField("id", queueID).Error("failed to connect to the queue after retrying 5 times")
				break
			}
		} else {
			sent = true
			logrus.WithField("id", queueID).Info("sent to queue")
		}
	}
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
			// start := time.Now()
			records := []record.TOIDRecord{}
			err := record.TOIDGobToRecordArray(buf[1:totalLength], &records)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to record")
				continue
			}
			logrus.WithField("records", records).Info("received incoming record")
			TOIDarrival(records)
			// log.Printf("TIMESTAMP %s:HandleRequest took %s\n", info.GetName(), time.Since(start))
		} else if buf[0] == 'q' { // received queue hosts
			queuePool = append(queuePool, string(buf[1:totalLength]))
			queueConn = make([]net.Conn, len(queuePool))
			logrus.WithField("queue", string(buf[1:totalLength])).Info("received new queue update")
		} else {
			logrus.WithField("header", buf[0]).Warning("couldn't understand request")
		}
	}
}
