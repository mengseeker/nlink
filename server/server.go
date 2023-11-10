package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/mengseeker/nlink/core/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ServerConfig struct {
	Addr           string
	TLS_CA         string
	TLS_Cert       string
	TLS_Key        string
	ReadBufferSize int
}

type Server struct {
	api.UnimplementedProxyServer
	Config *ServerConfig

	httpClient http.Client
}

func NewServer(cfg ServerConfig) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:8899"
	}
	if cfg.TLS_CA+cfg.TLS_Cert+cfg.TLS_Key == "" {
		return nil, fmt.Errorf("invalid tls config")
	}
	if cfg.ReadBufferSize <= 0 {
		cfg.ReadBufferSize = 4 << 10
	}

	s := Server{
		Config: &cfg,
	}
	s.httpClient = http.Client{
		Transport: &http.Transport{},
	}
	return &s, nil
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
	api.RegisterProxyServer(gs, s)

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

func (s *Server) HTTPCall(stream api.Proxy_HTTPCallServer) (err error) {
	r, err := stream.Recv()
	if err != nil {
		return err
	}
	req := r.GetRequest()
	ir, iw := io.Pipe()
	defer iw.Close()
	proxyReq, err := http.NewRequestWithContext(stream.Context(), req.Method, req.Url, ir)
	for i := range req.Headers {
		proxyReq.Header.Add(req.Headers[i].Key, req.Headers[i].Value)
	}

	// handle read client stream write to remote body
	bodyDone := make(chan any)
	go func() {
		defer close(bodyDone)
		data := r.GetBody()
		_, err = iw.Write(data)
		if err != nil {
			slog.Error("write request data err", "error", err)
			return
		}

		defer iw.Close()
		for {
			br, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				slog.Error("recv request stream data err", "error", err)
				return
			}
			_, err = iw.Write(br.GetBody())
			if err != nil {
				slog.Error("write request stream data err", "error", err)
				return
			}
		}
	}()

	resp, err := s.httpClient.Do(proxyReq)
	if err != nil {
		slog.Error("http call err", "error", err)
		return err
	}
	defer resp.Body.Close()

	proxyResp := api.HTTPResponse_Response{
		Code:          int32(resp.StatusCode),
		ContentLength: resp.ContentLength,
	}
	for k, v := range resp.Header {
		for i := range v {
			proxyResp.Headers = append(proxyResp.Headers, &api.Header{Key: k, Value: v[i]})
		}
	}

	bf := make([]byte, s.Config.ReadBufferSize)
	n, err := resp.Body.Read(bf)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			slog.Error("read response body err", "error", err)
			return
		}
		// err = nil
	}
	err = stream.Send(&api.HTTPResponse{
		Response: &proxyResp,
		Body:     bf[:n],
	})
	if err != nil {
		slog.Error("write response err", "error", err)
		return
	}

	for {
		n, err = resp.Body.Read(bf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			} else {
				slog.Error("read response body err", "error", err)
				return
			}
		}
		err = stream.Send(&api.HTTPResponse{
			Body: bf[:n],
		})
		if err != nil {
			slog.Error("write response err", "error", err)
			return
		}
	}
	resp.Body.Close()
	<-bodyDone
	return nil
}

func (s *Server) TCPCall(stream api.Proxy_TCPCallServer) (err error) {
	r, err := stream.Recv()
	if err != nil {
		return err
	}
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.Req.Host, r.Req.Port))
	if err != nil {
		return fmt.Errorf("dial remote err %v", err)
	}
	defer remoteConn.Close()
	// try notify client
	err = stream.Send(&api.SockData{})
	if err != nil {
		return fmt.Errorf("notify client err %v", err)
	}

	// read remote data and write back
	done := make(chan int)
	go func() {
		defer close(done)
		remoteReadBuff := make([]byte, s.Config.ReadBufferSize)
		for {
			n, err := remoteConn.Read(remoteReadBuff)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				slog.Error("read remote err", "error", err)
				return
			}
			back := api.SockData{
				Data: remoteReadBuff[:n],
			}
			err = stream.Send(&back)
			if err != nil {
				slog.Error("write back err", "error", err)
				return
			}
		}
	}()

	// read data and write to remote
	if r.Data != nil {
		_, err = remoteConn.Write(r.Data.Data)
		if err != nil {
			slog.Error("write remote err", "error", err)
			return
		}
	}
	for {
		r, err = stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			}
			return err
		}
		slog.Error("read req err", "error", err)
		if r.Data != nil {
			_, err = remoteConn.Write(r.Data.Data)
			if err != nil {
				slog.Error("write remote err", "error", err)
				return err
			}
		}
	}

	<-done
	return
}
