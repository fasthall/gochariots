package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/fasthall/gochariots/record"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const DB_NAME string = "gochariots"
const COLLECTION_NAME string = "record"

var session *mgo.Session
var batchSize = 1000
var batch = 1000

func init() {
	var err error
	session, err = mgo.Dial("169.231.235.70")
	session.SetMode(mgo.Monotonic, true)
	session.SetPoolLimit(50)
	if err != nil {
		panic(err)
	}
}

func PutRecord(r record.Record) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)
	_, err := c.UpsertId(r.Id, &r)
	return err
}

func PutRecords(records []record.Record) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)
	if len(records) > 1000 {
		PutRecords(records[1000:])
		records = records[:1000]
	}
	bulk := c.Bulk()
	objs := make([]interface{}, len(records)*2)
	for i, r := range records {
		objs[2*i] = bson.M{"_id": r.Id}
		objs[2*i+1] = bson.M{
			"$set":         bson.M{"host": r.Host, "tags": r.Tags, "parent": r.Parent, "seed": r.Seed, "timestamp": r.Timestamp},
			"$setOnInsert": bson.M{"lid": r.LId}}
	}
	bulk.Upsert(objs...)
	_, err := bulk.Run()
	return err
}

func UpdateLId(id string, lid uint32) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)
	_, err := c.Upsert(bson.M{"_id": id}, bson.M{"$set": bson.M{"lid": lid}})
	return err
}

func UpdateLIds(id []string, lid []uint32) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)
	if len(id) != len(lid) {
		return errors.New("length doesn't match")
	}
	for len(id) > 1000 {
		UpdateLIds(id[:1000], lid[:1000])
		id = id[1000:]
		lid = lid[1000:]
	}
	bulk := c.Bulk()
	q := make([]interface{}, len(id)*2)
	for i := range id {
		q[2*i] = bson.M{"_id": id[i]}
		q[2*i+1] = bson.M{"$set": bson.M{"lid": lid[i]}}
	}
	bulk.Upsert(q...)
	_, err := bulk.Run()
	return err
}

func QueryDB(queries []string) ([]bool, error) {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)
	existed := make([]bool, len(queries))
	for i, id := range queries {
		if id == "" {
			existed[i] = true
		} else {
			cnt, err := c.Find(bson.M{"_id": id, "lid": bson.M{"$gt": 0}}).Count()
			if err != nil {
				return nil, err
			}
			if cnt > 0 {
				existed[i] = true
			}
		}
	}
	return existed, nil
}

func main() {
	var wg sync.WaitGroup
	wg.Add(batch)

	for b := 0; b < batch; b++ {
		id := []string{}
		lid := []uint32{}
		for i := 0; i < batchSize; i++ {
			id = append(id, strconv.Itoa(b*batchSize+i))
			lid = append(lid, uint32(b*batchSize+i+1))
		}
		go func() {
			defer wg.Done()
			err := UpdateLIds(id, lid)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}

	// for b := 0; b < batch; b++ {
	// 	records := make([]record.Record, batchSize)
	// 	for i := 0; i < batchSize; i++ {
	// 		records[i] = record.Record{
	// 			Id:     strconv.Itoa(b*batchSize + i),
	// 			Host:   1,
	// 			Tags:   map[string]string{},
	// 			Parent: "",
	// 			Seed:   uuid.NewV4().String(),
	// 		}
	// 	}
	// 	go func() {
	// 		defer wg.Done()
	// 		err := PutRecords(records)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 	}()
	// }

	wg.Wait()
}
