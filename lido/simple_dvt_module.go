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

// SimpleDVTModuleMetaData contains all meta data concerning the SimpleDVTModule contract.
var SimpleDVTModuleMetaData = &bind.MetaData{
	ABI: "[{\"constant\":true,\"inputs\":[],\"name\":\"hasInitialized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_targetLimitMode\",\"type\":\"uint256\"},{\"name\":\"_targetLimit\",\"type\":\"uint256\"}],\"name\":\"updateTargetValidatorsLimits\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_keysCount\",\"type\":\"uint256\"},{\"name\":\"_publicKeys\",\"type\":\"bytes\"},{\"name\":\"_signatures\",\"type\":\"bytes\"}],\"name\":\"addSigningKeys\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getType\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"exitDeadlineThreshold\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_script\",\"type\":\"bytes\"}],\"name\":\"getEVMScriptExecutor\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getRecoveryVault\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_offset\",\"type\":\"uint256\"},{\"name\":\"_limit\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorIds\",\"outputs\":[{\"name\":\"nodeOperatorIds\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_proofSlotTimestamp\",\"type\":\"uint256\"},{\"name\":\"_publicKey\",\"type\":\"bytes\"},{\"name\":\"_eligibleToExitInSec\",\"type\":\"uint256\"}],\"name\":\"reportValidatorExitDelay\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_offset\",\"type\":\"uint256\"},{\"name\":\"_limit\",\"type\":\"uint256\"}],\"name\":\"getSigningKeys\",\"outputs\":[{\"name\":\"pubkeys\",\"type\":\"bytes\"},{\"name\":\"signatures\",\"type\":\"bytes\"},{\"name\":\"used\",\"type\":\"bool[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_fromIndex\",\"type\":\"uint256\"},{\"name\":\"_keysCount\",\"type\":\"uint256\"}],\"name\":\"removeSigningKeysOperatorBH\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorIsActive\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_name\",\"type\":\"string\"}],\"name\":\"setNodeOperatorName\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_totalRewardShares\",\"type\":\"uint256\"}],\"name\":\"getRewardsDistribution\",\"outputs\":[{\"name\":\"recipients\",\"type\":\"address[]\"},{\"name\":\"shares\",\"type\":\"uint256[]\"},{\"name\":\"penalized\",\"type\":\"bool[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_indexFrom\",\"type\":\"uint256\"},{\"name\":\"_indexTo\",\"type\":\"uint256\"}],\"name\":\"invalidateReadyToDepositKeysRange\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_locator\",\"type\":\"address\"},{\"name\":\"_type\",\"type\":\"bytes32\"},{\"name\":\"_exitDeadlineThresholdInSeconds\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_publicKey\",\"type\":\"bytes\"},{\"name\":\"_withdrawalRequestPaidFee\",\"type\":\"uint256\"},{\"name\":\"_exitType\",\"type\":\"uint256\"}],\"name\":\"onValidatorExitTriggered\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStuckPenaltyDelay\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"removeSigningKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getRewardDistributionState\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_fromIndex\",\"type\":\"uint256\"},{\"name\":\"_keysCount\",\"type\":\"uint256\"}],\"name\":\"removeSigningKeys\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"isOperatorPenalized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"deactivateNodeOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"allowRecoverability\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"STAKING_ROUTER_ROLE\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_keysCount\",\"type\":\"uint256\"},{\"name\":\"_publicKeys\",\"type\":\"bytes\"},{\"name\":\"_signatures\",\"type\":\"bytes\"}],\"name\":\"addSigningKeysOperatorBH\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"appId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"_proofSlotTimestamp\",\"type\":\"uint256\"},{\"name\":\"_publicKey\",\"type\":\"bytes\"},{\"name\":\"_eligibleToExitInSec\",\"type\":\"uint256\"}],\"name\":\"isValidatorExitDelayPenaltyApplicable\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getActiveNodeOperatorsCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_name\",\"type\":\"string\"},{\"name\":\"_rewardAddress\",\"type\":\"address\"}],\"name\":\"addNodeOperator\",\"outputs\":[{\"name\":\"id\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getContractVersion\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getInitializationBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getUnusedSigningKeyCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"onRewardsMinted\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MANAGE_NODE_OPERATOR_ROLE\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_exitDeadlineThresholdInSeconds\",\"type\":\"uint256\"}],\"name\":\"finalizeUpgrade_v4\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"distributeReward\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"onWithdrawalCredentialsChanged\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"activateNodeOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_exitedValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"unsafeUpdateValidatorsCount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_rewardAddress\",\"type\":\"address\"}],\"name\":\"setNodeOperatorRewardAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_fullInfo\",\"type\":\"bool\"}],\"name\":\"getNodeOperator\",\"outputs\":[{\"name\":\"active\",\"type\":\"bool\"},{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"rewardAddress\",\"type\":\"address\"},{\"name\":\"totalVettedValidators\",\"type\":\"uint64\"},{\"name\":\"totalExitedValidators\",\"type\":\"uint64\"},{\"name\":\"totalAddedValidators\",\"type\":\"uint64\"},{\"name\":\"totalDepositedValidators\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStakingModuleSummary\",\"outputs\":[{\"name\":\"totalExitedValidators\",\"type\":\"uint256\"},{\"name\":\"totalDepositedValidators\",\"type\":\"uint256\"},{\"name\":\"depositableValidatorsCount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorIds\",\"type\":\"bytes\"},{\"name\":\"_exitedValidatorsCounts\",\"type\":\"bytes\"}],\"name\":\"updateExitedValidatorsCount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"transferToVault\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_sender\",\"type\":\"address\"},{\"name\":\"_role\",\"type\":\"bytes32\"},{\"name\":\"_params\",\"type\":\"uint256[]\"}],\"name\":\"canPerform\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getEVMScriptRegistry\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getNodeOperatorsCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_isTargetLimitActive\",\"type\":\"bool\"},{\"name\":\"_targetLimit\",\"type\":\"uint256\"}],\"name\":\"updateTargetValidatorsLimits\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_vettedSigningKeysCount\",\"type\":\"uint64\"}],\"name\":\"setNodeOperatorStakingLimit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorSummary\",\"outputs\":[{\"name\":\"targetLimitMode\",\"type\":\"uint256\"},{\"name\":\"targetValidatorsCount\",\"type\":\"uint256\"},{\"name\":\"stuckValidatorsCount\",\"type\":\"uint256\"},{\"name\":\"refundedValidatorsCount\",\"type\":\"uint256\"},{\"name\":\"stuckPenaltyEndTimestamp\",\"type\":\"uint256\"},{\"name\":\"totalExitedValidators\",\"type\":\"uint256\"},{\"name\":\"totalDepositedValidators\",\"type\":\"uint256\"},{\"name\":\"depositableValidatorsCount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"getSigningKey\",\"outputs\":[{\"name\":\"key\",\"type\":\"bytes\"},{\"name\":\"depositSignature\",\"type\":\"bytes\"},{\"name\":\"used\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_NODE_OPERATOR_NAME_LENGTH\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorIds\",\"type\":\"bytes\"},{\"name\":\"_vettedSigningKeysCounts\",\"type\":\"bytes\"}],\"name\":\"decreaseVettedSigningKeysCount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_publicKey\",\"type\":\"bytes\"}],\"name\":\"isValidatorExitingKeyReported\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_threshold\",\"type\":\"uint256\"},{\"name\":\"_lateReportingWindow\",\"type\":\"uint256\"}],\"name\":\"setExitDeadlineThreshold\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_depositsCount\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"obtainDepositData\",\"outputs\":[{\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"name\":\"signatures\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"exitPenaltyCutoffTimestamp\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getKeysOpIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"kernel\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getLocator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"SET_NODE_OPERATOR_LIMIT_ROLE\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getTotalSigningKeyCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isPetrified\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_STUCK_PENALTY_DELAY\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"onExitedAndStuckValidatorsCountsUpdated\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_NODE_OPERATORS_COUNT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_nodeOperatorId\",\"type\":\"uint256\"},{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"removeSigningKeyOperatorBH\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MANAGE_SIGNING_KEYS\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"isOperatorPenaltyCleared\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"rewardAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"stakingLimit\",\"type\":\"uint64\"}],\"name\":\"NodeOperatorAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"active\",\"type\":\"bool\"}],\"name\":\"NodeOperatorActiveSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"name\",\"type\":\"string\"}],\"name\":\"NodeOperatorNameSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"rewardAddress\",\"type\":\"address\"}],\"name\":\"NodeOperatorRewardAddressSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalKeysTrimmed\",\"type\":\"uint64\"}],\"name\":\"NodeOperatorTotalKeysTrimmed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"keysOpIndex\",\"type\":\"uint256\"}],\"name\":\"KeysOpIndexSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"moduleType\",\"type\":\"bytes32\"}],\"name\":\"StakingModuleTypeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"rewardAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"sharesAmount\",\"type\":\"uint256\"}],\"name\":\"RewardsDistributed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"state\",\"type\":\"uint8\"}],\"name\":\"RewardDistributionStateChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"locatorAddress\",\"type\":\"address\"}],\"name\":\"LocatorContractSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"approvedValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"VettedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"depositedValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"DepositedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"exitedValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"ExitedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"TotalSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"NonceChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"targetValidatorsCount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"targetLimitMode\",\"type\":\"uint256\"}],\"name\":\"TargetValidatorsCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"eligibleToExitInSec\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"proofSlotTimestamp\",\"type\":\"uint256\"}],\"name\":\"ValidatorExitStatusUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"withdrawalRequestPaidFee\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"exitType\",\"type\":\"uint256\"}],\"name\":\"ValidatorExitTriggered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"threshold\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"reportingWindow\",\"type\":\"uint256\"}],\"name\":\"ExitDeadlineThresholdChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"ContractVersionSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"executor\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"script\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"input\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"returnData\",\"type\":\"bytes\"}],\"name\":\"ScriptResult\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"vault\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"RecoverToVault\",\"type\":\"event\"}]",
}

// SimpleDVTModuleABI is the input ABI used to generate the binding from.
// Deprecated: Use SimpleDVTModuleMetaData.ABI instead.
var SimpleDVTModuleABI = SimpleDVTModuleMetaData.ABI

// SimpleDVTModule is an auto generated Go binding around an Ethereum contract.
type SimpleDVTModule struct {
	SimpleDVTModuleCaller     // Read-only binding to the contract
	SimpleDVTModuleTransactor // Write-only binding to the contract
	SimpleDVTModuleFilterer   // Log filterer for contract events
}

// SimpleDVTModuleCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleDVTModuleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleDVTModuleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleDVTModuleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleDVTModuleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleDVTModuleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleDVTModuleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleDVTModuleSession struct {
	Contract     *SimpleDVTModule  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SimpleDVTModuleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleDVTModuleCallerSession struct {
	Contract *SimpleDVTModuleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// SimpleDVTModuleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleDVTModuleTransactorSession struct {
	Contract     *SimpleDVTModuleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// SimpleDVTModuleRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleDVTModuleRaw struct {
	Contract *SimpleDVTModule // Generic contract binding to access the raw methods on
}

// SimpleDVTModuleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleDVTModuleCallerRaw struct {
	Contract *SimpleDVTModuleCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleDVTModuleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleDVTModuleTransactorRaw struct {
	Contract *SimpleDVTModuleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleDVTModule creates a new instance of SimpleDVTModule, bound to a specific deployed contract.
func NewSimpleDVTModule(address common.Address, backend bind.ContractBackend) (*SimpleDVTModule, error) {
	contract, err := bindSimpleDVTModule(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModule{SimpleDVTModuleCaller: SimpleDVTModuleCaller{contract: contract}, SimpleDVTModuleTransactor: SimpleDVTModuleTransactor{contract: contract}, SimpleDVTModuleFilterer: SimpleDVTModuleFilterer{contract: contract}}, nil
}

// NewSimpleDVTModuleCaller creates a new read-only instance of SimpleDVTModule, bound to a specific deployed contract.
func NewSimpleDVTModuleCaller(address common.Address, caller bind.ContractCaller) (*SimpleDVTModuleCaller, error) {
	contract, err := bindSimpleDVTModule(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleCaller{contract: contract}, nil
}

// NewSimpleDVTModuleTransactor creates a new write-only instance of SimpleDVTModule, bound to a specific deployed contract.
func NewSimpleDVTModuleTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleDVTModuleTransactor, error) {
	contract, err := bindSimpleDVTModule(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleTransactor{contract: contract}, nil
}

// NewSimpleDVTModuleFilterer creates a new log filterer instance of SimpleDVTModule, bound to a specific deployed contract.
func NewSimpleDVTModuleFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleDVTModuleFilterer, error) {
	contract, err := bindSimpleDVTModule(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleFilterer{contract: contract}, nil
}

// bindSimpleDVTModule binds a generic wrapper to an already deployed contract.
func bindSimpleDVTModule(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SimpleDVTModuleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleDVTModule *SimpleDVTModuleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SimpleDVTModule.Contract.SimpleDVTModuleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleDVTModule *SimpleDVTModuleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SimpleDVTModuleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleDVTModule *SimpleDVTModuleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SimpleDVTModuleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleDVTModule *SimpleDVTModuleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SimpleDVTModule.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleDVTModule *SimpleDVTModuleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleDVTModule *SimpleDVTModuleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.contract.Transact(opts, method, params...)
}

// MANAGENODEOPERATORROLE is a free data retrieval call binding the contract method 0x8ece9995.
//
// Solidity: function MANAGE_NODE_OPERATOR_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCaller) MANAGENODEOPERATORROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "MANAGE_NODE_OPERATOR_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MANAGENODEOPERATORROLE is a free data retrieval call binding the contract method 0x8ece9995.
//
// Solidity: function MANAGE_NODE_OPERATOR_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleSession) MANAGENODEOPERATORROLE() ([32]byte, error) {
	return _SimpleDVTModule.Contract.MANAGENODEOPERATORROLE(&_SimpleDVTModule.CallOpts)
}

// MANAGENODEOPERATORROLE is a free data retrieval call binding the contract method 0x8ece9995.
//
// Solidity: function MANAGE_NODE_OPERATOR_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) MANAGENODEOPERATORROLE() ([32]byte, error) {
	return _SimpleDVTModule.Contract.MANAGENODEOPERATORROLE(&_SimpleDVTModule.CallOpts)
}

// MANAGESIGNINGKEYS is a free data retrieval call binding the contract method 0xf31bd9c1.
//
// Solidity: function MANAGE_SIGNING_KEYS() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCaller) MANAGESIGNINGKEYS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "MANAGE_SIGNING_KEYS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MANAGESIGNINGKEYS is a free data retrieval call binding the contract method 0xf31bd9c1.
//
// Solidity: function MANAGE_SIGNING_KEYS() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleSession) MANAGESIGNINGKEYS() ([32]byte, error) {
	return _SimpleDVTModule.Contract.MANAGESIGNINGKEYS(&_SimpleDVTModule.CallOpts)
}

// MANAGESIGNINGKEYS is a free data retrieval call binding the contract method 0xf31bd9c1.
//
// Solidity: function MANAGE_SIGNING_KEYS() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) MANAGESIGNINGKEYS() ([32]byte, error) {
	return _SimpleDVTModule.Contract.MANAGESIGNINGKEYS(&_SimpleDVTModule.CallOpts)
}

// MAXNODEOPERATORSCOUNT is a free data retrieval call binding the contract method 0xec5af3a4.
//
// Solidity: function MAX_NODE_OPERATORS_COUNT() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) MAXNODEOPERATORSCOUNT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "MAX_NODE_OPERATORS_COUNT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXNODEOPERATORSCOUNT is a free data retrieval call binding the contract method 0xec5af3a4.
//
// Solidity: function MAX_NODE_OPERATORS_COUNT() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) MAXNODEOPERATORSCOUNT() (*big.Int, error) {
	return _SimpleDVTModule.Contract.MAXNODEOPERATORSCOUNT(&_SimpleDVTModule.CallOpts)
}

// MAXNODEOPERATORSCOUNT is a free data retrieval call binding the contract method 0xec5af3a4.
//
// Solidity: function MAX_NODE_OPERATORS_COUNT() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) MAXNODEOPERATORSCOUNT() (*big.Int, error) {
	return _SimpleDVTModule.Contract.MAXNODEOPERATORSCOUNT(&_SimpleDVTModule.CallOpts)
}

// MAXNODEOPERATORNAMELENGTH is a free data retrieval call binding the contract method 0xb4971833.
//
// Solidity: function MAX_NODE_OPERATOR_NAME_LENGTH() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) MAXNODEOPERATORNAMELENGTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "MAX_NODE_OPERATOR_NAME_LENGTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXNODEOPERATORNAMELENGTH is a free data retrieval call binding the contract method 0xb4971833.
//
// Solidity: function MAX_NODE_OPERATOR_NAME_LENGTH() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) MAXNODEOPERATORNAMELENGTH() (*big.Int, error) {
	return _SimpleDVTModule.Contract.MAXNODEOPERATORNAMELENGTH(&_SimpleDVTModule.CallOpts)
}

// MAXNODEOPERATORNAMELENGTH is a free data retrieval call binding the contract method 0xb4971833.
//
// Solidity: function MAX_NODE_OPERATOR_NAME_LENGTH() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) MAXNODEOPERATORNAMELENGTH() (*big.Int, error) {
	return _SimpleDVTModule.Contract.MAXNODEOPERATORNAMELENGTH(&_SimpleDVTModule.CallOpts)
}

// MAXSTUCKPENALTYDELAY is a free data retrieval call binding the contract method 0xe204d09b.
//
// Solidity: function MAX_STUCK_PENALTY_DELAY() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) MAXSTUCKPENALTYDELAY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "MAX_STUCK_PENALTY_DELAY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXSTUCKPENALTYDELAY is a free data retrieval call binding the contract method 0xe204d09b.
//
// Solidity: function MAX_STUCK_PENALTY_DELAY() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) MAXSTUCKPENALTYDELAY() (*big.Int, error) {
	return _SimpleDVTModule.Contract.MAXSTUCKPENALTYDELAY(&_SimpleDVTModule.CallOpts)
}

// MAXSTUCKPENALTYDELAY is a free data retrieval call binding the contract method 0xe204d09b.
//
// Solidity: function MAX_STUCK_PENALTY_DELAY() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) MAXSTUCKPENALTYDELAY() (*big.Int, error) {
	return _SimpleDVTModule.Contract.MAXSTUCKPENALTYDELAY(&_SimpleDVTModule.CallOpts)
}

// SETNODEOPERATORLIMITROLE is a free data retrieval call binding the contract method 0xd8e71cd1.
//
// Solidity: function SET_NODE_OPERATOR_LIMIT_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCaller) SETNODEOPERATORLIMITROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "SET_NODE_OPERATOR_LIMIT_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SETNODEOPERATORLIMITROLE is a free data retrieval call binding the contract method 0xd8e71cd1.
//
// Solidity: function SET_NODE_OPERATOR_LIMIT_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleSession) SETNODEOPERATORLIMITROLE() ([32]byte, error) {
	return _SimpleDVTModule.Contract.SETNODEOPERATORLIMITROLE(&_SimpleDVTModule.CallOpts)
}

// SETNODEOPERATORLIMITROLE is a free data retrieval call binding the contract method 0xd8e71cd1.
//
// Solidity: function SET_NODE_OPERATOR_LIMIT_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) SETNODEOPERATORLIMITROLE() ([32]byte, error) {
	return _SimpleDVTModule.Contract.SETNODEOPERATORLIMITROLE(&_SimpleDVTModule.CallOpts)
}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCaller) STAKINGROUTERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "STAKING_ROUTER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleSession) STAKINGROUTERROLE() ([32]byte, error) {
	return _SimpleDVTModule.Contract.STAKINGROUTERROLE(&_SimpleDVTModule.CallOpts)
}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) STAKINGROUTERROLE() ([32]byte, error) {
	return _SimpleDVTModule.Contract.STAKINGROUTERROLE(&_SimpleDVTModule.CallOpts)
}

// AllowRecoverability is a free data retrieval call binding the contract method 0x7e7db6e1.
//
// Solidity: function allowRecoverability(address token) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) AllowRecoverability(opts *bind.CallOpts, token common.Address) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "allowRecoverability", token)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllowRecoverability is a free data retrieval call binding the contract method 0x7e7db6e1.
//
// Solidity: function allowRecoverability(address token) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) AllowRecoverability(token common.Address) (bool, error) {
	return _SimpleDVTModule.Contract.AllowRecoverability(&_SimpleDVTModule.CallOpts, token)
}

// AllowRecoverability is a free data retrieval call binding the contract method 0x7e7db6e1.
//
// Solidity: function allowRecoverability(address token) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) AllowRecoverability(token common.Address) (bool, error) {
	return _SimpleDVTModule.Contract.AllowRecoverability(&_SimpleDVTModule.CallOpts, token)
}

// AppId is a free data retrieval call binding the contract method 0x80afdea8.
//
// Solidity: function appId() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCaller) AppId(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "appId")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// AppId is a free data retrieval call binding the contract method 0x80afdea8.
//
// Solidity: function appId() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleSession) AppId() ([32]byte, error) {
	return _SimpleDVTModule.Contract.AppId(&_SimpleDVTModule.CallOpts)
}

// AppId is a free data retrieval call binding the contract method 0x80afdea8.
//
// Solidity: function appId() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) AppId() ([32]byte, error) {
	return _SimpleDVTModule.Contract.AppId(&_SimpleDVTModule.CallOpts)
}

// CanPerform is a free data retrieval call binding the contract method 0xa1658fad.
//
// Solidity: function canPerform(address _sender, bytes32 _role, uint256[] _params) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) CanPerform(opts *bind.CallOpts, _sender common.Address, _role [32]byte, _params []*big.Int) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "canPerform", _sender, _role, _params)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanPerform is a free data retrieval call binding the contract method 0xa1658fad.
//
// Solidity: function canPerform(address _sender, bytes32 _role, uint256[] _params) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) CanPerform(_sender common.Address, _role [32]byte, _params []*big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.CanPerform(&_SimpleDVTModule.CallOpts, _sender, _role, _params)
}

// CanPerform is a free data retrieval call binding the contract method 0xa1658fad.
//
// Solidity: function canPerform(address _sender, bytes32 _role, uint256[] _params) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) CanPerform(_sender common.Address, _role [32]byte, _params []*big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.CanPerform(&_SimpleDVTModule.CallOpts, _sender, _role, _params)
}

// ExitDeadlineThreshold is a free data retrieval call binding the contract method 0x28d6d36b.
//
// Solidity: function exitDeadlineThreshold(uint256 ) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) ExitDeadlineThreshold(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "exitDeadlineThreshold", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ExitDeadlineThreshold is a free data retrieval call binding the contract method 0x28d6d36b.
//
// Solidity: function exitDeadlineThreshold(uint256 ) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) ExitDeadlineThreshold(arg0 *big.Int) (*big.Int, error) {
	return _SimpleDVTModule.Contract.ExitDeadlineThreshold(&_SimpleDVTModule.CallOpts, arg0)
}

// ExitDeadlineThreshold is a free data retrieval call binding the contract method 0x28d6d36b.
//
// Solidity: function exitDeadlineThreshold(uint256 ) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) ExitDeadlineThreshold(arg0 *big.Int) (*big.Int, error) {
	return _SimpleDVTModule.Contract.ExitDeadlineThreshold(&_SimpleDVTModule.CallOpts, arg0)
}

// ExitPenaltyCutoffTimestamp is a free data retrieval call binding the contract method 0xcfe58712.
//
// Solidity: function exitPenaltyCutoffTimestamp() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) ExitPenaltyCutoffTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "exitPenaltyCutoffTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ExitPenaltyCutoffTimestamp is a free data retrieval call binding the contract method 0xcfe58712.
//
// Solidity: function exitPenaltyCutoffTimestamp() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) ExitPenaltyCutoffTimestamp() (*big.Int, error) {
	return _SimpleDVTModule.Contract.ExitPenaltyCutoffTimestamp(&_SimpleDVTModule.CallOpts)
}

// ExitPenaltyCutoffTimestamp is a free data retrieval call binding the contract method 0xcfe58712.
//
// Solidity: function exitPenaltyCutoffTimestamp() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) ExitPenaltyCutoffTimestamp() (*big.Int, error) {
	return _SimpleDVTModule.Contract.ExitPenaltyCutoffTimestamp(&_SimpleDVTModule.CallOpts)
}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetActiveNodeOperatorsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getActiveNodeOperatorsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetActiveNodeOperatorsCount() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetActiveNodeOperatorsCount(&_SimpleDVTModule.CallOpts)
}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetActiveNodeOperatorsCount() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetActiveNodeOperatorsCount(&_SimpleDVTModule.CallOpts)
}

// GetContractVersion is a free data retrieval call binding the contract method 0x8aa10435.
//
// Solidity: function getContractVersion() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetContractVersion(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getContractVersion")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetContractVersion is a free data retrieval call binding the contract method 0x8aa10435.
//
// Solidity: function getContractVersion() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetContractVersion() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetContractVersion(&_SimpleDVTModule.CallOpts)
}

// GetContractVersion is a free data retrieval call binding the contract method 0x8aa10435.
//
// Solidity: function getContractVersion() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetContractVersion() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetContractVersion(&_SimpleDVTModule.CallOpts)
}

// GetEVMScriptExecutor is a free data retrieval call binding the contract method 0x2914b9bd.
//
// Solidity: function getEVMScriptExecutor(bytes _script) view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetEVMScriptExecutor(opts *bind.CallOpts, _script []byte) (common.Address, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getEVMScriptExecutor", _script)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetEVMScriptExecutor is a free data retrieval call binding the contract method 0x2914b9bd.
//
// Solidity: function getEVMScriptExecutor(bytes _script) view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetEVMScriptExecutor(_script []byte) (common.Address, error) {
	return _SimpleDVTModule.Contract.GetEVMScriptExecutor(&_SimpleDVTModule.CallOpts, _script)
}

// GetEVMScriptExecutor is a free data retrieval call binding the contract method 0x2914b9bd.
//
// Solidity: function getEVMScriptExecutor(bytes _script) view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetEVMScriptExecutor(_script []byte) (common.Address, error) {
	return _SimpleDVTModule.Contract.GetEVMScriptExecutor(&_SimpleDVTModule.CallOpts, _script)
}

// GetEVMScriptRegistry is a free data retrieval call binding the contract method 0xa479e508.
//
// Solidity: function getEVMScriptRegistry() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetEVMScriptRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getEVMScriptRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetEVMScriptRegistry is a free data retrieval call binding the contract method 0xa479e508.
//
// Solidity: function getEVMScriptRegistry() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetEVMScriptRegistry() (common.Address, error) {
	return _SimpleDVTModule.Contract.GetEVMScriptRegistry(&_SimpleDVTModule.CallOpts)
}

// GetEVMScriptRegistry is a free data retrieval call binding the contract method 0xa479e508.
//
// Solidity: function getEVMScriptRegistry() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetEVMScriptRegistry() (common.Address, error) {
	return _SimpleDVTModule.Contract.GetEVMScriptRegistry(&_SimpleDVTModule.CallOpts)
}

// GetInitializationBlock is a free data retrieval call binding the contract method 0x8b3dd749.
//
// Solidity: function getInitializationBlock() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetInitializationBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getInitializationBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetInitializationBlock is a free data retrieval call binding the contract method 0x8b3dd749.
//
// Solidity: function getInitializationBlock() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetInitializationBlock() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetInitializationBlock(&_SimpleDVTModule.CallOpts)
}

// GetInitializationBlock is a free data retrieval call binding the contract method 0x8b3dd749.
//
// Solidity: function getInitializationBlock() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetInitializationBlock() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetInitializationBlock(&_SimpleDVTModule.CallOpts)
}

// GetKeysOpIndex is a free data retrieval call binding the contract method 0xd07442f1.
//
// Solidity: function getKeysOpIndex() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetKeysOpIndex(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getKeysOpIndex")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetKeysOpIndex is a free data retrieval call binding the contract method 0xd07442f1.
//
// Solidity: function getKeysOpIndex() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetKeysOpIndex() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetKeysOpIndex(&_SimpleDVTModule.CallOpts)
}

// GetKeysOpIndex is a free data retrieval call binding the contract method 0xd07442f1.
//
// Solidity: function getKeysOpIndex() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetKeysOpIndex() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetKeysOpIndex(&_SimpleDVTModule.CallOpts)
}

// GetLocator is a free data retrieval call binding the contract method 0xd8343dcb.
//
// Solidity: function getLocator() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetLocator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getLocator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetLocator is a free data retrieval call binding the contract method 0xd8343dcb.
//
// Solidity: function getLocator() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetLocator() (common.Address, error) {
	return _SimpleDVTModule.Contract.GetLocator(&_SimpleDVTModule.CallOpts)
}

// GetLocator is a free data retrieval call binding the contract method 0xd8343dcb.
//
// Solidity: function getLocator() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetLocator() (common.Address, error) {
	return _SimpleDVTModule.Contract.GetLocator(&_SimpleDVTModule.CallOpts)
}

// GetNodeOperator is a free data retrieval call binding the contract method 0x9a56983c.
//
// Solidity: function getNodeOperator(uint256 _nodeOperatorId, bool _fullInfo) view returns(bool active, string name, address rewardAddress, uint64 totalVettedValidators, uint64 totalExitedValidators, uint64 totalAddedValidators, uint64 totalDepositedValidators)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetNodeOperator(opts *bind.CallOpts, _nodeOperatorId *big.Int, _fullInfo bool) (struct {
	Active                   bool
	Name                     string
	RewardAddress            common.Address
	TotalVettedValidators    uint64
	TotalExitedValidators    uint64
	TotalAddedValidators     uint64
	TotalDepositedValidators uint64
}, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getNodeOperator", _nodeOperatorId, _fullInfo)

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
func (_SimpleDVTModule *SimpleDVTModuleSession) GetNodeOperator(_nodeOperatorId *big.Int, _fullInfo bool) (struct {
	Active                   bool
	Name                     string
	RewardAddress            common.Address
	TotalVettedValidators    uint64
	TotalExitedValidators    uint64
	TotalAddedValidators     uint64
	TotalDepositedValidators uint64
}, error) {
	return _SimpleDVTModule.Contract.GetNodeOperator(&_SimpleDVTModule.CallOpts, _nodeOperatorId, _fullInfo)
}

// GetNodeOperator is a free data retrieval call binding the contract method 0x9a56983c.
//
// Solidity: function getNodeOperator(uint256 _nodeOperatorId, bool _fullInfo) view returns(bool active, string name, address rewardAddress, uint64 totalVettedValidators, uint64 totalExitedValidators, uint64 totalAddedValidators, uint64 totalDepositedValidators)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetNodeOperator(_nodeOperatorId *big.Int, _fullInfo bool) (struct {
	Active                   bool
	Name                     string
	RewardAddress            common.Address
	TotalVettedValidators    uint64
	TotalExitedValidators    uint64
	TotalAddedValidators     uint64
	TotalDepositedValidators uint64
}, error) {
	return _SimpleDVTModule.Contract.GetNodeOperator(&_SimpleDVTModule.CallOpts, _nodeOperatorId, _fullInfo)
}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 _offset, uint256 _limit) view returns(uint256[] nodeOperatorIds)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetNodeOperatorIds(opts *bind.CallOpts, _offset *big.Int, _limit *big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getNodeOperatorIds", _offset, _limit)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 _offset, uint256 _limit) view returns(uint256[] nodeOperatorIds)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetNodeOperatorIds(_offset *big.Int, _limit *big.Int) ([]*big.Int, error) {
	return _SimpleDVTModule.Contract.GetNodeOperatorIds(&_SimpleDVTModule.CallOpts, _offset, _limit)
}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 _offset, uint256 _limit) view returns(uint256[] nodeOperatorIds)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetNodeOperatorIds(_offset *big.Int, _limit *big.Int) ([]*big.Int, error) {
	return _SimpleDVTModule.Contract.GetNodeOperatorIds(&_SimpleDVTModule.CallOpts, _offset, _limit)
}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 _nodeOperatorId) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetNodeOperatorIsActive(opts *bind.CallOpts, _nodeOperatorId *big.Int) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getNodeOperatorIsActive", _nodeOperatorId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 _nodeOperatorId) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetNodeOperatorIsActive(_nodeOperatorId *big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.GetNodeOperatorIsActive(&_SimpleDVTModule.CallOpts, _nodeOperatorId)
}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 _nodeOperatorId) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetNodeOperatorIsActive(_nodeOperatorId *big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.GetNodeOperatorIsActive(&_SimpleDVTModule.CallOpts, _nodeOperatorId)
}

// GetNodeOperatorSummary is a free data retrieval call binding the contract method 0xb3076c3c.
//
// Solidity: function getNodeOperatorSummary(uint256 _nodeOperatorId) view returns(uint256 targetLimitMode, uint256 targetValidatorsCount, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp, uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetNodeOperatorSummary(opts *bind.CallOpts, _nodeOperatorId *big.Int) (struct {
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
	err := _SimpleDVTModule.contract.Call(opts, &out, "getNodeOperatorSummary", _nodeOperatorId)

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
func (_SimpleDVTModule *SimpleDVTModuleSession) GetNodeOperatorSummary(_nodeOperatorId *big.Int) (struct {
	TargetLimitMode            *big.Int
	TargetValidatorsCount      *big.Int
	StuckValidatorsCount       *big.Int
	RefundedValidatorsCount    *big.Int
	StuckPenaltyEndTimestamp   *big.Int
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _SimpleDVTModule.Contract.GetNodeOperatorSummary(&_SimpleDVTModule.CallOpts, _nodeOperatorId)
}

// GetNodeOperatorSummary is a free data retrieval call binding the contract method 0xb3076c3c.
//
// Solidity: function getNodeOperatorSummary(uint256 _nodeOperatorId) view returns(uint256 targetLimitMode, uint256 targetValidatorsCount, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp, uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetNodeOperatorSummary(_nodeOperatorId *big.Int) (struct {
	TargetLimitMode            *big.Int
	TargetValidatorsCount      *big.Int
	StuckValidatorsCount       *big.Int
	RefundedValidatorsCount    *big.Int
	StuckPenaltyEndTimestamp   *big.Int
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _SimpleDVTModule.Contract.GetNodeOperatorSummary(&_SimpleDVTModule.CallOpts, _nodeOperatorId)
}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetNodeOperatorsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getNodeOperatorsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetNodeOperatorsCount() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetNodeOperatorsCount(&_SimpleDVTModule.CallOpts)
}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetNodeOperatorsCount() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetNodeOperatorsCount(&_SimpleDVTModule.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetNonce() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetNonce(&_SimpleDVTModule.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetNonce() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetNonce(&_SimpleDVTModule.CallOpts)
}

// GetRecoveryVault is a free data retrieval call binding the contract method 0x32f0a3b5.
//
// Solidity: function getRecoveryVault() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetRecoveryVault(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getRecoveryVault")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRecoveryVault is a free data retrieval call binding the contract method 0x32f0a3b5.
//
// Solidity: function getRecoveryVault() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetRecoveryVault() (common.Address, error) {
	return _SimpleDVTModule.Contract.GetRecoveryVault(&_SimpleDVTModule.CallOpts)
}

// GetRecoveryVault is a free data retrieval call binding the contract method 0x32f0a3b5.
//
// Solidity: function getRecoveryVault() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetRecoveryVault() (common.Address, error) {
	return _SimpleDVTModule.Contract.GetRecoveryVault(&_SimpleDVTModule.CallOpts)
}

// GetRewardDistributionState is a free data retrieval call binding the contract method 0x6f817294.
//
// Solidity: function getRewardDistributionState() view returns(uint8)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetRewardDistributionState(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getRewardDistributionState")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetRewardDistributionState is a free data retrieval call binding the contract method 0x6f817294.
//
// Solidity: function getRewardDistributionState() view returns(uint8)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetRewardDistributionState() (uint8, error) {
	return _SimpleDVTModule.Contract.GetRewardDistributionState(&_SimpleDVTModule.CallOpts)
}

// GetRewardDistributionState is a free data retrieval call binding the contract method 0x6f817294.
//
// Solidity: function getRewardDistributionState() view returns(uint8)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetRewardDistributionState() (uint8, error) {
	return _SimpleDVTModule.Contract.GetRewardDistributionState(&_SimpleDVTModule.CallOpts)
}

// GetRewardsDistribution is a free data retrieval call binding the contract method 0x62dcfda1.
//
// Solidity: function getRewardsDistribution(uint256 _totalRewardShares) view returns(address[] recipients, uint256[] shares, bool[] penalized)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetRewardsDistribution(opts *bind.CallOpts, _totalRewardShares *big.Int) (struct {
	Recipients []common.Address
	Shares     []*big.Int
	Penalized  []bool
}, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getRewardsDistribution", _totalRewardShares)

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
func (_SimpleDVTModule *SimpleDVTModuleSession) GetRewardsDistribution(_totalRewardShares *big.Int) (struct {
	Recipients []common.Address
	Shares     []*big.Int
	Penalized  []bool
}, error) {
	return _SimpleDVTModule.Contract.GetRewardsDistribution(&_SimpleDVTModule.CallOpts, _totalRewardShares)
}

// GetRewardsDistribution is a free data retrieval call binding the contract method 0x62dcfda1.
//
// Solidity: function getRewardsDistribution(uint256 _totalRewardShares) view returns(address[] recipients, uint256[] shares, bool[] penalized)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetRewardsDistribution(_totalRewardShares *big.Int) (struct {
	Recipients []common.Address
	Shares     []*big.Int
	Penalized  []bool
}, error) {
	return _SimpleDVTModule.Contract.GetRewardsDistribution(&_SimpleDVTModule.CallOpts, _totalRewardShares)
}

// GetSigningKey is a free data retrieval call binding the contract method 0xb449402a.
//
// Solidity: function getSigningKey(uint256 _nodeOperatorId, uint256 _index) view returns(bytes key, bytes depositSignature, bool used)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetSigningKey(opts *bind.CallOpts, _nodeOperatorId *big.Int, _index *big.Int) (struct {
	Key              []byte
	DepositSignature []byte
	Used             bool
}, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getSigningKey", _nodeOperatorId, _index)

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
func (_SimpleDVTModule *SimpleDVTModuleSession) GetSigningKey(_nodeOperatorId *big.Int, _index *big.Int) (struct {
	Key              []byte
	DepositSignature []byte
	Used             bool
}, error) {
	return _SimpleDVTModule.Contract.GetSigningKey(&_SimpleDVTModule.CallOpts, _nodeOperatorId, _index)
}

// GetSigningKey is a free data retrieval call binding the contract method 0xb449402a.
//
// Solidity: function getSigningKey(uint256 _nodeOperatorId, uint256 _index) view returns(bytes key, bytes depositSignature, bool used)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetSigningKey(_nodeOperatorId *big.Int, _index *big.Int) (struct {
	Key              []byte
	DepositSignature []byte
	Used             bool
}, error) {
	return _SimpleDVTModule.Contract.GetSigningKey(&_SimpleDVTModule.CallOpts, _nodeOperatorId, _index)
}

// GetSigningKeys is a free data retrieval call binding the contract method 0x59e25c12.
//
// Solidity: function getSigningKeys(uint256 _nodeOperatorId, uint256 _offset, uint256 _limit) view returns(bytes pubkeys, bytes signatures, bool[] used)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetSigningKeys(opts *bind.CallOpts, _nodeOperatorId *big.Int, _offset *big.Int, _limit *big.Int) (struct {
	Pubkeys    []byte
	Signatures []byte
	Used       []bool
}, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getSigningKeys", _nodeOperatorId, _offset, _limit)

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
func (_SimpleDVTModule *SimpleDVTModuleSession) GetSigningKeys(_nodeOperatorId *big.Int, _offset *big.Int, _limit *big.Int) (struct {
	Pubkeys    []byte
	Signatures []byte
	Used       []bool
}, error) {
	return _SimpleDVTModule.Contract.GetSigningKeys(&_SimpleDVTModule.CallOpts, _nodeOperatorId, _offset, _limit)
}

// GetSigningKeys is a free data retrieval call binding the contract method 0x59e25c12.
//
// Solidity: function getSigningKeys(uint256 _nodeOperatorId, uint256 _offset, uint256 _limit) view returns(bytes pubkeys, bytes signatures, bool[] used)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetSigningKeys(_nodeOperatorId *big.Int, _offset *big.Int, _limit *big.Int) (struct {
	Pubkeys    []byte
	Signatures []byte
	Used       []bool
}, error) {
	return _SimpleDVTModule.Contract.GetSigningKeys(&_SimpleDVTModule.CallOpts, _nodeOperatorId, _offset, _limit)
}

// GetStakingModuleSummary is a free data retrieval call binding the contract method 0x9abddf09.
//
// Solidity: function getStakingModuleSummary() view returns(uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetStakingModuleSummary(opts *bind.CallOpts) (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getStakingModuleSummary")

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
func (_SimpleDVTModule *SimpleDVTModuleSession) GetStakingModuleSummary() (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _SimpleDVTModule.Contract.GetStakingModuleSummary(&_SimpleDVTModule.CallOpts)
}

// GetStakingModuleSummary is a free data retrieval call binding the contract method 0x9abddf09.
//
// Solidity: function getStakingModuleSummary() view returns(uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetStakingModuleSummary() (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _SimpleDVTModule.Contract.GetStakingModuleSummary(&_SimpleDVTModule.CallOpts)
}

// GetStuckPenaltyDelay is a free data retrieval call binding the contract method 0x6da7d0a7.
//
// Solidity: function getStuckPenaltyDelay() pure returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetStuckPenaltyDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getStuckPenaltyDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetStuckPenaltyDelay is a free data retrieval call binding the contract method 0x6da7d0a7.
//
// Solidity: function getStuckPenaltyDelay() pure returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetStuckPenaltyDelay() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetStuckPenaltyDelay(&_SimpleDVTModule.CallOpts)
}

// GetStuckPenaltyDelay is a free data retrieval call binding the contract method 0x6da7d0a7.
//
// Solidity: function getStuckPenaltyDelay() pure returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetStuckPenaltyDelay() (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetStuckPenaltyDelay(&_SimpleDVTModule.CallOpts)
}

// GetTotalSigningKeyCount is a free data retrieval call binding the contract method 0xdb9887ea.
//
// Solidity: function getTotalSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetTotalSigningKeyCount(opts *bind.CallOpts, _nodeOperatorId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getTotalSigningKeyCount", _nodeOperatorId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalSigningKeyCount is a free data retrieval call binding the contract method 0xdb9887ea.
//
// Solidity: function getTotalSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetTotalSigningKeyCount(_nodeOperatorId *big.Int) (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetTotalSigningKeyCount(&_SimpleDVTModule.CallOpts, _nodeOperatorId)
}

// GetTotalSigningKeyCount is a free data retrieval call binding the contract method 0xdb9887ea.
//
// Solidity: function getTotalSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetTotalSigningKeyCount(_nodeOperatorId *big.Int) (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetTotalSigningKeyCount(&_SimpleDVTModule.CallOpts, _nodeOperatorId)
}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetType(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getType")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetType() ([32]byte, error) {
	return _SimpleDVTModule.Contract.GetType(&_SimpleDVTModule.CallOpts)
}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetType() ([32]byte, error) {
	return _SimpleDVTModule.Contract.GetType(&_SimpleDVTModule.CallOpts)
}

// GetUnusedSigningKeyCount is a free data retrieval call binding the contract method 0x8ca7c052.
//
// Solidity: function getUnusedSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCaller) GetUnusedSigningKeyCount(opts *bind.CallOpts, _nodeOperatorId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "getUnusedSigningKeyCount", _nodeOperatorId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUnusedSigningKeyCount is a free data retrieval call binding the contract method 0x8ca7c052.
//
// Solidity: function getUnusedSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleSession) GetUnusedSigningKeyCount(_nodeOperatorId *big.Int) (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetUnusedSigningKeyCount(&_SimpleDVTModule.CallOpts, _nodeOperatorId)
}

// GetUnusedSigningKeyCount is a free data retrieval call binding the contract method 0x8ca7c052.
//
// Solidity: function getUnusedSigningKeyCount(uint256 _nodeOperatorId) view returns(uint256)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) GetUnusedSigningKeyCount(_nodeOperatorId *big.Int) (*big.Int, error) {
	return _SimpleDVTModule.Contract.GetUnusedSigningKeyCount(&_SimpleDVTModule.CallOpts, _nodeOperatorId)
}

// HasInitialized is a free data retrieval call binding the contract method 0x0803fac0.
//
// Solidity: function hasInitialized() view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) HasInitialized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "hasInitialized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasInitialized is a free data retrieval call binding the contract method 0x0803fac0.
//
// Solidity: function hasInitialized() view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) HasInitialized() (bool, error) {
	return _SimpleDVTModule.Contract.HasInitialized(&_SimpleDVTModule.CallOpts)
}

// HasInitialized is a free data retrieval call binding the contract method 0x0803fac0.
//
// Solidity: function hasInitialized() view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) HasInitialized() (bool, error) {
	return _SimpleDVTModule.Contract.HasInitialized(&_SimpleDVTModule.CallOpts)
}

// IsOperatorPenalized is a free data retrieval call binding the contract method 0x75049ad8.
//
// Solidity: function isOperatorPenalized(uint256 ) pure returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) IsOperatorPenalized(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "isOperatorPenalized", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorPenalized is a free data retrieval call binding the contract method 0x75049ad8.
//
// Solidity: function isOperatorPenalized(uint256 ) pure returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) IsOperatorPenalized(arg0 *big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.IsOperatorPenalized(&_SimpleDVTModule.CallOpts, arg0)
}

// IsOperatorPenalized is a free data retrieval call binding the contract method 0x75049ad8.
//
// Solidity: function isOperatorPenalized(uint256 ) pure returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) IsOperatorPenalized(arg0 *big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.IsOperatorPenalized(&_SimpleDVTModule.CallOpts, arg0)
}

// IsOperatorPenaltyCleared is a free data retrieval call binding the contract method 0xfbc77ef1.
//
// Solidity: function isOperatorPenaltyCleared(uint256 ) pure returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) IsOperatorPenaltyCleared(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "isOperatorPenaltyCleared", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorPenaltyCleared is a free data retrieval call binding the contract method 0xfbc77ef1.
//
// Solidity: function isOperatorPenaltyCleared(uint256 ) pure returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) IsOperatorPenaltyCleared(arg0 *big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.IsOperatorPenaltyCleared(&_SimpleDVTModule.CallOpts, arg0)
}

// IsOperatorPenaltyCleared is a free data retrieval call binding the contract method 0xfbc77ef1.
//
// Solidity: function isOperatorPenaltyCleared(uint256 ) pure returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) IsOperatorPenaltyCleared(arg0 *big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.IsOperatorPenaltyCleared(&_SimpleDVTModule.CallOpts, arg0)
}

// IsPetrified is a free data retrieval call binding the contract method 0xde4796ed.
//
// Solidity: function isPetrified() view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) IsPetrified(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "isPetrified")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsPetrified is a free data retrieval call binding the contract method 0xde4796ed.
//
// Solidity: function isPetrified() view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) IsPetrified() (bool, error) {
	return _SimpleDVTModule.Contract.IsPetrified(&_SimpleDVTModule.CallOpts)
}

// IsPetrified is a free data retrieval call binding the contract method 0xde4796ed.
//
// Solidity: function isPetrified() view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) IsPetrified() (bool, error) {
	return _SimpleDVTModule.Contract.IsPetrified(&_SimpleDVTModule.CallOpts)
}

// IsValidatorExitDelayPenaltyApplicable is a free data retrieval call binding the contract method 0x83b57a4e.
//
// Solidity: function isValidatorExitDelayPenaltyApplicable(uint256 , uint256 _proofSlotTimestamp, bytes _publicKey, uint256 _eligibleToExitInSec) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) IsValidatorExitDelayPenaltyApplicable(opts *bind.CallOpts, arg0 *big.Int, _proofSlotTimestamp *big.Int, _publicKey []byte, _eligibleToExitInSec *big.Int) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "isValidatorExitDelayPenaltyApplicable", arg0, _proofSlotTimestamp, _publicKey, _eligibleToExitInSec)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidatorExitDelayPenaltyApplicable is a free data retrieval call binding the contract method 0x83b57a4e.
//
// Solidity: function isValidatorExitDelayPenaltyApplicable(uint256 , uint256 _proofSlotTimestamp, bytes _publicKey, uint256 _eligibleToExitInSec) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) IsValidatorExitDelayPenaltyApplicable(arg0 *big.Int, _proofSlotTimestamp *big.Int, _publicKey []byte, _eligibleToExitInSec *big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.IsValidatorExitDelayPenaltyApplicable(&_SimpleDVTModule.CallOpts, arg0, _proofSlotTimestamp, _publicKey, _eligibleToExitInSec)
}

// IsValidatorExitDelayPenaltyApplicable is a free data retrieval call binding the contract method 0x83b57a4e.
//
// Solidity: function isValidatorExitDelayPenaltyApplicable(uint256 , uint256 _proofSlotTimestamp, bytes _publicKey, uint256 _eligibleToExitInSec) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) IsValidatorExitDelayPenaltyApplicable(arg0 *big.Int, _proofSlotTimestamp *big.Int, _publicKey []byte, _eligibleToExitInSec *big.Int) (bool, error) {
	return _SimpleDVTModule.Contract.IsValidatorExitDelayPenaltyApplicable(&_SimpleDVTModule.CallOpts, arg0, _proofSlotTimestamp, _publicKey, _eligibleToExitInSec)
}

// IsValidatorExitingKeyReported is a free data retrieval call binding the contract method 0xba2406fd.
//
// Solidity: function isValidatorExitingKeyReported(bytes _publicKey) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCaller) IsValidatorExitingKeyReported(opts *bind.CallOpts, _publicKey []byte) (bool, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "isValidatorExitingKeyReported", _publicKey)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidatorExitingKeyReported is a free data retrieval call binding the contract method 0xba2406fd.
//
// Solidity: function isValidatorExitingKeyReported(bytes _publicKey) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleSession) IsValidatorExitingKeyReported(_publicKey []byte) (bool, error) {
	return _SimpleDVTModule.Contract.IsValidatorExitingKeyReported(&_SimpleDVTModule.CallOpts, _publicKey)
}

// IsValidatorExitingKeyReported is a free data retrieval call binding the contract method 0xba2406fd.
//
// Solidity: function isValidatorExitingKeyReported(bytes _publicKey) view returns(bool)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) IsValidatorExitingKeyReported(_publicKey []byte) (bool, error) {
	return _SimpleDVTModule.Contract.IsValidatorExitingKeyReported(&_SimpleDVTModule.CallOpts, _publicKey)
}

// Kernel is a free data retrieval call binding the contract method 0xd4aae0c4.
//
// Solidity: function kernel() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCaller) Kernel(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SimpleDVTModule.contract.Call(opts, &out, "kernel")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Kernel is a free data retrieval call binding the contract method 0xd4aae0c4.
//
// Solidity: function kernel() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleSession) Kernel() (common.Address, error) {
	return _SimpleDVTModule.Contract.Kernel(&_SimpleDVTModule.CallOpts)
}

// Kernel is a free data retrieval call binding the contract method 0xd4aae0c4.
//
// Solidity: function kernel() view returns(address)
func (_SimpleDVTModule *SimpleDVTModuleCallerSession) Kernel() (common.Address, error) {
	return _SimpleDVTModule.Contract.Kernel(&_SimpleDVTModule.CallOpts)
}

// ActivateNodeOperator is a paid mutator transaction binding the contract method 0x91dcd6b2.
//
// Solidity: function activateNodeOperator(uint256 _nodeOperatorId) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) ActivateNodeOperator(opts *bind.TransactOpts, _nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "activateNodeOperator", _nodeOperatorId)
}

// ActivateNodeOperator is a paid mutator transaction binding the contract method 0x91dcd6b2.
//
// Solidity: function activateNodeOperator(uint256 _nodeOperatorId) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) ActivateNodeOperator(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.ActivateNodeOperator(&_SimpleDVTModule.TransactOpts, _nodeOperatorId)
}

// ActivateNodeOperator is a paid mutator transaction binding the contract method 0x91dcd6b2.
//
// Solidity: function activateNodeOperator(uint256 _nodeOperatorId) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) ActivateNodeOperator(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.ActivateNodeOperator(&_SimpleDVTModule.TransactOpts, _nodeOperatorId)
}

// AddNodeOperator is a paid mutator transaction binding the contract method 0x85fa63d7.
//
// Solidity: function addNodeOperator(string _name, address _rewardAddress) returns(uint256 id)
func (_SimpleDVTModule *SimpleDVTModuleTransactor) AddNodeOperator(opts *bind.TransactOpts, _name string, _rewardAddress common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "addNodeOperator", _name, _rewardAddress)
}

// AddNodeOperator is a paid mutator transaction binding the contract method 0x85fa63d7.
//
// Solidity: function addNodeOperator(string _name, address _rewardAddress) returns(uint256 id)
func (_SimpleDVTModule *SimpleDVTModuleSession) AddNodeOperator(_name string, _rewardAddress common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.AddNodeOperator(&_SimpleDVTModule.TransactOpts, _name, _rewardAddress)
}

// AddNodeOperator is a paid mutator transaction binding the contract method 0x85fa63d7.
//
// Solidity: function addNodeOperator(string _name, address _rewardAddress) returns(uint256 id)
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) AddNodeOperator(_name string, _rewardAddress common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.AddNodeOperator(&_SimpleDVTModule.TransactOpts, _name, _rewardAddress)
}

// AddSigningKeys is a paid mutator transaction binding the contract method 0x096b7b35.
//
// Solidity: function addSigningKeys(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) AddSigningKeys(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "addSigningKeys", _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeys is a paid mutator transaction binding the contract method 0x096b7b35.
//
// Solidity: function addSigningKeys(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) AddSigningKeys(_nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.AddSigningKeys(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeys is a paid mutator transaction binding the contract method 0x096b7b35.
//
// Solidity: function addSigningKeys(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) AddSigningKeys(_nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.AddSigningKeys(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x805911ae.
//
// Solidity: function addSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) AddSigningKeysOperatorBH(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "addSigningKeysOperatorBH", _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x805911ae.
//
// Solidity: function addSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) AddSigningKeysOperatorBH(_nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.AddSigningKeysOperatorBH(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// AddSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x805911ae.
//
// Solidity: function addSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _keysCount, bytes _publicKeys, bytes _signatures) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) AddSigningKeysOperatorBH(_nodeOperatorId *big.Int, _keysCount *big.Int, _publicKeys []byte, _signatures []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.AddSigningKeysOperatorBH(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _keysCount, _publicKeys, _signatures)
}

// DeactivateNodeOperator is a paid mutator transaction binding the contract method 0x75a080d5.
//
// Solidity: function deactivateNodeOperator(uint256 _nodeOperatorId) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) DeactivateNodeOperator(opts *bind.TransactOpts, _nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "deactivateNodeOperator", _nodeOperatorId)
}

// DeactivateNodeOperator is a paid mutator transaction binding the contract method 0x75a080d5.
//
// Solidity: function deactivateNodeOperator(uint256 _nodeOperatorId) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) DeactivateNodeOperator(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.DeactivateNodeOperator(&_SimpleDVTModule.TransactOpts, _nodeOperatorId)
}

// DeactivateNodeOperator is a paid mutator transaction binding the contract method 0x75a080d5.
//
// Solidity: function deactivateNodeOperator(uint256 _nodeOperatorId) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) DeactivateNodeOperator(_nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.DeactivateNodeOperator(&_SimpleDVTModule.TransactOpts, _nodeOperatorId)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes _nodeOperatorIds, bytes _vettedSigningKeysCounts) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) DecreaseVettedSigningKeysCount(opts *bind.TransactOpts, _nodeOperatorIds []byte, _vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "decreaseVettedSigningKeysCount", _nodeOperatorIds, _vettedSigningKeysCounts)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes _nodeOperatorIds, bytes _vettedSigningKeysCounts) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) DecreaseVettedSigningKeysCount(_nodeOperatorIds []byte, _vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.DecreaseVettedSigningKeysCount(&_SimpleDVTModule.TransactOpts, _nodeOperatorIds, _vettedSigningKeysCounts)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes _nodeOperatorIds, bytes _vettedSigningKeysCounts) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) DecreaseVettedSigningKeysCount(_nodeOperatorIds []byte, _vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.DecreaseVettedSigningKeysCount(&_SimpleDVTModule.TransactOpts, _nodeOperatorIds, _vettedSigningKeysCounts)
}

// DistributeReward is a paid mutator transaction binding the contract method 0x8f73c5ae.
//
// Solidity: function distributeReward() returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) DistributeReward(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "distributeReward")
}

// DistributeReward is a paid mutator transaction binding the contract method 0x8f73c5ae.
//
// Solidity: function distributeReward() returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) DistributeReward() (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.DistributeReward(&_SimpleDVTModule.TransactOpts)
}

// DistributeReward is a paid mutator transaction binding the contract method 0x8f73c5ae.
//
// Solidity: function distributeReward() returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) DistributeReward() (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.DistributeReward(&_SimpleDVTModule.TransactOpts)
}

// FinalizeUpgradeV4 is a paid mutator transaction binding the contract method 0x8ee1c0a8.
//
// Solidity: function finalizeUpgrade_v4(uint256 _exitDeadlineThresholdInSeconds) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) FinalizeUpgradeV4(opts *bind.TransactOpts, _exitDeadlineThresholdInSeconds *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "finalizeUpgrade_v4", _exitDeadlineThresholdInSeconds)
}

// FinalizeUpgradeV4 is a paid mutator transaction binding the contract method 0x8ee1c0a8.
//
// Solidity: function finalizeUpgrade_v4(uint256 _exitDeadlineThresholdInSeconds) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) FinalizeUpgradeV4(_exitDeadlineThresholdInSeconds *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.FinalizeUpgradeV4(&_SimpleDVTModule.TransactOpts, _exitDeadlineThresholdInSeconds)
}

// FinalizeUpgradeV4 is a paid mutator transaction binding the contract method 0x8ee1c0a8.
//
// Solidity: function finalizeUpgrade_v4(uint256 _exitDeadlineThresholdInSeconds) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) FinalizeUpgradeV4(_exitDeadlineThresholdInSeconds *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.FinalizeUpgradeV4(&_SimpleDVTModule.TransactOpts, _exitDeadlineThresholdInSeconds)
}

// Initialize is a paid mutator transaction binding the contract method 0x684560a2.
//
// Solidity: function initialize(address _locator, bytes32 _type, uint256 _exitDeadlineThresholdInSeconds) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) Initialize(opts *bind.TransactOpts, _locator common.Address, _type [32]byte, _exitDeadlineThresholdInSeconds *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "initialize", _locator, _type, _exitDeadlineThresholdInSeconds)
}

// Initialize is a paid mutator transaction binding the contract method 0x684560a2.
//
// Solidity: function initialize(address _locator, bytes32 _type, uint256 _exitDeadlineThresholdInSeconds) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) Initialize(_locator common.Address, _type [32]byte, _exitDeadlineThresholdInSeconds *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.Initialize(&_SimpleDVTModule.TransactOpts, _locator, _type, _exitDeadlineThresholdInSeconds)
}

// Initialize is a paid mutator transaction binding the contract method 0x684560a2.
//
// Solidity: function initialize(address _locator, bytes32 _type, uint256 _exitDeadlineThresholdInSeconds) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) Initialize(_locator common.Address, _type [32]byte, _exitDeadlineThresholdInSeconds *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.Initialize(&_SimpleDVTModule.TransactOpts, _locator, _type, _exitDeadlineThresholdInSeconds)
}

// InvalidateReadyToDepositKeysRange is a paid mutator transaction binding the contract method 0x65cc369a.
//
// Solidity: function invalidateReadyToDepositKeysRange(uint256 _indexFrom, uint256 _indexTo) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) InvalidateReadyToDepositKeysRange(opts *bind.TransactOpts, _indexFrom *big.Int, _indexTo *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "invalidateReadyToDepositKeysRange", _indexFrom, _indexTo)
}

// InvalidateReadyToDepositKeysRange is a paid mutator transaction binding the contract method 0x65cc369a.
//
// Solidity: function invalidateReadyToDepositKeysRange(uint256 _indexFrom, uint256 _indexTo) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) InvalidateReadyToDepositKeysRange(_indexFrom *big.Int, _indexTo *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.InvalidateReadyToDepositKeysRange(&_SimpleDVTModule.TransactOpts, _indexFrom, _indexTo)
}

// InvalidateReadyToDepositKeysRange is a paid mutator transaction binding the contract method 0x65cc369a.
//
// Solidity: function invalidateReadyToDepositKeysRange(uint256 _indexFrom, uint256 _indexTo) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) InvalidateReadyToDepositKeysRange(_indexFrom *big.Int, _indexTo *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.InvalidateReadyToDepositKeysRange(&_SimpleDVTModule.TransactOpts, _indexFrom, _indexTo)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 _depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_SimpleDVTModule *SimpleDVTModuleTransactor) ObtainDepositData(opts *bind.TransactOpts, _depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "obtainDepositData", _depositsCount, arg1)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 _depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_SimpleDVTModule *SimpleDVTModuleSession) ObtainDepositData(_depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.ObtainDepositData(&_SimpleDVTModule.TransactOpts, _depositsCount, arg1)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 _depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) ObtainDepositData(_depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.ObtainDepositData(&_SimpleDVTModule.TransactOpts, _depositsCount, arg1)
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) OnExitedAndStuckValidatorsCountsUpdated(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "onExitedAndStuckValidatorsCountsUpdated")
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) OnExitedAndStuckValidatorsCountsUpdated() (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.OnExitedAndStuckValidatorsCountsUpdated(&_SimpleDVTModule.TransactOpts)
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) OnExitedAndStuckValidatorsCountsUpdated() (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.OnExitedAndStuckValidatorsCountsUpdated(&_SimpleDVTModule.TransactOpts)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 ) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) OnRewardsMinted(opts *bind.TransactOpts, arg0 *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "onRewardsMinted", arg0)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 ) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) OnRewardsMinted(arg0 *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.OnRewardsMinted(&_SimpleDVTModule.TransactOpts, arg0)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 ) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) OnRewardsMinted(arg0 *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.OnRewardsMinted(&_SimpleDVTModule.TransactOpts, arg0)
}

// OnValidatorExitTriggered is a paid mutator transaction binding the contract method 0x693cc600.
//
// Solidity: function onValidatorExitTriggered(uint256 _nodeOperatorId, bytes _publicKey, uint256 _withdrawalRequestPaidFee, uint256 _exitType) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) OnValidatorExitTriggered(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _publicKey []byte, _withdrawalRequestPaidFee *big.Int, _exitType *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "onValidatorExitTriggered", _nodeOperatorId, _publicKey, _withdrawalRequestPaidFee, _exitType)
}

// OnValidatorExitTriggered is a paid mutator transaction binding the contract method 0x693cc600.
//
// Solidity: function onValidatorExitTriggered(uint256 _nodeOperatorId, bytes _publicKey, uint256 _withdrawalRequestPaidFee, uint256 _exitType) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) OnValidatorExitTriggered(_nodeOperatorId *big.Int, _publicKey []byte, _withdrawalRequestPaidFee *big.Int, _exitType *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.OnValidatorExitTriggered(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _publicKey, _withdrawalRequestPaidFee, _exitType)
}

// OnValidatorExitTriggered is a paid mutator transaction binding the contract method 0x693cc600.
//
// Solidity: function onValidatorExitTriggered(uint256 _nodeOperatorId, bytes _publicKey, uint256 _withdrawalRequestPaidFee, uint256 _exitType) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) OnValidatorExitTriggered(_nodeOperatorId *big.Int, _publicKey []byte, _withdrawalRequestPaidFee *big.Int, _exitType *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.OnValidatorExitTriggered(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _publicKey, _withdrawalRequestPaidFee, _exitType)
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) OnWithdrawalCredentialsChanged(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "onWithdrawalCredentialsChanged")
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) OnWithdrawalCredentialsChanged() (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.OnWithdrawalCredentialsChanged(&_SimpleDVTModule.TransactOpts)
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) OnWithdrawalCredentialsChanged() (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.OnWithdrawalCredentialsChanged(&_SimpleDVTModule.TransactOpts)
}

// RemoveSigningKey is a paid mutator transaction binding the contract method 0x6ef355f1.
//
// Solidity: function removeSigningKey(uint256 _nodeOperatorId, uint256 _index) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) RemoveSigningKey(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "removeSigningKey", _nodeOperatorId, _index)
}

// RemoveSigningKey is a paid mutator transaction binding the contract method 0x6ef355f1.
//
// Solidity: function removeSigningKey(uint256 _nodeOperatorId, uint256 _index) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) RemoveSigningKey(_nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.RemoveSigningKey(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _index)
}

// RemoveSigningKey is a paid mutator transaction binding the contract method 0x6ef355f1.
//
// Solidity: function removeSigningKey(uint256 _nodeOperatorId, uint256 _index) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) RemoveSigningKey(_nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.RemoveSigningKey(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _index)
}

// RemoveSigningKeyOperatorBH is a paid mutator transaction binding the contract method 0xed5cfa41.
//
// Solidity: function removeSigningKeyOperatorBH(uint256 _nodeOperatorId, uint256 _index) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) RemoveSigningKeyOperatorBH(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "removeSigningKeyOperatorBH", _nodeOperatorId, _index)
}

// RemoveSigningKeyOperatorBH is a paid mutator transaction binding the contract method 0xed5cfa41.
//
// Solidity: function removeSigningKeyOperatorBH(uint256 _nodeOperatorId, uint256 _index) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) RemoveSigningKeyOperatorBH(_nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.RemoveSigningKeyOperatorBH(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _index)
}

// RemoveSigningKeyOperatorBH is a paid mutator transaction binding the contract method 0xed5cfa41.
//
// Solidity: function removeSigningKeyOperatorBH(uint256 _nodeOperatorId, uint256 _index) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) RemoveSigningKeyOperatorBH(_nodeOperatorId *big.Int, _index *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.RemoveSigningKeyOperatorBH(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _index)
}

// RemoveSigningKeys is a paid mutator transaction binding the contract method 0x7038b141.
//
// Solidity: function removeSigningKeys(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) RemoveSigningKeys(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "removeSigningKeys", _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeys is a paid mutator transaction binding the contract method 0x7038b141.
//
// Solidity: function removeSigningKeys(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) RemoveSigningKeys(_nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.RemoveSigningKeys(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeys is a paid mutator transaction binding the contract method 0x7038b141.
//
// Solidity: function removeSigningKeys(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) RemoveSigningKeys(_nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.RemoveSigningKeys(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x5ddde810.
//
// Solidity: function removeSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) RemoveSigningKeysOperatorBH(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "removeSigningKeysOperatorBH", _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x5ddde810.
//
// Solidity: function removeSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) RemoveSigningKeysOperatorBH(_nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.RemoveSigningKeysOperatorBH(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _fromIndex, _keysCount)
}

// RemoveSigningKeysOperatorBH is a paid mutator transaction binding the contract method 0x5ddde810.
//
// Solidity: function removeSigningKeysOperatorBH(uint256 _nodeOperatorId, uint256 _fromIndex, uint256 _keysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) RemoveSigningKeysOperatorBH(_nodeOperatorId *big.Int, _fromIndex *big.Int, _keysCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.RemoveSigningKeysOperatorBH(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _fromIndex, _keysCount)
}

// ReportValidatorExitDelay is a paid mutator transaction binding the contract method 0x57f9c341.
//
// Solidity: function reportValidatorExitDelay(uint256 _nodeOperatorId, uint256 _proofSlotTimestamp, bytes _publicKey, uint256 _eligibleToExitInSec) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) ReportValidatorExitDelay(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _proofSlotTimestamp *big.Int, _publicKey []byte, _eligibleToExitInSec *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "reportValidatorExitDelay", _nodeOperatorId, _proofSlotTimestamp, _publicKey, _eligibleToExitInSec)
}

// ReportValidatorExitDelay is a paid mutator transaction binding the contract method 0x57f9c341.
//
// Solidity: function reportValidatorExitDelay(uint256 _nodeOperatorId, uint256 _proofSlotTimestamp, bytes _publicKey, uint256 _eligibleToExitInSec) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) ReportValidatorExitDelay(_nodeOperatorId *big.Int, _proofSlotTimestamp *big.Int, _publicKey []byte, _eligibleToExitInSec *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.ReportValidatorExitDelay(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _proofSlotTimestamp, _publicKey, _eligibleToExitInSec)
}

// ReportValidatorExitDelay is a paid mutator transaction binding the contract method 0x57f9c341.
//
// Solidity: function reportValidatorExitDelay(uint256 _nodeOperatorId, uint256 _proofSlotTimestamp, bytes _publicKey, uint256 _eligibleToExitInSec) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) ReportValidatorExitDelay(_nodeOperatorId *big.Int, _proofSlotTimestamp *big.Int, _publicKey []byte, _eligibleToExitInSec *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.ReportValidatorExitDelay(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _proofSlotTimestamp, _publicKey, _eligibleToExitInSec)
}

// SetExitDeadlineThreshold is a paid mutator transaction binding the contract method 0xbe1575d3.
//
// Solidity: function setExitDeadlineThreshold(uint256 _threshold, uint256 _lateReportingWindow) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) SetExitDeadlineThreshold(opts *bind.TransactOpts, _threshold *big.Int, _lateReportingWindow *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "setExitDeadlineThreshold", _threshold, _lateReportingWindow)
}

// SetExitDeadlineThreshold is a paid mutator transaction binding the contract method 0xbe1575d3.
//
// Solidity: function setExitDeadlineThreshold(uint256 _threshold, uint256 _lateReportingWindow) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) SetExitDeadlineThreshold(_threshold *big.Int, _lateReportingWindow *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SetExitDeadlineThreshold(&_SimpleDVTModule.TransactOpts, _threshold, _lateReportingWindow)
}

// SetExitDeadlineThreshold is a paid mutator transaction binding the contract method 0xbe1575d3.
//
// Solidity: function setExitDeadlineThreshold(uint256 _threshold, uint256 _lateReportingWindow) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) SetExitDeadlineThreshold(_threshold *big.Int, _lateReportingWindow *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SetExitDeadlineThreshold(&_SimpleDVTModule.TransactOpts, _threshold, _lateReportingWindow)
}

// SetNodeOperatorName is a paid mutator transaction binding the contract method 0x5e57d742.
//
// Solidity: function setNodeOperatorName(uint256 _nodeOperatorId, string _name) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) SetNodeOperatorName(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _name string) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "setNodeOperatorName", _nodeOperatorId, _name)
}

// SetNodeOperatorName is a paid mutator transaction binding the contract method 0x5e57d742.
//
// Solidity: function setNodeOperatorName(uint256 _nodeOperatorId, string _name) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) SetNodeOperatorName(_nodeOperatorId *big.Int, _name string) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SetNodeOperatorName(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _name)
}

// SetNodeOperatorName is a paid mutator transaction binding the contract method 0x5e57d742.
//
// Solidity: function setNodeOperatorName(uint256 _nodeOperatorId, string _name) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) SetNodeOperatorName(_nodeOperatorId *big.Int, _name string) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SetNodeOperatorName(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _name)
}

// SetNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x973e9328.
//
// Solidity: function setNodeOperatorRewardAddress(uint256 _nodeOperatorId, address _rewardAddress) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) SetNodeOperatorRewardAddress(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _rewardAddress common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "setNodeOperatorRewardAddress", _nodeOperatorId, _rewardAddress)
}

// SetNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x973e9328.
//
// Solidity: function setNodeOperatorRewardAddress(uint256 _nodeOperatorId, address _rewardAddress) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) SetNodeOperatorRewardAddress(_nodeOperatorId *big.Int, _rewardAddress common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SetNodeOperatorRewardAddress(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _rewardAddress)
}

// SetNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x973e9328.
//
// Solidity: function setNodeOperatorRewardAddress(uint256 _nodeOperatorId, address _rewardAddress) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) SetNodeOperatorRewardAddress(_nodeOperatorId *big.Int, _rewardAddress common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SetNodeOperatorRewardAddress(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _rewardAddress)
}

// SetNodeOperatorStakingLimit is a paid mutator transaction binding the contract method 0xae962acf.
//
// Solidity: function setNodeOperatorStakingLimit(uint256 _nodeOperatorId, uint64 _vettedSigningKeysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) SetNodeOperatorStakingLimit(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _vettedSigningKeysCount uint64) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "setNodeOperatorStakingLimit", _nodeOperatorId, _vettedSigningKeysCount)
}

// SetNodeOperatorStakingLimit is a paid mutator transaction binding the contract method 0xae962acf.
//
// Solidity: function setNodeOperatorStakingLimit(uint256 _nodeOperatorId, uint64 _vettedSigningKeysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) SetNodeOperatorStakingLimit(_nodeOperatorId *big.Int, _vettedSigningKeysCount uint64) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SetNodeOperatorStakingLimit(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _vettedSigningKeysCount)
}

// SetNodeOperatorStakingLimit is a paid mutator transaction binding the contract method 0xae962acf.
//
// Solidity: function setNodeOperatorStakingLimit(uint256 _nodeOperatorId, uint64 _vettedSigningKeysCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) SetNodeOperatorStakingLimit(_nodeOperatorId *big.Int, _vettedSigningKeysCount uint64) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.SetNodeOperatorStakingLimit(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _vettedSigningKeysCount)
}

// TransferToVault is a paid mutator transaction binding the contract method 0x9d4941d8.
//
// Solidity: function transferToVault(address _token) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) TransferToVault(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "transferToVault", _token)
}

// TransferToVault is a paid mutator transaction binding the contract method 0x9d4941d8.
//
// Solidity: function transferToVault(address _token) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) TransferToVault(_token common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.TransferToVault(&_SimpleDVTModule.TransactOpts, _token)
}

// TransferToVault is a paid mutator transaction binding the contract method 0x9d4941d8.
//
// Solidity: function transferToVault(address _token) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) TransferToVault(_token common.Address) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.TransferToVault(&_SimpleDVTModule.TransactOpts, _token)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0x94120368.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 _nodeOperatorId, uint256 _exitedValidatorsCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) UnsafeUpdateValidatorsCount(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _exitedValidatorsCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "unsafeUpdateValidatorsCount", _nodeOperatorId, _exitedValidatorsCount)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0x94120368.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 _nodeOperatorId, uint256 _exitedValidatorsCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) UnsafeUpdateValidatorsCount(_nodeOperatorId *big.Int, _exitedValidatorsCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.UnsafeUpdateValidatorsCount(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _exitedValidatorsCount)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0x94120368.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 _nodeOperatorId, uint256 _exitedValidatorsCount) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) UnsafeUpdateValidatorsCount(_nodeOperatorId *big.Int, _exitedValidatorsCount *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.UnsafeUpdateValidatorsCount(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _exitedValidatorsCount)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes _nodeOperatorIds, bytes _exitedValidatorsCounts) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) UpdateExitedValidatorsCount(opts *bind.TransactOpts, _nodeOperatorIds []byte, _exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "updateExitedValidatorsCount", _nodeOperatorIds, _exitedValidatorsCounts)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes _nodeOperatorIds, bytes _exitedValidatorsCounts) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) UpdateExitedValidatorsCount(_nodeOperatorIds []byte, _exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.UpdateExitedValidatorsCount(&_SimpleDVTModule.TransactOpts, _nodeOperatorIds, _exitedValidatorsCounts)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes _nodeOperatorIds, bytes _exitedValidatorsCounts) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) UpdateExitedValidatorsCount(_nodeOperatorIds []byte, _exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.UpdateExitedValidatorsCount(&_SimpleDVTModule.TransactOpts, _nodeOperatorIds, _exitedValidatorsCounts)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, uint256 _targetLimitMode, uint256 _targetLimit) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) UpdateTargetValidatorsLimits(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _targetLimitMode *big.Int, _targetLimit *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "updateTargetValidatorsLimits", _nodeOperatorId, _targetLimitMode, _targetLimit)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, uint256 _targetLimitMode, uint256 _targetLimit) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) UpdateTargetValidatorsLimits(_nodeOperatorId *big.Int, _targetLimitMode *big.Int, _targetLimit *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.UpdateTargetValidatorsLimits(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _targetLimitMode, _targetLimit)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, uint256 _targetLimitMode, uint256 _targetLimit) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) UpdateTargetValidatorsLimits(_nodeOperatorId *big.Int, _targetLimitMode *big.Int, _targetLimit *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.UpdateTargetValidatorsLimits(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _targetLimitMode, _targetLimit)
}

// UpdateTargetValidatorsLimits0 is a paid mutator transaction binding the contract method 0xa9e7a846.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, bool _isTargetLimitActive, uint256 _targetLimit) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactor) UpdateTargetValidatorsLimits0(opts *bind.TransactOpts, _nodeOperatorId *big.Int, _isTargetLimitActive bool, _targetLimit *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.contract.Transact(opts, "updateTargetValidatorsLimits0", _nodeOperatorId, _isTargetLimitActive, _targetLimit)
}

// UpdateTargetValidatorsLimits0 is a paid mutator transaction binding the contract method 0xa9e7a846.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, bool _isTargetLimitActive, uint256 _targetLimit) returns()
func (_SimpleDVTModule *SimpleDVTModuleSession) UpdateTargetValidatorsLimits0(_nodeOperatorId *big.Int, _isTargetLimitActive bool, _targetLimit *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.UpdateTargetValidatorsLimits0(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _isTargetLimitActive, _targetLimit)
}

// UpdateTargetValidatorsLimits0 is a paid mutator transaction binding the contract method 0xa9e7a846.
//
// Solidity: function updateTargetValidatorsLimits(uint256 _nodeOperatorId, bool _isTargetLimitActive, uint256 _targetLimit) returns()
func (_SimpleDVTModule *SimpleDVTModuleTransactorSession) UpdateTargetValidatorsLimits0(_nodeOperatorId *big.Int, _isTargetLimitActive bool, _targetLimit *big.Int) (*types.Transaction, error) {
	return _SimpleDVTModule.Contract.UpdateTargetValidatorsLimits0(&_SimpleDVTModule.TransactOpts, _nodeOperatorId, _isTargetLimitActive, _targetLimit)
}

// SimpleDVTModuleContractVersionSetIterator is returned from FilterContractVersionSet and is used to iterate over the raw logs and unpacked data for ContractVersionSet events raised by the SimpleDVTModule contract.
type SimpleDVTModuleContractVersionSetIterator struct {
	Event *SimpleDVTModuleContractVersionSet // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleContractVersionSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleContractVersionSet)
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
		it.Event = new(SimpleDVTModuleContractVersionSet)
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
func (it *SimpleDVTModuleContractVersionSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleContractVersionSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleContractVersionSet represents a ContractVersionSet event raised by the SimpleDVTModule contract.
type SimpleDVTModuleContractVersionSet struct {
	Version *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterContractVersionSet is a free log retrieval operation binding the contract event 0xfddcded6b4f4730c226821172046b48372d3cd963c159701ae1b7c3bcac541bb.
//
// Solidity: event ContractVersionSet(uint256 version)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterContractVersionSet(opts *bind.FilterOpts) (*SimpleDVTModuleContractVersionSetIterator, error) {

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "ContractVersionSet")
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleContractVersionSetIterator{contract: _SimpleDVTModule.contract, event: "ContractVersionSet", logs: logs, sub: sub}, nil
}

// WatchContractVersionSet is a free log subscription operation binding the contract event 0xfddcded6b4f4730c226821172046b48372d3cd963c159701ae1b7c3bcac541bb.
//
// Solidity: event ContractVersionSet(uint256 version)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchContractVersionSet(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleContractVersionSet) (event.Subscription, error) {

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "ContractVersionSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleContractVersionSet)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "ContractVersionSet", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseContractVersionSet(log types.Log) (*SimpleDVTModuleContractVersionSet, error) {
	event := new(SimpleDVTModuleContractVersionSet)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "ContractVersionSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleDepositedSigningKeysCountChangedIterator is returned from FilterDepositedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for DepositedSigningKeysCountChanged events raised by the SimpleDVTModule contract.
type SimpleDVTModuleDepositedSigningKeysCountChangedIterator struct {
	Event *SimpleDVTModuleDepositedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleDepositedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleDepositedSigningKeysCountChanged)
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
		it.Event = new(SimpleDVTModuleDepositedSigningKeysCountChanged)
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
func (it *SimpleDVTModuleDepositedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleDepositedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleDepositedSigningKeysCountChanged represents a DepositedSigningKeysCountChanged event raised by the SimpleDVTModule contract.
type SimpleDVTModuleDepositedSigningKeysCountChanged struct {
	NodeOperatorId           *big.Int
	DepositedValidatorsCount *big.Int
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterDepositedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x24eb1c9e765ba41accf9437300ea91ece5ed3f897ec3cdee0e9debd7fe309b78.
//
// Solidity: event DepositedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositedValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterDepositedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleDepositedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "DepositedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleDepositedSigningKeysCountChangedIterator{contract: _SimpleDVTModule.contract, event: "DepositedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchDepositedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x24eb1c9e765ba41accf9437300ea91ece5ed3f897ec3cdee0e9debd7fe309b78.
//
// Solidity: event DepositedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositedValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchDepositedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleDepositedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "DepositedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleDepositedSigningKeysCountChanged)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "DepositedSigningKeysCountChanged", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseDepositedSigningKeysCountChanged(log types.Log) (*SimpleDVTModuleDepositedSigningKeysCountChanged, error) {
	event := new(SimpleDVTModuleDepositedSigningKeysCountChanged)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "DepositedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleExitDeadlineThresholdChangedIterator is returned from FilterExitDeadlineThresholdChanged and is used to iterate over the raw logs and unpacked data for ExitDeadlineThresholdChanged events raised by the SimpleDVTModule contract.
type SimpleDVTModuleExitDeadlineThresholdChangedIterator struct {
	Event *SimpleDVTModuleExitDeadlineThresholdChanged // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleExitDeadlineThresholdChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleExitDeadlineThresholdChanged)
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
		it.Event = new(SimpleDVTModuleExitDeadlineThresholdChanged)
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
func (it *SimpleDVTModuleExitDeadlineThresholdChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleExitDeadlineThresholdChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleExitDeadlineThresholdChanged represents a ExitDeadlineThresholdChanged event raised by the SimpleDVTModule contract.
type SimpleDVTModuleExitDeadlineThresholdChanged struct {
	Threshold       *big.Int
	ReportingWindow *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterExitDeadlineThresholdChanged is a free log retrieval operation binding the contract event 0xcdd9ecf0c02860ba025eac9d0b0c3fad00cc8a9139ca9eec4076fe8ae57d43af.
//
// Solidity: event ExitDeadlineThresholdChanged(uint256 threshold, uint256 reportingWindow)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterExitDeadlineThresholdChanged(opts *bind.FilterOpts) (*SimpleDVTModuleExitDeadlineThresholdChangedIterator, error) {

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "ExitDeadlineThresholdChanged")
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleExitDeadlineThresholdChangedIterator{contract: _SimpleDVTModule.contract, event: "ExitDeadlineThresholdChanged", logs: logs, sub: sub}, nil
}

// WatchExitDeadlineThresholdChanged is a free log subscription operation binding the contract event 0xcdd9ecf0c02860ba025eac9d0b0c3fad00cc8a9139ca9eec4076fe8ae57d43af.
//
// Solidity: event ExitDeadlineThresholdChanged(uint256 threshold, uint256 reportingWindow)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchExitDeadlineThresholdChanged(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleExitDeadlineThresholdChanged) (event.Subscription, error) {

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "ExitDeadlineThresholdChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleExitDeadlineThresholdChanged)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "ExitDeadlineThresholdChanged", log); err != nil {
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

// ParseExitDeadlineThresholdChanged is a log parse operation binding the contract event 0xcdd9ecf0c02860ba025eac9d0b0c3fad00cc8a9139ca9eec4076fe8ae57d43af.
//
// Solidity: event ExitDeadlineThresholdChanged(uint256 threshold, uint256 reportingWindow)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseExitDeadlineThresholdChanged(log types.Log) (*SimpleDVTModuleExitDeadlineThresholdChanged, error) {
	event := new(SimpleDVTModuleExitDeadlineThresholdChanged)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "ExitDeadlineThresholdChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleExitedSigningKeysCountChangedIterator is returned from FilterExitedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for ExitedSigningKeysCountChanged events raised by the SimpleDVTModule contract.
type SimpleDVTModuleExitedSigningKeysCountChangedIterator struct {
	Event *SimpleDVTModuleExitedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleExitedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleExitedSigningKeysCountChanged)
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
		it.Event = new(SimpleDVTModuleExitedSigningKeysCountChanged)
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
func (it *SimpleDVTModuleExitedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleExitedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleExitedSigningKeysCountChanged represents a ExitedSigningKeysCountChanged event raised by the SimpleDVTModule contract.
type SimpleDVTModuleExitedSigningKeysCountChanged struct {
	NodeOperatorId        *big.Int
	ExitedValidatorsCount *big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterExitedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x0f67960648751434ae86bf350db61194f387fda387e7f568b0ccd0ae0c220166.
//
// Solidity: event ExitedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 exitedValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterExitedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleExitedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "ExitedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleExitedSigningKeysCountChangedIterator{contract: _SimpleDVTModule.contract, event: "ExitedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchExitedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x0f67960648751434ae86bf350db61194f387fda387e7f568b0ccd0ae0c220166.
//
// Solidity: event ExitedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 exitedValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchExitedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleExitedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "ExitedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleExitedSigningKeysCountChanged)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "ExitedSigningKeysCountChanged", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseExitedSigningKeysCountChanged(log types.Log) (*SimpleDVTModuleExitedSigningKeysCountChanged, error) {
	event := new(SimpleDVTModuleExitedSigningKeysCountChanged)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "ExitedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleKeysOpIndexSetIterator is returned from FilterKeysOpIndexSet and is used to iterate over the raw logs and unpacked data for KeysOpIndexSet events raised by the SimpleDVTModule contract.
type SimpleDVTModuleKeysOpIndexSetIterator struct {
	Event *SimpleDVTModuleKeysOpIndexSet // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleKeysOpIndexSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleKeysOpIndexSet)
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
		it.Event = new(SimpleDVTModuleKeysOpIndexSet)
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
func (it *SimpleDVTModuleKeysOpIndexSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleKeysOpIndexSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleKeysOpIndexSet represents a KeysOpIndexSet event raised by the SimpleDVTModule contract.
type SimpleDVTModuleKeysOpIndexSet struct {
	KeysOpIndex *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterKeysOpIndexSet is a free log retrieval operation binding the contract event 0xfb992daec9d46d64898e3a9336d02811349df6cbea8b95d4deb2fa6c7b454f0d.
//
// Solidity: event KeysOpIndexSet(uint256 keysOpIndex)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterKeysOpIndexSet(opts *bind.FilterOpts) (*SimpleDVTModuleKeysOpIndexSetIterator, error) {

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "KeysOpIndexSet")
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleKeysOpIndexSetIterator{contract: _SimpleDVTModule.contract, event: "KeysOpIndexSet", logs: logs, sub: sub}, nil
}

// WatchKeysOpIndexSet is a free log subscription operation binding the contract event 0xfb992daec9d46d64898e3a9336d02811349df6cbea8b95d4deb2fa6c7b454f0d.
//
// Solidity: event KeysOpIndexSet(uint256 keysOpIndex)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchKeysOpIndexSet(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleKeysOpIndexSet) (event.Subscription, error) {

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "KeysOpIndexSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleKeysOpIndexSet)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "KeysOpIndexSet", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseKeysOpIndexSet(log types.Log) (*SimpleDVTModuleKeysOpIndexSet, error) {
	event := new(SimpleDVTModuleKeysOpIndexSet)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "KeysOpIndexSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleLocatorContractSetIterator is returned from FilterLocatorContractSet and is used to iterate over the raw logs and unpacked data for LocatorContractSet events raised by the SimpleDVTModule contract.
type SimpleDVTModuleLocatorContractSetIterator struct {
	Event *SimpleDVTModuleLocatorContractSet // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleLocatorContractSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleLocatorContractSet)
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
		it.Event = new(SimpleDVTModuleLocatorContractSet)
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
func (it *SimpleDVTModuleLocatorContractSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleLocatorContractSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleLocatorContractSet represents a LocatorContractSet event raised by the SimpleDVTModule contract.
type SimpleDVTModuleLocatorContractSet struct {
	LocatorAddress common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterLocatorContractSet is a free log retrieval operation binding the contract event 0xa44aa4b7320163340e971b1f22f153bbb8a0151d783bd58377018ea5bc96d0c9.
//
// Solidity: event LocatorContractSet(address locatorAddress)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterLocatorContractSet(opts *bind.FilterOpts) (*SimpleDVTModuleLocatorContractSetIterator, error) {

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "LocatorContractSet")
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleLocatorContractSetIterator{contract: _SimpleDVTModule.contract, event: "LocatorContractSet", logs: logs, sub: sub}, nil
}

// WatchLocatorContractSet is a free log subscription operation binding the contract event 0xa44aa4b7320163340e971b1f22f153bbb8a0151d783bd58377018ea5bc96d0c9.
//
// Solidity: event LocatorContractSet(address locatorAddress)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchLocatorContractSet(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleLocatorContractSet) (event.Subscription, error) {

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "LocatorContractSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleLocatorContractSet)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "LocatorContractSet", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseLocatorContractSet(log types.Log) (*SimpleDVTModuleLocatorContractSet, error) {
	event := new(SimpleDVTModuleLocatorContractSet)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "LocatorContractSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleNodeOperatorActiveSetIterator is returned from FilterNodeOperatorActiveSet and is used to iterate over the raw logs and unpacked data for NodeOperatorActiveSet events raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorActiveSetIterator struct {
	Event *SimpleDVTModuleNodeOperatorActiveSet // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleNodeOperatorActiveSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleNodeOperatorActiveSet)
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
		it.Event = new(SimpleDVTModuleNodeOperatorActiveSet)
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
func (it *SimpleDVTModuleNodeOperatorActiveSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleNodeOperatorActiveSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleNodeOperatorActiveSet represents a NodeOperatorActiveSet event raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorActiveSet struct {
	NodeOperatorId *big.Int
	Active         bool
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorActiveSet is a free log retrieval operation binding the contract event 0xecdf08e8a6c4493efb460f6abc7d14532074fa339c3a6410623a1d3ee0fb2cac.
//
// Solidity: event NodeOperatorActiveSet(uint256 indexed nodeOperatorId, bool active)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterNodeOperatorActiveSet(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleNodeOperatorActiveSetIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "NodeOperatorActiveSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleNodeOperatorActiveSetIterator{contract: _SimpleDVTModule.contract, event: "NodeOperatorActiveSet", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorActiveSet is a free log subscription operation binding the contract event 0xecdf08e8a6c4493efb460f6abc7d14532074fa339c3a6410623a1d3ee0fb2cac.
//
// Solidity: event NodeOperatorActiveSet(uint256 indexed nodeOperatorId, bool active)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchNodeOperatorActiveSet(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleNodeOperatorActiveSet, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "NodeOperatorActiveSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleNodeOperatorActiveSet)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorActiveSet", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseNodeOperatorActiveSet(log types.Log) (*SimpleDVTModuleNodeOperatorActiveSet, error) {
	event := new(SimpleDVTModuleNodeOperatorActiveSet)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorActiveSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleNodeOperatorAddedIterator is returned from FilterNodeOperatorAdded and is used to iterate over the raw logs and unpacked data for NodeOperatorAdded events raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorAddedIterator struct {
	Event *SimpleDVTModuleNodeOperatorAdded // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleNodeOperatorAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleNodeOperatorAdded)
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
		it.Event = new(SimpleDVTModuleNodeOperatorAdded)
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
func (it *SimpleDVTModuleNodeOperatorAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleNodeOperatorAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleNodeOperatorAdded represents a NodeOperatorAdded event raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorAdded struct {
	NodeOperatorId *big.Int
	Name           string
	RewardAddress  common.Address
	StakingLimit   uint64
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorAdded is a free log retrieval operation binding the contract event 0xc52ec0ad7872dae440d886040390c13677df7bf3cca136d8d81e5e5e7dd62ff1.
//
// Solidity: event NodeOperatorAdded(uint256 nodeOperatorId, string name, address rewardAddress, uint64 stakingLimit)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterNodeOperatorAdded(opts *bind.FilterOpts) (*SimpleDVTModuleNodeOperatorAddedIterator, error) {

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "NodeOperatorAdded")
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleNodeOperatorAddedIterator{contract: _SimpleDVTModule.contract, event: "NodeOperatorAdded", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorAdded is a free log subscription operation binding the contract event 0xc52ec0ad7872dae440d886040390c13677df7bf3cca136d8d81e5e5e7dd62ff1.
//
// Solidity: event NodeOperatorAdded(uint256 nodeOperatorId, string name, address rewardAddress, uint64 stakingLimit)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchNodeOperatorAdded(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleNodeOperatorAdded) (event.Subscription, error) {

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "NodeOperatorAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleNodeOperatorAdded)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorAdded", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseNodeOperatorAdded(log types.Log) (*SimpleDVTModuleNodeOperatorAdded, error) {
	event := new(SimpleDVTModuleNodeOperatorAdded)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleNodeOperatorNameSetIterator is returned from FilterNodeOperatorNameSet and is used to iterate over the raw logs and unpacked data for NodeOperatorNameSet events raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorNameSetIterator struct {
	Event *SimpleDVTModuleNodeOperatorNameSet // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleNodeOperatorNameSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleNodeOperatorNameSet)
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
		it.Event = new(SimpleDVTModuleNodeOperatorNameSet)
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
func (it *SimpleDVTModuleNodeOperatorNameSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleNodeOperatorNameSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleNodeOperatorNameSet represents a NodeOperatorNameSet event raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorNameSet struct {
	NodeOperatorId *big.Int
	Name           string
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorNameSet is a free log retrieval operation binding the contract event 0xcb16868f4831cc58a28d413f658752a2958bd1f50e94ed6391716b936c48093b.
//
// Solidity: event NodeOperatorNameSet(uint256 indexed nodeOperatorId, string name)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterNodeOperatorNameSet(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleNodeOperatorNameSetIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "NodeOperatorNameSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleNodeOperatorNameSetIterator{contract: _SimpleDVTModule.contract, event: "NodeOperatorNameSet", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorNameSet is a free log subscription operation binding the contract event 0xcb16868f4831cc58a28d413f658752a2958bd1f50e94ed6391716b936c48093b.
//
// Solidity: event NodeOperatorNameSet(uint256 indexed nodeOperatorId, string name)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchNodeOperatorNameSet(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleNodeOperatorNameSet, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "NodeOperatorNameSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleNodeOperatorNameSet)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorNameSet", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseNodeOperatorNameSet(log types.Log) (*SimpleDVTModuleNodeOperatorNameSet, error) {
	event := new(SimpleDVTModuleNodeOperatorNameSet)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorNameSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleNodeOperatorRewardAddressSetIterator is returned from FilterNodeOperatorRewardAddressSet and is used to iterate over the raw logs and unpacked data for NodeOperatorRewardAddressSet events raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorRewardAddressSetIterator struct {
	Event *SimpleDVTModuleNodeOperatorRewardAddressSet // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleNodeOperatorRewardAddressSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleNodeOperatorRewardAddressSet)
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
		it.Event = new(SimpleDVTModuleNodeOperatorRewardAddressSet)
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
func (it *SimpleDVTModuleNodeOperatorRewardAddressSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleNodeOperatorRewardAddressSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleNodeOperatorRewardAddressSet represents a NodeOperatorRewardAddressSet event raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorRewardAddressSet struct {
	NodeOperatorId *big.Int
	RewardAddress  common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorRewardAddressSet is a free log retrieval operation binding the contract event 0x9a52205165d510fc1e428886d52108725dc01ed544da1702dc7bd3fdb3f243b2.
//
// Solidity: event NodeOperatorRewardAddressSet(uint256 indexed nodeOperatorId, address rewardAddress)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterNodeOperatorRewardAddressSet(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleNodeOperatorRewardAddressSetIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "NodeOperatorRewardAddressSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleNodeOperatorRewardAddressSetIterator{contract: _SimpleDVTModule.contract, event: "NodeOperatorRewardAddressSet", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorRewardAddressSet is a free log subscription operation binding the contract event 0x9a52205165d510fc1e428886d52108725dc01ed544da1702dc7bd3fdb3f243b2.
//
// Solidity: event NodeOperatorRewardAddressSet(uint256 indexed nodeOperatorId, address rewardAddress)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchNodeOperatorRewardAddressSet(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleNodeOperatorRewardAddressSet, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "NodeOperatorRewardAddressSet", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleNodeOperatorRewardAddressSet)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorRewardAddressSet", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseNodeOperatorRewardAddressSet(log types.Log) (*SimpleDVTModuleNodeOperatorRewardAddressSet, error) {
	event := new(SimpleDVTModuleNodeOperatorRewardAddressSet)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorRewardAddressSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleNodeOperatorTotalKeysTrimmedIterator is returned from FilterNodeOperatorTotalKeysTrimmed and is used to iterate over the raw logs and unpacked data for NodeOperatorTotalKeysTrimmed events raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorTotalKeysTrimmedIterator struct {
	Event *SimpleDVTModuleNodeOperatorTotalKeysTrimmed // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleNodeOperatorTotalKeysTrimmedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleNodeOperatorTotalKeysTrimmed)
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
		it.Event = new(SimpleDVTModuleNodeOperatorTotalKeysTrimmed)
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
func (it *SimpleDVTModuleNodeOperatorTotalKeysTrimmedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleNodeOperatorTotalKeysTrimmedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleNodeOperatorTotalKeysTrimmed represents a NodeOperatorTotalKeysTrimmed event raised by the SimpleDVTModule contract.
type SimpleDVTModuleNodeOperatorTotalKeysTrimmed struct {
	NodeOperatorId   *big.Int
	TotalKeysTrimmed uint64
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorTotalKeysTrimmed is a free log retrieval operation binding the contract event 0x9824694569ba758f8872bb150515caaf8f1e2cc27e6805679c4ac8c3b9b83d87.
//
// Solidity: event NodeOperatorTotalKeysTrimmed(uint256 indexed nodeOperatorId, uint64 totalKeysTrimmed)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterNodeOperatorTotalKeysTrimmed(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleNodeOperatorTotalKeysTrimmedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "NodeOperatorTotalKeysTrimmed", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleNodeOperatorTotalKeysTrimmedIterator{contract: _SimpleDVTModule.contract, event: "NodeOperatorTotalKeysTrimmed", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorTotalKeysTrimmed is a free log subscription operation binding the contract event 0x9824694569ba758f8872bb150515caaf8f1e2cc27e6805679c4ac8c3b9b83d87.
//
// Solidity: event NodeOperatorTotalKeysTrimmed(uint256 indexed nodeOperatorId, uint64 totalKeysTrimmed)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchNodeOperatorTotalKeysTrimmed(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleNodeOperatorTotalKeysTrimmed, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "NodeOperatorTotalKeysTrimmed", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleNodeOperatorTotalKeysTrimmed)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorTotalKeysTrimmed", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseNodeOperatorTotalKeysTrimmed(log types.Log) (*SimpleDVTModuleNodeOperatorTotalKeysTrimmed, error) {
	event := new(SimpleDVTModuleNodeOperatorTotalKeysTrimmed)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "NodeOperatorTotalKeysTrimmed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleNonceChangedIterator is returned from FilterNonceChanged and is used to iterate over the raw logs and unpacked data for NonceChanged events raised by the SimpleDVTModule contract.
type SimpleDVTModuleNonceChangedIterator struct {
	Event *SimpleDVTModuleNonceChanged // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleNonceChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleNonceChanged)
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
		it.Event = new(SimpleDVTModuleNonceChanged)
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
func (it *SimpleDVTModuleNonceChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleNonceChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleNonceChanged represents a NonceChanged event raised by the SimpleDVTModule contract.
type SimpleDVTModuleNonceChanged struct {
	Nonce *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNonceChanged is a free log retrieval operation binding the contract event 0x7220970e1f1f12864ecccd8942690a837c7a8dd45d158cb891eb45a8a69134aa.
//
// Solidity: event NonceChanged(uint256 nonce)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterNonceChanged(opts *bind.FilterOpts) (*SimpleDVTModuleNonceChangedIterator, error) {

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "NonceChanged")
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleNonceChangedIterator{contract: _SimpleDVTModule.contract, event: "NonceChanged", logs: logs, sub: sub}, nil
}

// WatchNonceChanged is a free log subscription operation binding the contract event 0x7220970e1f1f12864ecccd8942690a837c7a8dd45d158cb891eb45a8a69134aa.
//
// Solidity: event NonceChanged(uint256 nonce)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchNonceChanged(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleNonceChanged) (event.Subscription, error) {

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "NonceChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleNonceChanged)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "NonceChanged", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseNonceChanged(log types.Log) (*SimpleDVTModuleNonceChanged, error) {
	event := new(SimpleDVTModuleNonceChanged)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "NonceChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleRecoverToVaultIterator is returned from FilterRecoverToVault and is used to iterate over the raw logs and unpacked data for RecoverToVault events raised by the SimpleDVTModule contract.
type SimpleDVTModuleRecoverToVaultIterator struct {
	Event *SimpleDVTModuleRecoverToVault // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleRecoverToVaultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleRecoverToVault)
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
		it.Event = new(SimpleDVTModuleRecoverToVault)
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
func (it *SimpleDVTModuleRecoverToVaultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleRecoverToVaultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleRecoverToVault represents a RecoverToVault event raised by the SimpleDVTModule contract.
type SimpleDVTModuleRecoverToVault struct {
	Vault  common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRecoverToVault is a free log retrieval operation binding the contract event 0x596caf56044b55fb8c4ca640089bbc2b63cae3e978b851f5745cbb7c5b288e02.
//
// Solidity: event RecoverToVault(address indexed vault, address indexed token, uint256 amount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterRecoverToVault(opts *bind.FilterOpts, vault []common.Address, token []common.Address) (*SimpleDVTModuleRecoverToVaultIterator, error) {

	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "RecoverToVault", vaultRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleRecoverToVaultIterator{contract: _SimpleDVTModule.contract, event: "RecoverToVault", logs: logs, sub: sub}, nil
}

// WatchRecoverToVault is a free log subscription operation binding the contract event 0x596caf56044b55fb8c4ca640089bbc2b63cae3e978b851f5745cbb7c5b288e02.
//
// Solidity: event RecoverToVault(address indexed vault, address indexed token, uint256 amount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchRecoverToVault(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleRecoverToVault, vault []common.Address, token []common.Address) (event.Subscription, error) {

	var vaultRule []interface{}
	for _, vaultItem := range vault {
		vaultRule = append(vaultRule, vaultItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "RecoverToVault", vaultRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleRecoverToVault)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "RecoverToVault", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseRecoverToVault(log types.Log) (*SimpleDVTModuleRecoverToVault, error) {
	event := new(SimpleDVTModuleRecoverToVault)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "RecoverToVault", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleRewardDistributionStateChangedIterator is returned from FilterRewardDistributionStateChanged and is used to iterate over the raw logs and unpacked data for RewardDistributionStateChanged events raised by the SimpleDVTModule contract.
type SimpleDVTModuleRewardDistributionStateChangedIterator struct {
	Event *SimpleDVTModuleRewardDistributionStateChanged // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleRewardDistributionStateChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleRewardDistributionStateChanged)
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
		it.Event = new(SimpleDVTModuleRewardDistributionStateChanged)
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
func (it *SimpleDVTModuleRewardDistributionStateChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleRewardDistributionStateChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleRewardDistributionStateChanged represents a RewardDistributionStateChanged event raised by the SimpleDVTModule contract.
type SimpleDVTModuleRewardDistributionStateChanged struct {
	State uint8
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRewardDistributionStateChanged is a free log retrieval operation binding the contract event 0x7545d380f29a8ae65fafb1acdf2c7762ec02d5607fecbea9dd8d8245e1616d93.
//
// Solidity: event RewardDistributionStateChanged(uint8 state)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterRewardDistributionStateChanged(opts *bind.FilterOpts) (*SimpleDVTModuleRewardDistributionStateChangedIterator, error) {

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "RewardDistributionStateChanged")
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleRewardDistributionStateChangedIterator{contract: _SimpleDVTModule.contract, event: "RewardDistributionStateChanged", logs: logs, sub: sub}, nil
}

// WatchRewardDistributionStateChanged is a free log subscription operation binding the contract event 0x7545d380f29a8ae65fafb1acdf2c7762ec02d5607fecbea9dd8d8245e1616d93.
//
// Solidity: event RewardDistributionStateChanged(uint8 state)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchRewardDistributionStateChanged(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleRewardDistributionStateChanged) (event.Subscription, error) {

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "RewardDistributionStateChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleRewardDistributionStateChanged)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "RewardDistributionStateChanged", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseRewardDistributionStateChanged(log types.Log) (*SimpleDVTModuleRewardDistributionStateChanged, error) {
	event := new(SimpleDVTModuleRewardDistributionStateChanged)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "RewardDistributionStateChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleRewardsDistributedIterator is returned from FilterRewardsDistributed and is used to iterate over the raw logs and unpacked data for RewardsDistributed events raised by the SimpleDVTModule contract.
type SimpleDVTModuleRewardsDistributedIterator struct {
	Event *SimpleDVTModuleRewardsDistributed // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleRewardsDistributedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleRewardsDistributed)
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
		it.Event = new(SimpleDVTModuleRewardsDistributed)
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
func (it *SimpleDVTModuleRewardsDistributedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleRewardsDistributedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleRewardsDistributed represents a RewardsDistributed event raised by the SimpleDVTModule contract.
type SimpleDVTModuleRewardsDistributed struct {
	RewardAddress common.Address
	SharesAmount  *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRewardsDistributed is a free log retrieval operation binding the contract event 0xdf29796aad820e4bb192f3a8d631b76519bcd2cbe77cc85af20e9df53cece086.
//
// Solidity: event RewardsDistributed(address indexed rewardAddress, uint256 sharesAmount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterRewardsDistributed(opts *bind.FilterOpts, rewardAddress []common.Address) (*SimpleDVTModuleRewardsDistributedIterator, error) {

	var rewardAddressRule []interface{}
	for _, rewardAddressItem := range rewardAddress {
		rewardAddressRule = append(rewardAddressRule, rewardAddressItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "RewardsDistributed", rewardAddressRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleRewardsDistributedIterator{contract: _SimpleDVTModule.contract, event: "RewardsDistributed", logs: logs, sub: sub}, nil
}

// WatchRewardsDistributed is a free log subscription operation binding the contract event 0xdf29796aad820e4bb192f3a8d631b76519bcd2cbe77cc85af20e9df53cece086.
//
// Solidity: event RewardsDistributed(address indexed rewardAddress, uint256 sharesAmount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchRewardsDistributed(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleRewardsDistributed, rewardAddress []common.Address) (event.Subscription, error) {

	var rewardAddressRule []interface{}
	for _, rewardAddressItem := range rewardAddress {
		rewardAddressRule = append(rewardAddressRule, rewardAddressItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "RewardsDistributed", rewardAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleRewardsDistributed)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "RewardsDistributed", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseRewardsDistributed(log types.Log) (*SimpleDVTModuleRewardsDistributed, error) {
	event := new(SimpleDVTModuleRewardsDistributed)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "RewardsDistributed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleScriptResultIterator is returned from FilterScriptResult and is used to iterate over the raw logs and unpacked data for ScriptResult events raised by the SimpleDVTModule contract.
type SimpleDVTModuleScriptResultIterator struct {
	Event *SimpleDVTModuleScriptResult // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleScriptResultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleScriptResult)
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
		it.Event = new(SimpleDVTModuleScriptResult)
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
func (it *SimpleDVTModuleScriptResultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleScriptResultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleScriptResult represents a ScriptResult event raised by the SimpleDVTModule contract.
type SimpleDVTModuleScriptResult struct {
	Executor   common.Address
	Script     []byte
	Input      []byte
	ReturnData []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterScriptResult is a free log retrieval operation binding the contract event 0x5229a5dba83a54ae8cb5b51bdd6de9474cacbe9dd332f5185f3a4f4f2e3f4ad9.
//
// Solidity: event ScriptResult(address indexed executor, bytes script, bytes input, bytes returnData)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterScriptResult(opts *bind.FilterOpts, executor []common.Address) (*SimpleDVTModuleScriptResultIterator, error) {

	var executorRule []interface{}
	for _, executorItem := range executor {
		executorRule = append(executorRule, executorItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "ScriptResult", executorRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleScriptResultIterator{contract: _SimpleDVTModule.contract, event: "ScriptResult", logs: logs, sub: sub}, nil
}

// WatchScriptResult is a free log subscription operation binding the contract event 0x5229a5dba83a54ae8cb5b51bdd6de9474cacbe9dd332f5185f3a4f4f2e3f4ad9.
//
// Solidity: event ScriptResult(address indexed executor, bytes script, bytes input, bytes returnData)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchScriptResult(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleScriptResult, executor []common.Address) (event.Subscription, error) {

	var executorRule []interface{}
	for _, executorItem := range executor {
		executorRule = append(executorRule, executorItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "ScriptResult", executorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleScriptResult)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "ScriptResult", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseScriptResult(log types.Log) (*SimpleDVTModuleScriptResult, error) {
	event := new(SimpleDVTModuleScriptResult)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "ScriptResult", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleStakingModuleTypeSetIterator is returned from FilterStakingModuleTypeSet and is used to iterate over the raw logs and unpacked data for StakingModuleTypeSet events raised by the SimpleDVTModule contract.
type SimpleDVTModuleStakingModuleTypeSetIterator struct {
	Event *SimpleDVTModuleStakingModuleTypeSet // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleStakingModuleTypeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleStakingModuleTypeSet)
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
		it.Event = new(SimpleDVTModuleStakingModuleTypeSet)
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
func (it *SimpleDVTModuleStakingModuleTypeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleStakingModuleTypeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleStakingModuleTypeSet represents a StakingModuleTypeSet event raised by the SimpleDVTModule contract.
type SimpleDVTModuleStakingModuleTypeSet struct {
	ModuleType [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStakingModuleTypeSet is a free log retrieval operation binding the contract event 0xdb042010b15d1321c99552200b350bba0a95dfa3d0b43869983ce74b44d644ee.
//
// Solidity: event StakingModuleTypeSet(bytes32 moduleType)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterStakingModuleTypeSet(opts *bind.FilterOpts) (*SimpleDVTModuleStakingModuleTypeSetIterator, error) {

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "StakingModuleTypeSet")
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleStakingModuleTypeSetIterator{contract: _SimpleDVTModule.contract, event: "StakingModuleTypeSet", logs: logs, sub: sub}, nil
}

// WatchStakingModuleTypeSet is a free log subscription operation binding the contract event 0xdb042010b15d1321c99552200b350bba0a95dfa3d0b43869983ce74b44d644ee.
//
// Solidity: event StakingModuleTypeSet(bytes32 moduleType)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchStakingModuleTypeSet(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleStakingModuleTypeSet) (event.Subscription, error) {

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "StakingModuleTypeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleStakingModuleTypeSet)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "StakingModuleTypeSet", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseStakingModuleTypeSet(log types.Log) (*SimpleDVTModuleStakingModuleTypeSet, error) {
	event := new(SimpleDVTModuleStakingModuleTypeSet)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "StakingModuleTypeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleTargetValidatorsCountChangedIterator is returned from FilterTargetValidatorsCountChanged and is used to iterate over the raw logs and unpacked data for TargetValidatorsCountChanged events raised by the SimpleDVTModule contract.
type SimpleDVTModuleTargetValidatorsCountChangedIterator struct {
	Event *SimpleDVTModuleTargetValidatorsCountChanged // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleTargetValidatorsCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleTargetValidatorsCountChanged)
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
		it.Event = new(SimpleDVTModuleTargetValidatorsCountChanged)
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
func (it *SimpleDVTModuleTargetValidatorsCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleTargetValidatorsCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleTargetValidatorsCountChanged represents a TargetValidatorsCountChanged event raised by the SimpleDVTModule contract.
type SimpleDVTModuleTargetValidatorsCountChanged struct {
	NodeOperatorId        *big.Int
	TargetValidatorsCount *big.Int
	TargetLimitMode       *big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterTargetValidatorsCountChanged is a free log retrieval operation binding the contract event 0xf92eb109ce5b449e9b121c352c6aeb4319538a90738cb95d84f08e41274e92d2.
//
// Solidity: event TargetValidatorsCountChanged(uint256 indexed nodeOperatorId, uint256 targetValidatorsCount, uint256 targetLimitMode)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterTargetValidatorsCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleTargetValidatorsCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "TargetValidatorsCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleTargetValidatorsCountChangedIterator{contract: _SimpleDVTModule.contract, event: "TargetValidatorsCountChanged", logs: logs, sub: sub}, nil
}

// WatchTargetValidatorsCountChanged is a free log subscription operation binding the contract event 0xf92eb109ce5b449e9b121c352c6aeb4319538a90738cb95d84f08e41274e92d2.
//
// Solidity: event TargetValidatorsCountChanged(uint256 indexed nodeOperatorId, uint256 targetValidatorsCount, uint256 targetLimitMode)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchTargetValidatorsCountChanged(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleTargetValidatorsCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "TargetValidatorsCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleTargetValidatorsCountChanged)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "TargetValidatorsCountChanged", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseTargetValidatorsCountChanged(log types.Log) (*SimpleDVTModuleTargetValidatorsCountChanged, error) {
	event := new(SimpleDVTModuleTargetValidatorsCountChanged)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "TargetValidatorsCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleTotalSigningKeysCountChangedIterator is returned from FilterTotalSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for TotalSigningKeysCountChanged events raised by the SimpleDVTModule contract.
type SimpleDVTModuleTotalSigningKeysCountChangedIterator struct {
	Event *SimpleDVTModuleTotalSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleTotalSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleTotalSigningKeysCountChanged)
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
		it.Event = new(SimpleDVTModuleTotalSigningKeysCountChanged)
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
func (it *SimpleDVTModuleTotalSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleTotalSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleTotalSigningKeysCountChanged represents a TotalSigningKeysCountChanged event raised by the SimpleDVTModule contract.
type SimpleDVTModuleTotalSigningKeysCountChanged struct {
	NodeOperatorId       *big.Int
	TotalValidatorsCount *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterTotalSigningKeysCountChanged is a free log retrieval operation binding the contract event 0xdd01838a366ae4dc9a86e1922512c0716abebc9a440baae0e22d2dec578223f0.
//
// Solidity: event TotalSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 totalValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterTotalSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleTotalSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "TotalSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleTotalSigningKeysCountChangedIterator{contract: _SimpleDVTModule.contract, event: "TotalSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchTotalSigningKeysCountChanged is a free log subscription operation binding the contract event 0xdd01838a366ae4dc9a86e1922512c0716abebc9a440baae0e22d2dec578223f0.
//
// Solidity: event TotalSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 totalValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchTotalSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleTotalSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "TotalSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleTotalSigningKeysCountChanged)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "TotalSigningKeysCountChanged", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseTotalSigningKeysCountChanged(log types.Log) (*SimpleDVTModuleTotalSigningKeysCountChanged, error) {
	event := new(SimpleDVTModuleTotalSigningKeysCountChanged)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "TotalSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleValidatorExitStatusUpdatedIterator is returned from FilterValidatorExitStatusUpdated and is used to iterate over the raw logs and unpacked data for ValidatorExitStatusUpdated events raised by the SimpleDVTModule contract.
type SimpleDVTModuleValidatorExitStatusUpdatedIterator struct {
	Event *SimpleDVTModuleValidatorExitStatusUpdated // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleValidatorExitStatusUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleValidatorExitStatusUpdated)
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
		it.Event = new(SimpleDVTModuleValidatorExitStatusUpdated)
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
func (it *SimpleDVTModuleValidatorExitStatusUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleValidatorExitStatusUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleValidatorExitStatusUpdated represents a ValidatorExitStatusUpdated event raised by the SimpleDVTModule contract.
type SimpleDVTModuleValidatorExitStatusUpdated struct {
	NodeOperatorId      *big.Int
	PublicKey           []byte
	EligibleToExitInSec *big.Int
	ProofSlotTimestamp  *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterValidatorExitStatusUpdated is a free log retrieval operation binding the contract event 0x7f781065728a5d3f9c1ee91ef0c0b8a9455ebc228dae6d35514b32c7370349c9.
//
// Solidity: event ValidatorExitStatusUpdated(uint256 indexed nodeOperatorId, bytes publicKey, uint256 eligibleToExitInSec, uint256 proofSlotTimestamp)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterValidatorExitStatusUpdated(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleValidatorExitStatusUpdatedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "ValidatorExitStatusUpdated", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleValidatorExitStatusUpdatedIterator{contract: _SimpleDVTModule.contract, event: "ValidatorExitStatusUpdated", logs: logs, sub: sub}, nil
}

// WatchValidatorExitStatusUpdated is a free log subscription operation binding the contract event 0x7f781065728a5d3f9c1ee91ef0c0b8a9455ebc228dae6d35514b32c7370349c9.
//
// Solidity: event ValidatorExitStatusUpdated(uint256 indexed nodeOperatorId, bytes publicKey, uint256 eligibleToExitInSec, uint256 proofSlotTimestamp)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchValidatorExitStatusUpdated(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleValidatorExitStatusUpdated, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "ValidatorExitStatusUpdated", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleValidatorExitStatusUpdated)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "ValidatorExitStatusUpdated", log); err != nil {
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

// ParseValidatorExitStatusUpdated is a log parse operation binding the contract event 0x7f781065728a5d3f9c1ee91ef0c0b8a9455ebc228dae6d35514b32c7370349c9.
//
// Solidity: event ValidatorExitStatusUpdated(uint256 indexed nodeOperatorId, bytes publicKey, uint256 eligibleToExitInSec, uint256 proofSlotTimestamp)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseValidatorExitStatusUpdated(log types.Log) (*SimpleDVTModuleValidatorExitStatusUpdated, error) {
	event := new(SimpleDVTModuleValidatorExitStatusUpdated)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "ValidatorExitStatusUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleValidatorExitTriggeredIterator is returned from FilterValidatorExitTriggered and is used to iterate over the raw logs and unpacked data for ValidatorExitTriggered events raised by the SimpleDVTModule contract.
type SimpleDVTModuleValidatorExitTriggeredIterator struct {
	Event *SimpleDVTModuleValidatorExitTriggered // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleValidatorExitTriggeredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleValidatorExitTriggered)
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
		it.Event = new(SimpleDVTModuleValidatorExitTriggered)
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
func (it *SimpleDVTModuleValidatorExitTriggeredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleValidatorExitTriggeredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleValidatorExitTriggered represents a ValidatorExitTriggered event raised by the SimpleDVTModule contract.
type SimpleDVTModuleValidatorExitTriggered struct {
	NodeOperatorId           *big.Int
	PublicKey                []byte
	WithdrawalRequestPaidFee *big.Int
	ExitType                 *big.Int
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterValidatorExitTriggered is a free log retrieval operation binding the contract event 0xbc3a9275fc3a367bea48e061d624141756b6b28245cee6c39703d9289aa90f62.
//
// Solidity: event ValidatorExitTriggered(uint256 indexed nodeOperatorId, bytes publicKey, uint256 withdrawalRequestPaidFee, uint256 exitType)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterValidatorExitTriggered(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleValidatorExitTriggeredIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "ValidatorExitTriggered", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleValidatorExitTriggeredIterator{contract: _SimpleDVTModule.contract, event: "ValidatorExitTriggered", logs: logs, sub: sub}, nil
}

// WatchValidatorExitTriggered is a free log subscription operation binding the contract event 0xbc3a9275fc3a367bea48e061d624141756b6b28245cee6c39703d9289aa90f62.
//
// Solidity: event ValidatorExitTriggered(uint256 indexed nodeOperatorId, bytes publicKey, uint256 withdrawalRequestPaidFee, uint256 exitType)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchValidatorExitTriggered(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleValidatorExitTriggered, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "ValidatorExitTriggered", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleValidatorExitTriggered)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "ValidatorExitTriggered", log); err != nil {
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

// ParseValidatorExitTriggered is a log parse operation binding the contract event 0xbc3a9275fc3a367bea48e061d624141756b6b28245cee6c39703d9289aa90f62.
//
// Solidity: event ValidatorExitTriggered(uint256 indexed nodeOperatorId, bytes publicKey, uint256 withdrawalRequestPaidFee, uint256 exitType)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseValidatorExitTriggered(log types.Log) (*SimpleDVTModuleValidatorExitTriggered, error) {
	event := new(SimpleDVTModuleValidatorExitTriggered)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "ValidatorExitTriggered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleDVTModuleVettedSigningKeysCountChangedIterator is returned from FilterVettedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for VettedSigningKeysCountChanged events raised by the SimpleDVTModule contract.
type SimpleDVTModuleVettedSigningKeysCountChangedIterator struct {
	Event *SimpleDVTModuleVettedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *SimpleDVTModuleVettedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleDVTModuleVettedSigningKeysCountChanged)
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
		it.Event = new(SimpleDVTModuleVettedSigningKeysCountChanged)
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
func (it *SimpleDVTModuleVettedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleDVTModuleVettedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleDVTModuleVettedSigningKeysCountChanged represents a VettedSigningKeysCountChanged event raised by the SimpleDVTModule contract.
type SimpleDVTModuleVettedSigningKeysCountChanged struct {
	NodeOperatorId          *big.Int
	ApprovedValidatorsCount *big.Int
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterVettedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x947f955eec7e1f626bee3afd2aa47b5de04ddcdd3fe78dc8838213015ef58dfd.
//
// Solidity: event VettedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 approvedValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) FilterVettedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*SimpleDVTModuleVettedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.FilterLogs(opts, "VettedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &SimpleDVTModuleVettedSigningKeysCountChangedIterator{contract: _SimpleDVTModule.contract, event: "VettedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchVettedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x947f955eec7e1f626bee3afd2aa47b5de04ddcdd3fe78dc8838213015ef58dfd.
//
// Solidity: event VettedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 approvedValidatorsCount)
func (_SimpleDVTModule *SimpleDVTModuleFilterer) WatchVettedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *SimpleDVTModuleVettedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _SimpleDVTModule.contract.WatchLogs(opts, "VettedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleDVTModuleVettedSigningKeysCountChanged)
				if err := _SimpleDVTModule.contract.UnpackLog(event, "VettedSigningKeysCountChanged", log); err != nil {
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
func (_SimpleDVTModule *SimpleDVTModuleFilterer) ParseVettedSigningKeysCountChanged(log types.Log) (*SimpleDVTModuleVettedSigningKeysCountChanged, error) {
	event := new(SimpleDVTModuleVettedSigningKeysCountChanged)
	if err := _SimpleDVTModule.contract.UnpackLog(event, "VettedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
