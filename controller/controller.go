package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/misc"
	"github.com/fasthall/gochariots/queue"
	"github.com/gin-gonic/gin"
)

var apps []string
var appsVersion uint32
var batchers []string
var batcherClient []batcherrpc.BatcherRPCClient
var batchersVersion uint32
var queues []string
var queueClient []queue.QueueClient
var queuesVersion uint32
var maintainers []string
var maintainerClient []maintainer.MaintainerClient
var maintainersVersion uint32
var indexers []string
var indexersVersion uint32
var remoteBatcher []string
var remoteBatcherVer uint32
var mutex sync.Mutex

// StartController starts controller's REST API server on sepcified port
func StartController(port string) {
	router := gin.Default()

	router.GET("/", getInfo)
	router.POST("/app", addApps)
	router.GET("/app", getApps)
	router.POST("/batcher", addBatchers)
	router.GET("/batcher", getBatchers)
	router.POST("/queue", addQueue)
	router.GET("/queue", getQueues)
	router.POST("/maintainer", addMaintainer)
	router.GET("/maintainer", getMaintainers)
	router.POST("/indexer", addIndexer)
	router.GET("/indexer", getIndexers)
	router.POST("/remote/batcher", addRemoteBatcher)
	router.GET("/remote/batcher", getRemoteBatcher)

	router.Run(":" + port)
}

func getInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"apps":           apps,
		"batchers":       batchers,
		"queues":         queues,
		"maintainers":    maintainers,
		"indexers":       indexers,
		"remoteBatchers": remoteBatcher,
	})
}

func addApps(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	apps = append(apps, c.Query("host"))
	appsVersion++
	mutex.Unlock()

	go informAppBatcher(c.Query("host"))
	go informAppIndexer(c.Query("host"))

	c.String(http.StatusOK, c.Query("host")+" added")
}

func informAppBatcher(host string) {
	jsonBatchers, err := json.Marshal(batchers)
	if err != nil {
		logrus.WithError(err).Error("couldn't convert batchers to bytes")
		panic("failing to update cluster may cause unexpected error")
	}
	p := misc.NewParams()
	p.AddParam("host", string(jsonBatchers))
	p.AddParam("ver", strconv.Itoa(int(batchersVersion)))
	code := http.StatusBadRequest
	for code != http.StatusOK {
		time.Sleep(1 * time.Second)
		code, _, err = misc.Report(host, "batcher", p)
		if err != nil {
			logrus.WithError(err).Error("couldn't inform app about batchers")
		}
	}
	logrus.WithFields(logrus.Fields{"host": host, "batchers": batchers}).Info("successfully informed app about batcher list")
}

func informAppIndexer(host string) {
	jsonIndexers, err := json.Marshal(indexers)
	if err != nil {
		logrus.WithError(err).Error("couldn't convert indexers to bytes")
		panic("failing to update cluster may cause unexpected error")
	}
	p := misc.NewParams()
	p.AddParam("host", string(jsonIndexers))
	p.AddParam("ver", strconv.Itoa(int(indexersVersion)))
	code := http.StatusBadRequest
	for code != http.StatusOK {
		time.Sleep(1 * time.Second)
		code, _, err = misc.Report(host, "indexer", p)
		if err != nil {
			logrus.WithError(err).Error("couldn't inform app about indexers")
		}
	}
	logrus.WithFields(logrus.Fields{"host": host, "indexers": indexers}).Info("successfully informed app about indexer list")
}

func getApps(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"apps": apps,
	})
}

func addBatchers(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	batchers = append(batchers, c.Query("host"))
	batchersVersion++
	conn, err := grpc.Dial(c.Query("host"), grpc.WithInsecure())
	if err != nil {
		c.String(http.StatusBadRequest, "couldn't connect to batcher")
		return
	}
	cli := batcherrpc.NewBatcherRPCClient(conn)
	batcherClient = append(batcherClient, cli)
	mutex.Unlock()
	for _, host := range apps {
		informAppBatcher(host)
	}
	for i, cli := range batcherClient {
		rpcQueues := batcherrpc.RPCQueues{
			Version: queuesVersion,
			Queues:  queues,
		}
		_, err := cli.UpdateQueue(context.Background(), &rpcQueues)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send queue list to batcher")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"id": i, "queues": queues}).Info("successfully informed batcher about queue list")
	}
	c.String(http.StatusOK, c.Query("host")+" added")
}

func getBatchers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": batchers,
	})
}

func addQueue(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	queues = append(queues, c.Query("host"))
	conn, err := grpc.Dial(c.Query("host"), grpc.WithInsecure())
	cli := queue.NewQueueClient(conn)
	queueClient = append(queueClient, cli)
	queuesVersion++
	mutex.Unlock()

	c.String(http.StatusOK, c.Query("host")+" added")
	// update "next queue" for each queue
	for i := range queueClient {
		host := queues[(i+1)%len(queues)]
		rpcQueue := queue.RPCQueue{
			Version: queuesVersion,
			Queue:   host,
		}
		_, err := queueClient[i].UpdateNextQueue(context.Background(), &rpcQueue)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send next queue host")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"id": i, "queue": queues[(i+1)%len(queues)]}).Info("successfully informed queue about next queue host")
	}
	// update batchers about queues
	for i, cli := range batcherClient {
		rpcQueues := batcherrpc.RPCQueues{}
		_, err := cli.UpdateQueue(context.Background(), &rpcQueues)
		if err != nil {
			logrus.WithField("id", i).WithError(err).Error("couldn't send new queue host")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"batcher": i, "queues": queues}).Info("successfully informed batcher about new queue host")
	}

	// update queues' maintainer list
	rpcMaintainers := queue.RPCMaintainers{
		Version:    maintainersVersion,
		Maintainer: maintainers,
	}
	_, err = cli.UpdateMaintainers(context.Background(), &rpcMaintainers)
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send maintainer list to new queue")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("maintainers", maintainers).Info("successfully informed new queue about maintainer list")

	// tell queue about indexer
	rpcIndexers := queue.RPCIndexers{
		Version: indexersVersion,
		Indexer: indexers,
	}
	_, err = cli.UpdateIndexers(context.Background(), &rpcIndexers)
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send indexer list to queue")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("indexers", indexers).Info("successfully informed queue about indexer list")
}

func getQueues(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"queues": queues,
	})
}

func addMaintainer(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	maintainers = append(maintainers, c.Query("host"))
	conn, err := grpc.Dial(c.Query("host"), grpc.WithInsecure())
	cli := maintainer.NewMaintainerClient(conn)
	maintainerClient = append(maintainerClient, cli)
	maintainersVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

	// tell maintainer its indexer
	i := len(maintainers) - 1
	if i < len(indexers) {
		rpcIndexer := maintainer.RPCIndexer{
			Version: indexersVersion,
			Indexer: indexers[i],
		}
		_, err := cli.UpdateIndexer(context.Background(), &rpcIndexer)
		if err != nil {
			logrus.WithFields(logrus.Fields{"maintainer": c.Query("host"), "indexer": indexers[i]}).Error("couldn't notify maintainer its indexer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"maintainer": c.Query("host"), "indexer": indexers[i]}).Info("successfully informed maintainer its indexer")
	}

	// update queues' maintainer list
	for _, cli := range queueClient {
		rpcMaintainers := queue.RPCMaintainers{
			Version:    maintainersVersion,
			Maintainer: maintainers,
		}
		_, err := cli.UpdateMaintainers(context.Background(), &rpcMaintainers)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send maintainer list to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("maintainers", maintainers).Info("successfully informed queues about maintainer list")
	}

	// update remote batchers
	rpcBatchers := maintainer.RPCBatchers{
		Version: remoteBatcherVer,
		Batcher: remoteBatcher,
	}
	_, err = cli.UpdateBatchers(context.Background(), &rpcBatchers)
	if err != nil {
		logrus.WithField("id", i).Error("couldn't send remoteBatcher to maintainer")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("batchers", remoteBatcher).Info("successfully informed maintainers about new remote batchers")
}

func getMaintainers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"maintainers": maintainers,
	})
}

func addIndexer(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	indexers = append(indexers, c.Query("host"))
	indexersVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

	for _, host := range apps {
		informAppIndexer(host)
	}

	// tell maintainer its indexer
	i := len(indexers) - 1
	if i < len(maintainers) {
		rpcIndexer := maintainer.RPCIndexer{
			Version: indexersVersion,
			Indexer: indexers[i],
		}
		_, err := maintainerClient[i].UpdateIndexer(context.Background(), &rpcIndexer)
		if err != nil {
			logrus.WithFields(logrus.Fields{"maintainer": c.Query("host"), "indexer": indexers[i]}).Error("couldn't notify maintainer its indexer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"maintainer": maintainers[i], "indexer": indexers[i]}).Info("successfully informed maintainer its indexer")
	}

	// update queues' indexer list
	for i, cli := range queueClient {
		rpcIndexers := queue.RPCIndexers{
			Version: indexersVersion,
			Indexer: indexers,
		}
		_, err := cli.UpdateIndexers(context.Background(), &rpcIndexers)
		if err != nil {
			logrus.WithField("host", queues[i]).Error("couldn't send indexer list to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("indexers", indexers).Info("successfully informed queues about indexer list")
	}
}

func getIndexers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"indexers": indexers,
	})
}

func addRemoteBatcher(c *gin.Context) {
	dc, err := strconv.Atoi(c.Query("dc"))
	if err != nil {
		logrus.WithField("parameter", c.Query("dc")).Warning("received invalid parameter")
		c.String(http.StatusBadRequest, "invalid parameter")
		return
	}
	mutex.Lock()
	for len(remoteBatcher) <= dc {
		remoteBatcher = append(remoteBatcher, "")
	}
	remoteBatcher[dc] = c.Query("host")
	remoteBatcherVer++
	mutex.Unlock()

	for i, cli := range maintainerClient {
		rpcBatchers := maintainer.RPCBatchers{
			Version: remoteBatcherVer,
			Batcher: remoteBatcher,
		}
		_, err := cli.UpdateBatchers(context.Background(), &rpcBatchers)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send remoteBatcher to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("batchers", remoteBatcher).Info("successfully informed maintainers about new remote batchers")
	}
	c.String(http.StatusOK, "remoteBatcher["+c.Query("dc")+"] = "+c.Query("host")+" updated")
}

func getRemoteBatcher(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": remoteBatcher,
	})
}
