package main

import (
	"fmt"
	"strconv"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/fasthall/gochariots/record"
)

const DB_NAME string = "gochariots"
const COLLECTION_NAME string = "record"

var c *mgo.Collection

var batch = 100000

func connect() {
	session, err := mgo.Dial("169.231.235.70")
	if err != nil {
		panic(err)
	}
	c = session.DB(DB_NAME).C(COLLECTION_NAME)
}

func GetRecordByID(id string) (record.Record, error) {
	if c == nil {
		connect()
	}

	var r record.Record
	err := c.Find(bson.M{"_id": id}).One(&r)
	return r, err
}

func GetRecord(lid uint32) (record.Record, error) {
	if c == nil {
		connect()
	}

	var r record.Record
	err := c.Find(bson.M{"lid": lid}).One(&r)
	return r, err
}

func main() {
	r1, err := GetRecordByID("0")
	if err != nil {
		panic("r1" + err.Error())
	}
	fmt.Println(r1.Timestamp)
	r2, err := GetRecordByID(strconv.Itoa(batch - 1))
	if err != nil {
		panic("r2" + err.Error())
	}
	fmt.Println(r2.Timestamp)
	fmt.Println(strconv.FormatInt(int64((r2.Timestamp-r1.Timestamp)/1000000), 10) + "ms")
}
