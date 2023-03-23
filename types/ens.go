package types

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

type EnsNameRegisteredIndexed struct {
	ParentHash       []byte               `json:"parent_hash,omitempty"`
	BlockNumber      uint64               `json:"block_number,omitempty"`
	RegisterContract []byte               `json:"register_contract,omitempty"`
	Time             *timestamp.Timestamp `json:"time,omitempty"`
	Label            []byte               `json:"label,omitempty"`
	Owner            []byte               `json:"owner,omitempty"`
	Resolver         []byte               `json:"resolver ,omitempty"`
	Name             []byte               `json:"name,omitempty"`
	Expires          *timestamp.Timestamp `json:"expires,omitempty"`
}
type EnsNameRenewedIndexed struct {
	ParentHash  []byte               `json:"parent_hash,omitempty"`
	BlockNumber uint64               `json:"block_number,omitempty"`
	Time        *timestamp.Timestamp `json:"time,omitempty"`
	Label       []byte               `json:"label,omitempty"`
	Name        []byte               `json:"name,omitempty"`
	Expires     *timestamp.Timestamp `json:"expires,omitempty"`
}
