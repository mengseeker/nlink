package connpool

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/mengseeker/nlink/core/log"
)

type TCPConnectPool struct {
	Name      string
	container chan net.Conn
	address   string
	tlsConfig *tls.Config
	done      chan any
	log       *log.Logger
	once      sync.Once
}

func NewTCPConnectPool(name, address string, tlsc *tls.Config, size int, l *log.Logger) (pl *TCPConnectPool, err error) {
	pl = &TCPConnectPool{
		Name:      name,
		container: make(chan net.Conn, size),
		address:   address,
		tlsConfig: tlsc,
		done:      make(chan any),
		log:       l,
		once:      sync.Once{},
	}
	return
}

func (p *TCPConnectPool) Get(ctx context.Context) (conn net.Conn, err error) {
	p.once.Do(func() {
		go p.handleNewResource()
	})
	select {
	case <-ctx.Done():
		return nil, errors.New("timeout to get conn")
	case conn = <-p.container:
		return conn, nil
	}
}

func (p *TCPConnectPool) Release() {
	p.done <- nil
	close(p.done)
	close(p.container)
	for conn := range p.container {
		conn.Close()
	}
}

func (p *TCPConnectPool) handleNewResource() {
	for {
		select {
		case <-p.done:
			return
		default:
			conn, err := tls.Dial("tcp", p.address, p.tlsConfig)
			if err != nil {
				p.log.Errorf("dial tcp server %q, err: %v", p.address, err)
				time.Sleep(time.Second * 30)
				continue
			}
			p.container <- conn
		}
	}
}
