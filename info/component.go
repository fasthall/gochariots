package info

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/misc"
)

// Name is the name of this running component
var name string
var port string

func SetName(n string) {
	name = n
}

func GetName() string {
	return name
}

func SetPort(p string) {
	port = p
}

func GetPort() string {
	return port
}

func RedirectLog(name string, level logrus.Level) {
	// logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	err := os.Mkdir("logs", 0755)
	filepath := path.Join("logs", name)
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Error opening log file %s\n", filepath)
		fmt.Println(err)
		logrus.SetOutput(os.Stdout)
	} else {
		logrus.SetOutput(f)
	}
}

func Config(file, component string) {
	config, err := misc.ReadConfig(file)
	if err != nil {
		logrus.WithError(err).Warn("read config file failed")
		return
	}
	if config.Controller == "" {
		logrus.Error("No controller information found in config file")
		return
	}
	addr, err := misc.GetHostIP()
	if err != nil {
		logrus.WithError(err).Error("couldn't find local IP address")
		return
	}
	p := misc.NewParams()
	p.AddParam("host", addr+":"+GetPort())
	logrus.WithFields(logrus.Fields{"controller": config.Controller}).Info("Config file read")

	err = errors.New("")
	for err != nil {
		time.Sleep(1 * time.Second)
		code := http.StatusBadRequest
		response := ""
		for code != http.StatusOK {
			code, response, err = misc.Report(config.Controller, component, p)
			if err != nil {
				logrus.WithError(err).Error("couldn't report to the controller")
			}
		}
		logrus.WithField("response", response).Info("reported to controller")
	}
}

func LogTimestamp(action string) {
	log.Println(action, time.Now())
}
