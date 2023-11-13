package client

import (
	"context"
)

// TODO
type QuicForwardClient struct {
	name string
}

func (cli *QuicForwardClient) ServerName() string {
	return cli.name
}

func (cli *QuicForwardClient) HTTPCall(ctx context.Context) (stream Proxy_HTTPCallClient, err error) {
	return
}

func (cli *QuicForwardClient) TCPCall(ctx context.Context) (stream Proxy_TCPCallClient, err error) {
	return
}
func DialQuicServer(ctx context.Context, sc ServerConfig) (cli *QuicForwardClient, err error) {

	return
}
