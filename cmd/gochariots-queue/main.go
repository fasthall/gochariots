package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/queue"
)

func main() {
	v := flag.Bool("v", false, "Turn on all logging")
	flag.Parse()
	if len(flag.Args()) < 4 {
		fmt.Println("Usage: gochariots-queue port num_dc dc_id token(true|false)")
		return
	}
	numDc, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Println("Usage: gochariots-queue port num_dc dc_id token(true|false)")
		return
	}
	dcID, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		fmt.Println("Usage: gochariots-queue port num_dc dc_id token(true|false)")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("queue" + flag.Arg(0))
	info.RedirectLog(info.GetName()+".log", *v)
	queue.InitQueue(flag.Arg(3) == "true")
	ln, err := net.Listen("tcp", ":"+flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println(info.GetName()+" is listening to port", flag.Arg(0))
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
