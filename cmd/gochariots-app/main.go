package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fasthall/gochariots/app"
	"github.com/fasthall/gochariots/info"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: gochariots-app port num_dc dc_id")
		return
	}
	numDc, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Usage: gochariots-app port num_dc dc_id")
		return
	}
	dcID, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Usage: gochariots-app port num_dc dc_id")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("app" + os.Args[1])
	info.WritePID()
	app.Run(os.Args[1])
}
