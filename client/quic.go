package client

import (
	"context"
	"fmt"
	"time"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/quics"
	"github.com/quic-go/quic-go"
)

type QuicForwardClient struct {
	name                  string
	conn                  quic.Connection
	QuicOpenStreamTimeout time.Duration
}

func (cli *QuicForwardClient) ServerName() string {
	return cli.name
}

func (cli *QuicForwardClient) HTTPCall(ctx context.Context) (stream Proxy_HTTPCallClient, err error) {
	var header quics.StreamHeader
	header.SetStreamType(quics.StreamType_HTTP)
	qstream, err := cli.NewStream(ctx, header)
	if err != nil {
		return
	}
	stream = &QuicProxy_HTTPCallClient{stream: qstream}
	return
}

func (cli *QuicForwardClient) TCPCall(ctx context.Context) (stream Proxy_TCPCallClient, err error) {
	var header quics.StreamHeader
	header.SetStreamType(quics.StreamType_TCP)
	qstream, err := cli.NewStream(ctx, header)
	if err != nil {
		return
	}
	stream = &QuicProxy_TCPCallClient{stream: qstream}
	return
}

func (cli *QuicForwardClient) NewStream(ctx context.Context, h quics.StreamHeader) (stream quic.Stream, err error) {
	stream, err = cli.conn.OpenStreamSync(ctx)
	if err != nil {
		return
	}
	err = quics.WriteHeader(stream, h)
	return
}

func DialQuicServer(ctx context.Context, sc ServerConfig) (cli *QuicForwardClient, err error) {
	tc, err := NewClientTls(sc.Cert, sc.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %v", err)
	}
	conn, err := quic.DialAddr(ctx, sc.Addr, tc, &quic.Config{})
	if err != nil {
		return
	}
	cli = &QuicForwardClient{
		name:                  sc.Name,
		conn:                  conn,
		QuicOpenStreamTimeout: time.Second,
	}
	return
}

type QuicProxy_HTTPCallClient struct {
	stream quic.Stream
}

func (cli *QuicProxy_HTTPCallClient) Send(data *api.HTTPRequest) error {
	return quics.SendMsg(cli.stream, data)
}
func (cli *QuicProxy_HTTPCallClient) Recv() (*api.HTTPResponse, error) {
	var req api.HTTPResponse
	return &req, quics.RecvMsg(cli.stream, &req)
}
func (cli *QuicProxy_HTTPCallClient) Context() context.Context {
	return cli.stream.Context()
}
func (cli *QuicProxy_HTTPCallClient) CloseSend() error {
	return cli.stream.Close()
}

type QuicProxy_TCPCallClient struct {
	stream quic.Stream
}

func (cli *QuicProxy_TCPCallClient) Send(data *api.SockRequest) error {
	return quics.SendMsg(cli.stream, data)
}
func (cli *QuicProxy_TCPCallClient) Recv() (*api.SockData, error) {
	var req api.SockData
	return &req, quics.RecvMsg(cli.stream, &req)
}
func (cli *QuicProxy_TCPCallClient) Context() context.Context {
	return cli.stream.Context()
}
func (cli *QuicProxy_TCPCallClient) CloseSend() error {
	return cli.stream.Close()
}
