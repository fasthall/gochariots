package mongodb

import (
	"errors"
	"os"

	"github.com/Sirupsen/logrus"

	"github.com/fasthall/gochariots/record"
	"github.com/satori/go.uuid"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const DB_NAME string = "gochariots"
const COLLECTION_NAME string = "record"

var c *mgo.Collection

func connect() {
	if os.Getenv("MONGODB_HOST") != "" {
		session, err := mgo.Dial(os.Getenv("MONGODB_HOST"))
		if err != nil {
			panic(err)
		}
		c = session.DB(DB_NAME).C(COLLECTION_NAME)
	}
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
	if len(id) != len(lid) {
		return errors.New("length doesn't match")
	}
	for len(id) > 1000 {
		UpdateLIds(id[:1000], lid[:1000])
		id = id[1000:]
		lid = lid[1000:]
	}
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
		logrus.WithError(err).Error("mgo.ErrNotFound")
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

func PutTOIDRecord(r record.TOIDRecord) error {
	if c == nil {
		connect()
	}

	if r.Id == "" {
		r.Id = uuid.NewV4().String()
	}
	return c.Insert(&r)
}

func PutTOIDRecords(records []record.TOIDRecord) error {
	if c == nil {
		connect()
	}

	objs := make([]interface{}, len(records))
	for i, r := range records {
		if r.Id == "" {
			r.Id = uuid.NewV4().String()
		}
		objs[i] = r
	}

	return c.Insert(objs...)
}

func QueryDB(queries *map[string]bool) error {
	if c == nil {
		connect()
	}

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
