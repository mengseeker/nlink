package client

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/mengseeker/nlink/core/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcForwardClient struct {
	api.ProxyClient
	name string
	conn *grpc.ClientConn
}

func (cli *GrpcForwardClient) ServerName() string {
	return cli.name
}

func (cli *GrpcForwardClient) HTTPCall(ctx context.Context) (Proxy_HTTPCallClient, error) {
	return cli.ProxyClient.HTTPCall(ctx)
}

func (cli *GrpcForwardClient) TCPCall(ctx context.Context) (Proxy_TCPCallClient, error) {
	return cli.ProxyClient.TCPCall(ctx)
}

func DialGrpcServer(ctx context.Context, sc ServerConfig) (cli *GrpcForwardClient, err error) {
	cert, err := tls.LoadX509KeyPair(sc.Cert, sc.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %v", err)
	}
	tls := credentials.NewTLS(&tls.Config{
		ServerName:         "x.test.example.com",
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	})
	conn, err := grpc.Dial(sc.Addr, grpc.WithTransportCredentials(tls))
	if err != nil {
		return nil, fmt.Errorf("dial server err: %v", err)
	}
	cli = &GrpcForwardClient{
		conn:        conn,
		ProxyClient: api.NewProxyClient(conn),
		name:        sc.Name,
	}
	return
}
