package client

type FunctionProvider struct {
}

type Country string

const (
	China Country = "CN"
)

func (pv *FunctionProvider) GEOIP(ip string) Country {
	// TODO
	return China
}

func (pv *FunctionProvider) Resolver(domain string) (IP string) {
	// TODO
	return
}

func (pv *FunctionProvider) Forward(name string) {
	// TODO
}
