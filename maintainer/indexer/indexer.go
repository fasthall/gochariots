package indexer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net"
	"strconv"

	"encoding/gob"

	"github.com/fasthall/gochariots/info"
)

var indexes = make(map[uint64][]int)
var Subscriber net.Conn

func InitIndexer(p string) {

}

func toHash(b []byte) uint64 {
	hash := fnv.New64a()
	hash.Write(b)
	return hash.Sum64()
}

func Insert(key, value string, LId int) {
	h := toHash([]byte(key + ":" + value))
	indexes[h] = append(indexes[h], LId)
	notify(h, LId)
}

func GetByTag(key, value string) []int {
	h := toHash([]byte(key + ":" + value))
	return indexes[h]
}

func GetByTags(tags map[string]string) []int {
	result := map[int]int{}
	for key, value := range tags {
		tmp := GetByTag(key, value)
		for _, tmpv := range tmp {
			result[tmpv]++
		}
	}
	ans := []int{}
	for k, v := range result {
		if v == len(tags) {
			ans = append(ans, k)
		}
	}
	return ans
}

func notify(hash uint64, LId int) {
	if Subscriber != nil {
		payload := map[string]string{"hash": fmt.Sprint(hash), "LId": strconv.Itoa(LId)}
		bytes, err := json.Marshal(payload)
		if err != nil {
			log.Println(err)
		}
		_, err = Subscriber.Write(append(bytes, byte('\n')))
		if err != nil {
			log.Println(err)
		}
	}
}

// HandleRequest handles incoming connection
func HandleRequest(conn net.Conn) {
	lenbuf := make([]byte, 4)
	buf := make([]byte, 1024*1024*32)
	for {
		remain := 4
		head := 0
		for remain > 0 {
			l, err := conn.Read(lenbuf[head : head+remain])
			if err == io.EOF {
				return
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
			log.Println(info.GetName(), "couldn't read incoming request length")
			break
		}
		totalLength := int(binary.BigEndian.Uint32(lenbuf))
		if totalLength > cap(buf) {
			log.Println(info.GetName(), "buffer is not large enough, allocate more", totalLength)
			buf = make([]byte, totalLength)
		}
		remain = totalLength
		head = 0
		for remain > 0 {
			l, err := conn.Read(buf[head : head+remain])
			if err == io.EOF {
				return
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
		if buf[0] == 'i' { // get LId by tags
			var tags map[string]string
			err := json.Unmarshal(buf[1:totalLength], &tags)
			if err != nil {
				log.Println(info.GetName(), "couldn't unmarshal tags:", string(buf[1:totalLength]))
				log.Panicln(err)
			}
			lids := GetByTags(tags)
			tmp, err := json.Marshal(lids)
			if err != nil {
				tmp = []byte(fmt.Sprintln(err))
			}
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, uint32(len(tmp)))
			conn.Write(append(b, tmp...))
		} else if buf[0] == 't' { // insert tags into hash table
			lid := int(binary.BigEndian.Uint32(buf[1:5]))
			var tags map[string]string
			dec := gob.NewDecoder(bytes.NewBuffer(buf[5:totalLength]))
			err := dec.Decode(&tags)
			if err != nil {
				log.Println(info.GetName(), "couldn't decode tags")
				log.Panicln(err)
			}
			for key, value := range tags {
				Insert(key, value, lid)
			}
		} else if buf[0] == 'g' { // get LIds by tags
			var tags map[string]string
			err := json.Unmarshal(buf[1:totalLength], &tags)
			if err != nil {
				log.Println(info.GetName(), "couldn't unmarshal tags:", string(buf[1:totalLength]))
				log.Panicln(err)
			}
			lids := GetByTags(tags)
			b, err := json.Marshal(lids)
			if err != nil {
				conn.Write([]byte(fmt.Sprintln(err)))
			} else {
				conn.Write(b)
			}
		} else if buf[0] == 's' {
			log.Println(info.GetName(), "got subscription")
			Subscriber = conn
		} else {
			log.Println(info.GetName(), "couldn't understand", string(buf))
		}
	}
}
