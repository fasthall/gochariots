package connection

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/Sirupsen/logrus"
)

func Read(conn net.Conn, buf *[]byte) (int, error) {
	lenbuf := make([]byte, 4)
	remain := 4
	head := 0
	for remain > 0 {
		l, err := conn.Read(lenbuf[head : head+remain])
		if err != nil {
			return 0, err
		}
		remain -= l
		head += l
	}
	if remain != 0 {
		return 0, errors.New("Incorrect length")
	}
	totalLength := int(binary.BigEndian.Uint32(lenbuf))
	if totalLength > cap(*buf) {
		*buf = make([]byte, totalLength)
		logrus.WithFields(logrus.Fields{"old": cap(*buf), "new": totalLength}).Warning("buffer is not large enough, allocate more")
	}
	remain = totalLength
	head = 0
	for remain > 0 {
		l, err := conn.Read((*buf)[head : head+remain])
		if err != nil {
			return 0, err
		}
		remain -= l
		head += l
	}
	if remain != 0 {
		return 0, errors.New("Couldn't read incoming request")
	}
	return totalLength, nil
}
