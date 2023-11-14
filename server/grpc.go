package server

import (
	"context"
	"fmt"
	"net"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcServer struct {
	api.UnimplementedProxyServer
	Handler Handler
	Config  *ServerConfig
	Log     *log.Logger
}

func NewGrpcServer(cfg ServerConfig, handler Handler, log *log.Logger) (*GrpcServer, error) {
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:8899"
	}
	if cfg.TLS_CA+cfg.TLS_Cert+cfg.TLS_Key == "" {
		return nil, fmt.Errorf("invalid tls config")
	}
	if cfg.ReadBufferSize <= 0 {
		cfg.ReadBufferSize = 4 << 10
	}

	s := GrpcServer{
		Config:  &cfg,
		Handler: handler,
		Log:     log.With("Unit", "GrpcServer"),
	}
	return &s, nil
}

func (s *GrpcServer) Start(c context.Context) (err error) {
	tls, err := NewServerTls(s.Config.TLS_Cert, s.Config.TLS_Key, s.Config.TLS_CA)
	if err != nil {
		return
	}

	gs := grpc.NewServer(grpc.Creds(credentials.NewTLS(tls)))

	// register grpc services
	api.RegisterProxyServer(gs, s)

	// listen
	lis, err := net.Listen("tcp", s.Config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s.Log.Infof("server listening at %v", lis.Addr())
	if err := gs.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return
}

func (s *GrpcServer) HTTPCall(stream api.Proxy_HTTPCallServer) (err error) {
	return s.Handler.HTTPCall(stream)
}

func (s *GrpcServer) TCPCall(stream api.Proxy_TCPCallServer) (err error) {
	return s.Handler.TCPCall(stream)
}
