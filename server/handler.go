package server

import (
	"net"

	"github.com/mengseeker/nlink/core/transform"
)

type Conn interface {
	net.Conn

	// CloseWrite shuts down the writing side of the connection.
	CloseWrite() error

	// Disconnect closes the connection and cleans up any resources
	// must be called when the connection is no longer needed
	Disconnect(reason string) error
}

func ServeConn(conn net.Conn) {

}

func (s *Server) Serve(conn net.Conn) {
	defer conn.Close()
	pc, err := transform.AcceptPackConn(conn)
	if err != nil {
		logger.Error("ac pack conn", err)
		return
	}
	defer pc.Disconnect("serve done")

	for {
		meta, err := pc.Accept()
		if err != nil {
			logger.Error("accept ", err)
			return
		}
		logger.Infof("accept %v", meta)
		s.handleConnect(pc, meta)
	}

}

func (s *Server) handleConnect(conn Conn, meta *transform.Meta) {
	defer conn.Close()
	remoteConn, err := net.DialTimeout(meta.Net, meta.Addr, DialTimeout)
	if err != nil {
		logger.Error("dial remote: ", err)
		return
	}
	defer remoteConn.Close()

	transform.TransformConn(conn, remoteConn, logger)
}
