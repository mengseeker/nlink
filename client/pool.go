package client

import (
	"net"
	"time"

	"github.com/mengseeker/nlink/core/transform"
	"go.uber.org/zap"
)

type Conn interface {
	net.Conn

	// CloseWrite shuts down the writing side of the connection.
	CloseWrite() error

	// Bind binds the connection to the given address
	//
	// Binding the connection allows to receive data from the address
	// without sending data to it.
	Bind(*transform.Meta) error

	// Reset should reset the connection to be reused
	//
	// If keepAlive is true, the connection will be kept open and re-used for
	// other requests
	//
	// If keepAlive is false, the connection will be closed after this request
	// and cannot be re-used for other requests
	Reset() error

	// Disconnect closes the connection and cleans up any resources
	// must be called when the connection is no longer needed
	Disconnect(reason string) error
}

// use for http proxy client
type httpConn struct {
	Conn
	pl *ConnPool
}

func (hc *httpConn) Close() error {
	hc.Conn.Close()
	hc.pl.Put(hc.Conn)
	return nil
}

type ConnPool struct {
	// Dialer is used to create a new connection to server
	Dialer func() (Conn, error)

	// max conns
	MaxConns int

	// max idls conns
	MaxIdle int

	// conn timeout and remove from pool
	IdleTimeout time.Duration

	conns   chan Conn
	putChan chan *putBackConn

	disconnectNum int
	dialServerNum int
	dialRemoteNum int
	recoverNum    int
}

const (
	DefaultMaxConns    = 200
	DefaultIdleTimeout = 10 * time.Minute
	// DefaultIdleTimeout = 10 * time.Second
)

type PoolConfig struct {
	MaxConns    int
	IdleTimeout time.Duration
}

func NewConnPool(cfg PoolConfig, dialer func() (Conn, error)) *ConnPool {
	if cfg.MaxConns == 0 {
		cfg.MaxConns = DefaultMaxConns
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = DefaultIdleTimeout
	}

	pl := &ConnPool{
		MaxConns:    cfg.MaxConns,
		IdleTimeout: cfg.IdleTimeout,
		conns:       make(chan Conn),
		putChan:     make(chan *putBackConn, DefaultMaxConns),
	}

	pl.Dialer = func() (Conn, error) {
		pl.dialServerNum++
		return dialer()
	}

	go pl.handlePut()
	return pl
}

func (p *ConnPool) ConnCount() int {
	return len(p.putChan)
}

func (p *ConnPool) get() (Conn, error) {
	select {
	case conn := <-p.conns:
		return conn, nil
	default:
		return p.Dialer()
	}
}

func (p *ConnPool) DialRemote(remote *transform.Meta) (Conn, error) {
	conn, err := p.get()
	if err != nil {
		return nil, err
	}
	if err := conn.Bind(remote); err != nil {
		p.DisconnectConn(conn, "bind error")
		return nil, err
	}

	p.dialRemoteNum++
	logger.Infof("dial remote %s", remote.String())

	logger.With(
		zap.Int("idle_conn", len(p.putChan)),
		zap.Int("dial_remote", p.dialRemoteNum),
		zap.Int("dial_server", p.dialServerNum),
		zap.Int("recover", p.recoverNum),
		zap.Int("disconnect", p.disconnectNum),
		zap.Int("used_conn", p.dialServerNum-p.disconnectNum),
	).Debug("dial status")
	return conn, err
}

type putBackConn struct {
	conn    Conn
	lastUse time.Time
}

func (p *ConnPool) handlePut() {
	tm := time.NewTimer(p.IdleTimeout)
	var leftTime time.Duration
	for {
		conn := <-p.putChan
		leftTime = time.Until(conn.lastUse.Add(p.IdleTimeout))
		if leftTime <= time.Second {
			p.DisconnectConn(conn.conn, "idle timeout")
			continue
		}

		tm.Reset(leftTime)

		select {
		case p.conns <- conn.conn:
			continue
		case <-tm.C:
			p.DisconnectConn(conn.conn, "idle timeout")
		}
	}
}

func (p *ConnPool) Put(conn Conn) {
	p.recoverNum++
	if err := conn.Reset(); err != nil {
		p.DisconnectConn(conn, "reset error")
		return
	}

	select {
	case p.putChan <- &putBackConn{conn: conn, lastUse: time.Now()}:
		return
	default:
		p.DisconnectConn(conn, "pool is full")
	}
}

func (p *ConnPool) DisconnectConn(conn Conn, reason string) {
	p.disconnectNum++
	conn.Disconnect(reason)
}
