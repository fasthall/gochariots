package main

import (
	"fmt"
	"net"
	"os"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
)

func main() {
	fmt.Println(os.Getpid())
	info.InitChariots(1, 0)

	if len(os.Args) < 2 {
		fmt.Println("Usage: gochariots-maintainer port")
		return
	}

	info.SetName("maintainer" + os.Args[1])
	info.WritePID()
	log.InitLogMaintainer("flstore/")
	ln, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println(info.GetName()+" is listening to port", os.Args[1])
	for {
		// Listen for an incoming connection.
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		// Handle connections in a new goroutine.
		go log.HandleRequest(conn)
	}
}
