package client

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
)

const (
	ForwardConnPoolSize = 10
)

type Forward interface {
	RuleHandler
}

type ForwardClient struct {
	Log           *log.Logger
	ForwardConfig ServerConfig
	connPool      ConnPool
	cancel        context.CancelFunc
	httpClient    http.Client
	connCount     atomic.Int32
}

func NewForwardClient(ctx context.Context, sc ServerConfig, l *log.Logger) (*ForwardClient, error) {
	ctx, cancel := context.WithCancel(ctx)
	c := ForwardClient{
		Log:           l.With("server.name", sc.Name, "server.net", sc.Net),
		ForwardConfig: sc,
		cancel:        cancel,
	}
	if sc.Net == "tcp" {
		connPool, err := NewTCPConnectPool(sc.Name, sc.Addr, sc.Cert, sc.Key, ForwardConnPoolSize, l)
		if err != nil {
			return nil, err
		}
		c.connPool = connPool
	} else {
		connPool, err := NewUDPConnectPool(sc.Name, sc.Addr, sc.Cert, sc.Key, ForwardConnPoolSize, l)
		if err != nil {
			return nil, err
		}
		c.connPool = connPool
	}
	c.httpClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    1000,
			IdleConnTimeout: 3 * time.Minute,
			Dial: func(network, addr string) (net.Conn, error) {
				return c.Dial(&api.ForwardMeta{
					Network: "tcp",
					Address: addr,
					ID:      uuid.NewString(),
				})
			},
		},
	}
	go func() {
		<-ctx.Done()
		c.connPool.Release()
	}()

	return &c, nil
}

func (f *ForwardClient) Dial(remote *api.ForwardMeta) (net.Conn, error) {
	remote.ID = uuid.NewString()
	f.Log.With("remote.network", remote.Network, "remote.address", remote.Address, "conn.id", remote.ID).Info("dial forward")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	remoteConn, err := f.connPool.Get(ctx)
	if err != nil {
		return nil, errors.New("unable to get remote conn, err: " + err.Error())
	}
	err = transform.SendMsg(remoteConn, remote)
	if err != nil {
		remoteConn.Close()
		return nil, errors.New("send metadata err: " + err.Error())
	}
	f.connCount.Add(1)
	return &transform.Conn{
		Conn: remoteConn,
		AfterCloseHook: func() {
			f.connCount.Add(-1)
			f.Log.With("remote.network", remote.Network, "remote.address", remote.Address, "conn.id", remote.ID).Info("close forward")
		},
	}, nil
}

func (f *ForwardClient) HTTPRequest(w http.ResponseWriter, r *http.Request) {
	l := f.Log.With("remote.network", "tcp", "remote.address", r.URL.Host)
	l.Info("forward http")
	resp, err := f.httpClient.Do(r)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = errors.New("remote server close connection")
		}
		l.Errorf("request call err: %v", err)
		l.With("url", r.URL.String()).Errorf("http call err: %v", err)
		ResponseError(w, err)
		return
	}
	defer resp.Body.Close()
	CopyHTTPResponse(w, resp)
}

func (f *ForwardClient) Conn(conn net.Conn, remote *api.ForwardMeta) {
	defer conn.Close()

	remoteConn, err := f.Dial(remote)
	if err != nil {
		f.Log.Errorf("dial remote err: %v", err)
		return
	}
	defer remoteConn.Close()
	l := f.Log.With("remote.network", remote.Network, "remote.address", remote.Address, "conn.id", remote.ID)
	l.Info("forward conn")

	transform.ConnCopyAndWait(conn, remoteConn, l)
}

func (f *ForwardClient) Close() error {
	f.cancel()
	return nil
}
