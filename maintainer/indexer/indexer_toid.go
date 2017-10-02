package indexer

import (
	"sync"

	"golang.org/x/net/context"
)

var indexMutex sync.Mutex
var TOIDindexes = make(map[uint64]TOIDIndexTableEntry)

func (s *Server) TOIDQuery(ctx context.Context, in *RPCTOIdQuery) (*RPCTOIDQueryReply, error) {
	id := in.GetId()
	indexMutex.Lock()
	entry, ok := TOIDindexes[id]
	indexMutex.Unlock()
	return &RPCTOIDQueryReply{
		Existed: ok,
		Lid:     int32(entry.LId),
		Toid:    int32(entry.TOId),
		Host:    int32(entry.Host)}, nil
}

func (s *Server) TOIDInsertTags(ctx context.Context, in *RPCTOIdTags) (*RPCReply, error) {
	id := in.GetId()
	lid := int(in.GetLid())
	toid := int(in.GetToid())
	host := int(in.GetHost())
	entry := TOIDIndexTableEntry{
		LId:  lid,
		TOId: toid,
		Host: host,
	}
	indexMutex.Lock()
	TOIDindexes[id] = entry
	indexMutex.Unlock()
	return &RPCReply{Message: "ok"}, nil
}

type TOIDIndexTableEntry struct {
	LId  int
	Host int
	TOId int
}

func TOIDInitIndexer(p string) {

}
