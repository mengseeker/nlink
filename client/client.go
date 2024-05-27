package client

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ServerConfig struct {
	Name string
	Addr string
	Cert string
	Key  string

	MaxConns    int
	IdleTimeout time.Duration
	MaxIdle     int
}

type ResolverConfig struct {
	DNS string
	DoT string
}

type ProxyConfig struct {
	Listen   string
	Net      string
	Cert     string
	Key      string
	Rules    []string
	Servers  []ServerConfig
	Resolver []ResolverConfig
	Groups   []ForwardGroupConfig
}

func Start(cfg ProxyConfig) error {
	if cfg.Listen == "" {
		cfg.Listen = ":7890"
	}

	for i := range cfg.Servers {
		if cfg.Servers[i].Cert == "" {
			cfg.Servers[i].Cert = cfg.Cert
		}
		if cfg.Servers[i].Key == "" {
			cfg.Servers[i].Key = cfg.Key
		}
	}

	fmt.Printf("start proxy with config:\n")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(cfg)

	provider, err := NewFuncProvider(cfg.Resolver, cfg.Servers)
	if err != nil {
		return err
	}

	forwards := map[string]Forward{}
	forwardClients := map[string]*ForwardClient{}
	for _, sc := range cfg.Servers {
		forward, err := NewForwardClient(sc)
		if err != nil {
			return fmt.Errorf("new forwardclient %s err: %v", sc.Name, err)
		}
		forwards[sc.Name] = forward
		forwardClients[sc.Name] = forward
	}

	for _, gc := range cfg.Groups {
		if _, ok := forwards[gc.Name]; ok {
			return fmt.Errorf("group name conflict with server name: %s", gc.Name)
		}
		g, err := NewForwardGroup(forwardClients, gc)
		if err != nil {
			return fmt.Errorf("new group %s err: %v", gc.Name, err)
		}
		forwards[gc.Name] = g
	}

	mapper, err := NewRuleMapper(cfg.Rules, provider, forwards)
	if err != nil {
		return fmt.Errorf("parse rule err: %v", err)
	}
	httpHandler := NewHTTPHandler(mapper)
	socks4Handler := NewSocks4Handler(mapper)
	socks5Handler := NewSocks5Handler(mapper)

	lis := &Listener{
		Address:       cfg.Listen,
		HTTPHandler:   httpHandler,
		Socks4Handler: socks4Handler,
		Socks5Handler: socks5Handler,
	}
	return lis.ListenAndServe()
}
