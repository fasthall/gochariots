package maintainer

import (
	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
	"golang.org/x/net/context"
)

// Propagate sends the local record to remote datacenter's batcher
func TOIDPropagate(rs []record.TOIDRecord) {
	for dc, host := range remoteBatchers {
		if dc != info.ID && host != "" {
			rpcRecords := batcherrpc.RPCRecords{
				Records: make([]*batcherrpc.RPCRecord, len(rs)),
			}
			for i, r := range rs {
				rpcRecords.Records[i] = &batcherrpc.RPCRecord{
					Timestamp: r.Timestamp,
					Host:      r.Host,
					Toid:      r.TOId,
					Lid:       r.LId,
					Tags:      r.Tags,
					Causality: &batcherrpc.RPCCausality{
						Host: r.Pre.Host,
						Toid: r.Pre.TOId,
					},
				}
			}
			_, err := remoteBatchersClient[dc].TOIDReceiveRecords(context.Background(), &rpcRecords)
			if err != nil {
				logrus.WithError(err).Error("couldn't send to remote batcher")
			} else {
				logrus.WithFields(logrus.Fields{"batcher": host, "records": len(rs)}).Debug("sent to remote batcher")
			}
		}
	}
}
