package transform

import (
	"bytes"
	"errors"
)

type Meta struct {
	Network string // tcp, udp
	Address string // host:port
}

func (m *Meta) Marshal() []byte {
	return []byte(m.Network + "://" + m.Address)
}

func (m *Meta) String() string {
	return m.Network + "://" + m.Address
}

func (m *Meta) Unmarshal(data []byte) error {
	parts := bytes.SplitN(data, []byte("://"), 2)
	if len(parts) != 2 {
		return errors.New("invalid meta data")
	}
	m.Network = string(parts[0])
	m.Address = string(parts[1])
	return nil
}
