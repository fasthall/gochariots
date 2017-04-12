package maintainer

import (
	"log"
	"net"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
)

var lastSendLId int
var lastSentTOId int
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
			log.Printf("%s is propagatin record to remoteBatchers[%d] %s", info.GetName(), dc, host)
			b := []byte{'r'}
			jsonBytes, err := record.ToJSON(r)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert record to bytes:", r)
				log.Panicln(err)
			}

			if remoteBatchersConn[dc] == nil {
				err = dialRemoteBatchers(dc)
				if err != nil {
					log.Printf("%s couldn't connect to remoteBatchers[%d] %s\n", info.GetName(), dc, host)
				} else {
					log.Printf("%s is connected to remoteBatchers[%d] %s\n", info.GetName(), dc, host)
				}
			}
			cnt := 5
			sent := false
			for sent == false {
				_, err = remoteBatchersConn[dc].Write(append(b, jsonBytes...))
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
