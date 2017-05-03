package index

import "hash/fnv"

var indexes = make(map[uint64][]int)

func toHash(b []byte) uint64 {
	hash := fnv.New64a()
	hash.Write(b)
	return hash.Sum64()
}

func Insert(key, value string, LId int) {
	h := toHash([]byte(key + ":" + value))
	indexes[h] = append(indexes[h], LId)
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
