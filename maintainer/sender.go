package maintainer

import (
	"fmt"
	"net"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/record"
)

var lastSendLId int
var lastSentTOId int

func Propagate(r record.Record) {
	for dc, host := range remoteBatchers {
		fmt.Println(dc, info.ID, host)
		if dc != info.ID && host != "" {
			b := []byte{'r'}
			jsonBytes, err := record.ToJSON(r)
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
		fmt.Println(info.GetName(), "propagates to", host, r)
	}
}
