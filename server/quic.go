package server

import (
	"context"
	"errors"
	"time"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/quics"
	"github.com/quic-go/quic-go"
)

type QuicServer struct {
	Handler Handler
	Config  *ServerConfig
	Log     *log.Logger
}

func NewQuicServer(cfg ServerConfig, handler Handler, log *log.Logger) (*QuicServer, error) {
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:8899"
	}
	if cfg.ReadBufferSize <= 0 {
		cfg.ReadBufferSize = 4 << 10
	}

	s := QuicServer{
		Config:  &cfg,
		Handler: handler,
		Log:     log.With("Unit", "QuicServer"),
	}
	return &s, nil
}

func (s *QuicServer) Start(c context.Context) (err error) {
	tls, err := NewServerTls(s.Config.TLS_Cert, s.Config.TLS_Key, s.Config.TLS_CA)
	if err != nil {
		return
	}
	lis, err := quic.ListenAddr(s.Config.Addr, tls, &quic.Config{
		KeepAlivePeriod: time.Second,
	})
	if err != nil {
		return
	}
	s.Log.Infof("server listening at %v", lis.Addr())

	// handle
	for {
		conn, err := lis.Accept(c)
		if err != nil {
			if errors.Is(err, quic.ErrServerClosed) {
				return nil
			}
			return err
		}
		go s.handleClient(c, conn)
	}
}

func (s *QuicServer) handleClient(c context.Context, conn quic.Connection) {
	for {
		stream, err := conn.AcceptStream(c)
		if err != nil {
			s.Log.Errorf("AcceptStream err: %v", err)
			return
		}
		go s.handleStream(c, stream)
	}
}

func (s *QuicServer) handleStream(c context.Context, stream quic.Stream) {
	streamLog := s.Log.With("streamID", stream.StreamID())
	defer stream.Close()
	defer streamLog.Debug("server close stream")
	header, err := quics.ReadHeader(stream)
	if err != nil {
		streamLog.Errorf("read stream header err: %v", err)
		return
	}
	switch header.StreamType() {
	case quics.StreamType_TCP:
		if err = s.Handler.TCPCall(&QuicTCPCallStream{stream: stream}); err != nil {
			streamLog.Errorf("handle tcp stream err: %v", err)
		}
	case quics.StreamType_HTTP:
		if err = s.Handler.HTTPCall(&QuicHTTPCallStream{stream: stream}); err != nil {
			streamLog.Errorf("handle http stream err: %v", err)
		}
	default:
		streamLog.Errorf("invalid streamtype: %x", header)
	}
}

type QuicHTTPCallStream struct {
	stream quic.Stream
}

func (c *QuicHTTPCallStream) Context() context.Context {
	return c.stream.Context()
}

func (c *QuicHTTPCallStream) Send(data *api.HTTPResponse) error {
	return quics.SendMsg(c.stream, data)
}

func (c *QuicHTTPCallStream) Recv() (*api.HTTPRequest, error) {
	var req api.HTTPRequest
	return &req, quics.RecvMsg(c.stream, &req)
}

type QuicTCPCallStream struct {
	stream quic.Stream
}

func (c *QuicTCPCallStream) Context() context.Context {
	return c.stream.Context()
}

func (c *QuicTCPCallStream) Send(data *api.SockData) error {
	return quics.SendMsg(c.stream, data)
}

func (c *QuicTCPCallStream) Recv() (*api.SockRequest, error) {
	var req api.SockRequest
	return &req, quics.RecvMsg(c.stream, &req)
}
