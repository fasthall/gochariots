package info

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var batchers []string
var filters []string
var queues []string
var maintainers []string
var indexers []string
var remoteBatcher []string

// StartController starts controller's REST API server on sepcified port
func StartController(port string) {
	router := gin.Default()

	router.POST("/batcher", addBatchers)
	router.GET("/batcher", getBatchers)
	router.POST("/filter", addFilter)
	router.GET("/filter", getFilters)
	router.POST("/queue", addQueue)
	router.GET("/queue", getQueues)
	router.POST("/maintainer", addMaintainer)
	router.GET("/maintainer", getMaintainers)
	router.POST("/remote/batcher", addRemoteBatcher)
	router.GET("/remote/batcher", getRemoteBatcher)

	router.Run(":" + port)
}

func addBatchers(c *gin.Context) {
	batchers = append(batchers, c.Query("host"))
	c.String(http.StatusOK, c.Query("host")+" added\n")
}

func getBatchers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": batchers,
	})
}

func addFilter(c *gin.Context) {
	filters = append(filters, c.Query("host"))
	c.String(http.StatusOK, c.Query("host")+" added\n")
	for i, host := range batchers {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Printf("%s couldn't connect to batchers[%d] %s\n", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(filters)
		if err != nil {
			log.Println(GetName(), "couldn't convert filter list to bytes:", filters)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		b := make([]byte, 5)
		b[4] = byte('f')
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			log.Printf("%s couldn't send filter list to batchers[%d] %s", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		log.Printf("%s successfully informs batchers[%d] about filter list %s\n", GetName(), i, filters)
	}
}

func getFilters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"filters": filters,
	})
}

func addQueue(c *gin.Context) {
	queues = append(queues, c.Query("host"))
	c.String(http.StatusOK, c.Query("host")+" added\n")
	// update "next queue" for each queue
	for i, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Printf("%s couldn't connect to queues[%d] %s\n", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		b := make([]byte, 5)
		b[4] = byte('q')
		binary.BigEndian.PutUint32(b, uint32(len(queues[(i+1)%len(queues)])+1))
		_, err = conn.Write(append(b, []byte(queues[(i+1)%len(queues)])...))
		if err != nil {
			log.Printf("%s couldn't send next queue host to queues[%d] %s", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		log.Printf("%s successfully informs queues[%d] about next queue host %s\n", GetName(), i, queues[(i+1)%len(queues)])
	}
	// update filter about queues
	for i, host := range filters {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Printf("%s couldn't connect to filters[%d] %s\n", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		b := make([]byte, 5)
		b[4] = byte('q')
		binary.BigEndian.PutUint32(b, uint32(len(c.Query("host"))+1))
		_, err = conn.Write(append(b, []byte(c.Query("host"))...))
		if err != nil {
			log.Printf("%s couldn't send new queue host to filters[%d] %s", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		log.Printf("%s successfully informs filters[%d] about new queue host %s\n", GetName(), i, c.Query("host"))
	}
	// update queues' maintainer list
	conn, err := net.Dial("tcp", c.Query("host"))
	if err != nil {
		log.Println(GetName(), "couldn't connect to queue", c.Query("host"))
		log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
	}
	defer conn.Close()
	jsonBytes, err := json.Marshal(maintainers)
	b := make([]byte, 5)
	b[4] = byte('m')
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
	_, err = conn.Write(append(b, jsonBytes...))
	if err != nil {
		log.Printf("%s couldn't send maintainer list to new queue %s", GetName(), c.Query("host"))
		log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
	}
	log.Printf("%s successfully informs new queue about maintainer list %s\n", GetName(), maintainers)
}

func getQueues(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"queues": queues,
	})
}

func addMaintainer(c *gin.Context) {
	if c.Query("host") == "" || c.Query("indexer") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host and $indexer\n")
		return
	}
	maintainers = append(maintainers, c.Query("host"))
	indexers = append(indexers, c.Query("indexer"))

	conn, err := net.Dial("tcp", c.Query("host"))
	if err != nil {
		log.Printf("%s couldn't connect to maintainer %s\n", GetName(), c.Query("host"))
		log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
	}
	b := make([]byte, 5)
	b[4] = byte('i')
	binary.BigEndian.PutUint32(b, uint32(len(c.Query("indexer"))+1))
	_, err = conn.Write(append(b, []byte(c.Query("indexer"))...))
	if err != nil {
		log.Printf("%s couldn't notify maintainer %s its indexer %s", GetName(), c.Query("host"), c.Query("indexer"))
		log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
	}
	conn.Close()
	// update queues' maintainer list
	for i, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Printf("%s couldn't connect to queues[%d] %s\n", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(maintainers)
		b := make([]byte, 5)
		b[4] = byte('m')
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			log.Printf("%s couldn't send maintainer list to queues[%d] %s", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		log.Printf("%s successfully informs new maintainer about maintainer list %s\n", GetName(), maintainers)
	}
	c.String(http.StatusOK, c.Query("host")+" added\n")
}

func getMaintainers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"maintainers": maintainers,
	})
}

func addRemoteBatcher(c *gin.Context) {
	dc, err := strconv.Atoi(c.Query("dc"))
	if err != nil {
		log.Println(GetName(), "received invalid parameter:", c.Query("dc"))
		return
	}
	for len(remoteBatcher) <= dc {
		remoteBatcher = append(remoteBatcher, "")
	}
	remoteBatcher[dc] = c.Query("host")
	for i, host := range maintainers {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Printf("%s couldn't connect to maintainers[%d] %s\n", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(remoteBatcher)
		b := make([]byte, 5)
		b[4] = byte('b')
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			log.Printf("%s couldn't send remoteBatcher to maintainers[%d] %s", GetName(), i, host)
			log.Panicln(GetName(), "failing to update cluster may cause unexpected error")
		}
		log.Printf("%s successfully informs maintainers[%d] about new remote batchers %s\n", GetName(), i, remoteBatcher)
	}
	c.String(http.StatusOK, "remoteBatcher["+c.Query("dc")+"] = "+c.Query("host")+" updated\n")
}

func getRemoteBatcher(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": remoteBatcher,
	})
}
