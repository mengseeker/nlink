package server

import (
	"context"

	"github.com/mengseeker/nlink/core/log"
)

type ServerConfig struct {
	Addr     string
	TLS_CA   string
	TLS_Cert string
	TLS_Key  string
}

func Start(c context.Context, cfg ServerConfig) {
	l := log.NewLogger()
	gs, err := NewTCPServer(cfg, l)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := gs.Start(c); err != nil {
			panic(err)
		}
	}()

	qs, err := NewQuicServer(cfg, l)
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
