package server

import (
	"io"
	"net"
	"sync"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
)

type Handler struct {
	Log *log.Logger
	net.Dialer
}

func (h *Handler) HandleConnect(conn io.ReadWriteCloser) {
	defer conn.Close()
	var meta api.ForwardMeta
	err := transform.RecvMsg(conn, &meta)
	if err != nil {
		h.Log.Errorf("handleConnect read conn meta err: %v", err)
		return
	}
	l := h.Log.With("network", meta.Network, "address", meta.Address)
	l.Info("dial remote")
	remote, err := h.Dial(meta.Network, meta.Address)
	if err != nil {
		l.Errorf("handleConnect dial remote err: %v", err)
		return
	}
	defer remote.Close()
	wg := sync.WaitGroup{}
	copy := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		_, err := io.Copy(w, r)
		if err != nil {
			l.Errorf("copy data err: %v", err)
		}
	}
	wg.Add(2)
	go copy(remote, conn)
	go copy(conn, remote)
	wg.Wait()
}
