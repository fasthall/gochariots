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
var remoteBatcherVer int

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
				logrus.WithError(err).Error("couldn't convert record to bytes")
				panic(err)
			}
			b := make([]byte, 5)
			b[4] = byte('r')
			binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))

			connMutex.Lock()
			if remoteBatchersConn[dc] == nil {
				err = dialRemoteBatchers(dc)
				if err != nil {
					logrus.WithField("batcher", host).Error("couldn't connect to remoteBatcher")
					panic(err)
				} else {
					logrus.WithField("batcher", host).Info("connected to remoteBatcher")
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
							logrus.WithField("attempt", cnt).Warning("couldn't connect to remote batcher, retrying...")
						} else {
							logrus.WithField("batcher", host).Info("connected to remote batcher")
						}
					} else {
						logrus.WithField("batcher", host).Error("failed to connect to the remote batcher after retrying 5 times")
						break
					}
				} else {
					sent = true
					// log.Println(string(append(b, append(jsonBytes, byte('\n'))...)))
					// log.Println(info.GetName(), "propagates to", host, r)
				}

				if err != nil {
					if cnt == 0 {
						logrus.WithField("batcher", host).Error("failed to connect to the remote batcher after retrying 5 times")
						break
					}
					logrus.WithField("attempt", cnt).Error("couldn't send to remote batcher")
					cnt--
					continue
				}
			}
			logrus.WithFields(logrus.Fields{"batcher": host, "record": r}).Debug("sent to remote batcher")
		}
	}
}
