package client

import (
	"context"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/mengseeker/nlink/core/log"

	"github.com/mengseeker/nlink/core/geoip"
	"github.com/mengseeker/nlink/core/resolver"
)

type IP struct {
	net.IP
	lastUsedTime int64
}

type FunctionProvider struct {
	log         *log.Logger
	resolvers   []resolver.Resolver
	domainCache map[string]*IP
	dcLock      sync.Mutex
}

func NewFunctionProvider(ctx context.Context, rc []ResolverConfig, l *log.Logger) (pv *FunctionProvider, err error) {
	pv = &FunctionProvider{
		log:         l,
		dcLock:      sync.Mutex{},
		resolvers:   make([]resolver.Resolver, 0, len(rc)),
		domainCache: make(map[string]*IP),
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
	tk := time.NewTicker(time.Minute)
	defer tk.Stop()
	clean := func() {
		pv.dcLock.Lock()
		defer pv.dcLock.Unlock()
		expireTime := time.Now().Add(-time.Minute * 10).Unix()
		for k, v := range pv.domainCache {
			if v.lastUsedTime < expireTime {
				delete(pv.domainCache, k)
			}
		}
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			clean()
		}
	}
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
	pv.dcLock.Lock()
	defer pv.dcLock.Unlock()
	ip := pv.resolv(ctx, domain)
	pv.domainCache[domain] = &IP{
		IP:           ip,
		lastUsedTime: time.Now().Unix(),
	}
	return ip
}

func (pv *FunctionProvider) resolv(ctx context.Context, domain string) (IP net.IP) {
	records := make(chan net.IP)
	wg := sync.WaitGroup{}
	tctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	call := func(rl resolver.Resolver) {
		defer recover()
		defer wg.Done()
		ip, err := rl.Resolv(tctx, domain)
		if err != nil {
			pv.log.Debugf("lookup %s err: %v", domain, err)
			return
		}
		records <- ip
	}
	wg.Add(len(pv.resolvers))
	go func() {
		wg.Wait()
		close(records)
	}()
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
