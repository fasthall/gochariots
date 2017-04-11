package queue

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
)

var buffered []map[int]log.Record
var sameDCBuffered []log.Record
var logMaintainerConn net.Conn
var logMaintainerHost string
var nextQueueConn net.Conn
var nextQueueHost string
var mutex sync.Mutex

type queueHost int

// Token is used by queues to ensure causality of LId assignment
type Token struct {
	MaxTOId         []int
	LastLId         int
	DeferredRecords []log.Record
}

// InitQueue initializes the buffer and hashmap for queued records
func InitQueue(hasToken bool) {
	buffered = make([]map[int]log.Record, info.NumDC)
	for i := range buffered {
		buffered[i] = make(map[int]log.Record)
	}
	if hasToken {
		var token Token
		token.InitToken(make([]int, info.NumDC), 0)
		fmt.Println(token)
		TokenArrival(token)
	}
}

// InitToken intializes a token. The IDs info should be accuired from log maintainers
func (token *Token) InitToken(maxTOId []int, lastLId int) {
	token.MaxTOId = maxTOId
	token.LastLId = lastLId
}

// recordsArrival deals with the records received from filters
func recordsArrival(records []log.Record) {
	mutex.Lock()
	for _, record := range records {
		if record.Host == info.ID {
			sameDCBuffered = append(sameDCBuffered, record)
		} else {
			buffered[record.Host][record.TOId] = record
		}
	}
	mutex.Unlock()
}

// TokenArrival function deals with token received.
// For each deferred records in the token, check if the current max TOId in shared log satisfies the dependency.
// If so, the deferred records are sent to the log maintainers.
func TokenArrival(token Token) {
	// append buffered records to the token in order
	for dc := range buffered {
		keys := []int{}
		for k := range buffered[dc] {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			v := buffered[dc][k]
			token.DeferredRecords = append(token.DeferredRecords, v)
		}
		buffered[dc] = map[int]log.Record{}
	}
	token.DeferredRecords = append(token.DeferredRecords, sameDCBuffered...)
	sameDCBuffered = []log.Record{}

	// put the deffered records with dependency satisfied into dispatch slice
	dispatch := []log.Record{}
	head := 0
	for i := 0; i < len(token.DeferredRecords); i++ {
		v := token.DeferredRecords[i]
		if v.Host != info.ID {
			// default value of TOId is 0 so no need to check if the record has dependency or not
			if v.TOId == token.MaxTOId[v.Host]+1 && v.Pre.TOId <= token.MaxTOId[v.Pre.Host] {
				dispatch = append(dispatch, v)
				token.MaxTOId[v.Host] = v.TOId
			} else {
				token.DeferredRecords[head] = v
				head++
			}
		} else {
			// if it's from the same DC, TOId needs to be assigned
			if v.Pre.TOId <= token.MaxTOId[v.Pre.Host] {
				v.TOId = token.MaxTOId[v.Host] + 1
				dispatch = append(dispatch, v)
				token.MaxTOId[v.Host] = v.TOId
			} else {
				token.DeferredRecords[head] = v
				head++
			}
		}
	}
	token.DeferredRecords = token.DeferredRecords[:head]

	if len(dispatch) > 0 {
		// assign LId and send to log maintainers
		lastID := assignLId(dispatch, token.LastLId)
		token.LastLId = lastID
		dispatchRecords(dispatch)
	}
	go passToken(&token)
}

// assignLId assigns LId to all the records to be sent to log maintainers
// return the last LId assigned
func assignLId(records []log.Record, lastLId int) int {
	for i := range records {
		lastLId++
		records[i].LId = lastLId
	}
	return lastLId
}

func dialNextQueue() {
	var err error
	nextQueueConn, err = net.Dial("tcp", nextQueueHost)
	if err != nil {
		fmt.Println(info.GetName(), "couldn't connect to", nextQueueHost)
		panic(err)
	}
}

// passToken sends the token to the next queue in the ring
func passToken(token *Token) {
	time.Sleep(500 * time.Millisecond)
	if nextQueueHost == "" {
		TokenArrival(*token)
	} else {
		b := []byte{'t'}
		jsonBytes, err := json.Marshal(token)
		if err != nil {
			fmt.Println("Couldn't convert token to bytes")
			panic(err)
		}
		if nextQueueConn == nil {
			dialNextQueue()
		}
		sent := false
		for sent == false {
			_, err = nextQueueConn.Write(append(b, jsonBytes...))
			if err != nil {
				dialNextQueue()
			} else {
				sent = true
			}
		}
		if len(token.DeferredRecords) > 0 {
			fmt.Println(info.GetName(), "sent to", nextQueueHost)
			fmt.Println(string(jsonBytes))
		}
	}
}

func dialLogMaintainer() {
	var err error
	logMaintainerConn, err = net.Dial("tcp", logMaintainerHost)
	if err != nil {
		fmt.Println("Couldn't connect to", logMaintainerHost)
		return
	}
}

// dispatchRecords sends the ready records to log maintainers
func dispatchRecords(records []log.Record) {
	b := []byte{'r'}
	jsonBytes, err := log.ToJSONArray(records)
	if err != nil {
		panic(err)
	}
	if logMaintainerConn == nil {
		dialLogMaintainer()
	}
	sent := false
	for sent == false {
		_, err = logMaintainerConn.Write(append(b, jsonBytes...))
		if err != nil {
			dialLogMaintainer()
		} else {
			sent = true
		}
	}
	fmt.Println(info.GetName(), "sent to", logMaintainerHost)
	fmt.Println(string(jsonBytes))
}

func HandleRequest(conn net.Conn) {
	for {
		// Read the incoming connection into the buffer.
		buf := make([]byte, 2048)
		// buf, err := ioutil.ReadAll(conn)
		l, err := conn.Read(buf)
		if err == io.EOF {
			return
		} else if err != nil {
			fmt.Println("Error during reading buffer")
			panic(err)
		}
		if buf[0] == 'r' { // received records
			records, err := log.ToRecordArray(buf[1:l])
			if err != nil {
				fmt.Println("Couldn't convert received bytes to records:", string(buf))
				panic(err)
			}
			fmt.Println(info.GetName(), "received:", records)
			recordsArrival(records)
		} else if buf[0] == 'q' { // received next host update
			nextQueueHost = string(buf[1:l])
			if nextQueueConn != nil {
				nextQueueConn.Close()
				nextQueueConn = nil
			}
			fmt.Println(info.GetName(), "set next host:", nextQueueHost)
		} else if buf[0] == 't' { // received token
			var token Token
			err := json.Unmarshal(buf[1:l], &token)
			if err != nil {
				fmt.Println("Couldn't convert received bytes to token:", string(buf))
				panic(err)
			}
			TokenArrival(token)
			if len(token.DeferredRecords) > 0 {
				fmt.Println(info.GetName(), "received:", token)
			}
		} else if buf[0] == 'm' { // received maintainer update
			var hosts []string
			err := json.Unmarshal(buf[1:l], &hosts)
			if err != nil {
				fmt.Println("Couldn't convert received bytes to maintainer hosts")
				panic(err)
			}
			if len(hosts) > 0 {
				logMaintainerHost = hosts[0]
				if logMaintainerConn != nil {
					logMaintainerConn.Close()
					logMaintainerConn = nil
				}
			}
		}
	}
}
