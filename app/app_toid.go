package app

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/fasthall/gochariots/maintainer/indexer"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/misc/connection"
	"github.com/fasthall/gochariots/record"
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
	router.GET("/record/:id", TOIDgetRecord)
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
	r := record.TOIDRecord{
		ID:   jsonRecord.ID,
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ID",
			"error":   err,
		})
		return
	}
	timeoutStr := c.DefaultQuery("timeout", "10")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid timeout value",
			"error":   err,
		})
		return
	}
	start := time.Now()
	done := false
	for !done {
		for i := range indexerConn {
			entry, ok, err := queryIndexer(id, i)
			if err == nil && ok {
				c.JSON(http.StatusOK, gin.H{
					"LId":  entry.LId,
					"TOId": entry.TOId,
					"Host": entry.Host,
				})
				done = true
			} else if err != nil {
				logrus.WithError(err).Error("query indexer failed")
			}
		}
		if time.Since(start) > time.Duration(timeout)*time.Second {
			c.Status(http.StatusRequestTimeout)
			return
		}
	}
}

func queryIndexer(id uint64, indexerID int) (indexer.TOIDIndexTableEntry, bool, error) {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, uint64(id))
	b := make([]byte, 5)
	b[4] = byte('q')
	binary.BigEndian.PutUint32(b, 9)

	indexerConnMutex.Lock()
	if indexerConn[indexerID] == nil {
		err := dialIndexer(indexerID)
		if err != nil {
			logrus.WithFields(logrus.Fields{"id": indexerID, "host": indexerPool[indexerID]}).Error("couldn't connect to indexer")
		} else {
			logrus.WithFields(logrus.Fields{"id": indexerID, "host": indexerPool[indexerID]}).Info("connected to indexer")
		}
	}
	indexerConnMutex.Unlock()

	var err error
	cnt := 5
	sent := false
	for sent == false {
		indexerConnMutex.Lock()
		if indexerConn[indexerID] != nil {
			_, err = indexerConn[indexerID].Write(append(b, idBytes...))
		} else {
			err = errors.New("indexerConn[indexerID] == nil")
		}
		indexerConnMutex.Unlock()
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialIndexer(indexerID)
				if err != nil {
					logrus.WithField("attempt", cnt).Warning("couldn't connect to indexer, retrying...")
				} else {
					logrus.WithField("id", indexerID).Info("connected to indexer")
				}
			} else {
				logrus.WithField("id", indexerID).Error("failed to connect to indexer after retrying 5 times")
				break
			}
		} else {
			sent = true
		}
	}
	if err != nil {
		return indexer.TOIDIndexTableEntry{}, false, err
	}

	buf := make([]byte, 13)
	indexerConnMutex.Lock()
	totalLength, err := connection.Read(indexerConn[indexerID], &buf)
	indexerConnMutex.Unlock()
	if err != nil || totalLength != 13 {
		return indexer.TOIDIndexTableEntry{}, false, err
	}
	if buf[0] == byte(0) {
		return indexer.TOIDIndexTableEntry{}, false, nil
	}
	lid := int(binary.BigEndian.Uint32(buf[1:5]))
	toid := int(binary.BigEndian.Uint32(buf[5:9]))
	host := int(binary.BigEndian.Uint32(buf[9:13]))
	entry := indexer.TOIDIndexTableEntry{
		LId:  lid,
		TOId: toid,
		Host: host,
	}
	return entry, true, nil
}
