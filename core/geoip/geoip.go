package geoip

import (
	"bytes"
	_ "embed"
	"io"
	"net"

	"github.com/oschwald/geoip2-golang"
	"github.com/xi2/xz"
)

//go:embed Country.mmdb.xz
var dbBytes []byte

var db *geoip2.Reader

func InitDB() (err error) {
	r, err := xz.NewReader(bytes.NewReader(dbBytes), 0)
	if err != nil {
		return
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		return
	}
	db, err = geoip2.FromBytes(raw)
	return
}

func Country(ip net.IP) string {
	c, _ := db.Country(ip)
	if c != nil {
		return c.Country.IsoCode
	}
	return ""
}
