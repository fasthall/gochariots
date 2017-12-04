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
const COLLECTION_RECORD string = "record"
const COLLECTION_MAPPING string = "mapping"

var mismatched = 0
var cnt = 0

var session *mgo.Session

type Mapping struct {
	Id  string `bson:"_id,omitempty"`
	LId uint32
}

func init() {
	if os.Getenv("MONGODB_HOST") != "" {
		var err error
		session, err = mgo.Dial(os.Getenv("MONGODB_HOST"))
		session.SetPoolLimit(50)
		session.SetMode(mgo.Strong, true)
		if err != nil {
			panic(err)
		}
	}
}

func PutRecords(records []record.Record) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)
	if len(records) > 1000 {
		PutRecords(records[1000:])
		records = records[:1000]
	}
	bulk := c.Bulk()
	objs := make([]interface{}, len(records))
	for i, r := range records {
		objs[i] = r
	}
	bulk.Unordered()
	bulk.Insert(objs...)
	_, err := bulk.Run()
	return err
}

func PutTOIDRecords(records []record.TOIDRecord) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)

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

func InsertLIds(id []string, lid []uint32) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_MAPPING)
	if len(id) != len(lid) {
		return errors.New("length doesn't match")
	}
	for len(id) > 1000 {
		InsertLIds(id[:1000], lid[:1000])
		id = id[1000:]
		lid = lid[1000:]
	}
	bulk := c.Bulk()
	q := make([]interface{}, len(id))
	for i := range id {
		q[i] = Mapping{Id: id[i], LId: lid[i]}
	}
	bulk.Unordered()
	bulk.Insert(q...)
	_, err := bulk.Run()
	return err
}

func GetRecord(id string) (record.Record, error) {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)

	var r record.Record
	err := c.FindId(id).One(&r)
	return r, err
}

func PutTOIDRecord(r record.TOIDRecord) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)

	if r.Id == "" {
		r.Id = uuid.NewV4().String()
	}
	return c.Insert(&r)
}

func QueryDB(queries *map[string]bool) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_MAPPING)

	for key, _ := range *queries {
		cnt, err := c.Find(bson.M{"_id": key}).Count()
		if err != nil {
			return err
		}
		if cnt > 0 {
			(*queries)[key] = true
		}
	}
	return nil
}
