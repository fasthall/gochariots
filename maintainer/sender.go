package maintainer

import (
	"fmt"
	"net"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
)

var lastSendLId int
var lastSentTOId int

func Propagate(record log.Record) {
	for dc, host := range remoteBatchers {
		fmt.Println(dc, info.ID, host)
		if dc != info.ID && host != "" {
			b := []byte{'r'}
			jsonBytes, err := log.ToJSON(record)
			if err != nil {
				panic(err)
			}

			conn, err := net.Dial("tcp", host)
			if err != nil {
				fmt.Println("Couldn't connect to", host)
			}
			defer conn.Close()
			conn.Write(append(b, jsonBytes...))
		}
		fmt.Println(info.GetName(), "propagates to", host, record)
	}
}
