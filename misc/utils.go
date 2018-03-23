package misc

import (
	"errors"
	"hash/crc32"
	"net"
)

// GetHostIP returns host IP address that eth0 uses
func GetHostIP() (string, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, ifi := range ifs {
		if ifi.Name == "eth0" {
			addrs, err := ifi.Addrs()
			if err == nil {
				for _, addr := range addrs {
					if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
						if ipnet.IP.To4() != nil {
							return ipnet.IP.String(), nil
						}
					}
				}
			}
		}
	}
	return "", errors.New("IP not found")
}

// HashID hashes string using CRC32 and modulate it by size
func HashID(id string, size int) uint32 {
	return crc32.ChecksumIEEE([]byte(id)) % uint32(size)
}
