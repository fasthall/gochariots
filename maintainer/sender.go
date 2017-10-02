package maintainer

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer/remotebatcher"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

var lastSendLId int
var lastSentTOId int
var connMutex sync.Mutex
var remoteBatchersClient []remotebatcher.BatcherClient
var remoteBatchers []string
var remoteBatcherVer int

// Propagate sends the local record to remote datacenter's batcher
func Propagate(r record.Record) {
	for dc, host := range remoteBatchers {
		if dc != info.ID && host != "" {
			// log.Printf("%s is propagating record to remoteBatchers[%d] %s", info.GetName(), dc, host)
			rpcRecord := remotebatcher.RPCRecord{
				Timestamp: r.Timestamp,
				Host:      int32(r.Host),
				Lid:       int32(r.LId),
				Tags:      r.Tags,
				Hash:      r.Hash,
				Seed:      r.Seed,
			}
			remoteBatchersClient[dc].ReceiveRecord(context.Background(), &rpcRecord)
			logrus.WithFields(logrus.Fields{"batcher": host, "record": r}).Debug("sent to remote batcher")
		}
	}
}
