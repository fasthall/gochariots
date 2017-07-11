package main

import (
	"fmt"
	"net"
	"os"

	"github.com/fasthall/gochariots/app"
	"github.com/fasthall/gochariots/batcher"
	"github.com/fasthall/gochariots/filter"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/maintainer/indexer"
	"github.com/fasthall/gochariots/queue"

	"github.com/Sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	gochariots = kingpin.New("gochariots", "A distributed shared log system.")

	appCommand = gochariots.Command("app", "Start an app instance.")
	appNumDC   = appCommand.Arg("num_dc", "The port app listens to.").Required().Int()
	appID      = appCommand.Arg("id", "The port app listens to.").Required().Int()
	appPort    = appCommand.Flag("port", "The port app listens to. By default it's 8080").Short('p').String()
	appTOId    = appCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	appConfig  = appCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	appInfo    = appCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	appDebug   = appCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	batcherCommand = gochariots.Command("batcher", "Start a batcher instance.")
	batcherNumDC   = batcherCommand.Arg("num_dc", "The port batcher listens to.").Required().Int()
	batcherID      = batcherCommand.Arg("id", "The port batcher listens to.").Required().Int()
	batcherPort    = batcherCommand.Flag("port", "The port app listens to. By default it's 9000").Short('p').String()
	batcherTOId    = batcherCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	batcherConfig  = batcherCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	batcherInfo    = batcherCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	batcherDebug   = batcherCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	controllerCommand = gochariots.Command("controller", "Start a controller instance.")
	controllerNumDC   = controllerCommand.Arg("num_dc", "The port controller listens to.").Required().Int()
	controllerID      = controllerCommand.Arg("id", "The port controller listens to.").Required().Int()
	controllerPort    = controllerCommand.Flag("port", "The port app listens to. By default it's 8081").Short('p').String()
	controllerInfo    = controllerCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	controllerDebug   = controllerCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	filterCommand = gochariots.Command("filter", "Start a filter instance.")
	filterNumDC   = filterCommand.Arg("num_dc", "The port filter listens to.").Required().Int()
	filterID      = filterCommand.Arg("id", "The port contrfilteroller listens to.").Required().Int()
	filterPort    = filterCommand.Flag("port", "The port app listens to. By default it's 9010").Short('p').String()
	filterTOId    = filterCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	filterConfig  = filterCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	filterInfo    = filterCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	filterDebug   = filterCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	queueCommand = gochariots.Command("queue", "Start a queue instance.")
	queueNumDC   = queueCommand.Arg("num_dc", "The port queue listens to.").Required().Int()
	queueID      = queueCommand.Arg("id", "The port queue listens to.").Required().Int()
	queuePort    = queueCommand.Flag("port", "The port app listens to. By default it's 9020").Short('p').String()
	queueHold    = queueCommand.Flag("hold", "Whether this queue instance holds a token when launched.").Required().Short('h').Bool()
	queueTOId    = queueCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	queueCarry   = queueCommand.Flag("carry", "Carry deferred records with token. Only work when toid is on.").Short('c').Bool()
	queueConfig  = queueCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	queueInfo    = queueCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	queueDebug   = queueCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	maintainerCommand = gochariots.Command("maintainer", "Start a maintainer instance.")
	maintainerNumDC   = maintainerCommand.Arg("num_dc", "The port maintainer listens to.").Required().Int()
	maintainerID      = maintainerCommand.Arg("id", "The port maintainer listens to.").Required().Int()
	maintainerInstN   = maintainerCommand.Arg("ntime", "Record the time maintainer takes to append n records").Int()
	maintainerPort    = maintainerCommand.Flag("port", "The port app listens to. By default it's 9030").Short('p').String()
	maintainerTOId    = maintainerCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	maintainerConfig  = maintainerCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	maintainerInfo    = maintainerCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	maintainerDebug   = maintainerCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	indexerCommand = gochariots.Command("indexer", "Start an indexer instance.")
	indexerNumDC   = indexerCommand.Arg("num_dc", "The port indexer listens to.").Required().Int()
	indexerID      = indexerCommand.Arg("id", "The port indexer listens to.").Required().Int()
	indexerPort    = indexerCommand.Flag("port", "The port app listens to. By default it's 9040").Short('p').String()
	indexerTOId    = indexerCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	indexerBoltDB  = indexerCommand.Flag("boltdb", "Use BoltDB. Only work when not using TOId version.").Short('b').Bool()
	indexerConfig  = indexerCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	indexerInfo    = indexerCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	indexerDebug   = indexerCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()
)

func main() {
	switch kingpin.MustParse(gochariots.Parse(os.Args[1:])) {
	case appCommand.FullCommand():
		if *appPort == "" {
			*appPort = "8080"
		}
		info.InitChariots(*appNumDC, *appID)
		indexerCommand.GetArg("info")
		info.SetName("app" + *appPort)
		info.SetPort(*appPort)
		level := logrus.WarnLevel
		if *appDebug {
			level = logrus.DebugLevel
		} else if *appInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		if *appConfig != "" {
			info.Config(*appConfig, "app")
		}
		if *appTOId {
			app.TOIDRun(*appPort)
		} else {
			app.Run(*appPort)
		}
	case controllerCommand.FullCommand():
		if *controllerPort == "" {
			*controllerPort = "8081"
		}
		info.InitChariots(*controllerNumDC, *controllerID)
		info.SetName("controller" + *controllerPort)
		info.SetPort(*controllerPort)
		level := logrus.WarnLevel
		if *controllerDebug {
			level = logrus.DebugLevel
		} else if *controllerInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		info.StartController(*controllerPort)
	case batcherCommand.FullCommand():
		if *batcherPort == "" {
			*batcherPort = "9000"
		}
		info.InitChariots(*batcherNumDC, *batcherID)
		info.SetName("batcher" + *batcherPort)
		info.SetPort(*batcherPort)
		level := logrus.WarnLevel
		if *batcherDebug {
			level = logrus.DebugLevel
		} else if *batcherInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		if *batcherTOId {
			batcher.TOIDInitBatcher(*batcherNumDC)
		} else {
			batcher.InitBatcher(*batcherNumDC)
		}
		ln, err := net.Listen("tcp", ":"+*batcherPort)
		if err != nil {
			panic(err)
		}
		defer ln.Close()
		if *batcherConfig != "" {
			info.Config(*batcherConfig, "batcher")
		}
		fmt.Println(info.GetName()+" is listening to port", *batcherPort)
		if *batcherTOId {
			go batcher.TOIDSweeper()
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					fmt.Println(info.GetName(), "Couldn't handle more connection", err)
					continue
				}
				// Handle connections in a new goroutine.
				go batcher.TOIDHandleRequest(conn)
			}
		} else {
			go batcher.Sweeper()
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					fmt.Println(info.GetName(), "Couldn't handle more connection", err)
					continue
				}
				// Handle connections in a new goroutine.
				go batcher.HandleRequest(conn)
			}
		}
	case filterCommand.FullCommand():
		if *filterPort == "" {
			*filterPort = "9010"
		}
		info.InitChariots(*filterNumDC, *filterID)
		info.SetName("filter" + *filterPort)
		info.SetPort(*filterPort)
		level := logrus.WarnLevel
		if *filterDebug {
			level = logrus.DebugLevel
		} else if *filterInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		if *filterTOId {
			filter.TOIDInitFilter(info.NumDC)
		} else {
			filter.InitFilter(info.NumDC)
		}
		ln, err := net.Listen("tcp", ":"+*filterPort)
		if err != nil {
			panic(err)
		}
		defer ln.Close()
		if *filterConfig != "" {
			info.Config(*filterConfig, "filter")
		}
		fmt.Println(info.GetName()+" is listening to port", *filterPort)
		if *filterTOId {
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
	case queueCommand.FullCommand():
		if *queuePort == "" {
			*queuePort = "9020"
		}
		info.InitChariots(*queueNumDC, *queueID)
		info.SetName("queue" + *queuePort)
		info.SetPort(*queuePort)
		level := logrus.WarnLevel
		if *queueDebug {
			level = logrus.DebugLevel
		} else if *queueInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		if *queueTOId {
			queue.TOIDInitQueue(*queueHold, *queueCarry)
		} else {
			queue.InitQueue(*queueHold)
		}
		ln, err := net.Listen("tcp", ":"+*queuePort)
		if err != nil {
			panic(err)
		}
		defer ln.Close()
		if *queueConfig != "" {
			info.Config(*queueConfig, "queue")
		}
		fmt.Println(info.GetName()+" is listening to port", *queuePort)
		if *queueTOId {
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					panic(err)
				}
				// Handle connections in a new goroutine.
				go queue.TOIDHandleRequest(conn)
			}
		} else {
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					panic(err)
				}
				// Handle connections in a new goroutine.
				go queue.HandleRequest(conn)
			}
		}
	case maintainerCommand.FullCommand():
		if *maintainerPort == "" {
			*maintainerPort = "9030"
		}
		info.InitChariots(*maintainerNumDC, *maintainerID)
		info.SetName("maintainer" + *maintainerPort)
		info.SetPort(*maintainerPort)
		level := logrus.WarnLevel
		if *maintainerDebug {
			level = logrus.DebugLevel
		} else if *maintainerInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		maintainer.InitLogMaintainer(info.GetName(), *maintainerInstN)
		ln, err := net.Listen("tcp", ":"+*maintainerPort)
		if err != nil {
			fmt.Println(info.GetName() + "couldn't listen on port " + *maintainerPort)
			panic(err)
		}
		defer ln.Close()
		if *maintainerConfig != "" {
			info.Config(*maintainerConfig, "maintainer")
		}
		fmt.Println(info.GetName()+" is listening to port", *maintainerPort)
		if *maintainerTOId {
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
	case indexerCommand.FullCommand():
		if *indexerPort == "" {
			*indexerPort = "9040"
		}
		info.InitChariots(*indexerNumDC, *indexerID)
		info.SetName("indexer" + *indexerPort)
		info.SetPort(*indexerPort)
		level := logrus.WarnLevel
		if *indexerDebug {
			level = logrus.DebugLevel
		} else if *indexerInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		if *indexerTOId {
			indexer.TOIDInitIndexer(info.GetName())
		} else {
			indexer.InitIndexer(info.GetName(), *indexerBoltDB)
		}
		ln, err := net.Listen("tcp", ":"+*indexerPort)
		if err != nil {
			fmt.Println(info.GetName() + "couldn't listen on port " + *indexerPort)
			panic(err)
		}
		defer ln.Close()
		if *indexerConfig != "" {
			info.Config(*indexerConfig, "indexer")
		}
		fmt.Println(info.GetName()+" is listening to port", *indexerPort)
		if *indexerTOId {
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					panic(err)
				}
				// Handle connections in a new goroutine.
				go indexer.TOIDHandleRequest(conn)
			}
		} else {
			for {
				// Listen for an incoming connection.
				conn, err := ln.Accept()
				if err != nil {
					panic(err)
				}
				// Handle connections in a new goroutine.
				go indexer.HandleRequest(conn)
			}
		}
	}
}
