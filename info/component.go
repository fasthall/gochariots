package info

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

// Name is the name of this running component
var name string

func SetName(n string) {
	name = n
}

func GetName() string {
	return name
}

func RedirectLog(name string) {
	filepath := path.Join("logs", name)
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file %s\n", name)
	} else {
		log.SetOutput(f)
	}
}

func LogTimestamp(action string) {
	log.Println(action, time.Now())
}
