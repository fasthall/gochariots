package indexer

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/misc"
	"github.com/fasthall/gochariots/misc/connection"
	cache "github.com/patrickmn/go-cache"
)

var Subscriber net.Conn
var boltDB bool
var db *bolt.DB
var gocache *cache.Cache
var indexes IndexTable

const bucket = "index"

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
	return it.GetByHash(hash)
}

func (it *IndexTable) GetByHash(hash uint64) []IndexTableEntry {
	it.mutex.Lock()
	result := it.table[hash]
	it.mutex.Unlock()
	return result
}

func bytesToEntryList(bytes []byte) ([]IndexTableEntry, error) {
	var result []IndexTableEntry
	if bytes != nil {
		err := json.Unmarshal(bytes, &result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func getBytesFromDB(hash uint64) []byte {
	logrus.WithField("hash", hash).Info("getBytesFromDB")
	var buf []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		k := make([]byte, 8)
		binary.BigEndian.PutUint64(k, hash)
		tmp := b.Get(k)
		if tmp != nil {
			buf = make([]byte, len(tmp))
			copy(buf, tmp)
		}
		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("couldn't read db")
		panic(err)
	}
	return buf
}

func insertIndexEntry(key, value string, lid int, seed uint64) {
	entry := IndexTableEntry{
		LId:  lid,
		Seed: seed,
	}
	hash := tagToHash(key, value)
	if !boltDB {
		indexes.Insert(key, value, lid, seed)
	} else {
		logrus.WithFields(logrus.Fields{"hash": hash, "entry": entry}).Info("insertIndexEntry")
		k := make([]byte, 8)
		binary.BigEndian.PutUint64(k, hash)
		tmp, found := gocache.Get(string(hash))
		if !found {
			logrus.WithField("hash", hash).Warning("key not found in cache")
			tmp = getBytesFromDB(hash)
		}
		result, err := bytesToEntryList(tmp.([]byte))
		if err != nil {
			logrus.WithError(err).Error("couldn't convert DB entry to IndexEntry")
			panic(err)
		}
		result = append(result, entry)
		bytes, err := json.Marshal(result)
		if err != nil {
			logrus.WithError(err).Error("couldn't marshal IndexEntry list")
			panic(err)
		}

		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}

			err = b.Put(k, bytes)
			if err != nil {
				return err
			}
			gocache.Set(string(hash), bytes, cache.DefaultExpiration)
			return nil
		})
		if err != nil {
			logrus.WithError(err).Error("couldn't insert entry")
		}
	}
	logrus.WithField("entry", entry).Debug("entry inserted")
}

func getIndexEntry(hash uint64) ([]IndexTableEntry, error) {
	if boltDB {
		bytes := getBytesFromDB(hash)
		list, err := bytesToEntryList(bytes)
		logrus.WithField("result", list).Info("getIndexEntry")
		return list, err
	}
	return indexes.GetByHash(hash), nil
}

func InitIndexer(p string, boltdb bool) {
	boltDB = boltdb
	if !boltdb {
		indexes = NewIndexTable()
	} else {
		var err error
		db, err = bolt.Open(info.GetName()+".db", 0644, nil)
		if err != nil {
			logrus.WithError(err).Error("couldn't open db")
			panic(err)
		}
		err = db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			logrus.WithError(err).Error("couldn't create bucket")
			panic(err)
		}
		gocache = cache.New(5*time.Minute, 10*time.Minute)
	}
	logrus.Info("indexer initialized")
}

func Config(file string) {
	config, err := misc.ReadConfig(file)
	if err != nil {
		logrus.WithError(err).Warn("read config file failed")
		return
	}
	if config.Controller == "" {
		logrus.Error("No controller information found in config file")
		return
	}
	addr, err := misc.GetHostIP()
	if err != nil {
		logrus.WithError(err).Error("couldn't find local IP address")
		return
	}
	p := misc.NewParams()
	p.AddParam("host", addr+":"+info.GetPort())
	logrus.WithFields(logrus.Fields{"controller": config.Controller}).Info("Config file read")

	err = errors.New("")
	for err != nil {
		time.Sleep(3 * time.Second)
		err = misc.Report(config.Controller, "indexer", p)
		if err != nil {
			logrus.WithError(err).Error("couldn't report to the controller")
		}
	}
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
	hash := tagToHash(key, value)
	return getLIdByHash(hash)
}

func getLIdByTags(tags map[string]string) []int {
	ans := []int{}
	for k, v := range tags {
		ans = append(ans, getLIdByTag(k, v)...)
	}
	return ans
}

func getLIdByHash(hash uint64) []int {
	result, err := getIndexEntry(hash)
	if err != nil {
		logrus.WithError(err).Error("couldn't read from indexer")
		panic(err)
	}
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
			logrus.Error("couldn't read incoming request")
			break
		}
		if buf[0] == 'q' { // query from queue
			var query []Query
			dec := gob.NewDecoder(bytes.NewBuffer(buf[1:totalLength]))
			err = dec.Decode(&query)
			if err != nil {
				logrus.WithField("buffer", buf[1:totalLength]).Error("couldn't decode query")
				break
			}
			ans := make([]bool, len(query))
			for i, q := range query {
				result, err := getIndexEntry(q.Hash)
				if err != nil {
					logrus.WithError(err).Error("couldn't read from indexer")
					panic(err)
				}
				for _, s := range result {
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
				logrus.WithError(err).Error("couldn't encode answer boolean slice")
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
				logrus.WithField("buffer", buf[1:totalLength]).Error("couldn't decode tags")
				panic(err)
			}
			for key, value := range tags {
				insertIndexEntry(key, value, lid, seed)
			}
		} else if buf[0] == 'g' { // get LIds by tags
			var tags map[string]string
			err := json.Unmarshal(buf[1:totalLength], &tags)
			if err != nil {
				logrus.WithField("buffer", buf[1:totalLength]).Error("couldn't unmarshal tags")
				panic(err)
			}
			lids := getLIdByTags(tags)
			b, err := json.Marshal(lids)
			if err != nil {
				conn.Write([]byte(fmt.Sprintln(err)))
			} else {
				conn.Write(b)
			}
		} else if buf[0] == 's' {
			logrus.Info("got subscription")
			Subscriber = conn
		} else {
			logrus.WithField("header", buf[0]).Warning("couldn't understand request")
		}
	}
}
