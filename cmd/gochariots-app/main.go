package main

import (
	"fmt"
	"os"

	"github.com/fasthall/gochariots/app"
	"github.com/fasthall/gochariots/info"
)

func main() {
	fmt.Println(os.Getpid())
	info.InitChariots(1, 0)

	if len(os.Args) < 2 {
		fmt.Println("Usage: gochariots-app port")
		return
	}

	info.SetName("app" + os.Args[1])
	info.WritePID()
	app.Run(os.Args[1])
}
