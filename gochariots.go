package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/fasthall/gochariots/app"
	"github.com/fasthall/gochariots/batcher"
	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/fasthall/gochariots/controller"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/queue"

	"github.com/Sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	gochariots = kingpin.New("gochariots", "A distributed shared log system.")

	appCommand = gochariots.Command("app", "Start an app instance.")
	appNumDC   = appCommand.Flag("num_dc", "The port app listens to.").Int()
	appID      = appCommand.Flag("id", "The port app listens to.").Int()
	appPort    = appCommand.Flag("port", "The port app listens to. By default it's 8080").Short('p').String()
	appTOId    = appCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	appConfig  = appCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	appInfo    = appCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	appDebug   = appCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	batcherCommand    = gochariots.Command("batcher", "Start a batcher instance.")
	batcherNumDC      = batcherCommand.Flag("num_dc", "The port batcher listens to.").Int()
	batcherID         = batcherCommand.Flag("id", "The port batcher listens to.").Int()
	batcherBufferSize = batcherCommand.Flag("buffer_size", "The buffer size.").Default("2000").Int()
	batcherPort       = batcherCommand.Flag("port", "The port app listens to. By default it's 9000").Short('p').String()
	batcherTOId       = batcherCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	batcherConfig     = batcherCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	batcherInfo       = batcherCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	batcherDebug      = batcherCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()
	batcherBenchmark  = batcherCommand.Flag("benchmark_accuracy", "How many records betweeen each throughput logging.").Int()

	controllerCommand = gochariots.Command("controller", "Start a controller instance.")
	controllerNumDC   = controllerCommand.Flag("num_dc", "The port controller listens to.").Int()
	controllerID      = controllerCommand.Flag("id", "The port controller listens to.").Int()
	controllerPort    = controllerCommand.Flag("port", "The port app listens to. By default it's 8081").Short('p').String()
	controllerInfo    = controllerCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	controllerDebug   = controllerCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	filterCommand = gochariots.Command("filter", "Start a filter instance.")
	filterNumDC   = filterCommand.Flag("num_dc", "The port filter listens to.").Int()
	filterID      = filterCommand.Flag("id", "The port contrfilteroller listens to.").Int()
	filterPort    = filterCommand.Flag("port", "The port app listens to. By default it's 9010").Short('p').String()
	filterTOId    = filterCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	filterConfig  = filterCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	filterInfo    = filterCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	filterDebug   = filterCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()

	queueCommand        = gochariots.Command("queue", "Start a queue instance.")
	queueNumDC          = queueCommand.Flag("num_dc", "The port queue listens to.").Int()
	queueID             = queueCommand.Flag("id", "The port queue listens to.").Int()
	queuePort           = queueCommand.Flag("port", "The port app listens to. By default it's 9020").Short('p').String()
	queueHold           = queueCommand.Flag("hold", "Whether this queue instance holds a token when launched.").Required().Short('h').Bool()
	queueTOId           = queueCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	queueCarry          = queueCommand.Flag("carry", "Carry deferred records with token. Only work when toid is on.").Short('c').Bool()
	queueConfig         = queueCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	queueInfo           = queueCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	queueDebug          = queueCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()
	queueTwoPhaseAppend = queueCommand.Flag("two_phase_append", "Enable two phase append.").Bool()
	queueQuerySizeLimit = queueCommand.Flag("query_size", "Query size of MongoDB. Only work when using two phase append.").Short('q').Int()
	queueBenchmark      = queueCommand.Flag("benchmark_accuracy", "How many records betweeen each throughput logging.").Int()

	maintainerCommand   = gochariots.Command("maintainer", "Start a maintainer instance.")
	maintainerNumDC     = maintainerCommand.Flag("num_dc", "The port maintainer listens to.").Int()
	maintainerID        = maintainerCommand.Flag("id", "The port maintainer listens to.").Int()
	maintainerPort      = maintainerCommand.Flag("port", "The port app listens to. By default it's 9030").Short('p').String()
	maintainerTOId      = maintainerCommand.Flag("toid", "Use TOId version.").Short('t').Bool()
	maintainerConfig    = maintainerCommand.Flag("config_file", "Configuration file to read.").Short('f').String()
	maintainerInfo      = maintainerCommand.Flag("info", "Turn on info level logging.").Short('i').Bool()
	maintainerDebug     = maintainerCommand.Flag("debug", "Turn on debug level logging.").Short('d').Bool()
	maintainerDynamoDB  = maintainerCommand.Flag("dynamodb", "Use DynamoDB as physical storage.").Bool()
	maintainerDatastore = maintainerCommand.Flag("datastore", "Use Datastore as physical storage.").Bool()
	maintainerCosmosDB  = maintainerCommand.Flag("cosmosdb", "Use CosmosDB as physical storage.").Bool()
	maintainerMongoDB   = maintainerCommand.Flag("mongodb", "Use MongoDB as physical storage.").Bool()
	maintainerBenchmark = maintainerCommand.Flag("benchmark_accuracy", "How many records betweeen each throughput logging.").Int()
)

func main() {
	switch kingpin.MustParse(gochariots.Parse(os.Args[1:])) {
	case appCommand.FullCommand():
		if *appPort == "" {
			*appPort = "8080"
		}
		if *appNumDC == 0 {
			*appNumDC = 1
		}
		info.NumDC = *appNumDC
		info.ID = *appID
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
		if *controllerNumDC == 0 {
			*controllerNumDC = 1
		}
		info.NumDC = *controllerNumDC
		info.ID = *controllerID
		info.SetName("controller" + *controllerPort)
		info.SetPort(*controllerPort)
		level := logrus.WarnLevel
		if *controllerDebug {
			level = logrus.DebugLevel
		} else if *controllerInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		controller.StartController(*controllerPort)
	case batcherCommand.FullCommand():
		if *batcherPort == "" {
			*batcherPort = "9000"
		}
		if *batcherNumDC == 0 {
			*batcherNumDC = 1
		}
		info.NumDC = *batcherNumDC
		info.ID = *batcherID
		info.SetName("batcher" + *batcherPort)
		info.SetPort(*batcherPort)
		level := logrus.WarnLevel
		if *batcherDebug {
			level = logrus.DebugLevel
		} else if *batcherInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		if *batcherConfig != "" {
			info.Config(*batcherConfig, "batcher")
		}
		if *batcherTOId {
			batcher.TOIDInitBatcher(*batcherBufferSize, *batcherBenchmark)
		} else {
			batcher.InitBatcher(*batcherBufferSize, *batcherBenchmark)
		}
		ln, err := net.Listen("tcp", ":"+*batcherPort)
		if err != nil {
			panic(err)
		}
		defer ln.Close()
		fmt.Println(info.GetName()+" is listening to port", *batcherPort)
		s := grpc.NewServer()
		batcherrpc.RegisterBatcherRPCServer(s, &batcher.Server{})
		reflection.Register(s)
		if err := s.Serve(ln); err != nil {
			logrus.Fatalf("failed to serve: %v", err)
		}
	case queueCommand.FullCommand():
		if *queuePort == "" {
			*queuePort = "9020"
		}
		if *queueNumDC == 0 {
			*queueNumDC = 1
		}
		if *queueQuerySizeLimit == 0 {
			*queueQuerySizeLimit = 1000
		}
		info.NumDC = *queueNumDC
		info.ID = *queueID
		info.SetName("queue" + *queuePort)
		info.SetPort(*queuePort)
		level := logrus.WarnLevel
		if *queueDebug {
			level = logrus.DebugLevel
		} else if *queueInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		if *queueConfig != "" {
			info.Config(*queueConfig, "queue")
		}
		if *queueTOId {
			queue.TOIDInitQueue(*queueHold, *queueCarry, *queueBenchmark)
		} else {
			queue.InitQueue(*queueHold, *queueQuerySizeLimit, *queueBenchmark)
		}
		ln, err := net.Listen("tcp", ":"+*queuePort)
		if err != nil {
			panic(err)
		}
		defer ln.Close()
		fmt.Println(info.GetName()+" is listening to port", *queuePort)
		s := grpc.NewServer()
		queue.RegisterQueueServer(s, &queue.Server{})
		reflection.Register(s)
		if err := s.Serve(ln); err != nil {
			logrus.Fatalf("failed to serve: %v", err)
		}
	case maintainerCommand.FullCommand():
		if *maintainerPort == "" {
			*maintainerPort = "9030"
		}
		if *maintainerNumDC == 0 {
			*maintainerNumDC = 1
		}
		info.NumDC = *maintainerNumDC
		info.ID = *maintainerID
		info.SetName("maintainer" + *maintainerPort)
		info.SetPort(*maintainerPort)
		level := logrus.WarnLevel
		if *maintainerDebug {
			level = logrus.DebugLevel
		} else if *maintainerInfo {
			level = logrus.InfoLevel
		}
		info.RedirectLog(info.GetName()+".log", level)
		if *maintainerConfig != "" {
			info.Config(*maintainerConfig, "maintainer")
		}
		maintainer.InitLogMaintainer(*maintainerBenchmark)
		ln, err := net.Listen("tcp", ":"+*maintainerPort)
		if err != nil {
			fmt.Println(info.GetName() + "couldn't listen on port " + *maintainerPort)
			panic(err)
		}
		defer ln.Close()
		fmt.Println(info.GetName()+" is listening to port", *maintainerPort)
		s := grpc.NewServer()
		maintainer.RegisterMaintainerServer(s, &maintainer.Server{})
		reflection.Register(s)
		if err := s.Serve(ln); err != nil {
			logrus.Fatalf("failed to serve: %v", err)
		}
	}
}
