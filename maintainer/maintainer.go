package maintainer

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"sync"

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
		log.Println(info.GetName(), "couldn't access path", path)
		log.Panicln(err)
	}
	f, err = os.Create(filepath.Join(path, p))
	if err != nil {
		log.Println(info.GetName(), "couldn't create file", p)
		log.Panicln(err)
	}
}

func dialConn() error {
	var err error
	indexerConn, err = net.Dial("tcp", indexerHost)
	return err
}

// Append appends a new record to the maintainer.
func Append(r record.Record) error {
	// info.LogTimestamp("Append")
	r.Timestamp = time.Now().UnixNano()
	if LogRecordNth > 0 {
		if r.LId == 1 {
			logFirstTime = time.Now()
		} else if r.LId == LogRecordNth {
			log.Println("Appending", LogRecordNth, "records took", time.Since(logFirstTime))
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
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(r.Tags)
	if err != nil {
		log.Println(info.GetName(), "couldn't encode tags")
		return
	}
	b := make([]byte, 5)
	b[4] = byte('t')
	binary.BigEndian.PutUint32(b, uint32(len(lidBytes)+buf.Len()+1))

	indexerConnMutex.Lock()
	if indexerConn == nil {
		err := dialConn()
		if err != nil {
			log.Printf("%s couldn't connect to indexerHost %s\n", info.GetName(), indexerHost)
		} else {
			log.Printf("%s is connected to indexerHost %s\n", info.GetName(), indexerHost)
		}
	}
	indexerConnMutex.Unlock()
	cnt := 5
	sent := false
	for sent == false {
		var err error
		indexerConnMutex.Lock()
		if indexerConn != nil {
			_, err = indexerConn.Write(append(append(b, lidBytes...), buf.Bytes()...))
		} else {
			err = errors.New("indexerConn == nil")
		}
		indexerConnMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialConn()
				if err != nil {
					log.Printf("%s couldn't connect to indexerHost %s, retrying...\n", info.GetName(), indexerHost)
				}
			} else {
				log.Printf("%s failed to connect to indexerHost %s after retrying 5 times\n", info.GetName(), indexerHost)
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
			log.Println(info.GetName(), "couldn't append the record to the store")
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
			log.Println(info.GetName(), "couldn't read incoming request")
			log.Println(info.GetName(), err)
			break
		}
		if buf[0] == 'b' { // received remote batchers update
			var batchers []string
			err := json.Unmarshal(buf[1:totalLength], &batchers)
			remoteBatchers = batchers
			remoteBatchersConn = make([]net.Conn, len(remoteBatchers))
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to string list:", string(buf[1:totalLength]))
				log.Panicln(err)
			}
			log.Println(info.GetName(), "received remote batchers update:", remoteBatchers)
		} else if buf[0] == 'r' { // received records from queue
			// info.LogTimestamp("HandleRequest")
			records := []record.Record{}
			err := record.GobToRecordArray(buf[1:totalLength], &records)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to records:", string(buf[1:totalLength]))
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
					log.Println(info.GetName(), "couldn't convert record to bytes", r)
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
			log.Println(info.GetName(), "updates indexer host to:", indexerHost)
		} else {
			log.Println(info.GetName(), "couldn't understand", string(buf))
		}
	}
}

func AssignToMaintainer(LId, numMaintainers int) int {
	return ((LId - 1) / batchSize) % numMaintainers
}
