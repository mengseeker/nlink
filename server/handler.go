package server

import (
	"net"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
)

type Handler struct {
	Log *log.Logger
	net.Dialer
}

func (h *Handler) HandleConnect(conn net.Conn) {
	defer conn.Close()
	var meta api.ForwardMeta
	err := transform.RecvMsg(conn, &meta)
	if err != nil {
		h.Log.Errorf("handleConnect read conn meta err: %v", err)
		return
	}
	l := h.Log.With("remote.network", meta.Network, "remote.address", meta.Address, "conn.id", meta.ID)
	l.Info("HandleConnect")
	remote, err := h.Dial(meta.Network, meta.Address)
	if err != nil {
		l.Errorf("handleConnect dial remote err: %v", err)
		return
	}

	defer remote.Close()
	transform.ConnCopyAndWait(conn, remote, l)
	l.Debugf("HandleConnect end")
}
