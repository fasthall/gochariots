package indexer

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/misc/connection"
)

var indexes IndexTable
var Subscriber net.Conn

type Query struct {
	Hash uint64
	Seed uint64
}

type IndexTableEntry struct {
	LId  int
	Seed uint64
}

type IndexTable struct {
	table map[uint64][]IndexTableEntry
	mutex sync.Mutex
}

func NewIndexTable() IndexTable {
	index := IndexTable{
		table: map[uint64][]IndexTableEntry{},
	}
	return index
}

func (it *IndexTable) Insert(key, value string, lid int, seed uint64) {
	hash := tagToHash(key, value)
	it.mutex.Lock()
	it.table[hash] = append(it.table[hash], IndexTableEntry{
		LId:  lid,
		Seed: seed,
	})
	it.mutex.Unlock()
}

func (it *IndexTable) Get(key, value string) []IndexTableEntry {
	hash := tagToHash(key, value)
	return it.table[hash]
}

func (it *IndexTable) GetByHash(hash uint64) []IndexTableEntry {
	it.mutex.Lock()
	result := it.table[hash]
	it.mutex.Unlock()
	return result
}

func InitIndexer(p string) {
	indexes = NewIndexTable()
	logrus.Info("indexer initialized")
}

func tagToHash(key, value string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(key + ":" + value))
	return hash.Sum64()
}

func ToHash(b []byte) uint64 {
	hash := fnv.New64a()
	hash.Write(b)
	return hash.Sum64()
}

func getLIdByTag(key, value string) []int {
	result := indexes.Get(key, value)
	ans := []int{}
	for _, r := range result {
		ans = append(ans, r.LId)
	}
	return ans
}

func getLIdByTags(tags map[string]string) []int {
	ans := []int{}
	for k, v := range tags {
		ans = append(ans, getLIdByTag(k, v)...)
	}
	return ans
}

func getLIdByHash(hash uint64) []int {
	result := indexes.GetByHash(hash)
	ans := []int{}
	for _, r := range result {
		ans = append(ans, r.LId)
	}
	return ans
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
		if buf[0] == 'q' { // query from queue
			var query []Query
			dec := gob.NewDecoder(bytes.NewBuffer(buf[1:totalLength]))
			err = dec.Decode(&query)
			if err != nil {
				log.Println(info.GetName(), "couldn't decode query")
				log.Println(info.GetName(), err)
				break
			}
			ans := make([]bool, len(query))
			for i, q := range query {
				for _, s := range indexes.GetByHash(q.Hash) {
					if s.Seed == q.Seed {
						ans[i] = true
						break
					}
				}
			}
			var tmp bytes.Buffer
			enc := gob.NewEncoder(&tmp)
			err := enc.Encode(ans)
			if err != nil {
				log.Println(info.GetName(), "couldn't encode answer boolean slice")
				log.Println(info.GetName(), err)
				break
			}
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, uint32(len(tmp.Bytes())))
			conn.Write(append(b, tmp.Bytes()...))
		} else if buf[0] == 't' { // insert tags into hash table
			lid := int(binary.BigEndian.Uint32(buf[1:5]))
			seed := binary.BigEndian.Uint64(buf[5:13])
			var tags map[string]string
			dec := gob.NewDecoder(bytes.NewBuffer(buf[13:totalLength]))
			err := dec.Decode(&tags)
			if err != nil {
				log.Println(info.GetName(), "couldn't decode tags")
				log.Panicln(err)
			}
			for key, value := range tags {
				indexes.Insert(key, value, lid, seed)
			}
		} else if buf[0] == 'g' { // get LIds by tags
			var tags map[string]string
			err := json.Unmarshal(buf[1:totalLength], &tags)
			if err != nil {
				log.Println(info.GetName(), "couldn't unmarshal tags:", string(buf[1:totalLength]))
				log.Panicln(err)
			}
			lids := getLIdByTags(tags)
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
