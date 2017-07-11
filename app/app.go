package app

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/record"
	"github.com/gin-gonic/gin"
)

var batcherConn []net.Conn
var batcherPool []string
var batchersVer int
var connMutex sync.Mutex

type JsonRecord struct {
	Tags    map[string]string `json:"tags"`
	PreHash uint64            `json:"prehash"`
	Seed    uint64            `json:"seed"`
}

func Run(port string) {
	router := gin.Default()
	router.POST("/record", postRecord)
	router.GET("/record/:lid", getRecord)
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
		batcherConn = make([]net.Conn, len(batcherPool))
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

func dialConn(hostID int) error {
	var err error
	batcherConn[hostID], err = net.Dial("tcp", batcherPool[hostID])
	return err
}

func postRecord(c *gin.Context) {
	logrus.WithField("timestamp", time.Now()).Info("postRecord")
	var jsonRecord JsonRecord
	err := c.Bind(&jsonRecord)
	if err != nil {
		panic(err)
	}

	// send to batcher
	r := record.Record{
		Host: info.ID,
		Tags: jsonRecord.Tags,
		Pre: record.Causality{
			Hash: jsonRecord.PreHash,
		},
		Seed: jsonRecord.Seed,
	}
	jsonBytes, err := record.ToJSON(r)
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
			logrus.WithField("host", batcherPool[hostID]).Error("couldn't connect to batcher")
		} else {
			logrus.WithField("host", batcherPool[hostID]).Info("connected to batcher")
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

func getRecord(c *gin.Context) {
	lid, err := strconv.Atoi(c.Param("lid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid LId",
			"error":   err,
		})
		return
	}
	// shouldn't be called like this, app should directly ask remote maintainer
	r, err := maintainer.ReadByLId(lid)
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
		"Seed":      r.Seed,
	})
}
