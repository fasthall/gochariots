package info

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
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
			logrus.WithField("id", i).Error("couldn't connect to batcher")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(filters)
		if err != nil {
			logrus.WithField("filters", filters).Error("couldn't convert filter list to bytes")
			panic("failing to update cluster may cause unexpected error")
		}
		b := make([]byte, 5)
		b[4] = byte('f')
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send filter list to batcher")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"id": i, "filters": filters}).Info("successfully informs batcher about filter list")
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
			logrus.WithField("id", i).Error("couldn't connect to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		b := make([]byte, 5)
		b[4] = byte('q')
		binary.BigEndian.PutUint32(b, uint32(len(queues[(i+1)%len(queues)])+1))
		_, err = conn.Write(append(b, []byte(queues[(i+1)%len(queues)])...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send next queue host")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"id": i, "queue": queues[(i+1)%len(queues)]}).Info("successfully informs queue about next queue host")
	}
	// update filter about queues
	for i, host := range filters {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't connect to filter")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		b := make([]byte, 5)
		b[4] = byte('q')
		binary.BigEndian.PutUint32(b, uint32(len(c.Query("host"))+1))
		_, err = conn.Write(append(b, []byte(c.Query("host"))...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send new queue host")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"id": i, "host": c.Query("host")}).Info("successfully informs filter about new queue host")
	}
	// update queues' maintainer list
	conn, err := net.Dial("tcp", c.Query("host"))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't connect to queue")
		panic("failing to update cluster may cause unexpected error")
	}
	defer conn.Close()
	jsonBytes, err := json.Marshal(maintainers)
	b := make([]byte, 5)
	b[4] = byte('m')
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
	_, err = conn.Write(append(b, jsonBytes...))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send maintainer list to new queue")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("maintainers", maintainers).Info("successfully informs new queue about maintainer list")
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
		logrus.WithField("host", c.Query("host")).Error("couldn't connect to maintainer")
		panic("failing to update cluster may cause unexpected error")
	}
	b := make([]byte, 5)
	b[4] = byte('i')
	binary.BigEndian.PutUint32(b, uint32(len(c.Query("indexer"))+1))
	_, err = conn.Write(append(b, []byte(c.Query("indexer"))...))
	if err != nil {
		logrus.WithFields(logrus.Fields{"maintainer": c.Query("host"), "indexer": c.Query("indexer")}).Error("couldn't notify maintainer its indexer")
		panic("failing to update cluster may cause unexpected error")
	}
	conn.Close()
	// update queues' maintainer list
	for i, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't connect to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(maintainers)
		b := make([]byte, 5)
		b[4] = byte('m')
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send maintainer list to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("maintainers", maintainers).Info("successfully informs queues about maintainer list")
	}
	// update queues' indexer list
	for i, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't connect to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(indexers)
		b := make([]byte, 5)
		b[4] = byte('i')
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send indexer list to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("indexers", indexers).Info("successfully informs queues about indexer list")
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
		logrus.WithField("parameter", c.Query("dc")).Warning("received invalid parameter")
		c.String(http.StatusBadRequest, "invalid parameter")
		return
	}
	for len(remoteBatcher) <= dc {
		remoteBatcher = append(remoteBatcher, "")
	}
	remoteBatcher[dc] = c.Query("host")
	for i, host := range maintainers {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't connect to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(remoteBatcher)
		b := make([]byte, 5)
		b[4] = byte('b')
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+1))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send remoteBatcher to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("batchers", remoteBatcher).Info("successfully informs maintainers about new remote batchers")
	}
	c.String(http.StatusOK, "remoteBatcher["+c.Query("dc")+"] = "+c.Query("host")+" updated\n")
}

func getRemoteBatcher(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": remoteBatcher,
	})
}
