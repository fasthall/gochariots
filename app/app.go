package app

import (
	"fmt"
	"math/rand"
	"net"

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
	router.POST("/", postRecord)
	router.Run(":" + port)
}

func AddBatcher(host string) {
	batcherPool = append(batcherPool, host)
}

func postRecord(c *gin.Context) {
	var jsonRecord JsonRecord
	err := c.Bind(&jsonRecord)
	if err != nil {
		panic(err)
	}
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
	jsonBytes, err := log.ToJSON(record)
	if err != nil {
		panic(err)
	}

	conn, _ := net.Dial("tcp", host)
	defer conn.Close()
	conn.Write(jsonBytes)
}

func randomHost() string {
	return batcherPool[rand.Intn(len(batcherPool))]
}
