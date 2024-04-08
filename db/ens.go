package db

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"eth2-exporter/ens"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"golang.org/x/sync/errgroup"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	go_ens "github.com/wealdtech/go-ens/v3"
	"github.com/wealdtech/go-ens/v3/contracts/registry"
)

// https://etherscan.io/tx/0x9fec76750a504e5610643d1882e3b07f4fc786acf7b9e6680697bb7165de1165#eventlog
// TransformEnsNameRegistered accepts an eth1 block and creates bigtable mutations for ENS Name events.
// It transforms the logs contained within a block and indexes ens relevant transactions and tags changes (to be verified from the node in a separate process)
// ==================================================
//
// It indexes transactions
//
// - by hashed ens name
// Row:    <chainID>:ENS:I:H:<nameHash>:<txHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:I:H:4ae569dd0aa2f6e9207e41423c956d0d27cbc376a499ee8d90fe1d84489ae9d1:e627ae94bd16eb1ed8774cd4003fc25625159f13f8a2612cc1c7f8d2ab11b1d7"
//
// - by address
// Row:    <chainID>:ENS:I:A:<address>:<txHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:I:A:05579fadcf7cc6544f7aa018a2726c85251600c5:e627ae94bd16eb1ed8774cd4003fc25625159f13f8a2612cc1c7f8d2ab11b1d7"
//
// ==================================================
//
// Track for later verification via the node ("set dirty")
//
// - by name
// Row:    <chainID>:ENS:V:N:<name>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:V:N:somename"
//
// - by name hash
// Row:    <chainID>:ENS:V:H:<nameHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:V:H:6f5d9cc23e60abe836401b4fd386ec9280a1f671d47d9bf3ec75dab76380d845"
//
// - by address
// Row:    <chainID>:ENS:V:A:<address>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:V:A:27234cb8734d5b1fac0521c6f5dc5aebc6e839b6"
//
// ==================================================

/*
0x335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a0 NewResolverTopic
0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f NameRegisteredTopic
0x69e37f151eb98a09618ddaa80c8cfaf1ce5996867c489f45b555b412271ebf27 NameRegisteredV2Topic
0x3da24c024582931cfaf8267d8ed24d13a82a8068d5bd337d30ec45cea4e506ae NameRenewedTopic
0x65412581168e88a1e60c6459d7f44ae83ad0832e670826c05a4e2476b57af752 AddressChangedTopic
0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7 NameChangedTopic
0xce0457fe73731f824cc272376169235128c118b49d344817417c6d108d155e82 NewOwnerTopic

0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5
0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e
0x253553366Da8546fC250F225fe3d25d0C782303b


// https://go.dev/play/p/F-TVEguGcpK

Registry                       E 0xce0457fe73731f824cc272376169235128c118b49d344817417c6d108d155e82 NewOwner(bytes32,bytes32,address)
Registry                       E 0x335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a0 NewResolver(bytes32,address)
Registry                       E 0x1d4f9bbfc9cab89d66e1a1562f2233ccbf1308cb4f63de2ead5787adddb8fa68 NewTTL(bytes32,uint64)
Registry                       E 0xd4735d920b0f87494915f556dd9b54c8f309026070caea5c737245152564d266 Transfer(bytes32,address)
Registry                       E 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31 ApprovalForAll(address,address,bool)
Registry                       M 0x14ab9038 setTTL(bytes32,uint64)
Registry                       M 0xe985e9c5 isApprovedForAll(address,address)
Registry                       M 0xb83f8663 old()
Registry                       M 0xf79fe538 recordExists(bytes32)
Registry                       M 0x0178b8bf resolver(bytes32)
Registry                       M 0xa22cb465 setApprovalForAll(address,bool)
Registry                       M 0xcf408823 setRecord(bytes32,address,address,uint64)
Registry                       M 0x06ab5923 setSubnodeOwner(bytes32,bytes32,address)
Registry                       M 0x16a25cbd ttl(bytes32)
Registry                       M 0x02571be3 owner(bytes32)
Registry                       M 0x5b0fc9c3 setOwner(bytes32,address)
Registry                       M 0x1896f70a setResolver(bytes32,address)
Registry                       M 0x5ef2c7f0 setSubnodeRecord(bytes32,bytes32,address,address,uint64)
BaseRegistrar                  E 0xb3d987963d01b2f68493b4bdb130988f157ea43070d4ad840fee0466ed9370d9 NameRegistered(uint256,address,uint256)
BaseRegistrar                  E 0x9b87a00e30f1ac65d898f070f8a3488fe60517182d0a2098e1b4b93a54aa9bd6 NameRenewed(uint256,uint256)
BaseRegistrar                  E 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef Transfer(address,address,uint256)
BaseRegistrar                  E 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31 ApprovalForAll(address,address,bool)
BaseRegistrar                  E 0x33d83959be2573f5453b12eb9d43b3499bc57d96bd2f067ba44803c859e81113 ControllerRemoved(address)
BaseRegistrar                  E 0xea3d7e1195a15d2ddcd859b01abd4c6b960fa9f9264e499a70a90c7f0c64b717 NameMigrated(uint256,address,uint256)
BaseRegistrar                  E 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925 Approval(address,address,uint256)
BaseRegistrar                  E 0x0a8bb31534c0ed46f380cb867bd5c803a189ced9a764e30b3a4991a9901d7474 ControllerAdded(address)
BaseRegistrar                  E 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0 OwnershipTransferred(address,address)
BaseRegistrar                  M 0xa22cb465 setApprovalForAll(address,bool)
BaseRegistrar                  M 0x4e543b26 setResolver(address)
BaseRegistrar                  M 0x01ffc9a7 supportsInterface(bytes4)
BaseRegistrar                  M 0xf2fde38b transferOwnership(address)
BaseRegistrar                  M 0xc1a287e2 GRACE_PERIOD()
BaseRegistrar                  M 0xa7fc7a07 addController(address)
BaseRegistrar                  M 0x96e494e8 available(uint256)
BaseRegistrar                  M 0x081812fc getApproved(uint256)
BaseRegistrar                  M 0xc475abff renew(uint256,uint256)
BaseRegistrar                  M 0x70a08231 balanceOf(address)
BaseRegistrar                  M 0xda8c229e controllers(address)
BaseRegistrar                  M 0xe985e9c5 isApprovedForAll(address,address)
BaseRegistrar                  M 0x3f15457f ens()
BaseRegistrar                  M 0xb88d4fde safeTransferFrom(address,address,uint256,bytes)
BaseRegistrar                  M 0x23b872dd transferFrom(address,address,uint256)
BaseRegistrar                  M 0xddf7fcb0 baseNode()
BaseRegistrar                  M 0x6352211e ownerOf(uint256)
BaseRegistrar                  M 0x095ea7b3 approve(address,uint256)
BaseRegistrar                  M 0x0e297b45 registerOnly(uint256,address,uint256)
BaseRegistrar                  M 0x42842e0e safeTransferFrom(address,address,uint256)
BaseRegistrar                  M 0x8f32d59b isOwner()
BaseRegistrar                  M 0xd6e4fa86 nameExpires(uint256)
BaseRegistrar                  M 0x8da5cb5b owner()
BaseRegistrar                  M 0xfca247ac register(uint256,address,uint256)
BaseRegistrar                  M 0x715018a6 renounceOwnership()
BaseRegistrar                  M 0x28ed4f6c reclaim(uint256,address)
BaseRegistrar                  M 0xf6a74ed7 removeController(address)
ETHRegistrarController         E 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0 OwnershipTransferred(address,address)
ETHRegistrarController         E 0x69e37f151eb98a09618ddaa80c8cfaf1ce5996867c489f45b555b412271ebf27 NameRegistered(string,bytes32,address,uint256,uint256,uint256)
ETHRegistrarController         E 0x3da24c024582931cfaf8267d8ed24d13a82a8068d5bd337d30ec45cea4e506ae NameRenewed(string,bytes32,uint256,uint256)
ETHRegistrarController         M 0x5d3590d5 recoverFunds(address,address,uint256)
ETHRegistrarController         M 0x715018a6 renounceOwnership()
ETHRegistrarController         M 0x83e7f6ff rentPrice(string,uint256)
ETHRegistrarController         M 0x9791c097 valid(string)
ETHRegistrarController         M 0x3ccfd60b withdraw()
ETHRegistrarController         M 0xa8e5fbc0 nameWrapper()
ETHRegistrarController         M 0x839df945 commitments(bytes32)
ETHRegistrarController         M 0x8d839ffe minCommitmentAge()
ETHRegistrarController         M 0x8da5cb5b owner()
ETHRegistrarController         M 0x74694a2b register(string,address,uint256,bytes32,address,bytes[],bool,uint16)
ETHRegistrarController         M 0x8a95b09f MIN_REGISTRATION_DURATION()
ETHRegistrarController         M 0xacf1a841 renew(string,uint256)
ETHRegistrarController         M 0x01ffc9a7 supportsInterface(bytes4)
ETHRegistrarController         M 0xf2fde38b transferOwnership(address)
ETHRegistrarController         M 0xf14fcbc8 commit(bytes32)
ETHRegistrarController         M 0x65a69dcf makeCommitment(string,address,uint256,bytes32,address,bytes[],bool,uint16)
ETHRegistrarController         M 0xce1e09c0 maxCommitmentAge()
ETHRegistrarController         M 0xd3419bf3 prices()
ETHRegistrarController         M 0x80869853 reverseRegistrar()
ETHRegistrarController         M 0xaeb8ce9b available(string)
DNSRegistrar                   E 0x87db02a0e483e2818060eddcbb3488ce44e35aff49a70d92c2aa6c8046cf01e2 Claim(bytes32,address,bytes,uint32)
DNSRegistrar                   E 0x9176b7f47e4504df5e5516c99d90d82ac7cbd49cc77e7f22ba2ac2f2e3a3eba8 NewPublicSuffixList(address)
DNSRegistrar                   M 0x6f951221 enableNode(bytes)
DNSRegistrar                   M 0x3f15457f ens()
DNSRegistrar                   M 0x25916d41 inceptions(bytes32)
DNSRegistrar                   M 0x7dc0d1d0 oracle()
DNSRegistrar                   M 0x29d56630 proveAndClaim(bytes,(bytes,bytes)[])
DNSRegistrar                   M 0x04f3bcec resolver()
DNSRegistrar                   M 0x1ecfc411 setPublicSuffixList(address)
DNSRegistrar                   M 0x30349ebe suffixes()
DNSRegistrar                   M 0x01ffc9a7 supportsInterface(bytes4)
DNSRegistrar                   M 0xab14ec59 previousRegistrar()
DNSRegistrar                   M 0x06963218 proveAndClaimWithResolver(bytes,(bytes,bytes)[],address,address)
ReverseRegistrar               E 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87 ControllerChanged(address,bool)
ReverseRegistrar               E 0xeae17a84d9eb83d8c8eb317f9e7d64857bc363fa51674d996c023f4340c577cf DefaultResolverChanged(address)
ReverseRegistrar               E 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0 OwnershipTransferred(address,address)
ReverseRegistrar               E 0x6ada868dd3058cf77a48a74489fd7963688e5464b2b0fa957ace976243270e92 ReverseClaimed(address,bytes32)
ReverseRegistrar               M 0xc66485b2 setDefaultResolver(address)
ReverseRegistrar               M 0x0f5a5466 claimWithResolver(address,address)
ReverseRegistrar               M 0xda8c229e controllers(address)
ReverseRegistrar               M 0xe0dba60f setController(address,bool)
ReverseRegistrar               M 0x715018a6 renounceOwnership()
ReverseRegistrar               M 0x828eab0e defaultResolver()
ReverseRegistrar               M 0x3f15457f ens()
ReverseRegistrar               M 0x8da5cb5b owner()
ReverseRegistrar               M 0x7a806d6b setNameForAddr(address,address,address,string)
ReverseRegistrar               M 0xc47f0027 setName(string)
ReverseRegistrar               M 0xf2fde38b transferOwnership(address)
ReverseRegistrar               M 0x1e83409a claim(address)
ReverseRegistrar               M 0x65669631 claimForAddr(address,address,address)
ReverseRegistrar               M 0xbffbe61c node(address)
NameWrapper                    E 0xee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c4 NameUnwrapped(bytes32,address)
NameWrapper                    E 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb TransferBatch(address,address,address,uint256[],uint256[])
NameWrapper                    E 0x6bb7ff708619ba0610cba295a58592e0451dee2622938c8755667688daf3529b URI(string,uint256)
NameWrapper                    E 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925 Approval(address,address,uint256)
NameWrapper                    E 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31 ApprovalForAll(address,address,bool)
NameWrapper                    E 0xf675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b132 ExpiryExtended(bytes32,uint64)
NameWrapper                    E 0x39873f00c80f4f94b7bd1594aebcf650f003545b74824d57ddf4939e3ff3a34b FusesSet(bytes32,uint32)
NameWrapper                    E 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87 ControllerChanged(address,bool)
NameWrapper                    E 0x8ce7013e8abebc55c3890a68f5a27c67c3f7efa64e584de5fb22363c606fd340 NameWrapped(bytes32,bytes,address,uint32,uint64)
NameWrapper                    E 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0 OwnershipTransferred(address,address)
NameWrapper                    E 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62 TransferSingle(address,address,address,uint256,uint256)
NameWrapper                    M 0xadf4960a allFusesBurned(bytes32,uint32)
NameWrapper                    M 0x00fdd58e balanceOf(address,uint256)
NameWrapper                    M 0xc475abff renew(uint256,uint256)
NameWrapper                    M 0xd8c9921a unwrap(bytes32,bytes32,address)
NameWrapper                    M 0x8b4dfa75 unwrapETH2LD(bytes32,address,address)
NameWrapper                    M 0x8cf8b41e wrapETH2LD(string,address,uint16,address)
NameWrapper                    M 0xeb8ae530 wrap(bytes,address,address)
NameWrapper                    M 0x3f15457f ens()
NameWrapper                    M 0x081812fc getApproved(uint256)
NameWrapper                    M 0x20c38e2b names(bytes32)
NameWrapper                    M 0x150b7a02 onERC721Received(address,address,uint256,bytes)
NameWrapper                    M 0xc658e086 setSubnodeOwner(bytes32,string,address,uint32,uint64)
NameWrapper                    M 0x24c1af44 setSubnodeRecord(bytes32,string,address,address,uint64,uint32,uint64)
NameWrapper                    M 0xc93ab3fd upgrade(bytes,bytes)
NameWrapper                    M 0xed70554d _tokens(uint256)
NameWrapper                    M 0x6e5d6ad2 extendExpiry(bytes32,bytes32,uint64)
NameWrapper                    M 0xd9a50c12 isWrapped(bytes32,bytes32)
NameWrapper                    M 0xfd0cd0d9 isWrapped(bytes32)
NameWrapper                    M 0x6352211e ownerOf(uint256)
NameWrapper                    M 0x0e89341c uri(uint256)
NameWrapper                    M 0x53095467 metadataService()
NameWrapper                    M 0x2b20e397 registrar()
NameWrapper                    M 0x2eb2c2d6 safeBatchTransferFrom(address,address,uint256[],uint256[],bytes)
NameWrapper                    M 0xa22cb465 setApprovalForAll(address,bool)
NameWrapper                    M 0xe0dba60f setController(address,bool)
NameWrapper                    M 0xcf408823 setRecord(bytes32,address,address,uint64)
NameWrapper                    M 0x1896f70a setResolver(bytes32,address)
NameWrapper                    M 0x1f4e1504 upgradeContract()
NameWrapper                    M 0x41415eab canModifyName(bytes32,address)
NameWrapper                    M 0xda8c229e controllers(address)
NameWrapper                    M 0x0178fe3f getData(uint256)
NameWrapper                    M 0x06fdde03 name()
NameWrapper                    M 0x5d3590d5 recoverFunds(address,address,uint256)
NameWrapper                    M 0xf242432a safeTransferFrom(address,address,uint256,uint256,bytes)
NameWrapper                    M 0x402906fc setFuses(bytes32,uint16)
NameWrapper                    M 0xa4014982 registerAndWrapETH2LD(string,address,uint256,address,uint16)
NameWrapper                    M 0xb6bcad26 setUpgradeContract(address)
NameWrapper                    M 0x095ea7b3 approve(address,uint256)
NameWrapper                    M 0x8da5cb5b owner()
NameWrapper                    M 0x715018a6 renounceOwnership()
NameWrapper                    M 0x1534e177 setMetadataService(address)
NameWrapper                    M 0x01ffc9a7 supportsInterface(bytes4)
NameWrapper                    M 0xf2fde38b transferOwnership(address)
NameWrapper                    M 0x4e1273f4 balanceOfBatch(address[],uint256[])
NameWrapper                    M 0x0e4cd725 canExtendSubnames(bytes32,address)
NameWrapper                    M 0xe985e9c5 isApprovedForAll(address,address)
NameWrapper                    M 0x33c69ea9 setChildFuses(bytes32,bytes32,uint32,uint64)
NameWrapper                    M 0x14ab9038 setTTL(bytes32,uint64)
PublicResolver                 E 0xc6621ccb8f3f5a04bb6502154b2caf6adf5983fe76dfef1cfc9c42e3579db444 VersionChanged(bytes32,uint64)
PublicResolver                 E 0x52d7d861f09ab3d26239d492e8968629f95e9e318cf0b73bfddc441522a15fd2 AddrChanged(bytes32,address)
PublicResolver                 E 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31 ApprovalForAll(address,address,bool)
PublicResolver                 E 0xf0ddb3b04746704017f9aa8bd728fcc2c1d11675041205350018915f5e4750a0 Approved(address,bytes32,address,bool)
PublicResolver                 E 0x8f15ed4b723ef428f250961da8315675b507046737e19319fc1a4d81bfe87f85 DNSZonehashChanged(bytes32,bytes,bytes)
PublicResolver                 E 0x52a608b3303a48862d07a73d82fa221318c0027fbbcfb1b2329bface3f19ff2b DNSRecordChanged(bytes32,bytes,uint16,bytes)
PublicResolver                 E 0x03528ed0c2a3ebc993b12ce3c16bb382f9c7d88ef7d8a1bf290eaf35955a1207 DNSRecordDeleted(bytes32,bytes,uint16)
PublicResolver                 E 0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7 NameChanged(bytes32,string)
PublicResolver                 E 0x1d6f5e03d3f63eb58751986629a5439baee5079ff04f345becb66e23eb154e46 PubkeyChanged(bytes32,bytes32,bytes32)
PublicResolver                 E 0x448bc014f1536726cf8d54ff3d6481ed3cbc683c2591ca204274009afa09b1a1 TextChanged(bytes32,string,string,string)
PublicResolver                 E 0xaa121bbeef5f32f5961a2a28966e769023910fc9479059ee3495d4c1a696efe3 ABIChanged(bytes32,uint256)
PublicResolver                 E 0x65412581168e88a1e60c6459d7f44ae83ad0832e670826c05a4e2476b57af752 AddressChanged(bytes32,uint256,bytes)
PublicResolver                 E 0xe379c1624ed7e714cc0937528a32359d69d5281337765313dba4e081b72d7578 ContenthashChanged(bytes32,bytes)
PublicResolver                 E 0x7c69f06bea0bdef565b709e93a147836b0063ba2dd89f02d0b7e8d931e6a6daa InterfaceChanged(bytes32,bytes4,address)
PublicResolver                 M 0xa4b91a01 approve(bytes32,address,bool)
PublicResolver                 M 0x4cbf6ba4 hasDNSRecords(bytes32,bytes32)
PublicResolver                 M 0x124a319c interfaceImplementer(bytes32,bytes4)
PublicResolver                 M 0x8b95dd71 setAddr(bytes32,uint256,bytes)
PublicResolver                 M 0xd5fa2b00 setAddr(bytes32,address)
PublicResolver                 M 0xe59d895d setInterface(bytes32,bytes4,address)
PublicResolver                 M 0x29cd62ea setPubkey(bytes32,bytes32,bytes32)
PublicResolver                 M 0x3603d758 clearRecords(bytes32)
PublicResolver                 M 0xa8fa5682 dnsRecord(bytes32,bytes32,uint16)
PublicResolver                 M 0x77372213 setName(bytes32,string)
PublicResolver                 M 0x10f13a8c setText(bytes32,string,string)
PublicResolver                 M 0x59d1d43c text(bytes32,string)
PublicResolver                 M 0xf1cb7e06 addr(bytes32,uint256)
PublicResolver                 M 0xe985e9c5 isApprovedForAll(address,address)
PublicResolver                 M 0x3b3b57de addr(bytes32)
PublicResolver                 M 0xbc1c58d1 contenthash(bytes32)
PublicResolver                 M 0xe32954eb multicallWithNodeCheck(bytes32,bytes[])
PublicResolver                 M 0x691f3431 name(bytes32)
PublicResolver                 M 0xc8690233 pubkey(bytes32)
PublicResolver                 M 0x304e6ade setContenthash(bytes32,bytes)
PublicResolver                 M 0x0af179d7 setDNSRecords(bytes32,bytes)
PublicResolver                 M 0x5c98042b zonehash(bytes32)
PublicResolver                 M 0x2203ab56 ABI(bytes32,uint256)
PublicResolver                 M 0xac9650d8 multicall(bytes[])
PublicResolver                 M 0xa22cb465 setApprovalForAll(address,bool)
PublicResolver                 M 0x01ffc9a7 supportsInterface(bytes4)
PublicResolver                 M 0xa9784b3e isApprovedFor(address,bytes32,address)
PublicResolver                 M 0xd700ff33 recordVersions(bytes32)
PublicResolver                 M 0x623195b0 setABI(bytes32,uint256,bytes)
PublicResolver                 M 0xce3decdc setZonehash(bytes32,bytes)
OldEnsRegistrarController      E 0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f NameRegistered(string,bytes32,address,uint256,uint256)
OldEnsRegistrarController      E 0x3da24c024582931cfaf8267d8ed24d13a82a8068d5bd337d30ec45cea4e506ae NameRenewed(string,bytes32,uint256,uint256)
OldEnsRegistrarController      E 0xf261845a790fe29bbd6631e2ca4a5bdc83e6eed7c3271d9590d97287e00e9123 NewPriceOracle(address)
OldEnsRegistrarController      E 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0 OwnershipTransferred(address,address)
OldEnsRegistrarController      M 0x01ffc9a7 supportsInterface(bytes4)
OldEnsRegistrarController      M 0x3ccfd60b withdraw()
OldEnsRegistrarController      M 0x8a95b09f MIN_REGISTRATION_DURATION()
OldEnsRegistrarController      M 0xaeb8ce9b available(string)
OldEnsRegistrarController      M 0x8d839ffe minCommitmentAge()
OldEnsRegistrarController      M 0xacf1a841 renew(string,uint256)
OldEnsRegistrarController      M 0xf2fde38b transferOwnership(address)
OldEnsRegistrarController      M 0x9791c097 valid(string)
OldEnsRegistrarController      M 0x8f32d59b isOwner()
OldEnsRegistrarController      M 0x8da5cb5b owner()
OldEnsRegistrarController      M 0xf7a16963 registerWithConfig(string,address,uint256,bytes32,address,address)
OldEnsRegistrarController      M 0x715018a6 renounceOwnership()
OldEnsRegistrarController      M 0x83e7f6ff rentPrice(string,uint256)
OldEnsRegistrarController      M 0x7e324479 setCommitmentAges(uint256,uint256)
OldEnsRegistrarController      M 0x530e784f setPriceOracle(address)
OldEnsRegistrarController      M 0xf14fcbc8 commit(bytes32)
OldEnsRegistrarController      M 0xf49826be makeCommitment(string,address,bytes32)
OldEnsRegistrarController      M 0x3d86c52f makeCommitmentWithConfig(string,address,bytes32,address,address)
OldEnsRegistrarController      M 0x839df945 commitments(bytes32)
OldEnsRegistrarController      M 0xce1e09c0 maxCommitmentAge()
OldEnsRegistrarController      M 0x85f6d155 register(string,address,uint256,bytes32)

*/

var topicsMap = map[string]string{
	// old-registrar
	"0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f": "NameRegistered(string,bytes32,address,uint256,uint256)",

	// BaseRegistrarImplementation
	"0x69e37f151eb98a09618ddaa80c8cfaf1ce5996867c489f45b555b412271ebf27": "NameRegistered(uint256,address,uint256)",
	"0x3da24c024582931cfaf8267d8ed24d13a82a8068d5bd337d30ec45cea4e506ae": "NameRenewed",
	"0x65412581168e88a1e60c6459d7f44ae83ad0832e670826c05a4e2476b57af752": "AddressChanged",
	"0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7": "NameChanged",

	// ETHRegistrarController

	// ENSRegistry
	"0xce0457fe73731f824cc272376169235128c118b49d344817417c6d108d155e82": "NewOwner(bytes32,bytes32,address)",
	"0x335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a0": "NewResolver(bytes32,address)",
	"0x1d4f9bbfc9cab89d66e1a1562f2233ccbf1308cb4f63de2ead5787adddb8fa68": "NewTTL(bytes32,uint64)",
	"0xd4735d920b0f87494915f556dd9b54c8f309026070caea5c737245152564d266": "Transfer(bytes32,address)",
}

var ensCrontractAddresses = map[string]string{
	"0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e": "Registry", // <-
	"0x57f1887a8BF19b14fC0dF6Fd9B2acc9Af147eA85": "BaseRegistrar",
	"0x253553366Da8546fC250F225fe3d25d0C782303b": "ETHRegistrarController", // <-
	"0xB32cB5677a7C971689228EC835800432B339bA2B": "DNSRegistrar",
	"0xa58E81fe9b61B5c3fE2AFD33CF304c454AbFc7Cb": "ReverseRegistrar",
	"0xD4416b13d2b3a9aBae7AcD5D6C2BbDBE25686401": "NameWrapper",
	"0x231b0Ee14048e9dCcD1d247744d114a4EB5E8E63": "PublicResolver",
	"0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5": "OldEnsRegistrarController", // <-
}

func (bigtable *Bigtable) TransformEnsNameRegistered(blk *types.Eth1Block, cache *freecache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error) {
	bulkData = &types.BulkMutations{}
	bulkMetadataUpdates = &types.BulkMutations{}

	filterer, err := ens.NewEnsRegistrarFilterer(common.Address{}, nil)
	if err != nil {
		log.Printf("error creating filterer: %v", err)
		return nil, nil, err
	}
	keys := make(map[string]bool)

	for i, tx := range blk.GetTransactions() {
		if i >= TX_PER_BLOCK_LIMIT {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, tx: %x", TX_PER_BLOCK_LIMIT-1, i, tx.GetHash())
		}

		foundNameIndex := -1
		foundResolverIndex := -1
		foundNameRenewedIndex := -1
		foundAddressChangedIndices := []int{}
		foundNameChangedIndex := -1
		foundNewOwnerIndex := -1
		logs := tx.GetLogs()

		for j, log := range logs {
			if j >= ITX_PER_TX_LIMIT {
				return nil, nil, fmt.Errorf("unexpected number of logs in block expected at most %d but got: %v tx: %x", ITX_PER_TX_LIMIT-1, j, tx.GetHash())
			}
			// isRegistarContract := utils.SliceContains(utils.Config.Indexer.EnsTransformer.ValidRegistrarContracts, common.BytesToAddress(log.Address).String())
			ensContract := ensCrontractAddresses[common.BytesToAddress(log.Address).String()]
			for k, lTopic := range log.GetTopics() {
				if ensContract == "Registry" {
					// 0x335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a0 NewResolver
					// 0xce0457fe73731f824cc272376169235128c118b49d344817417c6d108d155e82 NewOwner
				} else if ensContract == "BaseRegistrar" {
					// 0xb3d987963d01b2f68493b4bdb130988f157ea43070d4ad840fee0466ed9370d9 NameRegistered
					// 0x9b87a00e30f1ac65d898f070f8a3488fe60517182d0a2098e1b4b93a54aa9bd6 NameRenewed
					// 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0 OwnershipTransferred
					// 0xea3d7e1195a15d2ddcd859b01abd4c6b960fa9f9264e499a70a90c7f0c64b717 NameMigrated
				} else if ensContract == "OldEnsRegistrarController" {
					// 0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f NameRegistered
					// 0x3da24c024582931cfaf8267d8ed24d13a82a8068d5bd337d30ec45cea4e506ae NameRenewed
					// 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0 OwnershipTransferred
				} else {
					// 0x65412581168e88a1e60c6459d7f44ae83ad0832e670826c05a4e2476b57af752 AddressChanged
					// 0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7 NameChanged
				}

				if isRegistarContract {

					if bytes.Equal(lTopic, ens.NameRegisteredTopic) || bytes.Equal(lTopic, ens.NameRegisteredV2Topic) {
						foundNameIndex = j
					} else if bytes.Equal(lTopic, ens.NewResolverTopic) {
						foundResolverIndex = j
					} else if bytes.Equal(lTopic, ens.NameRenewedTopic) {
						foundNameRenewedIndex = j
					}
				} else if bytes.Equal(lTopic, ens.AddressChangedTopic) {
					foundAddressChangedIndices = append(foundAddressChangedIndices, j)
				} else if bytes.Equal(lTopic, ens.NameChangedTopic) {
					foundNameChangedIndex = j
				}

				if false && topicsMap[fmt.Sprintf("%#x", lTopic)] != "" {
					fmt.Printf("DEBUG: tx:%#x %v: %v: %#x: %v\n", tx.Hash, j, k, lTopic, topicsMap[fmt.Sprintf("%#x", lTopic)])
				}
			}
		}

		if true {
			switch {
			case fmt.Sprintf("%#x", tx.Hash) == "0x6cb6cd3b5ceb992a6110bf2de0508989e12ac03876846bdbb3184ff4915b39c5":
				fmt.Printf("DEBUG: itxs: %#x %v\n", tx.Hash, len(tx.Itx))
				for i, itx := range tx.Itx {
					fmt.Printf("DEBUG: itx: %v: %#x -> %#x %s %v\n", i, itx.From, itx.To, common.BytesToAddress(itx.To).String(), utils.SliceContains(utils.Config.Indexer.EnsTransformer.ValidRegistrarContracts, common.BytesToAddress(itx.To).String()))
				}
			case fmt.Sprintf("%#x", tx.Hash) == "0x30d164ba6a9f8b45229d98f2c324fdc38b958d24d424ab6b2ea064a9045754d6":
				fmt.Printf("DEBUG: itxs: %#x %v\n", tx.Hash, len(tx.Itx))
				for i, itx := range tx.Itx {
					fmt.Printf("DEBUG: itx: %v: %#x -> %#x %s %v\n", i, itx.From, itx.To, common.BytesToAddress(itx.To).String(), utils.SliceContains(utils.Config.Indexer.EnsTransformer.ValidRegistrarContracts, common.BytesToAddress(itx.To).String()))
				}
			default:
			}
		}

		// We found a register name event
		if foundNameIndex > -1 && foundResolverIndex > -1 {

			log := logs[foundNameIndex]
			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			nameLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundNameIndex),
				Removed:     log.GetRemoved(),
			}

			log = logs[foundResolverIndex]
			topics = make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			resolverLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundResolverIndex),
				Removed:     log.GetRemoved(),
			}

			var owner common.Address
			var name string

			nameRegistered, err := filterer.ParseNameRegistered(nameLog)
			if err != nil {
				nameRegisteredV2, err := filterer.ParseNameRegisteredV2(nameLog)
				if err != nil {
					utils.LogError(err, fmt.Sprintf("indexing of register event failed parse register event at tx [%v] index [%v] on block [%v]", i, foundNameIndex, blk.Number), 0)
					continue
				}
				owner = nameRegisteredV2.Owner
				name = nameRegisteredV2.Name
			} else {
				owner = nameRegistered.Owner
				name = nameRegistered.Name
			}

			if err = verifyName(name); err != nil {
				logger.Warnf("indexing of register event failed because of invalid name at tx [%v] index [%v] on block [%v]: %v", i, foundNameIndex, blk.Number, err)
				continue
			}

			resolver, err := filterer.ParseNewResolver(resolverLog)
			if err != nil {
				utils.LogError(err, fmt.Sprintf("indexing of register event failed parse resolver event at tx [%v] index [%v] on block [%v]", i, foundNameIndex, blk.Number), 0)
				continue
			}

			keys[fmt.Sprintf("%s:ENS:I:H:%x:%x", bigtable.chainId, resolver.Node, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:I:A:%x:%x", bigtable.chainId, owner, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:V:A:%x", bigtable.chainId, owner)] = true
			keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, name)] = true

		} else if foundNameRenewedIndex > -1 { // We found a renew name event
			log := logs[foundNameRenewedIndex]
			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			nameRenewedLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundNameRenewedIndex),
				Removed:     log.GetRemoved(),
			}

			nameRenewed, err := filterer.ParseNameRenewed(nameRenewedLog)
			if err != nil {
				utils.LogError(err, fmt.Sprintf("indexing of renew event failed parse event at tx [%v] index [%v] on block [%v]", i, foundNameRenewedIndex, blk.Number), 0)
				continue
			}

			if err = verifyName(nameRenewed.Name); err != nil {
				logger.Warnf("indexing of renew event failed because of invalid name at tx [%v] index [%v] on block [%v]: %v", i, foundNameIndex, blk.Number, err)
				continue
			}

			nameHash, err := go_ens.NameHash(nameRenewed.Name)
			if err != nil {
				utils.LogError(err, fmt.Sprintf("error hashing ens name [%v] at tx [%v] index [%v] on block [%v]", nameRenewed.Name, i, foundNameRenewedIndex, blk.Number), 0)
				continue
			}
			keys[fmt.Sprintf("%s:ENS:I:H:%x:%x", bigtable.chainId, nameHash, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, nameRenewed.Name)] = true

		} else if foundNameChangedIndex > -1 && foundNewOwnerIndex > -1 { // we found a name change event
			log := logs[foundNewOwnerIndex]
			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}
			newOwnerLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundNewOwnerIndex),
				Removed:     log.GetRemoved(),
			}

			newOwner, err := filterer.ParseNewOwner(newOwnerLog)
			if err != nil {
				utils.LogError(err, fmt.Errorf("indexing of new owner event failed parse event at index %v on block [%v]", foundNewOwnerIndex, blk.Number), 0)
				continue
			}

			nameChangedLog := logs[foundNameChangedIndex]
			nameChangedTopics := make([]common.Hash, 0, len(nameChangedLog.GetTopics()))
			for _, t := range nameChangedLog.GetTopics() {
				nameChangedTopics = append(nameChangedTopics, common.BytesToHash(t))
			}
			nameChangedLogT := eth_types.Log{
				Address:     common.BytesToAddress(nameChangedLog.GetAddress()),
				Data:        nameChangedLog.Data,
				Topics:      nameChangedTopics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundNameChangedIndex),
				Removed:     nameChangedLog.GetRemoved(),
			}
			newName, err := filterer.ParseNameChanged(nameChangedLogT)
			if err != nil {
				utils.LogError(err, fmt.Errorf("indexing of NameChanged event failed parse event at index %v on block [%v]", foundNameChangedIndex, blk.Number), 0)
				continue
			}

			keys[fmt.Sprintf("%s:ENS:I:A:%x:%x", bigtable.chainId, newOwner.Owner, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:V:A:%x", bigtable.chainId, newOwner.Owner)] = true
			keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, newName.Name)] = true
		}
		// We found a change address event, there can be multiple within one transaction
		for _, addressChangeIndex := range foundAddressChangedIndices {

			log := logs[addressChangeIndex]
			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			addressChangedLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(addressChangeIndex),
				Removed:     log.GetRemoved(),
			}

			addressChanged, err := filterer.ParseAddressChanged(addressChangedLog)
			if err != nil {
				utils.LogError(err, fmt.Sprintf("indexing of address change event failed parse event at index [%v] on block [%v]", addressChangeIndex, blk.Number), 0)
				continue
			}

			keys[fmt.Sprintf("%s:ENS:I:H:%x:%x", bigtable.chainId, addressChanged.Node, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:V:H:%x", bigtable.chainId, addressChanged.Node)] = true

		}
	}
	for key := range keys {
		mut := gcp_bigtable.NewMutation()
		mut.Set(DEFAULT_FAMILY, key, gcp_bigtable.Timestamp(0), nil)

		bulkData.Keys = append(bulkData.Keys, key)
		bulkData.Muts = append(bulkData.Muts, mut)
	}

	return bulkData, bulkMetadataUpdates, nil
}

func verifyName(name string) error {
	// limited by max capacity of db (caused by btrees of indexes); tests showed maximum of 2684 (added buffer)
	if len(name) > 2048 {
		return fmt.Errorf("name too long: %v", name)
	}
	return nil
}

type EnsCheckedDictionary struct {
	mux     sync.Mutex
	address map[common.Address]bool
	name    map[string]bool
}

func (bigtable *Bigtable) GetRowsByPrefix(prefix string) ([]string, error) {
	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	rowRange := gcp_bigtable.PrefixRange(prefix)
	keys := []string{}

	err := bigtable.tableData.ReadRows(ctx, rowRange, func(row gcp_bigtable.Row) bool {
		row_ := row[DEFAULT_FAMILY][0]
		keys = append(keys, row_.Row)
		return true
	}, gcp_bigtable.LimitRows(1000))
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (bigtable *Bigtable) ImportEnsUpdates(client *ethclient.Client, readBatchSize int64) error {
	key := fmt.Sprintf("%s:ENS:V", bigtable.chainId)

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	rowRange := gcp_bigtable.PrefixRange(key)
	keys := []string{}

	err := bigtable.tableData.ReadRows(ctx, rowRange, func(row gcp_bigtable.Row) bool {
		row_ := row[DEFAULT_FAMILY][0]
		keys = append(keys, row_.Row)
		return true
	}, gcp_bigtable.LimitRows(readBatchSize)) // limit to max 1000 entries to avoid blocking the import of new blocks
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		logger.Info("No ENS entries to validate")
		return nil
	}

	logger.Infof("Validating %v ENS entries", len(keys))
	alreadyChecked := EnsCheckedDictionary{
		address: make(map[common.Address]bool),
		name:    make(map[string]bool),
	}

	mutDelete := gcp_bigtable.NewMutation()
	mutDelete.DeleteRow()

	batchSize := 100
	total := len(keys)
	for i := 0; i < total; i += batchSize {
		to := i + batchSize
		if to > total {
			to = total
		}
		batch := keys[i:to]
		logger.Infof("Batching ENS entries %v:%v of %v", i, to, total)

		g := new(errgroup.Group)
		g.SetLimit(10) // limit load on the node
		mutsDelete := &types.BulkMutations{
			Keys: make([]string, 0, 1),
			Muts: make([]*gcp_bigtable.Mutation, 0, 1),
		}

		for _, k := range batch {
			key := k
			var name string
			var address *common.Address
			split := strings.Split(key, ":")
			value := split[4]

			switch split[3] {
			case "H":
				// if we have a hash we look if we find a name in the db. If not we can ignore it.
				nameHash, err := hex.DecodeString(value)
				if err != nil {
					utils.LogError(err, fmt.Errorf("name hash could not be decoded: %v", value), 0)
				} else {
					err := ReaderDb.Get(&name, `
					SELECT
						ens_name
					FROM ens
					WHERE name_hash = $1
					`, nameHash[:])
					if err != nil && err != sql.ErrNoRows {
						return err
					}
				}
			case "A":
				addressHash, err := hex.DecodeString(value)
				if err != nil {
					utils.LogError(err, fmt.Errorf("address hash could not be decoded: %v", value), 0)
				} else {
					add := common.BytesToAddress(addressHash)
					address = &add
				}
			case "N":
				name = value
			}

			g.Go(func() error {
				if name != "" {
					err := validateEnsName(client, name, &alreadyChecked, nil)
					if err != nil {
						return fmt.Errorf("error validating new name [%v]: %w", name, err)
					}
				} else if address != nil {
					err := validateEnsAddress(client, *address, &alreadyChecked)
					if err != nil {
						return fmt.Errorf("error validating new address [%v]: %w", address, err)
					}
				}
				return nil
			})

			mutsDelete.Keys = append(mutsDelete.Keys, key)
			mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
		}

		if err := g.Wait(); err != nil {
			return err
		}

		// After processing a batch of keys we remove them from bigtable
		err = bigtable.WriteBulk(mutsDelete, bigtable.tableData, DEFAULT_BATCH_INSERTS)
		if err != nil {
			return err
		}

		// give node some time for other stuff between batches
		time.Sleep(time.Millisecond * 100)
	}

	logger.Info("Import of ENS updates completed")
	return nil
}

func validateEnsAddress(client *ethclient.Client, address common.Address, alreadyChecked *EnsCheckedDictionary) error {
	alreadyChecked.mux.Lock()
	if alreadyChecked.address[address] {
		alreadyChecked.mux.Unlock()
		return nil
	}
	alreadyChecked.address[address] = true
	alreadyChecked.mux.Unlock()

	name, err := go_ens.ReverseResolve(client, address)
	if err != nil {
		if err.Error() == "not a resolver" || err.Error() == "no resolution" {
			logger.Warnf("reverse resolving address [%v] resulted in a skippable error [%s], skipping it", address, err.Error())
			return nil
		}

		return fmt.Errorf("error could not reverse resolve address [%v]: %w", address, err)
	}

	currentName, err := GetEnsNameForAddress(address)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	isPrimary := false
	if currentName != nil {
		if *currentName == name {
			return nil
		}
		logger.Infof("Address [%x] has a new main name from %x to: %v", address, *currentName, name)
		err := validateEnsName(client, *currentName, alreadyChecked, &isPrimary)
		if err != nil {
			return fmt.Errorf("error validating new name [%v]: %w", *currentName, err)
		}
	}
	isPrimary = true
	logger.Infof("Address [%x] has a primary name: %v", address, name)
	return validateEnsName(client, name, alreadyChecked, &isPrimary)
}

func validateEnsName(client *ethclient.Client, name string, alreadyChecked *EnsCheckedDictionary, isPrimaryName *bool) error {
	// For now only .eth is supported other ens domains use different techniques and require and individual implementation
	if !strings.HasSuffix(name, ".eth") {
		name = fmt.Sprintf("%s.eth", name)
	}
	alreadyChecked.mux.Lock()
	if alreadyChecked.name[name] {
		alreadyChecked.mux.Unlock()
		return nil
	}
	alreadyChecked.name[name] = true
	alreadyChecked.mux.Unlock()

	nameHash, err := go_ens.NameHash(name)
	if err != nil {
		logger.Errorf("error could not hash name [%v]: %v -> removing ens entry", name, err)

		err = removeEnsName(client, name)
		if err != nil {
			return fmt.Errorf("error removing ens name [%v]: %w", name, err)
		}
		return nil

		//return fmt.Errorf("error could not hash name [%v]: %w", name, err)
	}

	addr, err := go_ens.Resolve(client, name)
	if err != nil {
		if err.Error() == "unregistered name" ||
			err.Error() == "no address" ||
			err.Error() == "no resolver" ||
			err.Error() == "abi: attempting to unmarshall an empty string while arguments are expected" ||
			strings.Contains(err.Error(), "execution reverted") ||
			err.Error() == "invalid jump destination" ||
			err.Error() == "invalid opcode: INVALID" {
			// the given name is not available anymore or resolving it did not work properly => we can remove it from the db (if it is there)
			logger.WithField("error", err).WithField("name", name).Warnf("could not resolve name")
			err = removeEnsName(client, name)
			if err != nil {
				return fmt.Errorf("error removing ens name after resolve failed [%v]: %w", name, err)
			}
			return nil
		}
		return fmt.Errorf("error could not resolve name [%v]: %w", name, err)
	}

	// we need to get the main domain to get the expiration date
	parts := strings.Split(name, ".")
	mainName := strings.Join(parts[len(parts)-2:], ".")
	ensName, err := go_ens.NewName(client, mainName)
	if err != nil {
		return fmt.Errorf("error could not create name via go_ens.NewName for [%v]: %w", name, err)
	}

	expires, err := ensName.Expires()
	if err != nil {
		return fmt.Errorf("error could not get ens expire date for [%v]: %w", name, err)
	}
	isPrimary := false
	if isPrimaryName == nil {
		reverseName, err := go_ens.ReverseResolve(client, addr)
		if err != nil {
			if err.Error() == "not a resolver" || err.Error() == "no resolution" {
				logger.Warnf("reverse resolving address [%v] for name [%v] resulted in an error [%s], marking entry as not primary", addr, name, err.Error())
			} else {
				return fmt.Errorf("error could not reverse resolve address [%v]: %w", addr, err)
			}
		}
		if reverseName == name {
			isPrimary = true
		}
	} else if *isPrimaryName {
		isPrimary = true
	}
	_, err = WriterDb.Exec(`
	INSERT INTO ens (
		name_hash, 
		ens_name, 
		address,
		is_primary_name, 
		valid_to)
	VALUES ($1, $2, $3, $4, $5) 
	ON CONFLICT 
		(name_hash) 
	DO UPDATE SET 
		ens_name = excluded.ens_name,
		address = excluded.address,
		is_primary_name = excluded.is_primary_name,
		valid_to = excluded.valid_to
	`, nameHash[:], name, addr.Bytes(), isPrimary, expires)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "invalid byte sequence") {
			logger.Warnf("could not insert ens name [%v]: %v", name, err)
			return nil
		}
		return fmt.Errorf("error writing ens data for name [%v]: %w", name, err)
	}

	logger.Infof("Name [%v] resolved -> %x, expires: %v, is primary: %v", name, addr, expires, isPrimary)
	return nil
}

func GetAddressForEnsName(name string) (address *common.Address, err error) {
	addressBytes := []byte{}
	err = ReaderDb.Get(&addressBytes, `
	SELECT address 
	FROM ens
	WHERE
		ens_name = $1 AND
		valid_to >= now()
	`, name)
	if err == nil && addressBytes != nil {
		add := common.BytesToAddress(addressBytes)
		address = &add
	}
	return address, err
}

func GetEnsNameForAddress(address common.Address) (name *string, err error) {
	err = ReaderDb.Get(&name, `
	SELECT ens_name 
	FROM ens
	WHERE
		address = $1 AND
		is_primary_name AND
		valid_to >= now()
	;`, address.Bytes())
	return name, err
}

func GetEnsNamesForAddress(addressMap map[string]string) error {
	if len(addressMap) == 0 {
		return nil
	}
	type pair struct {
		Address []byte `db:"address"`
		EnsName string `db:"ens_name"`
	}
	dbAddresses := []pair{}
	addresses := make([][]byte, 0, len(addressMap))
	for add := range addressMap {
		addresses = append(addresses, []byte(add))
	}

	err := ReaderDb.Select(&dbAddresses, `
	SELECT address, ens_name 
	FROM ens
	WHERE
		address = ANY($1) AND
		is_primary_name AND
		valid_to >= now()
	;`, addresses)
	if err != nil {
		return err
	}
	for _, foundling := range dbAddresses {
		addressMap[string(foundling.Address)] = foundling.EnsName
	}
	return nil
}

func removeEnsName(client *ethclient.Client, name string) error {
	_, err := WriterDb.Exec(`
	DELETE FROM ens 
	WHERE 
		ens_name = $1
	;`, name)
	if err != nil && strings.Contains(fmt.Sprintf("%v", err), "invalid byte sequence") {
		logger.Warnf("could not delete ens name [%v]: %v", name, err)
		return nil
	} else if err != nil {
		return fmt.Errorf("error deleting ens name [%v]: %v", name, err)
	}
	logger.Infof("Ens name removed from db: %v", name)
	return nil
}

func (bigtable *Bigtable) TransformEnsNameRegistered2(blk *types.Eth1Block, cache *freecache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error) {
	registrarFilterer, _ := registry.NewContractFilterer(common.Address{}, nil)
	_ = registrarFilterer
	return nil, nil, nil
}
