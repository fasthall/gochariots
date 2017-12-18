package maintainer

import (
	"encoding/binary"
	"encoding/json"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer/adapter/mongodb"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

var xxx = 0
var mu sync.Mutex

func (s *Server) TOIDReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.TOIDRecord, len(in.GetRecords()))
	propagated := []record.TOIDRecord{}
	for i, ri := range in.GetRecords() {
		records[i] = record.TOIDRecord{
			Id:        ri.GetId(),
			Timestamp: ri.GetTimestamp(),
			Host:      ri.GetHost(),
			TOId:      ri.GetToid(),
			LId:       ri.GetLid(),
			Tags:      ri.GetTags(),
			Pre: record.TOIDCausality{
				Host: ri.GetCausality().GetHost(),
				TOId: ri.GetCausality().GetToid(),
			},
		}
		if records[i].Host == uint32(info.ID) {
			propagated = append(propagated, records[i])
		}
	}
	go func() {
		err := mongodb.PutTOIDRecords(records)
		if err != nil {
			logrus.WithError(err).Error("couldn't put records to mongodb")
		} else {
			benchmark.Logging(len(records))
		}
	}()
	go TOIDPropagate(propagated)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) TOIDUpdateBatchers(ctx context.Context, in *RPCBatchers) (*RPCReply, error) {
	return s.UpdateBatchers(ctx, in)
}

func (s *Server) TOIDReadByLId(ctx context.Context, in *RPCLId) (*RPCReply, error) {
	lid := in.GetLid()
	r, err := TOIDReadByLId(lid)
	if err != nil {
		return nil, err
	}
	j, err := json.Marshal(r)
	return &RPCReply{Message: string(j)}, nil
}

// // Append appends a new record to the maintainer.
// func TOIDAppend(r record.TOIDRecord) error {
// 	// info.LogTimestamp("Append")
// 	r.Timestamp = time.Now().UnixNano()

// 	if maintainerInterface == adapter.DYNAMODB {
// 		err := dynamodb.PutTOIDRecord(r)
// 		if err != nil {
// 			return err
// 		}
// 	} else if maintainerInterface == adapter.DATASTORE {
// 		err := datastore.PutTOIDRecord(r)
// 		if err != nil {
// 			return err
// 		}
// 	} else if maintainerInterface == adapter.FLSTORE {
// 		b, err := record.TOIDToJSON(r)
// 		if err != nil {
// 			return err
// 		}
// 		lenbuf := make([]byte, 4)
// 		binary.BigEndian.PutUint32(lenbuf, uint32(len(b)))
// 		lid := r.LId
// 		_, err = f.WriteAt(append(lenbuf, b...), int64(512*lid))
// 		if err != nil {
// 			return err
// 		}
// 		// log.Println(info.GetName(), "wrote record ", lid)
// 	} else if maintainerInterface == adapter.COSMOSDB {
// 		err := cosmos.PutTOIDRecord(r)
// 		if err != nil {
// 			return err
// 		}
// 	} else if maintainerInterface == adapter.MONGODB {
// 		err := mongodb.PutTOIDRecord(r)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	// log.Println(info.GetName(), "wrote record ", lid)

// 	if r.Host == uint32(info.ID) {
// 		TOIDPropagate(r)
// 	}
// 	return nil
// }

// ReadByLId reads from the maintainer according to LId.
func TOIDReadByLId(lid uint32) (record.TOIDRecord, error) {
	lenbuf := make([]byte, 4)
	_, err := f.ReadAt(lenbuf, int64(512*lid))
	if err != nil {
		return record.TOIDRecord{}, err
	}
	length := int(binary.BigEndian.Uint32(lenbuf))
	buf := make([]byte, length)
	_, err = f.ReadAt(buf, int64(512*lid+4))
	if err != nil {
		return record.TOIDRecord{}, err
	}
	var r record.TOIDRecord
	err = record.TOIDJSONToRecord(buf, &r)
	return r, err
}

// ReadByLIds reads multiple records
func TOIDReadByLIds(lids []uint32) ([]record.TOIDRecord, error) {
	result := []record.TOIDRecord{}
	for _, lid := range lids {
		r, err := TOIDReadByLId(lid)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
