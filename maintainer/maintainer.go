package maintainer

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"fmt"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
)

const batchSize int = 1000

// LastLId records the last LID maintained by this maintainer
var LastLId int
var path = "flstore"
var f *os.File

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

// Append appends a new record to the maintainer.
func Append(r record.Record) error {
	info.LogTimestamp("Append")
	b, err := record.ToJSON(r)
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
	log.Println(info.GetName(), "wrote record ", lid)
	LastLId = r.LId
	if r.Host == info.ID {
		Propagate(r)
	}
	return nil
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
	err = record.ToRecord(buf, &r)
	return r, err
}

func recordsArrival(records []record.Record) {
	info.LogTimestamp("recordsArrival")
	for _, record := range records {
		err := Append(record)
		if err != nil {
			log.Println(info.GetName(), "couldn't append the record to the store")
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
		} else if buf[0] == 'r' {
			info.LogTimestamp("HandleRequest")
			records, err := record.ToRecordArray(buf[1:totalLength])
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to records:", string(buf[1:totalLength]))
				log.Panicln(binary.BigEndian.Uint32(lenbuf), len(buf), err)
			}
			log.Println(info.GetName(), "received records:", records)
			recordsArrival(records)
		} else if buf[0] == 'l' {
			lid := int(binary.BigEndian.Uint32(buf[1:totalLength]))
			r, err := ReadByLId(lid)
			if err != nil {
				conn.Write([]byte(fmt.Sprint(err)))
			} else {
				conn.Write([]byte(fmt.Sprint(r)))
			}
		} else {
			log.Println(info.GetName(), "couldn't understand", string(buf))
		}
	}
}

func AssignToMaintainer(LId, numMaintainers int) int {
	return ((LId - 1) / batchSize) % numMaintainers
}
