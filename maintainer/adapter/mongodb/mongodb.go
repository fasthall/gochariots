package mongodb

import (
	"os"

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
