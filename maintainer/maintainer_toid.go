package maintainer

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/misc/connection"
	"github.com/fasthall/gochariots/record"
)

// Append appends a new record to the maintainer.
func TOIDAppend(r record.TOIDRecord) error {
	// info.LogTimestamp("Append")
	r.Timestamp = time.Now().UnixNano()
	if logRecordNth > 0 {
		if r.LId == 1 {
			logFirstTime = time.Now()
		} else if r.LId == logRecordNth {
			logrus.WithField("duration", time.Since(logFirstTime)).Info("appended", logRecordNth, "records")
		}
	}
	b, err := record.TOIDToJSON(r)
	if err != nil {
		return err
	}
	lenbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenbuf, uint32(len(b)))
	lid := r.LId
	_, err = f.WriteAt(lenbuf, int64(512*lid))
	if err != nil {
		return err
	}
	_, err = f.WriteAt(b, int64(512*lid+4))
	if err != nil {
		return err
	}
	// log.Println(info.GetName(), "wrote record ", lid)
	TOIDInsertIndexer(r)

	LastLId = r.LId
	if r.Host == info.ID {
		TOIDPropagate(r)
	}
	return nil
}

func TOIDInsertIndexer(r record.TOIDRecord) {
	tmp := make([]byte, 20)
	binary.BigEndian.PutUint64(tmp, r.ID)
	binary.BigEndian.PutUint32(tmp[8:], uint32(r.LId))
	binary.BigEndian.PutUint32(tmp[12:], uint32(r.TOId))
	binary.BigEndian.PutUint32(tmp[16:], uint32(r.Host))
	b := make([]byte, 5)
	b[4] = byte('t')
	binary.BigEndian.PutUint32(b, uint32(21))

	indexerConnMutex.Lock()
	if indexerConn == nil {
		err := dialConn()
		if err != nil {
			logrus.WithField("host", indexerHost).Error("couldn't connect to indexer")
		} else {
			logrus.WithField("host", indexerHost).Info("connected to indexer")
		}
	}
	indexerConnMutex.Unlock()
	cnt := 5
	sent := false
	for sent == false {
		var err error
		indexerConnMutex.Lock()
		if indexerConn != nil {
			_, err = indexerConn.Write(append(b, tmp...))
		} else {
			err = errors.New("indexerConn == nil")
		}
		indexerConnMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialConn()
				if err != nil {
					logrus.WithField("attempt", cnt).Warning("couldn't connect to indexer, retrying...")
				} else {
					logrus.WithField("host", indexerHost).Info("connected to indexer")
				}
			} else {
				logrus.WithField("host", indexerHost).Error("failed to connect to indexer after retrying 5 times")
				break
			}
		} else {
			sent = true
		}
	}
}

// ReadByLId reads from the maintainer according to LId.
func TOIDReadByLId(lid int) (record.TOIDRecord, error) {
	lenbuf := make([]byte, 4)
	_, err := f.ReadAt(lenbuf, int64(512*lid))
	if err != nil {
		return record.TOIDRecord{}, err
	}
	length := int(binary.BigEndian.Uint32(lenbuf))
	buf := make([]byte, length)
	_, err = f.ReadAt(buf, int64(512*lid+4))
	if err != nil {
		return record.TOIDRecord{}, err
	}
	var r record.TOIDRecord
	err = record.TOIDJSONToRecord(buf, &r)
	return r, err
}

// ReadByLIds reads multiple records
func TOIDReadByLIds(lids []int) ([]record.TOIDRecord, error) {
	result := []record.TOIDRecord{}
	for _, lid := range lids {
		r, err := TOIDReadByLId(lid)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func TOIDrecordsArrival(records []record.TOIDRecord) {
	// info.LogTimestamp("recordsArrival")
	for _, record := range records {
		err := TOIDAppend(record)
		if err != nil {
			logrus.WithError(err).Error("couldn't append the record to the store")
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
		if buf[0] == 'b' { // received remote batchers update
			ver := int(binary.BigEndian.Uint32(buf[1:5]))
			if ver > remoteBatcherVer {
				remoteBatcherVer = ver
				err := json.Unmarshal(buf[5:totalLength], &remoteBatchers)
				if err != nil {
					logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to batcher list")
					panic(err)
				} else {
					remoteBatchersConn = make([]net.Conn, len(remoteBatchers))
					logrus.WithField("batchers", remoteBatchers).Info("received remote batchers update")
				}
			} else {
				logrus.WithFields(logrus.Fields{"current": remoteBatcherVer, "received": ver}).Debug("receiver older version of remote batcher")
			}
		} else if buf[0] == 'r' { // received records from queue
			// info.LogTimestamp("HandleRequest")
			records := []record.TOIDRecord{}
			err := record.TOIDGobToRecordArray(buf[1:totalLength], &records)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to records")
			}
			// log.Println(info.GetName(), "received records:", records)
			TOIDrecordsArrival(records)
		} else if buf[0] == 'l' { // read records by LIds
			lid := int(binary.BigEndian.Uint32(buf[1:totalLength]))
			r, err := TOIDReadByLId(lid)
			if err != nil {
				conn.Write([]byte(fmt.Sprint(err)))
			} else {
				b, err := json.Marshal(r)
				if err != nil {
					logrus.WithField("record", r).Error("couldn't convert record to bytes")
					conn.Write([]byte(fmt.Sprint(err)))
				} else {
					conn.Write(b)
				}
			}
		} else if buf[0] == 'i' { // received indexer update
			ver := int(binary.BigEndian.Uint32(buf[1:5]))
			if ver > indexerVer {
				indexerVer = ver
				indexerHost = string(buf[5:totalLength])
				if indexerConn != nil {
					indexerConn.Close()
					indexerConn = nil
				}
				logrus.WithField("host", indexerHost).Info("updates indexer host")
			} else {
				logrus.WithFields(logrus.Fields{"current": indexerVer, "received": ver}).Debug("receiver older version of indexer host")
			}
		} else {
			logrus.WithField("header", buf[0]).Warning("couldn't understand request")
		}
	}
}
