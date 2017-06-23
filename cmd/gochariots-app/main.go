package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/fasthall/gochariots/app"
	"github.com/fasthall/gochariots/info"
)

func main() {
	v := flag.Bool("v", false, "Turn on all logging")
	toid := flag.Bool("toid", false, "TOId version")
	flag.Parse()
	if len(flag.Args()) < 3 {
		fmt.Println("Usage: gochariots-app port num_dc dc_id")
		return
	}
	numDc, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Println("Usage: gochariots-app port num_dc dc_id")
		return
	}
	dcID, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		fmt.Println("Usage: gochariots-app port num_dc dc_id")
		return
	}
	info.InitChariots(numDc, dcID)
	info.SetName("app" + flag.Arg(0))
	info.RedirectLog(info.GetName()+".log", *v)
	if *toid {
		app.TOIDRun(flag.Arg(0))
	} else {
		app.Run(flag.Arg(0))
	}
}
