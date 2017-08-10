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
	Timestamp int64  `datastore:"Timestamp"`
	Host      int    `datastore:"Host"`
	LId       int    `datastore:"LId"`
	Tags      string `datastore:"Tags"`
	Hash      string `datastore:"Hash"`
	Seed      string `datastore:"Seed"`
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
		Hash:      fmt.Sprintf("%v", r.Hash),
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
