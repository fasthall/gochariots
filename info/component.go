package info

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"net/http"

	"strconv"

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
		fmt.Println("Read config file failed")
		logrus.WithError(err).Warn("read config file failed")
		return
	}
	if config.Controller == "" {
		fmt.Println("No controller information found in config file")
		logrus.Error("no controller information found in config file")
		return
	}
	if config.NumDC != "" {
		fmt.Println("Found num_dc in config file")
		logrus.Info("found num_dc in config file")
		NumDC, err = strconv.Atoi(config.NumDC)
		if err != nil {
			fmt.Println("Read config file failed")
			logrus.WithError(err).Warn("read config file failed")
			return
		}
	}
	if config.ID != "" {
		fmt.Println("Found id in config file")
		logrus.Info("found id in config file")
		ID, err = strconv.Atoi(config.ID)
		if err != nil {
			fmt.Println("Read config file failed")
			logrus.WithError(err).Warn("read config file failed")
			return
		}
	}
	addr, err := misc.GetHostIP()
	if err != nil {
		fmt.Println("Couldn't find local IP address")
		logrus.WithError(err).Error("couldn't find local IP address")
		return
	}
	p := misc.NewParams()
	p.AddParam("host", addr+":"+GetPort())
	fmt.Println("Config file read")
	logrus.WithFields(logrus.Fields{"controller": config.Controller}).Info("config file read")

	go report(config.Controller, component, p)
}

func report(host, path string, p misc.Params) {
	// wait for a second to avoid the component is not ready which causes controller not able to connect to it
	// it should had been handled in controller
	time.Sleep(1 * time.Second)

	err := errors.New("")
	for err != nil {
		code := http.StatusBadRequest
		response := ""
		for code != http.StatusOK {
			code, response, err = misc.Report(host, path, p)
			if err != nil {
				time.Sleep(1 * time.Second)
				fmt.Println("Couldn't report to the controller")
				logrus.WithError(err).Error("couldn't report to the controller")
			}
		}
		fmt.Println("Reported to controller")
		logrus.WithField("response", response).Info("reported to controller")
	}
}

func LogTimestamp(action string) {
	log.Println(action, time.Now())
}
