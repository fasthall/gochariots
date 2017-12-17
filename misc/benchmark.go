package misc

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type Benchmark struct {
	start           time.Time
	lastCount       int
	recordCount     int
	loggingAccuracy int
	mutex           sync.Mutex
}

func NewBenchmark(accuracy int) Benchmark {
	b := Benchmark{
		lastCount:       0,
		loggingAccuracy: accuracy,
		start:           time.Time{},
	}
	return b
}

func (b *Benchmark) Logging(incoming int) {
	b.mutex.Lock()
	if b.start.Equal(time.Time{}) {
		b.start = time.Now()
	}
	b.recordCount += incoming
	if b.loggingAccuracy > 0 && b.recordCount-b.lastCount >= b.loggingAccuracy {
		throughput := float64(b.recordCount) / time.Since(b.start).Seconds()
		logrus.WithFields(logrus.Fields{"throughput": throughput, "records": b.recordCount, "timestamp": time.Now()}).Info("benchmark logging")
		b.lastCount = b.recordCount
	}
	b.mutex.Unlock()
}
