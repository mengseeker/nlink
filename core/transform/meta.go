package transform

import (
	"bytes"
	"errors"
)

type Meta struct {
	Net  string // tcp, udp
	Addr string // host:port
}

func (m *Meta) Marshal() []byte {
	return []byte(m.Net + "://" + m.Addr)
}

func (m *Meta) String() string {
	return m.Net + "://" + m.Addr
}

func (m *Meta) Network() string {
	return m.Net
}

func (m *Meta) Unmarshal(data []byte) error {
	parts := bytes.SplitN(data, []byte("://"), 2)
	if len(parts) != 2 {
		return errors.New("invalid meta data")
	}
	m.Net = string(parts[0])
	m.Addr = string(parts[1])
	return nil
}
