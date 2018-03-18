package app

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/satori/go.uuid"

	"google.golang.org/grpc"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/gin-gonic/gin"
)

var batcherClient []batcherrpc.BatcherRPCClient
var batcherPool []string
var batchersVer int

type JsonRecord struct {
	ID    string            `json:"id"`
	Tags  map[string]string `json:"tags"`
	Trace string            `json:"trace"`
	SeqID int               `json:"seqid"`
	Depth int               `json:"depth"`
}

func Run(port string) {
	router := gin.Default()
	router.POST("/record", postRecord)
	router.POST("/batcher", addBatchers)
	router.GET("/batcher", getBatchers)
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

func postRecord(c *gin.Context) {
	logrus.WithField("timestamp", time.Now()).Info("postRecord")
	var jsonRecord JsonRecord
	err := c.Bind(&jsonRecord)
	if err != nil {
		panic(err)
	}

	// send to batcher
	r := batcherrpc.RPCRecord{
		Id:    jsonRecord.ID,
		Host:  uint32(info.ID),
		Tags:  jsonRecord.Tags,
		Trace: jsonRecord.Trace,
		Seqid: int64(jsonRecord.SeqID),
		Depth: uint32(jsonRecord.Depth),
	}
	if r.Id == "" {
		r.Id = uuid.NewV4().String()
	}

	hostID := rand.Intn(len(batcherPool))
	_, err = batcherClient[hostID].ReceiveRecord(context.Background(), &r)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't connect to batcher")
		return
	}
	logrus.WithField("timestamp", time.Now()).Info("sendToBatcher")
	logrus.WithField("id", hostID).Info("sent record to batcher")
	c.String(http.StatusOK, r.Id)
}
