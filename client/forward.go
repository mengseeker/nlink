package client

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
)

const (
	ForwardConnPoolSize = 10
)

type ForwardClient struct {
	Log           *log.Logger
	ForwardConfig ServerConfig
	connPool      ConnPool
	cancel        context.CancelFunc
}

func NewForwardClient(ctx context.Context, sc ServerConfig, l *log.Logger) (*ForwardClient, error) {
	ctx, cancel := context.WithCancel(ctx)
	c := ForwardClient{
		Log:           l,
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
	go func() {
		<-ctx.Done()
		c.connPool.Release()
	}()

	return &c, nil
}

func (f *ForwardClient) HTTPRequest(req *http.Request) (resp *http.Response) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	remoteConn, err := f.connPool.Get(ctx)
	if err != nil {
		f.Log.Errorf("unable to get remote conn, err: %v", err)
		return NewErrHTTPResponse(req, err.Error())
	}
	remote := api.ForwardMeta{
		Network: "tcp",
		Address: req.URL.Host,
	}
	err = transform.SendMsg(remoteConn, &remote)
	if err != nil {
		f.Log.Errorf("send metadata err: %v", err)
		return NewErrHTTPResponse(req, err.Error())
	}
	dialConn := remoteConn.(net.Conn)
	// dialConn will auto close after call
	hc := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialConn, nil
			},
		},
	}
	deleteRequestHeaders(req)
	resp, err = hc.Do(req)
	if err != nil && !errors.Is(err, io.EOF) {
		f.Log.Errorf("request call err: %v", err)
		return NewErrHTTPResponse(req, err.Error())
	}
	return resp
}

func (f *ForwardClient) Conn(conn net.Conn, remote *api.ForwardMeta) {
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	remoteConn, err := f.connPool.Get(ctx)
	if err != nil {
		f.Log.Errorf("unable to get remote conn, err: %v", err)
		return
	}
	defer remoteConn.Close()
	err = transform.SendMsg(remoteConn, remote)
	if err != nil {
		f.Log.Errorf("send metadata err: %v", err)
		return
	}
	wg := sync.WaitGroup{}
	copy := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		_, err := io.Copy(w, r)
		if err != nil {
			f.Log.Errorf("copy data err: %v", err)
		}
	}
	wg.Add(2)
	go copy(remoteConn, conn)
	go copy(conn, remoteConn)
	wg.Wait()
}

func (f *ForwardClient) Close() error {
	f.cancel()
	return nil
}
