// Package batcher batches records sent from applications or other datacenters.
// When the buffer is full the records then are sent to corresponding filter.
package batcher

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/misc/connection"
	"github.com/fasthall/gochariots/record"
)

var TOIDbuffer []record.TOIDRecord

// InitBatcher allocates n buffers, where n is the number of filters
func TOIDInitBatcher() {
	numFilters = info.NumDC
	TOIDbuffer = make([]record.TOIDRecord, 0, bufferSize)
}

// arrival buffers arriving records.
// Upon records arrive, depends on where the record origins it goes to a certain buffer.
// When a buffer is full, all the records in the buffer will be sent to the corresponding filter.
// BUG(fasthall) In Arrival(), the mechanism to match the record and filter needs to be done. Currently the number of filters needs to be equal to datacenters.
func TOIDarrival(r record.TOIDRecord) {
	// r.Timestamp = time.Now()
	bufMutex.Lock()
	TOIDbuffer = append(TOIDbuffer, r)

	// if the buffer is full, send all records to the filter
	if len(TOIDbuffer) == cap(TOIDbuffer) {
		TOIDsendToQueue()
	}
	bufMutex.Unlock()
}

func TOIDsendToQueue() {
	if len(TOIDbuffer) == 0 {
		return
	}
	// logrus.WithField("timestamp", time.Now()).Debug("sendToQueue")
	bytes, err := record.TOIDToGobArray(TOIDbuffer)
	if err != nil {
		panic(err)
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))
	queueID := rand.Intn(len(queuePool))

	TOIDbuffer = TOIDbuffer[:0]
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
			logrus.WithField("id", queueID).Debug("sent to queue")
		}
	}
}

// Sweeper periodcally sends the buffer content to filters
func TOIDSweeper() {
	for {
		time.Sleep(10 * time.Millisecond)
		bufMutex.Lock()
		TOIDsendToQueue()
		bufMutex.Unlock()
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
			// logrus.WithField("timestamp", time.Now()).Info("HandleRequest")
			// start := time.Now()
			var r record.TOIDRecord
			err := record.TOIDJSONToRecord(buf[1:totalLength], &r)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to record")
				continue
			}
			logrus.WithField("record", r).Info("received incoming record")
			TOIDarrival(r)
			// elapsed := time.Since(start)
			// log.Printf("TIMESTAMP %s:HandleRequest took %s\n", info.GetName(), elapsed)
		} else if buf[0] == 's' { // received records
			// logrus.WithField("timestamp", time.Now()).Info("HandleRequest")
			// start := time.Now()
			var records []record.TOIDRecord
			err := record.TOIDJSONToRecordArray(buf[1:totalLength], &records)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to record")
				continue
			}
			// log.Println(info.GetName(), "received incoming record:", string(buf[1:totalLength]))
			for _, r := range records {
				TOIDarrival(r)
			}
			// elapsed := time.Since(start)
			// log.Printf("TIMESTAMP %s:HandleRequest took %s\n", info.GetName(), elapsed)
		} else if buf[0] == 'q' { // received queue hosts
			ver := int(binary.BigEndian.Uint32(buf[1:5]))
			if ver > queuePoolVer {
				queuePoolVer = ver
				err := json.Unmarshal(buf[5:totalLength], &queuePool)
				if err != nil {
					logrus.WithField("buffer", string(buf[5:totalLength])).Error("couldn't convert read buffer to queue list")
					continue
				}
				queueConn = make([]net.Conn, len(queuePool))
				logrus.WithField("queues", queuePool).Info("received new queue update")
			} else {
				logrus.WithFields(logrus.Fields{"current": queuePoolVer, "received": ver}).Debug("receiver older version of queue list")
			}
		} else {
			logrus.WithField("header", buf[0]).Warning("couldn't understand request")
		}
	}
}
