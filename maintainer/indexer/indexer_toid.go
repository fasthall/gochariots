package indexer

import (
	"encoding/binary"
	"io"
	"net"
	"sync"

	"github.com/Sirupsen/logrus"
)

var indexMutex sync.Mutex
var TOIDindexes = make(map[uint64]TOIDIndexTableEntry)

type TOIDIndexTableEntry struct {
	LId  int
	Host int
	TOId int
}

func TOIDInitIndexer(p string) {

}

// HandleRequest handles incoming connection
func TOIDHandleRequest(conn net.Conn) {
	lenbuf := make([]byte, 4)
	buf := make([]byte, 1024*1024*32)
	for {
		remain := 4
		head := 0
		for remain > 0 {
			l, err := conn.Read(lenbuf[head : head+remain])
			if err == io.EOF {
				return
			} else if err != nil {
				logrus.WithError(err).Error("couldn't read incoming request")
				break
			} else {
				remain -= l
				head += l
			}
		}
		if remain != 0 {
			logrus.WithField("remain", remain).Error("couldn't read incoming request length")
			break
		}
		totalLength := int(binary.BigEndian.Uint32(lenbuf))
		if totalLength > cap(buf) {
			logrus.WithFields(logrus.Fields{"old": cap(buf), "new": totalLength}).Warning("buffer is not large enough, allocate more")
			buf = make([]byte, totalLength)
		}
		remain = totalLength
		head = 0
		for remain > 0 {
			l, err := conn.Read(buf[head : head+remain])
			if err == io.EOF {
				return
			} else if err != nil {
				logrus.WithError(err).Error("couldn't read incoming request")
				break
			} else {
				remain -= l
				head += l
			}
		}
		if remain != 0 {
			logrus.WithField("remain", remain).Error("couldn't read incoming request length")
			break
		}
		if buf[0] == 't' { // insert tags into hash table
			id := binary.BigEndian.Uint64(buf[1:9])
			lid := int(binary.BigEndian.Uint32(buf[9:13]))
			toid := int(binary.BigEndian.Uint32(buf[13:17]))
			host := int(binary.BigEndian.Uint32(buf[17:]))
			entry := TOIDIndexTableEntry{
				LId:  lid,
				TOId: toid,
				Host: host,
			}
			indexMutex.Lock()
			TOIDindexes[id] = entry
			indexMutex.Unlock()
		} else if buf[0] == 'q' {
			id := binary.BigEndian.Uint64(buf[1:9])
			indexMutex.Lock()
			entry, ok := TOIDindexes[id]
			indexMutex.Unlock()
			tmp := make([]byte, 13)
			if ok {
				tmp[0] = byte(1)
			} else {
				tmp[0] = byte(0)
			}
			binary.BigEndian.PutUint32(tmp[1:5], uint32(entry.LId))
			binary.BigEndian.PutUint32(tmp[5:9], uint32(entry.TOId))
			binary.BigEndian.PutUint32(tmp[9:13], uint32(entry.Host))
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, 13)
			conn.Write(append(b, tmp...))
		} else {
			logrus.WithField("header", buf[0]).Warning("couldn't understand request")
		}
	}
}
