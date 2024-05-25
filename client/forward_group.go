package client

import (
	"errors"
	"net"
	"net/http"

	"github.com/mengseeker/nlink/core/transform"
)

type ForwardGroup struct {
	cfg      ForwardGroupConfig
	servers  []*ForwardClient
	selecter Selecter
}

type ForwardGroupConfig struct {
	Name     string
	Servers  []string
	Selecter SelecterConfig
}

func NewForwardGroup(clients map[string]*ForwardClient, config ForwardGroupConfig) (*ForwardGroup, error) {
	servers := []*ForwardClient{}
	for _, server := range config.Servers {
		servers = append(servers, clients[server])
		if clients[server] == nil {
			return nil, errors.New("not found server: " + server)
		}
	}
	selecter, err := NewSelecter(servers, config.Selecter)
	if err != nil {
		return nil, err
	}
	return &ForwardGroup{
		cfg:      config,
		servers:  servers,
		selecter: selecter,
	}, nil
}

func (f *ForwardGroup) HTTPRequest(w http.ResponseWriter, r *http.Request) {
	remote := &transform.Meta{
		Net:  "tcp",
		Addr: r.URL.Host,
	}
	fc, err := f.selecter(remote)
	if err != nil {
		ResponseError(w, err)
		return
	}
	fc.HTTPRequest(w, r)
}

func (f *ForwardGroup) Conn(conn net.Conn, remote *transform.Meta) {
	fc, err := f.selecter(remote)
	if err != nil {
		conn.Close()
		return
	}
	fc.Conn(conn, remote)
}
