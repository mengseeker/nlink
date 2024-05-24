package client

import (
	"errors"
	"net"
	"net/http"

	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/transform"
)

type ForwardGroup struct {
	cfg      ForwardGroupConfig
	log      *log.Logger
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
		log:      log.With("group", config.Name),
		servers:  servers,
		selecter: selecter,
	}, nil
}

func (f *ForwardGroup) HTTPRequest(w http.ResponseWriter, r *http.Request) {
	remote := &transform.Meta{
		Network: "tcp",
		Address: r.URL.Host,
	}
	fc, err := f.selecter(remote)
	if err != nil {
		f.log.Errorf("select err: %v", err)
		ResponseError(w, err)
		return
	}
	fc.HTTPRequest(w, r)
}

func (f *ForwardGroup) Conn(conn net.Conn, remote *transform.Meta) {
	fc, err := f.selecter(remote)
	if err != nil {
		f.log.Errorf("select err: %v", err)
		conn.Close()
		return
	}
	fc.Conn(conn, remote)
}
