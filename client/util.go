package client

import (
	"crypto/tls"
)

func NewClientTls(certFile, keyFile string) (tc *tls.Config, err error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	tc = &tls.Config{
		ServerName:         "x.test.example.com",
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	return
}
