package client

import (
	"context"
	"crypto/tls"
	"errors"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
	"github.com/quic-go/quic-go"
)

type ConnPool interface {
	Get(ctx context.Context) (conn net.Conn, err error)
	Release()
}

type TCPConnectPool struct {
	Name      string
	container chan net.Conn
	address   string
	tlsConfig *tls.Config
	done      chan any
	log       *log.Logger
	once      sync.Once
}

func NewTCPConnectPool(name, address, cert, key string, size int, l *log.Logger) (pl *TCPConnectPool, err error) {
	tc, err := NewClientTls(cert, key)
	if err != nil {
		return
	}
	pl = &TCPConnectPool{
		Name:      name,
		container: make(chan net.Conn, size),
		address:   address,
		tlsConfig: tc,
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

type UDPConnectPool struct {
	Name      string
	container chan transform.QUICConn
	address   string
	tlsConfig *tls.Config
	done      chan any
	wg        sync.WaitGroup
	log       *log.Logger
	once      sync.Once
}

func NewUDPConnectPool(name, address, cert, key string, size int, l *log.Logger) (pl *UDPConnectPool, err error) {
	tc, err := NewClientTls(cert, key)
	if err != nil {
		return
	}
	pl = &UDPConnectPool{
		Name:      name,
		container: make(chan transform.QUICConn, size),
		address:   address,
		tlsConfig: tc,
		done:      make(chan any),
		wg:        sync.WaitGroup{},
		log:       l,
		once:      sync.Once{},
	}
	return
}

func (p *UDPConnectPool) Get(ctx context.Context) (conn net.Conn, err error) {
	p.once.Do(func() {
		p.wg.Add(cap(p.container))
		for i := 0; i < cap(p.container); i++ {
			go p.handleNewResource()
		}
	})
	select {
	case <-ctx.Done():
		return nil, errors.New("timeout to get conn")
	case conn = <-p.container:
		return conn, nil
	}
}

func (p *UDPConnectPool) Release() {
	close(p.done)
	p.wg.Wait()
	close(p.container)
	for conn := range p.container {
		conn.Close()
	}
}

func (p *UDPConnectPool) handleNewResource() {
	defer p.wg.Done()
	var lastErrTimes time.Duration
	for {
		select {
		case <-p.done:
			return
		default:
			conn, err := p.dial()
			if err != nil {
				p.log.Errorf("dial udp server %s, err: %v", p.address, err)
				lastErrTimes++
				if lastErrTimes >= 10 {
					lastErrTimes = 10
				}
				time.Sleep(time.Second * (lastErrTimes*30 + time.Duration(rand.Intn(30))))
				continue
			}
			p.handleConn(conn)
		}
	}
}

func (p *UDPConnectPool) dial() (conn quic.Connection, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return quic.DialAddr(ctx, p.address, p.tlsConfig, &quic.Config{
		KeepAlivePeriod: time.Second * 3,
	})
}

func (p *UDPConnectPool) handleConn(conn quic.Connection) {
	defer conn.CloseWithError(0, "just close")
	for {
		select {
		case <-p.done:
			return
		default:
			stream, err := conn.OpenStream()
			if err != nil {
				p.log.Errorf("open udp stream err: %v", err)
				return
			}
			p.container <- transform.QUICConn{Stream: stream, Conn: conn}
		}

	}
}
