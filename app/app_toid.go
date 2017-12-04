package app

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/fasthall/gochariots/batcher/batcherrpc"
	"github.com/satori/go.uuid"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/gin-gonic/gin"
)

type TOIDJsonRecord struct {
	ID      string            `json:"id"`
	Tags    map[string]string `json:"tags"`
	PreHost uint32            `json:"prehost"`
	PreTOId uint32            `json:"pretoid"`
}

func TOIDRun(port string) {
	router := gin.Default()
	router.POST("/record", TOIDpostRecord)
	router.POST("/batcher", addBatchers)
	router.GET("/batcher", getBatchers)
	router.Run(":" + port)
}

func TOIDpostRecord(c *gin.Context) {
	logrus.WithField("timestamp", time.Now()).Info("postRecord")
	var jsonRecord TOIDJsonRecord
	err := c.Bind(&jsonRecord)
	if err != nil {
		panic(err)
	}
	// send to batcher
	r := batcherrpc.RPCRecord{
		Id:   jsonRecord.ID,
		Host: uint32(info.ID),
		Tags: jsonRecord.Tags,
		Causality: &batcherrpc.RPCCausality{
			Host: jsonRecord.PreHost,
			Toid: jsonRecord.PreTOId,
		},
	}
	if r.Id == "" {
		r.Id = uuid.NewV4().String()
	}

	hostID := rand.Intn(len(batcherPool))
	_, err = batcherClient[hostID].TOIDReceiveRecord(context.Background(), &r)
	if err != nil {
		c.String(http.StatusInternalServerError, "couldn't connect to batcher")
		return
	}
	logrus.WithField("timestamp", time.Now()).Info("sendToBatcher")
	logrus.WithField("id", hostID).Info("sent record to batcher")
	c.String(http.StatusOK, r.Id)
}
