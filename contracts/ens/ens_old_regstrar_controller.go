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

// ENSOldRegistrarControllerMetaData contains all meta data concerning the ENSOldRegistrarController contract.
var ENSOldRegistrarControllerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractBaseRegistrar\",\"name\":\"_base\",\"type\":\"address\"},{\"internalType\":\"contractPriceOracle\",\"name\":\"_prices\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_minCommitmentAge\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_maxCommitmentAge\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"label\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cost\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"NameRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"label\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cost\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"NameRenewed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oracle\",\"type\":\"address\"}],\"name\":\"NewPriceOracle\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[],\"name\":\"MIN_REGISTRATION_DURATION\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"available\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"}],\"name\":\"commit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"commitments\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"}],\"name\":\"makeCommitment\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"makeCommitmentWithConfig\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxCommitmentAge\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minCommitmentAge\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"registerWithConfig\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"renew\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"rentPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minCommitmentAge\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_maxCommitmentAge\",\"type\":\"uint256\"}],\"name\":\"setCommitmentAges\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"contractPriceOracle\",\"name\":\"_prices\",\"type\":\"address\"}],\"name\":\"setPriceOracle\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceID\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"valid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ENSOldRegistrarControllerABI is the input ABI used to generate the binding from.
// Deprecated: Use ENSOldRegistrarControllerMetaData.ABI instead.
var ENSOldRegistrarControllerABI = ENSOldRegistrarControllerMetaData.ABI

// ENSOldRegistrarController is an auto generated Go binding around an Ethereum contract.
type ENSOldRegistrarController struct {
	ENSOldRegistrarControllerCaller     // Read-only binding to the contract
	ENSOldRegistrarControllerTransactor // Write-only binding to the contract
	ENSOldRegistrarControllerFilterer   // Log filterer for contract events
}

// ENSOldRegistrarControllerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ENSOldRegistrarControllerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSOldRegistrarControllerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ENSOldRegistrarControllerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSOldRegistrarControllerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ENSOldRegistrarControllerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ENSOldRegistrarControllerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ENSOldRegistrarControllerSession struct {
	Contract     *ENSOldRegistrarController // Generic contract binding to set the session for
	CallOpts     bind.CallOpts              // Call options to use throughout this session
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ENSOldRegistrarControllerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ENSOldRegistrarControllerCallerSession struct {
	Contract *ENSOldRegistrarControllerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                    // Call options to use throughout this session
}

// ENSOldRegistrarControllerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ENSOldRegistrarControllerTransactorSession struct {
	Contract     *ENSOldRegistrarControllerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                    // Transaction auth options to use throughout this session
}

// ENSOldRegistrarControllerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ENSOldRegistrarControllerRaw struct {
	Contract *ENSOldRegistrarController // Generic contract binding to access the raw methods on
}

// ENSOldRegistrarControllerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ENSOldRegistrarControllerCallerRaw struct {
	Contract *ENSOldRegistrarControllerCaller // Generic read-only contract binding to access the raw methods on
}

// ENSOldRegistrarControllerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ENSOldRegistrarControllerTransactorRaw struct {
	Contract *ENSOldRegistrarControllerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewENSOldRegistrarController creates a new instance of ENSOldRegistrarController, bound to a specific deployed contract.
func NewENSOldRegistrarController(address common.Address, backend bind.ContractBackend) (*ENSOldRegistrarController, error) {
	contract, err := bindENSOldRegistrarController(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ENSOldRegistrarController{ENSOldRegistrarControllerCaller: ENSOldRegistrarControllerCaller{contract: contract}, ENSOldRegistrarControllerTransactor: ENSOldRegistrarControllerTransactor{contract: contract}, ENSOldRegistrarControllerFilterer: ENSOldRegistrarControllerFilterer{contract: contract}}, nil
}

// NewENSOldRegistrarControllerCaller creates a new read-only instance of ENSOldRegistrarController, bound to a specific deployed contract.
func NewENSOldRegistrarControllerCaller(address common.Address, caller bind.ContractCaller) (*ENSOldRegistrarControllerCaller, error) {
	contract, err := bindENSOldRegistrarController(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ENSOldRegistrarControllerCaller{contract: contract}, nil
}

// NewENSOldRegistrarControllerTransactor creates a new write-only instance of ENSOldRegistrarController, bound to a specific deployed contract.
func NewENSOldRegistrarControllerTransactor(address common.Address, transactor bind.ContractTransactor) (*ENSOldRegistrarControllerTransactor, error) {
	contract, err := bindENSOldRegistrarController(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ENSOldRegistrarControllerTransactor{contract: contract}, nil
}

// NewENSOldRegistrarControllerFilterer creates a new log filterer instance of ENSOldRegistrarController, bound to a specific deployed contract.
func NewENSOldRegistrarControllerFilterer(address common.Address, filterer bind.ContractFilterer) (*ENSOldRegistrarControllerFilterer, error) {
	contract, err := bindENSOldRegistrarController(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ENSOldRegistrarControllerFilterer{contract: contract}, nil
}

// bindENSOldRegistrarController binds a generic wrapper to an already deployed contract.
func bindENSOldRegistrarController(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ENSOldRegistrarControllerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ENSOldRegistrarController *ENSOldRegistrarControllerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ENSOldRegistrarController.Contract.ENSOldRegistrarControllerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ENSOldRegistrarController *ENSOldRegistrarControllerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.ENSOldRegistrarControllerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ENSOldRegistrarController *ENSOldRegistrarControllerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.ENSOldRegistrarControllerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ENSOldRegistrarController.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.contract.Transact(opts, method, params...)
}

// MINREGISTRATIONDURATION is a free data retrieval call binding the contract method 0x8a95b09f.
//
// Solidity: function MIN_REGISTRATION_DURATION() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) MINREGISTRATIONDURATION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "MIN_REGISTRATION_DURATION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINREGISTRATIONDURATION is a free data retrieval call binding the contract method 0x8a95b09f.
//
// Solidity: function MIN_REGISTRATION_DURATION() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) MINREGISTRATIONDURATION() (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.MINREGISTRATIONDURATION(&_ENSOldRegistrarController.CallOpts)
}

// MINREGISTRATIONDURATION is a free data retrieval call binding the contract method 0x8a95b09f.
//
// Solidity: function MIN_REGISTRATION_DURATION() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) MINREGISTRATIONDURATION() (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.MINREGISTRATIONDURATION(&_ENSOldRegistrarController.CallOpts)
}

// Available is a free data retrieval call binding the contract method 0xaeb8ce9b.
//
// Solidity: function available(string name) view returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) Available(opts *bind.CallOpts, name string) (bool, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "available", name)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Available is a free data retrieval call binding the contract method 0xaeb8ce9b.
//
// Solidity: function available(string name) view returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) Available(name string) (bool, error) {
	return _ENSOldRegistrarController.Contract.Available(&_ENSOldRegistrarController.CallOpts, name)
}

// Available is a free data retrieval call binding the contract method 0xaeb8ce9b.
//
// Solidity: function available(string name) view returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) Available(name string) (bool, error) {
	return _ENSOldRegistrarController.Contract.Available(&_ENSOldRegistrarController.CallOpts, name)
}

// Commitments is a free data retrieval call binding the contract method 0x839df945.
//
// Solidity: function commitments(bytes32 ) view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) Commitments(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "commitments", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Commitments is a free data retrieval call binding the contract method 0x839df945.
//
// Solidity: function commitments(bytes32 ) view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) Commitments(arg0 [32]byte) (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.Commitments(&_ENSOldRegistrarController.CallOpts, arg0)
}

// Commitments is a free data retrieval call binding the contract method 0x839df945.
//
// Solidity: function commitments(bytes32 ) view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) Commitments(arg0 [32]byte) (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.Commitments(&_ENSOldRegistrarController.CallOpts, arg0)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) IsOwner(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "isOwner")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) IsOwner() (bool, error) {
	return _ENSOldRegistrarController.Contract.IsOwner(&_ENSOldRegistrarController.CallOpts)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) IsOwner() (bool, error) {
	return _ENSOldRegistrarController.Contract.IsOwner(&_ENSOldRegistrarController.CallOpts)
}

// MakeCommitment is a free data retrieval call binding the contract method 0xf49826be.
//
// Solidity: function makeCommitment(string name, address owner, bytes32 secret) pure returns(bytes32)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) MakeCommitment(opts *bind.CallOpts, name string, owner common.Address, secret [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "makeCommitment", name, owner, secret)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MakeCommitment is a free data retrieval call binding the contract method 0xf49826be.
//
// Solidity: function makeCommitment(string name, address owner, bytes32 secret) pure returns(bytes32)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) MakeCommitment(name string, owner common.Address, secret [32]byte) ([32]byte, error) {
	return _ENSOldRegistrarController.Contract.MakeCommitment(&_ENSOldRegistrarController.CallOpts, name, owner, secret)
}

// MakeCommitment is a free data retrieval call binding the contract method 0xf49826be.
//
// Solidity: function makeCommitment(string name, address owner, bytes32 secret) pure returns(bytes32)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) MakeCommitment(name string, owner common.Address, secret [32]byte) ([32]byte, error) {
	return _ENSOldRegistrarController.Contract.MakeCommitment(&_ENSOldRegistrarController.CallOpts, name, owner, secret)
}

// MakeCommitmentWithConfig is a free data retrieval call binding the contract method 0x3d86c52f.
//
// Solidity: function makeCommitmentWithConfig(string name, address owner, bytes32 secret, address resolver, address addr) pure returns(bytes32)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) MakeCommitmentWithConfig(opts *bind.CallOpts, name string, owner common.Address, secret [32]byte, resolver common.Address, addr common.Address) ([32]byte, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "makeCommitmentWithConfig", name, owner, secret, resolver, addr)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MakeCommitmentWithConfig is a free data retrieval call binding the contract method 0x3d86c52f.
//
// Solidity: function makeCommitmentWithConfig(string name, address owner, bytes32 secret, address resolver, address addr) pure returns(bytes32)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) MakeCommitmentWithConfig(name string, owner common.Address, secret [32]byte, resolver common.Address, addr common.Address) ([32]byte, error) {
	return _ENSOldRegistrarController.Contract.MakeCommitmentWithConfig(&_ENSOldRegistrarController.CallOpts, name, owner, secret, resolver, addr)
}

// MakeCommitmentWithConfig is a free data retrieval call binding the contract method 0x3d86c52f.
//
// Solidity: function makeCommitmentWithConfig(string name, address owner, bytes32 secret, address resolver, address addr) pure returns(bytes32)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) MakeCommitmentWithConfig(name string, owner common.Address, secret [32]byte, resolver common.Address, addr common.Address) ([32]byte, error) {
	return _ENSOldRegistrarController.Contract.MakeCommitmentWithConfig(&_ENSOldRegistrarController.CallOpts, name, owner, secret, resolver, addr)
}

// MaxCommitmentAge is a free data retrieval call binding the contract method 0xce1e09c0.
//
// Solidity: function maxCommitmentAge() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) MaxCommitmentAge(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "maxCommitmentAge")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxCommitmentAge is a free data retrieval call binding the contract method 0xce1e09c0.
//
// Solidity: function maxCommitmentAge() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) MaxCommitmentAge() (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.MaxCommitmentAge(&_ENSOldRegistrarController.CallOpts)
}

// MaxCommitmentAge is a free data retrieval call binding the contract method 0xce1e09c0.
//
// Solidity: function maxCommitmentAge() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) MaxCommitmentAge() (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.MaxCommitmentAge(&_ENSOldRegistrarController.CallOpts)
}

// MinCommitmentAge is a free data retrieval call binding the contract method 0x8d839ffe.
//
// Solidity: function minCommitmentAge() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) MinCommitmentAge(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "minCommitmentAge")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinCommitmentAge is a free data retrieval call binding the contract method 0x8d839ffe.
//
// Solidity: function minCommitmentAge() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) MinCommitmentAge() (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.MinCommitmentAge(&_ENSOldRegistrarController.CallOpts)
}

// MinCommitmentAge is a free data retrieval call binding the contract method 0x8d839ffe.
//
// Solidity: function minCommitmentAge() view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) MinCommitmentAge() (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.MinCommitmentAge(&_ENSOldRegistrarController.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) Owner() (common.Address, error) {
	return _ENSOldRegistrarController.Contract.Owner(&_ENSOldRegistrarController.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) Owner() (common.Address, error) {
	return _ENSOldRegistrarController.Contract.Owner(&_ENSOldRegistrarController.CallOpts)
}

// RentPrice is a free data retrieval call binding the contract method 0x83e7f6ff.
//
// Solidity: function rentPrice(string name, uint256 duration) view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) RentPrice(opts *bind.CallOpts, name string, duration *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "rentPrice", name, duration)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RentPrice is a free data retrieval call binding the contract method 0x83e7f6ff.
//
// Solidity: function rentPrice(string name, uint256 duration) view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) RentPrice(name string, duration *big.Int) (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.RentPrice(&_ENSOldRegistrarController.CallOpts, name, duration)
}

// RentPrice is a free data retrieval call binding the contract method 0x83e7f6ff.
//
// Solidity: function rentPrice(string name, uint256 duration) view returns(uint256)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) RentPrice(name string, duration *big.Int) (*big.Int, error) {
	return _ENSOldRegistrarController.Contract.RentPrice(&_ENSOldRegistrarController.CallOpts, name, duration)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) SupportsInterface(opts *bind.CallOpts, interfaceID [4]byte) (bool, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "supportsInterface", interfaceID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _ENSOldRegistrarController.Contract.SupportsInterface(&_ENSOldRegistrarController.CallOpts, interfaceID)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _ENSOldRegistrarController.Contract.SupportsInterface(&_ENSOldRegistrarController.CallOpts, interfaceID)
}

// Valid is a free data retrieval call binding the contract method 0x9791c097.
//
// Solidity: function valid(string name) pure returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCaller) Valid(opts *bind.CallOpts, name string) (bool, error) {
	var out []interface{}
	err := _ENSOldRegistrarController.contract.Call(opts, &out, "valid", name)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Valid is a free data retrieval call binding the contract method 0x9791c097.
//
// Solidity: function valid(string name) pure returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) Valid(name string) (bool, error) {
	return _ENSOldRegistrarController.Contract.Valid(&_ENSOldRegistrarController.CallOpts, name)
}

// Valid is a free data retrieval call binding the contract method 0x9791c097.
//
// Solidity: function valid(string name) pure returns(bool)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerCallerSession) Valid(name string) (bool, error) {
	return _ENSOldRegistrarController.Contract.Valid(&_ENSOldRegistrarController.CallOpts, name)
}

// Commit is a paid mutator transaction binding the contract method 0xf14fcbc8.
//
// Solidity: function commit(bytes32 commitment) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) Commit(opts *bind.TransactOpts, commitment [32]byte) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "commit", commitment)
}

// Commit is a paid mutator transaction binding the contract method 0xf14fcbc8.
//
// Solidity: function commit(bytes32 commitment) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) Commit(commitment [32]byte) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.Commit(&_ENSOldRegistrarController.TransactOpts, commitment)
}

// Commit is a paid mutator transaction binding the contract method 0xf14fcbc8.
//
// Solidity: function commit(bytes32 commitment) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) Commit(commitment [32]byte) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.Commit(&_ENSOldRegistrarController.TransactOpts, commitment)
}

// Register is a paid mutator transaction binding the contract method 0x85f6d155.
//
// Solidity: function register(string name, address owner, uint256 duration, bytes32 secret) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) Register(opts *bind.TransactOpts, name string, owner common.Address, duration *big.Int, secret [32]byte) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "register", name, owner, duration, secret)
}

// Register is a paid mutator transaction binding the contract method 0x85f6d155.
//
// Solidity: function register(string name, address owner, uint256 duration, bytes32 secret) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) Register(name string, owner common.Address, duration *big.Int, secret [32]byte) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.Register(&_ENSOldRegistrarController.TransactOpts, name, owner, duration, secret)
}

// Register is a paid mutator transaction binding the contract method 0x85f6d155.
//
// Solidity: function register(string name, address owner, uint256 duration, bytes32 secret) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) Register(name string, owner common.Address, duration *big.Int, secret [32]byte) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.Register(&_ENSOldRegistrarController.TransactOpts, name, owner, duration, secret)
}

// RegisterWithConfig is a paid mutator transaction binding the contract method 0xf7a16963.
//
// Solidity: function registerWithConfig(string name, address owner, uint256 duration, bytes32 secret, address resolver, address addr) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) RegisterWithConfig(opts *bind.TransactOpts, name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, addr common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "registerWithConfig", name, owner, duration, secret, resolver, addr)
}

// RegisterWithConfig is a paid mutator transaction binding the contract method 0xf7a16963.
//
// Solidity: function registerWithConfig(string name, address owner, uint256 duration, bytes32 secret, address resolver, address addr) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) RegisterWithConfig(name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, addr common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.RegisterWithConfig(&_ENSOldRegistrarController.TransactOpts, name, owner, duration, secret, resolver, addr)
}

// RegisterWithConfig is a paid mutator transaction binding the contract method 0xf7a16963.
//
// Solidity: function registerWithConfig(string name, address owner, uint256 duration, bytes32 secret, address resolver, address addr) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) RegisterWithConfig(name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, addr common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.RegisterWithConfig(&_ENSOldRegistrarController.TransactOpts, name, owner, duration, secret, resolver, addr)
}

// Renew is a paid mutator transaction binding the contract method 0xacf1a841.
//
// Solidity: function renew(string name, uint256 duration) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) Renew(opts *bind.TransactOpts, name string, duration *big.Int) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "renew", name, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xacf1a841.
//
// Solidity: function renew(string name, uint256 duration) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) Renew(name string, duration *big.Int) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.Renew(&_ENSOldRegistrarController.TransactOpts, name, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xacf1a841.
//
// Solidity: function renew(string name, uint256 duration) payable returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) Renew(name string, duration *big.Int) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.Renew(&_ENSOldRegistrarController.TransactOpts, name, duration)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) RenounceOwnership() (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.RenounceOwnership(&_ENSOldRegistrarController.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.RenounceOwnership(&_ENSOldRegistrarController.TransactOpts)
}

// SetCommitmentAges is a paid mutator transaction binding the contract method 0x7e324479.
//
// Solidity: function setCommitmentAges(uint256 _minCommitmentAge, uint256 _maxCommitmentAge) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) SetCommitmentAges(opts *bind.TransactOpts, _minCommitmentAge *big.Int, _maxCommitmentAge *big.Int) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "setCommitmentAges", _minCommitmentAge, _maxCommitmentAge)
}

// SetCommitmentAges is a paid mutator transaction binding the contract method 0x7e324479.
//
// Solidity: function setCommitmentAges(uint256 _minCommitmentAge, uint256 _maxCommitmentAge) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) SetCommitmentAges(_minCommitmentAge *big.Int, _maxCommitmentAge *big.Int) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.SetCommitmentAges(&_ENSOldRegistrarController.TransactOpts, _minCommitmentAge, _maxCommitmentAge)
}

// SetCommitmentAges is a paid mutator transaction binding the contract method 0x7e324479.
//
// Solidity: function setCommitmentAges(uint256 _minCommitmentAge, uint256 _maxCommitmentAge) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) SetCommitmentAges(_minCommitmentAge *big.Int, _maxCommitmentAge *big.Int) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.SetCommitmentAges(&_ENSOldRegistrarController.TransactOpts, _minCommitmentAge, _maxCommitmentAge)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address _prices) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) SetPriceOracle(opts *bind.TransactOpts, _prices common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "setPriceOracle", _prices)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address _prices) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) SetPriceOracle(_prices common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.SetPriceOracle(&_ENSOldRegistrarController.TransactOpts, _prices)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address _prices) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) SetPriceOracle(_prices common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.SetPriceOracle(&_ENSOldRegistrarController.TransactOpts, _prices)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.TransferOwnership(&_ENSOldRegistrarController.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.TransferOwnership(&_ENSOldRegistrarController.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ENSOldRegistrarController.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerSession) Withdraw() (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.Withdraw(&_ENSOldRegistrarController.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_ENSOldRegistrarController *ENSOldRegistrarControllerTransactorSession) Withdraw() (*types.Transaction, error) {
	return _ENSOldRegistrarController.Contract.Withdraw(&_ENSOldRegistrarController.TransactOpts)
}

// ENSOldRegistrarControllerNameRegisteredIterator is returned from FilterNameRegistered and is used to iterate over the raw logs and unpacked data for NameRegistered events raised by the ENSOldRegistrarController contract.
type ENSOldRegistrarControllerNameRegisteredIterator struct {
	Event *ENSOldRegistrarControllerNameRegistered // Event containing the contract specifics and raw log

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
func (it *ENSOldRegistrarControllerNameRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSOldRegistrarControllerNameRegistered)
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
		it.Event = new(ENSOldRegistrarControllerNameRegistered)
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
func (it *ENSOldRegistrarControllerNameRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSOldRegistrarControllerNameRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSOldRegistrarControllerNameRegistered represents a NameRegistered event raised by the ENSOldRegistrarController contract.
type ENSOldRegistrarControllerNameRegistered struct {
	Name    string
	Label   [32]byte
	Owner   common.Address
	Cost    *big.Int
	Expires *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNameRegistered is a free log retrieval operation binding the contract event 0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f.
//
// Solidity: event NameRegistered(string name, bytes32 indexed label, address indexed owner, uint256 cost, uint256 expires)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) FilterNameRegistered(opts *bind.FilterOpts, label [][32]byte, owner []common.Address) (*ENSOldRegistrarControllerNameRegisteredIterator, error) {

	var labelRule []interface{}
	for _, labelItem := range label {
		labelRule = append(labelRule, labelItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _ENSOldRegistrarController.contract.FilterLogs(opts, "NameRegistered", labelRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &ENSOldRegistrarControllerNameRegisteredIterator{contract: _ENSOldRegistrarController.contract, event: "NameRegistered", logs: logs, sub: sub}, nil
}

// WatchNameRegistered is a free log subscription operation binding the contract event 0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f.
//
// Solidity: event NameRegistered(string name, bytes32 indexed label, address indexed owner, uint256 cost, uint256 expires)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) WatchNameRegistered(opts *bind.WatchOpts, sink chan<- *ENSOldRegistrarControllerNameRegistered, label [][32]byte, owner []common.Address) (event.Subscription, error) {

	var labelRule []interface{}
	for _, labelItem := range label {
		labelRule = append(labelRule, labelItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _ENSOldRegistrarController.contract.WatchLogs(opts, "NameRegistered", labelRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSOldRegistrarControllerNameRegistered)
				if err := _ENSOldRegistrarController.contract.UnpackLog(event, "NameRegistered", log); err != nil {
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

// ParseNameRegistered is a log parse operation binding the contract event 0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f.
//
// Solidity: event NameRegistered(string name, bytes32 indexed label, address indexed owner, uint256 cost, uint256 expires)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) ParseNameRegistered(log types.Log) (*ENSOldRegistrarControllerNameRegistered, error) {
	event := new(ENSOldRegistrarControllerNameRegistered)
	if err := _ENSOldRegistrarController.contract.UnpackLog(event, "NameRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSOldRegistrarControllerNameRenewedIterator is returned from FilterNameRenewed and is used to iterate over the raw logs and unpacked data for NameRenewed events raised by the ENSOldRegistrarController contract.
type ENSOldRegistrarControllerNameRenewedIterator struct {
	Event *ENSOldRegistrarControllerNameRenewed // Event containing the contract specifics and raw log

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
func (it *ENSOldRegistrarControllerNameRenewedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSOldRegistrarControllerNameRenewed)
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
		it.Event = new(ENSOldRegistrarControllerNameRenewed)
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
func (it *ENSOldRegistrarControllerNameRenewedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSOldRegistrarControllerNameRenewedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSOldRegistrarControllerNameRenewed represents a NameRenewed event raised by the ENSOldRegistrarController contract.
type ENSOldRegistrarControllerNameRenewed struct {
	Name    string
	Label   [32]byte
	Cost    *big.Int
	Expires *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNameRenewed is a free log retrieval operation binding the contract event 0x3da24c024582931cfaf8267d8ed24d13a82a8068d5bd337d30ec45cea4e506ae.
//
// Solidity: event NameRenewed(string name, bytes32 indexed label, uint256 cost, uint256 expires)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) FilterNameRenewed(opts *bind.FilterOpts, label [][32]byte) (*ENSOldRegistrarControllerNameRenewedIterator, error) {

	var labelRule []interface{}
	for _, labelItem := range label {
		labelRule = append(labelRule, labelItem)
	}

	logs, sub, err := _ENSOldRegistrarController.contract.FilterLogs(opts, "NameRenewed", labelRule)
	if err != nil {
		return nil, err
	}
	return &ENSOldRegistrarControllerNameRenewedIterator{contract: _ENSOldRegistrarController.contract, event: "NameRenewed", logs: logs, sub: sub}, nil
}

// WatchNameRenewed is a free log subscription operation binding the contract event 0x3da24c024582931cfaf8267d8ed24d13a82a8068d5bd337d30ec45cea4e506ae.
//
// Solidity: event NameRenewed(string name, bytes32 indexed label, uint256 cost, uint256 expires)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) WatchNameRenewed(opts *bind.WatchOpts, sink chan<- *ENSOldRegistrarControllerNameRenewed, label [][32]byte) (event.Subscription, error) {

	var labelRule []interface{}
	for _, labelItem := range label {
		labelRule = append(labelRule, labelItem)
	}

	logs, sub, err := _ENSOldRegistrarController.contract.WatchLogs(opts, "NameRenewed", labelRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSOldRegistrarControllerNameRenewed)
				if err := _ENSOldRegistrarController.contract.UnpackLog(event, "NameRenewed", log); err != nil {
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

// ParseNameRenewed is a log parse operation binding the contract event 0x3da24c024582931cfaf8267d8ed24d13a82a8068d5bd337d30ec45cea4e506ae.
//
// Solidity: event NameRenewed(string name, bytes32 indexed label, uint256 cost, uint256 expires)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) ParseNameRenewed(log types.Log) (*ENSOldRegistrarControllerNameRenewed, error) {
	event := new(ENSOldRegistrarControllerNameRenewed)
	if err := _ENSOldRegistrarController.contract.UnpackLog(event, "NameRenewed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSOldRegistrarControllerNewPriceOracleIterator is returned from FilterNewPriceOracle and is used to iterate over the raw logs and unpacked data for NewPriceOracle events raised by the ENSOldRegistrarController contract.
type ENSOldRegistrarControllerNewPriceOracleIterator struct {
	Event *ENSOldRegistrarControllerNewPriceOracle // Event containing the contract specifics and raw log

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
func (it *ENSOldRegistrarControllerNewPriceOracleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSOldRegistrarControllerNewPriceOracle)
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
		it.Event = new(ENSOldRegistrarControllerNewPriceOracle)
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
func (it *ENSOldRegistrarControllerNewPriceOracleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSOldRegistrarControllerNewPriceOracleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSOldRegistrarControllerNewPriceOracle represents a NewPriceOracle event raised by the ENSOldRegistrarController contract.
type ENSOldRegistrarControllerNewPriceOracle struct {
	Oracle common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterNewPriceOracle is a free log retrieval operation binding the contract event 0xf261845a790fe29bbd6631e2ca4a5bdc83e6eed7c3271d9590d97287e00e9123.
//
// Solidity: event NewPriceOracle(address indexed oracle)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) FilterNewPriceOracle(opts *bind.FilterOpts, oracle []common.Address) (*ENSOldRegistrarControllerNewPriceOracleIterator, error) {

	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _ENSOldRegistrarController.contract.FilterLogs(opts, "NewPriceOracle", oracleRule)
	if err != nil {
		return nil, err
	}
	return &ENSOldRegistrarControllerNewPriceOracleIterator{contract: _ENSOldRegistrarController.contract, event: "NewPriceOracle", logs: logs, sub: sub}, nil
}

// WatchNewPriceOracle is a free log subscription operation binding the contract event 0xf261845a790fe29bbd6631e2ca4a5bdc83e6eed7c3271d9590d97287e00e9123.
//
// Solidity: event NewPriceOracle(address indexed oracle)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) WatchNewPriceOracle(opts *bind.WatchOpts, sink chan<- *ENSOldRegistrarControllerNewPriceOracle, oracle []common.Address) (event.Subscription, error) {

	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _ENSOldRegistrarController.contract.WatchLogs(opts, "NewPriceOracle", oracleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSOldRegistrarControllerNewPriceOracle)
				if err := _ENSOldRegistrarController.contract.UnpackLog(event, "NewPriceOracle", log); err != nil {
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

// ParseNewPriceOracle is a log parse operation binding the contract event 0xf261845a790fe29bbd6631e2ca4a5bdc83e6eed7c3271d9590d97287e00e9123.
//
// Solidity: event NewPriceOracle(address indexed oracle)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) ParseNewPriceOracle(log types.Log) (*ENSOldRegistrarControllerNewPriceOracle, error) {
	event := new(ENSOldRegistrarControllerNewPriceOracle)
	if err := _ENSOldRegistrarController.contract.UnpackLog(event, "NewPriceOracle", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ENSOldRegistrarControllerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ENSOldRegistrarController contract.
type ENSOldRegistrarControllerOwnershipTransferredIterator struct {
	Event *ENSOldRegistrarControllerOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ENSOldRegistrarControllerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ENSOldRegistrarControllerOwnershipTransferred)
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
		it.Event = new(ENSOldRegistrarControllerOwnershipTransferred)
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
func (it *ENSOldRegistrarControllerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ENSOldRegistrarControllerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ENSOldRegistrarControllerOwnershipTransferred represents a OwnershipTransferred event raised by the ENSOldRegistrarController contract.
type ENSOldRegistrarControllerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ENSOldRegistrarControllerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ENSOldRegistrarController.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ENSOldRegistrarControllerOwnershipTransferredIterator{contract: _ENSOldRegistrarController.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ENSOldRegistrarControllerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ENSOldRegistrarController.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ENSOldRegistrarControllerOwnershipTransferred)
				if err := _ENSOldRegistrarController.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ENSOldRegistrarController *ENSOldRegistrarControllerFilterer) ParseOwnershipTransferred(log types.Log) (*ENSOldRegistrarControllerOwnershipTransferred, error) {
	event := new(ENSOldRegistrarControllerOwnershipTransferred)
	if err := _ENSOldRegistrarController.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
