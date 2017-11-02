package maintainer

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

var lastSendLId uint32
var lastSentTOId uint32
var connMutex sync.Mutex
var remoteBatchersClient []batcherrpc.BatcherRPCClient
var remoteBatchers []string
var remoteBatcherVer uint32

// Propagate sends the local record to remote datacenter's batcher
func Propagate(rs []record.Record) {
	for dc, host := range remoteBatchers {
		if dc != info.ID && host != "" {
			rpcRecords := batcherrpc.RPCRecords{
				Records: make([]*batcherrpc.RPCRecord, len(rs)),
			}
			for i, r := range rs {
				rpcRecords.Records[i] = &batcherrpc.RPCRecord{
					Timestamp: r.Timestamp,
					Host:      r.Host,
					Lid:       r.LId,
					Tags:      r.Tags,
					Parent:    r.Parent,
					Seed:      r.Seed,
				}
			}
			// log.Printf("%s is propagating record to remoteBatchers[%d] %s", info.GetName(), dc, host)
			_, err := remoteBatchersClient[dc].ReceiveRecords(context.Background(), &rpcRecords)
			if err != nil {
				logrus.WithFields(logrus.Fields{"batcher": host, "error": err}).Error("couldn't send to remote batcher")
			} else {
				logrus.WithFields(logrus.Fields{"batcher": host, "records": len(rs)}).Debug("sent to remote batcher")
			}
		}
	}
}
