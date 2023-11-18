package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/socks/transport/socks4"
	"github.com/mengseeker/nlink/core/socks/transport/socks5"
)

type Listener struct {
	Address       string
	Log           *log.Logger
	HTTPHandler   http.Handler
	Socks4Handler SockesHandler
	Socks5Handler SockesHandler
	TunnelHandler any // TODO

	lis         net.Listener
	ctx         context.Context
	wg          sync.WaitGroup
	httpConns   chan net.Conn
	socks4Conns chan net.Conn
	socks5Conns chan net.Conn
}

func (l *Listener) ListenAndServe(ctx context.Context) (err error) {
	l.ctx = ctx
	lis, err := net.Listen("tcp", l.Address)
	if err != nil {
		return fmt.Errorf("listen address %s err: %s", l.Address, err)
	}
	l.lis = lis
	defer lis.Close()
	l.wg = sync.WaitGroup{}
	defer l.wg.Wait()
	if l.HTTPHandler != nil {
		l.wg.Add(1)
		l.httpConns = make(chan net.Conn, 1)
		defer close(l.httpConns)
		go l.handleHTTPServer()
	}
	if l.Socks4Handler != nil {
		l.wg.Add(1)
		l.socks4Conns = make(chan net.Conn, 1)
		defer close(l.socks4Conns)
		go l.handleSocks4()
	}
	if l.Socks5Handler != nil {
		l.wg.Add(1)
		l.socks5Conns = make(chan net.Conn, 1)
		defer close(l.socks5Conns)
		go l.handleSocks5()
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := lis.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				return err
			}
			go l.handleTCPConn(conn)
		}
	}
}

func (l *Listener) handleSocks4() {
	defer l.wg.Done()
	for conn := range l.socks4Conns {
		go l.Socks4Handler.HandleConn(conn)
	}
}

func (l *Listener) handleSocks5() {
	defer l.wg.Done()
	for conn := range l.socks5Conns {
		go l.Socks5Handler.HandleConn(conn)
	}
}

func (l *Listener) handleHTTPServer() {
	defer l.wg.Done()
	if l.HTTPHandler != nil {
		err := http.Serve(&HTTPListener{l}, l.HTTPHandler)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			l.Log.Errorf("unexpected server close: %v", err)
		}
	}
}

func (l *Listener) handleTCPConn(conn net.Conn) {
	conn.(*net.TCPConn).SetKeepAlive(true)

	bufConn := NewPeekConn(conn)
	head, err := bufConn.Peek(1)
	if err != nil {
		return
	}

	switch head[0] {
	case socks4.Version:
		if l.socks4Conns != nil {
			l.socks4Conns <- bufConn
		}
	case socks5.Version:
		if l.socks5Conns != nil {
			l.socks5Conns <- bufConn
		}
	default:
		l.httpConns <- bufConn
	}
}

type HTTPListener struct {
	*Listener
}

func (l *HTTPListener) Accept() (net.Conn, error) {
	c, exist := <-l.httpConns
	if !exist {
		return nil, net.ErrClosed
	}
	return c, nil
}

func (l *HTTPListener) Close() error {
	return l.lis.Close()
}

func (l *HTTPListener) Addr() net.Addr {
	return l.lis.Addr()
}
