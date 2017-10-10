package mongodb

import (
	"os"

	"github.com/fasthall/gochariots/record"
	"github.com/satori/go.uuid"
	mgo "gopkg.in/mgo.v2"
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

	if r.Id == "" {
		r.Id = uuid.NewV4().String()
	}
	return c.Insert(&r)
}

func PutRecords(records []record.Record) error {
	if c == nil {
		connect()
	}

	for _, r := range records {
		if r.Id == "" {
			r.Id = uuid.NewV4().String()
		}
		err := c.Insert(&r)
		if err != nil {
			return err
		}
	}
	return nil
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

	for _, r := range records {
		if r.Id == "" {
			r.Id = uuid.NewV4().String()
		}
		err := c.Insert(&r)
		if err != nil {
			return err
		}
	}
	return nil
}
