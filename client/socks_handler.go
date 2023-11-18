package client

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/socks"
	"github.com/mengseeker/nlink/core/socks/transport/socks4"
	"github.com/mengseeker/nlink/core/socks/transport/socks5"
)

type SocksCommand uint8

type SockesHandler interface {
	HandleConn(net.Conn)
}

type Socks4Handler struct {
	mapper *RuleMapper
}

func (h *Socks4Handler) HandleConn(conn net.Conn) {
	addr, _, err := socks4.ServerHandshake(conn, nil)
	if err != nil {
		conn.Close()
		return
	}
	meta := socks.ParseSocksAddr(socks5.ParseAddr(addr))
	remote := api.ForwardMeta{
		Network: "tcp",
	}
	if meta.Host != "" {
		remote.Address = meta.Host
	} else {
		dst := meta.DstIP.String()
		if strings.Contains(dst, ":") {
			remote.Address = fmt.Sprintf("[%s]:%s", dst, meta.DstPort)
		}
		remote.Address = fmt.Sprintf("%s:%s", dst, meta.DstPort)
	}
	h.mapper.Match(NewMatchMetaFromSocksMeta(meta)).Conn(conn, &remote)
}

func NewSocks4Handler(mapper *RuleMapper) SockesHandler {
	return &Socks4Handler{mapper: mapper}
}

type Socks5Handler struct {
	mapper *RuleMapper
}

func (h *Socks5Handler) HandleConn(conn net.Conn) {
	target, command, err := socks5.ServerHandshake(conn, nil)
	if err != nil {
		conn.Close()
		return
	}
	if command == socks5.CmdUDPAssociate {
		defer conn.Close()
		io.Copy(io.Discard, conn)
		return
	}
	meta := socks.ParseSocksAddr(target)
	remote := api.ForwardMeta{
		Network: "tcp",
	}
	if meta.Host != "" {
		remote.Address = meta.Host + ":" + meta.DstPort
	} else {
		dst := meta.DstIP.String()
		if strings.Contains(dst, ":") {
			remote.Address = fmt.Sprintf("[%s]:%s", dst, meta.DstPort)
		}
		remote.Address = fmt.Sprintf("%s:%s", dst, meta.DstPort)
	}
	h.mapper.Match(NewMatchMetaFromSocksMeta(meta)).Conn(conn, &remote)
}

func NewSocks5Handler(mapper *RuleMapper) SockesHandler {
	return &Socks5Handler{mapper: mapper}
}
