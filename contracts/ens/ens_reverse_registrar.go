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

// ENSReverseRegistrarMetaData contains all meta data concerning the ENSReverseRegistrar contract.
var ENSReverseRegistrarMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractENS\",\"name\":\"ensAddr\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"}],\"name\":\"ControllerChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractNameResolver\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"DefaultResolverChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"ReverseClaimed\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"claim\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"claimForAddr\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"claimWithResolver\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"controllers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"defaultResolver\",\"outputs\":[{\"internalType\":\"contractNameResolver\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ens\",\"outputs\":[{\"internalType\":\"contractENS\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"node\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"}],\"name\":\"setController\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"setDefaultResolver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"setName\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"setNameForAddr\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ENSReverseRegistrarABI is the input ABI used to generate the binding from.
// Deprecated: Use ENSReverseRegistrarMetaData.ABI instead.
var ENSReverseRegistrarABI = ENSReverseRegistrarMetaData.ABI

// ENSReverseRegistrar is an auto generated Go binding around an Ethereum contract.
type ENSReverseRegistrar struct {
	ENSReverseRegistrarCaller     // Read-only binding to the contract
	ENSReverseRegistrarTransactor // Write-only binding to the contract
	ENSReverseRegistrarFilterer   // Log filterer for contract events
}

// ENSReverseRegistrarCaller is an auto generated read-only Go binding around an Ethereum contract.
type ENSReverseRegistrarCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSReverseRegistrarTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ENSReverseRegistrarTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSReverseRegistrarFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ENSReverseRegistrarFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSReverseRegistrarSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ENSReverseRegistrarSession struct {
	Contract     *ENSReverseRegistrar // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ENSReverseRegistrarCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ENSReverseRegistrarCallerSession struct {
	Contract *ENSReverseRegistrarCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// ENSReverseRegistrarTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ENSReverseRegistrarTransactorSession struct {
	Contract     *ENSReverseRegistrarTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// ENSReverseRegistrarRaw is an auto generated low-level Go binding around an Ethereum contract.
type ENSReverseRegistrarRaw struct {
	Contract *ENSReverseRegistrar // Generic contract binding to access the raw methods on
}

// ENSReverseRegistrarCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ENSReverseRegistrarCallerRaw struct {
	Contract *ENSReverseRegistrarCaller // Generic read-only contract binding to access the raw methods on
}

// ENSReverseRegistrarTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ENSReverseRegistrarTransactorRaw struct {
	Contract *ENSReverseRegistrarTransactor // Generic write-only contract binding to access the raw methods on
}

// NewENSReverseRegistrar creates a new instance of ENSReverseRegistrar, bound to a specific deployed contract.
func NewENSReverseRegistrar(address common.Address, backend bind.ContractBackend) (*ENSReverseRegistrar, error) {
	contract, err := bindENSReverseRegistrar(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ENSReverseRegistrar{ENSReverseRegistrarCaller: ENSReverseRegistrarCaller{contract: contract}, ENSReverseRegistrarTransactor: ENSReverseRegistrarTransactor{contract: contract}, ENSReverseRegistrarFilterer: ENSReverseRegistrarFilterer{contract: contract}}, nil
}

// NewENSReverseRegistrarCaller creates a new read-only instance of ENSReverseRegistrar, bound to a specific deployed contract.
func NewENSReverseRegistrarCaller(address common.Address, caller bind.ContractCaller) (*ENSReverseRegistrarCaller, error) {
	contract, err := bindENSReverseRegistrar(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ENSReverseRegistrarCaller{contract: contract}, nil
}

// NewENSReverseRegistrarTransactor creates a new write-only instance of ENSReverseRegistrar, bound to a specific deployed contract.
func NewENSReverseRegistrarTransactor(address common.Address, transactor bind.ContractTransactor) (*ENSReverseRegistrarTransactor, error) {
	contract, err := bindENSReverseRegistrar(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ENSReverseRegistrarTransactor{contract: contract}, nil
}

// NewENSReverseRegistrarFilterer creates a new log filterer instance of ENSReverseRegistrar, bound to a specific deployed contract.
func NewENSReverseRegistrarFilterer(address common.Address, filterer bind.ContractFilterer) (*ENSReverseRegistrarFilterer, error) {
	contract, err := bindENSReverseRegistrar(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ENSReverseRegistrarFilterer{contract: contract}, nil
}

// bindENSReverseRegistrar binds a generic wrapper to an already deployed contract.
func bindENSReverseRegistrar(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ENSReverseRegistrarMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ENSReverseRegistrar *ENSReverseRegistrarRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ENSReverseRegistrar.Contract.ENSReverseRegistrarCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ENSReverseRegistrar *ENSReverseRegistrarRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.ENSReverseRegistrarTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ENSReverseRegistrar *ENSReverseRegistrarRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.ENSReverseRegistrarTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ENSReverseRegistrar *ENSReverseRegistrarCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ENSReverseRegistrar.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.contract.Transact(opts, method, params...)
}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_ENSReverseRegistrar *ENSReverseRegistrarCaller) Controllers(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ENSReverseRegistrar.contract.Call(opts, &out, "controllers", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) Controllers(arg0 common.Address) (bool, error) {
	return _ENSReverseRegistrar.Contract.Controllers(&_ENSReverseRegistrar.CallOpts, arg0)
}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_ENSReverseRegistrar *ENSReverseRegistrarCallerSession) Controllers(arg0 common.Address) (bool, error) {
	return _ENSReverseRegistrar.Contract.Controllers(&_ENSReverseRegistrar.CallOpts, arg0)
}

// DefaultResolver is a free data retrieval call binding the contract method 0x828eab0e.
//
// Solidity: function defaultResolver() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarCaller) DefaultResolver(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSReverseRegistrar.contract.Call(opts, &out, "defaultResolver")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DefaultResolver is a free data retrieval call binding the contract method 0x828eab0e.
//
// Solidity: function defaultResolver() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) DefaultResolver() (common.Address, error) {
	return _ENSReverseRegistrar.Contract.DefaultResolver(&_ENSReverseRegistrar.CallOpts)
}

// DefaultResolver is a free data retrieval call binding the contract method 0x828eab0e.
//
// Solidity: function defaultResolver() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarCallerSession) DefaultResolver() (common.Address, error) {
	return _ENSReverseRegistrar.Contract.DefaultResolver(&_ENSReverseRegistrar.CallOpts)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarCaller) Ens(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSReverseRegistrar.contract.Call(opts, &out, "ens")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) Ens() (common.Address, error) {
	return _ENSReverseRegistrar.Contract.Ens(&_ENSReverseRegistrar.CallOpts)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarCallerSession) Ens() (common.Address, error) {
	return _ENSReverseRegistrar.Contract.Ens(&_ENSReverseRegistrar.CallOpts)
}

// Node is a free data retrieval call binding the contract method 0xbffbe61c.
//
// Solidity: function node(address addr) pure returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarCaller) Node(opts *bind.CallOpts, addr common.Address) ([32]byte, error) {
	var out []interface{}
	err := _ENSReverseRegistrar.contract.Call(opts, &out, "node", addr)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Node is a free data retrieval call binding the contract method 0xbffbe61c.
//
// Solidity: function node(address addr) pure returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) Node(addr common.Address) ([32]byte, error) {
	return _ENSReverseRegistrar.Contract.Node(&_ENSReverseRegistrar.CallOpts, addr)
}

// Node is a free data retrieval call binding the contract method 0xbffbe61c.
//
// Solidity: function node(address addr) pure returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarCallerSession) Node(addr common.Address) ([32]byte, error) {
	return _ENSReverseRegistrar.Contract.Node(&_ENSReverseRegistrar.CallOpts, addr)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSReverseRegistrar.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) Owner() (common.Address, error) {
	return _ENSReverseRegistrar.Contract.Owner(&_ENSReverseRegistrar.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSReverseRegistrar *ENSReverseRegistrarCallerSession) Owner() (common.Address, error) {
	return _ENSReverseRegistrar.Contract.Owner(&_ENSReverseRegistrar.CallOpts)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address owner) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) Claim(opts *bind.TransactOpts, owner common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "claim", owner)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address owner) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) Claim(owner common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.Claim(&_ENSReverseRegistrar.TransactOpts, owner)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address owner) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) Claim(owner common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.Claim(&_ENSReverseRegistrar.TransactOpts, owner)
}

// ClaimForAddr is a paid mutator transaction binding the contract method 0x65669631.
//
// Solidity: function claimForAddr(address addr, address owner, address resolver) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) ClaimForAddr(opts *bind.TransactOpts, addr common.Address, owner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "claimForAddr", addr, owner, resolver)
}

// ClaimForAddr is a paid mutator transaction binding the contract method 0x65669631.
//
// Solidity: function claimForAddr(address addr, address owner, address resolver) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) ClaimForAddr(addr common.Address, owner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.ClaimForAddr(&_ENSReverseRegistrar.TransactOpts, addr, owner, resolver)
}

// ClaimForAddr is a paid mutator transaction binding the contract method 0x65669631.
//
// Solidity: function claimForAddr(address addr, address owner, address resolver) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) ClaimForAddr(addr common.Address, owner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.ClaimForAddr(&_ENSReverseRegistrar.TransactOpts, addr, owner, resolver)
}

// ClaimWithResolver is a paid mutator transaction binding the contract method 0x0f5a5466.
//
// Solidity: function claimWithResolver(address owner, address resolver) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) ClaimWithResolver(opts *bind.TransactOpts, owner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "claimWithResolver", owner, resolver)
}

// ClaimWithResolver is a paid mutator transaction binding the contract method 0x0f5a5466.
//
// Solidity: function claimWithResolver(address owner, address resolver) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) ClaimWithResolver(owner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.ClaimWithResolver(&_ENSReverseRegistrar.TransactOpts, owner, resolver)
}

// ClaimWithResolver is a paid mutator transaction binding the contract method 0x0f5a5466.
//
// Solidity: function claimWithResolver(address owner, address resolver) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) ClaimWithResolver(owner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.ClaimWithResolver(&_ENSReverseRegistrar.TransactOpts, owner, resolver)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) RenounceOwnership() (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.RenounceOwnership(&_ENSReverseRegistrar.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.RenounceOwnership(&_ENSReverseRegistrar.TransactOpts)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool enabled) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) SetController(opts *bind.TransactOpts, controller common.Address, enabled bool) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "setController", controller, enabled)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool enabled) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) SetController(controller common.Address, enabled bool) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.SetController(&_ENSReverseRegistrar.TransactOpts, controller, enabled)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool enabled) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) SetController(controller common.Address, enabled bool) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.SetController(&_ENSReverseRegistrar.TransactOpts, controller, enabled)
}

// SetDefaultResolver is a paid mutator transaction binding the contract method 0xc66485b2.
//
// Solidity: function setDefaultResolver(address resolver) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) SetDefaultResolver(opts *bind.TransactOpts, resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "setDefaultResolver", resolver)
}

// SetDefaultResolver is a paid mutator transaction binding the contract method 0xc66485b2.
//
// Solidity: function setDefaultResolver(address resolver) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) SetDefaultResolver(resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.SetDefaultResolver(&_ENSReverseRegistrar.TransactOpts, resolver)
}

// SetDefaultResolver is a paid mutator transaction binding the contract method 0xc66485b2.
//
// Solidity: function setDefaultResolver(address resolver) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) SetDefaultResolver(resolver common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.SetDefaultResolver(&_ENSReverseRegistrar.TransactOpts, resolver)
}

// SetName is a paid mutator transaction binding the contract method 0xc47f0027.
//
// Solidity: function setName(string name) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) SetName(opts *bind.TransactOpts, name string) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "setName", name)
}

// SetName is a paid mutator transaction binding the contract method 0xc47f0027.
//
// Solidity: function setName(string name) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) SetName(name string) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.SetName(&_ENSReverseRegistrar.TransactOpts, name)
}

// SetName is a paid mutator transaction binding the contract method 0xc47f0027.
//
// Solidity: function setName(string name) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) SetName(name string) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.SetName(&_ENSReverseRegistrar.TransactOpts, name)
}

// SetNameForAddr is a paid mutator transaction binding the contract method 0x7a806d6b.
//
// Solidity: function setNameForAddr(address addr, address owner, address resolver, string name) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) SetNameForAddr(opts *bind.TransactOpts, addr common.Address, owner common.Address, resolver common.Address, name string) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "setNameForAddr", addr, owner, resolver, name)
}

// SetNameForAddr is a paid mutator transaction binding the contract method 0x7a806d6b.
//
// Solidity: function setNameForAddr(address addr, address owner, address resolver, string name) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) SetNameForAddr(addr common.Address, owner common.Address, resolver common.Address, name string) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.SetNameForAddr(&_ENSReverseRegistrar.TransactOpts, addr, owner, resolver, name)
}

// SetNameForAddr is a paid mutator transaction binding the contract method 0x7a806d6b.
//
// Solidity: function setNameForAddr(address addr, address owner, address resolver, string name) returns(bytes32)
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) SetNameForAddr(addr common.Address, owner common.Address, resolver common.Address, name string) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.SetNameForAddr(&_ENSReverseRegistrar.TransactOpts, addr, owner, resolver, name)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.TransferOwnership(&_ENSReverseRegistrar.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSReverseRegistrar *ENSReverseRegistrarTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ENSReverseRegistrar.Contract.TransferOwnership(&_ENSReverseRegistrar.TransactOpts, newOwner)
}

// ENSReverseRegistrarControllerChangedIterator is returned from FilterControllerChanged and is used to iterate over the raw logs and unpacked data for ControllerChanged events raised by the ENSReverseRegistrar contract.
type ENSReverseRegistrarControllerChangedIterator struct {
	Event *ENSReverseRegistrarControllerChanged // Event containing the contract specifics and raw log

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
func (it *ENSReverseRegistrarControllerChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSReverseRegistrarControllerChanged)
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
		it.Event = new(ENSReverseRegistrarControllerChanged)
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
func (it *ENSReverseRegistrarControllerChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSReverseRegistrarControllerChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSReverseRegistrarControllerChanged represents a ControllerChanged event raised by the ENSReverseRegistrar contract.
type ENSReverseRegistrarControllerChanged struct {
	Controller common.Address
	Enabled    bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterControllerChanged is a free log retrieval operation binding the contract event 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87.
//
// Solidity: event ControllerChanged(address indexed controller, bool enabled)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) FilterControllerChanged(opts *bind.FilterOpts, controller []common.Address) (*ENSReverseRegistrarControllerChangedIterator, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _ENSReverseRegistrar.contract.FilterLogs(opts, "ControllerChanged", controllerRule)
	if err != nil {
		return nil, err
	}
	return &ENSReverseRegistrarControllerChangedIterator{contract: _ENSReverseRegistrar.contract, event: "ControllerChanged", logs: logs, sub: sub}, nil
}

// WatchControllerChanged is a free log subscription operation binding the contract event 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87.
//
// Solidity: event ControllerChanged(address indexed controller, bool enabled)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) WatchControllerChanged(opts *bind.WatchOpts, sink chan<- *ENSReverseRegistrarControllerChanged, controller []common.Address) (event.Subscription, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _ENSReverseRegistrar.contract.WatchLogs(opts, "ControllerChanged", controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSReverseRegistrarControllerChanged)
				if err := _ENSReverseRegistrar.contract.UnpackLog(event, "ControllerChanged", log); err != nil {
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
// Solidity: event ControllerChanged(address indexed controller, bool enabled)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) ParseControllerChanged(log types.Log) (*ENSReverseRegistrarControllerChanged, error) {
	event := new(ENSReverseRegistrarControllerChanged)
	if err := _ENSReverseRegistrar.contract.UnpackLog(event, "ControllerChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSReverseRegistrarDefaultResolverChangedIterator is returned from FilterDefaultResolverChanged and is used to iterate over the raw logs and unpacked data for DefaultResolverChanged events raised by the ENSReverseRegistrar contract.
type ENSReverseRegistrarDefaultResolverChangedIterator struct {
	Event *ENSReverseRegistrarDefaultResolverChanged // Event containing the contract specifics and raw log

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
func (it *ENSReverseRegistrarDefaultResolverChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSReverseRegistrarDefaultResolverChanged)
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
		it.Event = new(ENSReverseRegistrarDefaultResolverChanged)
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
func (it *ENSReverseRegistrarDefaultResolverChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSReverseRegistrarDefaultResolverChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSReverseRegistrarDefaultResolverChanged represents a DefaultResolverChanged event raised by the ENSReverseRegistrar contract.
type ENSReverseRegistrarDefaultResolverChanged struct {
	Resolver common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDefaultResolverChanged is a free log retrieval operation binding the contract event 0xeae17a84d9eb83d8c8eb317f9e7d64857bc363fa51674d996c023f4340c577cf.
//
// Solidity: event DefaultResolverChanged(address indexed resolver)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) FilterDefaultResolverChanged(opts *bind.FilterOpts, resolver []common.Address) (*ENSReverseRegistrarDefaultResolverChangedIterator, error) {

	var resolverRule []interface{}
	for _, resolverItem := range resolver {
		resolverRule = append(resolverRule, resolverItem)
	}

	logs, sub, err := _ENSReverseRegistrar.contract.FilterLogs(opts, "DefaultResolverChanged", resolverRule)
	if err != nil {
		return nil, err
	}
	return &ENSReverseRegistrarDefaultResolverChangedIterator{contract: _ENSReverseRegistrar.contract, event: "DefaultResolverChanged", logs: logs, sub: sub}, nil
}

// WatchDefaultResolverChanged is a free log subscription operation binding the contract event 0xeae17a84d9eb83d8c8eb317f9e7d64857bc363fa51674d996c023f4340c577cf.
//
// Solidity: event DefaultResolverChanged(address indexed resolver)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) WatchDefaultResolverChanged(opts *bind.WatchOpts, sink chan<- *ENSReverseRegistrarDefaultResolverChanged, resolver []common.Address) (event.Subscription, error) {

	var resolverRule []interface{}
	for _, resolverItem := range resolver {
		resolverRule = append(resolverRule, resolverItem)
	}

	logs, sub, err := _ENSReverseRegistrar.contract.WatchLogs(opts, "DefaultResolverChanged", resolverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSReverseRegistrarDefaultResolverChanged)
				if err := _ENSReverseRegistrar.contract.UnpackLog(event, "DefaultResolverChanged", log); err != nil {
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

// ParseDefaultResolverChanged is a log parse operation binding the contract event 0xeae17a84d9eb83d8c8eb317f9e7d64857bc363fa51674d996c023f4340c577cf.
//
// Solidity: event DefaultResolverChanged(address indexed resolver)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) ParseDefaultResolverChanged(log types.Log) (*ENSReverseRegistrarDefaultResolverChanged, error) {
	event := new(ENSReverseRegistrarDefaultResolverChanged)
	if err := _ENSReverseRegistrar.contract.UnpackLog(event, "DefaultResolverChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSReverseRegistrarOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ENSReverseRegistrar contract.
type ENSReverseRegistrarOwnershipTransferredIterator struct {
	Event *ENSReverseRegistrarOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ENSReverseRegistrarOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSReverseRegistrarOwnershipTransferred)
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
		it.Event = new(ENSReverseRegistrarOwnershipTransferred)
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
func (it *ENSReverseRegistrarOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSReverseRegistrarOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSReverseRegistrarOwnershipTransferred represents a OwnershipTransferred event raised by the ENSReverseRegistrar contract.
type ENSReverseRegistrarOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ENSReverseRegistrarOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ENSReverseRegistrar.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ENSReverseRegistrarOwnershipTransferredIterator{contract: _ENSReverseRegistrar.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ENSReverseRegistrarOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ENSReverseRegistrar.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSReverseRegistrarOwnershipTransferred)
				if err := _ENSReverseRegistrar.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) ParseOwnershipTransferred(log types.Log) (*ENSReverseRegistrarOwnershipTransferred, error) {
	event := new(ENSReverseRegistrarOwnershipTransferred)
	if err := _ENSReverseRegistrar.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSReverseRegistrarReverseClaimedIterator is returned from FilterReverseClaimed and is used to iterate over the raw logs and unpacked data for ReverseClaimed events raised by the ENSReverseRegistrar contract.
type ENSReverseRegistrarReverseClaimedIterator struct {
	Event *ENSReverseRegistrarReverseClaimed // Event containing the contract specifics and raw log

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
func (it *ENSReverseRegistrarReverseClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSReverseRegistrarReverseClaimed)
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
		it.Event = new(ENSReverseRegistrarReverseClaimed)
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
func (it *ENSReverseRegistrarReverseClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSReverseRegistrarReverseClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSReverseRegistrarReverseClaimed represents a ReverseClaimed event raised by the ENSReverseRegistrar contract.
type ENSReverseRegistrarReverseClaimed struct {
	Addr common.Address
	Node [32]byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterReverseClaimed is a free log retrieval operation binding the contract event 0x6ada868dd3058cf77a48a74489fd7963688e5464b2b0fa957ace976243270e92.
//
// Solidity: event ReverseClaimed(address indexed addr, bytes32 indexed node)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) FilterReverseClaimed(opts *bind.FilterOpts, addr []common.Address, node [][32]byte) (*ENSReverseRegistrarReverseClaimedIterator, error) {

	var addrRule []interface{}
	for _, addrItem := range addr {
		addrRule = append(addrRule, addrItem)
	}
	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSReverseRegistrar.contract.FilterLogs(opts, "ReverseClaimed", addrRule, nodeRule)
	if err != nil {
		return nil, err
	}
	return &ENSReverseRegistrarReverseClaimedIterator{contract: _ENSReverseRegistrar.contract, event: "ReverseClaimed", logs: logs, sub: sub}, nil
}

// WatchReverseClaimed is a free log subscription operation binding the contract event 0x6ada868dd3058cf77a48a74489fd7963688e5464b2b0fa957ace976243270e92.
//
// Solidity: event ReverseClaimed(address indexed addr, bytes32 indexed node)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) WatchReverseClaimed(opts *bind.WatchOpts, sink chan<- *ENSReverseRegistrarReverseClaimed, addr []common.Address, node [][32]byte) (event.Subscription, error) {

	var addrRule []interface{}
	for _, addrItem := range addr {
		addrRule = append(addrRule, addrItem)
	}
	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ENSReverseRegistrar.contract.WatchLogs(opts, "ReverseClaimed", addrRule, nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSReverseRegistrarReverseClaimed)
				if err := _ENSReverseRegistrar.contract.UnpackLog(event, "ReverseClaimed", log); err != nil {
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

// ParseReverseClaimed is a log parse operation binding the contract event 0x6ada868dd3058cf77a48a74489fd7963688e5464b2b0fa957ace976243270e92.
//
// Solidity: event ReverseClaimed(address indexed addr, bytes32 indexed node)
func (_ENSReverseRegistrar *ENSReverseRegistrarFilterer) ParseReverseClaimed(log types.Log) (*ENSReverseRegistrarReverseClaimed, error) {
	event := new(ENSReverseRegistrarReverseClaimed)
	if err := _ENSReverseRegistrar.contract.UnpackLog(event, "ReverseClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
