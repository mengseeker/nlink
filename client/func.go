package client

import (
	"context"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/geoip"
	"github.com/mengseeker/nlink/core/resolver"
)

type ProxyStream interface {
	Context() context.Context
	CloseSend() error
}

type Proxy_HTTPCallClient interface {
	ProxyStream
	Send(*api.HTTPRequest) error
	Recv() (*api.HTTPResponse, error)
}

type Proxy_TCPCallClient interface {
	ProxyStream
	Send(*api.SockRequest) error
	Recv() (*api.SockData, error)
}

type ForwardClient interface {
	HTTPCall(ctx context.Context) (Proxy_HTTPCallClient, error)
	TCPCall(ctx context.Context) (Proxy_TCPCallClient, error)
	ServerName() string
}

type IP struct {
	net.IP
	lastUsedTime int64
}

type FunctionProvider struct {
	ServerConfigs  map[string]ServerConfig
	Forwards       map[string]ForwardClient
	ReadBufferSize int

	resolvers   []resolver.Resolver
	domainCache map[string]*IP
}

func NewFunctionProvider(ctx context.Context, sc []ServerConfig, rc []ResolverConfig) (pv *FunctionProvider, err error) {
	pv = &FunctionProvider{
		Forwards:      make(map[string]ForwardClient),
		ServerConfigs: make(map[string]ServerConfig),
	}
	for i := range sc {
		pv.ServerConfigs[sc[i].Name] = sc[i]
	}
	pv.ReadBufferSize = 4 << 10

	// init resolvers
	pv.resolvers = make([]resolver.Resolver, len(rc))
	for _, c := range rc {
		if c.DNS != "" {
			rcs, err := resolver.NewDNSResolver(c.DNS)
			if err != nil {
				return nil, fmt.Errorf("new dns resolver err: %s", err)
			}
			pv.resolvers = append(pv.resolvers, rcs)
		} else if c.DoT != "" {
			rcs, err := resolver.NewDoTResolver(c.DoT)
			if err != nil {
				return nil, fmt.Errorf("new DoT resolver err: %s", err)
			}
			pv.resolvers = append(pv.resolvers, rcs)
		}
	}

	if len(pv.resolvers) == 0 {
		rcs, err := resolver.NewLocalResolver()
		if err != nil {
			pv.resolvers = append(pv.resolvers, rcs)
		}
	}

	// init geoip
	err = geoip.InitDB()
	if err != nil {
		return
	}

	pv.domainCache = make(map[string]*IP)
	// load hosts
	hosts, err := resolver.LoadHosts()
	if err != nil {
		return
	}
	for k := range hosts {
		pv.domainCache[k] = &IP{
			IP:           hosts[k],
			lastUsedTime: math.MaxInt64,
		}
	}

	go pv.handleMaintenance(ctx)

	return pv, nil
}

func (pv *FunctionProvider) handleMaintenance(ctx context.Context) {
	// TODO
}

func (pv *FunctionProvider) dialProxyServer(ctx context.Context, name string) (err error) {
	sc, exist := pv.ServerConfigs[name]
	if !exist {
		return fmt.Errorf("forward server %q not fround", name)
	}
	if sc.Net == "tcp" {
		cli, err := DialGrpcServer(ctx, sc)
		if err != nil {
			return err
		}
		pv.Forwards[name] = cli
	} else {
		cli, err := DialQuicServer(ctx, sc)
		if err != nil {
			return err
		}
		pv.Forwards[name] = cli
	}
	return
}

func (pv *FunctionProvider) GEOIP(ip net.IP) string {
	return geoip.Country(ip)
}

func (pv *FunctionProvider) Resolv(ctx context.Context, domain string) net.IP {
	domain = strings.SplitN(domain, ":", 2)[0]
	if ip, exist := pv.domainCache[domain]; exist {
		now := time.Now().Unix()
		if pv.domainCache[domain].lastUsedTime < now {
			pv.domainCache[domain].lastUsedTime = now
		}
		return ip.IP
	}
	ip := pv.resolv(ctx, domain)
	pv.domainCache[domain] = &IP{
		IP:           ip,
		lastUsedTime: time.Now().Unix(),
	}
	return ip
}

func (pv *FunctionProvider) resolv(ctx context.Context, domain string) (IP net.IP) {
	records := make(chan net.IP)
	defer func() {
		for range records {
		}
		close(records)
	}()
	tctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	call := func(rl resolver.Resolver) {
		defer recover()
		ip, err := rl.Resolv(tctx, domain)
		if err != nil {
			l.Debugf("lookup %s err: %v", domain, err)
			return
		}
		records <- ip
	}

	for k := range pv.resolvers {
		go call(pv.resolvers[k])
	}

	return <-records
}

func (pv *FunctionProvider) getForwardProxyClient(ctx context.Context, name string) (cli ForwardClient, err error) {
	cli, ok := pv.Forwards[name]
	if !ok {
		err = pv.dialProxyServer(ctx, name)
		if err != nil {
			return
		}
		cli = pv.Forwards[name]
	}
	return
}

func (pv *FunctionProvider) DialHTTP(ctx context.Context, name string) (stream Proxy_HTTPCallClient, err error) {
	cli, err := pv.getForwardProxyClient(ctx, name)
	if err != nil {
		return
	}
	stream, err = cli.HTTPCall(ctx)
	return
}

func (pv *FunctionProvider) DialTCP(ctx context.Context, name string) (stream Proxy_TCPCallClient, err error) {
	cli, err := pv.getForwardProxyClient(ctx, name)
	if err != nil {
		return
	}
	stream, err = cli.TCPCall(ctx)
	return
}
