// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ens

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ENSNameWrapperMetaData contains all meta data concerning the ENSNameWrapper contract.
var ENSNameWrapperMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractENS\",\"name\":\"_ens\",\"type\":\"address\"},{\"internalType\":\"contractIBaseRegistrar\",\"name\":\"_registrar\",\"type\":\"address\"},{\"internalType\":\"contractIMetadataService\",\"name\":\"_metadataService\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"CannotUpgrade\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"IncompatibleParent\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"IncorrectTargetOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"IncorrectTokenType\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"labelHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"expectedLabelhash\",\"type\":\"bytes32\"}],\"name\":\"LabelMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"}],\"name\":\"LabelTooLong\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LabelTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NameIsNotWrapped\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"OperationProhibited\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"Unauthorised\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"}],\"name\":\"ControllerChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"ExpiryExtended\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"}],\"name\":\"FusesSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"NameUnwrapped\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"NameWrapped\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"name\":\"TransferBatch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TransferSingle\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"URI\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"_tokens\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"fuseMask\",\"type\":\"uint32\"}],\"name\":\"allFusesBurned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"accounts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"}],\"name\":\"balanceOfBatch\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"canExtendSubnames\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"canModifyName\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"controllers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ens\",\"outputs\":[{\"internalType\":\"contractENS\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"extendExpiry\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"getData\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"}],\"name\":\"isWrapped\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"isWrapped\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"metadataService\",\"outputs\":[{\"internalType\":\"contractIMetadataService\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"names\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"onERC721Received\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"recoverFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"wrappedOwner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"ownerControlledFuses\",\"type\":\"uint16\"}],\"name\":\"registerAndWrapETH2LD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"registrarExpiry\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registrar\",\"outputs\":[{\"internalType\":\"contractIBaseRegistrar\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"renew\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeBatchTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"setChildFuses\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"}],\"name\":\"setController\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint16\",\"name\":\"ownerControlledFuses\",\"type\":\"uint16\"}],\"name\":\"setFuses\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIMetadataService\",\"name\":\"_metadataService\",\"type\":\"address\"}],\"name\":\"setMetadataService\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"}],\"name\":\"setRecord\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"setResolver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"setSubnodeOwner\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"},{\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"setSubnodeRecord\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"}],\"name\":\"setTTL\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractINameWrapperUpgrade\",\"name\":\"_upgradeAddress\",\"type\":\"address\"}],\"name\":\"setUpgradeContract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"unwrap\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"registrant\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"unwrapETH2LD\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"upgrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"upgradeContract\",\"outputs\":[{\"internalType\":\"contractINameWrapperUpgrade\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"uri\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrappedOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"wrap\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"wrappedOwner\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"ownerControlledFuses\",\"type\":\"uint16\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"wrapETH2LD\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ENSNameWrapperABI is the input ABI used to generate the binding from.
// Deprecated: Use ENSNameWrapperMetaData.ABI instead.
var ENSNameWrapperABI = ENSNameWrapperMetaData.ABI

// ENSNameWrapper is an auto generated Go binding around an Ethereum contract.
type ENSNameWrapper struct {
	ENSNameWrapperCaller     // Read-only binding to the contract
	ENSNameWrapperTransactor // Write-only binding to the contract
	ENSNameWrapperFilterer   // Log filterer for contract events
}

// ENSNameWrapperCaller is an auto generated read-only Go binding around an Ethereum contract.
type ENSNameWrapperCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSNameWrapperTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ENSNameWrapperTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSNameWrapperFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ENSNameWrapperFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSNameWrapperSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ENSNameWrapperSession struct {
	Contract     *ENSNameWrapper   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ENSNameWrapperCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ENSNameWrapperCallerSession struct {
	Contract *ENSNameWrapperCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// ENSNameWrapperTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ENSNameWrapperTransactorSession struct {
	Contract     *ENSNameWrapperTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// ENSNameWrapperRaw is an auto generated low-level Go binding around an Ethereum contract.
type ENSNameWrapperRaw struct {
	Contract *ENSNameWrapper // Generic contract binding to access the raw methods on
}

// ENSNameWrapperCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ENSNameWrapperCallerRaw struct {
	Contract *ENSNameWrapperCaller // Generic read-only contract binding to access the raw methods on
}

// ENSNameWrapperTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ENSNameWrapperTransactorRaw struct {
	Contract *ENSNameWrapperTransactor // Generic write-only contract binding to access the raw methods on
}

// NewENSNameWrapper creates a new instance of ENSNameWrapper, bound to a specific deployed contract.
func NewENSNameWrapper(address common.Address, backend bind.ContractBackend) (*ENSNameWrapper, error) {
	contract, err := bindENSNameWrapper(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapper{ENSNameWrapperCaller: ENSNameWrapperCaller{contract: contract}, ENSNameWrapperTransactor: ENSNameWrapperTransactor{contract: contract}, ENSNameWrapperFilterer: ENSNameWrapperFilterer{contract: contract}}, nil
}

// NewENSNameWrapperCaller creates a new read-only instance of ENSNameWrapper, bound to a specific deployed contract.
func NewENSNameWrapperCaller(address common.Address, caller bind.ContractCaller) (*ENSNameWrapperCaller, error) {
	contract, err := bindENSNameWrapper(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperCaller{contract: contract}, nil
}

// NewENSNameWrapperTransactor creates a new write-only instance of ENSNameWrapper, bound to a specific deployed contract.
func NewENSNameWrapperTransactor(address common.Address, transactor bind.ContractTransactor) (*ENSNameWrapperTransactor, error) {
	contract, err := bindENSNameWrapper(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperTransactor{contract: contract}, nil
}

// NewENSNameWrapperFilterer creates a new log filterer instance of ENSNameWrapper, bound to a specific deployed contract.
func NewENSNameWrapperFilterer(address common.Address, filterer bind.ContractFilterer) (*ENSNameWrapperFilterer, error) {
	contract, err := bindENSNameWrapper(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperFilterer{contract: contract}, nil
}

// bindENSNameWrapper binds a generic wrapper to an already deployed contract.
func bindENSNameWrapper(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ENSNameWrapperMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ENSNameWrapper *ENSNameWrapperRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ENSNameWrapper.Contract.ENSNameWrapperCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ENSNameWrapper *ENSNameWrapperRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.ENSNameWrapperTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ENSNameWrapper *ENSNameWrapperRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.ENSNameWrapperTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ENSNameWrapper *ENSNameWrapperCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ENSNameWrapper.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ENSNameWrapper *ENSNameWrapperTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ENSNameWrapper *ENSNameWrapperTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.contract.Transact(opts, method, params...)
}

// Tokens is a free data retrieval call binding the contract method 0xed70554d.
//
// Solidity: function _tokens(uint256 ) view returns(uint256)
func (_ENSNameWrapper *ENSNameWrapperCaller) Tokens(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "_tokens", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Tokens is a free data retrieval call binding the contract method 0xed70554d.
//
// Solidity: function _tokens(uint256 ) view returns(uint256)
func (_ENSNameWrapper *ENSNameWrapperSession) Tokens(arg0 *big.Int) (*big.Int, error) {
	return _ENSNameWrapper.Contract.Tokens(&_ENSNameWrapper.CallOpts, arg0)
}

// Tokens is a free data retrieval call binding the contract method 0xed70554d.
//
// Solidity: function _tokens(uint256 ) view returns(uint256)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) Tokens(arg0 *big.Int) (*big.Int, error) {
	return _ENSNameWrapper.Contract.Tokens(&_ENSNameWrapper.CallOpts, arg0)
}

// AllFusesBurned is a free data retrieval call binding the contract method 0xadf4960a.
//
// Solidity: function allFusesBurned(bytes32 node, uint32 fuseMask) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCaller) AllFusesBurned(opts *bind.CallOpts, node [32]byte, fuseMask uint32) (bool, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "allFusesBurned", node, fuseMask)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllFusesBurned is a free data retrieval call binding the contract method 0xadf4960a.
//
// Solidity: function allFusesBurned(bytes32 node, uint32 fuseMask) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperSession) AllFusesBurned(node [32]byte, fuseMask uint32) (bool, error) {
	return _ENSNameWrapper.Contract.AllFusesBurned(&_ENSNameWrapper.CallOpts, node, fuseMask)
}

// AllFusesBurned is a free data retrieval call binding the contract method 0xadf4960a.
//
// Solidity: function allFusesBurned(bytes32 node, uint32 fuseMask) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) AllFusesBurned(node [32]byte, fuseMask uint32) (bool, error) {
	return _ENSNameWrapper.Contract.AllFusesBurned(&_ENSNameWrapper.CallOpts, node, fuseMask)
}

// BalanceOf is a free data retrieval call binding the contract method 0x00fdd58e.
//
// Solidity: function balanceOf(address account, uint256 id) view returns(uint256)
func (_ENSNameWrapper *ENSNameWrapperCaller) BalanceOf(opts *bind.CallOpts, account common.Address, id *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "balanceOf", account, id)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x00fdd58e.
//
// Solidity: function balanceOf(address account, uint256 id) view returns(uint256)
func (_ENSNameWrapper *ENSNameWrapperSession) BalanceOf(account common.Address, id *big.Int) (*big.Int, error) {
	return _ENSNameWrapper.Contract.BalanceOf(&_ENSNameWrapper.CallOpts, account, id)
}

// BalanceOf is a free data retrieval call binding the contract method 0x00fdd58e.
//
// Solidity: function balanceOf(address account, uint256 id) view returns(uint256)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) BalanceOf(account common.Address, id *big.Int) (*big.Int, error) {
	return _ENSNameWrapper.Contract.BalanceOf(&_ENSNameWrapper.CallOpts, account, id)
}

// BalanceOfBatch is a free data retrieval call binding the contract method 0x4e1273f4.
//
// Solidity: function balanceOfBatch(address[] accounts, uint256[] ids) view returns(uint256[])
func (_ENSNameWrapper *ENSNameWrapperCaller) BalanceOfBatch(opts *bind.CallOpts, accounts []common.Address, ids []*big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "balanceOfBatch", accounts, ids)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// BalanceOfBatch is a free data retrieval call binding the contract method 0x4e1273f4.
//
// Solidity: function balanceOfBatch(address[] accounts, uint256[] ids) view returns(uint256[])
func (_ENSNameWrapper *ENSNameWrapperSession) BalanceOfBatch(accounts []common.Address, ids []*big.Int) ([]*big.Int, error) {
	return _ENSNameWrapper.Contract.BalanceOfBatch(&_ENSNameWrapper.CallOpts, accounts, ids)
}

// BalanceOfBatch is a free data retrieval call binding the contract method 0x4e1273f4.
//
// Solidity: function balanceOfBatch(address[] accounts, uint256[] ids) view returns(uint256[])
func (_ENSNameWrapper *ENSNameWrapperCallerSession) BalanceOfBatch(accounts []common.Address, ids []*big.Int) ([]*big.Int, error) {
	return _ENSNameWrapper.Contract.BalanceOfBatch(&_ENSNameWrapper.CallOpts, accounts, ids)
}

// CanExtendSubnames is a free data retrieval call binding the contract method 0x0e4cd725.
//
// Solidity: function canExtendSubnames(bytes32 node, address addr) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCaller) CanExtendSubnames(opts *bind.CallOpts, node [32]byte, addr common.Address) (bool, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "canExtendSubnames", node, addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanExtendSubnames is a free data retrieval call binding the contract method 0x0e4cd725.
//
// Solidity: function canExtendSubnames(bytes32 node, address addr) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperSession) CanExtendSubnames(node [32]byte, addr common.Address) (bool, error) {
	return _ENSNameWrapper.Contract.CanExtendSubnames(&_ENSNameWrapper.CallOpts, node, addr)
}

// CanExtendSubnames is a free data retrieval call binding the contract method 0x0e4cd725.
//
// Solidity: function canExtendSubnames(bytes32 node, address addr) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) CanExtendSubnames(node [32]byte, addr common.Address) (bool, error) {
	return _ENSNameWrapper.Contract.CanExtendSubnames(&_ENSNameWrapper.CallOpts, node, addr)
}

// CanModifyName is a free data retrieval call binding the contract method 0x41415eab.
//
// Solidity: function canModifyName(bytes32 node, address addr) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCaller) CanModifyName(opts *bind.CallOpts, node [32]byte, addr common.Address) (bool, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "canModifyName", node, addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanModifyName is a free data retrieval call binding the contract method 0x41415eab.
//
// Solidity: function canModifyName(bytes32 node, address addr) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperSession) CanModifyName(node [32]byte, addr common.Address) (bool, error) {
	return _ENSNameWrapper.Contract.CanModifyName(&_ENSNameWrapper.CallOpts, node, addr)
}

// CanModifyName is a free data retrieval call binding the contract method 0x41415eab.
//
// Solidity: function canModifyName(bytes32 node, address addr) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) CanModifyName(node [32]byte, addr common.Address) (bool, error) {
	return _ENSNameWrapper.Contract.CanModifyName(&_ENSNameWrapper.CallOpts, node, addr)
}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCaller) Controllers(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "controllers", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperSession) Controllers(arg0 common.Address) (bool, error) {
	return _ENSNameWrapper.Contract.Controllers(&_ENSNameWrapper.CallOpts, arg0)
}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) Controllers(arg0 common.Address) (bool, error) {
	return _ENSNameWrapper.Contract.Controllers(&_ENSNameWrapper.CallOpts, arg0)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCaller) Ens(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "ens")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperSession) Ens() (common.Address, error) {
	return _ENSNameWrapper.Contract.Ens(&_ENSNameWrapper.CallOpts)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) Ens() (common.Address, error) {
	return _ENSNameWrapper.Contract.Ens(&_ENSNameWrapper.CallOpts)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 id) view returns(address operator)
func (_ENSNameWrapper *ENSNameWrapperCaller) GetApproved(opts *bind.CallOpts, id *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "getApproved", id)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 id) view returns(address operator)
func (_ENSNameWrapper *ENSNameWrapperSession) GetApproved(id *big.Int) (common.Address, error) {
	return _ENSNameWrapper.Contract.GetApproved(&_ENSNameWrapper.CallOpts, id)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 id) view returns(address operator)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) GetApproved(id *big.Int) (common.Address, error) {
	return _ENSNameWrapper.Contract.GetApproved(&_ENSNameWrapper.CallOpts, id)
}

// GetData is a free data retrieval call binding the contract method 0x0178fe3f.
//
// Solidity: function getData(uint256 id) view returns(address owner, uint32 fuses, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperCaller) GetData(opts *bind.CallOpts, id *big.Int) (struct {
	Owner  common.Address
	Fuses  uint32
	Expiry uint64
}, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "getData", id)

	outstruct := new(struct {
		Owner  common.Address
		Fuses  uint32
		Expiry uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Owner = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Fuses = *abi.ConvertType(out[1], new(uint32)).(*uint32)
	outstruct.Expiry = *abi.ConvertType(out[2], new(uint64)).(*uint64)

	return *outstruct, err

}

// GetData is a free data retrieval call binding the contract method 0x0178fe3f.
//
// Solidity: function getData(uint256 id) view returns(address owner, uint32 fuses, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperSession) GetData(id *big.Int) (struct {
	Owner  common.Address
	Fuses  uint32
	Expiry uint64
}, error) {
	return _ENSNameWrapper.Contract.GetData(&_ENSNameWrapper.CallOpts, id)
}

// GetData is a free data retrieval call binding the contract method 0x0178fe3f.
//
// Solidity: function getData(uint256 id) view returns(address owner, uint32 fuses, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) GetData(id *big.Int) (struct {
	Owner  common.Address
	Fuses  uint32
	Expiry uint64
}, error) {
	return _ENSNameWrapper.Contract.GetData(&_ENSNameWrapper.CallOpts, id)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address account, address operator) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCaller) IsApprovedForAll(opts *bind.CallOpts, account common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "isApprovedForAll", account, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address account, address operator) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperSession) IsApprovedForAll(account common.Address, operator common.Address) (bool, error) {
	return _ENSNameWrapper.Contract.IsApprovedForAll(&_ENSNameWrapper.CallOpts, account, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address account, address operator) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) IsApprovedForAll(account common.Address, operator common.Address) (bool, error) {
	return _ENSNameWrapper.Contract.IsApprovedForAll(&_ENSNameWrapper.CallOpts, account, operator)
}

// IsWrapped is a free data retrieval call binding the contract method 0xd9a50c12.
//
// Solidity: function isWrapped(bytes32 parentNode, bytes32 labelhash) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCaller) IsWrapped(opts *bind.CallOpts, parentNode [32]byte, labelhash [32]byte) (bool, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "isWrapped", parentNode, labelhash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsWrapped is a free data retrieval call binding the contract method 0xd9a50c12.
//
// Solidity: function isWrapped(bytes32 parentNode, bytes32 labelhash) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperSession) IsWrapped(parentNode [32]byte, labelhash [32]byte) (bool, error) {
	return _ENSNameWrapper.Contract.IsWrapped(&_ENSNameWrapper.CallOpts, parentNode, labelhash)
}

// IsWrapped is a free data retrieval call binding the contract method 0xd9a50c12.
//
// Solidity: function isWrapped(bytes32 parentNode, bytes32 labelhash) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) IsWrapped(parentNode [32]byte, labelhash [32]byte) (bool, error) {
	return _ENSNameWrapper.Contract.IsWrapped(&_ENSNameWrapper.CallOpts, parentNode, labelhash)
}

// IsWrapped0 is a free data retrieval call binding the contract method 0xfd0cd0d9.
//
// Solidity: function isWrapped(bytes32 node) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCaller) IsWrapped0(opts *bind.CallOpts, node [32]byte) (bool, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "isWrapped0", node)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsWrapped0 is a free data retrieval call binding the contract method 0xfd0cd0d9.
//
// Solidity: function isWrapped(bytes32 node) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperSession) IsWrapped0(node [32]byte) (bool, error) {
	return _ENSNameWrapper.Contract.IsWrapped0(&_ENSNameWrapper.CallOpts, node)
}

// IsWrapped0 is a free data retrieval call binding the contract method 0xfd0cd0d9.
//
// Solidity: function isWrapped(bytes32 node) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) IsWrapped0(node [32]byte) (bool, error) {
	return _ENSNameWrapper.Contract.IsWrapped0(&_ENSNameWrapper.CallOpts, node)
}

// MetadataService is a free data retrieval call binding the contract method 0x53095467.
//
// Solidity: function metadataService() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCaller) MetadataService(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "metadataService")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MetadataService is a free data retrieval call binding the contract method 0x53095467.
//
// Solidity: function metadataService() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperSession) MetadataService() (common.Address, error) {
	return _ENSNameWrapper.Contract.MetadataService(&_ENSNameWrapper.CallOpts)
}

// MetadataService is a free data retrieval call binding the contract method 0x53095467.
//
// Solidity: function metadataService() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) MetadataService() (common.Address, error) {
	return _ENSNameWrapper.Contract.MetadataService(&_ENSNameWrapper.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ENSNameWrapper *ENSNameWrapperCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ENSNameWrapper *ENSNameWrapperSession) Name() (string, error) {
	return _ENSNameWrapper.Contract.Name(&_ENSNameWrapper.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) Name() (string, error) {
	return _ENSNameWrapper.Contract.Name(&_ENSNameWrapper.CallOpts)
}

// Names is a free data retrieval call binding the contract method 0x20c38e2b.
//
// Solidity: function names(bytes32 ) view returns(bytes)
func (_ENSNameWrapper *ENSNameWrapperCaller) Names(opts *bind.CallOpts, arg0 [32]byte) ([]byte, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "names", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Names is a free data retrieval call binding the contract method 0x20c38e2b.
//
// Solidity: function names(bytes32 ) view returns(bytes)
func (_ENSNameWrapper *ENSNameWrapperSession) Names(arg0 [32]byte) ([]byte, error) {
	return _ENSNameWrapper.Contract.Names(&_ENSNameWrapper.CallOpts, arg0)
}

// Names is a free data retrieval call binding the contract method 0x20c38e2b.
//
// Solidity: function names(bytes32 ) view returns(bytes)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) Names(arg0 [32]byte) ([]byte, error) {
	return _ENSNameWrapper.Contract.Names(&_ENSNameWrapper.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperSession) Owner() (common.Address, error) {
	return _ENSNameWrapper.Contract.Owner(&_ENSNameWrapper.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) Owner() (common.Address, error) {
	return _ENSNameWrapper.Contract.Owner(&_ENSNameWrapper.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 id) view returns(address owner)
func (_ENSNameWrapper *ENSNameWrapperCaller) OwnerOf(opts *bind.CallOpts, id *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "ownerOf", id)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 id) view returns(address owner)
func (_ENSNameWrapper *ENSNameWrapperSession) OwnerOf(id *big.Int) (common.Address, error) {
	return _ENSNameWrapper.Contract.OwnerOf(&_ENSNameWrapper.CallOpts, id)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 id) view returns(address owner)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) OwnerOf(id *big.Int) (common.Address, error) {
	return _ENSNameWrapper.Contract.OwnerOf(&_ENSNameWrapper.CallOpts, id)
}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCaller) Registrar(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "registrar")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperSession) Registrar() (common.Address, error) {
	return _ENSNameWrapper.Contract.Registrar(&_ENSNameWrapper.CallOpts)
}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) Registrar() (common.Address, error) {
	return _ENSNameWrapper.Contract.Registrar(&_ENSNameWrapper.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _ENSNameWrapper.Contract.SupportsInterface(&_ENSNameWrapper.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _ENSNameWrapper.Contract.SupportsInterface(&_ENSNameWrapper.CallOpts, interfaceId)
}

// UpgradeContract is a free data retrieval call binding the contract method 0x1f4e1504.
//
// Solidity: function upgradeContract() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCaller) UpgradeContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "upgradeContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// UpgradeContract is a free data retrieval call binding the contract method 0x1f4e1504.
//
// Solidity: function upgradeContract() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperSession) UpgradeContract() (common.Address, error) {
	return _ENSNameWrapper.Contract.UpgradeContract(&_ENSNameWrapper.CallOpts)
}

// UpgradeContract is a free data retrieval call binding the contract method 0x1f4e1504.
//
// Solidity: function upgradeContract() view returns(address)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) UpgradeContract() (common.Address, error) {
	return _ENSNameWrapper.Contract.UpgradeContract(&_ENSNameWrapper.CallOpts)
}

// Uri is a free data retrieval call binding the contract method 0x0e89341c.
//
// Solidity: function uri(uint256 tokenId) view returns(string)
func (_ENSNameWrapper *ENSNameWrapperCaller) Uri(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _ENSNameWrapper.contract.Call(opts, &out, "uri", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Uri is a free data retrieval call binding the contract method 0x0e89341c.
//
// Solidity: function uri(uint256 tokenId) view returns(string)
func (_ENSNameWrapper *ENSNameWrapperSession) Uri(tokenId *big.Int) (string, error) {
	return _ENSNameWrapper.Contract.Uri(&_ENSNameWrapper.CallOpts, tokenId)
}

// Uri is a free data retrieval call binding the contract method 0x0e89341c.
//
// Solidity: function uri(uint256 tokenId) view returns(string)
func (_ENSNameWrapper *ENSNameWrapperCallerSession) Uri(tokenId *big.Int) (string, error) {
	return _ENSNameWrapper.Contract.Uri(&_ENSNameWrapper.CallOpts, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Approve(&_ENSNameWrapper.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Approve(&_ENSNameWrapper.TransactOpts, to, tokenId)
}

// ExtendExpiry is a paid mutator transaction binding the contract method 0x6e5d6ad2.
//
// Solidity: function extendExpiry(bytes32 parentNode, bytes32 labelhash, uint64 expiry) returns(uint64)
func (_ENSNameWrapper *ENSNameWrapperTransactor) ExtendExpiry(opts *bind.TransactOpts, parentNode [32]byte, labelhash [32]byte, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "extendExpiry", parentNode, labelhash, expiry)
}

// ExtendExpiry is a paid mutator transaction binding the contract method 0x6e5d6ad2.
//
// Solidity: function extendExpiry(bytes32 parentNode, bytes32 labelhash, uint64 expiry) returns(uint64)
func (_ENSNameWrapper *ENSNameWrapperSession) ExtendExpiry(parentNode [32]byte, labelhash [32]byte, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.ExtendExpiry(&_ENSNameWrapper.TransactOpts, parentNode, labelhash, expiry)
}

// ExtendExpiry is a paid mutator transaction binding the contract method 0x6e5d6ad2.
//
// Solidity: function extendExpiry(bytes32 parentNode, bytes32 labelhash, uint64 expiry) returns(uint64)
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) ExtendExpiry(parentNode [32]byte, labelhash [32]byte, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.ExtendExpiry(&_ENSNameWrapper.TransactOpts, parentNode, labelhash, expiry)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address to, address , uint256 tokenId, bytes data) returns(bytes4)
func (_ENSNameWrapper *ENSNameWrapperTransactor) OnERC721Received(opts *bind.TransactOpts, to common.Address, arg1 common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "onERC721Received", to, arg1, tokenId, data)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address to, address , uint256 tokenId, bytes data) returns(bytes4)
func (_ENSNameWrapper *ENSNameWrapperSession) OnERC721Received(to common.Address, arg1 common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.OnERC721Received(&_ENSNameWrapper.TransactOpts, to, arg1, tokenId, data)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address to, address , uint256 tokenId, bytes data) returns(bytes4)
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) OnERC721Received(to common.Address, arg1 common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.OnERC721Received(&_ENSNameWrapper.TransactOpts, to, arg1, tokenId, data)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) RecoverFunds(opts *bind.TransactOpts, _token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "recoverFunds", _token, _to, _amount)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) RecoverFunds(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.RecoverFunds(&_ENSNameWrapper.TransactOpts, _token, _to, _amount)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) RecoverFunds(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.RecoverFunds(&_ENSNameWrapper.TransactOpts, _token, _to, _amount)
}

// RegisterAndWrapETH2LD is a paid mutator transaction binding the contract method 0xa4014982.
//
// Solidity: function registerAndWrapETH2LD(string label, address wrappedOwner, uint256 duration, address resolver, uint16 ownerControlledFuses) returns(uint256 registrarExpiry)
func (_ENSNameWrapper *ENSNameWrapperTransactor) RegisterAndWrapETH2LD(opts *bind.TransactOpts, label string, wrappedOwner common.Address, duration *big.Int, resolver common.Address, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "registerAndWrapETH2LD", label, wrappedOwner, duration, resolver, ownerControlledFuses)
}

// RegisterAndWrapETH2LD is a paid mutator transaction binding the contract method 0xa4014982.
//
// Solidity: function registerAndWrapETH2LD(string label, address wrappedOwner, uint256 duration, address resolver, uint16 ownerControlledFuses) returns(uint256 registrarExpiry)
func (_ENSNameWrapper *ENSNameWrapperSession) RegisterAndWrapETH2LD(label string, wrappedOwner common.Address, duration *big.Int, resolver common.Address, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.RegisterAndWrapETH2LD(&_ENSNameWrapper.TransactOpts, label, wrappedOwner, duration, resolver, ownerControlledFuses)
}

// RegisterAndWrapETH2LD is a paid mutator transaction binding the contract method 0xa4014982.
//
// Solidity: function registerAndWrapETH2LD(string label, address wrappedOwner, uint256 duration, address resolver, uint16 ownerControlledFuses) returns(uint256 registrarExpiry)
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) RegisterAndWrapETH2LD(label string, wrappedOwner common.Address, duration *big.Int, resolver common.Address, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.RegisterAndWrapETH2LD(&_ENSNameWrapper.TransactOpts, label, wrappedOwner, duration, resolver, ownerControlledFuses)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 tokenId, uint256 duration) returns(uint256 expires)
func (_ENSNameWrapper *ENSNameWrapperTransactor) Renew(opts *bind.TransactOpts, tokenId *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "renew", tokenId, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 tokenId, uint256 duration) returns(uint256 expires)
func (_ENSNameWrapper *ENSNameWrapperSession) Renew(tokenId *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Renew(&_ENSNameWrapper.TransactOpts, tokenId, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 tokenId, uint256 duration) returns(uint256 expires)
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) Renew(tokenId *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Renew(&_ENSNameWrapper.TransactOpts, tokenId, duration)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSNameWrapper *ENSNameWrapperSession) RenounceOwnership() (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.RenounceOwnership(&_ENSNameWrapper.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.RenounceOwnership(&_ENSNameWrapper.TransactOpts)
}

// SafeBatchTransferFrom is a paid mutator transaction binding the contract method 0x2eb2c2d6.
//
// Solidity: function safeBatchTransferFrom(address from, address to, uint256[] ids, uint256[] amounts, bytes data) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SafeBatchTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, ids []*big.Int, amounts []*big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "safeBatchTransferFrom", from, to, ids, amounts, data)
}

// SafeBatchTransferFrom is a paid mutator transaction binding the contract method 0x2eb2c2d6.
//
// Solidity: function safeBatchTransferFrom(address from, address to, uint256[] ids, uint256[] amounts, bytes data) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SafeBatchTransferFrom(from common.Address, to common.Address, ids []*big.Int, amounts []*big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SafeBatchTransferFrom(&_ENSNameWrapper.TransactOpts, from, to, ids, amounts, data)
}

// SafeBatchTransferFrom is a paid mutator transaction binding the contract method 0x2eb2c2d6.
//
// Solidity: function safeBatchTransferFrom(address from, address to, uint256[] ids, uint256[] amounts, bytes data) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SafeBatchTransferFrom(from common.Address, to common.Address, ids []*big.Int, amounts []*big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SafeBatchTransferFrom(&_ENSNameWrapper.TransactOpts, from, to, ids, amounts, data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xf242432a.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 id, uint256 amount, bytes data) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, id *big.Int, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "safeTransferFrom", from, to, id, amount, data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xf242432a.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 id, uint256 amount, bytes data) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SafeTransferFrom(from common.Address, to common.Address, id *big.Int, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SafeTransferFrom(&_ENSNameWrapper.TransactOpts, from, to, id, amount, data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xf242432a.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 id, uint256 amount, bytes data) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SafeTransferFrom(from common.Address, to common.Address, id *big.Int, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SafeTransferFrom(&_ENSNameWrapper.TransactOpts, from, to, id, amount, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetApprovalForAll(&_ENSNameWrapper.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetApprovalForAll(&_ENSNameWrapper.TransactOpts, operator, approved)
}

// SetChildFuses is a paid mutator transaction binding the contract method 0x33c69ea9.
//
// Solidity: function setChildFuses(bytes32 parentNode, bytes32 labelhash, uint32 fuses, uint64 expiry) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetChildFuses(opts *bind.TransactOpts, parentNode [32]byte, labelhash [32]byte, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setChildFuses", parentNode, labelhash, fuses, expiry)
}

// SetChildFuses is a paid mutator transaction binding the contract method 0x33c69ea9.
//
// Solidity: function setChildFuses(bytes32 parentNode, bytes32 labelhash, uint32 fuses, uint64 expiry) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SetChildFuses(parentNode [32]byte, labelhash [32]byte, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetChildFuses(&_ENSNameWrapper.TransactOpts, parentNode, labelhash, fuses, expiry)
}

// SetChildFuses is a paid mutator transaction binding the contract method 0x33c69ea9.
//
// Solidity: function setChildFuses(bytes32 parentNode, bytes32 labelhash, uint32 fuses, uint64 expiry) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetChildFuses(parentNode [32]byte, labelhash [32]byte, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetChildFuses(&_ENSNameWrapper.TransactOpts, parentNode, labelhash, fuses, expiry)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool active) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetController(opts *bind.TransactOpts, controller common.Address, active bool) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setController", controller, active)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool active) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SetController(controller common.Address, active bool) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetController(&_ENSNameWrapper.TransactOpts, controller, active)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool active) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetController(controller common.Address, active bool) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetController(&_ENSNameWrapper.TransactOpts, controller, active)
}

// SetFuses is a paid mutator transaction binding the contract method 0x402906fc.
//
// Solidity: function setFuses(bytes32 node, uint16 ownerControlledFuses) returns(uint32)
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetFuses(opts *bind.TransactOpts, node [32]byte, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setFuses", node, ownerControlledFuses)
}

// SetFuses is a paid mutator transaction binding the contract method 0x402906fc.
//
// Solidity: function setFuses(bytes32 node, uint16 ownerControlledFuses) returns(uint32)
func (_ENSNameWrapper *ENSNameWrapperSession) SetFuses(node [32]byte, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetFuses(&_ENSNameWrapper.TransactOpts, node, ownerControlledFuses)
}

// SetFuses is a paid mutator transaction binding the contract method 0x402906fc.
//
// Solidity: function setFuses(bytes32 node, uint16 ownerControlledFuses) returns(uint32)
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetFuses(node [32]byte, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetFuses(&_ENSNameWrapper.TransactOpts, node, ownerControlledFuses)
}

// SetMetadataService is a paid mutator transaction binding the contract method 0x1534e177.
//
// Solidity: function setMetadataService(address _metadataService) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetMetadataService(opts *bind.TransactOpts, _metadataService common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setMetadataService", _metadataService)
}

// SetMetadataService is a paid mutator transaction binding the contract method 0x1534e177.
//
// Solidity: function setMetadataService(address _metadataService) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SetMetadataService(_metadataService common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetMetadataService(&_ENSNameWrapper.TransactOpts, _metadataService)
}

// SetMetadataService is a paid mutator transaction binding the contract method 0x1534e177.
//
// Solidity: function setMetadataService(address _metadataService) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetMetadataService(_metadataService common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetMetadataService(&_ENSNameWrapper.TransactOpts, _metadataService)
}

// SetRecord is a paid mutator transaction binding the contract method 0xcf408823.
//
// Solidity: function setRecord(bytes32 node, address owner, address resolver, uint64 ttl) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetRecord(opts *bind.TransactOpts, node [32]byte, owner common.Address, resolver common.Address, ttl uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setRecord", node, owner, resolver, ttl)
}

// SetRecord is a paid mutator transaction binding the contract method 0xcf408823.
//
// Solidity: function setRecord(bytes32 node, address owner, address resolver, uint64 ttl) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SetRecord(node [32]byte, owner common.Address, resolver common.Address, ttl uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetRecord(&_ENSNameWrapper.TransactOpts, node, owner, resolver, ttl)
}

// SetRecord is a paid mutator transaction binding the contract method 0xcf408823.
//
// Solidity: function setRecord(bytes32 node, address owner, address resolver, uint64 ttl) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetRecord(node [32]byte, owner common.Address, resolver common.Address, ttl uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetRecord(&_ENSNameWrapper.TransactOpts, node, owner, resolver, ttl)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetResolver(opts *bind.TransactOpts, node [32]byte, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setResolver", node, resolver)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SetResolver(node [32]byte, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetResolver(&_ENSNameWrapper.TransactOpts, node, resolver)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetResolver(node [32]byte, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetResolver(&_ENSNameWrapper.TransactOpts, node, resolver)
}

// SetSubnodeOwner is a paid mutator transaction binding the contract method 0xc658e086.
//
// Solidity: function setSubnodeOwner(bytes32 parentNode, string label, address owner, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetSubnodeOwner(opts *bind.TransactOpts, parentNode [32]byte, label string, owner common.Address, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setSubnodeOwner", parentNode, label, owner, fuses, expiry)
}

// SetSubnodeOwner is a paid mutator transaction binding the contract method 0xc658e086.
//
// Solidity: function setSubnodeOwner(bytes32 parentNode, string label, address owner, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_ENSNameWrapper *ENSNameWrapperSession) SetSubnodeOwner(parentNode [32]byte, label string, owner common.Address, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetSubnodeOwner(&_ENSNameWrapper.TransactOpts, parentNode, label, owner, fuses, expiry)
}

// SetSubnodeOwner is a paid mutator transaction binding the contract method 0xc658e086.
//
// Solidity: function setSubnodeOwner(bytes32 parentNode, string label, address owner, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetSubnodeOwner(parentNode [32]byte, label string, owner common.Address, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetSubnodeOwner(&_ENSNameWrapper.TransactOpts, parentNode, label, owner, fuses, expiry)
}

// SetSubnodeRecord is a paid mutator transaction binding the contract method 0x24c1af44.
//
// Solidity: function setSubnodeRecord(bytes32 parentNode, string label, address owner, address resolver, uint64 ttl, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetSubnodeRecord(opts *bind.TransactOpts, parentNode [32]byte, label string, owner common.Address, resolver common.Address, ttl uint64, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setSubnodeRecord", parentNode, label, owner, resolver, ttl, fuses, expiry)
}

// SetSubnodeRecord is a paid mutator transaction binding the contract method 0x24c1af44.
//
// Solidity: function setSubnodeRecord(bytes32 parentNode, string label, address owner, address resolver, uint64 ttl, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_ENSNameWrapper *ENSNameWrapperSession) SetSubnodeRecord(parentNode [32]byte, label string, owner common.Address, resolver common.Address, ttl uint64, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetSubnodeRecord(&_ENSNameWrapper.TransactOpts, parentNode, label, owner, resolver, ttl, fuses, expiry)
}

// SetSubnodeRecord is a paid mutator transaction binding the contract method 0x24c1af44.
//
// Solidity: function setSubnodeRecord(bytes32 parentNode, string label, address owner, address resolver, uint64 ttl, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetSubnodeRecord(parentNode [32]byte, label string, owner common.Address, resolver common.Address, ttl uint64, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetSubnodeRecord(&_ENSNameWrapper.TransactOpts, parentNode, label, owner, resolver, ttl, fuses, expiry)
}

// SetTTL is a paid mutator transaction binding the contract method 0x14ab9038.
//
// Solidity: function setTTL(bytes32 node, uint64 ttl) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetTTL(opts *bind.TransactOpts, node [32]byte, ttl uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setTTL", node, ttl)
}

// SetTTL is a paid mutator transaction binding the contract method 0x14ab9038.
//
// Solidity: function setTTL(bytes32 node, uint64 ttl) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SetTTL(node [32]byte, ttl uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetTTL(&_ENSNameWrapper.TransactOpts, node, ttl)
}

// SetTTL is a paid mutator transaction binding the contract method 0x14ab9038.
//
// Solidity: function setTTL(bytes32 node, uint64 ttl) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetTTL(node [32]byte, ttl uint64) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetTTL(&_ENSNameWrapper.TransactOpts, node, ttl)
}

// SetUpgradeContract is a paid mutator transaction binding the contract method 0xb6bcad26.
//
// Solidity: function setUpgradeContract(address _upgradeAddress) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) SetUpgradeContract(opts *bind.TransactOpts, _upgradeAddress common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "setUpgradeContract", _upgradeAddress)
}

// SetUpgradeContract is a paid mutator transaction binding the contract method 0xb6bcad26.
//
// Solidity: function setUpgradeContract(address _upgradeAddress) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) SetUpgradeContract(_upgradeAddress common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetUpgradeContract(&_ENSNameWrapper.TransactOpts, _upgradeAddress)
}

// SetUpgradeContract is a paid mutator transaction binding the contract method 0xb6bcad26.
//
// Solidity: function setUpgradeContract(address _upgradeAddress) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) SetUpgradeContract(_upgradeAddress common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.SetUpgradeContract(&_ENSNameWrapper.TransactOpts, _upgradeAddress)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.TransferOwnership(&_ENSNameWrapper.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.TransferOwnership(&_ENSNameWrapper.TransactOpts, newOwner)
}

// Unwrap is a paid mutator transaction binding the contract method 0xd8c9921a.
//
// Solidity: function unwrap(bytes32 parentNode, bytes32 labelhash, address controller) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) Unwrap(opts *bind.TransactOpts, parentNode [32]byte, labelhash [32]byte, controller common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "unwrap", parentNode, labelhash, controller)
}

// Unwrap is a paid mutator transaction binding the contract method 0xd8c9921a.
//
// Solidity: function unwrap(bytes32 parentNode, bytes32 labelhash, address controller) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) Unwrap(parentNode [32]byte, labelhash [32]byte, controller common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Unwrap(&_ENSNameWrapper.TransactOpts, parentNode, labelhash, controller)
}

// Unwrap is a paid mutator transaction binding the contract method 0xd8c9921a.
//
// Solidity: function unwrap(bytes32 parentNode, bytes32 labelhash, address controller) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) Unwrap(parentNode [32]byte, labelhash [32]byte, controller common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Unwrap(&_ENSNameWrapper.TransactOpts, parentNode, labelhash, controller)
}

// UnwrapETH2LD is a paid mutator transaction binding the contract method 0x8b4dfa75.
//
// Solidity: function unwrapETH2LD(bytes32 labelhash, address registrant, address controller) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) UnwrapETH2LD(opts *bind.TransactOpts, labelhash [32]byte, registrant common.Address, controller common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "unwrapETH2LD", labelhash, registrant, controller)
}

// UnwrapETH2LD is a paid mutator transaction binding the contract method 0x8b4dfa75.
//
// Solidity: function unwrapETH2LD(bytes32 labelhash, address registrant, address controller) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) UnwrapETH2LD(labelhash [32]byte, registrant common.Address, controller common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.UnwrapETH2LD(&_ENSNameWrapper.TransactOpts, labelhash, registrant, controller)
}

// UnwrapETH2LD is a paid mutator transaction binding the contract method 0x8b4dfa75.
//
// Solidity: function unwrapETH2LD(bytes32 labelhash, address registrant, address controller) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) UnwrapETH2LD(labelhash [32]byte, registrant common.Address, controller common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.UnwrapETH2LD(&_ENSNameWrapper.TransactOpts, labelhash, registrant, controller)
}

// Upgrade is a paid mutator transaction binding the contract method 0xc93ab3fd.
//
// Solidity: function upgrade(bytes name, bytes extraData) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) Upgrade(opts *bind.TransactOpts, name []byte, extraData []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "upgrade", name, extraData)
}

// Upgrade is a paid mutator transaction binding the contract method 0xc93ab3fd.
//
// Solidity: function upgrade(bytes name, bytes extraData) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) Upgrade(name []byte, extraData []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Upgrade(&_ENSNameWrapper.TransactOpts, name, extraData)
}

// Upgrade is a paid mutator transaction binding the contract method 0xc93ab3fd.
//
// Solidity: function upgrade(bytes name, bytes extraData) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) Upgrade(name []byte, extraData []byte) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Upgrade(&_ENSNameWrapper.TransactOpts, name, extraData)
}

// Wrap is a paid mutator transaction binding the contract method 0xeb8ae530.
//
// Solidity: function wrap(bytes name, address wrappedOwner, address resolver) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactor) Wrap(opts *bind.TransactOpts, name []byte, wrappedOwner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "wrap", name, wrappedOwner, resolver)
}

// Wrap is a paid mutator transaction binding the contract method 0xeb8ae530.
//
// Solidity: function wrap(bytes name, address wrappedOwner, address resolver) returns()
func (_ENSNameWrapper *ENSNameWrapperSession) Wrap(name []byte, wrappedOwner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Wrap(&_ENSNameWrapper.TransactOpts, name, wrappedOwner, resolver)
}

// Wrap is a paid mutator transaction binding the contract method 0xeb8ae530.
//
// Solidity: function wrap(bytes name, address wrappedOwner, address resolver) returns()
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) Wrap(name []byte, wrappedOwner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.Wrap(&_ENSNameWrapper.TransactOpts, name, wrappedOwner, resolver)
}

// WrapETH2LD is a paid mutator transaction binding the contract method 0x8cf8b41e.
//
// Solidity: function wrapETH2LD(string label, address wrappedOwner, uint16 ownerControlledFuses, address resolver) returns(uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperTransactor) WrapETH2LD(opts *bind.TransactOpts, label string, wrappedOwner common.Address, ownerControlledFuses uint16, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.contract.Transact(opts, "wrapETH2LD", label, wrappedOwner, ownerControlledFuses, resolver)
}

// WrapETH2LD is a paid mutator transaction binding the contract method 0x8cf8b41e.
//
// Solidity: function wrapETH2LD(string label, address wrappedOwner, uint16 ownerControlledFuses, address resolver) returns(uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperSession) WrapETH2LD(label string, wrappedOwner common.Address, ownerControlledFuses uint16, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.WrapETH2LD(&_ENSNameWrapper.TransactOpts, label, wrappedOwner, ownerControlledFuses, resolver)
}

// WrapETH2LD is a paid mutator transaction binding the contract method 0x8cf8b41e.
//
// Solidity: function wrapETH2LD(string label, address wrappedOwner, uint16 ownerControlledFuses, address resolver) returns(uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperTransactorSession) WrapETH2LD(label string, wrappedOwner common.Address, ownerControlledFuses uint16, resolver common.Address) (*types.Transaction, error) {
	return _ENSNameWrapper.Contract.WrapETH2LD(&_ENSNameWrapper.TransactOpts, label, wrappedOwner, ownerControlledFuses, resolver)
}

// ENSNameWrapperApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ENSNameWrapper contract.
type ENSNameWrapperApprovalIterator struct {
	Event *ENSNameWrapperApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperApproval represents a Approval event raised by the ENSNameWrapper contract.
type ENSNameWrapperApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*ENSNameWrapperApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperApprovalIterator{contract: _ENSNameWrapper.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperApproval)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseApproval(log types.Log) (*ENSNameWrapperApproval, error) {
	event := new(ENSNameWrapperApproval)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the ENSNameWrapper contract.
type ENSNameWrapperApprovalForAllIterator struct {
	Event *ENSNameWrapperApprovalForAll // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperApprovalForAll)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperApprovalForAll)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperApprovalForAll represents a ApprovalForAll event raised by the ENSNameWrapper contract.
type ENSNameWrapperApprovalForAll struct {
	Account  common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed account, address indexed operator, bool approved)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterApprovalForAll(opts *bind.FilterOpts, account []common.Address, operator []common.Address) (*ENSNameWrapperApprovalForAllIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "ApprovalForAll", accountRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperApprovalForAllIterator{contract: _ENSNameWrapper.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed account, address indexed operator, bool approved)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperApprovalForAll, account []common.Address, operator []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "ApprovalForAll", accountRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperApprovalForAll)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed account, address indexed operator, bool approved)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseApprovalForAll(log types.Log) (*ENSNameWrapperApprovalForAll, error) {
	event := new(ENSNameWrapperApprovalForAll)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperControllerChangedIterator is returned from FilterControllerChanged and is used to iterate over the raw logs and unpacked data for ControllerChanged events raised by the ENSNameWrapper contract.
type ENSNameWrapperControllerChangedIterator struct {
	Event *ENSNameWrapperControllerChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperControllerChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperControllerChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperControllerChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperControllerChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperControllerChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperControllerChanged represents a ControllerChanged event raised by the ENSNameWrapper contract.
type ENSNameWrapperControllerChanged struct {
	Controller common.Address
	Active     bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterControllerChanged is a free log retrieval operation binding the contract event 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87.
//
// Solidity: event ControllerChanged(address indexed controller, bool active)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterControllerChanged(opts *bind.FilterOpts, controller []common.Address) (*ENSNameWrapperControllerChangedIterator, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "ControllerChanged", controllerRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperControllerChangedIterator{contract: _ENSNameWrapper.contract, event: "ControllerChanged", logs: logs, sub: sub}, nil
}

// WatchControllerChanged is a free log subscription operation binding the contract event 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87.
//
// Solidity: event ControllerChanged(address indexed controller, bool active)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchControllerChanged(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperControllerChanged, controller []common.Address) (event.Subscription, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "ControllerChanged", controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperControllerChanged)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "ControllerChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseControllerChanged is a log parse operation binding the contract event 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87.
//
// Solidity: event ControllerChanged(address indexed controller, bool active)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseControllerChanged(log types.Log) (*ENSNameWrapperControllerChanged, error) {
	event := new(ENSNameWrapperControllerChanged)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "ControllerChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperExpiryExtendedIterator is returned from FilterExpiryExtended and is used to iterate over the raw logs and unpacked data for ExpiryExtended events raised by the ENSNameWrapper contract.
type ENSNameWrapperExpiryExtendedIterator struct {
	Event *ENSNameWrapperExpiryExtended // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperExpiryExtendedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperExpiryExtended)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperExpiryExtended)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperExpiryExtendedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperExpiryExtendedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperExpiryExtended represents a ExpiryExtended event raised by the ENSNameWrapper contract.
type ENSNameWrapperExpiryExtended struct {
	Node   [32]byte
	Expiry uint64
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExpiryExtended is a free log retrieval operation binding the contract event 0xf675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b132.
//
// Solidity: event ExpiryExtended(bytes32 indexed node, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterExpiryExtended(opts *bind.FilterOpts, node [][32]byte) (*ENSNameWrapperExpiryExtendedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "ExpiryExtended", nodeRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperExpiryExtendedIterator{contract: _ENSNameWrapper.contract, event: "ExpiryExtended", logs: logs, sub: sub}, nil
}

// WatchExpiryExtended is a free log subscription operation binding the contract event 0xf675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b132.
//
// Solidity: event ExpiryExtended(bytes32 indexed node, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchExpiryExtended(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperExpiryExtended, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "ExpiryExtended", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperExpiryExtended)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "ExpiryExtended", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseExpiryExtended is a log parse operation binding the contract event 0xf675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b132.
//
// Solidity: event ExpiryExtended(bytes32 indexed node, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseExpiryExtended(log types.Log) (*ENSNameWrapperExpiryExtended, error) {
	event := new(ENSNameWrapperExpiryExtended)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "ExpiryExtended", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperFusesSetIterator is returned from FilterFusesSet and is used to iterate over the raw logs and unpacked data for FusesSet events raised by the ENSNameWrapper contract.
type ENSNameWrapperFusesSetIterator struct {
	Event *ENSNameWrapperFusesSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperFusesSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperFusesSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperFusesSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperFusesSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperFusesSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperFusesSet represents a FusesSet event raised by the ENSNameWrapper contract.
type ENSNameWrapperFusesSet struct {
	Node  [32]byte
	Fuses uint32
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterFusesSet is a free log retrieval operation binding the contract event 0x39873f00c80f4f94b7bd1594aebcf650f003545b74824d57ddf4939e3ff3a34b.
//
// Solidity: event FusesSet(bytes32 indexed node, uint32 fuses)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterFusesSet(opts *bind.FilterOpts, node [][32]byte) (*ENSNameWrapperFusesSetIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "FusesSet", nodeRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperFusesSetIterator{contract: _ENSNameWrapper.contract, event: "FusesSet", logs: logs, sub: sub}, nil
}

// WatchFusesSet is a free log subscription operation binding the contract event 0x39873f00c80f4f94b7bd1594aebcf650f003545b74824d57ddf4939e3ff3a34b.
//
// Solidity: event FusesSet(bytes32 indexed node, uint32 fuses)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchFusesSet(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperFusesSet, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "FusesSet", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperFusesSet)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "FusesSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFusesSet is a log parse operation binding the contract event 0x39873f00c80f4f94b7bd1594aebcf650f003545b74824d57ddf4939e3ff3a34b.
//
// Solidity: event FusesSet(bytes32 indexed node, uint32 fuses)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseFusesSet(log types.Log) (*ENSNameWrapperFusesSet, error) {
	event := new(ENSNameWrapperFusesSet)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "FusesSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperNameUnwrappedIterator is returned from FilterNameUnwrapped and is used to iterate over the raw logs and unpacked data for NameUnwrapped events raised by the ENSNameWrapper contract.
type ENSNameWrapperNameUnwrappedIterator struct {
	Event *ENSNameWrapperNameUnwrapped // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperNameUnwrappedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperNameUnwrapped)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperNameUnwrapped)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperNameUnwrappedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperNameUnwrappedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperNameUnwrapped represents a NameUnwrapped event raised by the ENSNameWrapper contract.
type ENSNameWrapperNameUnwrapped struct {
	Node  [32]byte
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNameUnwrapped is a free log retrieval operation binding the contract event 0xee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c4.
//
// Solidity: event NameUnwrapped(bytes32 indexed node, address owner)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterNameUnwrapped(opts *bind.FilterOpts, node [][32]byte) (*ENSNameWrapperNameUnwrappedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "NameUnwrapped", nodeRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperNameUnwrappedIterator{contract: _ENSNameWrapper.contract, event: "NameUnwrapped", logs: logs, sub: sub}, nil
}

// WatchNameUnwrapped is a free log subscription operation binding the contract event 0xee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c4.
//
// Solidity: event NameUnwrapped(bytes32 indexed node, address owner)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchNameUnwrapped(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperNameUnwrapped, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "NameUnwrapped", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperNameUnwrapped)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "NameUnwrapped", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNameUnwrapped is a log parse operation binding the contract event 0xee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c4.
//
// Solidity: event NameUnwrapped(bytes32 indexed node, address owner)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseNameUnwrapped(log types.Log) (*ENSNameWrapperNameUnwrapped, error) {
	event := new(ENSNameWrapperNameUnwrapped)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "NameUnwrapped", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperNameWrappedIterator is returned from FilterNameWrapped and is used to iterate over the raw logs and unpacked data for NameWrapped events raised by the ENSNameWrapper contract.
type ENSNameWrapperNameWrappedIterator struct {
	Event *ENSNameWrapperNameWrapped // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperNameWrappedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperNameWrapped)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperNameWrapped)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperNameWrappedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperNameWrappedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperNameWrapped represents a NameWrapped event raised by the ENSNameWrapper contract.
type ENSNameWrapperNameWrapped struct {
	Node   [32]byte
	Name   []byte
	Owner  common.Address
	Fuses  uint32
	Expiry uint64
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterNameWrapped is a free log retrieval operation binding the contract event 0x8ce7013e8abebc55c3890a68f5a27c67c3f7efa64e584de5fb22363c606fd340.
//
// Solidity: event NameWrapped(bytes32 indexed node, bytes name, address owner, uint32 fuses, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterNameWrapped(opts *bind.FilterOpts, node [][32]byte) (*ENSNameWrapperNameWrappedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "NameWrapped", nodeRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperNameWrappedIterator{contract: _ENSNameWrapper.contract, event: "NameWrapped", logs: logs, sub: sub}, nil
}

// WatchNameWrapped is a free log subscription operation binding the contract event 0x8ce7013e8abebc55c3890a68f5a27c67c3f7efa64e584de5fb22363c606fd340.
//
// Solidity: event NameWrapped(bytes32 indexed node, bytes name, address owner, uint32 fuses, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchNameWrapped(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperNameWrapped, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "NameWrapped", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperNameWrapped)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "NameWrapped", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNameWrapped is a log parse operation binding the contract event 0x8ce7013e8abebc55c3890a68f5a27c67c3f7efa64e584de5fb22363c606fd340.
//
// Solidity: event NameWrapped(bytes32 indexed node, bytes name, address owner, uint32 fuses, uint64 expiry)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseNameWrapped(log types.Log) (*ENSNameWrapperNameWrapped, error) {
	event := new(ENSNameWrapperNameWrapped)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "NameWrapped", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ENSNameWrapper contract.
type ENSNameWrapperOwnershipTransferredIterator struct {
	Event *ENSNameWrapperOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperOwnershipTransferred represents a OwnershipTransferred event raised by the ENSNameWrapper contract.
type ENSNameWrapperOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ENSNameWrapperOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperOwnershipTransferredIterator{contract: _ENSNameWrapper.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperOwnershipTransferred)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseOwnershipTransferred(log types.Log) (*ENSNameWrapperOwnershipTransferred, error) {
	event := new(ENSNameWrapperOwnershipTransferred)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperTransferBatchIterator is returned from FilterTransferBatch and is used to iterate over the raw logs and unpacked data for TransferBatch events raised by the ENSNameWrapper contract.
type ENSNameWrapperTransferBatchIterator struct {
	Event *ENSNameWrapperTransferBatch // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperTransferBatchIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperTransferBatch)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperTransferBatch)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperTransferBatchIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperTransferBatchIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperTransferBatch represents a TransferBatch event raised by the ENSNameWrapper contract.
type ENSNameWrapperTransferBatch struct {
	Operator common.Address
	From     common.Address
	To       common.Address
	Ids      []*big.Int
	Values   []*big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterTransferBatch is a free log retrieval operation binding the contract event 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb.
//
// Solidity: event TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterTransferBatch(opts *bind.FilterOpts, operator []common.Address, from []common.Address, to []common.Address) (*ENSNameWrapperTransferBatchIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "TransferBatch", operatorRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperTransferBatchIterator{contract: _ENSNameWrapper.contract, event: "TransferBatch", logs: logs, sub: sub}, nil
}

// WatchTransferBatch is a free log subscription operation binding the contract event 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb.
//
// Solidity: event TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchTransferBatch(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperTransferBatch, operator []common.Address, from []common.Address, to []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "TransferBatch", operatorRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperTransferBatch)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "TransferBatch", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransferBatch is a log parse operation binding the contract event 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb.
//
// Solidity: event TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseTransferBatch(log types.Log) (*ENSNameWrapperTransferBatch, error) {
	event := new(ENSNameWrapperTransferBatch)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "TransferBatch", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperTransferSingleIterator is returned from FilterTransferSingle and is used to iterate over the raw logs and unpacked data for TransferSingle events raised by the ENSNameWrapper contract.
type ENSNameWrapperTransferSingleIterator struct {
	Event *ENSNameWrapperTransferSingle // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperTransferSingleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperTransferSingle)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperTransferSingle)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperTransferSingleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperTransferSingleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperTransferSingle represents a TransferSingle event raised by the ENSNameWrapper contract.
type ENSNameWrapperTransferSingle struct {
	Operator common.Address
	From     common.Address
	To       common.Address
	Id       *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterTransferSingle is a free log retrieval operation binding the contract event 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62.
//
// Solidity: event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterTransferSingle(opts *bind.FilterOpts, operator []common.Address, from []common.Address, to []common.Address) (*ENSNameWrapperTransferSingleIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "TransferSingle", operatorRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperTransferSingleIterator{contract: _ENSNameWrapper.contract, event: "TransferSingle", logs: logs, sub: sub}, nil
}

// WatchTransferSingle is a free log subscription operation binding the contract event 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62.
//
// Solidity: event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchTransferSingle(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperTransferSingle, operator []common.Address, from []common.Address, to []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "TransferSingle", operatorRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperTransferSingle)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "TransferSingle", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransferSingle is a log parse operation binding the contract event 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62.
//
// Solidity: event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseTransferSingle(log types.Log) (*ENSNameWrapperTransferSingle, error) {
	event := new(ENSNameWrapperTransferSingle)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "TransferSingle", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSNameWrapperURIIterator is returned from FilterURI and is used to iterate over the raw logs and unpacked data for URI events raised by the ENSNameWrapper contract.
type ENSNameWrapperURIIterator struct {
	Event *ENSNameWrapperURI // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ENSNameWrapperURIIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSNameWrapperURI)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ENSNameWrapperURI)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ENSNameWrapperURIIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSNameWrapperURIIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSNameWrapperURI represents a URI event raised by the ENSNameWrapper contract.
type ENSNameWrapperURI struct {
	Value string
	Id    *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterURI is a free log retrieval operation binding the contract event 0x6bb7ff708619ba0610cba295a58592e0451dee2622938c8755667688daf3529b.
//
// Solidity: event URI(string value, uint256 indexed id)
func (_ENSNameWrapper *ENSNameWrapperFilterer) FilterURI(opts *bind.FilterOpts, id []*big.Int) (*ENSNameWrapperURIIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.FilterLogs(opts, "URI", idRule)
	if err != nil {
		return nil, err
	}
	return &ENSNameWrapperURIIterator{contract: _ENSNameWrapper.contract, event: "URI", logs: logs, sub: sub}, nil
}

// WatchURI is a free log subscription operation binding the contract event 0x6bb7ff708619ba0610cba295a58592e0451dee2622938c8755667688daf3529b.
//
// Solidity: event URI(string value, uint256 indexed id)
func (_ENSNameWrapper *ENSNameWrapperFilterer) WatchURI(opts *bind.WatchOpts, sink chan<- *ENSNameWrapperURI, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _ENSNameWrapper.contract.WatchLogs(opts, "URI", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSNameWrapperURI)
				if err := _ENSNameWrapper.contract.UnpackLog(event, "URI", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseURI is a log parse operation binding the contract event 0x6bb7ff708619ba0610cba295a58592e0451dee2622938c8755667688daf3529b.
//
// Solidity: event URI(string value, uint256 indexed id)
func (_ENSNameWrapper *ENSNameWrapperFilterer) ParseURI(log types.Log) (*ENSNameWrapperURI, error) {
	event := new(ENSNameWrapperURI)
	if err := _ENSNameWrapper.contract.UnpackLog(event, "URI", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
