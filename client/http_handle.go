package client

import (
	"net/http"

	"github.com/mengseeker/nlink/core/api"
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
	ruleMapper *RuleMapper
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		h.handleHTTPS(w, r)
	} else {
		h.handleHTTP(w, r)
	}
}

func (h *HTTPHandler) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	hij, ok := w.(http.Hijacker)
	if !ok {
		panic("httpserver does not support hijacking")
	}

	proxyClient, _, e := hij.Hijack()
	if e != nil {
		panic("Cannot hijack connection " + e.Error())
	}
	proxyClient.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
	remote := api.ForwardMeta{
		Network: "tcp",
		Address: r.URL.Host,
	}
	h.ruleMapper.Match(NewMatchMetaFromHTTPSHost(r.URL.Host)).Conn(proxyClient, &remote)
}

func (h *HTTPHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if !r.URL.IsAbs() {
		http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", 500)
		return
	}
	deleteRequestHeaders(r)
	h.ruleMapper.Match(NewMatchMetaFromHTTPRequest(r)).HTTPRequest(w, r)
}

func deleteRequestHeaders(req *http.Request) {
	req.RequestURI = "" // this must be reset when serving a request with the client
	// req.Header.Del("Accept-Encoding")
	for k := range ProxyHeaders {
		req.Header.Del(k)
	}
}

func NewHTTPHandler(mapper *RuleMapper) *HTTPHandler {
	h := &HTTPHandler{
		ruleMapper: mapper,
	}
	return h
}
