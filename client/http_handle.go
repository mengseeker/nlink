package client

import (
	"net"
	"net/http"

	"github.com/mengseeker/nlink/core/api"
	"gopkg.in/elazarl/goproxy.v1"
)

var (
	ProxyHeaders = map[string]bool{
		"Proxy-Connection":    true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Connection":          true,
	}
)

type HTTPHandler struct {
	*goproxy.ProxyHttpServer
	// ruleMapper *RuleMapper
}

func NewHTTPHandler(mapper *RuleMapper) HTTPHandler {
	h := HTTPHandler{
		ProxyHttpServer: goproxy.NewProxyHttpServer(),
		// ruleMapper:      mapper,
	}
	h.ProxyHttpServer.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			return req, mapper.Match(NewMatchMetaFromHTTPRequest(req)).HTTPRequest(req)
		})
	h.ProxyHttpServer.OnRequest().HandleConnectFunc(
		func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
			return &goproxy.ConnectAction{
				Action: goproxy.ConnectHijack,
				Hijack: func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
					// client.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n"))
					remote := api.ForwardMeta{
						Network: "tcp",
						Address: req.URL.Host,
					}
					mapper.Match(NewMatchMetaFromHTTPSHost(host)).Conn(client, &remote)
				},
			}, host
		})
	return h
}
