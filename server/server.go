package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/mengseeker/nlink/core/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ServerConfig struct {
	Addr     string
	TLS_CA   string
	TLS_Cert string
	TLS_Key  string
}

type Server struct {
	api.UnimplementedHTTPCallerServer
	api.UnimplementedTCPCallerServer
	Config *ServerConfig

	httpClient http.Client
}

func (s *Server) Start(c context.Context) (err error) {
	// grpc options
	var sopts []grpc.ServerOption
	cert, err := tls.LoadX509KeyPair(s.Config.TLS_Cert, s.Config.TLS_Key)
	if err != nil {
		return fmt.Errorf("load tls err: %v", err)
	}
	ca := x509.NewCertPool()
	caBytes, err := os.ReadFile(s.Config.TLS_CA)
	if err != nil {
		return fmt.Errorf("load ca err: %v", err)
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return fmt.Errorf("failed to parse ca %q", s.Config.TLS_CA)
	}
	sopts = append(sopts, grpc.Creds(credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    ca,
	})))

	gs := grpc.NewServer(sopts...)

	// register grpc services
	api.RegisterHTTPCallerServer(gs, s)
	api.RegisterTCPCallerServer(gs, s)

	// listen
	lis, err := net.Listen("tcp", s.Config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	slog.Info(fmt.Sprintf("server listening at %v", lis.Addr()))
	gs.Serve(lis)
	if err := gs.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return
}

func (s *Server) HTTPCall(stream api.HTTPCaller_HTTPCallServer) (err error) {
	// r, err := stream.Recv()
	// if err != nil {
	// 	return err
	// }
	// req := r.GetRequest()

	// proxyReq := http.NewRequestWithContext(stream.Context(), req.Method, u, body)
	// s.httpClient.Do()
	return
}

func (s *Server) TCPCall(stream api.TCPCaller_TCPCallServer) (err error) {
	return
}
