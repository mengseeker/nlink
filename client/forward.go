package client

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
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
	pool       *ConnPool
}

func NewForwardClient(config ServerConfig) (*ForwardClient, error) {
	tlsConfig, err := NewClientTls(config.Cert, config.Key)
	if err != nil {
		return nil, fmt.Errorf("create tls config err: %v", err)
	}

	pool := NewConnPool(config.Pool, func() (Conn, error) {
		return transform.DialPackConn(config.Name, config.Addr, tlsConfig)
	})

	c := ForwardClient{
		Config: config,
		pool:   pool,
	}

	c.httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns: 1000,
			// IdleConnTimeout: 1 * time.Second,
			IdleConnTimeout: 3 * time.Minute,
			Dial: func(network, addr string) (net.Conn, error) {
				conn, err := c.Dial(&transform.Meta{
					Net:  "tcp",
					Addr: addr,
				})
				if err != nil {
					logger.Errorf("connect to remote failed: %v", err)
					return nil, err
				}
				return &httpConn{
					Conn: conn,
					pl:   pool,
				}, nil
			},
		},
	}

	return &c, nil
}

func (f *ForwardClient) Dial(remote *transform.Meta) (Conn, error) {
	return f.pool.DialRemote(remote)
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
		l.Errorf("connect to remote failed: %v", err)
		return
	}

	defer f.pool.Put(remoteConn)
	defer remoteConn.Close()

	transform.TransformConn(conn, remoteConn, l)
}
