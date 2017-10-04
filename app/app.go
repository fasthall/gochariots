package app

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/fasthall/gochariots/batcher/batcherrpc"

	"google.golang.org/grpc"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/gin-gonic/gin"
)

var batcherClient []batcherrpc.BatcherRPCClient
var batcherPool []string
var batchersVer int
var indexerPool []string
var indexersVer int

type JsonRecord struct {
	Tags    map[string]string `json:"tags"`
	Hash    []uint64          `json:"prehash"`
	StrHash []string          `json:"strhash"`
	Seed    uint64            `json:"seed"`
	StrSeed string            `json:"strseed"`
}

func Run(port string) {
	router := gin.Default()
	router.POST("/record", postRecord)
	router.POST("/batcher", addBatchers)
	router.GET("/batcher", getBatchers)
	router.POST("/indexer", addIndexer)
	router.GET("/indexer", getindexers)
	router.Run(":" + port)
}

func addBatchers(c *gin.Context) {
	ver, err := strconv.Atoi(c.Query("ver"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid version $ver")
		return
	}
	if ver > batchersVer {
		batchersVer = ver
		json.Unmarshal([]byte(c.Query("host")), &batcherPool)
		batcherClient = make([]batcherrpc.BatcherRPCClient, len(batcherPool))
		for i := range batcherPool {
			conn, err := grpc.Dial(batcherPool[i], grpc.WithInsecure())
			if err != nil {
				c.String(http.StatusBadRequest, "couldn't connect to batcher")
				return
			}
			batcherClient[i] = batcherrpc.NewBatcherRPCClient(conn)
		}
		c.String(http.StatusOK, c.Query("host")+" added")
		logrus.WithFields(logrus.Fields{"ver": batchersVer, "batchers": batcherPool}).Info("batcher list updated")
	} else {
		c.String(http.StatusOK, "received older version of batcher list")
		logrus.WithFields(logrus.Fields{"current": batchersVer, "received": ver}).Info("received older version of batcher list")
	}
}

func getBatchers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": batcherPool,
	})
}

func addIndexer(c *gin.Context) {
	ver, err := strconv.Atoi(c.Query("ver"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid version $ver")
		return
	}
	if ver > indexersVer {
		indexersVer = ver
		json.Unmarshal([]byte(c.Query("host")), &indexerPool)
		c.String(http.StatusOK, c.Query("host")+" added")
		logrus.WithFields(logrus.Fields{"ver": indexersVer, "indexers": indexerPool}).Info("indexer list updated")
	} else {
		c.String(http.StatusOK, "received older version of indexer list")
		logrus.WithFields(logrus.Fields{"current": indexersVer, "received": ver}).Info("received older version of indexer list")
	}
}

func getindexers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"indexer": indexerPool,
	})
}

func postRecord(c *gin.Context) {
	logrus.WithField("timestamp", time.Now()).Info("postRecord")
	var jsonRecord JsonRecord
	err := c.Bind(&jsonRecord)
	if err != nil {
		panic(err)
	}

	for _, strHash := range jsonRecord.StrHash {
		hash, err := strconv.ParseUint(strHash, 10, 64)
		if err != nil {
			panic(err)
		}
		jsonRecord.Hash = append(jsonRecord.Hash, hash)
	}
	if jsonRecord.StrSeed != "" {
		seed, err := strconv.ParseUint(jsonRecord.StrSeed, 10, 64)
		if err != nil {
			panic(err)
		}
		jsonRecord.Seed = seed
	}
	// send to batcher
	r := batcherrpc.RPCRecord{
		Host: uint32(info.ID),
		Tags: jsonRecord.Tags,
		Hash: jsonRecord.Hash,
		Seed: jsonRecord.Seed,
	}

	hostID := rand.Intn(len(batcherPool))
	_, err = batcherClient[hostID].ReceiveRecord(context.Background(), &r)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't connect to batcher")
		return
	}
	logrus.WithField("timestamp", time.Now()).Info("sendToBatcher")
	logrus.WithField("id", hostID).Info("sent record to batcher")
	c.String(http.StatusOK, "Record posted")
}
