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

// DNSSECRRSetWithSignature is an auto generated low-level Go binding around an user-defined struct.
type DNSSECRRSetWithSignature struct {
	Rrset []byte
	Sig   []byte
}

// ENSDNSRegistrarMetaData contains all meta data concerning the ENSDNSRegistrar contract.
var ENSDNSRegistrarMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_previousRegistrar\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_resolver\",\"type\":\"address\"},{\"internalType\":\"contractDNSSEC\",\"name\":\"_dnssec\",\"type\":\"address\"},{\"internalType\":\"contractPublicSuffixList\",\"name\":\"_suffixes\",\"type\":\"address\"},{\"internalType\":\"contractENS\",\"name\":\"_ens\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"}],\"name\":\"InvalidPublicSuffix\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NoOwnerRecordFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"OffsetOutOfBoundsError\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"PermissionDenied\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PreconditionNotMet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"StaleProof\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"dnsname\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"inception\",\"type\":\"uint32\"}],\"name\":\"Claim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"suffixes\",\"type\":\"address\"}],\"name\":\"NewPublicSuffixList\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"domain\",\"type\":\"bytes\"}],\"name\":\"enableNode\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ens\",\"outputs\":[{\"internalType\":\"contractENS\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"inceptions\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"oracle\",\"outputs\":[{\"internalType\":\"contractDNSSEC\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"previousRegistrar\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"rrset\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"internalType\":\"structDNSSEC.RRSetWithSignature[]\",\"name\":\"input\",\"type\":\"tuple[]\"}],\"name\":\"proveAndClaim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"rrset\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"internalType\":\"structDNSSEC.RRSetWithSignature[]\",\"name\":\"input\",\"type\":\"tuple[]\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"proveAndClaimWithResolver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"resolver\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractPublicSuffixList\",\"name\":\"_suffixes\",\"type\":\"address\"}],\"name\":\"setPublicSuffixList\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"suffixes\",\"outputs\":[{\"internalType\":\"contractPublicSuffixList\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceID\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// ENSDNSRegistrarABI is the input ABI used to generate the binding from.
// Deprecated: Use ENSDNSRegistrarMetaData.ABI instead.
var ENSDNSRegistrarABI = ENSDNSRegistrarMetaData.ABI

// ENSDNSRegistrar is an auto generated Go binding around an Ethereum contract.
type ENSDNSRegistrar struct {
	ENSDNSRegistrarCaller     // Read-only binding to the contract
	ENSDNSRegistrarTransactor // Write-only binding to the contract
	ENSDNSRegistrarFilterer   // Log filterer for contract events
}

// ENSDNSRegistrarCaller is an auto generated read-only Go binding around an Ethereum contract.
type ENSDNSRegistrarCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSDNSRegistrarTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ENSDNSRegistrarTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSDNSRegistrarFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ENSDNSRegistrarFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSDNSRegistrarSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ENSDNSRegistrarSession struct {
	Contract     *ENSDNSRegistrar  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ENSDNSRegistrarCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ENSDNSRegistrarCallerSession struct {
	Contract *ENSDNSRegistrarCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// ENSDNSRegistrarTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ENSDNSRegistrarTransactorSession struct {
	Contract     *ENSDNSRegistrarTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ENSDNSRegistrarRaw is an auto generated low-level Go binding around an Ethereum contract.
type ENSDNSRegistrarRaw struct {
	Contract *ENSDNSRegistrar // Generic contract binding to access the raw methods on
}

// ENSDNSRegistrarCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ENSDNSRegistrarCallerRaw struct {
	Contract *ENSDNSRegistrarCaller // Generic read-only contract binding to access the raw methods on
}

// ENSDNSRegistrarTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ENSDNSRegistrarTransactorRaw struct {
	Contract *ENSDNSRegistrarTransactor // Generic write-only contract binding to access the raw methods on
}

// NewENSDNSRegistrar creates a new instance of ENSDNSRegistrar, bound to a specific deployed contract.
func NewENSDNSRegistrar(address common.Address, backend bind.ContractBackend) (*ENSDNSRegistrar, error) {
	contract, err := bindENSDNSRegistrar(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ENSDNSRegistrar{ENSDNSRegistrarCaller: ENSDNSRegistrarCaller{contract: contract}, ENSDNSRegistrarTransactor: ENSDNSRegistrarTransactor{contract: contract}, ENSDNSRegistrarFilterer: ENSDNSRegistrarFilterer{contract: contract}}, nil
}

// NewENSDNSRegistrarCaller creates a new read-only instance of ENSDNSRegistrar, bound to a specific deployed contract.
func NewENSDNSRegistrarCaller(address common.Address, caller bind.ContractCaller) (*ENSDNSRegistrarCaller, error) {
	contract, err := bindENSDNSRegistrar(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ENSDNSRegistrarCaller{contract: contract}, nil
}

// NewENSDNSRegistrarTransactor creates a new write-only instance of ENSDNSRegistrar, bound to a specific deployed contract.
func NewENSDNSRegistrarTransactor(address common.Address, transactor bind.ContractTransactor) (*ENSDNSRegistrarTransactor, error) {
	contract, err := bindENSDNSRegistrar(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ENSDNSRegistrarTransactor{contract: contract}, nil
}

// NewENSDNSRegistrarFilterer creates a new log filterer instance of ENSDNSRegistrar, bound to a specific deployed contract.
func NewENSDNSRegistrarFilterer(address common.Address, filterer bind.ContractFilterer) (*ENSDNSRegistrarFilterer, error) {
	contract, err := bindENSDNSRegistrar(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ENSDNSRegistrarFilterer{contract: contract}, nil
}

// bindENSDNSRegistrar binds a generic wrapper to an already deployed contract.
func bindENSDNSRegistrar(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ENSDNSRegistrarMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ENSDNSRegistrar *ENSDNSRegistrarRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ENSDNSRegistrar.Contract.ENSDNSRegistrarCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ENSDNSRegistrar *ENSDNSRegistrarRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.ENSDNSRegistrarTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ENSDNSRegistrar *ENSDNSRegistrarRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.ENSDNSRegistrarTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ENSDNSRegistrar *ENSDNSRegistrarCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ENSDNSRegistrar.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.contract.Transact(opts, method, params...)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCaller) Ens(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSDNSRegistrar.contract.Call(opts, &out, "ens")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) Ens() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.Ens(&_ENSDNSRegistrar.CallOpts)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCallerSession) Ens() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.Ens(&_ENSDNSRegistrar.CallOpts)
}

// Inceptions is a free data retrieval call binding the contract method 0x25916d41.
//
// Solidity: function inceptions(bytes32 ) view returns(uint32)
func (_ENSDNSRegistrar *ENSDNSRegistrarCaller) Inceptions(opts *bind.CallOpts, arg0 [32]byte) (uint32, error) {
	var out []interface{}
	err := _ENSDNSRegistrar.contract.Call(opts, &out, "inceptions", arg0)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// Inceptions is a free data retrieval call binding the contract method 0x25916d41.
//
// Solidity: function inceptions(bytes32 ) view returns(uint32)
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) Inceptions(arg0 [32]byte) (uint32, error) {
	return _ENSDNSRegistrar.Contract.Inceptions(&_ENSDNSRegistrar.CallOpts, arg0)
}

// Inceptions is a free data retrieval call binding the contract method 0x25916d41.
//
// Solidity: function inceptions(bytes32 ) view returns(uint32)
func (_ENSDNSRegistrar *ENSDNSRegistrarCallerSession) Inceptions(arg0 [32]byte) (uint32, error) {
	return _ENSDNSRegistrar.Contract.Inceptions(&_ENSDNSRegistrar.CallOpts, arg0)
}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCaller) Oracle(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSDNSRegistrar.contract.Call(opts, &out, "oracle")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) Oracle() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.Oracle(&_ENSDNSRegistrar.CallOpts)
}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCallerSession) Oracle() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.Oracle(&_ENSDNSRegistrar.CallOpts)
}

// PreviousRegistrar is a free data retrieval call binding the contract method 0xab14ec59.
//
// Solidity: function previousRegistrar() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCaller) PreviousRegistrar(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSDNSRegistrar.contract.Call(opts, &out, "previousRegistrar")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PreviousRegistrar is a free data retrieval call binding the contract method 0xab14ec59.
//
// Solidity: function previousRegistrar() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) PreviousRegistrar() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.PreviousRegistrar(&_ENSDNSRegistrar.CallOpts)
}

// PreviousRegistrar is a free data retrieval call binding the contract method 0xab14ec59.
//
// Solidity: function previousRegistrar() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCallerSession) PreviousRegistrar() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.PreviousRegistrar(&_ENSDNSRegistrar.CallOpts)
}

// Resolver is a free data retrieval call binding the contract method 0x04f3bcec.
//
// Solidity: function resolver() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCaller) Resolver(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSDNSRegistrar.contract.Call(opts, &out, "resolver")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Resolver is a free data retrieval call binding the contract method 0x04f3bcec.
//
// Solidity: function resolver() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) Resolver() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.Resolver(&_ENSDNSRegistrar.CallOpts)
}

// Resolver is a free data retrieval call binding the contract method 0x04f3bcec.
//
// Solidity: function resolver() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCallerSession) Resolver() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.Resolver(&_ENSDNSRegistrar.CallOpts)
}

// Suffixes is a free data retrieval call binding the contract method 0x30349ebe.
//
// Solidity: function suffixes() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCaller) Suffixes(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSDNSRegistrar.contract.Call(opts, &out, "suffixes")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Suffixes is a free data retrieval call binding the contract method 0x30349ebe.
//
// Solidity: function suffixes() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) Suffixes() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.Suffixes(&_ENSDNSRegistrar.CallOpts)
}

// Suffixes is a free data retrieval call binding the contract method 0x30349ebe.
//
// Solidity: function suffixes() view returns(address)
func (_ENSDNSRegistrar *ENSDNSRegistrarCallerSession) Suffixes() (common.Address, error) {
	return _ENSDNSRegistrar.Contract.Suffixes(&_ENSDNSRegistrar.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_ENSDNSRegistrar *ENSDNSRegistrarCaller) SupportsInterface(opts *bind.CallOpts, interfaceID [4]byte) (bool, error) {
	var out []interface{}
	err := _ENSDNSRegistrar.contract.Call(opts, &out, "supportsInterface", interfaceID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _ENSDNSRegistrar.Contract.SupportsInterface(&_ENSDNSRegistrar.CallOpts, interfaceID)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_ENSDNSRegistrar *ENSDNSRegistrarCallerSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _ENSDNSRegistrar.Contract.SupportsInterface(&_ENSDNSRegistrar.CallOpts, interfaceID)
}

// EnableNode is a paid mutator transaction binding the contract method 0x6f951221.
//
// Solidity: function enableNode(bytes domain) returns(bytes32 node)
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactor) EnableNode(opts *bind.TransactOpts, domain []byte) (*types.Transaction, error) {
	return _ENSDNSRegistrar.contract.Transact(opts, "enableNode", domain)
}

// EnableNode is a paid mutator transaction binding the contract method 0x6f951221.
//
// Solidity: function enableNode(bytes domain) returns(bytes32 node)
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) EnableNode(domain []byte) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.EnableNode(&_ENSDNSRegistrar.TransactOpts, domain)
}

// EnableNode is a paid mutator transaction binding the contract method 0x6f951221.
//
// Solidity: function enableNode(bytes domain) returns(bytes32 node)
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactorSession) EnableNode(domain []byte) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.EnableNode(&_ENSDNSRegistrar.TransactOpts, domain)
}

// ProveAndClaim is a paid mutator transaction binding the contract method 0x29d56630.
//
// Solidity: function proveAndClaim(bytes name, (bytes,bytes)[] input) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactor) ProveAndClaim(opts *bind.TransactOpts, name []byte, input []DNSSECRRSetWithSignature) (*types.Transaction, error) {
	return _ENSDNSRegistrar.contract.Transact(opts, "proveAndClaim", name, input)
}

// ProveAndClaim is a paid mutator transaction binding the contract method 0x29d56630.
//
// Solidity: function proveAndClaim(bytes name, (bytes,bytes)[] input) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) ProveAndClaim(name []byte, input []DNSSECRRSetWithSignature) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.ProveAndClaim(&_ENSDNSRegistrar.TransactOpts, name, input)
}

// ProveAndClaim is a paid mutator transaction binding the contract method 0x29d56630.
//
// Solidity: function proveAndClaim(bytes name, (bytes,bytes)[] input) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactorSession) ProveAndClaim(name []byte, input []DNSSECRRSetWithSignature) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.ProveAndClaim(&_ENSDNSRegistrar.TransactOpts, name, input)
}

// ProveAndClaimWithResolver is a paid mutator transaction binding the contract method 0x06963218.
//
// Solidity: function proveAndClaimWithResolver(bytes name, (bytes,bytes)[] input, address resolver, address addr) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactor) ProveAndClaimWithResolver(opts *bind.TransactOpts, name []byte, input []DNSSECRRSetWithSignature, resolver common.Address, addr common.Address) (*types.Transaction, error) {
	return _ENSDNSRegistrar.contract.Transact(opts, "proveAndClaimWithResolver", name, input, resolver, addr)
}

// ProveAndClaimWithResolver is a paid mutator transaction binding the contract method 0x06963218.
//
// Solidity: function proveAndClaimWithResolver(bytes name, (bytes,bytes)[] input, address resolver, address addr) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) ProveAndClaimWithResolver(name []byte, input []DNSSECRRSetWithSignature, resolver common.Address, addr common.Address) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.ProveAndClaimWithResolver(&_ENSDNSRegistrar.TransactOpts, name, input, resolver, addr)
}

// ProveAndClaimWithResolver is a paid mutator transaction binding the contract method 0x06963218.
//
// Solidity: function proveAndClaimWithResolver(bytes name, (bytes,bytes)[] input, address resolver, address addr) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactorSession) ProveAndClaimWithResolver(name []byte, input []DNSSECRRSetWithSignature, resolver common.Address, addr common.Address) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.ProveAndClaimWithResolver(&_ENSDNSRegistrar.TransactOpts, name, input, resolver, addr)
}

// SetPublicSuffixList is a paid mutator transaction binding the contract method 0x1ecfc411.
//
// Solidity: function setPublicSuffixList(address _suffixes) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactor) SetPublicSuffixList(opts *bind.TransactOpts, _suffixes common.Address) (*types.Transaction, error) {
	return _ENSDNSRegistrar.contract.Transact(opts, "setPublicSuffixList", _suffixes)
}

// SetPublicSuffixList is a paid mutator transaction binding the contract method 0x1ecfc411.
//
// Solidity: function setPublicSuffixList(address _suffixes) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarSession) SetPublicSuffixList(_suffixes common.Address) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.SetPublicSuffixList(&_ENSDNSRegistrar.TransactOpts, _suffixes)
}

// SetPublicSuffixList is a paid mutator transaction binding the contract method 0x1ecfc411.
//
// Solidity: function setPublicSuffixList(address _suffixes) returns()
func (_ENSDNSRegistrar *ENSDNSRegistrarTransactorSession) SetPublicSuffixList(_suffixes common.Address) (*types.Transaction, error) {
	return _ENSDNSRegistrar.Contract.SetPublicSuffixList(&_ENSDNSRegistrar.TransactOpts, _suffixes)
}

// ENSDNSRegistrarClaimIterator is returned from FilterClaim and is used to iterate over the raw logs and unpacked data for Claim events raised by the ENSDNSRegistrar contract.
type ENSDNSRegistrarClaimIterator struct {
	Event *ENSDNSRegistrarClaim // Event containing the contract specifics and raw log

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
func (it *ENSDNSRegistrarClaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSDNSRegistrarClaim)
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
		it.Event = new(ENSDNSRegistrarClaim)
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
func (it *ENSDNSRegistrarClaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSDNSRegistrarClaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSDNSRegistrarClaim represents a Claim event raised by the ENSDNSRegistrar contract.
type ENSDNSRegistrarClaim struct {
	Node      [32]byte
	Owner     common.Address
	Dnsname   []byte
	Inception uint32
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterClaim is a free log retrieval operation binding the contract event 0x87db02a0e483e2818060eddcbb3488ce44e35aff49a70d92c2aa6c8046cf01e2.
//
// Solidity: event Claim(bytes32 indexed node, address indexed owner, bytes dnsname, uint32 inception)
func (_ENSDNSRegistrar *ENSDNSRegistrarFilterer) FilterClaim(opts *bind.FilterOpts, node [][32]byte, owner []common.Address) (*ENSDNSRegistrarClaimIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _ENSDNSRegistrar.contract.FilterLogs(opts, "Claim", nodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &ENSDNSRegistrarClaimIterator{contract: _ENSDNSRegistrar.contract, event: "Claim", logs: logs, sub: sub}, nil
}

// WatchClaim is a free log subscription operation binding the contract event 0x87db02a0e483e2818060eddcbb3488ce44e35aff49a70d92c2aa6c8046cf01e2.
//
// Solidity: event Claim(bytes32 indexed node, address indexed owner, bytes dnsname, uint32 inception)
func (_ENSDNSRegistrar *ENSDNSRegistrarFilterer) WatchClaim(opts *bind.WatchOpts, sink chan<- *ENSDNSRegistrarClaim, node [][32]byte, owner []common.Address) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _ENSDNSRegistrar.contract.WatchLogs(opts, "Claim", nodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSDNSRegistrarClaim)
				if err := _ENSDNSRegistrar.contract.UnpackLog(event, "Claim", log); err != nil {
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

// ParseClaim is a log parse operation binding the contract event 0x87db02a0e483e2818060eddcbb3488ce44e35aff49a70d92c2aa6c8046cf01e2.
//
// Solidity: event Claim(bytes32 indexed node, address indexed owner, bytes dnsname, uint32 inception)
func (_ENSDNSRegistrar *ENSDNSRegistrarFilterer) ParseClaim(log types.Log) (*ENSDNSRegistrarClaim, error) {
	event := new(ENSDNSRegistrarClaim)
	if err := _ENSDNSRegistrar.contract.UnpackLog(event, "Claim", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSDNSRegistrarNewPublicSuffixListIterator is returned from FilterNewPublicSuffixList and is used to iterate over the raw logs and unpacked data for NewPublicSuffixList events raised by the ENSDNSRegistrar contract.
type ENSDNSRegistrarNewPublicSuffixListIterator struct {
	Event *ENSDNSRegistrarNewPublicSuffixList // Event containing the contract specifics and raw log

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
func (it *ENSDNSRegistrarNewPublicSuffixListIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSDNSRegistrarNewPublicSuffixList)
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
		it.Event = new(ENSDNSRegistrarNewPublicSuffixList)
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
func (it *ENSDNSRegistrarNewPublicSuffixListIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSDNSRegistrarNewPublicSuffixListIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSDNSRegistrarNewPublicSuffixList represents a NewPublicSuffixList event raised by the ENSDNSRegistrar contract.
type ENSDNSRegistrarNewPublicSuffixList struct {
	Suffixes common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNewPublicSuffixList is a free log retrieval operation binding the contract event 0x9176b7f47e4504df5e5516c99d90d82ac7cbd49cc77e7f22ba2ac2f2e3a3eba8.
//
// Solidity: event NewPublicSuffixList(address suffixes)
func (_ENSDNSRegistrar *ENSDNSRegistrarFilterer) FilterNewPublicSuffixList(opts *bind.FilterOpts) (*ENSDNSRegistrarNewPublicSuffixListIterator, error) {

	logs, sub, err := _ENSDNSRegistrar.contract.FilterLogs(opts, "NewPublicSuffixList")
	if err != nil {
		return nil, err
	}
	return &ENSDNSRegistrarNewPublicSuffixListIterator{contract: _ENSDNSRegistrar.contract, event: "NewPublicSuffixList", logs: logs, sub: sub}, nil
}

// WatchNewPublicSuffixList is a free log subscription operation binding the contract event 0x9176b7f47e4504df5e5516c99d90d82ac7cbd49cc77e7f22ba2ac2f2e3a3eba8.
//
// Solidity: event NewPublicSuffixList(address suffixes)
func (_ENSDNSRegistrar *ENSDNSRegistrarFilterer) WatchNewPublicSuffixList(opts *bind.WatchOpts, sink chan<- *ENSDNSRegistrarNewPublicSuffixList) (event.Subscription, error) {

	logs, sub, err := _ENSDNSRegistrar.contract.WatchLogs(opts, "NewPublicSuffixList")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSDNSRegistrarNewPublicSuffixList)
				if err := _ENSDNSRegistrar.contract.UnpackLog(event, "NewPublicSuffixList", log); err != nil {
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

// ParseNewPublicSuffixList is a log parse operation binding the contract event 0x9176b7f47e4504df5e5516c99d90d82ac7cbd49cc77e7f22ba2ac2f2e3a3eba8.
//
// Solidity: event NewPublicSuffixList(address suffixes)
func (_ENSDNSRegistrar *ENSDNSRegistrarFilterer) ParseNewPublicSuffixList(log types.Log) (*ENSDNSRegistrarNewPublicSuffixList, error) {
	event := new(ENSDNSRegistrarNewPublicSuffixList)
	if err := _ENSDNSRegistrar.contract.UnpackLog(event, "NewPublicSuffixList", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
