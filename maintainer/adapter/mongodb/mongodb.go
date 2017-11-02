package mongodb

import (
	"errors"
	"os"

	"github.com/fasthall/gochariots/record"
	"github.com/satori/go.uuid"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const DB_NAME string = "gochariots"
const COLLECTION_NAME string = "record"

var session *mgo.Session

func init() {
	if os.Getenv("MONGODB_HOST") != "" {
		var err error
		session, err = mgo.Dial(os.Getenv("MONGODB_HOST"))
		session.SetMode(mgo.Monotonic, true)
		session.SetPoolLimit(50)
		if err != nil {
			panic(err)
		}
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

func PutTOIDRecords(records []record.TOIDRecord) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)

	if len(records) > 1000 {
		PutTOIDRecords(records[1000:])
		records = records[:1000]
	}
	bulk := c.Bulk()
	objs := make([]interface{}, len(records)*2)
	for i, r := range records {
		objs[2*i] = bson.M{"_id": r.Id}
		objs[2*i+1] = bson.M{
			"$set": bson.M{"host": r.Host, "toid": r.TOId, "lid": r.LId, "tags": r.Tags, "prehost": r.Pre.Host, "pretoid": r.Pre.TOId, "timestamp": r.Timestamp},
		}
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
	if len(id) != len(lid) {
		return errors.New("length doesn't match")
	}
	for len(id) > 1000 {
		UpdateLIds(id[1000:], lid[1000:])
		id = id[:1000]
		lid = lid[:1000]
	}
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)

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

func GetRecord(id string) (record.Record, error) {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)

	var r record.Record
	err := c.FindId(id).One(&r)
	return r, err
}

func PutTOIDRecord(r record.TOIDRecord) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)

	if r.Id == "" {
		r.Id = uuid.NewV4().String()
	}
	return c.Insert(&r)
}

// func PutTOIDRecords(records []record.TOIDRecord) error {
// 	sessionCopy := session.Copy()
// 	defer sessionCopy.Close()
// 	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)

// 	objs := make([]interface{}, len(records))
// 	for i, r := range records {
// 		if r.Id == "" {
// 			r.Id = uuid.NewV4().String()
// 		}
// 		objs[i] = r
// 	}

// 	return c.Insert(objs...)
// }

func QueryDB(queries *map[string]bool) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_NAME)

	for key, _ := range *queries {
		cnt, err := c.Find(bson.M{"_id": key, "lid": bson.M{"$gt": 0}}).Count()
		if err != nil {
			return err
		}
		if cnt > 0 {
			(*queries)[key] = true
		}
	}
	return nil
}
