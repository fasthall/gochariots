package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/fasthall/gochariots/record"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const DB_NAME string = "gochariots"
const COLLECTION_NAME string = "record"

var c *mgo.Collection
var batchSize = 1000
var batch = 1000

func connect() {
	session, err := mgo.Dial("169.231.235.70")
	if err != nil {
		panic(err)
	}
	c = session.DB(DB_NAME).C(COLLECTION_NAME)
}

func PutRecord(r record.Record) error {
	if c == nil {
		connect()
	}

	err := c.Insert(&r)
	if mgo.IsDup(err) {
		err := c.UpdateId(r.Id, bson.M{"$set": bson.M{"timestamp": r.Timestamp, "host": r.Host, "tags": r.Tags, "parent": r.Parent, "seed": r.Seed}})
		return err
	}
	return err
}

func PutRecords(records []record.Record) error {
	if c == nil {
		connect()
	}

	objs := make([]interface{}, len(records))
	for i, r := range records {
		objs[i] = r
	}

	return c.Insert(objs...)
}

func UpdateLId(id string, lid uint32) error {
	if c == nil {
		connect()
	}

	err := c.Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"lid": lid}})
	if err == mgo.ErrNotFound {
		r := record.Record{
			Id:  id,
			LId: lid,
		}
		return PutRecord(r)
	}
	return err
}

func UpdateLIds(id []string, lid []uint32) error {
	if c == nil {
		connect()
	}
	bulk := c.Bulk()
	q := make([]interface{}, len(id)*2)
	for i := range id {
		q[2*i] = bson.M{"_id": id[i]}
		q[2*i+1] = bson.M{"$set": bson.M{"lid": lid[i]}}
	}
	bulk.Update(q...)
	_, err := bulk.Run()
	if err == mgo.ErrNotFound {
		fmt.Println(err)
	}
	return err
}

func GetRecord(id string) (record.Record, error) {
	if c == nil {
		connect()
	}

	var r record.Record
	err := c.FindId(id).One(&r)
	return r, err
}

func QueryDB(queries []string) ([]bool, error) {
	if c == nil {
		connect()
	}

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
