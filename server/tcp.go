package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/mengseeker/nlink/core/log"
)

type TCPServer struct {
	Config *ServerConfig

	log *log.Logger
	lis net.Listener
}

func NewTCPServer(cfg ServerConfig, log *log.Logger) (*TCPServer, error) {
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:8899"
	}
	s := TCPServer{
		Config: &cfg,
		log:    log.With("server.type", "tcp"),
	}
	return &s, nil
}

func (s *TCPServer) Start(c context.Context) (err error) {
	tc, err := NewServerTls(s.Config.TLS_Cert, s.Config.TLS_Key, s.Config.TLS_CA)
	if err != nil {
		return
	}
	lis, err := tls.Listen("tcp", s.Config.Addr, tc)
	if err != nil {
		return fmt.Errorf("fail to listen tcp, err: %v", err)
	}
	defer lis.Close()
	s.log.Infof("server listening at %v", lis.Addr())
	s.lis = lis
	handler := Handler{
		Log:    s.log,
		Dialer: net.Dialer{},
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			return fmt.Errorf("accept connect err: %v", err)
		}
		go handler.HandleConnect(conn)
	}
}
