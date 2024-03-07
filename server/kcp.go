package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/mengseeker/nlink/core/log"
	"github.com/xtaci/kcp-go/v5"
)

type KCPServer struct {
	Config *ServerConfig

	log *log.Logger
	lis net.Listener
}

func NewKCPServer(cfg ServerConfig, log *log.Logger) (*KCPServer, error) {
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:8898"
	}
	s := KCPServer{
		Config: &cfg,
		log:    log.With("server.net", "udp-kcp"),
	}
	return &s, nil
}

func (s *KCPServer) Start(c context.Context) (err error) {
	tc, err := NewServerTls(s.Config.TLS_Cert, s.Config.TLS_Key, s.Config.TLS_CA)
	if err != nil {
		return
	}
	lis, err := kcp.ListenWithOptions("127.0.0.1:12345", nil, 10, 3)
	if err != nil {
		return fmt.Errorf("listen err: %v", err)
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
		tconn := tls.Server(conn, tc)
		if err != nil {
			return fmt.Errorf("accept connect err: %v", err)
		}
		go handler.HandleConnect(tconn)
	}
}
