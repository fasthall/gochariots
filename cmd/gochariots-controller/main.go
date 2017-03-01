package main

import (
	"fmt"
	"os"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
)

func main() {
	fmt.Println(os.Getpid())
	info.InitChariots(1, 0)

	if len(os.Args) < 2 {
		fmt.Println("Usage: gochariots-controller port")
		return
	}

	info.SetName("controller" + os.Args[1])
	info.WritePID()
	log.StartController(os.Args[1])
}
