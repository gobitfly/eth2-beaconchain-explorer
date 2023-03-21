package types

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type EnsNameRegisteredIndexed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ParentHash      []byte               `protobuf:"bytes,1,opt,name=parent_hash,json=parentHash,proto3" json:"parent_hash,omitempty"`
	BlockNumber     uint64               `protobuf:"varint,2,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	ContractAddress []byte               `protobuf:"bytes,3,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	Time            *timestamp.Timestamp `protobuf:"bytes,4,opt,name=time,proto3" json:"time,omitempty"`
	Label           []byte               `protobuf:"bytes,7,opt,name=label,proto3" json:"label,omitempty"`
	Owner           []byte               `protobuf:"bytes,5,opt,name=owner,proto3" json:"owner,omitempty"`
	Resolver        []byte               `protobuf:"bytes,5,opt,name=resolver,proto3" json:"resolver ,omitempty"`
	Name            []byte               `protobuf:"bytes,6,opt,name=name,proto3" json:"name,omitempty"`
	Expires         *timestamp.Timestamp `protobuf:"bytes,7,opt,name=expires,proto3" json:"expires,omitempty"`
}

func (x *EnsNameRegisteredIndexed) Reset() {
	*x = EnsNameRegisteredIndexed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eth1_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EnsNameRegisteredIndexed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnsNameRegisteredIndexed) ProtoMessage() {}

func (x *EnsNameRegisteredIndexed) ProtoReflect() protoreflect.Message {
	mi := &file_eth1_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *EnsNameRegisteredIndexed) GetParentHash() []byte {
	if x != nil {
		return x.ParentHash
	}
	return nil
}

func (x *EnsNameRegisteredIndexed) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *EnsNameRegisteredIndexed) GetContractAddress() []byte {
	if x != nil {
		return x.ContractAddress
	}
	return nil
}

func (x *EnsNameRegisteredIndexed) GetTime() *timestamp.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

func (x *EnsNameRegisteredIndexed) GetOwner() []byte {
	if x != nil {
		return x.Owner
	}
	return nil
}

func (x *EnsNameRegisteredIndexed) GetName() []byte {
	if x != nil {
		return x.Name
	}
	return nil
}

func (x *EnsNameRegisteredIndexed) GetExpires() *timestamp.Timestamp {
	if x != nil {
		return x.Expires
	}
	return nil
}
