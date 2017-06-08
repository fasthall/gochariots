package app

import (
	"encoding/binary"
	"errors"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/maintainer"
	"github.com/fasthall/gochariots/record"
	"github.com/gin-gonic/gin"
)

var batcherConn []net.Conn
var batcherPool []string
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
	router.POST("/batcher", addBatcher)
	router.GET("/batcher", getBatchers)
	router.Run(":" + port)
}

func addBatcher(c *gin.Context) {
	batcherPool = append(batcherPool, c.Query("host"))
	batcherConn = make([]net.Conn, len(batcherPool))
	c.String(http.StatusOK, c.Query("host")+" added\n")
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
	info.LogTimestamp("postRecord")
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
		log.Panicln(info.GetName(), "couldn't convert record to bytes")
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
			log.Printf("%s couldn't connect to batcherPool[%d] %s\n", info.GetName(), hostID, batcherPool[hostID])
		} else {
			log.Printf("%s is connected to batcherPool[%d] %s\n", info.GetName(), hostID, batcherPool[hostID])
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
		info.LogTimestamp("sendToBatcher")
		if err != nil {
			if cnt >= 0 {
				cnt--
				err = dialConn(hostID)
				if err != nil {
					log.Printf("%s couldn't connect to the batcherPool[%d] %s, retrying...", info.GetName(), hostID, batcherPool[hostID])
				} else {
					log.Printf("%s is connected to batcherPool[%d] %s\n", info.GetName(), hostID, batcherPool[hostID])
				}
			} else {
				log.Printf("%s failed to connect to the batcherPool[%d] %s after retrying 5 times", info.GetName(), hostID, batcherPool[hostID])
				c.String(http.StatusServiceUnavailable, "Couldn't connect to the batcher")
				return
			}
		} else {
			sent = true
			log.Printf("%s sent to batcherPool[%d] %s\n", info.GetName(), hostID, batcherPool[hostID])
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
