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

// type ProxyHandler struct {
// 	log *log.Logger
// 	fws map[string]*ForwardClient
// }

// func (h *ProxyHandler) NewConnectDirect() goproxy.FuncHttpsHandler {
// 	return func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
// 		h.log.Info("direct connect", "host", host)
// 		return goproxy.OkConnect, ctx.Req.URL.Host
// 	}
// }

// func (h *ProxyHandler) NewConnectReject() goproxy.FuncHttpsHandler {
// 	return func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
// 		h.log.Info("reject connect", "host", host)
// 		return goproxy.RejectConnect, ctx.Req.URL.Host
// 	}
// }

// func (h *ProxyHandler) NewConnectForward(fc *ForwardClient) (handle goproxy.FuncHttpsHandler) {
// 	handle = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
// 		fc.Log.Info("forward connect", "host", host)
// 		return &goproxy.ConnectAction{
// 			Action: goproxy.ConnectHijack,
// 			Hijack: func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
// 				remote := api.ForwardMeta{
// 					Network: "tcp",
// 					Address: host,
// 				}
// 				fc.ForwardConn(client, &remote)
// 			},
// 		}, host
// 	}
// 	return
// }

// func (h *ProxyHandler) NewHTTPDirect() goproxy.FuncReqHandler {
// 	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
// 		h.log.Info("direct request", "url", ctx.Req.URL)
// 		deleteRequestHeaders(req)
// 		resp, err := ctx.RoundTrip(req)
// 		if err != nil {
// 			if resp == nil {
// 				h.log.Errorf("error read response %v %v:", req.URL.Host, err.Error())
// 				resp = goproxy.NewResponse(req,
// 					goproxy.ContentTypeText, http.StatusBadGateway,
// 					err.Error())
// 			}
// 		}
// 		return req, resp
// 	}
// }

// func (h *ProxyHandler) NewHTTPForward(fc *ForwardClient) (handle goproxy.FuncReqHandler) {
// 	handle = func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
// 		fc.Log.Info("forward http", "url", ctx.Req.URL)
// 		deleteRequestHeaders(req)
// 		resp := fc.ForwardHTTP(req)
// 		return req, resp
// 	}
// 	return
// }
