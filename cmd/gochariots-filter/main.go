package main

import (
	"fmt"
	"net"
	"os"

	"github.com/fasthall/gochariots/filter"
	"github.com/fasthall/gochariots/info"
)

func main() {
	fmt.Println(os.Getpid())
	info.InitChariots(1, 0)

	if len(os.Args) < 2 {
		fmt.Println("Usage: gochariots-filter port")
		return
	}

	info.SetName("filter" + os.Args[1])
	info.WritePID()
	filter.InitFilter(info.NumDC)
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
		go filter.HandleRequest(conn)
	}
}
