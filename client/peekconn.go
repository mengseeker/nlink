package client

import (
	"io"
	"net"
)

type PeekConn struct {
	net.Conn
	peeks []byte
}

func NewPeekConn(c net.Conn) *PeekConn {
	return &PeekConn{Conn: c, peeks: make([]byte, 0)}
}

// Peek returns the next n bytes without advancing the reader.
func (c *PeekConn) Peek(n int) ([]byte, error) {
	if len(c.peeks) < n {
		start := len(c.peeks)
		if cap(c.peeks) < n {
			np := make([]byte, n)
			copy(np, c.peeks)
			c.peeks = np
		}
		_, err := io.ReadFull(c.Conn, c.peeks[start:n])
		if err != nil {
			return c.peeks, err
		}
	}
	return c.peeks[:n], nil
}

func (c *PeekConn) Read(p []byte) (int, error) {
	if len(c.peeks) == 0 {
		return c.Conn.Read(p)
	}
	copyd := copy(p, c.peeks)
	if len(c.peeks) <= len(p) {
		c.peeks = nil
		n, err := c.Conn.Read(p[copyd:])
		return copyd + n, err
	}
	return copyd, nil
}
