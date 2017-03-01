package main

import (
	"fmt"
	"net"
	"os"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/queue"
)

func main() {
	fmt.Println(os.Getpid())
	info.InitChariots(1, 0)

	if len(os.Args) < 3 {
		fmt.Println("Usage: gochariots-queue port token(true|false)")
		return
	}

	info.SetName("queue" + os.Args[1])
	info.WritePID()
	queue.InitQueue(os.Args[2] == "true")
	queue.SetLogMaintainer("localhost:9030")
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
		go queue.HandleRequest(conn)
	}
}
