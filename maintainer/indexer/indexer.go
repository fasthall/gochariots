package indexer

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/misc/connection"
)

var Subscriber net.Conn
var db *bolt.DB

const bucket = "index"

type Query struct {
	Hash uint64
	Seed uint64
}

type IndexTableEntry struct {
	LId  int
	Seed uint64
}

func InsertEntry(key, value string, lid int, seed uint64) {
	entry := IndexTableEntry{
		LId:  lid,
		Seed: seed,
	}
	hash := tagToHash(key, value)
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		k := make([]byte, 8)
		binary.BigEndian.PutUint64(k, hash)
		tmp := b.Get(k)
		var result []IndexTableEntry
		if tmp != nil {
			err = json.Unmarshal(b.Get(k), result)
			if err != nil {
				return err
			}
		}
		result = append(result, entry)
		bytes, err := json.Marshal(result)
		if err != nil {
			return err
		}
		err = b.Put(k, bytes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("couldn't insert entry")
	}
}

func GetEntry(hash uint64) []IndexTableEntry {
	var val []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		k := make([]byte, 8)
		binary.BigEndian.PutUint64(k, hash)
		val = b.Get(k)
		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("couldn't read db")
		panic(err)
	}
	if val == nil {
		return []IndexTableEntry{}
	}
	var result []IndexTableEntry
	err = json.Unmarshal(val, &result)
	if err != nil {
		logrus.WithError(err).Error("couldn't decode entry")
		panic(err)
	}
	return result
}

func InitIndexer(p string) {
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
	result := GetEntry(hash)
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
				for _, s := range GetEntry(q.Hash) {
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
				InsertEntry(key, value, lid, seed)
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
