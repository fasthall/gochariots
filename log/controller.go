package log

import (
	"fmt"
	"net"
	"net/http"

	"encoding/json"

	"github.com/gin-gonic/gin"
)

var batchers []string
var filters []string
var queues []string
var maintainers []string

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
	for _, host := range batchers {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			fmt.Println("Couldn't connect to batcher", host)
			panic(err)
		}
		defer conn.Close()
		b := []byte{'f'}
		jsonBytes, err := json.Marshal(filters)
		if err != nil {
			fmt.Println("Couldn't convert filter list to bytes")
			panic(err)
		}
		conn.Write(append(b, jsonBytes...))
		fmt.Println("update ", host, "filters")
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
			fmt.Println("Couldn't connect to queue", host)
			panic(err)
		}
		defer conn.Close()
		b := []byte{'q'}
		conn.Write(append(b, []byte(queues[(i+1)%len(queues)])...))
		fmt.Println("update", host, "next queue")
	}
	// update filter about queues
	for _, host := range filters {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			fmt.Println("Couldn't connect to filter", host)
			panic(err)
		}
		defer conn.Close()
		b := []byte{'q'}
		conn.Write(append(b, []byte(c.Query("host"))...))
		fmt.Println("inform", host, "about queue", c.Query("host"))
	}
	// update queues' maintainer list
	conn, err := net.Dial("tcp", c.Query("host"))
	if err != nil {
		fmt.Println("Couldn't connect to queue", c.Query("host"))
		panic(err)
	}
	defer conn.Close()
	b := []byte{'m'}
	jsonBytes, err := json.Marshal(maintainers)
	conn.Write(append(b, jsonBytes...))
	fmt.Println("update ", c.Query("host"), "about maintainer")
}

func getQueues(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"queues": queues,
	})
}

func addMaintainer(c *gin.Context) {
	maintainers = append(maintainers, c.Query("host"))
	for _, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			fmt.Println("Couldn't connect to queue", host)
			panic(err)
		}
		defer conn.Close()
		b := []byte{'m'}
		jsonBytes, err := json.Marshal(maintainers)
		conn.Write(append(b, jsonBytes...))
		fmt.Println("update ", host, "about maintainer")
	}
	c.String(http.StatusOK, c.Query("host")+" added\n")
}

func getMaintainers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"maintainers": maintainers,
	})
}
