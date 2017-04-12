package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	numDc, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	dcID, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("maintainer" + os.Args[1])
	info.RedirectLog(info.GetName() + ".log")
	maintainer.InitLogMaintainer("flstore/" + info.GetName())
	ln, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		fmt.Println(info.GetName() + "couldn't listen on port " + os.Args[1])
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
		go maintainer.HandleRequest(conn)
	}
}
