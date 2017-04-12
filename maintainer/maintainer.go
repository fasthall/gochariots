package maintainer

import (
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
		buf := make([]byte, 2048)
		l, err := conn.Read(buf)
		// Read the incoming connection into the buffer.
		if err == io.EOF {
			return
		} else if err != nil {
			log.Println(info.GetName(), "couldn't read incoming buffer.")
			log.Println(info.GetName(), err)
			continue
		}
		if buf[0] == 'b' { // received remote batchers update
			var batchers []string
			err := json.Unmarshal(buf[1:l], &batchers)
			remoteBatchers = batchers
			remoteBatchersConn = make([]net.Conn, len(remoteBatchers))
			if err != nil {
				log.Println(info.GetName(), "couldn't convert received bytes to string list:", string(buf[1:l]))
				log.Panicln(err)
			}
			log.Println(info.GetName(), "received remote batchers update:", remoteBatchers)
		} else if buf[0] == 'r' {
			records, err := record.ToRecordArray(buf[1:l])
			if err != nil {
				panic(err)
			}
			log.Println(info.GetName(), "received records:", records)
			recordsArrival(records)
		}
	}
}
