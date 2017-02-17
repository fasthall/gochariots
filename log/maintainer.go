package log

import (
	"fmt"
	"net"
	"time"

	"github.com/fasthall/gochariots/info"
	"github.com/garyburd/redigo/redis"
)

var pool redis.Pool

func InitLogMaintainer() {
	fmt.Println("Trying to connect to redis...")
	pool = redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	fmt.Println("Connected to redis.")
}

// Info prints the information of Redis server
func Info() {
	conn := pool.Get()
	defer conn.Close()
	info, err := conn.Do("INFO")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", info)
}

// Append appends a new record to the log store.
func Append(record Record) (interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	b, _ := ToJSON(record)
	return conn.Do("SET", record.LId, b)
}

// ReadByLId reads from the log store according to LId.
func ReadByLId(LId int) (Record, error) {
	conn := pool.Get()
	defer conn.Close()
	b, _ := redis.Bytes(conn.Do("GET", LId))
	return ToRecord(b)
}

func recordsArrival(records []Record) {
	for _, record := range records {
		Append(record)
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
	fmt.Println(info.Name, "received:", records)
	conn.Close()
	recordsArrival(records)
}
