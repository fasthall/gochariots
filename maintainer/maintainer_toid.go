package maintainer

import (
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer/indexer"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

func (s *Server) TOIDReceiveRecords(ctx context.Context, in *RPCRecords) (*RPCReply, error) {
	records := make([]record.TOIDRecord, len(in.GetRecords()))
	for i, ri := range in.GetRecords() {
		records[i] = record.TOIDRecord{
			ID:        ri.GetId(),
			Timestamp: ri.GetTimestamp(),
			Host:      int(ri.GetHost()),
			TOId:      int(ri.GetToid()),
			LId:       int(ri.GetLid()),
			Tags:      ri.GetTags(),
			Pre: record.TOIDCausality{
				Host: int(ri.GetCausality().GetHost()),
				TOId: int(ri.GetCausality().GetToid()),
			},
		}
	}
	TOIDrecordsArrival(records)
	return &RPCReply{Message: "ok"}, nil
}

func (s *Server) TOIDUpdateBatchers(ctx context.Context, in *RPCBatchers) (*RPCReply, error) {
	return s.UpdateBatchers(ctx, in)
}

func (s *Server) TOIDUpdateIndexer(ctx context.Context, in *RPCIndexer) (*RPCReply, error) {
	return s.UpdateIndexer(ctx, in)
}

func (s *Server) TOIDReadByLId(ctx context.Context, in *RPCLId) (*RPCReply, error) {
	lid := int(in.GetLid())
	r, err := TOIDReadByLId(lid)
	if err != nil {
		return nil, err
	}
	j, err := json.Marshal(r)
	return &RPCReply{Message: string(j)}, nil
}

// Append appends a new record to the maintainer.
func TOIDAppend(r record.TOIDRecord) error {
	// info.LogTimestamp("Append")
	r.Timestamp = time.Now().UnixNano()
	if logRecordNth > 0 {
		if r.LId == 1 {
			logFirstTime = time.Now()
		} else if r.LId == logRecordNth {
			logrus.WithField("duration", time.Since(logFirstTime)).Info("appended", logRecordNth, "records")
		}
	}
	b, err := record.TOIDToJSON(r)
	if err != nil {
		return err
	}
	lenbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenbuf, uint32(len(b)))
	lid := r.LId
	_, err = f.WriteAt(lenbuf, int64(512*lid))
	if err != nil {
		return err
	}
	_, err = f.WriteAt(b, int64(512*lid+4))
	if err != nil {
		return err
	}
	// log.Println(info.GetName(), "wrote record ", lid)
	TOIDInsertIndexer(r)

	LastLId = r.LId
	if r.Host == info.ID {
		TOIDPropagate(r)
	}
	return nil
}

func TOIDInsertIndexer(r record.TOIDRecord) {
	rpcTags := indexer.RPCTOIdTags{
		Id:   r.ID,
		Lid:  int32(r.LId),
		Toid: int32(r.TOId),
		Host: int32(r.Host),
	}
	_, err := indexerClient.TOIDInsertTags(context.Background(), &rpcTags)
	if err != nil {
		logrus.WithField("host", indexerHost).Error("failed to connect to indexer")
	}
}

// ReadByLId reads from the maintainer according to LId.
func TOIDReadByLId(lid int) (record.TOIDRecord, error) {
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
func TOIDReadByLIds(lids []int) ([]record.TOIDRecord, error) {
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

func TOIDrecordsArrival(records []record.TOIDRecord) {
	// info.LogTimestamp("recordsArrival")
	for _, record := range records {
		err := TOIDAppend(record)
		if err != nil {
			logrus.WithError(err).Error("couldn't append the record to the store")
		}
	}
}
