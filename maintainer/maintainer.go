package maintainer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
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

// Append appends a new record to the maintainer.
func Append(r record.Record) error {
	_, err := record.ToJSON(r)
	if err != nil {
		return err
	}
	fpath := filepath.Join(path, strconv.Itoa(r.LId))
	// err = ioutil.WriteFile(fpath, b, 0644)
	// if err != nil {
	// 	return err
	// }
	fmt.Println("Wrote to", fpath)
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
			fmt.Println(err)
			panic(err)
		}
	}
}

func HandleRequest(conn net.Conn) {
	for {
		// Make a buffer to hold incoming data.
		buf := make([]byte, 2048)
		l, err := conn.Read(buf)
		// Read the incoming connection into the buffer.
		if err == io.EOF {
			return
		} else if err != nil {
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
			records, err := record.ToRecordArray(buf[1:l])
			if err != nil {
				panic(err)
			}
			fmt.Println(info.GetName(), "received:", records)
			recordsArrival(records)
		}
	}
}
