package maintainer

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sync"

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
			log.Printf("%s is propagating record to remoteBatchers[%d] %s", info.GetName(), dc, host)
			jsonBytes, err := record.ToJSON(r)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert record to bytes:", r)
				log.Panicln(err)
			}
			b := make([]byte, 5)
			b[4] = byte('r')
			binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))

			connMutex.Lock()
			if remoteBatchersConn[dc] == nil {
				err = dialRemoteBatchers(dc)
				if err != nil {
					log.Printf("%s couldn't connect to remoteBatchers[%d] %s\n", info.GetName(), dc, host)
				} else {
					log.Printf("%s is connected to remoteBatchers[%d] %s\n", info.GetName(), dc, host)
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
							log.Printf("%s couldn't connect to remoteBatchers[%d] %s, retrying\n", info.GetName(), dc, host)
						}
					} else {
						log.Printf("%s failed to connect to remoteBatchers[%d] %s after retrying 5 times\n", info.GetName(), dc, host)
						break
					}
				} else {
					sent = true
					log.Println(string(append(b, append(jsonBytes, byte('\n'))...)))
					log.Println(info.GetName(), "propagates to", host, r)
				}

				if err != nil {
					if cnt == 0 {
						log.Printf("%s failed to propagate to remoteBatchers[%d] %s after retrying 5 times", info.GetName(), dc, host)
						break
					}
					log.Printf("%s couldn't send to remoteBatchers[%d] %s\n", info.GetName(), dc, host)
					cnt--
					continue
				}
			}
		}
	}
}
