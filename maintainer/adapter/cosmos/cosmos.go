package cosmos

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/fasthall/gochariots/record"

	"gopkg.in/mgo.v2"
)

var (
	DATABASE string
	PASSWORD string
)

var session *mgo.Session

func init() {
	DATABASE = os.Getenv("COSMOS_USERNAME")
	PASSWORD = os.Getenv("COSMOS_PASSWORD")

	if DATABASE == "" || PASSWORD == "" {
		return
	}

	// DialInfo holds options for establishing a session with a MongoDB cluster.
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s.documents.azure.com:10255", DATABASE)}, // Get HOST + PORT
		Timeout:  60 * time.Second,
		Database: DATABASE, // It can be anything
		Username: DATABASE, // Username
		Password: PASSWORD, // PASSWORD
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		},
	}

	var err error
	session, err = mgo.DialWithInfo(dialInfo)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}

	session.SetSafe(&mgo.Safe{})
}

func PutRecord(r record.Record) error {
	collection := session.DB(DATABASE).C("record")
	err := collection.Insert(r)
	if err != nil {
		return err
	}
	return nil
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
	collection := session.DB(DATABASE).C("record")
	err := collection.Insert(r)
	if err != nil {
		return err
	}
	return nil
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
