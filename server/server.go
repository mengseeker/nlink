package server

import (
	"context"
	"net"

	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
)

var (
	logger = log.NewLogger()
)

type ServerConfig struct {
	Addr     string
	TLS_CA   string
	TLS_Cert string
	TLS_Key  string
}

func Start(c context.Context, cfg ServerConfig) {
	gs, err := NewServer(cfg)
	if err != nil {
		panic(err)
	}
	if err := gs.Start(c); err != nil {
		panic(err)
	}
}

type Server struct {
	Config *ServerConfig
}

func NewServer(cfg ServerConfig) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:8899"
	}
	s := Server{
		Config: &cfg,
	}
	return &s, nil
}

func (s *Server) Start(c context.Context) (err error) {
	tc, err := NewServerTls(s.Config.TLS_Cert, s.Config.TLS_Key, s.Config.TLS_CA)
	if err != nil {
		return
	}

	lis, err := transform.ListenPackConn(s.Config.Addr, tc)
	if err != nil {
		return
	}

	for {
		stream, err := lis.Accept()
		if err != nil {
			logger.Error("accept", err)
			continue
		}
		go s.handleStream(c, stream)
	}
}

func (s *Server) handleStream(_ context.Context, stream *transform.PackStream) {
	defer stream.Close()
	remoteConn, err := net.Dial(stream.ProxyMeta.Network, stream.ProxyMeta.Address)
	if err != nil {
		logger.Error("dial remote", err)
		return
	}
	defer remoteConn.Close()

	transform.TransformConn(stream, remoteConn, logger)
}
