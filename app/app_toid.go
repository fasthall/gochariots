package app

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/fasthall/gochariots/batcher"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/gin-gonic/gin"
)

type TOIDJsonRecord struct {
	ID      uint64            `json:"id"`
	StrID   string            `json:"strid"`
	Tags    map[string]string `json:"tags"`
	PreHost int               `json:"prehost"`
	PreTOId int               `json:"pretoid"`
}

func TOIDRun(port string) {
	router := gin.Default()
	router.POST("/record", TOIDpostRecord)
	router.POST("/batcher", addBatchers)
	router.GET("/batcher", getBatchers)
	router.POST("/indexer", addIndexer)
	router.GET("/indexer", getindexers)
	router.Run(":" + port)
}

func TOIDpostRecord(c *gin.Context) {
	logrus.WithField("timestamp", time.Now()).Info("postRecord")
	var jsonRecord TOIDJsonRecord
	err := c.Bind(&jsonRecord)
	if err != nil {
		panic(err)
	}
	if jsonRecord.StrID != "" {
		id, err := strconv.ParseUint(jsonRecord.StrID, 10, 64)
		if err != nil {
			panic(err)
		}
		jsonRecord.ID = id
	}
	// send to batcher
	r := batcher.RPCRecord{
		Id:   jsonRecord.ID,
		Host: int32(info.ID),
		Tags: jsonRecord.Tags,
		Causality: &batcher.RPCCausality{
			Host: int32(jsonRecord.PreHost),
			Toid: int32(jsonRecord.PreTOId),
		},
	}

	hostID := rand.Intn(len(batcherPool))
	_, err = batcherClient[hostID].TOIDReceiveRecord(context.Background(), &r)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't connect to batcher")
		return
	}
	logrus.WithField("timestamp", time.Now()).Info("sendToBatcher")
	logrus.WithField("id", hostID).Info("sent record to batcher")
	c.String(http.StatusOK, "Record posted")
}
