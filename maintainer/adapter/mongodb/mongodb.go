package mongodb

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/fasthall/gochariots/record"
	"github.com/google/uuid"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const DB_NAME string = "lowgo"
const COLLECTION_RECORD string = "record"

var INSERT_SIZE_LIMIT = 1000

type Client struct {
	session *mgo.Session
}

func NewClient(host string) Client {
	client := Client{}
	err := errors.New("")
	for err != nil {
		client.session, err = mgo.Dial(host)
		logrus.WithField("host", host).Warning("MongoDB not ready, retry in 1 second")
		time.Sleep(1 * time.Second)
	}
	client.session.SetPoolLimit(50)
	client.session.SetMode(mgo.Strong, true)
	logrus.WithField("host", host).Info("connected to MongoDB")
	return client
}

func (cli *Client) PutRecords(records []record.Record) error {
	sessionCopy := cli.session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)
	if len(records) > INSERT_SIZE_LIMIT {
		cli.PutRecords(records[INSERT_SIZE_LIMIT:])
		records = records[:INSERT_SIZE_LIMIT]
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

func (cli *Client) PutTOIDRecords(records []record.TOIDRecord) error {
	sessionCopy := cli.session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)
	if len(records) > INSERT_SIZE_LIMIT {
		cli.PutTOIDRecords(records[INSERT_SIZE_LIMIT:])
		records = records[:INSERT_SIZE_LIMIT]
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

func (cli *Client) GetRecord(id string) (record.Record, error) {
	sessionCopy := cli.session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)

	var r record.Record
	err := c.FindId(id).One(&r)
	return r, err
}

func (cli *Client) PutTOIDRecord(r record.TOIDRecord) error {
	sessionCopy := cli.session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)

	if r.Id == "" {
		r.Id = uuid.New().String()
	}
	return c.Insert(&r)
}

func (cli *Client) QueryDB(queries *map[string]bool) error {
	ids := make([]string, 0, len(*queries))
	for key, _ := range *queries {
		ids = append(ids, key)
	}
	sessionCopy := cli.session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)
	q := c.Find(bson.M{"_id": bson.M{"$in": ids}})
	var result []record.Record
	err := q.Select(bson.M{"_id": 1}).All(&result)
	for _, r := range result {
		_, found := (*queries)[r.ID]
		if found {
			(*queries)[r.ID] = true
		}
	}
	return err
}
