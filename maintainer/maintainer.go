package maintainer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
)

var remoteBatchers []string
var LastLId int
var path = "flstore"

func InitLogMaintainer(p string) {
	LastLId = 0
	path = p
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		fmt.Println("Couldn't access path", path)
		panic(err)
	}
}

// Append appends a new record to the log store.
func Append(record log.Record) error {
	b, err := log.ToJSON(record)
	if err != nil {
		return err
	}
	fpath := filepath.Join(path, strconv.Itoa(record.LId))
	err = ioutil.WriteFile(fpath, b, 0644)
	if err != nil {
		return err
	}
	fmt.Println("Wrote to", fpath)
	LastLId = record.LId
	if record.Host == info.ID {
		Propagate(record)
	}
	return nil
}

// ReadByLId reads from the log store according to LId.
func ReadByLId(LId int) (log.Record, error) {
	fpath := filepath.Join(path, strconv.Itoa(LId))
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return log.Record{}, err
	}
	return log.ToRecord(b)
}

func recordsArrival(records []log.Record) {
	for _, record := range records {
		err := Append(record)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
}

func HandleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	l, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error during reading buffer")
		panic(err)
	}
	if buf[0] == 'b' { // received remote batchers update
		var batchers []string
		err := json.Unmarshal(buf[1:l], &batchers)
		remoteBatchers = batchers
		if err != nil {
			fmt.Println("Couldn't convert received bytes to string list")
			panic(err)
		}
		fmt.Println(info.GetName(), "received:", remoteBatchers)
	} else if buf[0] == 'r' {
		records, err := log.ToRecordArray(buf[1:l])
		if err != nil {
			panic(err)
		}
		fmt.Println(info.GetName(), "received:", records)
		recordsArrival(records)
	}
	conn.Close()
}
