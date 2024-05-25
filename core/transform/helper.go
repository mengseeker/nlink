package transform

import (
	"io"
	"net"
	"sync"

	"github.com/mengseeker/nlink/core/log"
)

const (
	BUFFER_SIZE = PACK_MAX_DATA_LEN
)

var (
	logger = log.NewLogger()
)

func TransformConn(local, remote net.Conn, logger *log.Logger) {
	// logger.Infof("transforming %v <-> %v", local.RemoteAddr(), remote.RemoteAddr())
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer CloseWrite(local)
		_, err := io.Copy(local, remote)
		if err != nil && logger != nil {
			logger.Debugf("copy from remote to local error: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		defer CloseWrite(remote)
		_, err := io.Copy(remote, local)
		if err != nil && logger != nil {
			logger.Debugf("copy from local to remote error: %v", err)
		}
	}()
	wg.Wait()
}

type ConnCloseWriter interface {
	CloseWrite() error
}

func CloseWrite(c net.Conn) error {
	if c == nil {
		return nil
	}
	if wc, ok := c.(ConnCloseWriter); ok {
		return wc.CloseWrite()
	}
	return c.Close()
}
