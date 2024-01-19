package client

import (
	"io"
	"net"
	"net/http"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
)

type RuleHandler interface {
	HTTPRequest(w http.ResponseWriter, r *http.Request)
	Conn(conn net.Conn, remote *api.ForwardMeta)
}

type RejectRuleHandler struct {
	log *log.Logger
}

func (h *RejectRuleHandler) HTTPRequest(w http.ResponseWriter, r *http.Request) {
	h.log.Info("reject request", "url", r.URL)
	http.Error(w, "reject", http.StatusForbidden)
}

func (h *RejectRuleHandler) Conn(conn net.Conn, remote *api.ForwardMeta) {
	h.log.Info("reject connect", "address", remote.Address)
	conn.Close()
}

type DirectRuleHandler struct {
	log *log.Logger
}

func (h *DirectRuleHandler) HTTPRequest(w http.ResponseWriter, r *http.Request) {
	h.log.Info("direct request", "url", r.URL)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		h.log.With("url", r.URL.String()).Errorf("http call err: %v", err)
		ResponseError(w, err)
		return
	}
	defer resp.Body.Close()
	CopyHTTPResponse(w, resp)
}

func (h *DirectRuleHandler) Conn(conn net.Conn, remote *api.ForwardMeta) {
	l := h.log.With("network", remote.Network, "address", remote.Address)
	l.Info("direct connect")
	defer conn.Close()
	remoteConn, err := net.Dial(remote.Network, remote.Address)
	if err != nil {
		l.Info("dial remote err: %v", err)
		return
	}
	defer remoteConn.Close()
	transform.ConnCopyAndWait(conn, remoteConn, l)
}

func ResponseError(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusBadGateway)
}

func CopyHTTPResponse(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
