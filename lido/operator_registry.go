// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package lido

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

// OperatorRegistryMetaData contains all meta data concerning the OperatorRegistry contract.
var OperatorRegistryMetaData = &bind.MetaData{
	ABI: "[{\"constant\":true,\"inputs\":[],\"name\":\"hasInitialized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_targetLimitMode\",\"type\":\"uint256\"},{\"name\":\"_targetLimit\",\"type\":\"uint256\"}],\"name\":\"updateTargetValidatorsLimits\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_keysCount\",\"type\":\"uint256\"},{\"name\":\"_publicKeys\",\"type\":\"bytes\"},{\"name\":\"_signatures\",\"type\":\"bytes\"}],\"name\":\"addSigningKeys\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getType\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_script\",\"type\":\"bytes\"}],\"name\":\"getEVMScriptExecutor\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"clearNodeOperatorPenalty\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getRecoveryVault\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_offset\",\"type\":\"uint256\"},{\"name\":\"_limit\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorIds\",\"outputs\":[{\"name\":\"nodeOperatorIds\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_offset\",\"type\":\"uint256\"},{\"name\":\"_limit\",\"type\":\"uint256\"}],\"name\":\"getSigningKeys\",\"outputs\":[{\"name\":\"pubkeys\",\"type\":\"bytes\"},{\"name\":\"signatures\",\"type\":\"bytes\"},{\"name\":\"used\",\"type\":\"bool[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_fromIndex\",\"type\":\"uint256\"},{\"name\":\"_keysCount\",\"type\":\"uint256\"}],\"name\":\"removeSigningKeysOperatorBH\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorIsActive\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_name\",\"type\":\"string\"}],\"name\":\"setNodeOperatorName\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_totalRewardShares\",\"type\":\"uint256\"}],\"name\":\"getRewardsDistribution\",\"outputs\":[{\"name\":\"recipients\",\"type\":\"address[]\"},{\"name\":\"shares\",\"type\":\"uint256[]\"},{\"name\":\"penalized\",\"type\":\"bool[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_indexFrom\",\"type\":\"uint256\"},{\"name\":\"_indexTo\",\"type\":\"uint256\"}],\"name\":\"invalidateReadyToDepositKeysRange\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_locator\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"bytes32\"},{\"name\":\"_stuckPenaltyDelay\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_delay\",\"type\":\"uint256\"}],\"name\":\"setStuckPenaltyDelay\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalizeUpgrade_v3\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStuckPenaltyDelay\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"removeSigningKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getRewardDistributionState\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_fromIndex\",\"type\":\"uint256\"},{\"name\":\"_keysCount\",\"type\":\"uint256\"}],\"name\":\"removeSigningKeys\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"isOperatorPenalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"deactivateNodeOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"allowRecoverability\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"STAKING_ROUTER_ROLE\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_keysCount\",\"type\":\"uint256\"},{\"name\":\"_publicKeys\",\"type\":\"bytes\"},{\"name\":\"_signatures\",\"type\":\"bytes\"}],\"name\":\"addSigningKeysOperatorBH\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"appId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getActiveNodeOperatorsCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_name\",\"type\":\"string\"},{\"name\":\"_rewardAddress\",\"type\":\"address\"}],\"name\":\"addNodeOperator\",\"outputs\":[{\"name\":\"id\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getContractVersion\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getInitializationBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getUnusedSigningKeyCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"onRewardsMinted\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MANAGE_NODE_OPERATOR_ROLE\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"distributeReward\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"onWithdrawalCredentialsChanged\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"activateNodeOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_rewardAddress\",\"type\":\"address\"}],\"name\":\"setNodeOperatorRewardAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_fullInfo\",\"type\":\"bool\"}],\"name\":\"getNodeOperator\",\"outputs\":[{\"name\":\"active\",\"type\":\"bool\"},{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"rewardAddress\",\"type\":\"address\"},{\"name\":\"totalVettedValidators\",\"type\":\"uint64\"},{\"name\":\"totalExitedValidators\",\"type\":\"uint64\"},{\"name\":\"totalAddedValidators\",\"type\":\"uint64\"},{\"name\":\"totalDepositedValidators\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_locator\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"bytes32\"},{\"name\":\"_stuckPenaltyDelay\",\"type\":\"uint256\"}],\"name\":\"finalizeUpgrade_v2\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStakingModuleSummary\",\"outputs\":[{\"name\":\"totalExitedValidators\",\"type\":\"uint256\"},{\"name\":\"totalDepositedValidators\",\"type\":\"uint256\"},{\"name\":\"depositableValidatorsCount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorIds\",\"type\":\"bytes\"},{\"name\":\"_exitedValidatorsCounts\",\"type\":\"bytes\"}],\"name\":\"updateExitedValidatorsCount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorIds\",\"type\":\"bytes\"},{\"name\":\"_stuckValidatorsCounts\",\"type\":\"bytes\"}],\"name\":\"updateStuckValidatorsCount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"transferToVault\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_sender\",\"type\":\"address\"},{\"name\":\"_role\",\"type\":\"bytes32\"},{\"name\":\"_params\",\"type\":\"uint256[]\"}],\"name\":\"canPerform\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_refundedValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"updateRefundedValidatorsCount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getEVMScriptRegistry\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getNodeOperatorsCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_isTargetLimitActive\",\"type\":\"bool\"},{\"name\":\"_targetLimit\",\"type\":\"uint256\"}],\"name\":\"updateTargetValidatorsLimits\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_vettedSigningKeysCount\",\"type\":\"uint64\"}],\"name\":\"setNodeOperatorStakingLimit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorSummary\",\"outputs\":[{\"name\":\"targetLimitMode\",\"type\":\"uint256\"},{\"name\":\"targetValidatorsCount\",\"type\":\"uint256\"},{\"name\":\"stuckValidatorsCount\",\"type\":\"uint256\"},{\"name\":\"refundedValidatorsCount\",\"type\":\"uint256\"},{\"name\":\"stuckPenaltyEndTimestamp\",\"type\":\"uint256\"},{\"name\":\"totalExitedValidators\",\"type\":\"uint256\"},{\"name\":\"totalDepositedValidators\",\"type\":\"uint256\"},{\"name\":\"depositableValidatorsCount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"getSigningKey\",\"outputs\":[{\"name\":\"key\",\"type\":\"bytes\"},{\"name\":\"depositSignature\",\"type\":\"bytes\"},{\"name\":\"used\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_NODE_OPERATOR_NAME_LENGTH\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorIds\",\"type\":\"bytes\"},{\"name\":\"_vettedSigningKeysCounts\",\"type\":\"bytes\"}],\"name\":\"decreaseVettedSigningKeysCount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_depositsCount\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"obtainDepositData\",\"outputs\":[{\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"name\":\"signatures\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getKeysOpIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"kernel\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getLocator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"SET_NODE_OPERATOR_LIMIT_ROLE\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getTotalSigningKeyCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isPetrified\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_STUCK_PENALTY_DELAY\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"onExitedAndStuckValidatorsCountsUpdated\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_NODE_OPERATORS_COUNT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"removeSigningKeyOperatorBH\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_exitedValidatorsCount\",\"type\":\"uint256\"},{\"name\":\"_stuckValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"unsafeUpdateValidatorsCount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MANAGE_SIGNING_KEYS\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"isOperatorPenaltyCleared\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"rewardAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"stakingLimit\",\"type\":\"uint64\"}],\"name\":\"NodeOperatorAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"active\",\"type\":\"bool\"}],\"name\":\"NodeOperatorActiveSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"name\",\"type\":\"string\"}],\"name\":\"NodeOperatorNameSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"rewardAddress\",\"type\":\"address\"}],\"name\":\"NodeOperatorRewardAddressSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalKeysTrimmed\",\"type\":\"uint64\"}],\"name\":\"NodeOperatorTotalKeysTrimmed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"keysOpIndex\",\"type\":\"uint256\"}],\"name\":\"KeysOpIndexSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"moduleType\",\"type\":\"bytes32\"}],\"name\":\"StakingModuleTypeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"rewardAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"sharesAmount\",\"type\":\"uint256\"}],\"name\":\"RewardsDistributed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"state\",\"type\":\"uint8\"}],\"name\":\"RewardDistributionStateChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"locatorAddress\",\"type\":\"address\"}],\"name\":\"LocatorContractSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"approvedValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"VettedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"depositedValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"DepositedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"exitedValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"ExitedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"TotalSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"NonceChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"stuckPenaltyDelay\",\"type\":\"uint256\"}],\"name\":\"StuckPenaltyDelayChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"stuckValidatorsCount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"refundedValidatorsCount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"stuckPenaltyEndTimestamp\",\"type\":\"uint256\"}],\"name\":\"StuckPenaltyStateChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"targetValidatorsCount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"targetLimitMode\",\"type\":\"uint256\"}],\"name\":\"TargetValidatorsCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"recipientAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"sharesPenalizedAmount\",\"type\":\"uint256\"}],\"name\":\"NodeOperatorPenalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"NodeOperatorPenaltyCleared\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"ContractVersionSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"executor\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"script\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"input\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"returnData\",\"type\":\"bytes\"}],\"name\":\"ScriptResult\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"vault\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"RecoverToVault\",\"type\":\"event\"}]",
}

// OperatorRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use OperatorRegistryMetaData.ABI instead.
var OperatorRegistryABI = OperatorRegistryMetaData.ABI

// OperatorRegistry is an auto generated Go binding around an Ethereum contract.
type OperatorRegistry struct {
	OperatorRegistryCaller     // Read-only binding to the contract
	OperatorRegistryTransactor // Write-only binding to the contract
	OperatorRegistryFilterer   // Log filterer for contract events
}

// OperatorRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type OperatorRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperatorRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OperatorRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperatorRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OperatorRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperatorRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OperatorRegistrySession struct {
	Contract     *OperatorRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OperatorRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OperatorRegistryCallerSession struct {
	Contract *OperatorRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// OperatorRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OperatorRegistryTransactorSession struct {
	Contract     *OperatorRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// OperatorRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type OperatorRegistryRaw struct {
	Contract *OperatorRegistry // Generic contract binding to access the raw methods on
}

// OperatorRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OperatorRegistryCallerRaw struct {
	Contract *OperatorRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// OperatorRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OperatorRegistryTransactorRaw struct {
	Contract *OperatorRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOperatorRegistry creates a new instance of OperatorRegistry, bound to a specific deployed contract.
func NewOperatorRegistry(address common.Address, backend bind.ContractBackend) (*OperatorRegistry, error) {
	contract, err := bindOperatorRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistry{OperatorRegistryCaller: OperatorRegistryCaller{contract: contract}, OperatorRegistryTransactor: OperatorRegistryTransactor{contract: contract}, OperatorRegistryFilterer: OperatorRegistryFilterer{contract: contract}}, nil
}

// NewOperatorRegistryCaller creates a new read-only instance of OperatorRegistry, bound to a specific deployed contract.
func NewOperatorRegistryCaller(address common.Address, caller bind.ContractCaller) (*OperatorRegistryCaller, error) {
	contract, err := bindOperatorRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryCaller{contract: contract}, nil
}

// NewOperatorRegistryTransactor creates a new write-only instance of OperatorRegistry, bound to a specific deployed contract.
func NewOperatorRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*OperatorRegistryTransactor, error) {
	contract, err := bindOperatorRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryTransactor{contract: contract}, nil
}

// NewOperatorRegistryFilterer creates a new log filterer instance of OperatorRegistry, bound to a specific deployed contract.
func NewOperatorRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*OperatorRegistryFilterer, error) {
	contract, err := bindOperatorRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryFilterer{contract: contract}, nil
}

// bindOperatorRegistry binds a generic wrapper to an already deployed contract.
func bindOperatorRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OperatorRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperatorRegistry *OperatorRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperatorRegistry.Contract.OperatorRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperatorRegistry *OperatorRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OperatorRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperatorRegistry *OperatorRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OperatorRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperatorRegistry *OperatorRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperatorRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperatorRegistry *OperatorRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperatorRegistry *OperatorRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.contract.Transact(opts, method, params...)
}

// MANAGENODEOPERATORROLE is a free data retrieval call binding the contract method 0x8ece9995.
//
// Solidity: function MANAGE_NODE_OPERATOR_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCaller) MANAGENODEOPERATORROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "MANAGE_NODE_OPERATOR_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MANAGENODEOPERATORROLE is a free data retrieval call binding the contract method 0x8ece9995.
//
// Solidity: function MANAGE_NODE_OPERATOR_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistrySession) MANAGENODEOPERATORROLE() ([32]byte, error) {
	return _OperatorRegistry.Contract.MANAGENODEOPERATORROLE(&_OperatorRegistry.CallOpts)
}

// MANAGENODEOPERATORROLE is a free data retrieval call binding the contract method 0x8ece9995.
//
// Solidity: function MANAGE_NODE_OPERATOR_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCallerSession) MANAGENODEOPERATORROLE() ([32]byte, error) {
	return _OperatorRegistry.Contract.MANAGENODEOPERATORROLE(&_OperatorRegistry.CallOpts)
}

// MANAGESIGNINGKEYS is a free data retrieval call binding the contract method 0xf31bd9c1.
//
// Solidity: function MANAGE_SIGNING_KEYS() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCaller) MANAGESIGNINGKEYS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "MANAGE_SIGNING_KEYS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MANAGESIGNINGKEYS is a free data retrieval call binding the contract method 0xf31bd9c1.
//
// Solidity: function MANAGE_SIGNING_KEYS() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistrySession) MANAGESIGNINGKEYS() ([32]byte, error) {
	return _OperatorRegistry.Contract.MANAGESIGNINGKEYS(&_OperatorRegistry.CallOpts)
}

// MANAGESIGNINGKEYS is a free data retrieval call binding the contract method 0xf31bd9c1.
//
// Solidity: function MANAGE_SIGNING_KEYS() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCallerSession) MANAGESIGNINGKEYS() ([32]byte, error) {
	return _OperatorRegistry.Contract.MANAGESIGNINGKEYS(&_OperatorRegistry.CallOpts)
}

// MAXNODEOPERATORSCOUNT is a free data retrieval call binding the contract method 0xec5af3a4.
//
// Solidity: function MAX_NODE_OPERATORS_COUNT() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) MAXNODEOPERATORSCOUNT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "MAX_NODE_OPERATORS_COUNT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXNODEOPERATORSCOUNT is a free data retrieval call binding the contract method 0xec5af3a4.
//
// Solidity: function MAX_NODE_OPERATORS_COUNT() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) MAXNODEOPERATORSCOUNT() (*big.Int, error) {
	return _OperatorRegistry.Contract.MAXNODEOPERATORSCOUNT(&_OperatorRegistry.CallOpts)
}

// MAXNODEOPERATORSCOUNT is a free data retrieval call binding the contract method 0xec5af3a4.
//
// Solidity: function MAX_NODE_OPERATORS_COUNT() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) MAXNODEOPERATORSCOUNT() (*big.Int, error) {
	return _OperatorRegistry.Contract.MAXNODEOPERATORSCOUNT(&_OperatorRegistry.CallOpts)
}

// MAXNODEOPERATORNAMELENGTH is a free data retrieval call binding the contract method 0xb4971833.
//
// Solidity: function MAX_NODE_OPERATOR_NAME_LENGTH() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) MAXNODEOPERATORNAMELENGTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "MAX_NODE_OPERATOR_NAME_LENGTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXNODEOPERATORNAMELENGTH is a free data retrieval call binding the contract method 0xb4971833.
//
// Solidity: function MAX_NODE_OPERATOR_NAME_LENGTH() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) MAXNODEOPERATORNAMELENGTH() (*big.Int, error) {
	return _OperatorRegistry.Contract.MAXNODEOPERATORNAMELENGTH(&_OperatorRegistry.CallOpts)
}

// MAXNODEOPERATORNAMELENGTH is a free data retrieval call binding the contract method 0xb4971833.
//
// Solidity: function MAX_NODE_OPERATOR_NAME_LENGTH() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) MAXNODEOPERATORNAMELENGTH() (*big.Int, error) {
	return _OperatorRegistry.Contract.MAXNODEOPERATORNAMELENGTH(&_OperatorRegistry.CallOpts)
}

// MAXSTUCKPENALTYDELAY is a free data retrieval call binding the contract method 0xe204d09b.
//
// Solidity: function MAX_STUCK_PENALTY_DELAY() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) MAXSTUCKPENALTYDELAY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "MAX_STUCK_PENALTY_DELAY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXSTUCKPENALTYDELAY is a free data retrieval call binding the contract method 0xe204d09b.
//
// Solidity: function MAX_STUCK_PENALTY_DELAY() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) MAXSTUCKPENALTYDELAY() (*big.Int, error) {
	return _OperatorRegistry.Contract.MAXSTUCKPENALTYDELAY(&_OperatorRegistry.CallOpts)
}

// MAXSTUCKPENALTYDELAY is a free data retrieval call binding the contract method 0xe204d09b.
//
// Solidity: function MAX_STUCK_PENALTY_DELAY() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) MAXSTUCKPENALTYDELAY() (*big.Int, error) {
	return _OperatorRegistry.Contract.MAXSTUCKPENALTYDELAY(&_OperatorRegistry.CallOpts)
}

// SETNODEOPERATORLIMITROLE is a free data retrieval call binding the contract method 0xd8e71cd1.
//
// Solidity: function SET_NODE_OPERATOR_LIMIT_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCaller) SETNODEOPERATORLIMITROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "SET_NODE_OPERATOR_LIMIT_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SETNODEOPERATORLIMITROLE is a free data retrieval call binding the contract method 0xd8e71cd1.
//
// Solidity: function SET_NODE_OPERATOR_LIMIT_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistrySession) SETNODEOPERATORLIMITROLE() ([32]byte, error) {
	return _OperatorRegistry.Contract.SETNODEOPERATORLIMITROLE(&_OperatorRegistry.CallOpts)
}

// SETNODEOPERATORLIMITROLE is a free data retrieval call binding the contract method 0xd8e71cd1.
//
// Solidity: function SET_NODE_OPERATOR_LIMIT_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCallerSession) SETNODEOPERATORLIMITROLE() ([32]byte, error) {
	return _OperatorRegistry.Contract.SETNODEOPERATORLIMITROLE(&_OperatorRegistry.CallOpts)
}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCaller) STAKINGROUTERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "STAKING_ROUTER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistrySession) STAKINGROUTERROLE() ([32]byte, error) {
	return _OperatorRegistry.Contract.STAKINGROUTERROLE(&_OperatorRegistry.CallOpts)
}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCallerSession) STAKINGROUTERROLE() ([32]byte, error) {
	return _OperatorRegistry.Contract.STAKINGROUTERROLE(&_OperatorRegistry.CallOpts)
}

// AllowRecoverability is a free data retrieval call binding the contract method 0x7e7db6e1.
//
// Solidity: function allowRecoverability(address token) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCaller) AllowRecoverability(opts *bind.CallOpts, token common.Address) (bool, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "allowRecoverability", token)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllowRecoverability is a free data retrieval call binding the contract method 0x7e7db6e1.
//
// Solidity: function allowRecoverability(address token) view returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) AllowRecoverability(token common.Address) (bool, error) {
	return _OperatorRegistry.Contract.AllowRecoverability(&_OperatorRegistry.CallOpts, token)
}

// AllowRecoverability is a free data retrieval call binding the contract method 0x7e7db6e1.
//
// Solidity: function allowRecoverability(address token) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCallerSession) AllowRecoverability(token common.Address) (bool, error) {
	return _OperatorRegistry.Contract.AllowRecoverability(&_OperatorRegistry.CallOpts, token)
}

// AppId is a free data retrieval call binding the contract method 0x80afdea8.
//
// Solidity: function appId() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCaller) AppId(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "appId")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// AppId is a free data retrieval call binding the contract method 0x80afdea8.
//
// Solidity: function appId() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistrySession) AppId() ([32]byte, error) {
	return _OperatorRegistry.Contract.AppId(&_OperatorRegistry.CallOpts)
}

// AppId is a free data retrieval call binding the contract method 0x80afdea8.
//
// Solidity: function appId() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCallerSession) AppId() ([32]byte, error) {
	return _OperatorRegistry.Contract.AppId(&_OperatorRegistry.CallOpts)
}

// CanPerform is a free data retrieval call binding the contract method 0xa1658fad.
//
// Solidity: function canPerform(address _sender, bytes32 _role, uint256[] _params) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCaller) CanPerform(opts *bind.CallOpts, _sender common.Address, _role [32]byte, _params []*big.Int) (bool, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "canPerform", _sender, _role, _params)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanPerform is a free data retrieval call binding the contract method 0xa1658fad.
//
// Solidity: function canPerform(address _sender, bytes32 _role, uint256[] _params) view returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) CanPerform(_sender common.Address, _role [32]byte, _params []*big.Int) (bool, error) {
	return _OperatorRegistry.Contract.CanPerform(&_OperatorRegistry.CallOpts, _sender, _role, _params)
}

// CanPerform is a free data retrieval call binding the contract method 0xa1658fad.
//
// Solidity: function canPerform(address _sender, bytes32 _role, uint256[] _params) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCallerSession) CanPerform(_sender common.Address, _role [32]byte, _params []*big.Int) (bool, error) {
	return _OperatorRegistry.Contract.CanPerform(&_OperatorRegistry.CallOpts, _sender, _role, _params)
}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetActiveNodeOperatorsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getActiveNodeOperatorsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetActiveNodeOperatorsCount() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetActiveNodeOperatorsCount(&_OperatorRegistry.CallOpts)
}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetActiveNodeOperatorsCount() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetActiveNodeOperatorsCount(&_OperatorRegistry.CallOpts)
}

// GetContractVersion is a free data retrieval call binding the contract method 0x8aa10435.
//
// Solidity: function getContractVersion() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetContractVersion(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getContractVersion")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetContractVersion is a free data retrieval call binding the contract method 0x8aa10435.
//
// Solidity: function getContractVersion() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetContractVersion() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetContractVersion(&_OperatorRegistry.CallOpts)
}

// GetContractVersion is a free data retrieval call binding the contract method 0x8aa10435.
//
// Solidity: function getContractVersion() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetContractVersion() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetContractVersion(&_OperatorRegistry.CallOpts)
}

// GetEVMScriptExecutor is a free data retrieval call binding the contract method 0x2914b9bd.
//
// Solidity: function getEVMScriptExecutor(bytes _script) view returns(address)
func (_OperatorRegistry *OperatorRegistryCaller) GetEVMScriptExecutor(opts *bind.CallOpts, _script []byte) (common.Address, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getEVMScriptExecutor", _script)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetEVMScriptExecutor is a free data retrieval call binding the contract method 0x2914b9bd.
//
// Solidity: function getEVMScriptExecutor(bytes _script) view returns(address)
func (_OperatorRegistry *OperatorRegistrySession) GetEVMScriptExecutor(_script []byte) (common.Address, error) {
	return _OperatorRegistry.Contract.GetEVMScriptExecutor(&_OperatorRegistry.CallOpts, _script)
}

// GetEVMScriptExecutor is a free data retrieval call binding the contract method 0x2914b9bd.
//
// Solidity: function getEVMScriptExecutor(bytes _script) view returns(address)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetEVMScriptExecutor(_script []byte) (common.Address, error) {
	return _OperatorRegistry.Contract.GetEVMScriptExecutor(&_OperatorRegistry.CallOpts, _script)
}

// GetEVMScriptRegistry is a free data retrieval call binding the contract method 0xa479e508.
//
// Solidity: function getEVMScriptRegistry() view returns(address)
func (_OperatorRegistry *OperatorRegistryCaller) GetEVMScriptRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getEVMScriptRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetEVMScriptRegistry is a free data retrieval call binding the contract method 0xa479e508.
//
// Solidity: function getEVMScriptRegistry() view returns(address)
func (_OperatorRegistry *OperatorRegistrySession) GetEVMScriptRegistry() (common.Address, error) {
	return _OperatorRegistry.Contract.GetEVMScriptRegistry(&_OperatorRegistry.CallOpts)
}

// GetEVMScriptRegistry is a free data retrieval call binding the contract method 0xa479e508.
//
// Solidity: function getEVMScriptRegistry() view returns(address)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetEVMScriptRegistry() (common.Address, error) {
	return _OperatorRegistry.Contract.GetEVMScriptRegistry(&_OperatorRegistry.CallOpts)
}

// GetInitializationBlock is a free data retrieval call binding the contract method 0x8b3dd749.
//
// Solidity: function getInitializationBlock() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetInitializationBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getInitializationBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetInitializationBlock is a free data retrieval call binding the contract method 0x8b3dd749.
//
// Solidity: function getInitializationBlock() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetInitializationBlock() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetInitializationBlock(&_OperatorRegistry.CallOpts)
}

// GetInitializationBlock is a free data retrieval call binding the contract method 0x8b3dd749.
//
// Solidity: function getInitializationBlock() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetInitializationBlock() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetInitializationBlock(&_OperatorRegistry.CallOpts)
}

// GetKeysOpIndex is a free data retrieval call binding the contract method 0xd07442f1.
//
// Solidity: function getKeysOpIndex() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetKeysOpIndex(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getKeysOpIndex")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetKeysOpIndex is a free data retrieval call binding the contract method 0xd07442f1.
//
// Solidity: function getKeysOpIndex() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetKeysOpIndex() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetKeysOpIndex(&_OperatorRegistry.CallOpts)
}

// GetKeysOpIndex is a free data retrieval call binding the contract method 0xd07442f1.
//
// Solidity: function getKeysOpIndex() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetKeysOpIndex() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetKeysOpIndex(&_OperatorRegistry.CallOpts)
}

// GetLocator is a free data retrieval call binding the contract method 0xd8343dcb.
//
// Solidity: function getLocator() view returns(address)
func (_OperatorRegistry *OperatorRegistryCaller) GetLocator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getLocator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetLocator is a free data retrieval call binding the contract method 0xd8343dcb.
//
// Solidity: function getLocator() view returns(address)
func (_OperatorRegistry *OperatorRegistrySession) GetLocator() (common.Address, error) {
	return _OperatorRegistry.Contract.GetLocator(&_OperatorRegistry.CallOpts)
}

// GetLocator is a free data retrieval call binding the contract method 0xd8343dcb.
//
// Solidity: function getLocator() view returns(address)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetLocator() (common.Address, error) {
	return _OperatorRegistry.Contract.GetLocator(&_OperatorRegistry.CallOpts)
}

// GetNodeOperator is a free data retrieval call binding the contract method 0x9a56983c.
//
// Solidity: function getNodeOperator(uint256 _nodeOperatorId, bool _fullInfo) view returns(bool active, string name, address rewardAddress, uint64 totalVettedValidators, uint64 totalExitedValidators, uint64 totalAddedValidators, uint64 totalDepositedValidators)
func (_OperatorRegistry *OperatorRegistryCaller) GetNodeOperator(opts *bind.CallOpts, _nodeOperatorId *big.Int, _fullInfo bool) (struct {
	Active                   bool
	Name                     string
	RewardAddress            common.Address
	TotalVettedValidators    uint64
	TotalExitedValidators    uint64
	TotalAddedValidators     uint64
	TotalDepositedValidators uint64
}, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getNodeOperator", _nodeOperatorId, _fullInfo)

	outstruct := new(struct {
		Active                   bool
		Name                     string
		RewardAddress            common.Address
		TotalVettedValidators    uint64
		TotalExitedValidators    uint64
		TotalAddedValidators     uint64
		TotalDepositedValidators uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Active = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.RewardAddress = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.TotalVettedValidators = *abi.ConvertType(out[3], new(uint64)).(*uint64)
	outstruct.TotalExitedValidators = *abi.ConvertType(out[4], new(uint64)).(*uint64)
	outstruct.TotalAddedValidators = *abi.ConvertType(out[5], new(uint64)).(*uint64)
	outstruct.TotalDepositedValidators = *abi.ConvertType(out[6], new(uint64)).(*uint64)

	return *outstruct, err

}

// GetNodeOperator is a free data retrieval call binding the contract method 0x9a56983c.
//
// Solidity: function getNodeOperator(uint256 _nodeOperatorId, bool _fullInfo) view returns(bool active, string name, address rewardAddress, uint64 totalVettedValidators, uint64 totalExitedValidators, uint64 totalAddedValidators, uint64 totalDepositedValidators)
func (_OperatorRegistry *OperatorRegistrySession) GetNodeOperator(_nodeOperatorId *big.Int, _fullInfo bool) (struct {
	Active                   bool
	Name                     string
	RewardAddress            common.Address
	TotalVettedValidators    uint64
	TotalExitedValidators    uint64
	TotalAddedValidators     uint64
	TotalDepositedValidators uint64
}, error) {
	return _OperatorRegistry.Contract.GetNodeOperator(&_OperatorRegistry.CallOpts, _nodeOperatorId, _fullInfo)
}

// GetNodeOperator is a free data retrieval call binding the contract method 0x9a56983c.
//
// Solidity: function getNodeOperator(uint256 _nodeOperatorId, bool _fullInfo) view returns(bool active, string name, address rewardAddress, uint64 totalVettedValidators, uint64 totalExitedValidators, uint64 totalAddedValidators, uint64 totalDepositedValidators)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetNodeOperator(_nodeOperatorId *big.Int, _fullInfo bool) (struct {
	Active                   bool
	Name                     string
	RewardAddress            common.Address
	TotalVettedValidators    uint64
	TotalExitedValidators    uint64
	TotalAddedValidators     uint64
	TotalDepositedValidators uint64
}, error) {
	return _OperatorRegistry.Contract.GetNodeOperator(&_OperatorRegistry.CallOpts, _nodeOperatorId, _fullInfo)
}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 _offset, uint256 _limit) view returns(uint256[] nodeOperatorIds)
func (_OperatorRegistry *OperatorRegistryCaller) GetNodeOperatorIds(opts *bind.CallOpts, _offset *big.Int, _limit *big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getNodeOperatorIds", _offset, _limit)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 _offset, uint256 _limit) view returns(uint256[] nodeOperatorIds)
func (_OperatorRegistry *OperatorRegistrySession) GetNodeOperatorIds(_offset *big.Int, _limit *big.Int) ([]*big.Int, error) {
	return _OperatorRegistry.Contract.GetNodeOperatorIds(&_OperatorRegistry.CallOpts, _offset, _limit)
}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 _offset, uint256 _limit) view returns(uint256[] nodeOperatorIds)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetNodeOperatorIds(_offset *big.Int, _limit *big.Int) ([]*big.Int, error) {
	return _OperatorRegistry.Contract.GetNodeOperatorIds(&_OperatorRegistry.CallOpts, _offset, _limit)
}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCaller) GetNodeOperatorIsActive(opts *bind.CallOpts, _nodeOperatorId *big.Int) (bool, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getNodeOperatorIsActive", _nodeOperatorId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) GetNodeOperatorIsActive(_nodeOperatorId *big.Int) (bool, error) {
	return _OperatorRegistry.Contract.GetNodeOperatorIsActive(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetNodeOperatorIsActive(_nodeOperatorId *big.Int) (bool, error) {
	return _OperatorRegistry.Contract.GetNodeOperatorIsActive(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// GetNodeOperatorSummary is a free data retrieval call binding the contract method 0xb3076c3c.
//
// Solidity: function getNodeOperatorSummary(uint256 _nodeOperatorId) view returns(uint256 targetLimitMode, uint256 targetValidatorsCount, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp, uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_OperatorRegistry *OperatorRegistryCaller) GetNodeOperatorSummary(opts *bind.CallOpts, _nodeOperatorId *big.Int) (struct {
	TargetLimitMode            *big.Int
	TargetValidatorsCount      *big.Int
	StuckValidatorsCount       *big.Int
	RefundedValidatorsCount    *big.Int
	StuckPenaltyEndTimestamp   *big.Int
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getNodeOperatorSummary", _nodeOperatorId)

	outstruct := new(struct {
		TargetLimitMode            *big.Int
		TargetValidatorsCount      *big.Int
		StuckValidatorsCount       *big.Int
		RefundedValidatorsCount    *big.Int
		StuckPenaltyEndTimestamp   *big.Int
		TotalExitedValidators      *big.Int
		TotalDepositedValidators   *big.Int
		DepositableValidatorsCount *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TargetLimitMode = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TargetValidatorsCount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.StuckValidatorsCount = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.RefundedValidatorsCount = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.StuckPenaltyEndTimestamp = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.TotalExitedValidators = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.TotalDepositedValidators = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.DepositableValidatorsCount = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetNodeOperatorSummary is a free data retrieval call binding the contract method 0xb3076c3c.
//
// Solidity: function getNodeOperatorSummary(uint256 _nodeOperatorId) view returns(uint256 targetLimitMode, uint256 targetValidatorsCount, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp, uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_OperatorRegistry *OperatorRegistrySession) GetNodeOperatorSummary(_nodeOperatorId *big.Int) (struct {
	TargetLimitMode            *big.Int
	TargetValidatorsCount      *big.Int
	StuckValidatorsCount       *big.Int
	RefundedValidatorsCount    *big.Int
	StuckPenaltyEndTimestamp   *big.Int
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _OperatorRegistry.Contract.GetNodeOperatorSummary(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// GetNodeOperatorSummary is a free data retrieval call binding the contract method 0xb3076c3c.
//
// Solidity: function getNodeOperatorSummary(uint256 _nodeOperatorId) view returns(uint256 targetLimitMode, uint256 targetValidatorsCount, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp, uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetNodeOperatorSummary(_nodeOperatorId *big.Int) (struct {
	TargetLimitMode            *big.Int
	TargetValidatorsCount      *big.Int
	StuckValidatorsCount       *big.Int
	RefundedValidatorsCount    *big.Int
	StuckPenaltyEndTimestamp   *big.Int
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _OperatorRegistry.Contract.GetNodeOperatorSummary(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetNodeOperatorsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getNodeOperatorsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetNodeOperatorsCount() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetNodeOperatorsCount(&_OperatorRegistry.CallOpts)
}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetNodeOperatorsCount() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetNodeOperatorsCount(&_OperatorRegistry.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetNonce() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetNonce(&_OperatorRegistry.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetNonce() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetNonce(&_OperatorRegistry.CallOpts)
}

// GetRecoveryVault is a free data retrieval call binding the contract method 0x32f0a3b5.
//
// Solidity: function getRecoveryVault() view returns(address)
func (_OperatorRegistry *OperatorRegistryCaller) GetRecoveryVault(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getRecoveryVault")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRecoveryVault is a free data retrieval call binding the contract method 0x32f0a3b5.
//
// Solidity: function getRecoveryVault() view returns(address)
func (_OperatorRegistry *OperatorRegistrySession) GetRecoveryVault() (common.Address, error) {
	return _OperatorRegistry.Contract.GetRecoveryVault(&_OperatorRegistry.CallOpts)
}

// GetRecoveryVault is a free data retrieval call binding the contract method 0x32f0a3b5.
//
// Solidity: function getRecoveryVault() view returns(address)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetRecoveryVault() (common.Address, error) {
	return _OperatorRegistry.Contract.GetRecoveryVault(&_OperatorRegistry.CallOpts)
}

// GetRewardDistributionState is a free data retrieval call binding the contract method 0x6f817294.
//
// Solidity: function getRewardDistributionState() view returns(uint8)
func (_OperatorRegistry *OperatorRegistryCaller) GetRewardDistributionState(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getRewardDistributionState")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetRewardDistributionState is a free data retrieval call binding the contract method 0x6f817294.
//
// Solidity: function getRewardDistributionState() view returns(uint8)
func (_OperatorRegistry *OperatorRegistrySession) GetRewardDistributionState() (uint8, error) {
	return _OperatorRegistry.Contract.GetRewardDistributionState(&_OperatorRegistry.CallOpts)
}

// GetRewardDistributionState is a free data retrieval call binding the contract method 0x6f817294.
//
// Solidity: function getRewardDistributionState() view returns(uint8)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetRewardDistributionState() (uint8, error) {
	return _OperatorRegistry.Contract.GetRewardDistributionState(&_OperatorRegistry.CallOpts)
}

// GetRewardsDistribution is a free data retrieval call binding the contract method 0x62dcfda1.
//
// Solidity: function getRewardsDistribution(uint256 _totalRewardShares) view returns(address[] recipients, uint256[] shares, bool[] penalized)
func (_OperatorRegistry *OperatorRegistryCaller) GetRewardsDistribution(opts *bind.CallOpts, _totalRewardShares *big.Int) (struct {
	Recipients []common.Address
	Shares     []*big.Int
	Penalized  []bool
}, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getRewardsDistribution", _totalRewardShares)

	outstruct := new(struct {
		Recipients []common.Address
		Shares     []*big.Int
		Penalized  []bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Recipients = *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	outstruct.Shares = *abi.ConvertType(out[1], new([]*big.Int)).(*[]*big.Int)
	outstruct.Penalized = *abi.ConvertType(out[2], new([]bool)).(*[]bool)

	return *outstruct, err

}

// GetRewardsDistribution is a free data retrieval call binding the contract method 0x62dcfda1.
//
// Solidity: function getRewardsDistribution(uint256 _totalRewardShares) view returns(address[] recipients, uint256[] shares, bool[] penalized)
func (_OperatorRegistry *OperatorRegistrySession) GetRewardsDistribution(_totalRewardShares *big.Int) (struct {
	Recipients []common.Address
	Shares     []*big.Int
	Penalized  []bool
}, error) {
	return _OperatorRegistry.Contract.GetRewardsDistribution(&_OperatorRegistry.CallOpts, _totalRewardShares)
}

// GetRewardsDistribution is a free data retrieval call binding the contract method 0x62dcfda1.
//
// Solidity: function getRewardsDistribution(uint256 _totalRewardShares) view returns(address[] recipients, uint256[] shares, bool[] penalized)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetRewardsDistribution(_totalRewardShares *big.Int) (struct {
	Recipients []common.Address
	Shares     []*big.Int
	Penalized  []bool
}, error) {
	return _OperatorRegistry.Contract.GetRewardsDistribution(&_OperatorRegistry.CallOpts, _totalRewardShares)
}

// GetSigningKey is a free data retrieval call binding the contract method 0xb449402a.
//
// Solidity: function getSigningKey(uint256 _nodeOperatorId, uint256 _index) view returns(bytes key, bytes depositSignature, bool used)
func (_OperatorRegistry *OperatorRegistryCaller) GetSigningKey(opts *bind.CallOpts, _nodeOperatorId *big.Int, _index *big.Int) (struct {
	Key              []byte
	DepositSignature []byte
	Used             bool
}, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getSigningKey", _nodeOperatorId, _index)

	outstruct := new(struct {
		Key              []byte
		DepositSignature []byte
		Used             bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Key = *abi.ConvertType(out[0], new([]byte)).(*[]byte)
	outstruct.DepositSignature = *abi.ConvertType(out[1], new([]byte)).(*[]byte)
	outstruct.Used = *abi.ConvertType(out[2], new(bool)).(*bool)

	return *outstruct, err

}

// GetSigningKey is a free data retrieval call binding the contract method 0xb449402a.
//
// Solidity: function getSigningKey(uint256 _nodeOperatorId, uint256 _index) view returns(bytes key, bytes depositSignature, bool used)
func (_OperatorRegistry *OperatorRegistrySession) GetSigningKey(_nodeOperatorId *big.Int, _index *big.Int) (struct {
	Key              []byte
	DepositSignature []byte
	Used             bool
}, error) {
	return _OperatorRegistry.Contract.GetSigningKey(&_OperatorRegistry.CallOpts, _nodeOperatorId, _index)
}

// GetSigningKey is a free data retrieval call binding the contract method 0xb449402a.
//
// Solidity: function getSigningKey(uint256 _nodeOperatorId, uint256 _index) view returns(bytes key, bytes depositSignature, bool used)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetSigningKey(_nodeOperatorId *big.Int, _index *big.Int) (struct {
	Key              []byte
	DepositSignature []byte
	Used             bool
}, error) {
	return _OperatorRegistry.Contract.GetSigningKey(&_OperatorRegistry.CallOpts, _nodeOperatorId, _index)
}

// GetSigningKeys is a free data retrieval call binding the contract method 0x59e25c12.
//
// Solidity: function getSigningKeys(uint256 _nodeOperatorId, uint256 _offset, uint256 _limit) view returns(bytes pubkeys, bytes signatures, bool[] used)
func (_OperatorRegistry *OperatorRegistryCaller) GetSigningKeys(opts *bind.CallOpts, _nodeOperatorId *big.Int, _offset *big.Int, _limit *big.Int) (struct {
	Pubkeys    []byte
	Signatures []byte
	Used       []bool
}, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getSigningKeys", _nodeOperatorId, _offset, _limit)

	outstruct := new(struct {
		Pubkeys    []byte
		Signatures []byte
		Used       []bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Pubkeys = *abi.ConvertType(out[0], new([]byte)).(*[]byte)
	outstruct.Signatures = *abi.ConvertType(out[1], new([]byte)).(*[]byte)
	outstruct.Used = *abi.ConvertType(out[2], new([]bool)).(*[]bool)

	return *outstruct, err

}

// GetSigningKeys is a free data retrieval call binding the contract method 0x59e25c12.
//
// Solidity: function getSigningKeys(uint256 _nodeOperatorId, uint256 _offset, uint256 _limit) view returns(bytes pubkeys, bytes signatures, bool[] used)
func (_OperatorRegistry *OperatorRegistrySession) GetSigningKeys(_nodeOperatorId *big.Int, _offset *big.Int, _limit *big.Int) (struct {
	Pubkeys    []byte
	Signatures []byte
	Used       []bool
}, error) {
	return _OperatorRegistry.Contract.GetSigningKeys(&_OperatorRegistry.CallOpts, _nodeOperatorId, _offset, _limit)
}

// GetSigningKeys is a free data retrieval call binding the contract method 0x59e25c12.
//
// Solidity: function getSigningKeys(uint256 _nodeOperatorId, uint256 _offset, uint256 _limit) view returns(bytes pubkeys, bytes signatures, bool[] used)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetSigningKeys(_nodeOperatorId *big.Int, _offset *big.Int, _limit *big.Int) (struct {
	Pubkeys    []byte
	Signatures []byte
	Used       []bool
}, error) {
	return _OperatorRegistry.Contract.GetSigningKeys(&_OperatorRegistry.CallOpts, _nodeOperatorId, _offset, _limit)
}

// GetStakingModuleSummary is a free data retrieval call binding the contract method 0x9abddf09.
//
// Solidity: function getStakingModuleSummary() view returns(uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_OperatorRegistry *OperatorRegistryCaller) GetStakingModuleSummary(opts *bind.CallOpts) (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getStakingModuleSummary")

	outstruct := new(struct {
		TotalExitedValidators      *big.Int
		TotalDepositedValidators   *big.Int
		DepositableValidatorsCount *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TotalExitedValidators = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TotalDepositedValidators = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.DepositableValidatorsCount = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetStakingModuleSummary is a free data retrieval call binding the contract method 0x9abddf09.
//
// Solidity: function getStakingModuleSummary() view returns(uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_OperatorRegistry *OperatorRegistrySession) GetStakingModuleSummary() (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _OperatorRegistry.Contract.GetStakingModuleSummary(&_OperatorRegistry.CallOpts)
}

// GetStakingModuleSummary is a free data retrieval call binding the contract method 0x9abddf09.
//
// Solidity: function getStakingModuleSummary() view returns(uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetStakingModuleSummary() (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _OperatorRegistry.Contract.GetStakingModuleSummary(&_OperatorRegistry.CallOpts)
}

// GetStuckPenaltyDelay is a free data retrieval call binding the contract method 0x6da7d0a7.
//
// Solidity: function getStuckPenaltyDelay() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetStuckPenaltyDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getStuckPenaltyDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetStuckPenaltyDelay is a free data retrieval call binding the contract method 0x6da7d0a7.
//
// Solidity: function getStuckPenaltyDelay() view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetStuckPenaltyDelay() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetStuckPenaltyDelay(&_OperatorRegistry.CallOpts)
}

// GetStuckPenaltyDelay is a free data retrieval call binding the contract method 0x6da7d0a7.
//
// Solidity: function getStuckPenaltyDelay() view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetStuckPenaltyDelay() (*big.Int, error) {
	return _OperatorRegistry.Contract.GetStuckPenaltyDelay(&_OperatorRegistry.CallOpts)
}

// GetTotalSigningKeyCount is a free data retrieval call binding the contract method 0xdb9887ea.
//
// Solidity: function getTotalSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetTotalSigningKeyCount(opts *bind.CallOpts, _nodeOperatorId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getTotalSigningKeyCount", _nodeOperatorId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalSigningKeyCount is a free data retrieval call binding the contract method 0xdb9887ea.
//
// Solidity: function getTotalSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetTotalSigningKeyCount(_nodeOperatorId *big.Int) (*big.Int, error) {
	return _OperatorRegistry.Contract.GetTotalSigningKeyCount(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// GetTotalSigningKeyCount is a free data retrieval call binding the contract method 0xdb9887ea.
//
// Solidity: function getTotalSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetTotalSigningKeyCount(_nodeOperatorId *big.Int) (*big.Int, error) {
	return _OperatorRegistry.Contract.GetTotalSigningKeyCount(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCaller) GetType(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getType")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistrySession) GetType() ([32]byte, error) {
	return _OperatorRegistry.Contract.GetType(&_OperatorRegistry.CallOpts)
}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetType() ([32]byte, error) {
	return _OperatorRegistry.Contract.GetType(&_OperatorRegistry.CallOpts)
}

// GetUnusedSigningKeyCount is a free data retrieval call binding the contract method 0x8ca7c052.
//
// Solidity: function getUnusedSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCaller) GetUnusedSigningKeyCount(opts *bind.CallOpts, _nodeOperatorId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "getUnusedSigningKeyCount", _nodeOperatorId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUnusedSigningKeyCount is a free data retrieval call binding the contract method 0x8ca7c052.
//
// Solidity: function getUnusedSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_OperatorRegistry *OperatorRegistrySession) GetUnusedSigningKeyCount(_nodeOperatorId *big.Int) (*big.Int, error) {
	return _OperatorRegistry.Contract.GetUnusedSigningKeyCount(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// GetUnusedSigningKeyCount is a free data retrieval call binding the contract method 0x8ca7c052.
//
// Solidity: function getUnusedSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_OperatorRegistry *OperatorRegistryCallerSession) GetUnusedSigningKeyCount(_nodeOperatorId *big.Int) (*big.Int, error) {
	return _OperatorRegistry.Contract.GetUnusedSigningKeyCount(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// HasInitialized is a free data retrieval call binding the contract method 0x0803fac0.
//
// Solidity: function hasInitialized() view returns(bool)
func (_OperatorRegistry *OperatorRegistryCaller) HasInitialized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "hasInitialized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasInitialized is a free data retrieval call binding the contract method 0x0803fac0.
//
// Solidity: function hasInitialized() view returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) HasInitialized() (bool, error) {
	return _OperatorRegistry.Contract.HasInitialized(&_OperatorRegistry.CallOpts)
}

// HasInitialized is a free data retrieval call binding the contract method 0x0803fac0.
//
// Solidity: function hasInitialized() view returns(bool)
func (_OperatorRegistry *OperatorRegistryCallerSession) HasInitialized() (bool, error) {
	return _OperatorRegistry.Contract.HasInitialized(&_OperatorRegistry.CallOpts)
}

// IsOperatorPenalized is a free data retrieval call binding the contract method 0x75049ad8.
//
// Solidity: function isOperatorPenalized(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCaller) IsOperatorPenalized(opts *bind.CallOpts, _nodeOperatorId *big.Int) (bool, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "isOperatorPenalized", _nodeOperatorId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorPenalized is a free data retrieval call binding the contract method 0x75049ad8.
//
// Solidity: function isOperatorPenalized(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) IsOperatorPenalized(_nodeOperatorId *big.Int) (bool, error) {
	return _OperatorRegistry.Contract.IsOperatorPenalized(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// IsOperatorPenalized is a free data retrieval call binding the contract method 0x75049ad8.
//
// Solidity: function isOperatorPenalized(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCallerSession) IsOperatorPenalized(_nodeOperatorId *big.Int) (bool, error) {
	return _OperatorRegistry.Contract.IsOperatorPenalized(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// IsOperatorPenaltyCleared is a free data retrieval call binding the contract method 0xfbc77ef1.
//
// Solidity: function isOperatorPenaltyCleared(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCaller) IsOperatorPenaltyCleared(opts *bind.CallOpts, _nodeOperatorId *big.Int) (bool, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "isOperatorPenaltyCleared", _nodeOperatorId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorPenaltyCleared is a free data retrieval call binding the contract method 0xfbc77ef1.
//
// Solidity: function isOperatorPenaltyCleared(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) IsOperatorPenaltyCleared(_nodeOperatorId *big.Int) (bool, error) {
	return _OperatorRegistry.Contract.IsOperatorPenaltyCleared(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// IsOperatorPenaltyCleared is a free data retrieval call binding the contract method 0xfbc77ef1.
//
// Solidity: function isOperatorPenaltyCleared(uint256 _nodeOperatorId) view returns(bool)
func (_OperatorRegistry *OperatorRegistryCallerSession) IsOperatorPenaltyCleared(_nodeOperatorId *big.Int) (bool, error) {
	return _OperatorRegistry.Contract.IsOperatorPenaltyCleared(&_OperatorRegistry.CallOpts, _nodeOperatorId)
}

// IsPetrified is a free data retrieval call binding the contract method 0xde4796ed.
//
// Solidity: function isPetrified() view returns(bool)
func (_OperatorRegistry *OperatorRegistryCaller) IsPetrified(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "isPetrified")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsPetrified is a free data retrieval call binding the contract method 0xde4796ed.
//
// Solidity: function isPetrified() view returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) IsPetrified() (bool, error) {
	return _OperatorRegistry.Contract.IsPetrified(&_OperatorRegistry.CallOpts)
}

// IsPetrified is a free data retrieval call binding the contract method 0xde4796ed.
//
// Solidity: function isPetrified() view returns(bool)
func (_OperatorRegistry *OperatorRegistryCallerSession) IsPetrified() (bool, error) {
	return _OperatorRegistry.Contract.IsPetrified(&_OperatorRegistry.CallOpts)
}

// Kernel is a free data retrieval call binding the contract method 0xd4aae0c4.
//
// Solidity: function kernel() view returns(address)
func (_OperatorRegistry *OperatorRegistryCaller) Kernel(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OperatorRegistry.contract.Call(opts, &out, "kernel")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Kernel is a free data retrieval call binding the contract method 0xd4aae0c4.
//
// Solidity: function kernel() view returns(address)
func (_OperatorRegistry *OperatorRegistrySession) Kernel() (common.Address, error) {
	return _OperatorRegistry.Contract.Kernel(&_OperatorRegistry.CallOpts)
}

// Kernel is a free data retrieval call binding the contract method 0xd4aae0c4.
//
// Solidity: function kernel() view returns(address)
func (_OperatorRegistry *OperatorRegistryCallerSession) Kernel() (common.Address, error) {
	return _OperatorRegistry.Contract.Kernel(&_OperatorRegistry.CallOpts)
}

// ActivateNodeOperator is a paid mutator transaction binding the contract method 0x91dcd6b2.
//
// Solidity: function activateNodeOperator(uint256 _nodeOperatorId) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) ActivateNodeOperator(opts *bind.TransactOpts, _nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "activateNodeOperator", _nodeOperatorId)
}

// ActivateNodeOperator is a paid mutator transaction binding the contract method 0x91dcd6b2.
//
// Solidity: function activateNodeOperator(uint256 _nodeOperatorId) returns()
func (_OperatorRegistry *OperatorRegistrySession) ActivateNodeOperator(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.ActivateNodeOperator(&_OperatorRegistry.TransactOpts, _nodeOperatorId)
}

// ActivateNodeOperator is a paid mutator transaction binding the contract method 0x91dcd6b2.
//
// Solidity: function activateNodeOperator(uint256 _nodeOperatorId) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) ActivateNodeOperator(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.ActivateNodeOperator(&_OperatorRegistry.TransactOpts, _nodeOperatorId)
}

// AddNodeOperator is a paid mutator transaction binding the contract method 0x85fa63d7.
//
// Solidity: function addNodeOperator(string _name, address _rewardAddress) returns(uint256 id)
func (_OperatorRegistry *OperatorRegistryTransactor) AddNodeOperator(opts *bind.TransactOpts, _name string, _rewardAddress common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "addNodeOperator", _name, _rewardAddress)
}

// AddNodeOperator is a paid mutator transaction binding the contract method 0x85fa63d7.
//
// Solidity: function addNodeOperator(string _name, address _rewardAddress) returns(uint256 id)
func (_OperatorRegistry *OperatorRegistrySession) AddNodeOperator(_name string, _rewardAddress common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.AddNodeOperator(&_OperatorRegistry.TransactOpts, _name, _rewardAddress)
}

// AddNodeOperator is a paid mutator transaction binding the contract method 0x85fa63d7.
//
// Solidity: function addNodeOperator(string _name, address _rewardAddress) returns(uint256 id)
func (_OperatorRegistry *OperatorRegistryTransactorSession) AddNodeOperator(_name string, _rewardAddress common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.AddNodeOperator(&_OperatorRegistry.TransactOpts, _name, _rewardAddress)
}

// AddSigningKeys is a paid mutator transaction binding the contract method 0x096b7b35.
//
// Solidity: function addSigningKeys(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) AddSigningKeys(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "addSigningKeys", _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeys is a paid mutator transaction binding the contract method 0x096b7b35.
//
// Solidity: function addSigningKeys(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_OperatorRegistry *OperatorRegistrySession) AddSigningKeys(_nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.AddSigningKeys(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeys is a paid mutator transaction binding the contract method 0x096b7b35.
//
// Solidity: function addSigningKeys(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) AddSigningKeys(_nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.AddSigningKeys(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x805911ae.
//
// Solidity: function addSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) AddSigningKeysOperatorBH(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "addSigningKeysOperatorBH", _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x805911ae.
//
// Solidity: function addSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_OperatorRegistry *OperatorRegistrySession) AddSigningKeysOperatorBH(_nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.AddSigningKeysOperatorBH(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x805911ae.
//
// Solidity: function addSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) AddSigningKeysOperatorBH(_nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.AddSigningKeysOperatorBH(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// ClearNodeOperatorPenalty is a paid mutator transaction binding the contract method 0x30a90f01.
//
// Solidity: function clearNodeOperatorPenalty(uint256 _nodeOperatorId) returns(bool)
func (_OperatorRegistry *OperatorRegistryTransactor) ClearNodeOperatorPenalty(opts *bind.TransactOpts, _nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "clearNodeOperatorPenalty", _nodeOperatorId)
}

// ClearNodeOperatorPenalty is a paid mutator transaction binding the contract method 0x30a90f01.
//
// Solidity: function clearNodeOperatorPenalty(uint256 _nodeOperatorId) returns(bool)
func (_OperatorRegistry *OperatorRegistrySession) ClearNodeOperatorPenalty(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.ClearNodeOperatorPenalty(&_OperatorRegistry.TransactOpts, _nodeOperatorId)
}

// ClearNodeOperatorPenalty is a paid mutator transaction binding the contract method 0x30a90f01.
//
// Solidity: function clearNodeOperatorPenalty(uint256 _nodeOperatorId) returns(bool)
func (_OperatorRegistry *OperatorRegistryTransactorSession) ClearNodeOperatorPenalty(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.ClearNodeOperatorPenalty(&_OperatorRegistry.TransactOpts, _nodeOperatorId)
}

// DeactivateNodeOperator is a paid mutator transaction binding the contract method 0x75a080d5.
//
// Solidity: function deactivateNodeOperator(uint256 _nodeOperatorId) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) DeactivateNodeOperator(opts *bind.TransactOpts, _nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "deactivateNodeOperator", _nodeOperatorId)
}

// DeactivateNodeOperator is a paid mutator transaction binding the contract method 0x75a080d5.
//
// Solidity: function deactivateNodeOperator(uint256 _nodeOperatorId) returns()
func (_OperatorRegistry *OperatorRegistrySession) DeactivateNodeOperator(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.DeactivateNodeOperator(&_OperatorRegistry.TransactOpts, _nodeOperatorId)
}

// DeactivateNodeOperator is a paid mutator transaction binding the contract method 0x75a080d5.
//
// Solidity: function deactivateNodeOperator(uint256 _nodeOperatorId) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) DeactivateNodeOperator(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.DeactivateNodeOperator(&_OperatorRegistry.TransactOpts, _nodeOperatorId)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes _nodeOperatorIds, bytes _vettedSigningKeysCounts) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) DecreaseVettedSigningKeysCount(opts *bind.TransactOpts, _nodeOperatorIds []byte, _vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "decreaseVettedSigningKeysCount", _nodeOperatorIds, _vettedSigningKeysCounts)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes _nodeOperatorIds, bytes _vettedSigningKeysCounts) returns()
func (_OperatorRegistry *OperatorRegistrySession) DecreaseVettedSigningKeysCount(_nodeOperatorIds []byte, _vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.DecreaseVettedSigningKeysCount(&_OperatorRegistry.TransactOpts, _nodeOperatorIds, _vettedSigningKeysCounts)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes _nodeOperatorIds, bytes _vettedSigningKeysCounts) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) DecreaseVettedSigningKeysCount(_nodeOperatorIds []byte, _vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.DecreaseVettedSigningKeysCount(&_OperatorRegistry.TransactOpts, _nodeOperatorIds, _vettedSigningKeysCounts)
}

// DistributeReward is a paid mutator transaction binding the contract method 0x8f73c5ae.
//
// Solidity: function distributeReward() returns()
func (_OperatorRegistry *OperatorRegistryTransactor) DistributeReward(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "distributeReward")
}

// DistributeReward is a paid mutator transaction binding the contract method 0x8f73c5ae.
//
// Solidity: function distributeReward() returns()
func (_OperatorRegistry *OperatorRegistrySession) DistributeReward() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.DistributeReward(&_OperatorRegistry.TransactOpts)
}

// DistributeReward is a paid mutator transaction binding the contract method 0x8f73c5ae.
//
// Solidity: function distributeReward() returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) DistributeReward() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.DistributeReward(&_OperatorRegistry.TransactOpts)
}

// FinalizeUpgradeV2 is a paid mutator transaction binding the contract method 0x9a7c2ade.
//
// Solidity: function finalizeUpgrade_v2(address _locator, bytes32 _type, uint256 _stuckPenaltyDelay) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) FinalizeUpgradeV2(opts *bind.TransactOpts, _locator common.Address, _type [32]byte, _stuckPenaltyDelay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "finalizeUpgrade_v2", _locator, _type, _stuckPenaltyDelay)
}

// FinalizeUpgradeV2 is a paid mutator transaction binding the contract method 0x9a7c2ade.
//
// Solidity: function finalizeUpgrade_v2(address _locator, bytes32 _type, uint256 _stuckPenaltyDelay) returns()
func (_OperatorRegistry *OperatorRegistrySession) FinalizeUpgradeV2(_locator common.Address, _type [32]byte, _stuckPenaltyDelay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.FinalizeUpgradeV2(&_OperatorRegistry.TransactOpts, _locator, _type, _stuckPenaltyDelay)
}

// FinalizeUpgradeV2 is a paid mutator transaction binding the contract method 0x9a7c2ade.
//
// Solidity: function finalizeUpgrade_v2(address _locator, bytes32 _type, uint256 _stuckPenaltyDelay) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) FinalizeUpgradeV2(_locator common.Address, _type [32]byte, _stuckPenaltyDelay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.FinalizeUpgradeV2(&_OperatorRegistry.TransactOpts, _locator, _type, _stuckPenaltyDelay)
}

// FinalizeUpgradeV3 is a paid mutator transaction binding the contract method 0x6d395b7e.
//
// Solidity: function finalizeUpgrade_v3() returns()
func (_OperatorRegistry *OperatorRegistryTransactor) FinalizeUpgradeV3(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "finalizeUpgrade_v3")
}

// FinalizeUpgradeV3 is a paid mutator transaction binding the contract method 0x6d395b7e.
//
// Solidity: function finalizeUpgrade_v3() returns()
func (_OperatorRegistry *OperatorRegistrySession) FinalizeUpgradeV3() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.FinalizeUpgradeV3(&_OperatorRegistry.TransactOpts)
}

// FinalizeUpgradeV3 is a paid mutator transaction binding the contract method 0x6d395b7e.
//
// Solidity: function finalizeUpgrade_v3() returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) FinalizeUpgradeV3() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.FinalizeUpgradeV3(&_OperatorRegistry.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x684560a2.
//
// Solidity: function initialize(address _locator, bytes32 _type, uint256 _stuckPenaltyDelay) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) Initialize(opts *bind.TransactOpts, _locator common.Address, _type [32]byte, _stuckPenaltyDelay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "initialize", _locator, _type, _stuckPenaltyDelay)
}

// Initialize is a paid mutator transaction binding the contract method 0x684560a2.
//
// Solidity: function initialize(address _locator, bytes32 _type, uint256 _stuckPenaltyDelay) returns()
func (_OperatorRegistry *OperatorRegistrySession) Initialize(_locator common.Address, _type [32]byte, _stuckPenaltyDelay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.Initialize(&_OperatorRegistry.TransactOpts, _locator, _type, _stuckPenaltyDelay)
}

// Initialize is a paid mutator transaction binding the contract method 0x684560a2.
//
// Solidity: function initialize(address _locator, bytes32 _type, uint256 _stuckPenaltyDelay) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) Initialize(_locator common.Address, _type [32]byte, _stuckPenaltyDelay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.Initialize(&_OperatorRegistry.TransactOpts, _locator, _type, _stuckPenaltyDelay)
}

// InvalidateReadyToDepositKeysRange is a paid mutator transaction binding the contract method 0x65cc369a.
//
// Solidity: function invalidateReadyToDepositKeysRange(uint256 _indexFrom, uint256 _indexTo) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) InvalidateReadyToDepositKeysRange(opts *bind.TransactOpts, _indexFrom *big.Int, _indexTo *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "invalidateReadyToDepositKeysRange", _indexFrom, _indexTo)
}

// InvalidateReadyToDepositKeysRange is a paid mutator transaction binding the contract method 0x65cc369a.
//
// Solidity: function invalidateReadyToDepositKeysRange(uint256 _indexFrom, uint256 _indexTo) returns()
func (_OperatorRegistry *OperatorRegistrySession) InvalidateReadyToDepositKeysRange(_indexFrom *big.Int, _indexTo *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.InvalidateReadyToDepositKeysRange(&_OperatorRegistry.TransactOpts, _indexFrom, _indexTo)
}

// InvalidateReadyToDepositKeysRange is a paid mutator transaction binding the contract method 0x65cc369a.
//
// Solidity: function invalidateReadyToDepositKeysRange(uint256 _indexFrom, uint256 _indexTo) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) InvalidateReadyToDepositKeysRange(_indexFrom *big.Int, _indexTo *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.InvalidateReadyToDepositKeysRange(&_OperatorRegistry.TransactOpts, _indexFrom, _indexTo)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 _depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_OperatorRegistry *OperatorRegistryTransactor) ObtainDepositData(opts *bind.TransactOpts, _depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "obtainDepositData", _depositsCount, arg1)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 _depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_OperatorRegistry *OperatorRegistrySession) ObtainDepositData(_depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.ObtainDepositData(&_OperatorRegistry.TransactOpts, _depositsCount, arg1)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 _depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_OperatorRegistry *OperatorRegistryTransactorSession) ObtainDepositData(_depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.ObtainDepositData(&_OperatorRegistry.TransactOpts, _depositsCount, arg1)
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_OperatorRegistry *OperatorRegistryTransactor) OnExitedAndStuckValidatorsCountsUpdated(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "onExitedAndStuckValidatorsCountsUpdated")
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_OperatorRegistry *OperatorRegistrySession) OnExitedAndStuckValidatorsCountsUpdated() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OnExitedAndStuckValidatorsCountsUpdated(&_OperatorRegistry.TransactOpts)
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) OnExitedAndStuckValidatorsCountsUpdated() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OnExitedAndStuckValidatorsCountsUpdated(&_OperatorRegistry.TransactOpts)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 ) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) OnRewardsMinted(opts *bind.TransactOpts, arg0 *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "onRewardsMinted", arg0)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 ) returns()
func (_OperatorRegistry *OperatorRegistrySession) OnRewardsMinted(arg0 *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OnRewardsMinted(&_OperatorRegistry.TransactOpts, arg0)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 ) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) OnRewardsMinted(arg0 *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OnRewardsMinted(&_OperatorRegistry.TransactOpts, arg0)
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_OperatorRegistry *OperatorRegistryTransactor) OnWithdrawalCredentialsChanged(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "onWithdrawalCredentialsChanged")
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_OperatorRegistry *OperatorRegistrySession) OnWithdrawalCredentialsChanged() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OnWithdrawalCredentialsChanged(&_OperatorRegistry.TransactOpts)
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) OnWithdrawalCredentialsChanged() (*types.Transaction, error) {
	return _OperatorRegistry.Contract.OnWithdrawalCredentialsChanged(&_OperatorRegistry.TransactOpts)
}

// RemoveSigningKey is a paid mutator transaction binding the contract method 0x6ef355f1.
//
// Solidity: function removeSigningKey(uint256 _nodeOperatorId, uint256 _index) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) RemoveSigningKey(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "removeSigningKey", _nodeOperatorId, _index)
}

// RemoveSigningKey is a paid mutator transaction binding the contract method 0x6ef355f1.
//
// Solidity: function removeSigningKey(uint256 _nodeOperatorId, uint256 _index) returns()
func (_OperatorRegistry *OperatorRegistrySession) RemoveSigningKey(_nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RemoveSigningKey(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _index)
}

// RemoveSigningKey is a paid mutator transaction binding the contract method 0x6ef355f1.
//
// Solidity: function removeSigningKey(uint256 _nodeOperatorId, uint256 _index) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) RemoveSigningKey(_nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RemoveSigningKey(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _index)
}

// RemoveSigningKeyOperatorBH is a paid mutator transaction binding the contract method 0xed5cfa41.
//
// Solidity: function removeSigningKeyOperatorBH(uint256 _nodeOperatorId, uint256 _index) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) RemoveSigningKeyOperatorBH(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "removeSigningKeyOperatorBH", _nodeOperatorId, _index)
}

// RemoveSigningKeyOperatorBH is a paid mutator transaction binding the contract method 0xed5cfa41.
//
// Solidity: function removeSigningKeyOperatorBH(uint256 _nodeOperatorId, uint256 _index) returns()
func (_OperatorRegistry *OperatorRegistrySession) RemoveSigningKeyOperatorBH(_nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RemoveSigningKeyOperatorBH(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _index)
}

// RemoveSigningKeyOperatorBH is a paid mutator transaction binding the contract method 0xed5cfa41.
//
// Solidity: function removeSigningKeyOperatorBH(uint256 _nodeOperatorId, uint256 _index) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) RemoveSigningKeyOperatorBH(_nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RemoveSigningKeyOperatorBH(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _index)
}

// RemoveSigningKeys is a paid mutator transaction binding the contract method 0x7038b141.
//
// Solidity: function removeSigningKeys(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) RemoveSigningKeys(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "removeSigningKeys", _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeys is a paid mutator transaction binding the contract method 0x7038b141.
//
// Solidity: function removeSigningKeys(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_OperatorRegistry *OperatorRegistrySession) RemoveSigningKeys(_nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RemoveSigningKeys(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeys is a paid mutator transaction binding the contract method 0x7038b141.
//
// Solidity: function removeSigningKeys(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) RemoveSigningKeys(_nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RemoveSigningKeys(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x5ddde810.
//
// Solidity: function removeSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) RemoveSigningKeysOperatorBH(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "removeSigningKeysOperatorBH", _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x5ddde810.
//
// Solidity: function removeSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_OperatorRegistry *OperatorRegistrySession) RemoveSigningKeysOperatorBH(_nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RemoveSigningKeysOperatorBH(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x5ddde810.
//
// Solidity: function removeSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) RemoveSigningKeysOperatorBH(_nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.RemoveSigningKeysOperatorBH(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _fromIndex, _keysCount)
}

// SetNodeOperatorName is a paid mutator transaction binding the contract method 0x5e57d742.
//
// Solidity: function setNodeOperatorName(uint256 _nodeOperatorId, string _name) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) SetNodeOperatorName(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _name string) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "setNodeOperatorName", _nodeOperatorId, _name)
}

// SetNodeOperatorName is a paid mutator transaction binding the contract method 0x5e57d742.
//
// Solidity: function setNodeOperatorName(uint256 _nodeOperatorId, string _name) returns()
func (_OperatorRegistry *OperatorRegistrySession) SetNodeOperatorName(_nodeOperatorId *big.Int, _name string) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.SetNodeOperatorName(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _name)
}

// SetNodeOperatorName is a paid mutator transaction binding the contract method 0x5e57d742.
//
// Solidity: function setNodeOperatorName(uint256 _nodeOperatorId, string _name) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) SetNodeOperatorName(_nodeOperatorId *big.Int, _name string) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.SetNodeOperatorName(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _name)
}

// SetNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x973e9328.
//
// Solidity: function setNodeOperatorRewardAddress(uint256 _nodeOperatorId, address _rewardAddress) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) SetNodeOperatorRewardAddress(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _rewardAddress common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "setNodeOperatorRewardAddress", _nodeOperatorId, _rewardAddress)
}

// SetNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x973e9328.
//
// Solidity: function setNodeOperatorRewardAddress(uint256 _nodeOperatorId, address _rewardAddress) returns()
func (_OperatorRegistry *OperatorRegistrySession) SetNodeOperatorRewardAddress(_nodeOperatorId *big.Int, _rewardAddress common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.SetNodeOperatorRewardAddress(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _rewardAddress)
}

// SetNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x973e9328.
//
// Solidity: function setNodeOperatorRewardAddress(uint256 _nodeOperatorId, address _rewardAddress) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) SetNodeOperatorRewardAddress(_nodeOperatorId *big.Int, _rewardAddress common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.SetNodeOperatorRewardAddress(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _rewardAddress)
}

// SetNodeOperatorStakingLimit is a paid mutator transaction binding the contract method 0xae962acf.
//
// Solidity: function setNodeOperatorStakingLimit(uint256 _nodeOperatorId, uint64 _vettedSigningKeysCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) SetNodeOperatorStakingLimit(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _vettedSigningKeysCount uint64) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "setNodeOperatorStakingLimit", _nodeOperatorId, _vettedSigningKeysCount)
}

// SetNodeOperatorStakingLimit is a paid mutator transaction binding the contract method 0xae962acf.
//
// Solidity: function setNodeOperatorStakingLimit(uint256 _nodeOperatorId, uint64 _vettedSigningKeysCount) returns()
func (_OperatorRegistry *OperatorRegistrySession) SetNodeOperatorStakingLimit(_nodeOperatorId *big.Int, _vettedSigningKeysCount uint64) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.SetNodeOperatorStakingLimit(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _vettedSigningKeysCount)
}

// SetNodeOperatorStakingLimit is a paid mutator transaction binding the contract method 0xae962acf.
//
// Solidity: function setNodeOperatorStakingLimit(uint256 _nodeOperatorId, uint64 _vettedSigningKeysCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) SetNodeOperatorStakingLimit(_nodeOperatorId *big.Int, _vettedSigningKeysCount uint64) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.SetNodeOperatorStakingLimit(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _vettedSigningKeysCount)
}

// SetStuckPenaltyDelay is a paid mutator transaction binding the contract method 0x6ccc7562.
//
// Solidity: function setStuckPenaltyDelay(uint256 _delay) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) SetStuckPenaltyDelay(opts *bind.TransactOpts, _delay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "setStuckPenaltyDelay", _delay)
}

// SetStuckPenaltyDelay is a paid mutator transaction binding the contract method 0x6ccc7562.
//
// Solidity: function setStuckPenaltyDelay(uint256 _delay) returns()
func (_OperatorRegistry *OperatorRegistrySession) SetStuckPenaltyDelay(_delay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.SetStuckPenaltyDelay(&_OperatorRegistry.TransactOpts, _delay)
}

// SetStuckPenaltyDelay is a paid mutator transaction binding the contract method 0x6ccc7562.
//
// Solidity: function setStuckPenaltyDelay(uint256 _delay) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) SetStuckPenaltyDelay(_delay *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.SetStuckPenaltyDelay(&_OperatorRegistry.TransactOpts, _delay)
}

// TransferToVault is a paid mutator transaction binding the contract method 0x9d4941d8.
//
// Solidity: function transferToVault(address _token) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) TransferToVault(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "transferToVault", _token)
}

// TransferToVault is a paid mutator transaction binding the contract method 0x9d4941d8.
//
// Solidity: function transferToVault(address _token) returns()
func (_OperatorRegistry *OperatorRegistrySession) TransferToVault(_token common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.TransferToVault(&_OperatorRegistry.TransactOpts, _token)
}

// TransferToVault is a paid mutator transaction binding the contract method 0x9d4941d8.
//
// Solidity: function transferToVault(address _token) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) TransferToVault(_token common.Address) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.TransferToVault(&_OperatorRegistry.TransactOpts, _token)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0xf2e2ca63.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 _nodeOperatorId, uint256 _exitedValidatorsCount, uint256 _stuckValidatorsCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) UnsafeUpdateValidatorsCount(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _exitedValidatorsCount *big.Int, _stuckValidatorsCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "unsafeUpdateValidatorsCount", _nodeOperatorId, _exitedValidatorsCount, _stuckValidatorsCount)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0xf2e2ca63.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 _nodeOperatorId, uint256 _exitedValidatorsCount, uint256 _stuckValidatorsCount) returns()
func (_OperatorRegistry *OperatorRegistrySession) UnsafeUpdateValidatorsCount(_nodeOperatorId *big.Int, _exitedValidatorsCount *big.Int, _stuckValidatorsCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UnsafeUpdateValidatorsCount(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _exitedValidatorsCount, _stuckValidatorsCount)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0xf2e2ca63.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 _nodeOperatorId, uint256 _exitedValidatorsCount, uint256 _stuckValidatorsCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) UnsafeUpdateValidatorsCount(_nodeOperatorId *big.Int, _exitedValidatorsCount *big.Int, _stuckValidatorsCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UnsafeUpdateValidatorsCount(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _exitedValidatorsCount, _stuckValidatorsCount)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes _nodeOperatorIds, bytes _exitedValidatorsCounts) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) UpdateExitedValidatorsCount(opts *bind.TransactOpts, _nodeOperatorIds []byte, _exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "updateExitedValidatorsCount", _nodeOperatorIds, _exitedValidatorsCounts)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes _nodeOperatorIds, bytes _exitedValidatorsCounts) returns()
func (_OperatorRegistry *OperatorRegistrySession) UpdateExitedValidatorsCount(_nodeOperatorIds []byte, _exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateExitedValidatorsCount(&_OperatorRegistry.TransactOpts, _nodeOperatorIds, _exitedValidatorsCounts)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes _nodeOperatorIds, bytes _exitedValidatorsCounts) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) UpdateExitedValidatorsCount(_nodeOperatorIds []byte, _exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateExitedValidatorsCount(&_OperatorRegistry.TransactOpts, _nodeOperatorIds, _exitedValidatorsCounts)
}

// UpdateRefundedValidatorsCount is a paid mutator transaction binding the contract method 0xa2e080f1.
//
// Solidity: function updateRefundedValidatorsCount(uint256 _nodeOperatorId, uint256 _refundedValidatorsCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) UpdateRefundedValidatorsCount(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _refundedValidatorsCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "updateRefundedValidatorsCount", _nodeOperatorId, _refundedValidatorsCount)
}

// UpdateRefundedValidatorsCount is a paid mutator transaction binding the contract method 0xa2e080f1.
//
// Solidity: function updateRefundedValidatorsCount(uint256 _nodeOperatorId, uint256 _refundedValidatorsCount) returns()
func (_OperatorRegistry *OperatorRegistrySession) UpdateRefundedValidatorsCount(_nodeOperatorId *big.Int, _refundedValidatorsCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateRefundedValidatorsCount(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _refundedValidatorsCount)
}

// UpdateRefundedValidatorsCount is a paid mutator transaction binding the contract method 0xa2e080f1.
//
// Solidity: function updateRefundedValidatorsCount(uint256 _nodeOperatorId, uint256 _refundedValidatorsCount) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) UpdateRefundedValidatorsCount(_nodeOperatorId *big.Int, _refundedValidatorsCount *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateRefundedValidatorsCount(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _refundedValidatorsCount)
}

// UpdateStuckValidatorsCount is a paid mutator transaction binding the contract method 0x9b3d1900.
//
// Solidity: function updateStuckValidatorsCount(bytes _nodeOperatorIds, bytes _stuckValidatorsCounts) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) UpdateStuckValidatorsCount(opts *bind.TransactOpts, _nodeOperatorIds []byte, _stuckValidatorsCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "updateStuckValidatorsCount", _nodeOperatorIds, _stuckValidatorsCounts)
}

// UpdateStuckValidatorsCount is a paid mutator transaction binding the contract method 0x9b3d1900.
//
// Solidity: function updateStuckValidatorsCount(bytes _nodeOperatorIds, bytes _stuckValidatorsCounts) returns()
func (_OperatorRegistry *OperatorRegistrySession) UpdateStuckValidatorsCount(_nodeOperatorIds []byte, _stuckValidatorsCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateStuckValidatorsCount(&_OperatorRegistry.TransactOpts, _nodeOperatorIds, _stuckValidatorsCounts)
}

// UpdateStuckValidatorsCount is a paid mutator transaction binding the contract method 0x9b3d1900.
//
// Solidity: function updateStuckValidatorsCount(bytes _nodeOperatorIds, bytes _stuckValidatorsCounts) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) UpdateStuckValidatorsCount(_nodeOperatorIds []byte, _stuckValidatorsCounts []byte) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateStuckValidatorsCount(&_OperatorRegistry.TransactOpts, _nodeOperatorIds, _stuckValidatorsCounts)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, uint256 _targetLimitMode, uint256 _targetLimit) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) UpdateTargetValidatorsLimits(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _targetLimitMode *big.Int, _targetLimit *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "updateTargetValidatorsLimits", _nodeOperatorId, _targetLimitMode, _targetLimit)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, uint256 _targetLimitMode, uint256 _targetLimit) returns()
func (_OperatorRegistry *OperatorRegistrySession) UpdateTargetValidatorsLimits(_nodeOperatorId *big.Int, _targetLimitMode *big.Int, _targetLimit *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateTargetValidatorsLimits(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _targetLimitMode, _targetLimit)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, uint256 _targetLimitMode, uint256 _targetLimit) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) UpdateTargetValidatorsLimits(_nodeOperatorId *big.Int, _targetLimitMode *big.Int, _targetLimit *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateTargetValidatorsLimits(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _targetLimitMode, _targetLimit)
}

// UpdateTargetValidatorsLimits0 is a paid mutator transaction binding the contract method 0xa9e7a846.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, bool _isTargetLimitActive, uint256 _targetLimit) returns()
func (_OperatorRegistry *OperatorRegistryTransactor) UpdateTargetValidatorsLimits0(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _isTargetLimitActive bool, _targetLimit *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.contract.Transact(opts, "updateTargetValidatorsLimits0", _nodeOperatorId, _isTargetLimitActive, _targetLimit)
}

// UpdateTargetValidatorsLimits0 is a paid mutator transaction binding the contract method 0xa9e7a846.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, bool _isTargetLimitActive, uint256 _targetLimit) returns()
func (_OperatorRegistry *OperatorRegistrySession) UpdateTargetValidatorsLimits0(_nodeOperatorId *big.Int, _isTargetLimitActive bool, _targetLimit *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateTargetValidatorsLimits0(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _isTargetLimitActive, _targetLimit)
}

// UpdateTargetValidatorsLimits0 is a paid mutator transaction binding the contract method 0xa9e7a846.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, bool _isTargetLimitActive, uint256 _targetLimit) returns()
func (_OperatorRegistry *OperatorRegistryTransactorSession) UpdateTargetValidatorsLimits0(_nodeOperatorId *big.Int, _isTargetLimitActive bool, _targetLimit *big.Int) (*types.Transaction, error) {
	return _OperatorRegistry.Contract.UpdateTargetValidatorsLimits0(&_OperatorRegistry.TransactOpts, _nodeOperatorId, _isTargetLimitActive, _targetLimit)
}

// OperatorRegistryContractVersionSetIterator is returned from FilterContractVersionSet and is used to iterate over the raw logs and unpacked data for ContractVersionSet events raised by the OperatorRegistry contract.
type OperatorRegistryContractVersionSetIterator struct {
	Event *OperatorRegistryContractVersionSet // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryContractVersionSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryContractVersionSet)
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
		it.Event = new(OperatorRegistryContractVersionSet)
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
func (it *OperatorRegistryContractVersionSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryContractVersionSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryContractVersionSet represents a ContractVersionSet event raised by the OperatorRegistry contract.
type OperatorRegistryContractVersionSet struct {
	Version *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterContractVersionSet is a free log retrieval operation binding the contract event 0xfddcded6b4f4730c226821172046b48372d3cd963c159701ae1b7c3bcac541bb.
//
// Solidity: event ContractVersionSet(uint256 version)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterContractVersionSet(opts *bind.FilterOpts) (*OperatorRegistryContractVersionSetIterator, error) {

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "ContractVersionSet")
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryContractVersionSetIterator{contract: _OperatorRegistry.contract, event: "ContractVersionSet", logs: logs, sub: sub}, nil
}

// WatchContractVersionSet is a free log subscription operation binding the contract event 0xfddcded6b4f4730c226821172046b48372d3cd963c159701ae1b7c3bcac541bb.
//
// Solidity: event ContractVersionSet(uint256 version)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchContractVersionSet(opts *bind.WatchOpts, sink chan<- *OperatorRegistryContractVersionSet) (event.Subscription, error) {

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "ContractVersionSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryContractVersionSet)
				if err := _OperatorRegistry.contract.UnpackLog(event, "ContractVersionSet", log); err != nil {
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

// ParseContractVersionSet is a log parse operation binding the contract event 0xfddcded6b4f4730c226821172046b48372d3cd963c159701ae1b7c3bcac541bb.
//
// Solidity: event ContractVersionSet(uint256 version)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseContractVersionSet(log types.Log) (*OperatorRegistryContractVersionSet, error) {
	event := new(OperatorRegistryContractVersionSet)
	if err := _OperatorRegistry.contract.UnpackLog(event, "ContractVersionSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryDepositedSigningKeysCountChangedIterator is returned from FilterDepositedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for DepositedSigningKeysCountChanged events raised by the OperatorRegistry contract.
type OperatorRegistryDepositedSigningKeysCountChangedIterator struct {
	Event *OperatorRegistryDepositedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryDepositedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryDepositedSigningKeysCountChanged)
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
		it.Event = new(OperatorRegistryDepositedSigningKeysCountChanged)
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
func (it *OperatorRegistryDepositedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryDepositedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryDepositedSigningKeysCountChanged represents a DepositedSigningKeysCountChanged event raised by the OperatorRegistry contract.
type OperatorRegistryDepositedSigningKeysCountChanged struct {
	NodeOperatorId           *big.Int
	DepositedValidatorsCount *big.Int
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterDepositedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x24eb1c9e765ba41accf9437300ea91ece5ed3f897ec3cdee0e9debd7fe309b78.
//
// Solidity: event DepositedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterDepositedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryDepositedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "DepositedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryDepositedSigningKeysCountChangedIterator{contract: _OperatorRegistry.contract, event: "DepositedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchDepositedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x24eb1c9e765ba41accf9437300ea91ece5ed3f897ec3cdee0e9debd7fe309b78.
//
// Solidity: event DepositedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchDepositedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryDepositedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "DepositedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryDepositedSigningKeysCountChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "DepositedSigningKeysCountChanged", log); err != nil {
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

// ParseDepositedSigningKeysCountChanged is a log parse operation binding the contract event 0x24eb1c9e765ba41accf9437300ea91ece5ed3f897ec3cdee0e9debd7fe309b78.
//
// Solidity: event DepositedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseDepositedSigningKeysCountChanged(log types.Log) (*OperatorRegistryDepositedSigningKeysCountChanged, error) {
	event := new(OperatorRegistryDepositedSigningKeysCountChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "DepositedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryExitedSigningKeysCountChangedIterator is returned from FilterExitedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for ExitedSigningKeysCountChanged events raised by the OperatorRegistry contract.
type OperatorRegistryExitedSigningKeysCountChangedIterator struct {
	Event *OperatorRegistryExitedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryExitedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryExitedSigningKeysCountChanged)
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
		it.Event = new(OperatorRegistryExitedSigningKeysCountChanged)
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
func (it *OperatorRegistryExitedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryExitedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryExitedSigningKeysCountChanged represents a ExitedSigningKeysCountChanged event raised by the OperatorRegistry contract.
type OperatorRegistryExitedSigningKeysCountChanged struct {
	NodeOperatorId        *big.Int
	ExitedValidatorsCount *big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterExitedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x0f67960648751434ae86bf350db61194f387fda387e7f568b0ccd0ae0c220166.
//
// Solidity: event ExitedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 exitedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterExitedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryExitedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "ExitedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryExitedSigningKeysCountChangedIterator{contract: _OperatorRegistry.contract, event: "ExitedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchExitedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x0f67960648751434ae86bf350db61194f387fda387e7f568b0ccd0ae0c220166.
//
// Solidity: event ExitedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 exitedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchExitedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryExitedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "ExitedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryExitedSigningKeysCountChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "ExitedSigningKeysCountChanged", log); err != nil {
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

// ParseExitedSigningKeysCountChanged is a log parse operation binding the contract event 0x0f67960648751434ae86bf350db61194f387fda387e7f568b0ccd0ae0c220166.
//
// Solidity: event ExitedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 exitedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseExitedSigningKeysCountChanged(log types.Log) (*OperatorRegistryExitedSigningKeysCountChanged, error) {
	event := new(OperatorRegistryExitedSigningKeysCountChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "ExitedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryKeysOpIndexSetIterator is returned from FilterKeysOpIndexSet and is used to iterate over the raw logs and unpacked data for KeysOpIndexSet events raised by the OperatorRegistry contract.
type OperatorRegistryKeysOpIndexSetIterator struct {
	Event *OperatorRegistryKeysOpIndexSet // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryKeysOpIndexSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryKeysOpIndexSet)
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
		it.Event = new(OperatorRegistryKeysOpIndexSet)
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
func (it *OperatorRegistryKeysOpIndexSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryKeysOpIndexSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryKeysOpIndexSet represents a KeysOpIndexSet event raised by the OperatorRegistry contract.
type OperatorRegistryKeysOpIndexSet struct {
	KeysOpIndex *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterKeysOpIndexSet is a free log retrieval operation binding the contract event 0xfb992daec9d46d64898e3a9336d02811349df6cbea8b95d4deb2fa6c7b454f0d.
//
// Solidity: event KeysOpIndexSet(uint256 keysOpIndex)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterKeysOpIndexSet(opts *bind.FilterOpts) (*OperatorRegistryKeysOpIndexSetIterator, error) {

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "KeysOpIndexSet")
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryKeysOpIndexSetIterator{contract: _OperatorRegistry.contract, event: "KeysOpIndexSet", logs: logs, sub: sub}, nil
}

// WatchKeysOpIndexSet is a free log subscription operation binding the contract event 0xfb992daec9d46d64898e3a9336d02811349df6cbea8b95d4deb2fa6c7b454f0d.
//
// Solidity: event KeysOpIndexSet(uint256 keysOpIndex)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchKeysOpIndexSet(opts *bind.WatchOpts, sink chan<- *OperatorRegistryKeysOpIndexSet) (event.Subscription, error) {

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "KeysOpIndexSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryKeysOpIndexSet)
				if err := _OperatorRegistry.contract.UnpackLog(event, "KeysOpIndexSet", log); err != nil {
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

// ParseKeysOpIndexSet is a log parse operation binding the contract event 0xfb992daec9d46d64898e3a9336d02811349df6cbea8b95d4deb2fa6c7b454f0d.
//
// Solidity: event KeysOpIndexSet(uint256 keysOpIndex)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseKeysOpIndexSet(log types.Log) (*OperatorRegistryKeysOpIndexSet, error) {
	event := new(OperatorRegistryKeysOpIndexSet)
	if err := _OperatorRegistry.contract.UnpackLog(event, "KeysOpIndexSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryLocatorContractSetIterator is returned from FilterLocatorContractSet and is used to iterate over the raw logs and unpacked data for LocatorContractSet events raised by the OperatorRegistry contract.
type OperatorRegistryLocatorContractSetIterator struct {
	Event *OperatorRegistryLocatorContractSet // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryLocatorContractSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryLocatorContractSet)
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
		it.Event = new(OperatorRegistryLocatorContractSet)
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
func (it *OperatorRegistryLocatorContractSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryLocatorContractSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryLocatorContractSet represents a LocatorContractSet event raised by the OperatorRegistry contract.
type OperatorRegistryLocatorContractSet struct {
	LocatorAddress common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterLocatorContractSet is a free log retrieval operation binding the contract event 0xa44aa4b7320163340e971b1f22f153bbb8a0151d783bd58377018ea5bc96d0c9.
//
// Solidity: event LocatorContractSet(address locatorAddress)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterLocatorContractSet(opts *bind.FilterOpts) (*OperatorRegistryLocatorContractSetIterator, error) {

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "LocatorContractSet")
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryLocatorContractSetIterator{contract: _OperatorRegistry.contract, event: "LocatorContractSet", logs: logs, sub: sub}, nil
}

// WatchLocatorContractSet is a free log subscription operation binding the contract event 0xa44aa4b7320163340e971b1f22f153bbb8a0151d783bd58377018ea5bc96d0c9.
//
// Solidity: event LocatorContractSet(address locatorAddress)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchLocatorContractSet(opts *bind.WatchOpts, sink chan<- *OperatorRegistryLocatorContractSet) (event.Subscription, error) {

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "LocatorContractSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryLocatorContractSet)
				if err := _OperatorRegistry.contract.UnpackLog(event, "LocatorContractSet", log); err != nil {
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

// ParseLocatorContractSet is a log parse operation binding the contract event 0xa44aa4b7320163340e971b1f22f153bbb8a0151d783bd58377018ea5bc96d0c9.
//
// Solidity: event LocatorContractSet(address locatorAddress)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseLocatorContractSet(log types.Log) (*OperatorRegistryLocatorContractSet, error) {
	event := new(OperatorRegistryLocatorContractSet)
	if err := _OperatorRegistry.contract.UnpackLog(event, "LocatorContractSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryNodeOperatorActiveSetIterator is returned from FilterNodeOperatorActiveSet and is used to iterate over the raw logs and unpacked data for NodeOperatorActiveSet events raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorActiveSetIterator struct {
	Event *OperatorRegistryNodeOperatorActiveSet // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryNodeOperatorActiveSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryNodeOperatorActiveSet)
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
		it.Event = new(OperatorRegistryNodeOperatorActiveSet)
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
func (it *OperatorRegistryNodeOperatorActiveSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryNodeOperatorActiveSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryNodeOperatorActiveSet represents a NodeOperatorActiveSet event raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorActiveSet struct {
	NodeOperatorId *big.Int
	Active         bool
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorActiveSet is a free log retrieval operation binding the contract event 0xecdf08e8a6c4493efb460f6abc7d14532074fa339c3a6410623a1d3ee0fb2cac.
//
// Solidity: event NodeOperatorActiveSet(uint256 indexed nodeOperatorId, bool active)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterNodeOperatorActiveSet(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryNodeOperatorActiveSetIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "NodeOperatorActiveSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryNodeOperatorActiveSetIterator{contract: _OperatorRegistry.contract, event: "NodeOperatorActiveSet", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorActiveSet is a free log subscription operation binding the contract event 0xecdf08e8a6c4493efb460f6abc7d14532074fa339c3a6410623a1d3ee0fb2cac.
//
// Solidity: event NodeOperatorActiveSet(uint256 indexed nodeOperatorId, bool active)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchNodeOperatorActiveSet(opts *bind.WatchOpts, sink chan<- *OperatorRegistryNodeOperatorActiveSet, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "NodeOperatorActiveSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryNodeOperatorActiveSet)
				if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorActiveSet", log); err != nil {
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

// ParseNodeOperatorActiveSet is a log parse operation binding the contract event 0xecdf08e8a6c4493efb460f6abc7d14532074fa339c3a6410623a1d3ee0fb2cac.
//
// Solidity: event NodeOperatorActiveSet(uint256 indexed nodeOperatorId, bool active)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseNodeOperatorActiveSet(log types.Log) (*OperatorRegistryNodeOperatorActiveSet, error) {
	event := new(OperatorRegistryNodeOperatorActiveSet)
	if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorActiveSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryNodeOperatorAddedIterator is returned from FilterNodeOperatorAdded and is used to iterate over the raw logs and unpacked data for NodeOperatorAdded events raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorAddedIterator struct {
	Event *OperatorRegistryNodeOperatorAdded // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryNodeOperatorAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryNodeOperatorAdded)
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
		it.Event = new(OperatorRegistryNodeOperatorAdded)
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
func (it *OperatorRegistryNodeOperatorAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryNodeOperatorAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryNodeOperatorAdded represents a NodeOperatorAdded event raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorAdded struct {
	NodeOperatorId *big.Int
	Name           string
	RewardAddress  common.Address
	StakingLimit   uint64
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorAdded is a free log retrieval operation binding the contract event 0xc52ec0ad7872dae440d886040390c13677df7bf3cca136d8d81e5e5e7dd62ff1.
//
// Solidity: event NodeOperatorAdded(uint256 nodeOperatorId, string name, address rewardAddress, uint64 stakingLimit)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterNodeOperatorAdded(opts *bind.FilterOpts) (*OperatorRegistryNodeOperatorAddedIterator, error) {

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "NodeOperatorAdded")
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryNodeOperatorAddedIterator{contract: _OperatorRegistry.contract, event: "NodeOperatorAdded", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorAdded is a free log subscription operation binding the contract event 0xc52ec0ad7872dae440d886040390c13677df7bf3cca136d8d81e5e5e7dd62ff1.
//
// Solidity: event NodeOperatorAdded(uint256 nodeOperatorId, string name, address rewardAddress, uint64 stakingLimit)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchNodeOperatorAdded(opts *bind.WatchOpts, sink chan<- *OperatorRegistryNodeOperatorAdded) (event.Subscription, error) {

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "NodeOperatorAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryNodeOperatorAdded)
				if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorAdded", log); err != nil {
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

// ParseNodeOperatorAdded is a log parse operation binding the contract event 0xc52ec0ad7872dae440d886040390c13677df7bf3cca136d8d81e5e5e7dd62ff1.
//
// Solidity: event NodeOperatorAdded(uint256 nodeOperatorId, string name, address rewardAddress, uint64 stakingLimit)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseNodeOperatorAdded(log types.Log) (*OperatorRegistryNodeOperatorAdded, error) {
	event := new(OperatorRegistryNodeOperatorAdded)
	if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryNodeOperatorNameSetIterator is returned from FilterNodeOperatorNameSet and is used to iterate over the raw logs and unpacked data for NodeOperatorNameSet events raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorNameSetIterator struct {
	Event *OperatorRegistryNodeOperatorNameSet // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryNodeOperatorNameSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryNodeOperatorNameSet)
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
		it.Event = new(OperatorRegistryNodeOperatorNameSet)
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
func (it *OperatorRegistryNodeOperatorNameSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryNodeOperatorNameSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryNodeOperatorNameSet represents a NodeOperatorNameSet event raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorNameSet struct {
	NodeOperatorId *big.Int
	Name           string
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorNameSet is a free log retrieval operation binding the contract event 0xcb16868f4831cc58a28d413f658752a2958bd1f50e94ed6391716b936c48093b.
//
// Solidity: event NodeOperatorNameSet(uint256 indexed nodeOperatorId, string name)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterNodeOperatorNameSet(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryNodeOperatorNameSetIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "NodeOperatorNameSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryNodeOperatorNameSetIterator{contract: _OperatorRegistry.contract, event: "NodeOperatorNameSet", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorNameSet is a free log subscription operation binding the contract event 0xcb16868f4831cc58a28d413f658752a2958bd1f50e94ed6391716b936c48093b.
//
// Solidity: event NodeOperatorNameSet(uint256 indexed nodeOperatorId, string name)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchNodeOperatorNameSet(opts *bind.WatchOpts, sink chan<- *OperatorRegistryNodeOperatorNameSet, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "NodeOperatorNameSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryNodeOperatorNameSet)
				if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorNameSet", log); err != nil {
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

// ParseNodeOperatorNameSet is a log parse operation binding the contract event 0xcb16868f4831cc58a28d413f658752a2958bd1f50e94ed6391716b936c48093b.
//
// Solidity: event NodeOperatorNameSet(uint256 indexed nodeOperatorId, string name)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseNodeOperatorNameSet(log types.Log) (*OperatorRegistryNodeOperatorNameSet, error) {
	event := new(OperatorRegistryNodeOperatorNameSet)
	if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorNameSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryNodeOperatorPenalizedIterator is returned from FilterNodeOperatorPenalized and is used to iterate over the raw logs and unpacked data for NodeOperatorPenalized events raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorPenalizedIterator struct {
	Event *OperatorRegistryNodeOperatorPenalized // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryNodeOperatorPenalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryNodeOperatorPenalized)
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
		it.Event = new(OperatorRegistryNodeOperatorPenalized)
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
func (it *OperatorRegistryNodeOperatorPenalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryNodeOperatorPenalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryNodeOperatorPenalized represents a NodeOperatorPenalized event raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorPenalized struct {
	RecipientAddress      common.Address
	SharesPenalizedAmount *big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorPenalized is a free log retrieval operation binding the contract event 0xe915a473fc2ef8e0231da98380f853b2aeea117a4392c67e753c54186bfbbd12.
//
// Solidity: event NodeOperatorPenalized(address indexed recipientAddress, uint256 sharesPenalizedAmount)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterNodeOperatorPenalized(opts *bind.FilterOpts, recipientAddress []common.Address) (*OperatorRegistryNodeOperatorPenalizedIterator, error) {

	var recipientAddressRule []interface{}
	for _, recipientAddressItem := range recipientAddress {
		recipientAddressRule = append(recipientAddressRule, recipientAddressItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "NodeOperatorPenalized", recipientAddressRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryNodeOperatorPenalizedIterator{contract: _OperatorRegistry.contract, event: "NodeOperatorPenalized", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorPenalized is a free log subscription operation binding the contract event 0xe915a473fc2ef8e0231da98380f853b2aeea117a4392c67e753c54186bfbbd12.
//
// Solidity: event NodeOperatorPenalized(address indexed recipientAddress, uint256 sharesPenalizedAmount)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchNodeOperatorPenalized(opts *bind.WatchOpts, sink chan<- *OperatorRegistryNodeOperatorPenalized, recipientAddress []common.Address) (event.Subscription, error) {

	var recipientAddressRule []interface{}
	for _, recipientAddressItem := range recipientAddress {
		recipientAddressRule = append(recipientAddressRule, recipientAddressItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "NodeOperatorPenalized", recipientAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryNodeOperatorPenalized)
				if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorPenalized", log); err != nil {
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

// ParseNodeOperatorPenalized is a log parse operation binding the contract event 0xe915a473fc2ef8e0231da98380f853b2aeea117a4392c67e753c54186bfbbd12.
//
// Solidity: event NodeOperatorPenalized(address indexed recipientAddress, uint256 sharesPenalizedAmount)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseNodeOperatorPenalized(log types.Log) (*OperatorRegistryNodeOperatorPenalized, error) {
	event := new(OperatorRegistryNodeOperatorPenalized)
	if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorPenalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryNodeOperatorPenaltyClearedIterator is returned from FilterNodeOperatorPenaltyCleared and is used to iterate over the raw logs and unpacked data for NodeOperatorPenaltyCleared events raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorPenaltyClearedIterator struct {
	Event *OperatorRegistryNodeOperatorPenaltyCleared // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryNodeOperatorPenaltyClearedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryNodeOperatorPenaltyCleared)
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
		it.Event = new(OperatorRegistryNodeOperatorPenaltyCleared)
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
func (it *OperatorRegistryNodeOperatorPenaltyClearedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryNodeOperatorPenaltyClearedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryNodeOperatorPenaltyCleared represents a NodeOperatorPenaltyCleared event raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorPenaltyCleared struct {
	NodeOperatorId *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorPenaltyCleared is a free log retrieval operation binding the contract event 0x8bc12804ec8f618bd18bedd0b06cba55d3775bf200242bf6cf93860e161c3f84.
//
// Solidity: event NodeOperatorPenaltyCleared(uint256 indexed nodeOperatorId)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterNodeOperatorPenaltyCleared(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryNodeOperatorPenaltyClearedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "NodeOperatorPenaltyCleared", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryNodeOperatorPenaltyClearedIterator{contract: _OperatorRegistry.contract, event: "NodeOperatorPenaltyCleared", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorPenaltyCleared is a free log subscription operation binding the contract event 0x8bc12804ec8f618bd18bedd0b06cba55d3775bf200242bf6cf93860e161c3f84.
//
// Solidity: event NodeOperatorPenaltyCleared(uint256 indexed nodeOperatorId)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchNodeOperatorPenaltyCleared(opts *bind.WatchOpts, sink chan<- *OperatorRegistryNodeOperatorPenaltyCleared, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "NodeOperatorPenaltyCleared", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryNodeOperatorPenaltyCleared)
				if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorPenaltyCleared", log); err != nil {
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

// ParseNodeOperatorPenaltyCleared is a log parse operation binding the contract event 0x8bc12804ec8f618bd18bedd0b06cba55d3775bf200242bf6cf93860e161c3f84.
//
// Solidity: event NodeOperatorPenaltyCleared(uint256 indexed nodeOperatorId)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseNodeOperatorPenaltyCleared(log types.Log) (*OperatorRegistryNodeOperatorPenaltyCleared, error) {
	event := new(OperatorRegistryNodeOperatorPenaltyCleared)
	if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorPenaltyCleared", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryNodeOperatorRewardAddressSetIterator is returned from FilterNodeOperatorRewardAddressSet and is used to iterate over the raw logs and unpacked data for NodeOperatorRewardAddressSet events raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorRewardAddressSetIterator struct {
	Event *OperatorRegistryNodeOperatorRewardAddressSet // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryNodeOperatorRewardAddressSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryNodeOperatorRewardAddressSet)
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
		it.Event = new(OperatorRegistryNodeOperatorRewardAddressSet)
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
func (it *OperatorRegistryNodeOperatorRewardAddressSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryNodeOperatorRewardAddressSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryNodeOperatorRewardAddressSet represents a NodeOperatorRewardAddressSet event raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorRewardAddressSet struct {
	NodeOperatorId *big.Int
	RewardAddress  common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorRewardAddressSet is a free log retrieval operation binding the contract event 0x9a52205165d510fc1e428886d52108725dc01ed544da1702dc7bd3fdb3f243b2.
//
// Solidity: event NodeOperatorRewardAddressSet(uint256 indexed nodeOperatorId, address rewardAddress)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterNodeOperatorRewardAddressSet(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryNodeOperatorRewardAddressSetIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "NodeOperatorRewardAddressSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryNodeOperatorRewardAddressSetIterator{contract: _OperatorRegistry.contract, event: "NodeOperatorRewardAddressSet", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorRewardAddressSet is a free log subscription operation binding the contract event 0x9a52205165d510fc1e428886d52108725dc01ed544da1702dc7bd3fdb3f243b2.
//
// Solidity: event NodeOperatorRewardAddressSet(uint256 indexed nodeOperatorId, address rewardAddress)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchNodeOperatorRewardAddressSet(opts *bind.WatchOpts, sink chan<- *OperatorRegistryNodeOperatorRewardAddressSet, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "NodeOperatorRewardAddressSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryNodeOperatorRewardAddressSet)
				if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorRewardAddressSet", log); err != nil {
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

// ParseNodeOperatorRewardAddressSet is a log parse operation binding the contract event 0x9a52205165d510fc1e428886d52108725dc01ed544da1702dc7bd3fdb3f243b2.
//
// Solidity: event NodeOperatorRewardAddressSet(uint256 indexed nodeOperatorId, address rewardAddress)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseNodeOperatorRewardAddressSet(log types.Log) (*OperatorRegistryNodeOperatorRewardAddressSet, error) {
	event := new(OperatorRegistryNodeOperatorRewardAddressSet)
	if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorRewardAddressSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryNodeOperatorTotalKeysTrimmedIterator is returned from FilterNodeOperatorTotalKeysTrimmed and is used to iterate over the raw logs and unpacked data for NodeOperatorTotalKeysTrimmed events raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorTotalKeysTrimmedIterator struct {
	Event *OperatorRegistryNodeOperatorTotalKeysTrimmed // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryNodeOperatorTotalKeysTrimmedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryNodeOperatorTotalKeysTrimmed)
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
		it.Event = new(OperatorRegistryNodeOperatorTotalKeysTrimmed)
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
func (it *OperatorRegistryNodeOperatorTotalKeysTrimmedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryNodeOperatorTotalKeysTrimmedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryNodeOperatorTotalKeysTrimmed represents a NodeOperatorTotalKeysTrimmed event raised by the OperatorRegistry contract.
type OperatorRegistryNodeOperatorTotalKeysTrimmed struct {
	NodeOperatorId   *big.Int
	TotalKeysTrimmed uint64
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorTotalKeysTrimmed is a free log retrieval operation binding the contract event 0x9824694569ba758f8872bb150515caaf8f1e2cc27e6805679c4ac8c3b9b83d87.
//
// Solidity: event NodeOperatorTotalKeysTrimmed(uint256 indexed nodeOperatorId, uint64 totalKeysTrimmed)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterNodeOperatorTotalKeysTrimmed(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryNodeOperatorTotalKeysTrimmedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "NodeOperatorTotalKeysTrimmed", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryNodeOperatorTotalKeysTrimmedIterator{contract: _OperatorRegistry.contract, event: "NodeOperatorTotalKeysTrimmed", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorTotalKeysTrimmed is a free log subscription operation binding the contract event 0x9824694569ba758f8872bb150515caaf8f1e2cc27e6805679c4ac8c3b9b83d87.
//
// Solidity: event NodeOperatorTotalKeysTrimmed(uint256 indexed nodeOperatorId, uint64 totalKeysTrimmed)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchNodeOperatorTotalKeysTrimmed(opts *bind.WatchOpts, sink chan<- *OperatorRegistryNodeOperatorTotalKeysTrimmed, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "NodeOperatorTotalKeysTrimmed", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryNodeOperatorTotalKeysTrimmed)
				if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorTotalKeysTrimmed", log); err != nil {
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

// ParseNodeOperatorTotalKeysTrimmed is a log parse operation binding the contract event 0x9824694569ba758f8872bb150515caaf8f1e2cc27e6805679c4ac8c3b9b83d87.
//
// Solidity: event NodeOperatorTotalKeysTrimmed(uint256 indexed nodeOperatorId, uint64 totalKeysTrimmed)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseNodeOperatorTotalKeysTrimmed(log types.Log) (*OperatorRegistryNodeOperatorTotalKeysTrimmed, error) {
	event := new(OperatorRegistryNodeOperatorTotalKeysTrimmed)
	if err := _OperatorRegistry.contract.UnpackLog(event, "NodeOperatorTotalKeysTrimmed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryNonceChangedIterator is returned from FilterNonceChanged and is used to iterate over the raw logs and unpacked data for NonceChanged events raised by the OperatorRegistry contract.
type OperatorRegistryNonceChangedIterator struct {
	Event *OperatorRegistryNonceChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryNonceChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryNonceChanged)
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
		it.Event = new(OperatorRegistryNonceChanged)
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
func (it *OperatorRegistryNonceChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryNonceChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryNonceChanged represents a NonceChanged event raised by the OperatorRegistry contract.
type OperatorRegistryNonceChanged struct {
	Nonce *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNonceChanged is a free log retrieval operation binding the contract event 0x7220970e1f1f12864ecccd8942690a837c7a8dd45d158cb891eb45a8a69134aa.
//
// Solidity: event NonceChanged(uint256 nonce)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterNonceChanged(opts *bind.FilterOpts) (*OperatorRegistryNonceChangedIterator, error) {

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "NonceChanged")
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryNonceChangedIterator{contract: _OperatorRegistry.contract, event: "NonceChanged", logs: logs, sub: sub}, nil
}

// WatchNonceChanged is a free log subscription operation binding the contract event 0x7220970e1f1f12864ecccd8942690a837c7a8dd45d158cb891eb45a8a69134aa.
//
// Solidity: event NonceChanged(uint256 nonce)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchNonceChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryNonceChanged) (event.Subscription, error) {

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "NonceChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryNonceChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "NonceChanged", log); err != nil {
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

// ParseNonceChanged is a log parse operation binding the contract event 0x7220970e1f1f12864ecccd8942690a837c7a8dd45d158cb891eb45a8a69134aa.
//
// Solidity: event NonceChanged(uint256 nonce)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseNonceChanged(log types.Log) (*OperatorRegistryNonceChanged, error) {
	event := new(OperatorRegistryNonceChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "NonceChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryRecoverToVaultIterator is returned from FilterRecoverToVault and is used to iterate over the raw logs and unpacked data for RecoverToVault events raised by the OperatorRegistry contract.
type OperatorRegistryRecoverToVaultIterator struct {
	Event *OperatorRegistryRecoverToVault // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryRecoverToVaultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryRecoverToVault)
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
		it.Event = new(OperatorRegistryRecoverToVault)
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
func (it *OperatorRegistryRecoverToVaultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryRecoverToVaultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryRecoverToVault represents a RecoverToVault event raised by the OperatorRegistry contract.
type OperatorRegistryRecoverToVault struct {
	Vault  common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRecoverToVault is a free log retrieval operation binding the contract event 0x596caf56044b55fb8c4ca640089bbc2b63cae3e978b851f5745cbb7c5b288e02.
//
// Solidity: event RecoverToVault(address indexed vault, address indexed token, uint256 amount)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterRecoverToVault(opts *bind.FilterOpts, vault []common.Address, token []common.Address) (*OperatorRegistryRecoverToVaultIterator, error) {

	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "RecoverToVault", vaultRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryRecoverToVaultIterator{contract: _OperatorRegistry.contract, event: "RecoverToVault", logs: logs, sub: sub}, nil
}

// WatchRecoverToVault is a free log subscription operation binding the contract event 0x596caf56044b55fb8c4ca640089bbc2b63cae3e978b851f5745cbb7c5b288e02.
//
// Solidity: event RecoverToVault(address indexed vault, address indexed token, uint256 amount)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchRecoverToVault(opts *bind.WatchOpts, sink chan<- *OperatorRegistryRecoverToVault, vault []common.Address, token []common.Address) (event.Subscription, error) {

	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "RecoverToVault", vaultRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryRecoverToVault)
				if err := _OperatorRegistry.contract.UnpackLog(event, "RecoverToVault", log); err != nil {
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

// ParseRecoverToVault is a log parse operation binding the contract event 0x596caf56044b55fb8c4ca640089bbc2b63cae3e978b851f5745cbb7c5b288e02.
//
// Solidity: event RecoverToVault(address indexed vault, address indexed token, uint256 amount)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseRecoverToVault(log types.Log) (*OperatorRegistryRecoverToVault, error) {
	event := new(OperatorRegistryRecoverToVault)
	if err := _OperatorRegistry.contract.UnpackLog(event, "RecoverToVault", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryRewardDistributionStateChangedIterator is returned from FilterRewardDistributionStateChanged and is used to iterate over the raw logs and unpacked data for RewardDistributionStateChanged events raised by the OperatorRegistry contract.
type OperatorRegistryRewardDistributionStateChangedIterator struct {
	Event *OperatorRegistryRewardDistributionStateChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryRewardDistributionStateChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryRewardDistributionStateChanged)
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
		it.Event = new(OperatorRegistryRewardDistributionStateChanged)
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
func (it *OperatorRegistryRewardDistributionStateChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryRewardDistributionStateChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryRewardDistributionStateChanged represents a RewardDistributionStateChanged event raised by the OperatorRegistry contract.
type OperatorRegistryRewardDistributionStateChanged struct {
	State uint8
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRewardDistributionStateChanged is a free log retrieval operation binding the contract event 0x7545d380f29a8ae65fafb1acdf2c7762ec02d5607fecbea9dd8d8245e1616d93.
//
// Solidity: event RewardDistributionStateChanged(uint8 state)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterRewardDistributionStateChanged(opts *bind.FilterOpts) (*OperatorRegistryRewardDistributionStateChangedIterator, error) {

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "RewardDistributionStateChanged")
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryRewardDistributionStateChangedIterator{contract: _OperatorRegistry.contract, event: "RewardDistributionStateChanged", logs: logs, sub: sub}, nil
}

// WatchRewardDistributionStateChanged is a free log subscription operation binding the contract event 0x7545d380f29a8ae65fafb1acdf2c7762ec02d5607fecbea9dd8d8245e1616d93.
//
// Solidity: event RewardDistributionStateChanged(uint8 state)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchRewardDistributionStateChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryRewardDistributionStateChanged) (event.Subscription, error) {

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "RewardDistributionStateChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryRewardDistributionStateChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "RewardDistributionStateChanged", log); err != nil {
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

// ParseRewardDistributionStateChanged is a log parse operation binding the contract event 0x7545d380f29a8ae65fafb1acdf2c7762ec02d5607fecbea9dd8d8245e1616d93.
//
// Solidity: event RewardDistributionStateChanged(uint8 state)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseRewardDistributionStateChanged(log types.Log) (*OperatorRegistryRewardDistributionStateChanged, error) {
	event := new(OperatorRegistryRewardDistributionStateChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "RewardDistributionStateChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryRewardsDistributedIterator is returned from FilterRewardsDistributed and is used to iterate over the raw logs and unpacked data for RewardsDistributed events raised by the OperatorRegistry contract.
type OperatorRegistryRewardsDistributedIterator struct {
	Event *OperatorRegistryRewardsDistributed // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryRewardsDistributedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryRewardsDistributed)
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
		it.Event = new(OperatorRegistryRewardsDistributed)
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
func (it *OperatorRegistryRewardsDistributedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryRewardsDistributedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryRewardsDistributed represents a RewardsDistributed event raised by the OperatorRegistry contract.
type OperatorRegistryRewardsDistributed struct {
	RewardAddress common.Address
	SharesAmount  *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRewardsDistributed is a free log retrieval operation binding the contract event 0xdf29796aad820e4bb192f3a8d631b76519bcd2cbe77cc85af20e9df53cece086.
//
// Solidity: event RewardsDistributed(address indexed rewardAddress, uint256 sharesAmount)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterRewardsDistributed(opts *bind.FilterOpts, rewardAddress []common.Address) (*OperatorRegistryRewardsDistributedIterator, error) {

	var rewardAddressRule []interface{}
	for _, rewardAddressItem := range rewardAddress {
		rewardAddressRule = append(rewardAddressRule, rewardAddressItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "RewardsDistributed", rewardAddressRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryRewardsDistributedIterator{contract: _OperatorRegistry.contract, event: "RewardsDistributed", logs: logs, sub: sub}, nil
}

// WatchRewardsDistributed is a free log subscription operation binding the contract event 0xdf29796aad820e4bb192f3a8d631b76519bcd2cbe77cc85af20e9df53cece086.
//
// Solidity: event RewardsDistributed(address indexed rewardAddress, uint256 sharesAmount)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchRewardsDistributed(opts *bind.WatchOpts, sink chan<- *OperatorRegistryRewardsDistributed, rewardAddress []common.Address) (event.Subscription, error) {

	var rewardAddressRule []interface{}
	for _, rewardAddressItem := range rewardAddress {
		rewardAddressRule = append(rewardAddressRule, rewardAddressItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "RewardsDistributed", rewardAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryRewardsDistributed)
				if err := _OperatorRegistry.contract.UnpackLog(event, "RewardsDistributed", log); err != nil {
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

// ParseRewardsDistributed is a log parse operation binding the contract event 0xdf29796aad820e4bb192f3a8d631b76519bcd2cbe77cc85af20e9df53cece086.
//
// Solidity: event RewardsDistributed(address indexed rewardAddress, uint256 sharesAmount)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseRewardsDistributed(log types.Log) (*OperatorRegistryRewardsDistributed, error) {
	event := new(OperatorRegistryRewardsDistributed)
	if err := _OperatorRegistry.contract.UnpackLog(event, "RewardsDistributed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryScriptResultIterator is returned from FilterScriptResult and is used to iterate over the raw logs and unpacked data for ScriptResult events raised by the OperatorRegistry contract.
type OperatorRegistryScriptResultIterator struct {
	Event *OperatorRegistryScriptResult // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryScriptResultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryScriptResult)
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
		it.Event = new(OperatorRegistryScriptResult)
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
func (it *OperatorRegistryScriptResultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryScriptResultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryScriptResult represents a ScriptResult event raised by the OperatorRegistry contract.
type OperatorRegistryScriptResult struct {
	Executor   common.Address
	Script     []byte
	Input      []byte
	ReturnData []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterScriptResult is a free log retrieval operation binding the contract event 0x5229a5dba83a54ae8cb5b51bdd6de9474cacbe9dd332f5185f3a4f4f2e3f4ad9.
//
// Solidity: event ScriptResult(address indexed executor, bytes script, bytes input, bytes returnData)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterScriptResult(opts *bind.FilterOpts, executor []common.Address) (*OperatorRegistryScriptResultIterator, error) {

	var executorRule []interface{}
	for _, executorItem := range executor {
		executorRule = append(executorRule, executorItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "ScriptResult", executorRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryScriptResultIterator{contract: _OperatorRegistry.contract, event: "ScriptResult", logs: logs, sub: sub}, nil
}

// WatchScriptResult is a free log subscription operation binding the contract event 0x5229a5dba83a54ae8cb5b51bdd6de9474cacbe9dd332f5185f3a4f4f2e3f4ad9.
//
// Solidity: event ScriptResult(address indexed executor, bytes script, bytes input, bytes returnData)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchScriptResult(opts *bind.WatchOpts, sink chan<- *OperatorRegistryScriptResult, executor []common.Address) (event.Subscription, error) {

	var executorRule []interface{}
	for _, executorItem := range executor {
		executorRule = append(executorRule, executorItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "ScriptResult", executorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryScriptResult)
				if err := _OperatorRegistry.contract.UnpackLog(event, "ScriptResult", log); err != nil {
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

// ParseScriptResult is a log parse operation binding the contract event 0x5229a5dba83a54ae8cb5b51bdd6de9474cacbe9dd332f5185f3a4f4f2e3f4ad9.
//
// Solidity: event ScriptResult(address indexed executor, bytes script, bytes input, bytes returnData)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseScriptResult(log types.Log) (*OperatorRegistryScriptResult, error) {
	event := new(OperatorRegistryScriptResult)
	if err := _OperatorRegistry.contract.UnpackLog(event, "ScriptResult", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryStakingModuleTypeSetIterator is returned from FilterStakingModuleTypeSet and is used to iterate over the raw logs and unpacked data for StakingModuleTypeSet events raised by the OperatorRegistry contract.
type OperatorRegistryStakingModuleTypeSetIterator struct {
	Event *OperatorRegistryStakingModuleTypeSet // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryStakingModuleTypeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryStakingModuleTypeSet)
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
		it.Event = new(OperatorRegistryStakingModuleTypeSet)
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
func (it *OperatorRegistryStakingModuleTypeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryStakingModuleTypeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryStakingModuleTypeSet represents a StakingModuleTypeSet event raised by the OperatorRegistry contract.
type OperatorRegistryStakingModuleTypeSet struct {
	ModuleType [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStakingModuleTypeSet is a free log retrieval operation binding the contract event 0xdb042010b15d1321c99552200b350bba0a95dfa3d0b43869983ce74b44d644ee.
//
// Solidity: event StakingModuleTypeSet(bytes32 moduleType)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterStakingModuleTypeSet(opts *bind.FilterOpts) (*OperatorRegistryStakingModuleTypeSetIterator, error) {

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "StakingModuleTypeSet")
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryStakingModuleTypeSetIterator{contract: _OperatorRegistry.contract, event: "StakingModuleTypeSet", logs: logs, sub: sub}, nil
}

// WatchStakingModuleTypeSet is a free log subscription operation binding the contract event 0xdb042010b15d1321c99552200b350bba0a95dfa3d0b43869983ce74b44d644ee.
//
// Solidity: event StakingModuleTypeSet(bytes32 moduleType)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchStakingModuleTypeSet(opts *bind.WatchOpts, sink chan<- *OperatorRegistryStakingModuleTypeSet) (event.Subscription, error) {

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "StakingModuleTypeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryStakingModuleTypeSet)
				if err := _OperatorRegistry.contract.UnpackLog(event, "StakingModuleTypeSet", log); err != nil {
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

// ParseStakingModuleTypeSet is a log parse operation binding the contract event 0xdb042010b15d1321c99552200b350bba0a95dfa3d0b43869983ce74b44d644ee.
//
// Solidity: event StakingModuleTypeSet(bytes32 moduleType)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseStakingModuleTypeSet(log types.Log) (*OperatorRegistryStakingModuleTypeSet, error) {
	event := new(OperatorRegistryStakingModuleTypeSet)
	if err := _OperatorRegistry.contract.UnpackLog(event, "StakingModuleTypeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryStuckPenaltyDelayChangedIterator is returned from FilterStuckPenaltyDelayChanged and is used to iterate over the raw logs and unpacked data for StuckPenaltyDelayChanged events raised by the OperatorRegistry contract.
type OperatorRegistryStuckPenaltyDelayChangedIterator struct {
	Event *OperatorRegistryStuckPenaltyDelayChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryStuckPenaltyDelayChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryStuckPenaltyDelayChanged)
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
		it.Event = new(OperatorRegistryStuckPenaltyDelayChanged)
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
func (it *OperatorRegistryStuckPenaltyDelayChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryStuckPenaltyDelayChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryStuckPenaltyDelayChanged represents a StuckPenaltyDelayChanged event raised by the OperatorRegistry contract.
type OperatorRegistryStuckPenaltyDelayChanged struct {
	StuckPenaltyDelay *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterStuckPenaltyDelayChanged is a free log retrieval operation binding the contract event 0x4cccd9748bff0341d9852cc61d82652a3003dcebea088f05388c0be1f26b4c8a.
//
// Solidity: event StuckPenaltyDelayChanged(uint256 stuckPenaltyDelay)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterStuckPenaltyDelayChanged(opts *bind.FilterOpts) (*OperatorRegistryStuckPenaltyDelayChangedIterator, error) {

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "StuckPenaltyDelayChanged")
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryStuckPenaltyDelayChangedIterator{contract: _OperatorRegistry.contract, event: "StuckPenaltyDelayChanged", logs: logs, sub: sub}, nil
}

// WatchStuckPenaltyDelayChanged is a free log subscription operation binding the contract event 0x4cccd9748bff0341d9852cc61d82652a3003dcebea088f05388c0be1f26b4c8a.
//
// Solidity: event StuckPenaltyDelayChanged(uint256 stuckPenaltyDelay)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchStuckPenaltyDelayChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryStuckPenaltyDelayChanged) (event.Subscription, error) {

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "StuckPenaltyDelayChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryStuckPenaltyDelayChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "StuckPenaltyDelayChanged", log); err != nil {
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

// ParseStuckPenaltyDelayChanged is a log parse operation binding the contract event 0x4cccd9748bff0341d9852cc61d82652a3003dcebea088f05388c0be1f26b4c8a.
//
// Solidity: event StuckPenaltyDelayChanged(uint256 stuckPenaltyDelay)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseStuckPenaltyDelayChanged(log types.Log) (*OperatorRegistryStuckPenaltyDelayChanged, error) {
	event := new(OperatorRegistryStuckPenaltyDelayChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "StuckPenaltyDelayChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryStuckPenaltyStateChangedIterator is returned from FilterStuckPenaltyStateChanged and is used to iterate over the raw logs and unpacked data for StuckPenaltyStateChanged events raised by the OperatorRegistry contract.
type OperatorRegistryStuckPenaltyStateChangedIterator struct {
	Event *OperatorRegistryStuckPenaltyStateChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryStuckPenaltyStateChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryStuckPenaltyStateChanged)
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
		it.Event = new(OperatorRegistryStuckPenaltyStateChanged)
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
func (it *OperatorRegistryStuckPenaltyStateChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryStuckPenaltyStateChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryStuckPenaltyStateChanged represents a StuckPenaltyStateChanged event raised by the OperatorRegistry contract.
type OperatorRegistryStuckPenaltyStateChanged struct {
	NodeOperatorId           *big.Int
	StuckValidatorsCount     *big.Int
	RefundedValidatorsCount  *big.Int
	StuckPenaltyEndTimestamp *big.Int
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterStuckPenaltyStateChanged is a free log retrieval operation binding the contract event 0x0ee42dd52dd2b8feb0fc9cc054a08162a23e022c177319db981cf339e5b8ffdb.
//
// Solidity: event StuckPenaltyStateChanged(uint256 indexed nodeOperatorId, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterStuckPenaltyStateChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryStuckPenaltyStateChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "StuckPenaltyStateChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryStuckPenaltyStateChangedIterator{contract: _OperatorRegistry.contract, event: "StuckPenaltyStateChanged", logs: logs, sub: sub}, nil
}

// WatchStuckPenaltyStateChanged is a free log subscription operation binding the contract event 0x0ee42dd52dd2b8feb0fc9cc054a08162a23e022c177319db981cf339e5b8ffdb.
//
// Solidity: event StuckPenaltyStateChanged(uint256 indexed nodeOperatorId, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchStuckPenaltyStateChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryStuckPenaltyStateChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "StuckPenaltyStateChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryStuckPenaltyStateChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "StuckPenaltyStateChanged", log); err != nil {
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

// ParseStuckPenaltyStateChanged is a log parse operation binding the contract event 0x0ee42dd52dd2b8feb0fc9cc054a08162a23e022c177319db981cf339e5b8ffdb.
//
// Solidity: event StuckPenaltyStateChanged(uint256 indexed nodeOperatorId, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseStuckPenaltyStateChanged(log types.Log) (*OperatorRegistryStuckPenaltyStateChanged, error) {
	event := new(OperatorRegistryStuckPenaltyStateChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "StuckPenaltyStateChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryTargetValidatorsCountChangedIterator is returned from FilterTargetValidatorsCountChanged and is used to iterate over the raw logs and unpacked data for TargetValidatorsCountChanged events raised by the OperatorRegistry contract.
type OperatorRegistryTargetValidatorsCountChangedIterator struct {
	Event *OperatorRegistryTargetValidatorsCountChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryTargetValidatorsCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryTargetValidatorsCountChanged)
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
		it.Event = new(OperatorRegistryTargetValidatorsCountChanged)
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
func (it *OperatorRegistryTargetValidatorsCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryTargetValidatorsCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryTargetValidatorsCountChanged represents a TargetValidatorsCountChanged event raised by the OperatorRegistry contract.
type OperatorRegistryTargetValidatorsCountChanged struct {
	NodeOperatorId        *big.Int
	TargetValidatorsCount *big.Int
	TargetLimitMode       *big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterTargetValidatorsCountChanged is a free log retrieval operation binding the contract event 0xf92eb109ce5b449e9b121c352c6aeb4319538a90738cb95d84f08e41274e92d2.
//
// Solidity: event TargetValidatorsCountChanged(uint256 indexed nodeOperatorId, uint256 targetValidatorsCount, uint256 targetLimitMode)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterTargetValidatorsCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryTargetValidatorsCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "TargetValidatorsCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryTargetValidatorsCountChangedIterator{contract: _OperatorRegistry.contract, event: "TargetValidatorsCountChanged", logs: logs, sub: sub}, nil
}

// WatchTargetValidatorsCountChanged is a free log subscription operation binding the contract event 0xf92eb109ce5b449e9b121c352c6aeb4319538a90738cb95d84f08e41274e92d2.
//
// Solidity: event TargetValidatorsCountChanged(uint256 indexed nodeOperatorId, uint256 targetValidatorsCount, uint256 targetLimitMode)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchTargetValidatorsCountChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryTargetValidatorsCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "TargetValidatorsCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryTargetValidatorsCountChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "TargetValidatorsCountChanged", log); err != nil {
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

// ParseTargetValidatorsCountChanged is a log parse operation binding the contract event 0xf92eb109ce5b449e9b121c352c6aeb4319538a90738cb95d84f08e41274e92d2.
//
// Solidity: event TargetValidatorsCountChanged(uint256 indexed nodeOperatorId, uint256 targetValidatorsCount, uint256 targetLimitMode)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseTargetValidatorsCountChanged(log types.Log) (*OperatorRegistryTargetValidatorsCountChanged, error) {
	event := new(OperatorRegistryTargetValidatorsCountChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "TargetValidatorsCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryTotalSigningKeysCountChangedIterator is returned from FilterTotalSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for TotalSigningKeysCountChanged events raised by the OperatorRegistry contract.
type OperatorRegistryTotalSigningKeysCountChangedIterator struct {
	Event *OperatorRegistryTotalSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryTotalSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryTotalSigningKeysCountChanged)
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
		it.Event = new(OperatorRegistryTotalSigningKeysCountChanged)
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
func (it *OperatorRegistryTotalSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryTotalSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryTotalSigningKeysCountChanged represents a TotalSigningKeysCountChanged event raised by the OperatorRegistry contract.
type OperatorRegistryTotalSigningKeysCountChanged struct {
	NodeOperatorId       *big.Int
	TotalValidatorsCount *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterTotalSigningKeysCountChanged is a free log retrieval operation binding the contract event 0xdd01838a366ae4dc9a86e1922512c0716abebc9a440baae0e22d2dec578223f0.
//
// Solidity: event TotalSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 totalValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterTotalSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryTotalSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "TotalSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryTotalSigningKeysCountChangedIterator{contract: _OperatorRegistry.contract, event: "TotalSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchTotalSigningKeysCountChanged is a free log subscription operation binding the contract event 0xdd01838a366ae4dc9a86e1922512c0716abebc9a440baae0e22d2dec578223f0.
//
// Solidity: event TotalSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 totalValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchTotalSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryTotalSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "TotalSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryTotalSigningKeysCountChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "TotalSigningKeysCountChanged", log); err != nil {
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

// ParseTotalSigningKeysCountChanged is a log parse operation binding the contract event 0xdd01838a366ae4dc9a86e1922512c0716abebc9a440baae0e22d2dec578223f0.
//
// Solidity: event TotalSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 totalValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseTotalSigningKeysCountChanged(log types.Log) (*OperatorRegistryTotalSigningKeysCountChanged, error) {
	event := new(OperatorRegistryTotalSigningKeysCountChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "TotalSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OperatorRegistryVettedSigningKeysCountChangedIterator is returned from FilterVettedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for VettedSigningKeysCountChanged events raised by the OperatorRegistry contract.
type OperatorRegistryVettedSigningKeysCountChangedIterator struct {
	Event *OperatorRegistryVettedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *OperatorRegistryVettedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OperatorRegistryVettedSigningKeysCountChanged)
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
		it.Event = new(OperatorRegistryVettedSigningKeysCountChanged)
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
func (it *OperatorRegistryVettedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OperatorRegistryVettedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OperatorRegistryVettedSigningKeysCountChanged represents a VettedSigningKeysCountChanged event raised by the OperatorRegistry contract.
type OperatorRegistryVettedSigningKeysCountChanged struct {
	NodeOperatorId          *big.Int
	ApprovedValidatorsCount *big.Int
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterVettedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x947f955eec7e1f626bee3afd2aa47b5de04ddcdd3fe78dc8838213015ef58dfd.
//
// Solidity: event VettedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 approvedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) FilterVettedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*OperatorRegistryVettedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.FilterLogs(opts, "VettedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &OperatorRegistryVettedSigningKeysCountChangedIterator{contract: _OperatorRegistry.contract, event: "VettedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchVettedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x947f955eec7e1f626bee3afd2aa47b5de04ddcdd3fe78dc8838213015ef58dfd.
//
// Solidity: event VettedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 approvedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) WatchVettedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *OperatorRegistryVettedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _OperatorRegistry.contract.WatchLogs(opts, "VettedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OperatorRegistryVettedSigningKeysCountChanged)
				if err := _OperatorRegistry.contract.UnpackLog(event, "VettedSigningKeysCountChanged", log); err != nil {
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

// ParseVettedSigningKeysCountChanged is a log parse operation binding the contract event 0x947f955eec7e1f626bee3afd2aa47b5de04ddcdd3fe78dc8838213015ef58dfd.
//
// Solidity: event VettedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 approvedValidatorsCount)
func (_OperatorRegistry *OperatorRegistryFilterer) ParseVettedSigningKeysCountChanged(log types.Log) (*OperatorRegistryVettedSigningKeysCountChanged, error) {
	event := new(OperatorRegistryVettedSigningKeysCountChanged)
	if err := _OperatorRegistry.contract.UnpackLog(event, "VettedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
