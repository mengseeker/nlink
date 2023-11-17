package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mengseeker/nlink/core/log"
	"gopkg.in/elazarl/goproxy.v1"
)

var (
	l = log.With("Uint", "client")
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

	proxy    *goproxy.ProxyHttpServer
	provider *FunctionProvider
}

func NewProxy(cfg ProxyConfig) (p *Proxy, err error) {
	if cfg.Listen == "" {
		cfg.Listen = ":7890"
	}
	p = &Proxy{
		Config: &cfg,
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
	p.provider, err = NewFunctionProvider(ctx, p.Config.Servers, p.Config.Resolver)
	if err != nil {
		return err
	}
	p.proxy = goproxy.NewProxyHttpServer()
	if err = p.applyRule(); err != nil {
		return
	}
	l.Infof("proxy listen at: %s", p.Config.Listen)
	return http.ListenAndServe(p.Config.Listen, p.proxy)
}

func (p *Proxy) applyRule() (err error) {
	for _, rStr := range p.Config.Rules {
		r, err := UnmashalProxyRule(rStr)
		if err != nil {
			return fmt.Errorf("unmashal rule err: %v", err)
		}
		conds, err := r.BuildProxyConds(p.provider)
		if err != nil {
			return fmt.Errorf("invalid rule %q, err: %v", rStr, err)
		}
		rh, ch, err := r.BuildProxyAction(p.provider)
		if err != nil {
			return fmt.Errorf("invalid rule %q, err: %v", rStr, err)
		}
		p.proxy.OnRequest(conds...).DoFunc(rh)
		p.proxy.OnRequest(conds...).HandleConnectFunc(ch)
	}
	return
}
