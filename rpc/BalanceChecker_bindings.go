// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package rpc

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
)

// BalanceMetaData contains all meta data concerning the Balance contract.
var BalanceMetaData = &bind.MetaData{
	ABI: "[{\"constant\":true,\"inputs\":[{\"name\":\"user\",\"type\":\"address\"},{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"tokenBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"users\",\"type\":\"address[]\"},{\"name\":\"tokens\",\"type\":\"address[]\"}],\"name\":\"balances\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"}]",
}

// BalanceABI is the input ABI used to generate the binding from.
// Deprecated: Use BalanceMetaData.ABI instead.
var BalanceABI = BalanceMetaData.ABI

// Balance is an auto generated Go binding around an Ethereum contract.
type Balance struct {
	BalanceCaller     // Read-only binding to the contract
	BalanceTransactor // Write-only binding to the contract
	BalanceFilterer   // Log filterer for contract events
}

// BalanceCaller is an auto generated read-only Go binding around an Ethereum contract.
type BalanceCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalanceTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BalanceTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalanceFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BalanceFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalanceSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BalanceSession struct {
	Contract     *Balance          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BalanceCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BalanceCallerSession struct {
	Contract *BalanceCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// BalanceTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BalanceTransactorSession struct {
	Contract     *BalanceTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// BalanceRaw is an auto generated low-level Go binding around an Ethereum contract.
type BalanceRaw struct {
	Contract *Balance // Generic contract binding to access the raw methods on
}

// BalanceCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BalanceCallerRaw struct {
	Contract *BalanceCaller // Generic read-only contract binding to access the raw methods on
}

// BalanceTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BalanceTransactorRaw struct {
	Contract *BalanceTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBalance creates a new instance of Balance, bound to a specific deployed contract.
func NewBalance(address common.Address, backend bind.ContractBackend) (*Balance, error) {
	contract, err := bindBalance(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Balance{BalanceCaller: BalanceCaller{contract: contract}, BalanceTransactor: BalanceTransactor{contract: contract}, BalanceFilterer: BalanceFilterer{contract: contract}}, nil
}

// NewBalanceCaller creates a new read-only instance of Balance, bound to a specific deployed contract.
func NewBalanceCaller(address common.Address, caller bind.ContractCaller) (*BalanceCaller, error) {
	contract, err := bindBalance(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BalanceCaller{contract: contract}, nil
}

// NewBalanceTransactor creates a new write-only instance of Balance, bound to a specific deployed contract.
func NewBalanceTransactor(address common.Address, transactor bind.ContractTransactor) (*BalanceTransactor, error) {
	contract, err := bindBalance(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BalanceTransactor{contract: contract}, nil
}

// NewBalanceFilterer creates a new log filterer instance of Balance, bound to a specific deployed contract.
func NewBalanceFilterer(address common.Address, filterer bind.ContractFilterer) (*BalanceFilterer, error) {
	contract, err := bindBalance(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BalanceFilterer{contract: contract}, nil
}

// bindBalance binds a generic wrapper to an already deployed contract.
func bindBalance(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BalanceABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Balance *BalanceRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Balance.Contract.BalanceCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Balance *BalanceRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Balance.Contract.BalanceTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Balance *BalanceRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Balance.Contract.BalanceTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Balance *BalanceCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Balance.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Balance *BalanceTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Balance.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Balance *BalanceTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Balance.Contract.contract.Transact(opts, method, params...)
}

// Balances is a free data retrieval call binding the contract method 0xf0002ea9.
//
// Solidity: function balances(address[] users, address[] tokens) view returns(uint256[])
func (_Balance *BalanceCaller) Balances(opts *bind.CallOpts, users []common.Address, tokens []common.Address) ([]*big.Int, error) {
	var out []interface{}
	err := _Balance.contract.Call(opts, &out, "balances", users, tokens)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// Balances is a free data retrieval call binding the contract method 0xf0002ea9.
//
// Solidity: function balances(address[] users, address[] tokens) view returns(uint256[])
func (_Balance *BalanceSession) Balances(users []common.Address, tokens []common.Address) ([]*big.Int, error) {
	return _Balance.Contract.Balances(&_Balance.CallOpts, users, tokens)
}

// Balances is a free data retrieval call binding the contract method 0xf0002ea9.
//
// Solidity: function balances(address[] users, address[] tokens) view returns(uint256[])
func (_Balance *BalanceCallerSession) Balances(users []common.Address, tokens []common.Address) ([]*big.Int, error) {
	return _Balance.Contract.Balances(&_Balance.CallOpts, users, tokens)
}

// TokenBalance is a free data retrieval call binding the contract method 0x1049334f.
//
// Solidity: function tokenBalance(address user, address token) view returns(uint256)
func (_Balance *BalanceCaller) TokenBalance(opts *bind.CallOpts, user common.Address, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Balance.contract.Call(opts, &out, "tokenBalance", user, token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenBalance is a free data retrieval call binding the contract method 0x1049334f.
//
// Solidity: function tokenBalance(address user, address token) view returns(uint256)
func (_Balance *BalanceSession) TokenBalance(user common.Address, token common.Address) (*big.Int, error) {
	return _Balance.Contract.TokenBalance(&_Balance.CallOpts, user, token)
}

// TokenBalance is a free data retrieval call binding the contract method 0x1049334f.
//
// Solidity: function tokenBalance(address user, address token) view returns(uint256)
func (_Balance *BalanceCallerSession) TokenBalance(user common.Address, token common.Address) (*big.Int, error) {
	return _Balance.Contract.TokenBalance(&_Balance.CallOpts, user, token)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Balance *BalanceTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Balance.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Balance *BalanceSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Balance.Contract.Fallback(&_Balance.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Balance *BalanceTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Balance.Contract.Fallback(&_Balance.TransactOpts, calldata)
}
