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
	"github.com/fasthall/gochariots/cache"
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
var storages []string
var storagesVersion uint32
var caches []string
var cacheClient []cache.CacheClient
var cachesVersion uint32
var remoteBatcher []string
var remoteBatcherVer uint32
var mutex sync.Mutex

// StartController starts controller's REST API server on sepcified port
func StartController(port string) {
	appsVersion = 1
	batchersVersion = 1
	queuesVersion = 1
	maintainersVersion = 1
	remoteBatcherVer = 1

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
	router.POST("/cache", addCache)
	router.GET("/cache", getCaches)
	router.POST("/storage", addStorage)
	router.GET("/storage", getStorages)
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
		"caches":         caches,
		"remoteBatchers": remoteBatcher,
	})
}

func addApps(c *gin.Context) {
	logrus.WithField("host", c.Query("host")).Info("addApps")
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	apps = append(apps, c.Query("host"))
	appsVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

	go informAppBatcher(c.Query("host"))
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

func getApps(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"apps": apps,
		"ver":  appsVersion,
	})
}

func addBatchers(c *gin.Context) {
	logrus.WithField("host", c.Query("host")).Info("addBatchers")
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	conn, err := grpc.Dial(c.Query("host"), grpc.WithInsecure())
	if err != nil {
		c.String(http.StatusBadRequest, "couldn't connect to batcher")
		logrus.WithError(err).Error("couldn't connect to " + c.Query("host"))
		return
	}
	cli := batcherrpc.NewBatcherRPCClient(conn)
	batcherClient = append(batcherClient, cli)
	batchers = append(batchers, c.Query("host"))
	batchersVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

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
}

func getBatchers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": batchers,
		"ver":      batchersVersion,
	})
}

func addQueue(c *gin.Context) {
	logrus.WithField("host", c.Query("host")).Info("addQueue")
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	conn, err := grpc.Dial(c.Query("host"), grpc.WithInsecure())
	if err != nil {
		c.String(http.StatusBadRequest, "couldn't connect to batcher")
		logrus.WithError(err).Error("couldn't connect to " + c.Query("host"))
		return
	}
	cli := queue.NewQueueClient(conn)
	queueClient = append(queueClient, cli)
	queues = append(queues, c.Query("host"))
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
	for i := range batcherClient {
		rpcQueues := batcherrpc.RPCQueues{
			Version: queuesVersion,
			Queues:  queues,
		}
		_, err := batcherClient[i].UpdateQueue(context.Background(), &rpcQueues)
		if err != nil {
			logrus.WithField("batcherID", i).WithError(err).Error("couldn't send new queue host")
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

	// update queues' cache list
	rpcCaches := queue.RPCCaches{
		Version: cachesVersion,
		Hosts:   caches,
	}
	_, err = cli.UpdateCaches(context.Background(), &rpcCaches)
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send cache list to new queue")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("caches", caches).Info("successfully informed new queue about cache list")
}

func getQueues(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"queues": queues,
		"ver":    queuesVersion,
	})
}

func addMaintainer(c *gin.Context) {
	logrus.WithField("host", c.Query("host")).Info("addMaintainer")
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	conn, err := grpc.Dial(c.Query("host"), grpc.WithInsecure())
	if err != nil {
		c.String(http.StatusBadRequest, "couldn't connect to batcher")
		logrus.WithError(err).Error("couldn't connect to " + c.Query("host"))
		return
	}
	cli := maintainer.NewMaintainerClient(conn)
	maintainerClient = append(maintainerClient, cli)
	maintainers = append(maintainers, c.Query("host"))
	maintainersVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

	// update queues' maintainer list
	for i := range queueClient {
		rpcMaintainers := queue.RPCMaintainers{
			Version:    maintainersVersion,
			Maintainer: maintainers,
		}
		_, err := queueClient[i].UpdateMaintainers(context.Background(), &rpcMaintainers)
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
		logrus.WithField("host", c.Query("host")).Error("couldn't send remoteBatcher to maintainer")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("batchers", remoteBatcher).Info("successfully informed maintainer about new remote batchers")

	// update cache list
	rpcCaches := maintainer.RPCCaches{
		Version: cachesVersion,
		Hosts:   caches,
	}
	_, err = cli.UpdateCaches(context.Background(), &rpcCaches)
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send cache list to maintainer")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("caches", caches).Info("successfully informed maintainer about cache list")

	rpcMongos := maintainer.RPCMongos{
		Version: storagesVersion,
		Hosts:   storages,
	}
	_, err = cli.UpdateMongos(context.Background(), &rpcMongos)
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send storage list to maintainer")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("storages", storages).Info("successfully informed maintainer about storage list")
}

func getMaintainers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"maintainers": maintainers,
		"ver":         maintainersVersion,
	})
}

func addCache(c *gin.Context) {
	logrus.WithField("host", c.Query("host")).Info("addCache")
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	conn, err := grpc.Dial(c.Query("host"), grpc.WithInsecure())
	if err != nil {
		c.String(http.StatusBadRequest, "couldn't connect to batcher")
		logrus.WithError(err).Error("couldn't connect to " + c.Query("host"))
		return
	}
	cli := cache.NewCacheClient(conn)
	cacheClient = append(cacheClient, cli)
	caches = append(caches, c.Query("host"))
	cachesVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

	// update queues' redisCaches list
	for i, cli := range queueClient {
		rpcCaches := queue.RPCCaches{
			Version: cachesVersion,
			Hosts:   caches,
		}
		_, err := cli.UpdateCaches(context.Background(), &rpcCaches)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send redisCaches list to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("caches", caches).Info("successfully informed queues about cache list")
	}

	// update maintainers' redisCaches list
	for i, cli := range maintainerClient {
		rpcCaches := maintainer.RPCCaches{
			Version: cachesVersion,
			Hosts:   caches,
		}
		_, err := cli.UpdateCaches(context.Background(), &rpcCaches)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send redisCaches list to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("caches", caches).Info("successfully informed maintainers about cache list")
	}

	// inform new cache about storage list
	rpcStorages := cache.RPCStorages{
		Version: storagesVersion,
		Hosts:   storages,
	}
	_, err = cli.UpdateStorage(context.Background(), &rpcStorages)
	if err != nil {
		logrus.WithError(err).Error("couldn't send storage list to cache")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("storages", storages).Info("successfully informed cache about storage list")
}

func getCaches(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"caches": caches,
		"ver":    cachesVersion,
	})
}

func addStorage(c *gin.Context) {
	logrus.WithField("host", c.Query("host")).Info("addStorage")
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	storages = append(storages, c.Query("host"))
	storagesVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

	// update maintainers' redisCaches list
	for _, cli := range maintainerClient {
		rpcMongos := maintainer.RPCMongos{
			Version: storagesVersion,
			Hosts:   storages,
		}
		_, err := cli.UpdateMongos(context.Background(), &rpcMongos)
		if err != nil {
			logrus.Error(len(maintainerClient))
			logrus.WithError(err).Error("couldn't send storage list to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("storages", storages).Info("successfully informed maintainers about storage list")
	}
}

func getStorages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"storages": storages,
		"ver":      cachesVersion,
	})
}

func addRemoteBatcher(c *gin.Context) {
	logrus.WithField("host", c.Query("host")).Info("addRemoteBatcher")
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
	c.String(http.StatusOK, "remoteBatcher["+c.Query("dc")+"] = "+c.Query("host")+" updated")

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
}

func getRemoteBatcher(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": remoteBatcher,
		"ver":      remoteBatcherVer,
	})
}
