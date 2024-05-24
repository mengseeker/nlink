package client

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/mengseeker/nlink/core/log"

	"github.com/mengseeker/nlink/core/geoip"
	"github.com/mengseeker/nlink/core/resolver"
)

type FuncProvider struct {
	log       *log.Logger
	resolvers []resolver.Resolver
	hosts     map[string]net.IP
	servers   map[string]bool
}

func NewFuncProvider(rc []ResolverConfig, servers []ServerConfig) (pv *FuncProvider, err error) {
	pv = &FuncProvider{
		resolvers: make([]resolver.Resolver, 0),
		servers:   map[string]bool{},
	}

	for _, sc := range servers {
		pv.servers[sc.Name] = true
	}

	// init resolvers
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

	// load hosts
	pv.hosts, err = resolver.LoadHosts()
	if err != nil {
		return
	}

	return pv, nil
}

func (pv *FuncProvider) GEOIP(ip net.IP) string {
	return geoip.Country(ip)
}

func (pv *FuncProvider) HasServer(name string) bool {
	return pv.servers[name]
}

func (pv *FuncProvider) Resolv(domain string) (IP net.IP) {
	records := make(chan net.IP)
	defer close(records)
	tctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var i atomic.Bool
	call := func(rl resolver.Resolver) {
		ip, err := rl.Resolv(tctx, domain)
		if err != nil {
			pv.log.Debugf("lookup %s err: %v", domain, err)
			return
		}
		if i.CompareAndSwap(false, true) {
			select {
			case records <- ip:
			case <-tctx.Done():
			}
		}
	}
	for i := range pv.resolvers {
		go call(pv.resolvers[i])
	}

	select {
	case ip := <-records:
		return ip
	case <-tctx.Done():
		return nil
	}
}
