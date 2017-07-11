package app

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/record"
	"github.com/gin-gonic/gin"
)

type TOIDJsonRecord struct {
	Tags    map[string]string `json:"tags"`
	PreHost int               `json:"prehost"`
	PreTOId int               `json:"pretoid"`
}

func TOIDRun(port string) {
	router := gin.Default()
	router.POST("/record", TOIDpostRecord)
	router.GET("/record/:lid", TOIDgetRecord)
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
	r := record.TOIDRecord{
		Host: info.ID,
		Tags: jsonRecord.Tags,
		Pre: record.TOIDCausality{
			Host: jsonRecord.PreHost,
			TOId: jsonRecord.PreTOId,
		},
	}
	jsonBytes, err := record.TOIDToJSON(r)
	if err != nil {
		logrus.WithField("timestamp", time.Now()).Error("couldn't convert record to bytes")
		c.String(http.StatusBadRequest, "Couldn't convert record to bytes")
		return
	}
	b := make([]byte, 5)
	b[4] = byte('r')
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))

	hostID := rand.Intn(len(batcherPool))
	connMutex.Lock()
	if batcherConn[hostID] == nil {
		err = dialConn(hostID)
		if err != nil {
			logrus.WithField("id", hostID).Error("couldn't connect to batcher")
		} else {
			logrus.WithField("id", hostID).Info("connected to batcher")
		}
	}
	connMutex.Unlock()
	cnt := 5
	sent := false
	for sent == false {
		connMutex.Lock()
		if batcherConn[hostID] != nil {
			_, err = batcherConn[hostID].Write(append(b, jsonBytes...))
		} else {
			err = errors.New("batcherConn[hostID] == nil")
		}
		connMutex.Unlock()
		logrus.WithField("timestamp", time.Now()).Info("sendToBatcher")
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialConn(hostID)
				if err != nil {
					logrus.WithField("attempt", cnt).Warning("couldn't connect to batcher, retrying...")
				} else {
					logrus.WithField("id", hostID).Info("connected to batcher")
				}
			} else {
				logrus.WithField("id", hostID).Error("failed to connect to the batcher after retrying 5 times")
				c.String(http.StatusServiceUnavailable, "Couldn't connect to the batcher")
				return
			}
		} else {
			sent = true
			logrus.WithField("id", hostID).Info("sent record to batcher")
		}
	}
	c.String(http.StatusOK, "Record posted")
}

func TOIDgetRecord(c *gin.Context) {
	lid, err := strconv.Atoi(c.Param("lid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid LId",
			"error":   err,
		})
		return
	}
	// shouldn't be called like this, app should directly ask remote maintainer
	r, err := maintainer.TOIDReadByLId(lid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Record not found",
			"error":   err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"LId":       r.LId,
		"Host":      r.Host,
		"Causality": r.Pre,
		"Tags":      r.Tags,
	})
}
