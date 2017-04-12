package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/fasthall/gochariots/filter"
	"github.com/fasthall/gochariots/info"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: gochariots-filter port num_dc dc_id")
		return
	}
	numDc, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Usage: gochariots-filter port num_dc dc_id")
		return
	}
	dcID, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Usage: gochariots-filter port num_dc dc_id")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("filter" + os.Args[1])
	info.RedirectLog(info.GetName() + ".log")
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
