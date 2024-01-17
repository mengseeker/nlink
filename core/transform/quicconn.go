package transform

import (
	"net"

	"github.com/quic-go/quic-go"
)

// net.Conn interface
type QUICConn struct {
	quic.Stream
	Conn quic.Connection
}

func (c QUICConn) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c QUICConn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
