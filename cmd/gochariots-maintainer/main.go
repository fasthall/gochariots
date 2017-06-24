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
	v := flag.Bool("v", false, "Turn on all logging")
	toid := flag.Bool("toid", false, "TOId version")
	flag.Parse()
	if len(flag.Args()) < 3 {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	numDc, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	dcID, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		fmt.Println("Usage: gochariots-maintainer port num_dc dc_id")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("maintainer" + flag.Arg(0))
	info.RedirectLog(info.GetName()+".log", *v)
	maintainer.InitLogMaintainer(info.GetName(), *n)
	ln, err := net.Listen("tcp", ":"+flag.Arg(0))
	if err != nil {
		fmt.Println(info.GetName() + "couldn't listen on port " + flag.Arg(0))
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
			go maintainer.TOIDHandleRequest(conn)
		}
	} else {
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
}
