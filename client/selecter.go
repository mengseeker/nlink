package client

import (
	"errors"
	"fmt"

	"github.com/mengseeker/nlink/core/api"
)

type Selecter func(remote *api.ForwardMeta) (selected *ForwardClient, err error)

type SelecterType string

const (
	SelecterType_FrontFirst        SelecterType = "frontfirst"
	SelecterType_RoundRobin        SelecterType = "roundrobin"
	SelecterType_Random            SelecterType = "random"
	SelecterType_Hash              SelecterType = "hash"
	SelecterType_LeastConn         SelecterType = "leastconn"
	SelecterType_LeastTTL          SelecterType = "leastttl"
	SelecterType_LeastConnWeighted SelecterType = "leastconnweighted"
	SelecterType_ConsistentHash    SelecterType = "consistenthash"
)

type SelecterConfig struct {
	Type SelecterType
}

func NewSelecter(clients []*ForwardClient, selecterConfig SelecterConfig) (Selecter, error) {
	switch selecterConfig.Type {
	case SelecterType_FrontFirst:
		return NewFrontFirstSelecter(clients), nil
	case SelecterType_RoundRobin:
		return nil, fmt.Errorf("selecter type %q not implemented", selecterConfig.Type)
	case SelecterType_Random:
		return nil, fmt.Errorf("selecter type %q not implemented", selecterConfig.Type)
	case SelecterType_Hash:
		return nil, fmt.Errorf("selecter type %q not implemented", selecterConfig.Type)
	case SelecterType_LeastConn:
		return nil, fmt.Errorf("selecter type %q not implemented", selecterConfig.Type)
	case SelecterType_LeastTTL:
		return nil, fmt.Errorf("selecter type %q not implemented", selecterConfig.Type)
	case SelecterType_LeastConnWeighted:
		return nil, fmt.Errorf("selecter type %q not implemented", selecterConfig.Type)
	case SelecterType_ConsistentHash:
		return nil, fmt.Errorf("selecter type %q not implemented", selecterConfig.Type)
	}
	return nil, errors.New("not support selecter type: " + string(selecterConfig.Type))
}

func NewFrontFirstSelecter(clients []*ForwardClient) Selecter {
	return func(remote *api.ForwardMeta) (selected *ForwardClient, err error) {
		return clients[0], nil
	}
}
