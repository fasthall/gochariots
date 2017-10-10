package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"

	"cloud.google.com/go/datastore"
	"github.com/fasthall/gochariots/record"
)

var ctx context.Context
var client *datastore.Client

type Record struct {
	Id        string `datastore:"Id"`
	Timestamp int64  `datastore:"Timestamp"`
	Host      uint32 `datastore:"Host"`
	LId       uint32 `datastore:"LId"`
	Tags      string `datastore:"Tags"`
	Parent    string `datastore:"Parent"`
	Seed      string `datastore:"Seed"`
}

type TOIDRecord struct {
	Id        string `datastore:"Id"`
	Timestamp int64  `datastore:"Timestamp"`
	Host      uint32 `datastore:"Host"`
	TOId      uint32 `datastore:"TOId"`
	LId       uint32 `datastore:"LId"`
	Tags      string `datastore:"Tags"`
	PreHost   uint32 `datastore:"PreHash"`
	PreTOId   uint32 `datastore:"PreTOId"`
}

func init() {
	ctx = context.Background()
	var err error
	projectID := os.Getenv("DATASTORE_PROJECT_ID")
	if projectID != "" {
		client, err = datastore.NewClient(ctx, os.Getenv("DATASTORE_PROJECT_ID"))
		if err != nil {
			logrus.WithError(err).Error("couldn't connect to datastore")
			panic(err)
		}
	}
}

func PutRecord(r record.Record) error {
	tags, err := json.Marshal(r.Tags)
	if err != nil {
		return err
	}
	entity := &Record{
		Timestamp: r.Timestamp,
		Host:      r.Host,
		LId:       r.LId,
		Tags:      string(tags),
		Parent:    fmt.Sprintf("%v", r.Parent),
		Seed:      fmt.Sprintf("%v", r.Seed),
	}
	key := datastore.IncompleteKey("Record", nil)
	key, err = client.Put(ctx, key, entity)
	return err
}

func PutRecords(records []record.Record) error {
	for _, r := range records {
		err := PutRecord(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func PutTOIDRecord(r record.TOIDRecord) error {
	tags, err := json.Marshal(r.Tags)
	if err != nil {
		return err
	}
	entity := &TOIDRecord{
		Id:        r.Id,
		Timestamp: r.Timestamp,
		Host:      r.Host,
		TOId:      r.TOId,
		LId:       r.LId,
		Tags:      string(tags),
		PreHost:   r.Pre.Host,
		PreTOId:   r.Pre.TOId,
	}
	key := datastore.IncompleteKey("Record", nil)
	key, err = client.Put(ctx, key, entity)
	return err
}

func PutTOIDRecords(records []record.TOIDRecord) error {
	for _, r := range records {
		err := PutTOIDRecord(r)
		if err != nil {
			return err
		}
	}
	return nil
}
