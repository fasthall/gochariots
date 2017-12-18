package mongodb

import (
	"errors"
	"os"
	"sync"

	"github.com/fasthall/gochariots/record"
	"github.com/satori/go.uuid"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const DB_NAME string = "gochariots"
const COLLECTION_RECORD string = "record"
const COLLECTION_MAPPING string = "mapping"

var INSERT_SIZE_LIMIT = 1000

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
	if len(records) > INSERT_SIZE_LIMIT {
		PutRecords(records[INSERT_SIZE_LIMIT:])
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

func PutTOIDRecords(records []record.TOIDRecord) error {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)
	if len(records) > INSERT_SIZE_LIMIT {
		PutTOIDRecords(records[INSERT_SIZE_LIMIT:])
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

func InsertLIds(id []string, lid []uint32) error {
	if len(id) == 0 {
		return nil
	}
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	c := sessionCopy.DB(DB_NAME).C(COLLECTION_MAPPING)
	if len(id) != len(lid) {
		return errors.New("length doesn't match")
	}
	for len(id) > INSERT_SIZE_LIMIT {
		InsertLIds(id[:INSERT_SIZE_LIMIT], lid[:INSERT_SIZE_LIMIT])
		id = id[INSERT_SIZE_LIMIT:]
		lid = lid[INSERT_SIZE_LIMIT:]
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

func ParellelQueryDBCausality(queries []record.Causality, querySize int) ([]record.Causality, []record.Causality, error) {
	nonexist := []record.Causality{}
	exist := []record.Causality{}

	subQueries := []map[string]bool{map[string]bool{}}
	subQuery := map[string]bool{}
	for _, q := range queries {
		if len(subQuery) == querySize {
			subQueries = append(subQueries, subQuery)
			subQuery = map[string]bool{}
		}
		if q.Parent != "" {
			subQuery[q.Parent] = false
		}
	}
	if len(subQuery) > 0 {
		subQueries = append(subQueries, subQuery)
	}

	var wg sync.WaitGroup
	wg.Add(len(subQueries))
	// map
	for i := range subQueries {
		subQ := subQueries[i]
		go func() {
			_ = QueryDB(&subQ, true)
			wg.Done()
		}()
	}

	wg.Wait()
	// reduce
	result := map[string]bool{}
	for i := range subQueries {
		for k, v := range subQueries[i] {
			result[k] = v
		}
	}

	for _, q := range queries {
		if q.Parent == "" {
			exist = append(exist, q)
		} else {
			value, found := result[q.Parent]
			if found && value {
				exist = append(exist, q)
			} else {
				nonexist = append(nonexist, q)
			}
		}
	}
	return exist, nonexist, nil
}

func ParellelQueryDBRecord(queries []record.Record, querySize int) ([]record.Record, []record.Record, error) {
	nonexist := []record.Record{}
	exist := []record.Record{}

	subQueries := []map[string]bool{map[string]bool{}}
	subQuery := map[string]bool{}
	for _, q := range queries {
		if len(subQuery) == querySize {
			subQueries = append(subQueries, subQuery)
			subQuery = map[string]bool{}
		}
		if q.Parent != "" {
			subQuery[q.Parent] = false
		}
	}
	if len(subQuery) > 0 {
		subQueries = append(subQueries, subQuery)
	}

	var wg sync.WaitGroup
	wg.Add(len(subQueries))
	// map
	for i := range subQueries {
		subQ := subQueries[i]
		go func() {
			_ = QueryDB(&subQ, false)
			wg.Done()
		}()
	}

	wg.Wait()
	// reduce
	result := map[string]bool{}
	for i := range subQueries {
		for k, v := range subQueries[i] {
			result[k] = v
		}
	}

	for _, q := range queries {
		if q.Parent == "" {
			exist = append(exist, q)
		} else {
			value, found := result[q.Parent]
			if found && value {
				exist = append(exist, q)
			} else {
				nonexist = append(nonexist, q)
			}
		}
	}
	return exist, nonexist, nil
}

func QueryDB(queries *map[string]bool, twoPhase bool) error {
	ids := make([]string, 0, len(*queries))
	for key, _ := range *queries {
		ids = append(ids, key)
	}
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	var c *mgo.Collection
	if twoPhase {
		c = sessionCopy.DB(DB_NAME).C(COLLECTION_MAPPING)
	} else {
		c = sessionCopy.DB(DB_NAME).C(COLLECTION_RECORD)
	}
	q := c.Find(bson.M{"_id": bson.M{"$in": ids}})
	var result []record.Record
	err := q.Select(bson.M{"_id": 1}).All(&result)
	for _, r := range result {
		_, found := (*queries)[r.Id]
		if found {
			(*queries)[r.Id] = true
		}
	}
	return err
}
