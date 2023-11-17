package client

import (
	"net"
	"net/http"

	"github.com/mengseeker/nlink/core/log"

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

type ProxyHandler struct {
	log *log.Logger
	fws map[string]*ForwardClient
}

func (h *ProxyHandler) NewConnectDirect() goproxy.FuncHttpsHandler {
	return func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		h.log.Info("direct connect", "host", host)
		return goproxy.OkConnect, ctx.Req.URL.Host
	}
}

func (h *ProxyHandler) NewConnectReject() goproxy.FuncHttpsHandler {
	return func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		h.log.Info("reject connect", "host", host)
		return goproxy.RejectConnect, ctx.Req.URL.Host
	}
}

func (h *ProxyHandler) NewConnectForward(fc *ForwardClient) (handle goproxy.FuncHttpsHandler) {
	handle = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		fc.Log.Info("forward connect", "host", host)
		return &goproxy.ConnectAction{
			Action: goproxy.ConnectHijack,
			Hijack: func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
				remote := api.ForwardMeta{
					Network: "tcp",
					Address: host,
				}
				fc.ForwardConn(client, &remote)
			},
		}, host
	}
	return
}

func (h *ProxyHandler) NewHTTPDirect() goproxy.FuncReqHandler {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		h.log.Info("direct request", "url", ctx.Req.URL)
		deleteRequestHeaders(req)
		resp, err := ctx.RoundTrip(req)
		if err != nil {
			if resp == nil {
				h.log.Errorf("error read response %v %v:", req.URL.Host, err.Error())
				resp = goproxy.NewResponse(req,
					goproxy.ContentTypeText, http.StatusBadGateway,
					err.Error())
			}
		}
		return req, resp
	}
}

func (h *ProxyHandler) NewHTTPReject() goproxy.FuncReqHandler {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		h.log.Info("reject request", "url", ctx.Req.URL)
		return req, goproxy.NewResponse(req,
			goproxy.ContentTypeText, http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
}

func (h *ProxyHandler) NewHTTPForward(fc *ForwardClient) (handle goproxy.FuncReqHandler) {
	handle = func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		fc.Log.Info("forward http", "url", ctx.Req.URL)
		deleteRequestHeaders(req)
		resp := fc.ForwardHTTP(req)
		return req, resp
	}
	return
}

func NewHTTPResponse(req *http.Request) (resp *http.Response) {
	resp = &http.Response{}
	resp.Request = req
	resp.TransferEncoding = req.TransferEncoding
	resp.Header = make(http.Header)
	return
}

func deleteRequestHeaders(req *http.Request) {
	req.RequestURI = "" // this must be reset when serving a request with the client
	// req.Header.Del("Accept-Encoding")
	for k := range ProxyHeaders {
		req.Header.Del(k)
	}
}
