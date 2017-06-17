package main

import (
	"flag"
	"fmt"
	"net"

	"strconv"

	"github.com/fasthall/gochariots/batcher"
	"github.com/fasthall/gochariots/info"
)

func main() {
	v := flag.Bool("v", false, "Turn on all logging")
	flag.Parse()
	if len(flag.Args()) < 3 {
		fmt.Println("Usage: gochariots-batcher port num_dc dc_id")
		return
	}
	numDc, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Println("Usage: gochariots-batcher port num_dc dc_id")
		return
	}
	dcID, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		fmt.Println("Usage: gochariots-batcher port num_dc dc_id")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("batcher" + flag.Arg(0))
	info.RedirectLog(info.GetName()+".log", *v)
	batcher.InitBatcher(numDc)
	ln, err := net.Listen("tcp", ":"+flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println(info.GetName()+" is listening to port", flag.Arg(0))
	go batcher.Sweeper()
	for {
		// Listen for an incoming connection.
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(info.GetName(), "Couldn't handle more connection", err)
			continue
		}
		// Handle connections in a new goroutine.
		go batcher.HandleRequest(conn)
	}
}
