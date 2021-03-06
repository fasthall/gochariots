package record

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// Record represents a single record in the shared log.
// 	Host: The datacenter where the record origins
// 	LId: The record number in this datacenter
//	TOId: The total order of the record in the origining datacenter
//	Tags: Consist of keys and values
type Record struct {
	ID        string `bson:"_id,omitempty"`
	LID       uint32
	Parent    string
	Timestamp int64
	Host      uint32
	Tags      map[string]string
	Trace     string
}

// ToJSON encodes a record into bytes
func ToJSON(r Record) ([]byte, error) {
	return json.Marshal(r)
}

// ToJSONArray encodes slice of records into JSON array
func ToJSONArray(records []Record) ([]byte, error) {
	return json.Marshal(records)
}

// JSONToRecord decodes bytes into record
func JSONToRecord(b []byte, r *Record) error {
	return json.Unmarshal(b, &r)
}

// JSONToRecordArray decodes json bytes into slice of records
func JSONToRecordArray(b []byte, records *[]Record) error {
	return json.Unmarshal(b, &records)
}

// ToGob encodes a record into gob bytes
func ToGob(r Record) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(r)
	return buf.Bytes(), err
}

// GobToRecord decodes gob bytes into record
func GobToRecord(b []byte, r *Record) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return dec.Decode(&r)
}

// ToGob encodes a record into gob bytes
func ToGobArray(records []Record) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(records)
	return buf.Bytes(), err
}

// GobToRecord decodes gob bytes into record
func GobToRecordArray(b []byte, r *[]Record) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return dec.Decode(&r)
}
