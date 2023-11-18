package client

import (
	"context"
	"fmt"

	"github.com/mengseeker/nlink/core/log"
)

type ServerConfig struct {
	Net  string
	Name string
	Addr string
	Cert string
	Key  string
}

type ResolverConfig struct {
	DNS string
	DoT string
}

type ProxyConfig struct {
	Listen   string
	Net      string
	Cert     string
	Key      string
	Rules    []string
	Servers  []ServerConfig
	Resolver []ResolverConfig
}

type Proxy struct {
	Config *ProxyConfig

	ctx    context.Context
	cancel context.CancelFunc
	lis    *Listener
	log    *log.Logger
}

func NewProxy(cfg ProxyConfig) (p *Proxy, err error) {
	if cfg.Listen == "" {
		cfg.Listen = ":7890"
	}
	p = &Proxy{
		Config: &cfg,
		log:    log.NewLogger().With("unit", "client"),
	}
	for i := range p.Config.Servers {
		if p.Config.Servers[i].Net == "" {
			p.Config.Servers[i].Net = p.Config.Net
		}
		if p.Config.Servers[i].Cert == "" {
			p.Config.Servers[i].Cert = p.Config.Cert
		}
		if p.Config.Servers[i].Key == "" {
			p.Config.Servers[i].Key = p.Config.Key
		}
	}
	return
}

func (p *Proxy) Start(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	p.ctx = ctx
	p.cancel = cancel
	provider, err := NewFuncProvider(p.Config.Resolver, p.log)
	if err != nil {
		return err
	}

	forwards := map[string]*ForwardClient{}
	for _, sc := range p.Config.Servers {
		forward, err := NewForwardClient(ctx, sc, p.log)
		if err != nil {
			return fmt.Errorf("new forwardclient %s err: %v", sc.Name, err)
		}
		forwards[sc.Name] = forward
	}

	mapper, err := NewRuleMapper(ctx, p.Config.Rules, provider, forwards)
	if err != nil {
		return fmt.Errorf("parse rule err: %v", err)
	}
	httpHandler := NewHTTPHandler(mapper)
	socks4Handler := NewSocks4Handler(mapper)
	socks5Handler := NewSocks5Handler(mapper)

	lis := &Listener{
		Address:       p.Config.Listen,
		Log:           p.log,
		HTTPHandler:   httpHandler,
		Socks4Handler: socks4Handler,
		Socks5Handler: socks5Handler,
	}
	p.lis = lis
	return p.lis.ListenAndServe(ctx)
}

func (p *Proxy) Stop() (err error) {
	p.cancel()
	return nil
}
