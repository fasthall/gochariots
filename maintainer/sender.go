package maintainer

import (
	"log"
	"net"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
)

var lastSendLId int
var lastSentTOId int

// Propagate sends the local record to remote datacenter's batcher
func Propagate(r record.Record) {
	for dc, host := range remoteBatchers {
		log.Printf("%s is propagatin record to remoteBatchers[%d] %s", info.GetName(), dc, host)
		if dc != info.ID && host != "" {
			b := []byte{'r'}
			jsonBytes, err := record.ToJSON(r)
			if err != nil {
				log.Println(info.GetName(), "couldn't convert record to bytes:", r)
				log.Panicln(err)
			}

			cnt := 5
			sent := false
			for sent == false {
				conn, err := net.Dial("tcp", host)
				if err != nil {
					if cnt == 0 {
						log.Printf("%s failed to propagate to remoteBatchers[%d] %s after retrying 5 times", info.GetName(), dc, host)
						break
					}
					log.Printf("%s couldn't connect to remoteBatchers[%d] %s\n", info.GetName(), dc, host)
					cnt--
					continue
				}
				defer conn.Close()
				_, err = conn.Write(append(b, jsonBytes...))
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
		log.Println(info.GetName(), "propagates to", host, r)
	}
}
