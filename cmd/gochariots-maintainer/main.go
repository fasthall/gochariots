package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
)

func main() {
	n := flag.Int("n", 0, "Log the time maintainer takes to append the nth record")
	flag.Parse()
	maintainer.LogRecordNth = *n
	fmt.Println(maintainer.LogRecordNth)
	if len(flag.Args()) < 3 {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	numDc, err := strconv.Atoi(flag.Args()[1])
	if err != nil {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	dcID, err := strconv.Atoi(flag.Args()[2])
	if err != nil {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("maintainer" + flag.Args()[0])
	info.RedirectLog(info.GetName() + ".log")
	maintainer.InitLogMaintainer(info.GetName())
	ln, err := net.Listen("tcp", ":"+flag.Args()[0])
	if err != nil {
		fmt.Println(info.GetName() + "couldn't listen on port " + flag.Args()[0])
		panic(err)
	}
	defer ln.Close()
	fmt.Println(info.GetName()+" is listening to port", flag.Args()[0])
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
