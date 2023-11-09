package client

import (
	"net/http"

	"gopkg.in/elazarl/goproxy.v1"
)

func DirectReqHandle(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	req.RequestURI = "" // this must be reset when serving a request with the client
	// req.Header.Del("Accept-Encoding")
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Connection")
	resp, err := ctx.RoundTrip(req)
	if err != nil {
		if resp == nil {
			ctx.Logf("error read response %v %v:", req.URL.Host, err.Error())
			resp = goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusBadGateway,
				err.Error())
		}
	}
	return req, resp
}

func DirectHandleConnect(req string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return goproxy.OkConnect, ctx.Req.URL.Host
}

func RejectReq(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	return req, goproxy.NewResponse(req,
		goproxy.ContentTypeText, http.StatusForbidden, http.StatusText(http.StatusForbidden))
}

func newForwardHandle(pv *FunctionProvider, name string) (reqHandle goproxy.FuncReqHandler, httpsHandle goproxy.FuncHttpsHandler, err error) {
	// TODO
	return
}
