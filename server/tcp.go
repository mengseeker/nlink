package server

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/mengseeker/nlink/core/api"
)

func (s *Server) TCPCall(stream api.Proxy_TCPCallServer) (err error) {
	r, err := stream.Recv()
	if err != nil {
		return err
	}
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.Req.Host, r.Req.Port))
	if err != nil {
		return fmt.Errorf("dial remote err %v", err)
	}
	defer remoteConn.Close()
	// try notify client
	err = stream.Send(&api.SockData{})
	if err != nil {
		return fmt.Errorf("notify client err %v", err)
	}

	// read remote data and write back
	done := make(chan int)
	go func() {
		defer close(done)
		remoteReadBuff := make([]byte, s.Config.ReadBufferSize)
		for {
			n, err := remoteConn.Read(remoteReadBuff)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				slog.Error("read remote err", "error", err)
				return
			}
			back := api.SockData{
				Data: remoteReadBuff[:n],
			}
			err = stream.Send(&back)
			if err != nil {
				slog.Error("write back err", "error", err)
				return
			}
		}
	}()

	// read data and write to remote
	if r.Data != nil {
		_, err = remoteConn.Write(r.Data.Data)
		if err != nil {
			slog.Error("write remote err", "error", err)
			return
		}
	}
	for {
		r, err = stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			}
			return err
		}
		slog.Error("read req err", "error", err)
		if r.Data != nil {
			_, err = remoteConn.Write(r.Data.Data)
			if err != nil {
				slog.Error("write remote err", "error", err)
				return err
			}
		}
	}

	<-done
	return
}
