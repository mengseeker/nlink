package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mengseeker/nlink/core/api"
	"gopkg.in/elazarl/goproxy.v1"
)

type ServerConfig struct {
	Addr string
	Cert string
	Key  string
}

type ProxyConfig struct {
	Name    string
	Addr    string
	Verbose bool
	Rules   []string
	Servers []ServerConfig
}

type Proxy struct {
	Config *ProxyConfig

	servers  map[string]api.ProxyServer
	proxy    *goproxy.ProxyHttpServer
	provider *FunctionProvider
}

func (p *Proxy) Start(ctx context.Context) (err error) {
	p.proxy = goproxy.NewProxyHttpServer()
	p.proxy.Verbose = p.Config.Verbose
	p.servers = make(map[string]api.ProxyServer)
	// TODO connect to proxy servers
	if err = p.applyRule(); err != nil {
		return
	}
	return http.ListenAndServe(p.Config.Addr, p.proxy)
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
