package misc

import (
	"errors"
	"net"
)

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
