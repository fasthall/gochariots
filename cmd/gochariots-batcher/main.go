package main

import (
	"fmt"
	"net"
	"os"

	"strconv"

	"github.com/fasthall/gochariots/batcher"
	"github.com/fasthall/gochariots/info"
)

func main() {
	fmt.Println(os.Getpid())
	info.InitChariots(1, 0)

	if len(os.Args) < 3 {
		fmt.Println("Usage: gochariots-batcher port num_datacenters")
		return
	}

	info.SetName("batcher" + os.Args[1])
	info.WritePID()
	n, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Usage: gochariots-batcher port num_datacenters")
		return
	}
	batcher.InitBatcher(n)
	ln, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println(info.GetName()+" is listening to port", os.Args[1])
	go batcher.Sweeper()
	for {
		// Listen for an incoming connection.
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		// Handle connections in a new goroutine.
		go batcher.HandleRequest(conn)
	}
}
