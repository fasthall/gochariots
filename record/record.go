// Package record provides the functions of log mainteiners in Chariots.
// It will be implemented using FLStore, but it's currently a mocked in-memory log for prototyping.
package record

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"time"
)

// Record represents a single record in the shared log.
// 	Host: The datacenter where the record origins
// 	LId: The record number in this datacenter
//	TOId: The total order of the record in the origining datacenter
//	Tags: Consist of keys and values
type Record struct {
	Timestamp time.Time
	Host      int
	TOId      int
	LId       int
	Tags      map[string]string
	Pre       Causality
}

// Causality structure is used in Record structure. It shows the record which should present before.
type Causality struct {
	Host int
	TOId int
	Tags map[string]string
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
