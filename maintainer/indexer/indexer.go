package indexer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/fasthall/gochariots/info"
	cache "github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
)

var boltDB bool
var db *bolt.DB
var gocache *cache.Cache
var indexes IndexTable

const bucket = "index"

type Query struct {
	Hash []uint64
	Seed uint64
}

type IndexTableEntry struct {
	LId  uint32
	Seed uint64
}

type IndexTable struct {
	table map[uint64][]IndexTableEntry
	mutex sync.Mutex
}

type Server struct{}

func (s *Server) Query(ctx context.Context, in *RPCQueries) (*RPCQueryReply, error) {
	ans := make([]bool, len(in.GetQueries()))
	for i, q := range in.GetQueries() {
		result := true
		for _, hash := range q.GetHash() {
			seeds, err := getIndexEntry(hash)
			if err != nil {
				logrus.WithError(err).Error("couldn't read from indexer")
				panic(err)
			}
			in := false
			for _, s := range seeds {
				if s.Seed == q.GetSeed() {
					in = true
					break
				}
			}
			if in == false {
				result = false
				break
			}
		}
		if result {
			ans[i] = true
		}

	}
	return &RPCQueryReply{Reply: ans}, nil
}

func (s *Server) InsertTags(ctx context.Context, in *RPCTags) (*RPCReply, error) {
	for key, value := range in.GetTags() {
		insertIndexEntry(key, value, in.GetLid(), in.GetSeed())
	}
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) GetLIds(ctx context.Context, in *RPCTags) (*RPCLIds, error) {
	lids := getLIdByTags(in.GetTags())
	return &RPCLIds{Lid: lids}, nil
}

func NewIndexTable() IndexTable {
	index := IndexTable{
		table: map[uint64][]IndexTableEntry{},
	}
	return index
}

func (it *IndexTable) Insert(key, value string, lid uint32, seed uint64) {
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
	logrus.WithFields(logrus.Fields{"hash": hash, "return": result}).Debug("GetByHash")
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

func insertIndexEntry(key, value string, lid uint32, seed uint64) {
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
	logrus.WithFields(logrus.Fields{"LId": entry.LId, "seed": entry.Seed, "hash": hash}).Debug("entry inserted")
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
	logrus.WithField("BoltDB", boltDB).Info("indexer initialized with BoltDB")
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

func getLIdByTag(key, value string) []uint32 {
	hash := tagToHash(key, value)
	return getLIdByHash(hash)
}

func getLIdByTags(tags map[string]string) []uint32 {
	ans := []uint32{}
	for k, v := range tags {
		ans = append(ans, getLIdByTag(k, v)...)
	}
	return ans
}

func getLIdByHash(hash uint64) []uint32 {
	result, err := getIndexEntry(hash)
	if err != nil {
		logrus.WithError(err).Error("couldn't read from indexer")
		panic(err)
	}
	ans := []uint32{}
	for _, r := range result {
		ans = append(ans, r.LId)
	}
	return ans
}
