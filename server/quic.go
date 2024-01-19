package server

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
	"github.com/quic-go/quic-go"
)

type QuicServer struct {
	Handler Handler
	Config  *ServerConfig
	Log     *log.Logger
}

func NewQuicServer(cfg ServerConfig, log *log.Logger) (*QuicServer, error) {
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:8899"
	}
	s := QuicServer{
		Config: &cfg,
		Log:    log.With("server.net", "udp"),
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
	handler := Handler{
		Log:    s.Log,
		Dialer: net.Dialer{},
	}
	for {
		stream, err := conn.AcceptStream(c)
		if err != nil {
			s.Log.Errorf("AcceptStream err: %v", err)
			return
		}
		go handler.HandleConnect(transform.QUICConn{Conn: conn, Stream: stream})
	}
}
