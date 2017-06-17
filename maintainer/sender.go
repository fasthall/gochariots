package maintainer

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
)

var lastSendLId int
var lastSentTOId int
var connMutex sync.Mutex
var remoteBatchersConn []net.Conn
var remoteBatchers []string

func dialRemoteBatchers(dc int) error {
	var err error
	remoteBatchersConn[dc], err = net.Dial("tcp", remoteBatchers[dc])
	return err
}

// Propagate sends the local record to remote datacenter's batcher
func Propagate(r record.Record) {
	for dc, host := range remoteBatchers {
		if dc != info.ID && host != "" {
			// log.Printf("%s is propagating record to remoteBatchers[%d] %s", info.GetName(), dc, host)
			jsonBytes, err := record.ToJSON(r)
			if err != nil {
				logrus.WithField("record", r).Error("couldn't convert record to bytes")
				panic(err)
			}
			b := make([]byte, 5)
			b[4] = byte('r')
			binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))

			connMutex.Lock()
			if remoteBatchersConn[dc] == nil {
				err = dialRemoteBatchers(dc)
				if err != nil {
					logrus.WithField("id", dc).Error("couldn't connect to batcher")
				} else {
					logrus.WithField("id", dc).Info("connected to batcher")
				}
			}
			connMutex.Unlock()
			cnt := 5
			sent := false
			for sent == false {
				connMutex.Lock()
				if remoteBatchersConn[dc] != nil {
					_, err = remoteBatchersConn[dc].Write(append(b, jsonBytes...))
				} else {
					err = errors.New("batcherConn[hostID] == nil")
				}
				connMutex.Unlock()
				if err != nil {
					if cnt >= 0 {
						cnt--
						err = dialRemoteBatchers(dc)
						if err != nil {
							logrus.WithField("attempt", cnt).Warning("couldn't connect to batcher, retrying...")
						} else {
							logrus.WithField("id", dc).Info("connected to batcher")
						}
					} else {
						logrus.WithField("id", dc).Error("failed to connect to batcher after retrying 5 times")
						break
					}
				} else {
					sent = true
					logrus.WithField("id", dc).Info("sucessfully propagated")
				}
			}
		}
	}
}
