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

	return c.Insert(&r)
}

func PutRecords(records []record.Record) error {
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

func GetRecord(id uint64) (record.Record, error) {
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

type Query struct {
	Id   []string
	Seed string
}

func QueryDB(queries []Query) ([]bool, error) {
	if c == nil {
		connect()
	}

	existed := make([]bool, len(queries))
	for i, q := range queries {
		bsonm := []bson.M{}
		for _, id := range q.Id {
			bsonm = append(bsonm, bson.M{"seed": q.Seed, "_id": id})
		}
		cnt, err := c.Find(bson.M{"$or": bsonm}).Count()
		if err != nil {
			return nil, err
		}
		if cnt == len(q.Id) {
			existed[i] = true
		}
	}
	return existed, nil
}
