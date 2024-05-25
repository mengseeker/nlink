package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/mengseeker/nlink/core/log"
)

var (
	logger = log.NewLogger()
)

const (
	DialTimeout = 3 * time.Second
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

	lis, err := tls.Listen("tcp", s.Config.Addr, tc)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			logger.Error("accept", err)
			continue
		}
		go s.Serve(conn)
	}
}
