package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"

	"github.com/fasthall/gochariots/filter"
	"github.com/fasthall/gochariots/info"
)

func main() {
	v := flag.Bool("v", false, "Turn on all logging")
	toid := flag.Bool("toid", false, "TOId version")
	flag.Parse()
	if len(flag.Args()) < 3 {
		fmt.Println("Usage: gochariots-filter port num_dc dc_id")
		return
	}
	numDc, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Println("Usage: gochariots-filter port num_dc dc_id")
		return
	}
	dcID, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		fmt.Println("Usage: gochariots-filter port num_dc dc_id")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("filter" + flag.Arg(0))
	info.RedirectLog(info.GetName()+".log", *v)
	if *toid {
		filter.TOIDInitFilter(info.NumDC)
	} else {
		filter.InitFilter(info.NumDC)
	}
	ln, err := net.Listen("tcp", ":"+flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println(info.GetName()+" is listening to port", flag.Arg(0))
	if *toid {
		for {
			// Listen for an incoming connection.
			conn, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			// Handle connections in a new goroutine.
			go filter.TOIDHandleRequest(conn)
		}
	} else {
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
}
