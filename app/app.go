package app

import (
	"fmt"
	"math/rand"
	"net"

	"strconv"

	"net/http"

	"github.com/fasthall/gochariots/info"
	"github.com/fasthall/gochariots/log"
	"github.com/gin-gonic/gin"
)

var batcherPool []string

type JsonRecord struct {
	Tags    map[string]string `json:"tags"`
	PreHost int               `json:"prehost"`
	PreTOId int               `json:"pretoid"`
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
	c.String(http.StatusOK, c.Query("host")+" added\n")
}

func getBatchers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": batcherPool,
	})
}

func postRecord(c *gin.Context) {
	var jsonRecord JsonRecord
	err := c.Bind(&jsonRecord)
	if err != nil {
		panic(err)
	}

	// send to batcher
	host := randomHost()
	fmt.Println("POST ", jsonRecord, " to ", host)
	record := log.Record{
		Host: info.ID,
		Tags: jsonRecord.Tags,
		Pre: log.Causality{
			Host: jsonRecord.PreHost,
			TOId: jsonRecord.PreTOId,
		},
	}
	b := []byte{'r'}
	jsonBytes, err := log.ToJSON(record)
	if err != nil {
		panic(err)
	}

	conn, _ := net.Dial("tcp", host)
	defer conn.Close()
	conn.Write(append(b, jsonBytes...))
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
	record, err := log.ReadByLId(lid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Record not found",
			"error":   err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"LId":       record.LId,
		"Host":      record.Host,
		"TOId":      record.TOId,
		"Causality": record.Pre,
		"Tags":      record.Tags,
	})
}

func randomHost() string {
	return batcherPool[rand.Intn(len(batcherPool))]
}
