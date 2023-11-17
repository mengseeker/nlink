package resolver

import (
	"context"
	"testing"
)

func TestDotResolver_Resolv(t *testing.T) {
	rl, err := NewDoTResolver("dns.alidns.com")
	if err != nil {
		t.Fatal(err)
	}
	domain := "www.baidu.com"
	ip, err := rl.Resolv(context.Background(), domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("lookup %s -> %s", domain, ip.String())
	domain = "123.242.123.1"
	ip, err = rl.Resolv(context.Background(), domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("lookup %s -> %s", domain, ip.String())

}

func TestDNSResolver_Resolv(t *testing.T) {
	rl, err := NewDNSResolver("114.114.114.114")
	if err != nil {
		t.Fatal(err)
	}
	domain := "www.baidu.com"
	ip, err := rl.Resolv(context.Background(), domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("lookup %s -> %s", domain, ip.String())
	domain = "123.242.123.1"
	ip, err = rl.Resolv(context.Background(), domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("lookup %s -> %s", domain, ip.String())

}

func TestLocalResolver_Resolv(t *testing.T) {
	rl, err := NewLocalResolver()
	if err != nil {
		t.Fatal(err)
	}
	domain := "www.baidu.com"
	ip, err := rl.Resolv(context.Background(), domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("lookup %s -> %s", domain, ip.String())
	domain = "123.242.123.1"
	ip, err = rl.Resolv(context.Background(), domain)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("lookup %s -> %s", domain, ip.String())

}
