package log

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fasthall/gochariots/info"
)

var LastLId int
var path string = "flstore"

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
func Append(record Record) error {
	b, err := ToJSON(record)
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
	return nil
}

// ReadByLId reads from the log store according to LId.
func ReadByLId(LId int) (Record, error) {
	fpath := filepath.Join(path, strconv.Itoa(LId))
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return Record{}, err
	}
	return ToRecord(b)
}

func recordsArrival(records []Record) {
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
		panic(err)
	}
	records, err := ToRecordArray(buf[:l])
	if err != nil {
		panic(err)
	}
	fmt.Println(info.GetName(), "received:", records)
	conn.Close()
	recordsArrival(records)
}
