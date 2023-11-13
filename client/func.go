package client

import (
	"context"
	"fmt"

	"github.com/mengseeker/nlink/core/api"
)

type ProxyStream interface {
	Context() context.Context
	CloseSend() error
}

type Proxy_HTTPCallClient interface {
	ProxyStream
	Send(*api.HTTPRequest) error
	Recv() (*api.HTTPResponse, error)
}

type Proxy_TCPCallClient interface {
	ProxyStream
	Send(*api.SockRequest) error
	Recv() (*api.SockData, error)
}

type ForwardClient interface {
	HTTPCall(ctx context.Context) (Proxy_HTTPCallClient, error)
	TCPCall(ctx context.Context) (Proxy_TCPCallClient, error)
	ServerName() string
}

type FunctionProvider struct {
	ServerConfigs  map[string]ServerConfig
	Forwards       map[string]ForwardClient
	ReadBufferSize int
}

func NewFunctionProvider(sc []ServerConfig) *FunctionProvider {
	pv := FunctionProvider{
		Forwards:      make(map[string]ForwardClient),
		ServerConfigs: make(map[string]ServerConfig),
	}
	for i := range sc {
		pv.ServerConfigs[sc[i].Name] = sc[i]
	}
	pv.ReadBufferSize = 4 << 10
	return &pv
}

func (pv *FunctionProvider) dialProxyServer(ctx context.Context, name string) (err error) {
	sc, exist := pv.ServerConfigs[name]
	if !exist {
		return fmt.Errorf("forward server %q not fround", name)
	}
	if sc.Net == "tcp" {
		cli, err := DialGrpcServer(ctx, sc)
		if err != nil {
			return err
		}
		pv.Forwards[name] = cli
	} else {
		cli, err := DialQuicServer(ctx, sc)
		if err != nil {
			return err
		}
		pv.Forwards[name] = cli
	}
	return
}

type Country string

const (
	China Country = "CN"
)

func (pv *FunctionProvider) GEOIP(ip string) Country {
	// TODO
	return China
}

func (pv *FunctionProvider) Resolver(domain string) (IP string) {
	// TODO
	return
}

func (pv *FunctionProvider) getForwardProxyClient(ctx context.Context, name string) (cli ForwardClient, err error) {
	cli, ok := pv.Forwards[name]
	if !ok {
		err = pv.dialProxyServer(ctx, name)
		if err != nil {
			return
		}
		cli = pv.Forwards[name]
	}
	return
}

func (pv *FunctionProvider) DialHTTP(ctx context.Context, name string) (stream Proxy_HTTPCallClient, err error) {
	cli, err := pv.getForwardProxyClient(ctx, name)
	if err != nil {
		return
	}
	stream, err = cli.HTTPCall(ctx)
	return
}

func (pv *FunctionProvider) DialTCP(ctx context.Context, name string) (stream Proxy_TCPCallClient, err error) {
	cli, err := pv.getForwardProxyClient(ctx, name)
	if err != nil {
		return
	}
	stream, err = cli.TCPCall(ctx)
	return
}
