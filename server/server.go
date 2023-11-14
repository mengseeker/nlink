package server

import (
	"context"
	"net/http"

	"github.com/mengseeker/nlink/core/log"
)

type ServerConfig struct {
	Addr           string
	TLS_CA         string
	TLS_Cert       string
	TLS_Key        string
	ReadBufferSize int
}

func Start(c context.Context, cfg ServerConfig) {
	l := log.NewLogger()
	handler := Handler{
		Log:            l,
		ReadBufferSize: cfg.ReadBufferSize,
		HTTPClient:     http.DefaultClient,
	}
	gs, err := NewGrpcServer(cfg, handler, l)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := gs.Start(c); err != nil {
			panic(err)
		}
	}()
	
	qs, err := NewQuicServer(cfg, handler, l)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := qs.Start(c); err != nil {
			panic(err)
		}
	}()
	<-c.Done()
}
