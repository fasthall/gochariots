package misc

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type Benchmark struct {
	lastCount       int
	recordCount     int
	loggingAccuracy int
	mutex           sync.Mutex
}

func NewBenchmark(accuracy int) Benchmark {
	b := Benchmark{
		lastCount:       0,
		loggingAccuracy: accuracy,
	}
	return b
}

func (b *Benchmark) Logging(incoming int) {
	b.mutex.Lock()
	b.recordCount += incoming
	if b.loggingAccuracy > 0 && b.recordCount-b.lastCount >= b.loggingAccuracy {
		logrus.WithFields(logrus.Fields{"records": b.recordCount, "timestamp": time.Now()}).Info("benchmark logging")
		b.lastCount = b.recordCount
	}
	b.mutex.Unlock()
}
