// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package oneinchoracle

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

// OneinchOracleMetaData contains all meta data concerning the OneinchOracle contract.
var OneinchOracleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractMultiWrapper\",\"name\":\"_multiWrapper\",\"type\":\"address\"},{\"internalType\":\"contractIOracle[]\",\"name\":\"existingOracles\",\"type\":\"address[]\"},{\"internalType\":\"enumOffchainOracle.OracleType[]\",\"name\":\"oracleTypes\",\"type\":\"uint8[]\"},{\"internalType\":\"contractIERC20[]\",\"name\":\"existingConnectors\",\"type\":\"address[]\"},{\"internalType\":\"contractIERC20\",\"name\":\"wBase\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"ArraysLengthMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ConnectorAlreadyAdded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidOracleTokenKind\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OracleAlreadyAdded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SameTokens\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooBigThreshold\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UnknownConnector\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UnknownOracle\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIERC20\",\"name\":\"connector\",\"type\":\"address\"}],\"name\":\"ConnectorAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIERC20\",\"name\":\"connector\",\"type\":\"address\"}],\"name\":\"ConnectorRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractMultiWrapper\",\"name\":\"multiWrapper\",\"type\":\"address\"}],\"name\":\"MultiWrapperUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIOracle\",\"name\":\"oracle\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"enumOffchainOracle.OracleType\",\"name\":\"oracleType\",\"type\":\"uint8\"}],\"name\":\"OracleAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIOracle\",\"name\":\"oracle\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"enumOffchainOracle.OracleType\",\"name\":\"oracleType\",\"type\":\"uint8\"}],\"name\":\"OracleRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"connector\",\"type\":\"address\"}],\"name\":\"addConnector\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIOracle\",\"name\":\"oracle\",\"type\":\"address\"},{\"internalType\":\"enumOffchainOracle.OracleType\",\"name\":\"oracleKind\",\"type\":\"uint8\"}],\"name\":\"addOracle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"connectors\",\"outputs\":[{\"internalType\":\"contractIERC20[]\",\"name\":\"allConnectors\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"dstToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"useWrappers\",\"type\":\"bool\"}],\"name\":\"getRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"useSrcWrappers\",\"type\":\"bool\"}],\"name\":\"getRateToEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"useSrcWrappers\",\"type\":\"bool\"},{\"internalType\":\"contractIERC20[]\",\"name\":\"customConnectors\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"thresholdFilter\",\"type\":\"uint256\"}],\"name\":\"getRateToEthWithCustomConnectors\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"useSrcWrappers\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"thresholdFilter\",\"type\":\"uint256\"}],\"name\":\"getRateToEthWithThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"dstToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"useWrappers\",\"type\":\"bool\"},{\"internalType\":\"contractIERC20[]\",\"name\":\"customConnectors\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"thresholdFilter\",\"type\":\"uint256\"}],\"name\":\"getRateWithCustomConnectors\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"dstToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"useWrappers\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"thresholdFilter\",\"type\":\"uint256\"}],\"name\":\"getRateWithThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"multiWrapper\",\"outputs\":[{\"internalType\":\"contractMultiWrapper\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"oracles\",\"outputs\":[{\"internalType\":\"contractIOracle[]\",\"name\":\"allOracles\",\"type\":\"address[]\"},{\"internalType\":\"enumOffchainOracle.OracleType[]\",\"name\":\"oracleTypes\",\"type\":\"uint8[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"connector\",\"type\":\"address\"}],\"name\":\"removeConnector\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIOracle\",\"name\":\"oracle\",\"type\":\"address\"},{\"internalType\":\"enumOffchainOracle.OracleType\",\"name\":\"oracleKind\",\"type\":\"uint8\"}],\"name\":\"removeOracle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractMultiWrapper\",\"name\":\"_multiWrapper\",\"type\":\"address\"}],\"name\":\"setMultiWrapper\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// OneinchOracleABI is the input ABI used to generate the binding from.
// Deprecated: Use OneinchOracleMetaData.ABI instead.
var OneinchOracleABI = OneinchOracleMetaData.ABI

// OneinchOracle is an auto generated Go binding around an Ethereum contract.
type OneinchOracle struct {
	OneinchOracleCaller     // Read-only binding to the contract
	OneinchOracleTransactor // Write-only binding to the contract
	OneinchOracleFilterer   // Log filterer for contract events
}

// OneinchOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type OneinchOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OneinchOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OneinchOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OneinchOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OneinchOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OneinchOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OneinchOracleSession struct {
	Contract     *OneinchOracle    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OneinchOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OneinchOracleCallerSession struct {
	Contract *OneinchOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// OneinchOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OneinchOracleTransactorSession struct {
	Contract     *OneinchOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// OneinchOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type OneinchOracleRaw struct {
	Contract *OneinchOracle // Generic contract binding to access the raw methods on
}

// OneinchOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OneinchOracleCallerRaw struct {
	Contract *OneinchOracleCaller // Generic read-only contract binding to access the raw methods on
}

// OneinchOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OneinchOracleTransactorRaw struct {
	Contract *OneinchOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOneinchOracle creates a new instance of OneinchOracle, bound to a specific deployed contract.
func NewOneinchOracle(address common.Address, backend bind.ContractBackend) (*OneinchOracle, error) {
	contract, err := bindOneinchOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OneinchOracle{OneinchOracleCaller: OneinchOracleCaller{contract: contract}, OneinchOracleTransactor: OneinchOracleTransactor{contract: contract}, OneinchOracleFilterer: OneinchOracleFilterer{contract: contract}}, nil
}

// NewOneinchOracleCaller creates a new read-only instance of OneinchOracle, bound to a specific deployed contract.
func NewOneinchOracleCaller(address common.Address, caller bind.ContractCaller) (*OneinchOracleCaller, error) {
	contract, err := bindOneinchOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OneinchOracleCaller{contract: contract}, nil
}

// NewOneinchOracleTransactor creates a new write-only instance of OneinchOracle, bound to a specific deployed contract.
func NewOneinchOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*OneinchOracleTransactor, error) {
	contract, err := bindOneinchOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OneinchOracleTransactor{contract: contract}, nil
}

// NewOneinchOracleFilterer creates a new log filterer instance of OneinchOracle, bound to a specific deployed contract.
func NewOneinchOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*OneinchOracleFilterer, error) {
	contract, err := bindOneinchOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OneinchOracleFilterer{contract: contract}, nil
}

// bindOneinchOracle binds a generic wrapper to an already deployed contract.
func bindOneinchOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OneinchOracleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OneinchOracle *OneinchOracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OneinchOracle.Contract.OneinchOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OneinchOracle *OneinchOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OneinchOracle.Contract.OneinchOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OneinchOracle *OneinchOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OneinchOracle.Contract.OneinchOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OneinchOracle *OneinchOracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OneinchOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OneinchOracle *OneinchOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OneinchOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OneinchOracle *OneinchOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OneinchOracle.Contract.contract.Transact(opts, method, params...)
}

// Connectors is a free data retrieval call binding the contract method 0x65050a68.
//
// Solidity: function connectors() view returns(address[] allConnectors)
func (_OneinchOracle *OneinchOracleCaller) Connectors(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "connectors")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// Connectors is a free data retrieval call binding the contract method 0x65050a68.
//
// Solidity: function connectors() view returns(address[] allConnectors)
func (_OneinchOracle *OneinchOracleSession) Connectors() ([]common.Address, error) {
	return _OneinchOracle.Contract.Connectors(&_OneinchOracle.CallOpts)
}

// Connectors is a free data retrieval call binding the contract method 0x65050a68.
//
// Solidity: function connectors() view returns(address[] allConnectors)
func (_OneinchOracle *OneinchOracleCallerSession) Connectors() ([]common.Address, error) {
	return _OneinchOracle.Contract.Connectors(&_OneinchOracle.CallOpts)
}

// GetRate is a free data retrieval call binding the contract method 0x802431fb.
//
// Solidity: function getRate(address srcToken, address dstToken, bool useWrappers) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCaller) GetRate(opts *bind.CallOpts, srcToken common.Address, dstToken common.Address, useWrappers bool) (*big.Int, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "getRate", srcToken, dstToken, useWrappers)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRate is a free data retrieval call binding the contract method 0x802431fb.
//
// Solidity: function getRate(address srcToken, address dstToken, bool useWrappers) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleSession) GetRate(srcToken common.Address, dstToken common.Address, useWrappers bool) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRate(&_OneinchOracle.CallOpts, srcToken, dstToken, useWrappers)
}

// GetRate is a free data retrieval call binding the contract method 0x802431fb.
//
// Solidity: function getRate(address srcToken, address dstToken, bool useWrappers) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCallerSession) GetRate(srcToken common.Address, dstToken common.Address, useWrappers bool) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRate(&_OneinchOracle.CallOpts, srcToken, dstToken, useWrappers)
}

// GetRateToEth is a free data retrieval call binding the contract method 0x7de4fd10.
//
// Solidity: function getRateToEth(address srcToken, bool useSrcWrappers) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCaller) GetRateToEth(opts *bind.CallOpts, srcToken common.Address, useSrcWrappers bool) (*big.Int, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "getRateToEth", srcToken, useSrcWrappers)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRateToEth is a free data retrieval call binding the contract method 0x7de4fd10.
//
// Solidity: function getRateToEth(address srcToken, bool useSrcWrappers) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleSession) GetRateToEth(srcToken common.Address, useSrcWrappers bool) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateToEth(&_OneinchOracle.CallOpts, srcToken, useSrcWrappers)
}

// GetRateToEth is a free data retrieval call binding the contract method 0x7de4fd10.
//
// Solidity: function getRateToEth(address srcToken, bool useSrcWrappers) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCallerSession) GetRateToEth(srcToken common.Address, useSrcWrappers bool) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateToEth(&_OneinchOracle.CallOpts, srcToken, useSrcWrappers)
}

// GetRateToEthWithCustomConnectors is a free data retrieval call binding the contract method 0xade8b048.
//
// Solidity: function getRateToEthWithCustomConnectors(address srcToken, bool useSrcWrappers, address[] customConnectors, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCaller) GetRateToEthWithCustomConnectors(opts *bind.CallOpts, srcToken common.Address, useSrcWrappers bool, customConnectors []common.Address, thresholdFilter *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "getRateToEthWithCustomConnectors", srcToken, useSrcWrappers, customConnectors, thresholdFilter)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRateToEthWithCustomConnectors is a free data retrieval call binding the contract method 0xade8b048.
//
// Solidity: function getRateToEthWithCustomConnectors(address srcToken, bool useSrcWrappers, address[] customConnectors, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleSession) GetRateToEthWithCustomConnectors(srcToken common.Address, useSrcWrappers bool, customConnectors []common.Address, thresholdFilter *big.Int) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateToEthWithCustomConnectors(&_OneinchOracle.CallOpts, srcToken, useSrcWrappers, customConnectors, thresholdFilter)
}

// GetRateToEthWithCustomConnectors is a free data retrieval call binding the contract method 0xade8b048.
//
// Solidity: function getRateToEthWithCustomConnectors(address srcToken, bool useSrcWrappers, address[] customConnectors, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCallerSession) GetRateToEthWithCustomConnectors(srcToken common.Address, useSrcWrappers bool, customConnectors []common.Address, thresholdFilter *big.Int) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateToEthWithCustomConnectors(&_OneinchOracle.CallOpts, srcToken, useSrcWrappers, customConnectors, thresholdFilter)
}

// GetRateToEthWithThreshold is a free data retrieval call binding the contract method 0x78159aae.
//
// Solidity: function getRateToEthWithThreshold(address srcToken, bool useSrcWrappers, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCaller) GetRateToEthWithThreshold(opts *bind.CallOpts, srcToken common.Address, useSrcWrappers bool, thresholdFilter *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "getRateToEthWithThreshold", srcToken, useSrcWrappers, thresholdFilter)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRateToEthWithThreshold is a free data retrieval call binding the contract method 0x78159aae.
//
// Solidity: function getRateToEthWithThreshold(address srcToken, bool useSrcWrappers, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleSession) GetRateToEthWithThreshold(srcToken common.Address, useSrcWrappers bool, thresholdFilter *big.Int) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateToEthWithThreshold(&_OneinchOracle.CallOpts, srcToken, useSrcWrappers, thresholdFilter)
}

// GetRateToEthWithThreshold is a free data retrieval call binding the contract method 0x78159aae.
//
// Solidity: function getRateToEthWithThreshold(address srcToken, bool useSrcWrappers, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCallerSession) GetRateToEthWithThreshold(srcToken common.Address, useSrcWrappers bool, thresholdFilter *big.Int) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateToEthWithThreshold(&_OneinchOracle.CallOpts, srcToken, useSrcWrappers, thresholdFilter)
}

// GetRateWithCustomConnectors is a free data retrieval call binding the contract method 0x6f9293b9.
//
// Solidity: function getRateWithCustomConnectors(address srcToken, address dstToken, bool useWrappers, address[] customConnectors, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCaller) GetRateWithCustomConnectors(opts *bind.CallOpts, srcToken common.Address, dstToken common.Address, useWrappers bool, customConnectors []common.Address, thresholdFilter *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "getRateWithCustomConnectors", srcToken, dstToken, useWrappers, customConnectors, thresholdFilter)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRateWithCustomConnectors is a free data retrieval call binding the contract method 0x6f9293b9.
//
// Solidity: function getRateWithCustomConnectors(address srcToken, address dstToken, bool useWrappers, address[] customConnectors, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleSession) GetRateWithCustomConnectors(srcToken common.Address, dstToken common.Address, useWrappers bool, customConnectors []common.Address, thresholdFilter *big.Int) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateWithCustomConnectors(&_OneinchOracle.CallOpts, srcToken, dstToken, useWrappers, customConnectors, thresholdFilter)
}

// GetRateWithCustomConnectors is a free data retrieval call binding the contract method 0x6f9293b9.
//
// Solidity: function getRateWithCustomConnectors(address srcToken, address dstToken, bool useWrappers, address[] customConnectors, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCallerSession) GetRateWithCustomConnectors(srcToken common.Address, dstToken common.Address, useWrappers bool, customConnectors []common.Address, thresholdFilter *big.Int) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateWithCustomConnectors(&_OneinchOracle.CallOpts, srcToken, dstToken, useWrappers, customConnectors, thresholdFilter)
}

// GetRateWithThreshold is a free data retrieval call binding the contract method 0x6744d6c7.
//
// Solidity: function getRateWithThreshold(address srcToken, address dstToken, bool useWrappers, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCaller) GetRateWithThreshold(opts *bind.CallOpts, srcToken common.Address, dstToken common.Address, useWrappers bool, thresholdFilter *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "getRateWithThreshold", srcToken, dstToken, useWrappers, thresholdFilter)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRateWithThreshold is a free data retrieval call binding the contract method 0x6744d6c7.
//
// Solidity: function getRateWithThreshold(address srcToken, address dstToken, bool useWrappers, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleSession) GetRateWithThreshold(srcToken common.Address, dstToken common.Address, useWrappers bool, thresholdFilter *big.Int) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateWithThreshold(&_OneinchOracle.CallOpts, srcToken, dstToken, useWrappers, thresholdFilter)
}

// GetRateWithThreshold is a free data retrieval call binding the contract method 0x6744d6c7.
//
// Solidity: function getRateWithThreshold(address srcToken, address dstToken, bool useWrappers, uint256 thresholdFilter) view returns(uint256 weightedRate)
func (_OneinchOracle *OneinchOracleCallerSession) GetRateWithThreshold(srcToken common.Address, dstToken common.Address, useWrappers bool, thresholdFilter *big.Int) (*big.Int, error) {
	return _OneinchOracle.Contract.GetRateWithThreshold(&_OneinchOracle.CallOpts, srcToken, dstToken, useWrappers, thresholdFilter)
}

// MultiWrapper is a free data retrieval call binding the contract method 0xb77910dc.
//
// Solidity: function multiWrapper() view returns(address)
func (_OneinchOracle *OneinchOracleCaller) MultiWrapper(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "multiWrapper")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MultiWrapper is a free data retrieval call binding the contract method 0xb77910dc.
//
// Solidity: function multiWrapper() view returns(address)
func (_OneinchOracle *OneinchOracleSession) MultiWrapper() (common.Address, error) {
	return _OneinchOracle.Contract.MultiWrapper(&_OneinchOracle.CallOpts)
}

// MultiWrapper is a free data retrieval call binding the contract method 0xb77910dc.
//
// Solidity: function multiWrapper() view returns(address)
func (_OneinchOracle *OneinchOracleCallerSession) MultiWrapper() (common.Address, error) {
	return _OneinchOracle.Contract.MultiWrapper(&_OneinchOracle.CallOpts)
}

// Oracles is a free data retrieval call binding the contract method 0x2857373a.
//
// Solidity: function oracles() view returns(address[] allOracles, uint8[] oracleTypes)
func (_OneinchOracle *OneinchOracleCaller) Oracles(opts *bind.CallOpts) (struct {
	AllOracles  []common.Address
	OracleTypes []uint8
}, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "oracles")

	outstruct := new(struct {
		AllOracles  []common.Address
		OracleTypes []uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.AllOracles = *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	outstruct.OracleTypes = *abi.ConvertType(out[1], new([]uint8)).(*[]uint8)

	return *outstruct, err

}

// Oracles is a free data retrieval call binding the contract method 0x2857373a.
//
// Solidity: function oracles() view returns(address[] allOracles, uint8[] oracleTypes)
func (_OneinchOracle *OneinchOracleSession) Oracles() (struct {
	AllOracles  []common.Address
	OracleTypes []uint8
}, error) {
	return _OneinchOracle.Contract.Oracles(&_OneinchOracle.CallOpts)
}

// Oracles is a free data retrieval call binding the contract method 0x2857373a.
//
// Solidity: function oracles() view returns(address[] allOracles, uint8[] oracleTypes)
func (_OneinchOracle *OneinchOracleCallerSession) Oracles() (struct {
	AllOracles  []common.Address
	OracleTypes []uint8
}, error) {
	return _OneinchOracle.Contract.Oracles(&_OneinchOracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_OneinchOracle *OneinchOracleCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OneinchOracle.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_OneinchOracle *OneinchOracleSession) Owner() (common.Address, error) {
	return _OneinchOracle.Contract.Owner(&_OneinchOracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_OneinchOracle *OneinchOracleCallerSession) Owner() (common.Address, error) {
	return _OneinchOracle.Contract.Owner(&_OneinchOracle.CallOpts)
}

// AddConnector is a paid mutator transaction binding the contract method 0xaa16d4c0.
//
// Solidity: function addConnector(address connector) returns()
func (_OneinchOracle *OneinchOracleTransactor) AddConnector(opts *bind.TransactOpts, connector common.Address) (*types.Transaction, error) {
	return _OneinchOracle.contract.Transact(opts, "addConnector", connector)
}

// AddConnector is a paid mutator transaction binding the contract method 0xaa16d4c0.
//
// Solidity: function addConnector(address connector) returns()
func (_OneinchOracle *OneinchOracleSession) AddConnector(connector common.Address) (*types.Transaction, error) {
	return _OneinchOracle.Contract.AddConnector(&_OneinchOracle.TransactOpts, connector)
}

// AddConnector is a paid mutator transaction binding the contract method 0xaa16d4c0.
//
// Solidity: function addConnector(address connector) returns()
func (_OneinchOracle *OneinchOracleTransactorSession) AddConnector(connector common.Address) (*types.Transaction, error) {
	return _OneinchOracle.Contract.AddConnector(&_OneinchOracle.TransactOpts, connector)
}

// AddOracle is a paid mutator transaction binding the contract method 0x9d4d7b1c.
//
// Solidity: function addOracle(address oracle, uint8 oracleKind) returns()
func (_OneinchOracle *OneinchOracleTransactor) AddOracle(opts *bind.TransactOpts, oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneinchOracle.contract.Transact(opts, "addOracle", oracle, oracleKind)
}

// AddOracle is a paid mutator transaction binding the contract method 0x9d4d7b1c.
//
// Solidity: function addOracle(address oracle, uint8 oracleKind) returns()
func (_OneinchOracle *OneinchOracleSession) AddOracle(oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneinchOracle.Contract.AddOracle(&_OneinchOracle.TransactOpts, oracle, oracleKind)
}

// AddOracle is a paid mutator transaction binding the contract method 0x9d4d7b1c.
//
// Solidity: function addOracle(address oracle, uint8 oracleKind) returns()
func (_OneinchOracle *OneinchOracleTransactorSession) AddOracle(oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneinchOracle.Contract.AddOracle(&_OneinchOracle.TransactOpts, oracle, oracleKind)
}

// RemoveConnector is a paid mutator transaction binding the contract method 0x1a6c6a98.
//
// Solidity: function removeConnector(address connector) returns()
func (_OneinchOracle *OneinchOracleTransactor) RemoveConnector(opts *bind.TransactOpts, connector common.Address) (*types.Transaction, error) {
	return _OneinchOracle.contract.Transact(opts, "removeConnector", connector)
}

// RemoveConnector is a paid mutator transaction binding the contract method 0x1a6c6a98.
//
// Solidity: function removeConnector(address connector) returns()
func (_OneinchOracle *OneinchOracleSession) RemoveConnector(connector common.Address) (*types.Transaction, error) {
	return _OneinchOracle.Contract.RemoveConnector(&_OneinchOracle.TransactOpts, connector)
}

// RemoveConnector is a paid mutator transaction binding the contract method 0x1a6c6a98.
//
// Solidity: function removeConnector(address connector) returns()
func (_OneinchOracle *OneinchOracleTransactorSession) RemoveConnector(connector common.Address) (*types.Transaction, error) {
	return _OneinchOracle.Contract.RemoveConnector(&_OneinchOracle.TransactOpts, connector)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0xf0b92e40.
//
// Solidity: function removeOracle(address oracle, uint8 oracleKind) returns()
func (_OneinchOracle *OneinchOracleTransactor) RemoveOracle(opts *bind.TransactOpts, oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneinchOracle.contract.Transact(opts, "removeOracle", oracle, oracleKind)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0xf0b92e40.
//
// Solidity: function removeOracle(address oracle, uint8 oracleKind) returns()
func (_OneinchOracle *OneinchOracleSession) RemoveOracle(oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneinchOracle.Contract.RemoveOracle(&_OneinchOracle.TransactOpts, oracle, oracleKind)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0xf0b92e40.
//
// Solidity: function removeOracle(address oracle, uint8 oracleKind) returns()
func (_OneinchOracle *OneinchOracleTransactorSession) RemoveOracle(oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneinchOracle.Contract.RemoveOracle(&_OneinchOracle.TransactOpts, oracle, oracleKind)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OneinchOracle *OneinchOracleTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OneinchOracle.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OneinchOracle *OneinchOracleSession) RenounceOwnership() (*types.Transaction, error) {
	return _OneinchOracle.Contract.RenounceOwnership(&_OneinchOracle.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OneinchOracle *OneinchOracleTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _OneinchOracle.Contract.RenounceOwnership(&_OneinchOracle.TransactOpts)
}

// SetMultiWrapper is a paid mutator transaction binding the contract method 0xd0626518.
//
// Solidity: function setMultiWrapper(address _multiWrapper) returns()
func (_OneinchOracle *OneinchOracleTransactor) SetMultiWrapper(opts *bind.TransactOpts, _multiWrapper common.Address) (*types.Transaction, error) {
	return _OneinchOracle.contract.Transact(opts, "setMultiWrapper", _multiWrapper)
}

// SetMultiWrapper is a paid mutator transaction binding the contract method 0xd0626518.
//
// Solidity: function setMultiWrapper(address _multiWrapper) returns()
func (_OneinchOracle *OneinchOracleSession) SetMultiWrapper(_multiWrapper common.Address) (*types.Transaction, error) {
	return _OneinchOracle.Contract.SetMultiWrapper(&_OneinchOracle.TransactOpts, _multiWrapper)
}

// SetMultiWrapper is a paid mutator transaction binding the contract method 0xd0626518.
//
// Solidity: function setMultiWrapper(address _multiWrapper) returns()
func (_OneinchOracle *OneinchOracleTransactorSession) SetMultiWrapper(_multiWrapper common.Address) (*types.Transaction, error) {
	return _OneinchOracle.Contract.SetMultiWrapper(&_OneinchOracle.TransactOpts, _multiWrapper)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_OneinchOracle *OneinchOracleTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _OneinchOracle.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_OneinchOracle *OneinchOracleSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _OneinchOracle.Contract.TransferOwnership(&_OneinchOracle.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_OneinchOracle *OneinchOracleTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _OneinchOracle.Contract.TransferOwnership(&_OneinchOracle.TransactOpts, newOwner)
}

// OneinchOracleConnectorAddedIterator is returned from FilterConnectorAdded and is used to iterate over the raw logs and unpacked data for ConnectorAdded events raised by the OneinchOracle contract.
type OneinchOracleConnectorAddedIterator struct {
	Event *OneinchOracleConnectorAdded // Event containing the contract specifics and raw log

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
func (it *OneinchOracleConnectorAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneinchOracleConnectorAdded)
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
		it.Event = new(OneinchOracleConnectorAdded)
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
func (it *OneinchOracleConnectorAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneinchOracleConnectorAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneinchOracleConnectorAdded represents a ConnectorAdded event raised by the OneinchOracle contract.
type OneinchOracleConnectorAdded struct {
	Connector common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterConnectorAdded is a free log retrieval operation binding the contract event 0xff88af5d962d47fd25d87755e8267a029fad5a91740c67d0dade2bdbe5268a1d.
//
// Solidity: event ConnectorAdded(address connector)
func (_OneinchOracle *OneinchOracleFilterer) FilterConnectorAdded(opts *bind.FilterOpts) (*OneinchOracleConnectorAddedIterator, error) {

	logs, sub, err := _OneinchOracle.contract.FilterLogs(opts, "ConnectorAdded")
	if err != nil {
		return nil, err
	}
	return &OneinchOracleConnectorAddedIterator{contract: _OneinchOracle.contract, event: "ConnectorAdded", logs: logs, sub: sub}, nil
}

// WatchConnectorAdded is a free log subscription operation binding the contract event 0xff88af5d962d47fd25d87755e8267a029fad5a91740c67d0dade2bdbe5268a1d.
//
// Solidity: event ConnectorAdded(address connector)
func (_OneinchOracle *OneinchOracleFilterer) WatchConnectorAdded(opts *bind.WatchOpts, sink chan<- *OneinchOracleConnectorAdded) (event.Subscription, error) {

	logs, sub, err := _OneinchOracle.contract.WatchLogs(opts, "ConnectorAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneinchOracleConnectorAdded)
				if err := _OneinchOracle.contract.UnpackLog(event, "ConnectorAdded", log); err != nil {
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

// ParseConnectorAdded is a log parse operation binding the contract event 0xff88af5d962d47fd25d87755e8267a029fad5a91740c67d0dade2bdbe5268a1d.
//
// Solidity: event ConnectorAdded(address connector)
func (_OneinchOracle *OneinchOracleFilterer) ParseConnectorAdded(log types.Log) (*OneinchOracleConnectorAdded, error) {
	event := new(OneinchOracleConnectorAdded)
	if err := _OneinchOracle.contract.UnpackLog(event, "ConnectorAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneinchOracleConnectorRemovedIterator is returned from FilterConnectorRemoved and is used to iterate over the raw logs and unpacked data for ConnectorRemoved events raised by the OneinchOracle contract.
type OneinchOracleConnectorRemovedIterator struct {
	Event *OneinchOracleConnectorRemoved // Event containing the contract specifics and raw log

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
func (it *OneinchOracleConnectorRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneinchOracleConnectorRemoved)
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
		it.Event = new(OneinchOracleConnectorRemoved)
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
func (it *OneinchOracleConnectorRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneinchOracleConnectorRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneinchOracleConnectorRemoved represents a ConnectorRemoved event raised by the OneinchOracle contract.
type OneinchOracleConnectorRemoved struct {
	Connector common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterConnectorRemoved is a free log retrieval operation binding the contract event 0x6825b26a0827e9c2ceca01d6289ce4a40e629dc074ec48ea4727d1afbff359f5.
//
// Solidity: event ConnectorRemoved(address connector)
func (_OneinchOracle *OneinchOracleFilterer) FilterConnectorRemoved(opts *bind.FilterOpts) (*OneinchOracleConnectorRemovedIterator, error) {

	logs, sub, err := _OneinchOracle.contract.FilterLogs(opts, "ConnectorRemoved")
	if err != nil {
		return nil, err
	}
	return &OneinchOracleConnectorRemovedIterator{contract: _OneinchOracle.contract, event: "ConnectorRemoved", logs: logs, sub: sub}, nil
}

// WatchConnectorRemoved is a free log subscription operation binding the contract event 0x6825b26a0827e9c2ceca01d6289ce4a40e629dc074ec48ea4727d1afbff359f5.
//
// Solidity: event ConnectorRemoved(address connector)
func (_OneinchOracle *OneinchOracleFilterer) WatchConnectorRemoved(opts *bind.WatchOpts, sink chan<- *OneinchOracleConnectorRemoved) (event.Subscription, error) {

	logs, sub, err := _OneinchOracle.contract.WatchLogs(opts, "ConnectorRemoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneinchOracleConnectorRemoved)
				if err := _OneinchOracle.contract.UnpackLog(event, "ConnectorRemoved", log); err != nil {
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

// ParseConnectorRemoved is a log parse operation binding the contract event 0x6825b26a0827e9c2ceca01d6289ce4a40e629dc074ec48ea4727d1afbff359f5.
//
// Solidity: event ConnectorRemoved(address connector)
func (_OneinchOracle *OneinchOracleFilterer) ParseConnectorRemoved(log types.Log) (*OneinchOracleConnectorRemoved, error) {
	event := new(OneinchOracleConnectorRemoved)
	if err := _OneinchOracle.contract.UnpackLog(event, "ConnectorRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneinchOracleMultiWrapperUpdatedIterator is returned from FilterMultiWrapperUpdated and is used to iterate over the raw logs and unpacked data for MultiWrapperUpdated events raised by the OneinchOracle contract.
type OneinchOracleMultiWrapperUpdatedIterator struct {
	Event *OneinchOracleMultiWrapperUpdated // Event containing the contract specifics and raw log

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
func (it *OneinchOracleMultiWrapperUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneinchOracleMultiWrapperUpdated)
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
		it.Event = new(OneinchOracleMultiWrapperUpdated)
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
func (it *OneinchOracleMultiWrapperUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneinchOracleMultiWrapperUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneinchOracleMultiWrapperUpdated represents a MultiWrapperUpdated event raised by the OneinchOracle contract.
type OneinchOracleMultiWrapperUpdated struct {
	MultiWrapper common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterMultiWrapperUpdated is a free log retrieval operation binding the contract event 0x1030152fe2062b574a830e6b9f13c65995990df31e4dc708d142533bb3ad0f52.
//
// Solidity: event MultiWrapperUpdated(address multiWrapper)
func (_OneinchOracle *OneinchOracleFilterer) FilterMultiWrapperUpdated(opts *bind.FilterOpts) (*OneinchOracleMultiWrapperUpdatedIterator, error) {

	logs, sub, err := _OneinchOracle.contract.FilterLogs(opts, "MultiWrapperUpdated")
	if err != nil {
		return nil, err
	}
	return &OneinchOracleMultiWrapperUpdatedIterator{contract: _OneinchOracle.contract, event: "MultiWrapperUpdated", logs: logs, sub: sub}, nil
}

// WatchMultiWrapperUpdated is a free log subscription operation binding the contract event 0x1030152fe2062b574a830e6b9f13c65995990df31e4dc708d142533bb3ad0f52.
//
// Solidity: event MultiWrapperUpdated(address multiWrapper)
func (_OneinchOracle *OneinchOracleFilterer) WatchMultiWrapperUpdated(opts *bind.WatchOpts, sink chan<- *OneinchOracleMultiWrapperUpdated) (event.Subscription, error) {

	logs, sub, err := _OneinchOracle.contract.WatchLogs(opts, "MultiWrapperUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneinchOracleMultiWrapperUpdated)
				if err := _OneinchOracle.contract.UnpackLog(event, "MultiWrapperUpdated", log); err != nil {
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

// ParseMultiWrapperUpdated is a log parse operation binding the contract event 0x1030152fe2062b574a830e6b9f13c65995990df31e4dc708d142533bb3ad0f52.
//
// Solidity: event MultiWrapperUpdated(address multiWrapper)
func (_OneinchOracle *OneinchOracleFilterer) ParseMultiWrapperUpdated(log types.Log) (*OneinchOracleMultiWrapperUpdated, error) {
	event := new(OneinchOracleMultiWrapperUpdated)
	if err := _OneinchOracle.contract.UnpackLog(event, "MultiWrapperUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneinchOracleOracleAddedIterator is returned from FilterOracleAdded and is used to iterate over the raw logs and unpacked data for OracleAdded events raised by the OneinchOracle contract.
type OneinchOracleOracleAddedIterator struct {
	Event *OneinchOracleOracleAdded // Event containing the contract specifics and raw log

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
func (it *OneinchOracleOracleAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneinchOracleOracleAdded)
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
		it.Event = new(OneinchOracleOracleAdded)
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
func (it *OneinchOracleOracleAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneinchOracleOracleAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneinchOracleOracleAdded represents a OracleAdded event raised by the OneinchOracle contract.
type OneinchOracleOracleAdded struct {
	Oracle     common.Address
	OracleType uint8
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOracleAdded is a free log retrieval operation binding the contract event 0x5874b2072ff37562df54063dd700c59d45f311bdf6f9cabb5a15f0ffb2e9f622.
//
// Solidity: event OracleAdded(address oracle, uint8 oracleType)
func (_OneinchOracle *OneinchOracleFilterer) FilterOracleAdded(opts *bind.FilterOpts) (*OneinchOracleOracleAddedIterator, error) {

	logs, sub, err := _OneinchOracle.contract.FilterLogs(opts, "OracleAdded")
	if err != nil {
		return nil, err
	}
	return &OneinchOracleOracleAddedIterator{contract: _OneinchOracle.contract, event: "OracleAdded", logs: logs, sub: sub}, nil
}

// WatchOracleAdded is a free log subscription operation binding the contract event 0x5874b2072ff37562df54063dd700c59d45f311bdf6f9cabb5a15f0ffb2e9f622.
//
// Solidity: event OracleAdded(address oracle, uint8 oracleType)
func (_OneinchOracle *OneinchOracleFilterer) WatchOracleAdded(opts *bind.WatchOpts, sink chan<- *OneinchOracleOracleAdded) (event.Subscription, error) {

	logs, sub, err := _OneinchOracle.contract.WatchLogs(opts, "OracleAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneinchOracleOracleAdded)
				if err := _OneinchOracle.contract.UnpackLog(event, "OracleAdded", log); err != nil {
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

// ParseOracleAdded is a log parse operation binding the contract event 0x5874b2072ff37562df54063dd700c59d45f311bdf6f9cabb5a15f0ffb2e9f622.
//
// Solidity: event OracleAdded(address oracle, uint8 oracleType)
func (_OneinchOracle *OneinchOracleFilterer) ParseOracleAdded(log types.Log) (*OneinchOracleOracleAdded, error) {
	event := new(OneinchOracleOracleAdded)
	if err := _OneinchOracle.contract.UnpackLog(event, "OracleAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneinchOracleOracleRemovedIterator is returned from FilterOracleRemoved and is used to iterate over the raw logs and unpacked data for OracleRemoved events raised by the OneinchOracle contract.
type OneinchOracleOracleRemovedIterator struct {
	Event *OneinchOracleOracleRemoved // Event containing the contract specifics and raw log

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
func (it *OneinchOracleOracleRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneinchOracleOracleRemoved)
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
		it.Event = new(OneinchOracleOracleRemoved)
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
func (it *OneinchOracleOracleRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneinchOracleOracleRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneinchOracleOracleRemoved represents a OracleRemoved event raised by the OneinchOracle contract.
type OneinchOracleOracleRemoved struct {
	Oracle     common.Address
	OracleType uint8
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOracleRemoved is a free log retrieval operation binding the contract event 0x7a7f56716fe703fb190529c336e57df71ab88188ba47e8d786bac684b61ab9a6.
//
// Solidity: event OracleRemoved(address oracle, uint8 oracleType)
func (_OneinchOracle *OneinchOracleFilterer) FilterOracleRemoved(opts *bind.FilterOpts) (*OneinchOracleOracleRemovedIterator, error) {

	logs, sub, err := _OneinchOracle.contract.FilterLogs(opts, "OracleRemoved")
	if err != nil {
		return nil, err
	}
	return &OneinchOracleOracleRemovedIterator{contract: _OneinchOracle.contract, event: "OracleRemoved", logs: logs, sub: sub}, nil
}

// WatchOracleRemoved is a free log subscription operation binding the contract event 0x7a7f56716fe703fb190529c336e57df71ab88188ba47e8d786bac684b61ab9a6.
//
// Solidity: event OracleRemoved(address oracle, uint8 oracleType)
func (_OneinchOracle *OneinchOracleFilterer) WatchOracleRemoved(opts *bind.WatchOpts, sink chan<- *OneinchOracleOracleRemoved) (event.Subscription, error) {

	logs, sub, err := _OneinchOracle.contract.WatchLogs(opts, "OracleRemoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneinchOracleOracleRemoved)
				if err := _OneinchOracle.contract.UnpackLog(event, "OracleRemoved", log); err != nil {
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

// ParseOracleRemoved is a log parse operation binding the contract event 0x7a7f56716fe703fb190529c336e57df71ab88188ba47e8d786bac684b61ab9a6.
//
// Solidity: event OracleRemoved(address oracle, uint8 oracleType)
func (_OneinchOracle *OneinchOracleFilterer) ParseOracleRemoved(log types.Log) (*OneinchOracleOracleRemoved, error) {
	event := new(OneinchOracleOracleRemoved)
	if err := _OneinchOracle.contract.UnpackLog(event, "OracleRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneinchOracleOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the OneinchOracle contract.
type OneinchOracleOwnershipTransferredIterator struct {
	Event *OneinchOracleOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *OneinchOracleOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneinchOracleOwnershipTransferred)
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
		it.Event = new(OneinchOracleOwnershipTransferred)
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
func (it *OneinchOracleOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneinchOracleOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneinchOracleOwnershipTransferred represents a OwnershipTransferred event raised by the OneinchOracle contract.
type OneinchOracleOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_OneinchOracle *OneinchOracleFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*OneinchOracleOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _OneinchOracle.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &OneinchOracleOwnershipTransferredIterator{contract: _OneinchOracle.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_OneinchOracle *OneinchOracleFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *OneinchOracleOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _OneinchOracle.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneinchOracleOwnershipTransferred)
				if err := _OneinchOracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_OneinchOracle *OneinchOracleFilterer) ParseOwnershipTransferred(log types.Log) (*OneinchOracleOwnershipTransferred, error) {
	event := new(OneinchOracleOwnershipTransferred)
	if err := _OneinchOracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
