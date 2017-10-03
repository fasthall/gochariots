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
		Lid:     entry.LId,
		Toid:    entry.TOId,
		Host:    entry.Host}, nil
}

func (s *Server) TOIDInsertTags(ctx context.Context, in *RPCTOIdTags) (*RPCReply, error) {
	entry := TOIDIndexTableEntry{
		LId:  in.GetLid(),
		TOId: in.GetToid(),
		Host: in.GetHost(),
	}
	indexMutex.Lock()
	TOIDindexes[in.GetId()] = entry
	indexMutex.Unlock()
	return &RPCReply{Message: "ok"}, nil
}

type TOIDIndexTableEntry struct {
	LId  uint32
	Host uint32
	TOId uint32
}

func TOIDInitIndexer(p string) {

}
