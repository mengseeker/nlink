package resolver

import (
	"context"
	"net"
	"strings"

	"github.com/ncruces/go-dns"
)

type Resolver interface {
	Resolv(ctx context.Context, domain string) (net.IP, error)
}

type resolver struct {
	*net.Resolver
	Server string
}

func (r *resolver) Resolv(ctx context.Context, domain string) (net.IP, error) {
	ips, err := r.Resolver.LookupIP(ctx, "ip", domain)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, &net.DNSError{Err: "no ip address", Server: r.Server, Name: domain}
	}
	for _, ip := range ips {
		if ip = ip.To4(); ip != nil {
			return ip, nil
		}
	}
	return ips[0], err
}

func NewDoTResolver(server string) (*resolver, error) {
	r, err := dns.NewDoTResolver(server)
	if err != nil {
		return nil, err
	}
	return &resolver{
		Resolver: r,
		Server:   server,
	}, nil
}

func NewLocalResolver() (*resolver, error) {
	return &resolver{
		Resolver: &net.Resolver{},
		Server:   "local",
	}, nil
}

func NewDNSResolver(server string) (*resolver, error) {
	if !strings.Contains(server, ":") {
		server = server + ":53"
	}
	var d net.Dialer
	return &resolver{
		Resolver: &net.Resolver{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return d.DialContext(ctx, network, server)
			},
		},
		Server: server,
	}, nil
}
