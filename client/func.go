package client

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/mengseeker/nlink/core/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ForwardClient struct {
	api.ProxyClient
	Name string
	conn *grpc.ClientConn
}

type FunctionProvider struct {
	ServerConfigs  map[string]ServerConfig
	Forwards       map[string]*ForwardClient
	ReadBufferSize int
}

func NewFunctionProvider(sc []ServerConfig) *FunctionProvider {
	pv := FunctionProvider{
		Forwards:      make(map[string]*ForwardClient),
		ServerConfigs: make(map[string]ServerConfig),
	}
	for i := range sc {
		pv.ServerConfigs[sc[i].Name] = sc[i]
	}
	pv.ReadBufferSize = 4 << 10
	return &pv
}

func (pv *FunctionProvider) dialProxyServer(ctx context.Context, name string) (err error) {
	sc := pv.ServerConfigs[name]
	cert, err := tls.LoadX509KeyPair(sc.Cert, sc.Key)
	if err != nil {
		return fmt.Errorf("failed to load client cert: %v", err)
	}
	tls := credentials.NewTLS(&tls.Config{
		ServerName:         "x.test.example.com",
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	})
	conn, err := grpc.Dial(sc.Addr, grpc.WithTransportCredentials(tls))
	// conn, err := grpc.Dial(sc.Addr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("dial server err: %v", err)
	}
	pv.Forwards[sc.Name] = &ForwardClient{
		ProxyClient: api.NewProxyClient(conn),
		Name:        sc.Name,
		conn:        conn,
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

func (pv *FunctionProvider) getForwardProxyClient(ctx context.Context, name string) (cli *ForwardClient, err error) {
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

func (pv *FunctionProvider) DialHTTP(ctx context.Context, name string) (stream api.Proxy_HTTPCallClient, err error) {
	cli, err := pv.getForwardProxyClient(ctx, name)
	if err != nil {
		return
	}
	stream, err = cli.ProxyClient.HTTPCall(ctx)
	return
}

func (pv *FunctionProvider) DialTCP(ctx context.Context, name string) (stream api.Proxy_TCPCallClient, err error) {
	cli, err := pv.getForwardProxyClient(ctx, name)
	if err != nil {
		return
	}
	stream, err = cli.TCPCall(ctx)
	return
}
