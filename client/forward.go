package client

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
)

var (
	logger = log.NewLogger()
)

type Forward interface {
	RuleHandler
}

type ForwardClient struct {
	Config     ServerConfig
	httpClient *http.Client
	conn       *transform.PackConn
	connErr    atomic.Bool
}

func NewForwardClient(sc ServerConfig) (*ForwardClient, error) {
	c := ForwardClient{
		Config: sc,
	}
	err := c.dialServer()
	go c.handlerReconnect()

	return &c, err
}

func (f *ForwardClient) Dial(remote *transform.Meta) (net.Conn, error) {
	logger.Infof("dial remote %s", remote.String())
	conn, err := f.conn.DialStream(remote)
	if err != nil {
		f.connErr.Store(true)
		return nil, err
	}
	return conn, err
}

func (f *ForwardClient) HTTPRequest(w http.ResponseWriter, r *http.Request) {
	resp, err := f.httpClient.Do(r)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = errors.New("remote server close connection")
		}
		ResponseError(w, err)
		return
	}
	defer resp.Body.Close()
	CopyHTTPResponse(w, resp)
}

func (f *ForwardClient) Conn(conn net.Conn, remote *transform.Meta) {
	defer conn.Close()
	l := logger.With("remote", remote.String())
	remoteConn, err := f.Dial(remote)
	if err != nil {
		l.Errorf("connect to remote %s failed: %v", remote.String(), err)
		return
	}
	defer remoteConn.Close()

	transform.TransformConn(conn, remoteConn, l)
}

func (f *ForwardClient) handlerReconnect() {
	tk := time.NewTicker(3 * time.Second)
	for range tk.C {
		if f.connErr.Load() {
			logger.Infof("reconnect server %s", f.Config.Addr)
			f.conn.Close()
			err := f.dialServer()
			if err == nil {
				f.connErr.Store(false)
			}
		}
	}
}

func (f *ForwardClient) dialServer() error {
	tlsConfig, err := NewClientTls(f.Config.Cert, f.Config.Key)
	if err != nil {
		return fmt.Errorf("create tls config err: %v", err)
	}

	conn, err := transform.DialPackConn(f.Config.Addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("connect to server err: %v", err)
	}

	f.conn = conn

	f.httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    1000,
			IdleConnTimeout: 3 * time.Minute,
			Dial: func(network, addr string) (net.Conn, error) {
				return f.Dial(&transform.Meta{
					Network: "tcp",
					Address: addr,
				})
			},
		},
	}

	return nil
}
