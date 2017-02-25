package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"strconv"

	"github.com/fasthall/gochariots/app"
	"github.com/fasthall/gochariots/batcher"
	"github.com/fasthall/gochariots/filter"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
	"github.com/fasthall/gochariots/queue"
)

func main() {
	fmt.Println(os.Getpid())
	initChariots(1, 0)

	if len(os.Args) >= 3 {
		switch os.Args[1] {
		case "app":
			info.Name = "App" + os.Args[2]
			writePID()
			app.Run(os.Args[2])
			break
		case "batcher":
			info.Name = "Bather" + os.Args[2]
			writePID()
			batcher.InitBatcher(1)
			batcher.SetFilterHost(0, "localhost:9010")
			ln, err := net.Listen("tcp", ":"+os.Args[2])
			if err != nil {
				panic(err)
			}
			defer ln.Close()
			fmt.Println(info.Name+" is listening to port", os.Args[2])
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
		case "filter":
			info.Name = "Filter" + os.Args[2]
			writePID()
			filter.InitFilter(info.NumDC)
			ln, err := net.Listen("tcp", ":"+os.Args[2])
			if err != nil {
				panic(err)
			}
			defer ln.Close()
			fmt.Println(info.Name+" is listening to port", os.Args[2])
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					panic(err)
				}
				// Handle connections in a new goroutine.
				go filter.HandleRequest(conn)
			}
		case "queue":
			info.Name = "Queue" + os.Args[2]
			writePID()
			queue.InitQueue(os.Args[3] == "true")
			queue.SetLogMaintainer("localhost:9030")
			ln, err := net.Listen("tcp", ":"+os.Args[2])
			if err != nil {
				panic(err)
			}
			defer ln.Close()
			fmt.Println(info.Name+" is listening to port", os.Args[2])
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					panic(err)
				}
				// Handle connections in a new goroutine.
				go queue.HandleRequest(conn)
			}
		case "controller":
			info.Name = "LogController" + os.Args[2]
			writePID()
			log.StartController(os.Args[2])
			break
		case "maintainer":
			info.Name = "LogMaintainer" + os.Args[2]
			writePID()
			log.InitLogMaintainer("flstore/")
			ln, err := net.Listen("tcp", ":"+os.Args[2])
			if err != nil {
				panic(err)
			}
			defer ln.Close()
			fmt.Println(info.Name+" is listening to port", os.Args[2])
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					panic(err)
				}
				// Handle connections in a new goroutine.
				go log.HandleRequest(conn)
			}
		}
	}
}

func initChariots(numDc int, id int) {
	info.NumDC = numDc
	info.ID = id
}

func writePID() {
	ioutil.WriteFile(info.Name+".pid", []byte(strconv.Itoa(os.Getpid())+"\n"), 0644)
}
