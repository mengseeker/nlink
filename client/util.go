package client

import (
	"strings"
)

func ParseHost(host string) (domain, port string) {
	if host == "" {
		return
	}
	domain = host
	if host[0] == ':' || host[0] == '[' {
		bs := strings.Split(host, "]")
		if len(bs) < 2 || len(bs[0]) < 2 || len(bs[1]) < 2 {
			return
		}
		domain = strings.TrimPrefix(bs[0], "[")
		port = strings.TrimPrefix(bs[1], ":")
		return
	}
	bs := strings.Split(host, ":")
	domain = bs[0]
	if len(bs) > 1 {
		port = bs[1]
	}
	return
}
