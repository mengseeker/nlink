package server

import (
	"context"
	"net/http"

	"github.com/mengseeker/nlink/core/log"
)

type ServerConfig struct {
	Net            string
	Addr           string
	TLS_CA         string
	TLS_Cert       string
	TLS_Key        string
	ReadBufferSize int
}

func Start(c context.Context, cfg ServerConfig) (err error) {
	log := log.NewLogger()
	handler := Handler{
		Log:            log,
		ReadBufferSize: cfg.ReadBufferSize,
		HTTPClient:     http.DefaultClient,
	}
	gs, err := NewGrpcServer(cfg, handler, log)
	if err != nil {
		return
	}
	return gs.Start(c)
}
