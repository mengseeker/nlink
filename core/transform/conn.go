package transform

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/mengseeker/nlink/core/log"
)

func ConnCopyAndWait(c1, c2 net.Conn, l *log.Logger) {
	wg := sync.WaitGroup{}
	copy := func(w, r net.Conn) {
		defer wg.Done()
		_, err := io.Copy(w, r)
		if err != nil {
			l.Errorf("copy data err: %v", err)
		}

		ConnCloseRead(r)
		ConnCloseWrite(w)
		l.Debugf("copy done %s -> %s", r.RemoteAddr(), w.RemoteAddr())
	}
	wg.Add(2)
	go copy(c1, c2)
	go copy(c2, c1)
	wg.Wait()
}

func ConnCloseRead(r net.Conn) {
	if c, ok := r.(*net.TCPConn); ok {
		c.CloseRead()
	} else if c, ok := r.(QUICConn); ok {
		c.CancelRead(0)
	} else if c, ok := r.(*tls.Conn); ok {
		c.NetConn().(*net.TCPConn).CloseRead()
	} else if c, ok := r.(*PeekConn); ok {
		ConnCloseRead(c.Conn)
	} else {
		fmt.Printf("----------%#v\n", r)
	}
}

func ConnCloseWrite(w net.Conn) {
	if c, ok := w.(*net.TCPConn); ok {
		c.CloseWrite()
	} else if c, ok := w.(QUICConn); ok {
		c.Close()
	} else if c, ok := w.(*tls.Conn); ok {
		c.NetConn().(*net.TCPConn).CloseWrite()
	} else if c, ok := w.(*PeekConn); ok {
		ConnCloseWrite(c.Conn)
	} else {
		fmt.Printf("----------%#v\n", w)
	}
}
