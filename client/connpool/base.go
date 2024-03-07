package connpool

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/mengseeker/nlink/core/log"
)

type ConnPool interface {
	Get(ctx context.Context) (conn net.Conn, err error)
	Release()
}

const (
	NET_TCP  = "tcp"
	NET_KCP  = "udp-kcp"
	NET_QUIC = "udp"

	PoolSize = 10
)

func NewConnPool(name, addr, cert, key, net string, l *log.Logger) (ConnPool, error) {
	tlsc, err := NewClientTls(cert, key)
	if err != nil {
		return nil, err
	}
	var pool ConnPool
	switch net {
	case NET_TCP:
		pool, err = NewTCPConnectPool(name, addr, tlsc, PoolSize, l)
	case NET_QUIC:
		pool, err = NewQUICConnectPool(name, addr, tlsc, PoolSize, l)
	case NET_KCP:
		pool, err = NewKCPConnectPool(name, addr, tlsc, PoolSize, l)
	default:
		pool, err = NewTCPConnectPool(name, addr, tlsc, PoolSize, l)
	}

	return pool, err
}

func NewClientTls(certFile, keyFile string) (tc *tls.Config, err error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	tc = &tls.Config{
		ServerName:         "x.test.example.com",
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	return
}
