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

type ProxyConfig struct {
	Listen  string
	Verbose bool
	Rules   []string
	Servers []ServerConfig
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
	p.provider = NewFunctionProvider(p.Config.Servers)
	return
}

func (p *Proxy) Start(ctx context.Context) (err error) {
	p.proxy = goproxy.NewProxyHttpServer()
	p.proxy.Verbose = p.Config.Verbose
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
