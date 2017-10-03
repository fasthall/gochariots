package maintainer

import (
	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer/remotebatcher"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

// Propagate sends the local record to remote datacenter's batcher
func TOIDPropagate(r record.TOIDRecord) {
	for dc, host := range remoteBatchers {
		if dc != info.ID && host != "" {
			// log.Printf("%s is propagating record to remoteBatchers[%d] %s", info.GetName(), dc, host)
			rpcRecord := remotebatcher.RPCRecord{
				Id:        r.Id,
				Timestamp: r.Timestamp,
				Host:      r.Host,
				Toid:      r.TOId,
				Lid:       r.LId,
				Tags:      r.Tags,
				Causality: &remotebatcher.RPCCausality{
					Host: r.Pre.Host,
					Toid: r.Pre.TOId,
				},
			}
			remoteBatchersClient[dc].TOIDReceiveRecord(context.Background(), &rpcRecord)
			logrus.WithFields(logrus.Fields{"batcher": host, "record": r}).Debug("sent to remote batcher")
		}
	}
}
