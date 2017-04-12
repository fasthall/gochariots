package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/queue"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: gochariots-queue port num_dc dc_id token(true|false)")
		return
	}
	numDc, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Usage: gochariots-queue port num_dc dc_id token(true|false)")
		return
	}
	dcID, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Usage: gochariots-queue port num_dc dc_id token(true|false)")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("queue" + os.Args[1])
	info.RedirectLog(info.GetName() + ".log")
	queue.InitQueue(os.Args[4] == "true")
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
