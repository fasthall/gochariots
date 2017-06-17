package maintainer

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/misc/connection"
	"github.com/fasthall/gochariots/record"
)

const batchSize int = 1000

// LastLId records the last LID maintained by this maintainer
var LastLId int
var path = "flstore"
var f *os.File
var indexerConnMutex sync.Mutex
var indexerConn net.Conn
var indexerHost string

var logFirstTime time.Time
var LogRecordNth int

// InitLogMaintainer initializes the maintainer and assign the path name to store the records
func InitLogMaintainer(p string) {
	LastLId = 0
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		logrus.WithField("path", path).Error("couldn't access path")
	}
	f, err = os.Create(filepath.Join(path, p))
	if err != nil {
		logrus.WithField("path", p).Error("couldn't create file")
		panic(err)
	}
}

func dialConn() error {
	var err error
	indexerConn, err = net.Dial("tcp", indexerHost)
	return err
}

// Append appends a new record to the maintainer.
func Append(r record.Record) error {
	logrus.WithField("timestamp", time.Now()).Info("Append")
	r.Timestamp = time.Now().UnixNano()
	if LogRecordNth > 0 {
		if r.LId == 1 {
			logFirstTime = time.Now()
		} else if r.LId == LogRecordNth {
			logrus.WithField("duration", time.Since(logFirstTime)).Info("appended", LogRecordNth, "records")
		}
	}
	b, err := record.ToJSON(r)
	if err != nil {
		return err
	}
	lenbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenbuf, uint32(len(b)))
	lid := r.LId
	_, err = f.WriteAt(append(lenbuf, b...), int64(512*lid))
	if err != nil {
		return err
	}
	// log.Println(info.GetName(), "wrote record ", lid)
	InsertIndexer(r)

	LastLId = r.LId
	if r.Host == info.ID {
		Propagate(r)
	}
	return nil
}

func InsertIndexer(r record.Record) {
	lidBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lidBytes, uint32(r.LId))
	seedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(seedBytes, r.Seed)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(r.Tags)
	if err != nil {
		logrus.WithField("tags", r.Tags).Error("couldn't encode tags")
		return
	}
	b := make([]byte, 5)
	b[4] = byte('t')
	binary.BigEndian.PutUint32(b, uint32(13+buf.Len()))

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
			_, err = indexerConn.Write(append(append(append(b, lidBytes...), seedBytes...), buf.Bytes()...))
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
func ReadByLId(lid int) (record.Record, error) {
	lenbuf := make([]byte, 4)
	_, err := f.ReadAt(lenbuf, int64(512*lid))
	if err != nil {
		return record.Record{}, err
	}
	length := int(binary.BigEndian.Uint32(lenbuf))
	buf := make([]byte, length)
	_, err = f.ReadAt(buf, int64(512*lid+4))
	if err != nil {
		return record.Record{}, err
	}
	var r record.Record
	err = record.JSONToRecord(buf, &r)
	return r, err
}

// ReadByLIds reads multiple records
func ReadByLIds(lids []int) ([]record.Record, error) {
	result := []record.Record{}
	for _, lid := range lids {
		r, err := ReadByLId(lid)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func recordsArrival(records []record.Record) {
	// info.LogTimestamp("recordsArrival")
	for _, record := range records {
		err := Append(record)
		if err != nil {
			logrus.WithError(err).Error("couldn't append the record to the store")
		}
	}
}

// HandleRequest handles incoming connection
func HandleRequest(conn net.Conn) {
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
			var batchers []string
			err := json.Unmarshal(buf[1:totalLength], &batchers)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to batcher list")
				panic(err)
			} else {
				remoteBatchers = batchers
				remoteBatchersConn = make([]net.Conn, len(remoteBatchers))
				logrus.WithField("batchers", remoteBatchers).Info("received remote batchers update")
			}
		} else if buf[0] == 'r' { // received records from queue
			// info.LogTimestamp("HandleRequest")
			records := []record.Record{}
			err := record.GobToRecordArray(buf[1:totalLength], &records)
			if err != nil {
				logrus.WithField("buffer", string(buf[1:totalLength])).Error("couldn't convert read buffer to records")
			}
			// log.Println(info.GetName(), "received records:", records)
			recordsArrival(records)
		} else if buf[0] == 'l' { // read records by LIds
			lid := int(binary.BigEndian.Uint32(buf[1:totalLength]))
			r, err := ReadByLId(lid)
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
			indexerHost = string(buf[1:totalLength])
			if indexerConn != nil {
				indexerConn.Close()
				indexerConn = nil
			}
			logrus.WithField("host", indexerHost).Info("updates indexer host")
		} else {
			logrus.WithField("header", buf[0]).Warning("couldn't understand request")
		}
	}
}

func AssignToMaintainer(LId, numMaintainers int) int {
	return ((LId - 1) / batchSize) % numMaintainers
}
