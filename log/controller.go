package log

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

var maintainers []string
var queues []string
var filters []string

func StartController(port string) {
	router := gin.Default()
	router.POST("/maintainer", addMaintainer)
	router.GET("/maintainer", getMaintainers)
	router.POST("/queue", addQueue)
	router.GET("/queue", getQueues)
	router.POST("/filter", addFilter)
	router.GET("/filter", getFilters)
	router.Run(":" + port)
}

func addMaintainer(c *gin.Context) {
	maintainers = append(maintainers, c.Query("host"))
	c.String(http.StatusOK, c.Query("host")+" added\n")
}

func getMaintainers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"maintainers": maintainers,
	})
}

func addQueue(c *gin.Context) {
	queues = append(queues, c.Query("host"))
	c.String(http.StatusOK, c.Query("host")+" added\n")
	for i, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			fmt.Println("Couldn't connect to queue", host)
			panic(err)
		}
		defer conn.Close()
		b := []byte{'q'}
		conn.Write(append(b, []byte(queues[(i+1)%len(queues)])...))
		fmt.Println("update ", host, "next queue")
	}
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
}

func getQueues(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"queues": queues,
	})
}

func addFilter(c *gin.Context) {
	filters = append(filters, c.Query("host"))
	c.String(http.StatusOK, c.Query("host")+" added\n")
}

func getFilters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"filters": filters,
	})
}
