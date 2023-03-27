package types

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

type EnsNameRegisteredIndexed struct {
	ParentHash       []byte               `json:"parent_hash,omitempty"`
	BlockNumber      uint64               `json:"block_number,omitempty"`
	RegisterContract []byte               `json:"register_contract,omitempty"`
	ResolveContract  []byte               `json:"resolve_contract,omitempty"`
	Time             *timestamp.Timestamp `json:"time,omitempty"`
	Label            []byte               `json:"label,omitempty"`
	Owner            []byte               `json:"owner,omitempty"`
	Resolver         []byte               `json:"resolver ,omitempty"`
	Node             []byte               `json:"node ,omitempty"`
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

type EnsAddressChangedIndexed struct {
	ParentHash      []byte               `json:"parent_hash,omitempty"`
	BlockNumber     uint64               `json:"block_number,omitempty"`
	ResolveContract []byte               `json:"resolve_contract,omitempty"`
	Time            *timestamp.Timestamp `json:"time,omitempty"`
	Node            [32]byte             `json:"node,omitempty"`
	CoinType        uint64               `json:"coin_type,omitempty"`
	NewAddress      []byte               `json:"new_address ,omitempty"`
}

type EnsNameChangedIndexed struct {
	ParentHash      []byte               `json:"parent_hash,omitempty"`
	BlockNumber     uint64               `json:"block_number,omitempty"`
	ResolveContract []byte               `json:"resolve_contract,omitempty"`
	Time            *timestamp.Timestamp `json:"time,omitempty"`
	Node            [32]byte             `json:"node,omitempty"`
	NewName         string               `json:"new_name ,omitempty"`
}
