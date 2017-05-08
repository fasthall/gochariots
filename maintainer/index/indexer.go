package index

import (
	"encoding/json"
	"hash/fnv"
	"log"
	"net"
	"strconv"
)

var indexes = make(map[uint64][]int)
var Subscriber net.Conn

func toHash(b []byte) uint64 {
	hash := fnv.New64a()
	hash.Write(b)
	return hash.Sum64()
}

func Insert(key, value string, LId int) {
	h := toHash([]byte(key + ":" + value))
	indexes[h] = append(indexes[h], LId)
	notify(key, value, LId)
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

func notify(key, value string, LId int) {
	if Subscriber != nil {
		payload := map[string]string{key: value, "LId": strconv.Itoa(LId)}
		bytes, err := json.Marshal(payload)
		if err != nil {
			log.Println(err)
		}
		// b := make([]byte, 4)
		// binary.BigEndian.PutUint32(b, uint32(len(bytes)+1))
		_, err = Subscriber.Write(append(bytes, byte('\n')))
		if err != nil {
			log.Println(err)
		}
	}
}
