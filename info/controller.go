package info

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fasthall/gochariots/misc"
	"github.com/gin-gonic/gin"
)

var apps []string
var appsVersion int
var batchers []string
var batchersVersion int
var filters []string
var filtersVersion int
var queues []string
var queuesVersion int
var maintainers []string
var maintainersVersion int
var indexers []string
var indexersVersion int
var remoteBatcher []string
var remoteBatcherVer int
var mutex sync.Mutex

// StartController starts controller's REST API server on sepcified port
func StartController(port string) {
	router := gin.Default()

	router.GET("/", getInfo)
	router.POST("/app", addApps)
	router.GET("/app", getApps)
	router.POST("/batcher", addBatchers)
	router.GET("/batcher", getBatchers)
	router.POST("/filter", addFilter)
	router.GET("/filter", getFilters)
	router.POST("/queue", addQueue)
	router.GET("/queue", getQueues)
	router.POST("/maintainer", addMaintainer)
	router.GET("/maintainer", getMaintainers)
	router.POST("/indexer", addIndexer)
	router.GET("/indexer", getIndexers)
	router.POST("/remote/batcher", addRemoteBatcher)
	router.GET("/remote/batcher", getRemoteBatcher)

	router.Run(":" + port)
}

func getInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"apps":           apps,
		"batchers":       batchers,
		"filters":        filters,
		"queues":         queues,
		"maintainers":    maintainers,
		"indexers":       indexers,
		"remoteBatchers": remoteBatcher,
	})
}

func addApps(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	apps = append(apps, c.Query("host"))
	appsVersion++
	mutex.Unlock()

	jsonBatchers, err := json.Marshal(batchers)
	if err != nil {
		logrus.WithError(err).Error("couldn't convert batchers to bytes")
		panic("failing to update cluster may cause unexpected error")
	}
	p := misc.NewParams()
	p.AddParam("host", string(jsonBatchers))
	p.AddParam("ver", strconv.Itoa(batchersVersion))
	misc.Report(c.Query("host"), "batcher", p)
	logrus.WithField("host", c.Query("host")).Info("successfully informed app about batcher list")

	c.String(http.StatusOK, c.Query("host")+" added")
}

func getApps(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"apps": apps,
	})
}

func addBatchers(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	batchers = append(batchers, c.Query("host"))
	batchersVersion++
	mutex.Unlock()
	for _, host := range apps {
		jsonBatchers, err := json.Marshal(batchers)
		if err != nil {
			logrus.WithError(err).Error("couldn't convert batchers to bytes")
			panic("failing to update cluster may cause unexpected error")
		}
		p := misc.NewParams()
		p.AddParam("host", string(jsonBatchers))
		p.AddParam("ver", strconv.Itoa(batchersVersion))
		misc.Report(host, "batcher", p)
		logrus.WithField("host", host).Info("successfully informed app about batcher list")
	}
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
		b := make([]byte, 9)
		b[4] = byte('f')
		binary.BigEndian.PutUint32(b[5:], uint32(filtersVersion))
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send filter list to batcher")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"id": i, "filters": filters}).Info("successfully informed batcher about filter list")
	}
	c.String(http.StatusOK, c.Query("host")+" added")
}

func getBatchers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": batchers,
	})
}

func addFilter(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	filters = append(filters, c.Query("host"))
	filtersVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")
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
		b := make([]byte, 9)
		b[4] = byte('f')
		binary.BigEndian.PutUint32(b[5:], uint32(filtersVersion))
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send filter list to batcher")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"id": i, "filters": filters}).Info("successfully informed batcher about filter list")
	}

	// update filter about queues
	conn, err := net.Dial("tcp", c.Query("host"))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't connect to filter")
		panic("failing to update cluster may cause unexpected error")
	}
	defer conn.Close()
	b := make([]byte, 9)
	b[4] = byte('q')
	jsonBytes, err := json.Marshal(queues)
	binary.BigEndian.PutUint32(b[5:], uint32(queuesVersion))
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
	_, err = conn.Write(append(b, []byte(jsonBytes)...))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send new queue host")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithFields(logrus.Fields{"filter": c.Query("host"), "queues": queues}).Info("successfully informed filter about new queue host")
}

func getFilters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"filters": filters,
	})
}

func addQueue(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	queues = append(queues, c.Query("host"))
	queuesVersion++
	mutex.Unlock()

	c.String(http.StatusOK, c.Query("host")+" added")
	// update "next queue" for each queue
	for i, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't connect to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		b := make([]byte, 9)
		b[4] = byte('q')
		binary.BigEndian.PutUint32(b[5:], uint32(queuesVersion))
		binary.BigEndian.PutUint32(b, uint32(len(queues[(i+1)%len(queues)])+5))
		_, err = conn.Write(append(b, []byte(queues[(i+1)%len(queues)])...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send next queue host")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"id": i, "queue": queues[(i+1)%len(queues)]}).Info("successfully informed queue about next queue host")
	}
	// update filter about queues
	for i, host := range filters {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't connect to filter")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		b := make([]byte, 9)
		b[4] = byte('q')
		jsonBytes, err := json.Marshal(queues)
		binary.BigEndian.PutUint32(b[5:], uint32(queuesVersion))
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
		_, err = conn.Write(append(b, []byte(jsonBytes)...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send new queue host")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"filter": host, "queues": queues}).Info("successfully informed filter about new queue host")
	}

	// update queues' maintainer list
	conn, err := net.Dial("tcp", c.Query("host"))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't connect to queue")
		panic("failing to update cluster may cause unexpected error")
	}
	jsonBytes, err := json.Marshal(maintainers)
	b := make([]byte, 9)
	b[4] = byte('m')
	binary.BigEndian.PutUint32(b[5:], uint32(maintainersVersion))
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
	_, err = conn.Write(append(b, jsonBytes...))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send maintainer list to new queue")
		panic("failing to update cluster may cause unexpected error")
	}
	conn.Close()
	logrus.WithField("maintainers", maintainers).Info("successfully informed new queue about maintainer list")

	// tell queue about indexer
	conn, err = net.Dial("tcp", c.Query("host"))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't connect to queue")
		panic("failing to update cluster may cause unexpected error")
	}
	defer conn.Close()
	jsonBytes, err = json.Marshal(indexers)
	if err != nil {
		logrus.WithError(err).Error("couldn't convert indexer list to json bytes")
		panic(err)
	}
	b = make([]byte, 9)
	b[4] = byte('i')
	binary.BigEndian.PutUint32(b[5:], uint32(indexersVersion))
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
	_, err = conn.Write(append(b, jsonBytes...))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't send indexer list to queue")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("indexers", indexers).Info("successfully informed queue about indexer list")
}

func getQueues(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"queues": queues,
	})
}

func addMaintainer(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	maintainers = append(maintainers, c.Query("host"))
	maintainersVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

	// tell maintainer its indexer
	i := len(maintainers) - 1
	if i < len(indexers) {
		conn, err := net.Dial("tcp", c.Query("host"))
		defer conn.Close()
		if err != nil {
			logrus.WithField("host", c.Query("host")).Error("couldn't connect to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		b := make([]byte, 9)
		b[4] = byte('i')
		binary.BigEndian.PutUint32(b[5:], uint32(indexersVersion))
		binary.BigEndian.PutUint32(b, uint32(len(indexers[i])+5))
		_, err = conn.Write(append(b, []byte(indexers[i])...))
		if err != nil {
			logrus.WithFields(logrus.Fields{"maintainer": c.Query("host"), "indexer": indexers[i]}).Error("couldn't notify maintainer its indexer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"maintainer": c.Query("host"), "indexer": indexers[i]}).Info("successfully informed maintainer its indexer")
	}

	// update queues' maintainer list
	for i, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithField("id", i).Error("couldn't connect to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(maintainers)
		b := make([]byte, 9)
		b[4] = byte('m')
		binary.BigEndian.PutUint32(b[5:], uint32(maintainersVersion))
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send maintainer list to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("maintainers", maintainers).Info("successfully informed queues about maintainer list")
	}

	// update remote batchers
	conn, err := net.Dial("tcp", c.Query("host"))
	if err != nil {
		logrus.WithField("host", c.Query("host")).Error("couldn't connect to maintainer")
		panic("failing to update cluster may cause unexpected error")
	}
	defer conn.Close()
	jsonBytes, err := json.Marshal(remoteBatcher)
	b := make([]byte, 9)
	b[4] = byte('b')
	binary.BigEndian.PutUint32(b[5:], uint32(remoteBatcherVer))
	binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
	_, err = conn.Write(append(b, jsonBytes...))
	if err != nil {
		logrus.WithField("id", i).Error("couldn't send remoteBatcher to maintainer")
		panic("failing to update cluster may cause unexpected error")
	}
	logrus.WithField("batchers", remoteBatcher).Info("successfully informed maintainers about new remote batchers")
}

func getMaintainers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"maintainers": maintainers,
	})
}

func addIndexer(c *gin.Context) {
	if c.Query("host") == "" {
		c.String(http.StatusBadRequest, "invalid parameter, needs $host")
		return
	}
	mutex.Lock()
	indexers = append(indexers, c.Query("host"))
	indexersVersion++
	mutex.Unlock()
	c.String(http.StatusOK, c.Query("host")+" added")

	// tell maintainer its indexer
	i := len(indexers) - 1
	if i < len(maintainers) {
		maintainer := maintainers[i]
		conn, err := net.Dial("tcp", maintainer)
		defer conn.Close()
		if err != nil {
			logrus.WithField("host", maintainer).Error("couldn't connect to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		b := make([]byte, 9)
		b[4] = byte('i')
		binary.BigEndian.PutUint32(b[5:], uint32(indexersVersion))
		binary.BigEndian.PutUint32(b, uint32(len(indexers[i])+5))
		_, err = conn.Write(append(b, []byte(indexers[i])...))
		if err != nil {
			logrus.WithFields(logrus.Fields{"maintainer": c.Query("host"), "indexer": indexers[i]}).Error("couldn't notify maintainer its indexer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithFields(logrus.Fields{"maintainer": maintainer, "indexer": indexers[i]}).Info("successfully informed maintainer its indexer")
	}

	// update queues' indexer list
	for _, host := range queues {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithField("host", host).Error("couldn't connect to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(indexers)
		if err != nil {
			logrus.WithError(err).Error("couldn't convert indexer list to json bytes")
			panic(err)
		}
		b := make([]byte, 9)
		b[4] = byte('i')
		binary.BigEndian.PutUint32(b[5:], uint32(indexersVersion))
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("host", host).Error("couldn't send indexer list to queue")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("indexers", indexers).Info("successfully informed queues about indexer list")
	}
}

func getIndexers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"indexers": indexers,
	})
}

func addRemoteBatcher(c *gin.Context) {
	dc, err := strconv.Atoi(c.Query("dc"))
	if err != nil {
		logrus.WithField("parameter", c.Query("dc")).Warning("received invalid parameter")
		c.String(http.StatusBadRequest, "invalid parameter")
		return
	}
	mutex.Lock()
	for len(remoteBatcher) <= dc {
		remoteBatcher = append(remoteBatcher, "")
	}
	remoteBatcher[dc] = c.Query("host")
	remoteBatcherVer++
	mutex.Unlock()

	for i, host := range maintainers {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logrus.WithFields(logrus.Fields{"id": i, "host": host}).Error("couldn't connect to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		defer conn.Close()
		jsonBytes, err := json.Marshal(remoteBatcher)
		b := make([]byte, 9)
		b[4] = byte('b')
		binary.BigEndian.PutUint32(b[5:], uint32(remoteBatcherVer))
		binary.BigEndian.PutUint32(b, uint32(len(jsonBytes)+5))
		_, err = conn.Write(append(b, jsonBytes...))
		if err != nil {
			logrus.WithField("id", i).Error("couldn't send remoteBatcher to maintainer")
			panic("failing to update cluster may cause unexpected error")
		}
		logrus.WithField("batchers", remoteBatcher).Info("successfully informed maintainers about new remote batchers")
	}
	c.String(http.StatusOK, "remoteBatcher["+c.Query("dc")+"] = "+c.Query("host")+" updated")
}

func getRemoteBatcher(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"batchers": remoteBatcher,
	})
}
