package maintainer

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
)

// LastLId records the last LID maintained by this maintainer
var LastLId int
var path = "flstore"

// InitLogMaintainer initializes the maintainer and assign the path name to store the records
func InitLogMaintainer(p string) {
	LastLId = 0
	path = p
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println(info.GetName(), "couldn't access path", path)
		panic(err)
	}
}

// Append appends a new record to the maintainer.
func Append(r record.Record) error {
	info.LogTimestamp("Append")
	b, err := record.ToJSON(r)
	if err != nil {
		return err
	}
	fpath := filepath.Join(path, strconv.Itoa(r.LId))
	err = ioutil.WriteFile(fpath, b, 0644)
	if err != nil {
		return err
	}
	log.Println(info.GetName(), "wrote record to", fpath)
	LastLId = r.LId
	if r.Host == info.ID {
		Propagate(r)
	}
	return nil
}

// ReadByLId reads from the maintainer according to LId.
func ReadByLId(LId int) (record.Record, error) {
	fpath := filepath.Join(path, strconv.Itoa(LId))
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return record.Record{}, err
	}
	return record.ToRecord(b)
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
	for {
		// Make a buffer to hold incoming data.
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
		l, err := conn.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(info.GetName(), "couldn't read incoming request")
			log.Println(info.GetName(), err)
			break
		}
		if buf[0] == 'b' { // received remote batchers update
			var batchers []string
			err := json.Unmarshal(buf[1:], &batchers)
			remoteBatchers = batchers
			remoteBatchersConn = make([]net.Conn, len(remoteBatchers))
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to string list:", string(buf[1:]))
				log.Panicln(err)
			}
			log.Println(info.GetName(), "received remote batchers update:", remoteBatchers)
		} else if buf[0] == 'r' {
			info.LogTimestamp("HandleRequest")
			records, err := record.ToRecordArray(buf[1:])
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to records:", string(buf[1:]))
				log.Panicln(binary.BigEndian.Uint32(lenbuf), len(buf), l, err)
			}
			log.Println(info.GetName(), "received records:", records)
			recordsArrival(records)
		}
	}
}
