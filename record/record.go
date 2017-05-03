// Package record provides the functions of log mainteiners in Chariots.
// It will be implemented using FLStore, but it's currently a mocked in-memory log for prototyping.
package record

import (
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
func ToJSON(record Record) ([]byte, error) {
	return json.Marshal(record)
}

// ToRecord decodes bytes into record
func ToRecord(b []byte, r *Record) error {
	err := json.Unmarshal(b, &r)
	return err
}

// ToJSONArray encodes slice of records into JSON array
func ToJSONArray(records []Record) ([]byte, error) {
	return json.Marshal(records)
}

// ToRecordArray decodes json bytes into slice of records
func ToRecordArray(b []byte) ([]Record, error) {
	var records []Record
	err := json.Unmarshal(b, &records)
	return records, err
}
