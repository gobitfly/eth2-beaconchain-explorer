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

// ICSAccountingPermitInput is an auto generated low-level Go binding around an user-defined struct.
type ICSAccountingPermitInput struct {
	Value    *big.Int
	Deadline *big.Int
	V        uint8
	R        [32]byte
	S        [32]byte
}

// NodeOperator is an auto generated low-level Go binding around an user-defined struct.
type NodeOperator struct {
	TotalAddedKeys             uint32
	TotalWithdrawnKeys         uint32
	TotalDepositedKeys         uint32
	TotalVettedKeys            uint32
	StuckValidatorsCount       uint32
	DepositableValidatorsCount uint32
	TargetLimit                uint32
	TargetLimitMode            uint8
	TotalExitedKeys            uint32
	EnqueuedCount              uint32
	ManagerAddress             common.Address
	ProposedManagerAddress     common.Address
	RewardAddress              common.Address
	ProposedRewardAddress      common.Address
	ExtendedManagerPermissions bool
}

// NodeOperatorManagementProperties is an auto generated low-level Go binding around an user-defined struct.
type NodeOperatorManagementProperties struct {
	ManagerAddress             common.Address
	RewardAddress              common.Address
	ExtendedManagerPermissions bool
}

// CSMModuleMetaData contains all meta data concerning the CSMModule contract.
var CSMModuleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"moduleType\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"minSlashingPenaltyQuotient\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"elRewardsStealingFine\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxKeysPerOperatorEA\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxKeyRemovalCharge\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"lidoLocator\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"AccessControlBadConfirmation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"neededRole\",\"type\":\"bytes32\"}],\"name\":\"AccessControlUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AlreadyActivated\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AlreadyProposed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AlreadySubmitted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EmptyKey\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ExitedKeysDecrease\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ExitedKeysHigherThanTotalDeposited\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedToSendEther\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInput\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidKeysCount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidLength\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidReportData\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidVetKeysPointer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MaxSigningKeysCountExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MethodCallIsNotAllowed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NodeOperatorDoesNotExist\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAllowedToJoinYet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAllowedToRecover\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughKeys\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSupported\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PauseUntilMustBeInFuture\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PausedExpected\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"QueueIsEmpty\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"QueueLookupNoLimit\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ResumedExpected\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SameAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderIsNotEligible\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderIsNotManagerAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderIsNotProposedAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderIsNotRewardAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SigningKeysInvalidOffset\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"StuckKeysHigherThanNonExited\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAccountingAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAdminAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroLocatorAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroPauseDuration\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroRewardAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"count\",\"type\":\"uint256\"}],\"name\":\"BatchEnqueued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"depositableKeysCount\",\"type\":\"uint256\"}],\"name\":\"DepositableSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"depositedKeysCount\",\"type\":\"uint256\"}],\"name\":\"DepositedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ELRewardsStealingPenaltyCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ELRewardsStealingPenaltyCompensated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"proposedBlockHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"stolenAmount\",\"type\":\"uint256\"}],\"name\":\"ELRewardsStealingPenaltyReported\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"ELRewardsStealingPenaltySettled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ERC1155Recovered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ERC20Recovered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"ERC721Recovered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"EtherRecovered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"exitedKeysCount\",\"type\":\"uint256\"}],\"name\":\"ExitedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"keyIndex\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"InitialSlashingSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"KeyRemovalChargeApplied\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"KeyRemovalChargeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"managerAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"rewardAddress\",\"type\":\"address\"}],\"name\":\"NodeOperatorAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldProposedAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newProposedAddress\",\"type\":\"address\"}],\"name\":\"NodeOperatorManagerAddressChangeProposed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"NodeOperatorManagerAddressChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldProposedAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newProposedAddress\",\"type\":\"address\"}],\"name\":\"NodeOperatorRewardAddressChangeProposed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"NodeOperatorRewardAddressChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"NonceChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"PublicRelease\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"referrer\",\"type\":\"address\"}],\"name\":\"ReferrerSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"Resumed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"SigningKeyAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"SigningKeyRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"shares\",\"type\":\"uint256\"}],\"name\":\"StETHSharesRecovered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"stuckKeysCount\",\"type\":\"uint256\"}],\"name\":\"StuckSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"targetLimitMode\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"targetValidatorsCount\",\"type\":\"uint256\"}],\"name\":\"TargetValidatorsCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalKeysCount\",\"type\":\"uint256\"}],\"name\":\"TotalSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"vettedKeysCount\",\"type\":\"uint256\"}],\"name\":\"VettedSigningKeysCountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"VettedSigningKeysCountDecreased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"keyIndex\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"WithdrawalSubmitted\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"EL_REWARDS_STEALING_FINE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"INITIAL_SLASHING_PENALTY\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"LIDO_LOCATOR\",\"outputs\":[{\"internalType\":\"contractILidoLocator\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_KEY_REMOVAL_CHARGE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_SIGNING_KEYS_PER_OPERATOR_BEFORE_PUBLIC_RELEASE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MODULE_MANAGER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PAUSE_INFINITELY\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PAUSE_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RECOVERER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"REPORT_EL_REWARDS_STEALING_PENALTY_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RESUME_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SETTLE_EL_REWARDS_STEALING_PENALTY_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"STAKING_ROUTER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"STETH\",\"outputs\":[{\"internalType\":\"contractIStETH\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VERIFIER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"accounting\",\"outputs\":[{\"internalType\":\"contractICSAccounting\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"activatePublicRelease\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"managerAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"rewardAddress\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"extendedManagerPermissions\",\"type\":\"bool\"}],\"internalType\":\"structNodeOperatorManagementProperties\",\"name\":\"managementProperties\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"eaProof\",\"type\":\"bytes32[]\"},{\"internalType\":\"address\",\"name\":\"referrer\",\"type\":\"address\"}],\"name\":\"addNodeOperatorETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"managerAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"rewardAddress\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"extendedManagerPermissions\",\"type\":\"bool\"}],\"internalType\":\"structNodeOperatorManagementProperties\",\"name\":\"managementProperties\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structICSAccounting.PermitInput\",\"name\":\"permit\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"eaProof\",\"type\":\"bytes32[]\"},{\"internalType\":\"address\",\"name\":\"referrer\",\"type\":\"address\"}],\"name\":\"addNodeOperatorStETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"managerAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"rewardAddress\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"extendedManagerPermissions\",\"type\":\"bool\"}],\"internalType\":\"structNodeOperatorManagementProperties\",\"name\":\"managementProperties\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structICSAccounting.PermitInput\",\"name\":\"permit\",\"type\":\"tuple\"},{\"internalType\":\"bytes32[]\",\"name\":\"eaProof\",\"type\":\"bytes32[]\"},{\"internalType\":\"address\",\"name\":\"referrer\",\"type\":\"address\"}],\"name\":\"addNodeOperatorWstETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"name\":\"addValidatorKeysETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structICSAccounting.PermitInput\",\"name\":\"permit\",\"type\":\"tuple\"}],\"name\":\"addValidatorKeysStETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structICSAccounting.PermitInput\",\"name\":\"permit\",\"type\":\"tuple\"}],\"name\":\"addValidatorKeysWstETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"cancelELRewardsStealingPenalty\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"changeNodeOperatorRewardAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stETHAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cumulativeFeeShares\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"rewardsProof\",\"type\":\"bytes32[]\"}],\"name\":\"claimRewardsStETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stEthAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cumulativeFeeShares\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"rewardsProof\",\"type\":\"bytes32[]\"}],\"name\":\"claimRewardsUnstETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"wstETHAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cumulativeFeeShares\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"rewardsProof\",\"type\":\"bytes32[]\"}],\"name\":\"claimRewardsWstETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"maxItems\",\"type\":\"uint256\"}],\"name\":\"cleanDepositQueue\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"removed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastRemovedAtDepth\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"compensateELRewardsStealingPenalty\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"confirmNodeOperatorManagerAddressChange\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"confirmNodeOperatorRewardAddressChange\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"nodeOperatorIds\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"vettedSigningKeysCounts\",\"type\":\"bytes\"}],\"name\":\"decreaseVettedSigningKeysCount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"depositETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"depositQueue\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"head\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"tail\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"index\",\"type\":\"uint128\"}],\"name\":\"depositQueueItem\",\"outputs\":[{\"internalType\":\"Batch\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stETHAmount\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structICSAccounting.PermitInput\",\"name\":\"permit\",\"type\":\"tuple\"}],\"name\":\"depositStETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"wstETHAmount\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structICSAccounting.PermitInput\",\"name\":\"permit\",\"type\":\"tuple\"}],\"name\":\"depositWstETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"earlyAdoption\",\"outputs\":[{\"internalType\":\"contractICSEarlyAdoption\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getActiveNodeOperatorsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getNodeOperator\",\"outputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"totalAddedKeys\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalWithdrawnKeys\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalDepositedKeys\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalVettedKeys\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"stuckValidatorsCount\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"depositableValidatorsCount\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"targetLimit\",\"type\":\"uint32\"},{\"internalType\":\"uint8\",\"name\":\"targetLimitMode\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"totalExitedKeys\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"enqueuedCount\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"managerAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"proposedManagerAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"rewardAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"proposedRewardAddress\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"extendedManagerPermissions\",\"type\":\"bool\"}],\"internalType\":\"structNodeOperator\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorIds\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"nodeOperatorIds\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorIsActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorNonWithdrawnKeys\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"getNodeOperatorSummary\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"targetLimitMode\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"targetValidatorsCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stuckValidatorsCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"refundedValidatorsCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stuckPenaltyEndTimestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalExitedValidators\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalDepositedValidators\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"depositableValidatorsCount\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNodeOperatorsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getResumeSinceTimestamp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getRoleMember\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleMemberCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"}],\"name\":\"getSigningKeys\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"}],\"name\":\"getSigningKeysWithSignatures\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"keys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getStakingModuleSummary\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"totalExitedValidators\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalDepositedValidators\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"depositableValidatorsCount\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getType\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_accounting\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_earlyAdoption\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_keyRemovalCharge\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"admin\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isPaused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keyIndex\",\"type\":\"uint256\"}],\"name\":\"isValidatorSlashed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keyIndex\",\"type\":\"uint256\"}],\"name\":\"isValidatorWithdrawn\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"keyRemovalCharge\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"normalizeQueue\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositsCount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"obtainDepositData\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"publicKeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"onExitedAndStuckValidatorsCountsUpdated\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"totalShares\",\"type\":\"uint256\"}],\"name\":\"onRewardsMinted\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"onWithdrawalCredentialsChanged\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"pauseFor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"proposedAddress\",\"type\":\"address\"}],\"name\":\"proposeNodeOperatorManagerAddressChange\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"proposedAddress\",\"type\":\"address\"}],\"name\":\"proposeNodeOperatorRewardAddressChange\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"publicRelease\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"recoverERC1155\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"recoverERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"recoverERC721\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"recoverEther\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keysCount\",\"type\":\"uint256\"}],\"name\":\"removeKeys\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"callerConfirmation\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"reportELRewardsStealingPenalty\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"}],\"name\":\"resetNodeOperatorManagerAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"resume\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"setKeyRemovalCharge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"nodeOperatorIds\",\"type\":\"uint256[]\"}],\"name\":\"settleELRewardsStealingPenalty\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keyIndex\",\"type\":\"uint256\"}],\"name\":\"submitInitialSlashing\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"keyIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isSlashed\",\"type\":\"bool\"}],\"name\":\"submitWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"exitedValidatorsKeysCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stuckValidatorsKeysCount\",\"type\":\"uint256\"}],\"name\":\"unsafeUpdateValidatorsCount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"nodeOperatorIds\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"exitedValidatorsCounts\",\"type\":\"bytes\"}],\"name\":\"updateExitedValidatorsCount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"updateRefundedValidatorsCount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"nodeOperatorIds\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"stuckValidatorsCounts\",\"type\":\"bytes\"}],\"name\":\"updateStuckValidatorsCount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeOperatorId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"targetLimitMode\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"targetLimit\",\"type\":\"uint256\"}],\"name\":\"updateTargetValidatorsLimits\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// CSMModuleABI is the input ABI used to generate the binding from.
// Deprecated: Use CSMModuleMetaData.ABI instead.
var CSMModuleABI = CSMModuleMetaData.ABI

// CSMModule is an auto generated Go binding around an Ethereum contract.
type CSMModule struct {
	CSMModuleCaller     // Read-only binding to the contract
	CSMModuleTransactor // Write-only binding to the contract
	CSMModuleFilterer   // Log filterer for contract events
}

// CSMModuleCaller is an auto generated read-only Go binding around an Ethereum contract.
type CSMModuleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CSMModuleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CSMModuleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CSMModuleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CSMModuleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CSMModuleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CSMModuleSession struct {
	Contract     *CSMModule        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CSMModuleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CSMModuleCallerSession struct {
	Contract *CSMModuleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// CSMModuleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CSMModuleTransactorSession struct {
	Contract     *CSMModuleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// CSMModuleRaw is an auto generated low-level Go binding around an Ethereum contract.
type CSMModuleRaw struct {
	Contract *CSMModule // Generic contract binding to access the raw methods on
}

// CSMModuleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CSMModuleCallerRaw struct {
	Contract *CSMModuleCaller // Generic read-only contract binding to access the raw methods on
}

// CSMModuleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CSMModuleTransactorRaw struct {
	Contract *CSMModuleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCSMModule creates a new instance of CSMModule, bound to a specific deployed contract.
func NewCSMModule(address common.Address, backend bind.ContractBackend) (*CSMModule, error) {
	contract, err := bindCSMModule(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CSMModule{CSMModuleCaller: CSMModuleCaller{contract: contract}, CSMModuleTransactor: CSMModuleTransactor{contract: contract}, CSMModuleFilterer: CSMModuleFilterer{contract: contract}}, nil
}

// NewCSMModuleCaller creates a new read-only instance of CSMModule, bound to a specific deployed contract.
func NewCSMModuleCaller(address common.Address, caller bind.ContractCaller) (*CSMModuleCaller, error) {
	contract, err := bindCSMModule(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CSMModuleCaller{contract: contract}, nil
}

// NewCSMModuleTransactor creates a new write-only instance of CSMModule, bound to a specific deployed contract.
func NewCSMModuleTransactor(address common.Address, transactor bind.ContractTransactor) (*CSMModuleTransactor, error) {
	contract, err := bindCSMModule(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CSMModuleTransactor{contract: contract}, nil
}

// NewCSMModuleFilterer creates a new log filterer instance of CSMModule, bound to a specific deployed contract.
func NewCSMModuleFilterer(address common.Address, filterer bind.ContractFilterer) (*CSMModuleFilterer, error) {
	contract, err := bindCSMModule(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CSMModuleFilterer{contract: contract}, nil
}

// bindCSMModule binds a generic wrapper to an already deployed contract.
func bindCSMModule(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CSMModuleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CSMModule *CSMModuleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CSMModule.Contract.CSMModuleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CSMModule *CSMModuleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CSMModule.Contract.CSMModuleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CSMModule *CSMModuleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CSMModule.Contract.CSMModuleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CSMModule *CSMModuleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CSMModule.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CSMModule *CSMModuleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CSMModule.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CSMModule *CSMModuleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CSMModule.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _CSMModule.Contract.DEFAULTADMINROLE(&_CSMModule.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _CSMModule.Contract.DEFAULTADMINROLE(&_CSMModule.CallOpts)
}

// ELREWARDSSTEALINGFINE is a free data retrieval call binding the contract method 0xbdac46a2.
//
// Solidity: function EL_REWARDS_STEALING_FINE() view returns(uint256)
func (_CSMModule *CSMModuleCaller) ELREWARDSSTEALINGFINE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "EL_REWARDS_STEALING_FINE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ELREWARDSSTEALINGFINE is a free data retrieval call binding the contract method 0xbdac46a2.
//
// Solidity: function EL_REWARDS_STEALING_FINE() view returns(uint256)
func (_CSMModule *CSMModuleSession) ELREWARDSSTEALINGFINE() (*big.Int, error) {
	return _CSMModule.Contract.ELREWARDSSTEALINGFINE(&_CSMModule.CallOpts)
}

// ELREWARDSSTEALINGFINE is a free data retrieval call binding the contract method 0xbdac46a2.
//
// Solidity: function EL_REWARDS_STEALING_FINE() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) ELREWARDSSTEALINGFINE() (*big.Int, error) {
	return _CSMModule.Contract.ELREWARDSSTEALINGFINE(&_CSMModule.CallOpts)
}

// INITIALSLASHINGPENALTY is a free data retrieval call binding the contract method 0xd6477919.
//
// Solidity: function INITIAL_SLASHING_PENALTY() view returns(uint256)
func (_CSMModule *CSMModuleCaller) INITIALSLASHINGPENALTY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "INITIAL_SLASHING_PENALTY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// INITIALSLASHINGPENALTY is a free data retrieval call binding the contract method 0xd6477919.
//
// Solidity: function INITIAL_SLASHING_PENALTY() view returns(uint256)
func (_CSMModule *CSMModuleSession) INITIALSLASHINGPENALTY() (*big.Int, error) {
	return _CSMModule.Contract.INITIALSLASHINGPENALTY(&_CSMModule.CallOpts)
}

// INITIALSLASHINGPENALTY is a free data retrieval call binding the contract method 0xd6477919.
//
// Solidity: function INITIAL_SLASHING_PENALTY() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) INITIALSLASHINGPENALTY() (*big.Int, error) {
	return _CSMModule.Contract.INITIALSLASHINGPENALTY(&_CSMModule.CallOpts)
}

// LIDOLOCATOR is a free data retrieval call binding the contract method 0xdbba4b48.
//
// Solidity: function LIDO_LOCATOR() view returns(address)
func (_CSMModule *CSMModuleCaller) LIDOLOCATOR(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "LIDO_LOCATOR")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// LIDOLOCATOR is a free data retrieval call binding the contract method 0xdbba4b48.
//
// Solidity: function LIDO_LOCATOR() view returns(address)
func (_CSMModule *CSMModuleSession) LIDOLOCATOR() (common.Address, error) {
	return _CSMModule.Contract.LIDOLOCATOR(&_CSMModule.CallOpts)
}

// LIDOLOCATOR is a free data retrieval call binding the contract method 0xdbba4b48.
//
// Solidity: function LIDO_LOCATOR() view returns(address)
func (_CSMModule *CSMModuleCallerSession) LIDOLOCATOR() (common.Address, error) {
	return _CSMModule.Contract.LIDOLOCATOR(&_CSMModule.CallOpts)
}

// MAXKEYREMOVALCHARGE is a free data retrieval call binding the contract method 0x4baf13cc.
//
// Solidity: function MAX_KEY_REMOVAL_CHARGE() view returns(uint256)
func (_CSMModule *CSMModuleCaller) MAXKEYREMOVALCHARGE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "MAX_KEY_REMOVAL_CHARGE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXKEYREMOVALCHARGE is a free data retrieval call binding the contract method 0x4baf13cc.
//
// Solidity: function MAX_KEY_REMOVAL_CHARGE() view returns(uint256)
func (_CSMModule *CSMModuleSession) MAXKEYREMOVALCHARGE() (*big.Int, error) {
	return _CSMModule.Contract.MAXKEYREMOVALCHARGE(&_CSMModule.CallOpts)
}

// MAXKEYREMOVALCHARGE is a free data retrieval call binding the contract method 0x4baf13cc.
//
// Solidity: function MAX_KEY_REMOVAL_CHARGE() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) MAXKEYREMOVALCHARGE() (*big.Int, error) {
	return _CSMModule.Contract.MAXKEYREMOVALCHARGE(&_CSMModule.CallOpts)
}

// MAXSIGNINGKEYSPEROPERATORBEFOREPUBLICRELEASE is a free data retrieval call binding the contract method 0x47faf311.
//
// Solidity: function MAX_SIGNING_KEYS_PER_OPERATOR_BEFORE_PUBLIC_RELEASE() view returns(uint256)
func (_CSMModule *CSMModuleCaller) MAXSIGNINGKEYSPEROPERATORBEFOREPUBLICRELEASE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "MAX_SIGNING_KEYS_PER_OPERATOR_BEFORE_PUBLIC_RELEASE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXSIGNINGKEYSPEROPERATORBEFOREPUBLICRELEASE is a free data retrieval call binding the contract method 0x47faf311.
//
// Solidity: function MAX_SIGNING_KEYS_PER_OPERATOR_BEFORE_PUBLIC_RELEASE() view returns(uint256)
func (_CSMModule *CSMModuleSession) MAXSIGNINGKEYSPEROPERATORBEFOREPUBLICRELEASE() (*big.Int, error) {
	return _CSMModule.Contract.MAXSIGNINGKEYSPEROPERATORBEFOREPUBLICRELEASE(&_CSMModule.CallOpts)
}

// MAXSIGNINGKEYSPEROPERATORBEFOREPUBLICRELEASE is a free data retrieval call binding the contract method 0x47faf311.
//
// Solidity: function MAX_SIGNING_KEYS_PER_OPERATOR_BEFORE_PUBLIC_RELEASE() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) MAXSIGNINGKEYSPEROPERATORBEFOREPUBLICRELEASE() (*big.Int, error) {
	return _CSMModule.Contract.MAXSIGNINGKEYSPEROPERATORBEFOREPUBLICRELEASE(&_CSMModule.CallOpts)
}

// MODULEMANAGERROLE is a free data retrieval call binding the contract method 0xcb17ed3e.
//
// Solidity: function MODULE_MANAGER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) MODULEMANAGERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "MODULE_MANAGER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MODULEMANAGERROLE is a free data retrieval call binding the contract method 0xcb17ed3e.
//
// Solidity: function MODULE_MANAGER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) MODULEMANAGERROLE() ([32]byte, error) {
	return _CSMModule.Contract.MODULEMANAGERROLE(&_CSMModule.CallOpts)
}

// MODULEMANAGERROLE is a free data retrieval call binding the contract method 0xcb17ed3e.
//
// Solidity: function MODULE_MANAGER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) MODULEMANAGERROLE() ([32]byte, error) {
	return _CSMModule.Contract.MODULEMANAGERROLE(&_CSMModule.CallOpts)
}

// PAUSEINFINITELY is a free data retrieval call binding the contract method 0xa302ee38.
//
// Solidity: function PAUSE_INFINITELY() view returns(uint256)
func (_CSMModule *CSMModuleCaller) PAUSEINFINITELY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "PAUSE_INFINITELY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PAUSEINFINITELY is a free data retrieval call binding the contract method 0xa302ee38.
//
// Solidity: function PAUSE_INFINITELY() view returns(uint256)
func (_CSMModule *CSMModuleSession) PAUSEINFINITELY() (*big.Int, error) {
	return _CSMModule.Contract.PAUSEINFINITELY(&_CSMModule.CallOpts)
}

// PAUSEINFINITELY is a free data retrieval call binding the contract method 0xa302ee38.
//
// Solidity: function PAUSE_INFINITELY() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) PAUSEINFINITELY() (*big.Int, error) {
	return _CSMModule.Contract.PAUSEINFINITELY(&_CSMModule.CallOpts)
}

// PAUSEROLE is a free data retrieval call binding the contract method 0x389ed267.
//
// Solidity: function PAUSE_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) PAUSEROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "PAUSE_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PAUSEROLE is a free data retrieval call binding the contract method 0x389ed267.
//
// Solidity: function PAUSE_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) PAUSEROLE() ([32]byte, error) {
	return _CSMModule.Contract.PAUSEROLE(&_CSMModule.CallOpts)
}

// PAUSEROLE is a free data retrieval call binding the contract method 0x389ed267.
//
// Solidity: function PAUSE_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) PAUSEROLE() ([32]byte, error) {
	return _CSMModule.Contract.PAUSEROLE(&_CSMModule.CallOpts)
}

// RECOVERERROLE is a free data retrieval call binding the contract method 0xacf1c948.
//
// Solidity: function RECOVERER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) RECOVERERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "RECOVERER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RECOVERERROLE is a free data retrieval call binding the contract method 0xacf1c948.
//
// Solidity: function RECOVERER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) RECOVERERROLE() ([32]byte, error) {
	return _CSMModule.Contract.RECOVERERROLE(&_CSMModule.CallOpts)
}

// RECOVERERROLE is a free data retrieval call binding the contract method 0xacf1c948.
//
// Solidity: function RECOVERER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) RECOVERERROLE() ([32]byte, error) {
	return _CSMModule.Contract.RECOVERERROLE(&_CSMModule.CallOpts)
}

// REPORTELREWARDSSTEALINGPENALTYROLE is a free data retrieval call binding the contract method 0x8573e351.
//
// Solidity: function REPORT_EL_REWARDS_STEALING_PENALTY_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) REPORTELREWARDSSTEALINGPENALTYROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "REPORT_EL_REWARDS_STEALING_PENALTY_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// REPORTELREWARDSSTEALINGPENALTYROLE is a free data retrieval call binding the contract method 0x8573e351.
//
// Solidity: function REPORT_EL_REWARDS_STEALING_PENALTY_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) REPORTELREWARDSSTEALINGPENALTYROLE() ([32]byte, error) {
	return _CSMModule.Contract.REPORTELREWARDSSTEALINGPENALTYROLE(&_CSMModule.CallOpts)
}

// REPORTELREWARDSSTEALINGPENALTYROLE is a free data retrieval call binding the contract method 0x8573e351.
//
// Solidity: function REPORT_EL_REWARDS_STEALING_PENALTY_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) REPORTELREWARDSSTEALINGPENALTYROLE() ([32]byte, error) {
	return _CSMModule.Contract.REPORTELREWARDSSTEALINGPENALTYROLE(&_CSMModule.CallOpts)
}

// RESUMEROLE is a free data retrieval call binding the contract method 0x2de03aa1.
//
// Solidity: function RESUME_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) RESUMEROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "RESUME_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RESUMEROLE is a free data retrieval call binding the contract method 0x2de03aa1.
//
// Solidity: function RESUME_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) RESUMEROLE() ([32]byte, error) {
	return _CSMModule.Contract.RESUMEROLE(&_CSMModule.CallOpts)
}

// RESUMEROLE is a free data retrieval call binding the contract method 0x2de03aa1.
//
// Solidity: function RESUME_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) RESUMEROLE() ([32]byte, error) {
	return _CSMModule.Contract.RESUMEROLE(&_CSMModule.CallOpts)
}

// SETTLEELREWARDSSTEALINGPENALTYROLE is a free data retrieval call binding the contract method 0x3f04f0c8.
//
// Solidity: function SETTLE_EL_REWARDS_STEALING_PENALTY_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) SETTLEELREWARDSSTEALINGPENALTYROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "SETTLE_EL_REWARDS_STEALING_PENALTY_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SETTLEELREWARDSSTEALINGPENALTYROLE is a free data retrieval call binding the contract method 0x3f04f0c8.
//
// Solidity: function SETTLE_EL_REWARDS_STEALING_PENALTY_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) SETTLEELREWARDSSTEALINGPENALTYROLE() ([32]byte, error) {
	return _CSMModule.Contract.SETTLEELREWARDSSTEALINGPENALTYROLE(&_CSMModule.CallOpts)
}

// SETTLEELREWARDSSTEALINGPENALTYROLE is a free data retrieval call binding the contract method 0x3f04f0c8.
//
// Solidity: function SETTLE_EL_REWARDS_STEALING_PENALTY_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) SETTLEELREWARDSSTEALINGPENALTYROLE() ([32]byte, error) {
	return _CSMModule.Contract.SETTLEELREWARDSSTEALINGPENALTYROLE(&_CSMModule.CallOpts)
}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) STAKINGROUTERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "STAKING_ROUTER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) STAKINGROUTERROLE() ([32]byte, error) {
	return _CSMModule.Contract.STAKINGROUTERROLE(&_CSMModule.CallOpts)
}

// STAKINGROUTERROLE is a free data retrieval call binding the contract method 0x80231f15.
//
// Solidity: function STAKING_ROUTER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) STAKINGROUTERROLE() ([32]byte, error) {
	return _CSMModule.Contract.STAKINGROUTERROLE(&_CSMModule.CallOpts)
}

// STETH is a free data retrieval call binding the contract method 0xe00bfe50.
//
// Solidity: function STETH() view returns(address)
func (_CSMModule *CSMModuleCaller) STETH(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "STETH")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// STETH is a free data retrieval call binding the contract method 0xe00bfe50.
//
// Solidity: function STETH() view returns(address)
func (_CSMModule *CSMModuleSession) STETH() (common.Address, error) {
	return _CSMModule.Contract.STETH(&_CSMModule.CallOpts)
}

// STETH is a free data retrieval call binding the contract method 0xe00bfe50.
//
// Solidity: function STETH() view returns(address)
func (_CSMModule *CSMModuleCallerSession) STETH() (common.Address, error) {
	return _CSMModule.Contract.STETH(&_CSMModule.CallOpts)
}

// VERIFIERROLE is a free data retrieval call binding the contract method 0xe7705db6.
//
// Solidity: function VERIFIER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) VERIFIERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "VERIFIER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VERIFIERROLE is a free data retrieval call binding the contract method 0xe7705db6.
//
// Solidity: function VERIFIER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleSession) VERIFIERROLE() ([32]byte, error) {
	return _CSMModule.Contract.VERIFIERROLE(&_CSMModule.CallOpts)
}

// VERIFIERROLE is a free data retrieval call binding the contract method 0xe7705db6.
//
// Solidity: function VERIFIER_ROLE() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) VERIFIERROLE() ([32]byte, error) {
	return _CSMModule.Contract.VERIFIERROLE(&_CSMModule.CallOpts)
}

// Accounting is a free data retrieval call binding the contract method 0x9624e83e.
//
// Solidity: function accounting() view returns(address)
func (_CSMModule *CSMModuleCaller) Accounting(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "accounting")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Accounting is a free data retrieval call binding the contract method 0x9624e83e.
//
// Solidity: function accounting() view returns(address)
func (_CSMModule *CSMModuleSession) Accounting() (common.Address, error) {
	return _CSMModule.Contract.Accounting(&_CSMModule.CallOpts)
}

// Accounting is a free data retrieval call binding the contract method 0x9624e83e.
//
// Solidity: function accounting() view returns(address)
func (_CSMModule *CSMModuleCallerSession) Accounting() (common.Address, error) {
	return _CSMModule.Contract.Accounting(&_CSMModule.CallOpts)
}

// DepositQueue is a free data retrieval call binding the contract method 0xf617eecc.
//
// Solidity: function depositQueue() view returns(uint128 head, uint128 tail)
func (_CSMModule *CSMModuleCaller) DepositQueue(opts *bind.CallOpts) (struct {
	Head *big.Int
	Tail *big.Int
}, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "depositQueue")

	outstruct := new(struct {
		Head *big.Int
		Tail *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Head = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Tail = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// DepositQueue is a free data retrieval call binding the contract method 0xf617eecc.
//
// Solidity: function depositQueue() view returns(uint128 head, uint128 tail)
func (_CSMModule *CSMModuleSession) DepositQueue() (struct {
	Head *big.Int
	Tail *big.Int
}, error) {
	return _CSMModule.Contract.DepositQueue(&_CSMModule.CallOpts)
}

// DepositQueue is a free data retrieval call binding the contract method 0xf617eecc.
//
// Solidity: function depositQueue() view returns(uint128 head, uint128 tail)
func (_CSMModule *CSMModuleCallerSession) DepositQueue() (struct {
	Head *big.Int
	Tail *big.Int
}, error) {
	return _CSMModule.Contract.DepositQueue(&_CSMModule.CallOpts)
}

// DepositQueueItem is a free data retrieval call binding the contract method 0x5e169fb8.
//
// Solidity: function depositQueueItem(uint128 index) view returns(uint256)
func (_CSMModule *CSMModuleCaller) DepositQueueItem(opts *bind.CallOpts, index *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "depositQueueItem", index)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DepositQueueItem is a free data retrieval call binding the contract method 0x5e169fb8.
//
// Solidity: function depositQueueItem(uint128 index) view returns(uint256)
func (_CSMModule *CSMModuleSession) DepositQueueItem(index *big.Int) (*big.Int, error) {
	return _CSMModule.Contract.DepositQueueItem(&_CSMModule.CallOpts, index)
}

// DepositQueueItem is a free data retrieval call binding the contract method 0x5e169fb8.
//
// Solidity: function depositQueueItem(uint128 index) view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) DepositQueueItem(index *big.Int) (*big.Int, error) {
	return _CSMModule.Contract.DepositQueueItem(&_CSMModule.CallOpts, index)
}

// EarlyAdoption is a free data retrieval call binding the contract method 0x26a666e4.
//
// Solidity: function earlyAdoption() view returns(address)
func (_CSMModule *CSMModuleCaller) EarlyAdoption(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "earlyAdoption")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EarlyAdoption is a free data retrieval call binding the contract method 0x26a666e4.
//
// Solidity: function earlyAdoption() view returns(address)
func (_CSMModule *CSMModuleSession) EarlyAdoption() (common.Address, error) {
	return _CSMModule.Contract.EarlyAdoption(&_CSMModule.CallOpts)
}

// EarlyAdoption is a free data retrieval call binding the contract method 0x26a666e4.
//
// Solidity: function earlyAdoption() view returns(address)
func (_CSMModule *CSMModuleCallerSession) EarlyAdoption() (common.Address, error) {
	return _CSMModule.Contract.EarlyAdoption(&_CSMModule.CallOpts)
}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_CSMModule *CSMModuleCaller) GetActiveNodeOperatorsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getActiveNodeOperatorsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_CSMModule *CSMModuleSession) GetActiveNodeOperatorsCount() (*big.Int, error) {
	return _CSMModule.Contract.GetActiveNodeOperatorsCount(&_CSMModule.CallOpts)
}

// GetActiveNodeOperatorsCount is a free data retrieval call binding the contract method 0x8469cbd3.
//
// Solidity: function getActiveNodeOperatorsCount() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) GetActiveNodeOperatorsCount() (*big.Int, error) {
	return _CSMModule.Contract.GetActiveNodeOperatorsCount(&_CSMModule.CallOpts)
}

// GetNodeOperator is a free data retrieval call binding the contract method 0x65c14dc7.
//
// Solidity: function getNodeOperator(uint256 nodeOperatorId) view returns((uint32,uint32,uint32,uint32,uint32,uint32,uint32,uint8,uint32,uint32,address,address,address,address,bool))
func (_CSMModule *CSMModuleCaller) GetNodeOperator(opts *bind.CallOpts, nodeOperatorId *big.Int) (NodeOperator, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getNodeOperator", nodeOperatorId)

	if err != nil {
		return *new(NodeOperator), err
	}

	out0 := *abi.ConvertType(out[0], new(NodeOperator)).(*NodeOperator)

	return out0, err

}

// GetNodeOperator is a free data retrieval call binding the contract method 0x65c14dc7.
//
// Solidity: function getNodeOperator(uint256 nodeOperatorId) view returns((uint32,uint32,uint32,uint32,uint32,uint32,uint32,uint8,uint32,uint32,address,address,address,address,bool))
func (_CSMModule *CSMModuleSession) GetNodeOperator(nodeOperatorId *big.Int) (NodeOperator, error) {
	return _CSMModule.Contract.GetNodeOperator(&_CSMModule.CallOpts, nodeOperatorId)
}

// GetNodeOperator is a free data retrieval call binding the contract method 0x65c14dc7.
//
// Solidity: function getNodeOperator(uint256 nodeOperatorId) view returns((uint32,uint32,uint32,uint32,uint32,uint32,uint32,uint8,uint32,uint32,address,address,address,address,bool))
func (_CSMModule *CSMModuleCallerSession) GetNodeOperator(nodeOperatorId *big.Int) (NodeOperator, error) {
	return _CSMModule.Contract.GetNodeOperator(&_CSMModule.CallOpts, nodeOperatorId)
}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 offset, uint256 limit) view returns(uint256[] nodeOperatorIds)
func (_CSMModule *CSMModuleCaller) GetNodeOperatorIds(opts *bind.CallOpts, offset *big.Int, limit *big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getNodeOperatorIds", offset, limit)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 offset, uint256 limit) view returns(uint256[] nodeOperatorIds)
func (_CSMModule *CSMModuleSession) GetNodeOperatorIds(offset *big.Int, limit *big.Int) ([]*big.Int, error) {
	return _CSMModule.Contract.GetNodeOperatorIds(&_CSMModule.CallOpts, offset, limit)
}

// GetNodeOperatorIds is a free data retrieval call binding the contract method 0x4febc81b.
//
// Solidity: function getNodeOperatorIds(uint256 offset, uint256 limit) view returns(uint256[] nodeOperatorIds)
func (_CSMModule *CSMModuleCallerSession) GetNodeOperatorIds(offset *big.Int, limit *big.Int) ([]*big.Int, error) {
	return _CSMModule.Contract.GetNodeOperatorIds(&_CSMModule.CallOpts, offset, limit)
}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 nodeOperatorId) view returns(bool)
func (_CSMModule *CSMModuleCaller) GetNodeOperatorIsActive(opts *bind.CallOpts, nodeOperatorId *big.Int) (bool, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getNodeOperatorIsActive", nodeOperatorId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 nodeOperatorId) view returns(bool)
func (_CSMModule *CSMModuleSession) GetNodeOperatorIsActive(nodeOperatorId *big.Int) (bool, error) {
	return _CSMModule.Contract.GetNodeOperatorIsActive(&_CSMModule.CallOpts, nodeOperatorId)
}

// GetNodeOperatorIsActive is a free data retrieval call binding the contract method 0x5e2fb908.
//
// Solidity: function getNodeOperatorIsActive(uint256 nodeOperatorId) view returns(bool)
func (_CSMModule *CSMModuleCallerSession) GetNodeOperatorIsActive(nodeOperatorId *big.Int) (bool, error) {
	return _CSMModule.Contract.GetNodeOperatorIsActive(&_CSMModule.CallOpts, nodeOperatorId)
}

// GetNodeOperatorNonWithdrawnKeys is a free data retrieval call binding the contract method 0x8ec69028.
//
// Solidity: function getNodeOperatorNonWithdrawnKeys(uint256 nodeOperatorId) view returns(uint256)
func (_CSMModule *CSMModuleCaller) GetNodeOperatorNonWithdrawnKeys(opts *bind.CallOpts, nodeOperatorId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getNodeOperatorNonWithdrawnKeys", nodeOperatorId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNodeOperatorNonWithdrawnKeys is a free data retrieval call binding the contract method 0x8ec69028.
//
// Solidity: function getNodeOperatorNonWithdrawnKeys(uint256 nodeOperatorId) view returns(uint256)
func (_CSMModule *CSMModuleSession) GetNodeOperatorNonWithdrawnKeys(nodeOperatorId *big.Int) (*big.Int, error) {
	return _CSMModule.Contract.GetNodeOperatorNonWithdrawnKeys(&_CSMModule.CallOpts, nodeOperatorId)
}

// GetNodeOperatorNonWithdrawnKeys is a free data retrieval call binding the contract method 0x8ec69028.
//
// Solidity: function getNodeOperatorNonWithdrawnKeys(uint256 nodeOperatorId) view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) GetNodeOperatorNonWithdrawnKeys(nodeOperatorId *big.Int) (*big.Int, error) {
	return _CSMModule.Contract.GetNodeOperatorNonWithdrawnKeys(&_CSMModule.CallOpts, nodeOperatorId)
}

// GetNodeOperatorSummary is a free data retrieval call binding the contract method 0xb3076c3c.
//
// Solidity: function getNodeOperatorSummary(uint256 nodeOperatorId) view returns(uint256 targetLimitMode, uint256 targetValidatorsCount, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp, uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_CSMModule *CSMModuleCaller) GetNodeOperatorSummary(opts *bind.CallOpts, nodeOperatorId *big.Int) (struct {
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
	err := _CSMModule.contract.Call(opts, &out, "getNodeOperatorSummary", nodeOperatorId)

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
// Solidity: function getNodeOperatorSummary(uint256 nodeOperatorId) view returns(uint256 targetLimitMode, uint256 targetValidatorsCount, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp, uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_CSMModule *CSMModuleSession) GetNodeOperatorSummary(nodeOperatorId *big.Int) (struct {
	TargetLimitMode            *big.Int
	TargetValidatorsCount      *big.Int
	StuckValidatorsCount       *big.Int
	RefundedValidatorsCount    *big.Int
	StuckPenaltyEndTimestamp   *big.Int
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _CSMModule.Contract.GetNodeOperatorSummary(&_CSMModule.CallOpts, nodeOperatorId)
}

// GetNodeOperatorSummary is a free data retrieval call binding the contract method 0xb3076c3c.
//
// Solidity: function getNodeOperatorSummary(uint256 nodeOperatorId) view returns(uint256 targetLimitMode, uint256 targetValidatorsCount, uint256 stuckValidatorsCount, uint256 refundedValidatorsCount, uint256 stuckPenaltyEndTimestamp, uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_CSMModule *CSMModuleCallerSession) GetNodeOperatorSummary(nodeOperatorId *big.Int) (struct {
	TargetLimitMode            *big.Int
	TargetValidatorsCount      *big.Int
	StuckValidatorsCount       *big.Int
	RefundedValidatorsCount    *big.Int
	StuckPenaltyEndTimestamp   *big.Int
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _CSMModule.Contract.GetNodeOperatorSummary(&_CSMModule.CallOpts, nodeOperatorId)
}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_CSMModule *CSMModuleCaller) GetNodeOperatorsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getNodeOperatorsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_CSMModule *CSMModuleSession) GetNodeOperatorsCount() (*big.Int, error) {
	return _CSMModule.Contract.GetNodeOperatorsCount(&_CSMModule.CallOpts)
}

// GetNodeOperatorsCount is a free data retrieval call binding the contract method 0xa70c70e4.
//
// Solidity: function getNodeOperatorsCount() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) GetNodeOperatorsCount() (*big.Int, error) {
	return _CSMModule.Contract.GetNodeOperatorsCount(&_CSMModule.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_CSMModule *CSMModuleCaller) GetNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_CSMModule *CSMModuleSession) GetNonce() (*big.Int, error) {
	return _CSMModule.Contract.GetNonce(&_CSMModule.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) GetNonce() (*big.Int, error) {
	return _CSMModule.Contract.GetNonce(&_CSMModule.CallOpts)
}

// GetResumeSinceTimestamp is a free data retrieval call binding the contract method 0x589ff76c.
//
// Solidity: function getResumeSinceTimestamp() view returns(uint256)
func (_CSMModule *CSMModuleCaller) GetResumeSinceTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getResumeSinceTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetResumeSinceTimestamp is a free data retrieval call binding the contract method 0x589ff76c.
//
// Solidity: function getResumeSinceTimestamp() view returns(uint256)
func (_CSMModule *CSMModuleSession) GetResumeSinceTimestamp() (*big.Int, error) {
	return _CSMModule.Contract.GetResumeSinceTimestamp(&_CSMModule.CallOpts)
}

// GetResumeSinceTimestamp is a free data retrieval call binding the contract method 0x589ff76c.
//
// Solidity: function getResumeSinceTimestamp() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) GetResumeSinceTimestamp() (*big.Int, error) {
	return _CSMModule.Contract.GetResumeSinceTimestamp(&_CSMModule.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_CSMModule *CSMModuleCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_CSMModule *CSMModuleSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _CSMModule.Contract.GetRoleAdmin(&_CSMModule.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _CSMModule.Contract.GetRoleAdmin(&_CSMModule.CallOpts, role)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_CSMModule *CSMModuleCaller) GetRoleMember(opts *bind.CallOpts, role [32]byte, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getRoleMember", role, index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_CSMModule *CSMModuleSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _CSMModule.Contract.GetRoleMember(&_CSMModule.CallOpts, role, index)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_CSMModule *CSMModuleCallerSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _CSMModule.Contract.GetRoleMember(&_CSMModule.CallOpts, role, index)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_CSMModule *CSMModuleCaller) GetRoleMemberCount(opts *bind.CallOpts, role [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getRoleMemberCount", role)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_CSMModule *CSMModuleSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _CSMModule.Contract.GetRoleMemberCount(&_CSMModule.CallOpts, role)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _CSMModule.Contract.GetRoleMemberCount(&_CSMModule.CallOpts, role)
}

// GetSigningKeys is a free data retrieval call binding the contract method 0x59e25c12.
//
// Solidity: function getSigningKeys(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) view returns(bytes)
func (_CSMModule *CSMModuleCaller) GetSigningKeys(opts *bind.CallOpts, nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) ([]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getSigningKeys", nodeOperatorId, startIndex, keysCount)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetSigningKeys is a free data retrieval call binding the contract method 0x59e25c12.
//
// Solidity: function getSigningKeys(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) view returns(bytes)
func (_CSMModule *CSMModuleSession) GetSigningKeys(nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) ([]byte, error) {
	return _CSMModule.Contract.GetSigningKeys(&_CSMModule.CallOpts, nodeOperatorId, startIndex, keysCount)
}

// GetSigningKeys is a free data retrieval call binding the contract method 0x59e25c12.
//
// Solidity: function getSigningKeys(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) view returns(bytes)
func (_CSMModule *CSMModuleCallerSession) GetSigningKeys(nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) ([]byte, error) {
	return _CSMModule.Contract.GetSigningKeys(&_CSMModule.CallOpts, nodeOperatorId, startIndex, keysCount)
}

// GetSigningKeysWithSignatures is a free data retrieval call binding the contract method 0x50388cb6.
//
// Solidity: function getSigningKeysWithSignatures(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) view returns(bytes keys, bytes signatures)
func (_CSMModule *CSMModuleCaller) GetSigningKeysWithSignatures(opts *bind.CallOpts, nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) (struct {
	Keys       []byte
	Signatures []byte
}, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getSigningKeysWithSignatures", nodeOperatorId, startIndex, keysCount)

	outstruct := new(struct {
		Keys       []byte
		Signatures []byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Keys = *abi.ConvertType(out[0], new([]byte)).(*[]byte)
	outstruct.Signatures = *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return *outstruct, err

}

// GetSigningKeysWithSignatures is a free data retrieval call binding the contract method 0x50388cb6.
//
// Solidity: function getSigningKeysWithSignatures(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) view returns(bytes keys, bytes signatures)
func (_CSMModule *CSMModuleSession) GetSigningKeysWithSignatures(nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) (struct {
	Keys       []byte
	Signatures []byte
}, error) {
	return _CSMModule.Contract.GetSigningKeysWithSignatures(&_CSMModule.CallOpts, nodeOperatorId, startIndex, keysCount)
}

// GetSigningKeysWithSignatures is a free data retrieval call binding the contract method 0x50388cb6.
//
// Solidity: function getSigningKeysWithSignatures(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) view returns(bytes keys, bytes signatures)
func (_CSMModule *CSMModuleCallerSession) GetSigningKeysWithSignatures(nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) (struct {
	Keys       []byte
	Signatures []byte
}, error) {
	return _CSMModule.Contract.GetSigningKeysWithSignatures(&_CSMModule.CallOpts, nodeOperatorId, startIndex, keysCount)
}

// GetStakingModuleSummary is a free data retrieval call binding the contract method 0x9abddf09.
//
// Solidity: function getStakingModuleSummary() view returns(uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_CSMModule *CSMModuleCaller) GetStakingModuleSummary(opts *bind.CallOpts) (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getStakingModuleSummary")

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
func (_CSMModule *CSMModuleSession) GetStakingModuleSummary() (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _CSMModule.Contract.GetStakingModuleSummary(&_CSMModule.CallOpts)
}

// GetStakingModuleSummary is a free data retrieval call binding the contract method 0x9abddf09.
//
// Solidity: function getStakingModuleSummary() view returns(uint256 totalExitedValidators, uint256 totalDepositedValidators, uint256 depositableValidatorsCount)
func (_CSMModule *CSMModuleCallerSession) GetStakingModuleSummary() (struct {
	TotalExitedValidators      *big.Int
	TotalDepositedValidators   *big.Int
	DepositableValidatorsCount *big.Int
}, error) {
	return _CSMModule.Contract.GetStakingModuleSummary(&_CSMModule.CallOpts)
}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_CSMModule *CSMModuleCaller) GetType(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "getType")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_CSMModule *CSMModuleSession) GetType() ([32]byte, error) {
	return _CSMModule.Contract.GetType(&_CSMModule.CallOpts)
}

// GetType is a free data retrieval call binding the contract method 0x15dae03e.
//
// Solidity: function getType() view returns(bytes32)
func (_CSMModule *CSMModuleCallerSession) GetType() ([32]byte, error) {
	return _CSMModule.Contract.GetType(&_CSMModule.CallOpts)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_CSMModule *CSMModuleCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_CSMModule *CSMModuleSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _CSMModule.Contract.HasRole(&_CSMModule.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_CSMModule *CSMModuleCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _CSMModule.Contract.HasRole(&_CSMModule.CallOpts, role, account)
}

// IsPaused is a free data retrieval call binding the contract method 0xb187bd26.
//
// Solidity: function isPaused() view returns(bool)
func (_CSMModule *CSMModuleCaller) IsPaused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "isPaused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsPaused is a free data retrieval call binding the contract method 0xb187bd26.
//
// Solidity: function isPaused() view returns(bool)
func (_CSMModule *CSMModuleSession) IsPaused() (bool, error) {
	return _CSMModule.Contract.IsPaused(&_CSMModule.CallOpts)
}

// IsPaused is a free data retrieval call binding the contract method 0xb187bd26.
//
// Solidity: function isPaused() view returns(bool)
func (_CSMModule *CSMModuleCallerSession) IsPaused() (bool, error) {
	return _CSMModule.Contract.IsPaused(&_CSMModule.CallOpts)
}

// IsValidatorSlashed is a free data retrieval call binding the contract method 0x3dbe8b5a.
//
// Solidity: function isValidatorSlashed(uint256 nodeOperatorId, uint256 keyIndex) view returns(bool)
func (_CSMModule *CSMModuleCaller) IsValidatorSlashed(opts *bind.CallOpts, nodeOperatorId *big.Int, keyIndex *big.Int) (bool, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "isValidatorSlashed", nodeOperatorId, keyIndex)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidatorSlashed is a free data retrieval call binding the contract method 0x3dbe8b5a.
//
// Solidity: function isValidatorSlashed(uint256 nodeOperatorId, uint256 keyIndex) view returns(bool)
func (_CSMModule *CSMModuleSession) IsValidatorSlashed(nodeOperatorId *big.Int, keyIndex *big.Int) (bool, error) {
	return _CSMModule.Contract.IsValidatorSlashed(&_CSMModule.CallOpts, nodeOperatorId, keyIndex)
}

// IsValidatorSlashed is a free data retrieval call binding the contract method 0x3dbe8b5a.
//
// Solidity: function isValidatorSlashed(uint256 nodeOperatorId, uint256 keyIndex) view returns(bool)
func (_CSMModule *CSMModuleCallerSession) IsValidatorSlashed(nodeOperatorId *big.Int, keyIndex *big.Int) (bool, error) {
	return _CSMModule.Contract.IsValidatorSlashed(&_CSMModule.CallOpts, nodeOperatorId, keyIndex)
}

// IsValidatorWithdrawn is a free data retrieval call binding the contract method 0x53433643.
//
// Solidity: function isValidatorWithdrawn(uint256 nodeOperatorId, uint256 keyIndex) view returns(bool)
func (_CSMModule *CSMModuleCaller) IsValidatorWithdrawn(opts *bind.CallOpts, nodeOperatorId *big.Int, keyIndex *big.Int) (bool, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "isValidatorWithdrawn", nodeOperatorId, keyIndex)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidatorWithdrawn is a free data retrieval call binding the contract method 0x53433643.
//
// Solidity: function isValidatorWithdrawn(uint256 nodeOperatorId, uint256 keyIndex) view returns(bool)
func (_CSMModule *CSMModuleSession) IsValidatorWithdrawn(nodeOperatorId *big.Int, keyIndex *big.Int) (bool, error) {
	return _CSMModule.Contract.IsValidatorWithdrawn(&_CSMModule.CallOpts, nodeOperatorId, keyIndex)
}

// IsValidatorWithdrawn is a free data retrieval call binding the contract method 0x53433643.
//
// Solidity: function isValidatorWithdrawn(uint256 nodeOperatorId, uint256 keyIndex) view returns(bool)
func (_CSMModule *CSMModuleCallerSession) IsValidatorWithdrawn(nodeOperatorId *big.Int, keyIndex *big.Int) (bool, error) {
	return _CSMModule.Contract.IsValidatorWithdrawn(&_CSMModule.CallOpts, nodeOperatorId, keyIndex)
}

// KeyRemovalCharge is a free data retrieval call binding the contract method 0xd9df8c92.
//
// Solidity: function keyRemovalCharge() view returns(uint256)
func (_CSMModule *CSMModuleCaller) KeyRemovalCharge(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "keyRemovalCharge")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// KeyRemovalCharge is a free data retrieval call binding the contract method 0xd9df8c92.
//
// Solidity: function keyRemovalCharge() view returns(uint256)
func (_CSMModule *CSMModuleSession) KeyRemovalCharge() (*big.Int, error) {
	return _CSMModule.Contract.KeyRemovalCharge(&_CSMModule.CallOpts)
}

// KeyRemovalCharge is a free data retrieval call binding the contract method 0xd9df8c92.
//
// Solidity: function keyRemovalCharge() view returns(uint256)
func (_CSMModule *CSMModuleCallerSession) KeyRemovalCharge() (*big.Int, error) {
	return _CSMModule.Contract.KeyRemovalCharge(&_CSMModule.CallOpts)
}

// PublicRelease is a free data retrieval call binding the contract method 0xe21a430b.
//
// Solidity: function publicRelease() view returns(bool)
func (_CSMModule *CSMModuleCaller) PublicRelease(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "publicRelease")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// PublicRelease is a free data retrieval call binding the contract method 0xe21a430b.
//
// Solidity: function publicRelease() view returns(bool)
func (_CSMModule *CSMModuleSession) PublicRelease() (bool, error) {
	return _CSMModule.Contract.PublicRelease(&_CSMModule.CallOpts)
}

// PublicRelease is a free data retrieval call binding the contract method 0xe21a430b.
//
// Solidity: function publicRelease() view returns(bool)
func (_CSMModule *CSMModuleCallerSession) PublicRelease() (bool, error) {
	return _CSMModule.Contract.PublicRelease(&_CSMModule.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_CSMModule *CSMModuleCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _CSMModule.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_CSMModule *CSMModuleSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _CSMModule.Contract.SupportsInterface(&_CSMModule.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_CSMModule *CSMModuleCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _CSMModule.Contract.SupportsInterface(&_CSMModule.CallOpts, interfaceId)
}

// ActivatePublicRelease is a paid mutator transaction binding the contract method 0xd3931457.
//
// Solidity: function activatePublicRelease() returns()
func (_CSMModule *CSMModuleTransactor) ActivatePublicRelease(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "activatePublicRelease")
}

// ActivatePublicRelease is a paid mutator transaction binding the contract method 0xd3931457.
//
// Solidity: function activatePublicRelease() returns()
func (_CSMModule *CSMModuleSession) ActivatePublicRelease() (*types.Transaction, error) {
	return _CSMModule.Contract.ActivatePublicRelease(&_CSMModule.TransactOpts)
}

// ActivatePublicRelease is a paid mutator transaction binding the contract method 0xd3931457.
//
// Solidity: function activatePublicRelease() returns()
func (_CSMModule *CSMModuleTransactorSession) ActivatePublicRelease() (*types.Transaction, error) {
	return _CSMModule.Contract.ActivatePublicRelease(&_CSMModule.TransactOpts)
}

// AddNodeOperatorETH is a paid mutator transaction binding the contract method 0x157a039b.
//
// Solidity: function addNodeOperatorETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, bytes32[] eaProof, address referrer) payable returns()
func (_CSMModule *CSMModuleTransactor) AddNodeOperatorETH(opts *bind.TransactOpts, keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "addNodeOperatorETH", keysCount, publicKeys, signatures, managementProperties, eaProof, referrer)
}

// AddNodeOperatorETH is a paid mutator transaction binding the contract method 0x157a039b.
//
// Solidity: function addNodeOperatorETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, bytes32[] eaProof, address referrer) payable returns()
func (_CSMModule *CSMModuleSession) AddNodeOperatorETH(keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.AddNodeOperatorETH(&_CSMModule.TransactOpts, keysCount, publicKeys, signatures, managementProperties, eaProof, referrer)
}

// AddNodeOperatorETH is a paid mutator transaction binding the contract method 0x157a039b.
//
// Solidity: function addNodeOperatorETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, bytes32[] eaProof, address referrer) payable returns()
func (_CSMModule *CSMModuleTransactorSession) AddNodeOperatorETH(keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.AddNodeOperatorETH(&_CSMModule.TransactOpts, keysCount, publicKeys, signatures, managementProperties, eaProof, referrer)
}

// AddNodeOperatorStETH is a paid mutator transaction binding the contract method 0x6a5f2c4a.
//
// Solidity: function addNodeOperatorStETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, (uint256,uint256,uint8,bytes32,bytes32) permit, bytes32[] eaProof, address referrer) returns()
func (_CSMModule *CSMModuleTransactor) AddNodeOperatorStETH(opts *bind.TransactOpts, keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, permit ICSAccountingPermitInput, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "addNodeOperatorStETH", keysCount, publicKeys, signatures, managementProperties, permit, eaProof, referrer)
}

// AddNodeOperatorStETH is a paid mutator transaction binding the contract method 0x6a5f2c4a.
//
// Solidity: function addNodeOperatorStETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, (uint256,uint256,uint8,bytes32,bytes32) permit, bytes32[] eaProof, address referrer) returns()
func (_CSMModule *CSMModuleSession) AddNodeOperatorStETH(keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, permit ICSAccountingPermitInput, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.AddNodeOperatorStETH(&_CSMModule.TransactOpts, keysCount, publicKeys, signatures, managementProperties, permit, eaProof, referrer)
}

// AddNodeOperatorStETH is a paid mutator transaction binding the contract method 0x6a5f2c4a.
//
// Solidity: function addNodeOperatorStETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, (uint256,uint256,uint8,bytes32,bytes32) permit, bytes32[] eaProof, address referrer) returns()
func (_CSMModule *CSMModuleTransactorSession) AddNodeOperatorStETH(keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, permit ICSAccountingPermitInput, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.AddNodeOperatorStETH(&_CSMModule.TransactOpts, keysCount, publicKeys, signatures, managementProperties, permit, eaProof, referrer)
}

// AddNodeOperatorWstETH is a paid mutator transaction binding the contract method 0xacc446eb.
//
// Solidity: function addNodeOperatorWstETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, (uint256,uint256,uint8,bytes32,bytes32) permit, bytes32[] eaProof, address referrer) returns()
func (_CSMModule *CSMModuleTransactor) AddNodeOperatorWstETH(opts *bind.TransactOpts, keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, permit ICSAccountingPermitInput, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "addNodeOperatorWstETH", keysCount, publicKeys, signatures, managementProperties, permit, eaProof, referrer)
}

// AddNodeOperatorWstETH is a paid mutator transaction binding the contract method 0xacc446eb.
//
// Solidity: function addNodeOperatorWstETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, (uint256,uint256,uint8,bytes32,bytes32) permit, bytes32[] eaProof, address referrer) returns()
func (_CSMModule *CSMModuleSession) AddNodeOperatorWstETH(keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, permit ICSAccountingPermitInput, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.AddNodeOperatorWstETH(&_CSMModule.TransactOpts, keysCount, publicKeys, signatures, managementProperties, permit, eaProof, referrer)
}

// AddNodeOperatorWstETH is a paid mutator transaction binding the contract method 0xacc446eb.
//
// Solidity: function addNodeOperatorWstETH(uint256 keysCount, bytes publicKeys, bytes signatures, (address,address,bool) managementProperties, (uint256,uint256,uint8,bytes32,bytes32) permit, bytes32[] eaProof, address referrer) returns()
func (_CSMModule *CSMModuleTransactorSession) AddNodeOperatorWstETH(keysCount *big.Int, publicKeys []byte, signatures []byte, managementProperties NodeOperatorManagementProperties, permit ICSAccountingPermitInput, eaProof [][32]byte, referrer common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.AddNodeOperatorWstETH(&_CSMModule.TransactOpts, keysCount, publicKeys, signatures, managementProperties, permit, eaProof, referrer)
}

// AddValidatorKeysETH is a paid mutator transaction binding the contract method 0xfe7ed3cd.
//
// Solidity: function addValidatorKeysETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures) payable returns()
func (_CSMModule *CSMModuleTransactor) AddValidatorKeysETH(opts *bind.TransactOpts, nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "addValidatorKeysETH", nodeOperatorId, keysCount, publicKeys, signatures)
}

// AddValidatorKeysETH is a paid mutator transaction binding the contract method 0xfe7ed3cd.
//
// Solidity: function addValidatorKeysETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures) payable returns()
func (_CSMModule *CSMModuleSession) AddValidatorKeysETH(nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.AddValidatorKeysETH(&_CSMModule.TransactOpts, nodeOperatorId, keysCount, publicKeys, signatures)
}

// AddValidatorKeysETH is a paid mutator transaction binding the contract method 0xfe7ed3cd.
//
// Solidity: function addValidatorKeysETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures) payable returns()
func (_CSMModule *CSMModuleTransactorSession) AddValidatorKeysETH(nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.AddValidatorKeysETH(&_CSMModule.TransactOpts, nodeOperatorId, keysCount, publicKeys, signatures)
}

// AddValidatorKeysStETH is a paid mutator transaction binding the contract method 0x946654ad.
//
// Solidity: function addValidatorKeysStETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleTransactor) AddValidatorKeysStETH(opts *bind.TransactOpts, nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "addValidatorKeysStETH", nodeOperatorId, keysCount, publicKeys, signatures, permit)
}

// AddValidatorKeysStETH is a paid mutator transaction binding the contract method 0x946654ad.
//
// Solidity: function addValidatorKeysStETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleSession) AddValidatorKeysStETH(nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.Contract.AddValidatorKeysStETH(&_CSMModule.TransactOpts, nodeOperatorId, keysCount, publicKeys, signatures, permit)
}

// AddValidatorKeysStETH is a paid mutator transaction binding the contract method 0x946654ad.
//
// Solidity: function addValidatorKeysStETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleTransactorSession) AddValidatorKeysStETH(nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.Contract.AddValidatorKeysStETH(&_CSMModule.TransactOpts, nodeOperatorId, keysCount, publicKeys, signatures, permit)
}

// AddValidatorKeysWstETH is a paid mutator transaction binding the contract method 0x9ec3c24c.
//
// Solidity: function addValidatorKeysWstETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleTransactor) AddValidatorKeysWstETH(opts *bind.TransactOpts, nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "addValidatorKeysWstETH", nodeOperatorId, keysCount, publicKeys, signatures, permit)
}

// AddValidatorKeysWstETH is a paid mutator transaction binding the contract method 0x9ec3c24c.
//
// Solidity: function addValidatorKeysWstETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleSession) AddValidatorKeysWstETH(nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.Contract.AddValidatorKeysWstETH(&_CSMModule.TransactOpts, nodeOperatorId, keysCount, publicKeys, signatures, permit)
}

// AddValidatorKeysWstETH is a paid mutator transaction binding the contract method 0x9ec3c24c.
//
// Solidity: function addValidatorKeysWstETH(uint256 nodeOperatorId, uint256 keysCount, bytes publicKeys, bytes signatures, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleTransactorSession) AddValidatorKeysWstETH(nodeOperatorId *big.Int, keysCount *big.Int, publicKeys []byte, signatures []byte, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.Contract.AddValidatorKeysWstETH(&_CSMModule.TransactOpts, nodeOperatorId, keysCount, publicKeys, signatures, permit)
}

// CancelELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x40044801.
//
// Solidity: function cancelELRewardsStealingPenalty(uint256 nodeOperatorId, uint256 amount) returns()
func (_CSMModule *CSMModuleTransactor) CancelELRewardsStealingPenalty(opts *bind.TransactOpts, nodeOperatorId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "cancelELRewardsStealingPenalty", nodeOperatorId, amount)
}

// CancelELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x40044801.
//
// Solidity: function cancelELRewardsStealingPenalty(uint256 nodeOperatorId, uint256 amount) returns()
func (_CSMModule *CSMModuleSession) CancelELRewardsStealingPenalty(nodeOperatorId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.CancelELRewardsStealingPenalty(&_CSMModule.TransactOpts, nodeOperatorId, amount)
}

// CancelELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x40044801.
//
// Solidity: function cancelELRewardsStealingPenalty(uint256 nodeOperatorId, uint256 amount) returns()
func (_CSMModule *CSMModuleTransactorSession) CancelELRewardsStealingPenalty(nodeOperatorId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.CancelELRewardsStealingPenalty(&_CSMModule.TransactOpts, nodeOperatorId, amount)
}

// ChangeNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x75a401da.
//
// Solidity: function changeNodeOperatorRewardAddress(uint256 nodeOperatorId, address newAddress) returns()
func (_CSMModule *CSMModuleTransactor) ChangeNodeOperatorRewardAddress(opts *bind.TransactOpts, nodeOperatorId *big.Int, newAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "changeNodeOperatorRewardAddress", nodeOperatorId, newAddress)
}

// ChangeNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x75a401da.
//
// Solidity: function changeNodeOperatorRewardAddress(uint256 nodeOperatorId, address newAddress) returns()
func (_CSMModule *CSMModuleSession) ChangeNodeOperatorRewardAddress(nodeOperatorId *big.Int, newAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.ChangeNodeOperatorRewardAddress(&_CSMModule.TransactOpts, nodeOperatorId, newAddress)
}

// ChangeNodeOperatorRewardAddress is a paid mutator transaction binding the contract method 0x75a401da.
//
// Solidity: function changeNodeOperatorRewardAddress(uint256 nodeOperatorId, address newAddress) returns()
func (_CSMModule *CSMModuleTransactorSession) ChangeNodeOperatorRewardAddress(nodeOperatorId *big.Int, newAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.ChangeNodeOperatorRewardAddress(&_CSMModule.TransactOpts, nodeOperatorId, newAddress)
}

// ClaimRewardsStETH is a paid mutator transaction binding the contract method 0x8409d4fe.
//
// Solidity: function claimRewardsStETH(uint256 nodeOperatorId, uint256 stETHAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleTransactor) ClaimRewardsStETH(opts *bind.TransactOpts, nodeOperatorId *big.Int, stETHAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "claimRewardsStETH", nodeOperatorId, stETHAmount, cumulativeFeeShares, rewardsProof)
}

// ClaimRewardsStETH is a paid mutator transaction binding the contract method 0x8409d4fe.
//
// Solidity: function claimRewardsStETH(uint256 nodeOperatorId, uint256 stETHAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleSession) ClaimRewardsStETH(nodeOperatorId *big.Int, stETHAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.Contract.ClaimRewardsStETH(&_CSMModule.TransactOpts, nodeOperatorId, stETHAmount, cumulativeFeeShares, rewardsProof)
}

// ClaimRewardsStETH is a paid mutator transaction binding the contract method 0x8409d4fe.
//
// Solidity: function claimRewardsStETH(uint256 nodeOperatorId, uint256 stETHAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleTransactorSession) ClaimRewardsStETH(nodeOperatorId *big.Int, stETHAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.Contract.ClaimRewardsStETH(&_CSMModule.TransactOpts, nodeOperatorId, stETHAmount, cumulativeFeeShares, rewardsProof)
}

// ClaimRewardsUnstETH is a paid mutator transaction binding the contract method 0x3df6c438.
//
// Solidity: function claimRewardsUnstETH(uint256 nodeOperatorId, uint256 stEthAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleTransactor) ClaimRewardsUnstETH(opts *bind.TransactOpts, nodeOperatorId *big.Int, stEthAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "claimRewardsUnstETH", nodeOperatorId, stEthAmount, cumulativeFeeShares, rewardsProof)
}

// ClaimRewardsUnstETH is a paid mutator transaction binding the contract method 0x3df6c438.
//
// Solidity: function claimRewardsUnstETH(uint256 nodeOperatorId, uint256 stEthAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleSession) ClaimRewardsUnstETH(nodeOperatorId *big.Int, stEthAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.Contract.ClaimRewardsUnstETH(&_CSMModule.TransactOpts, nodeOperatorId, stEthAmount, cumulativeFeeShares, rewardsProof)
}

// ClaimRewardsUnstETH is a paid mutator transaction binding the contract method 0x3df6c438.
//
// Solidity: function claimRewardsUnstETH(uint256 nodeOperatorId, uint256 stEthAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleTransactorSession) ClaimRewardsUnstETH(nodeOperatorId *big.Int, stEthAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.Contract.ClaimRewardsUnstETH(&_CSMModule.TransactOpts, nodeOperatorId, stEthAmount, cumulativeFeeShares, rewardsProof)
}

// ClaimRewardsWstETH is a paid mutator transaction binding the contract method 0x5097ef59.
//
// Solidity: function claimRewardsWstETH(uint256 nodeOperatorId, uint256 wstETHAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleTransactor) ClaimRewardsWstETH(opts *bind.TransactOpts, nodeOperatorId *big.Int, wstETHAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "claimRewardsWstETH", nodeOperatorId, wstETHAmount, cumulativeFeeShares, rewardsProof)
}

// ClaimRewardsWstETH is a paid mutator transaction binding the contract method 0x5097ef59.
//
// Solidity: function claimRewardsWstETH(uint256 nodeOperatorId, uint256 wstETHAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleSession) ClaimRewardsWstETH(nodeOperatorId *big.Int, wstETHAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.Contract.ClaimRewardsWstETH(&_CSMModule.TransactOpts, nodeOperatorId, wstETHAmount, cumulativeFeeShares, rewardsProof)
}

// ClaimRewardsWstETH is a paid mutator transaction binding the contract method 0x5097ef59.
//
// Solidity: function claimRewardsWstETH(uint256 nodeOperatorId, uint256 wstETHAmount, uint256 cumulativeFeeShares, bytes32[] rewardsProof) returns()
func (_CSMModule *CSMModuleTransactorSession) ClaimRewardsWstETH(nodeOperatorId *big.Int, wstETHAmount *big.Int, cumulativeFeeShares *big.Int, rewardsProof [][32]byte) (*types.Transaction, error) {
	return _CSMModule.Contract.ClaimRewardsWstETH(&_CSMModule.TransactOpts, nodeOperatorId, wstETHAmount, cumulativeFeeShares, rewardsProof)
}

// CleanDepositQueue is a paid mutator transaction binding the contract method 0x735dfa28.
//
// Solidity: function cleanDepositQueue(uint256 maxItems) returns(uint256 removed, uint256 lastRemovedAtDepth)
func (_CSMModule *CSMModuleTransactor) CleanDepositQueue(opts *bind.TransactOpts, maxItems *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "cleanDepositQueue", maxItems)
}

// CleanDepositQueue is a paid mutator transaction binding the contract method 0x735dfa28.
//
// Solidity: function cleanDepositQueue(uint256 maxItems) returns(uint256 removed, uint256 lastRemovedAtDepth)
func (_CSMModule *CSMModuleSession) CleanDepositQueue(maxItems *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.CleanDepositQueue(&_CSMModule.TransactOpts, maxItems)
}

// CleanDepositQueue is a paid mutator transaction binding the contract method 0x735dfa28.
//
// Solidity: function cleanDepositQueue(uint256 maxItems) returns(uint256 removed, uint256 lastRemovedAtDepth)
func (_CSMModule *CSMModuleTransactorSession) CleanDepositQueue(maxItems *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.CleanDepositQueue(&_CSMModule.TransactOpts, maxItems)
}

// CompensateELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x6efe37a2.
//
// Solidity: function compensateELRewardsStealingPenalty(uint256 nodeOperatorId) payable returns()
func (_CSMModule *CSMModuleTransactor) CompensateELRewardsStealingPenalty(opts *bind.TransactOpts, nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "compensateELRewardsStealingPenalty", nodeOperatorId)
}

// CompensateELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x6efe37a2.
//
// Solidity: function compensateELRewardsStealingPenalty(uint256 nodeOperatorId) payable returns()
func (_CSMModule *CSMModuleSession) CompensateELRewardsStealingPenalty(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.CompensateELRewardsStealingPenalty(&_CSMModule.TransactOpts, nodeOperatorId)
}

// CompensateELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x6efe37a2.
//
// Solidity: function compensateELRewardsStealingPenalty(uint256 nodeOperatorId) payable returns()
func (_CSMModule *CSMModuleTransactorSession) CompensateELRewardsStealingPenalty(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.CompensateELRewardsStealingPenalty(&_CSMModule.TransactOpts, nodeOperatorId)
}

// ConfirmNodeOperatorManagerAddressChange is a paid mutator transaction binding the contract method 0x6bb1bfdf.
//
// Solidity: function confirmNodeOperatorManagerAddressChange(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleTransactor) ConfirmNodeOperatorManagerAddressChange(opts *bind.TransactOpts, nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "confirmNodeOperatorManagerAddressChange", nodeOperatorId)
}

// ConfirmNodeOperatorManagerAddressChange is a paid mutator transaction binding the contract method 0x6bb1bfdf.
//
// Solidity: function confirmNodeOperatorManagerAddressChange(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleSession) ConfirmNodeOperatorManagerAddressChange(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.ConfirmNodeOperatorManagerAddressChange(&_CSMModule.TransactOpts, nodeOperatorId)
}

// ConfirmNodeOperatorManagerAddressChange is a paid mutator transaction binding the contract method 0x6bb1bfdf.
//
// Solidity: function confirmNodeOperatorManagerAddressChange(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleTransactorSession) ConfirmNodeOperatorManagerAddressChange(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.ConfirmNodeOperatorManagerAddressChange(&_CSMModule.TransactOpts, nodeOperatorId)
}

// ConfirmNodeOperatorRewardAddressChange is a paid mutator transaction binding the contract method 0x5204281c.
//
// Solidity: function confirmNodeOperatorRewardAddressChange(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleTransactor) ConfirmNodeOperatorRewardAddressChange(opts *bind.TransactOpts, nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "confirmNodeOperatorRewardAddressChange", nodeOperatorId)
}

// ConfirmNodeOperatorRewardAddressChange is a paid mutator transaction binding the contract method 0x5204281c.
//
// Solidity: function confirmNodeOperatorRewardAddressChange(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleSession) ConfirmNodeOperatorRewardAddressChange(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.ConfirmNodeOperatorRewardAddressChange(&_CSMModule.TransactOpts, nodeOperatorId)
}

// ConfirmNodeOperatorRewardAddressChange is a paid mutator transaction binding the contract method 0x5204281c.
//
// Solidity: function confirmNodeOperatorRewardAddressChange(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleTransactorSession) ConfirmNodeOperatorRewardAddressChange(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.ConfirmNodeOperatorRewardAddressChange(&_CSMModule.TransactOpts, nodeOperatorId)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes nodeOperatorIds, bytes vettedSigningKeysCounts) returns()
func (_CSMModule *CSMModuleTransactor) DecreaseVettedSigningKeysCount(opts *bind.TransactOpts, nodeOperatorIds []byte, vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "decreaseVettedSigningKeysCount", nodeOperatorIds, vettedSigningKeysCounts)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes nodeOperatorIds, bytes vettedSigningKeysCounts) returns()
func (_CSMModule *CSMModuleSession) DecreaseVettedSigningKeysCount(nodeOperatorIds []byte, vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.DecreaseVettedSigningKeysCount(&_CSMModule.TransactOpts, nodeOperatorIds, vettedSigningKeysCounts)
}

// DecreaseVettedSigningKeysCount is a paid mutator transaction binding the contract method 0xb643189b.
//
// Solidity: function decreaseVettedSigningKeysCount(bytes nodeOperatorIds, bytes vettedSigningKeysCounts) returns()
func (_CSMModule *CSMModuleTransactorSession) DecreaseVettedSigningKeysCount(nodeOperatorIds []byte, vettedSigningKeysCounts []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.DecreaseVettedSigningKeysCount(&_CSMModule.TransactOpts, nodeOperatorIds, vettedSigningKeysCounts)
}

// DepositETH is a paid mutator transaction binding the contract method 0x5358fbda.
//
// Solidity: function depositETH(uint256 nodeOperatorId) payable returns()
func (_CSMModule *CSMModuleTransactor) DepositETH(opts *bind.TransactOpts, nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "depositETH", nodeOperatorId)
}

// DepositETH is a paid mutator transaction binding the contract method 0x5358fbda.
//
// Solidity: function depositETH(uint256 nodeOperatorId) payable returns()
func (_CSMModule *CSMModuleSession) DepositETH(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.DepositETH(&_CSMModule.TransactOpts, nodeOperatorId)
}

// DepositETH is a paid mutator transaction binding the contract method 0x5358fbda.
//
// Solidity: function depositETH(uint256 nodeOperatorId) payable returns()
func (_CSMModule *CSMModuleTransactorSession) DepositETH(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.DepositETH(&_CSMModule.TransactOpts, nodeOperatorId)
}

// DepositStETH is a paid mutator transaction binding the contract method 0xe1aa105d.
//
// Solidity: function depositStETH(uint256 nodeOperatorId, uint256 stETHAmount, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleTransactor) DepositStETH(opts *bind.TransactOpts, nodeOperatorId *big.Int, stETHAmount *big.Int, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "depositStETH", nodeOperatorId, stETHAmount, permit)
}

// DepositStETH is a paid mutator transaction binding the contract method 0xe1aa105d.
//
// Solidity: function depositStETH(uint256 nodeOperatorId, uint256 stETHAmount, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleSession) DepositStETH(nodeOperatorId *big.Int, stETHAmount *big.Int, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.Contract.DepositStETH(&_CSMModule.TransactOpts, nodeOperatorId, stETHAmount, permit)
}

// DepositStETH is a paid mutator transaction binding the contract method 0xe1aa105d.
//
// Solidity: function depositStETH(uint256 nodeOperatorId, uint256 stETHAmount, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleTransactorSession) DepositStETH(nodeOperatorId *big.Int, stETHAmount *big.Int, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.Contract.DepositStETH(&_CSMModule.TransactOpts, nodeOperatorId, stETHAmount, permit)
}

// DepositWstETH is a paid mutator transaction binding the contract method 0x3f214bb2.
//
// Solidity: function depositWstETH(uint256 nodeOperatorId, uint256 wstETHAmount, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleTransactor) DepositWstETH(opts *bind.TransactOpts, nodeOperatorId *big.Int, wstETHAmount *big.Int, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "depositWstETH", nodeOperatorId, wstETHAmount, permit)
}

// DepositWstETH is a paid mutator transaction binding the contract method 0x3f214bb2.
//
// Solidity: function depositWstETH(uint256 nodeOperatorId, uint256 wstETHAmount, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleSession) DepositWstETH(nodeOperatorId *big.Int, wstETHAmount *big.Int, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.Contract.DepositWstETH(&_CSMModule.TransactOpts, nodeOperatorId, wstETHAmount, permit)
}

// DepositWstETH is a paid mutator transaction binding the contract method 0x3f214bb2.
//
// Solidity: function depositWstETH(uint256 nodeOperatorId, uint256 wstETHAmount, (uint256,uint256,uint8,bytes32,bytes32) permit) returns()
func (_CSMModule *CSMModuleTransactorSession) DepositWstETH(nodeOperatorId *big.Int, wstETHAmount *big.Int, permit ICSAccountingPermitInput) (*types.Transaction, error) {
	return _CSMModule.Contract.DepositWstETH(&_CSMModule.TransactOpts, nodeOperatorId, wstETHAmount, permit)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_CSMModule *CSMModuleTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_CSMModule *CSMModuleSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.GrantRole(&_CSMModule.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_CSMModule *CSMModuleTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.GrantRole(&_CSMModule.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0xbe203094.
//
// Solidity: function initialize(address _accounting, address _earlyAdoption, uint256 _keyRemovalCharge, address admin) returns()
func (_CSMModule *CSMModuleTransactor) Initialize(opts *bind.TransactOpts, _accounting common.Address, _earlyAdoption common.Address, _keyRemovalCharge *big.Int, admin common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "initialize", _accounting, _earlyAdoption, _keyRemovalCharge, admin)
}

// Initialize is a paid mutator transaction binding the contract method 0xbe203094.
//
// Solidity: function initialize(address _accounting, address _earlyAdoption, uint256 _keyRemovalCharge, address admin) returns()
func (_CSMModule *CSMModuleSession) Initialize(_accounting common.Address, _earlyAdoption common.Address, _keyRemovalCharge *big.Int, admin common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.Initialize(&_CSMModule.TransactOpts, _accounting, _earlyAdoption, _keyRemovalCharge, admin)
}

// Initialize is a paid mutator transaction binding the contract method 0xbe203094.
//
// Solidity: function initialize(address _accounting, address _earlyAdoption, uint256 _keyRemovalCharge, address admin) returns()
func (_CSMModule *CSMModuleTransactorSession) Initialize(_accounting common.Address, _earlyAdoption common.Address, _keyRemovalCharge *big.Int, admin common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.Initialize(&_CSMModule.TransactOpts, _accounting, _earlyAdoption, _keyRemovalCharge, admin)
}

// NormalizeQueue is a paid mutator transaction binding the contract method 0xb1520dc5.
//
// Solidity: function normalizeQueue(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleTransactor) NormalizeQueue(opts *bind.TransactOpts, nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "normalizeQueue", nodeOperatorId)
}

// NormalizeQueue is a paid mutator transaction binding the contract method 0xb1520dc5.
//
// Solidity: function normalizeQueue(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleSession) NormalizeQueue(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.NormalizeQueue(&_CSMModule.TransactOpts, nodeOperatorId)
}

// NormalizeQueue is a paid mutator transaction binding the contract method 0xb1520dc5.
//
// Solidity: function normalizeQueue(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleTransactorSession) NormalizeQueue(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.NormalizeQueue(&_CSMModule.TransactOpts, nodeOperatorId)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_CSMModule *CSMModuleTransactor) ObtainDepositData(opts *bind.TransactOpts, depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "obtainDepositData", depositsCount, arg1)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_CSMModule *CSMModuleSession) ObtainDepositData(depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.ObtainDepositData(&_CSMModule.TransactOpts, depositsCount, arg1)
}

// ObtainDepositData is a paid mutator transaction binding the contract method 0xbee41b58.
//
// Solidity: function obtainDepositData(uint256 depositsCount, bytes ) returns(bytes publicKeys, bytes signatures)
func (_CSMModule *CSMModuleTransactorSession) ObtainDepositData(depositsCount *big.Int, arg1 []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.ObtainDepositData(&_CSMModule.TransactOpts, depositsCount, arg1)
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_CSMModule *CSMModuleTransactor) OnExitedAndStuckValidatorsCountsUpdated(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "onExitedAndStuckValidatorsCountsUpdated")
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_CSMModule *CSMModuleSession) OnExitedAndStuckValidatorsCountsUpdated() (*types.Transaction, error) {
	return _CSMModule.Contract.OnExitedAndStuckValidatorsCountsUpdated(&_CSMModule.TransactOpts)
}

// OnExitedAndStuckValidatorsCountsUpdated is a paid mutator transaction binding the contract method 0xe864299e.
//
// Solidity: function onExitedAndStuckValidatorsCountsUpdated() returns()
func (_CSMModule *CSMModuleTransactorSession) OnExitedAndStuckValidatorsCountsUpdated() (*types.Transaction, error) {
	return _CSMModule.Contract.OnExitedAndStuckValidatorsCountsUpdated(&_CSMModule.TransactOpts)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 totalShares) returns()
func (_CSMModule *CSMModuleTransactor) OnRewardsMinted(opts *bind.TransactOpts, totalShares *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "onRewardsMinted", totalShares)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 totalShares) returns()
func (_CSMModule *CSMModuleSession) OnRewardsMinted(totalShares *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.OnRewardsMinted(&_CSMModule.TransactOpts, totalShares)
}

// OnRewardsMinted is a paid mutator transaction binding the contract method 0x8d7e4017.
//
// Solidity: function onRewardsMinted(uint256 totalShares) returns()
func (_CSMModule *CSMModuleTransactorSession) OnRewardsMinted(totalShares *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.OnRewardsMinted(&_CSMModule.TransactOpts, totalShares)
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_CSMModule *CSMModuleTransactor) OnWithdrawalCredentialsChanged(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "onWithdrawalCredentialsChanged")
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_CSMModule *CSMModuleSession) OnWithdrawalCredentialsChanged() (*types.Transaction, error) {
	return _CSMModule.Contract.OnWithdrawalCredentialsChanged(&_CSMModule.TransactOpts)
}

// OnWithdrawalCredentialsChanged is a paid mutator transaction binding the contract method 0x90c09bdb.
//
// Solidity: function onWithdrawalCredentialsChanged() returns()
func (_CSMModule *CSMModuleTransactorSession) OnWithdrawalCredentialsChanged() (*types.Transaction, error) {
	return _CSMModule.Contract.OnWithdrawalCredentialsChanged(&_CSMModule.TransactOpts)
}

// PauseFor is a paid mutator transaction binding the contract method 0xf3f449c7.
//
// Solidity: function pauseFor(uint256 duration) returns()
func (_CSMModule *CSMModuleTransactor) PauseFor(opts *bind.TransactOpts, duration *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "pauseFor", duration)
}

// PauseFor is a paid mutator transaction binding the contract method 0xf3f449c7.
//
// Solidity: function pauseFor(uint256 duration) returns()
func (_CSMModule *CSMModuleSession) PauseFor(duration *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.PauseFor(&_CSMModule.TransactOpts, duration)
}

// PauseFor is a paid mutator transaction binding the contract method 0xf3f449c7.
//
// Solidity: function pauseFor(uint256 duration) returns()
func (_CSMModule *CSMModuleTransactorSession) PauseFor(duration *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.PauseFor(&_CSMModule.TransactOpts, duration)
}

// ProposeNodeOperatorManagerAddressChange is a paid mutator transaction binding the contract method 0x8cabe959.
//
// Solidity: function proposeNodeOperatorManagerAddressChange(uint256 nodeOperatorId, address proposedAddress) returns()
func (_CSMModule *CSMModuleTransactor) ProposeNodeOperatorManagerAddressChange(opts *bind.TransactOpts, nodeOperatorId *big.Int, proposedAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "proposeNodeOperatorManagerAddressChange", nodeOperatorId, proposedAddress)
}

// ProposeNodeOperatorManagerAddressChange is a paid mutator transaction binding the contract method 0x8cabe959.
//
// Solidity: function proposeNodeOperatorManagerAddressChange(uint256 nodeOperatorId, address proposedAddress) returns()
func (_CSMModule *CSMModuleSession) ProposeNodeOperatorManagerAddressChange(nodeOperatorId *big.Int, proposedAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.ProposeNodeOperatorManagerAddressChange(&_CSMModule.TransactOpts, nodeOperatorId, proposedAddress)
}

// ProposeNodeOperatorManagerAddressChange is a paid mutator transaction binding the contract method 0x8cabe959.
//
// Solidity: function proposeNodeOperatorManagerAddressChange(uint256 nodeOperatorId, address proposedAddress) returns()
func (_CSMModule *CSMModuleTransactorSession) ProposeNodeOperatorManagerAddressChange(nodeOperatorId *big.Int, proposedAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.ProposeNodeOperatorManagerAddressChange(&_CSMModule.TransactOpts, nodeOperatorId, proposedAddress)
}

// ProposeNodeOperatorRewardAddressChange is a paid mutator transaction binding the contract method 0x1b40b231.
//
// Solidity: function proposeNodeOperatorRewardAddressChange(uint256 nodeOperatorId, address proposedAddress) returns()
func (_CSMModule *CSMModuleTransactor) ProposeNodeOperatorRewardAddressChange(opts *bind.TransactOpts, nodeOperatorId *big.Int, proposedAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "proposeNodeOperatorRewardAddressChange", nodeOperatorId, proposedAddress)
}

// ProposeNodeOperatorRewardAddressChange is a paid mutator transaction binding the contract method 0x1b40b231.
//
// Solidity: function proposeNodeOperatorRewardAddressChange(uint256 nodeOperatorId, address proposedAddress) returns()
func (_CSMModule *CSMModuleSession) ProposeNodeOperatorRewardAddressChange(nodeOperatorId *big.Int, proposedAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.ProposeNodeOperatorRewardAddressChange(&_CSMModule.TransactOpts, nodeOperatorId, proposedAddress)
}

// ProposeNodeOperatorRewardAddressChange is a paid mutator transaction binding the contract method 0x1b40b231.
//
// Solidity: function proposeNodeOperatorRewardAddressChange(uint256 nodeOperatorId, address proposedAddress) returns()
func (_CSMModule *CSMModuleTransactorSession) ProposeNodeOperatorRewardAddressChange(nodeOperatorId *big.Int, proposedAddress common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.ProposeNodeOperatorRewardAddressChange(&_CSMModule.TransactOpts, nodeOperatorId, proposedAddress)
}

// RecoverERC1155 is a paid mutator transaction binding the contract method 0x5c654ad9.
//
// Solidity: function recoverERC1155(address token, uint256 tokenId) returns()
func (_CSMModule *CSMModuleTransactor) RecoverERC1155(opts *bind.TransactOpts, token common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "recoverERC1155", token, tokenId)
}

// RecoverERC1155 is a paid mutator transaction binding the contract method 0x5c654ad9.
//
// Solidity: function recoverERC1155(address token, uint256 tokenId) returns()
func (_CSMModule *CSMModuleSession) RecoverERC1155(token common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.RecoverERC1155(&_CSMModule.TransactOpts, token, tokenId)
}

// RecoverERC1155 is a paid mutator transaction binding the contract method 0x5c654ad9.
//
// Solidity: function recoverERC1155(address token, uint256 tokenId) returns()
func (_CSMModule *CSMModuleTransactorSession) RecoverERC1155(token common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.RecoverERC1155(&_CSMModule.TransactOpts, token, tokenId)
}

// RecoverERC20 is a paid mutator transaction binding the contract method 0x8980f11f.
//
// Solidity: function recoverERC20(address token, uint256 amount) returns()
func (_CSMModule *CSMModuleTransactor) RecoverERC20(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "recoverERC20", token, amount)
}

// RecoverERC20 is a paid mutator transaction binding the contract method 0x8980f11f.
//
// Solidity: function recoverERC20(address token, uint256 amount) returns()
func (_CSMModule *CSMModuleSession) RecoverERC20(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.RecoverERC20(&_CSMModule.TransactOpts, token, amount)
}

// RecoverERC20 is a paid mutator transaction binding the contract method 0x8980f11f.
//
// Solidity: function recoverERC20(address token, uint256 amount) returns()
func (_CSMModule *CSMModuleTransactorSession) RecoverERC20(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.RecoverERC20(&_CSMModule.TransactOpts, token, amount)
}

// RecoverERC721 is a paid mutator transaction binding the contract method 0x819d4cc6.
//
// Solidity: function recoverERC721(address token, uint256 tokenId) returns()
func (_CSMModule *CSMModuleTransactor) RecoverERC721(opts *bind.TransactOpts, token common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "recoverERC721", token, tokenId)
}

// RecoverERC721 is a paid mutator transaction binding the contract method 0x819d4cc6.
//
// Solidity: function recoverERC721(address token, uint256 tokenId) returns()
func (_CSMModule *CSMModuleSession) RecoverERC721(token common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.RecoverERC721(&_CSMModule.TransactOpts, token, tokenId)
}

// RecoverERC721 is a paid mutator transaction binding the contract method 0x819d4cc6.
//
// Solidity: function recoverERC721(address token, uint256 tokenId) returns()
func (_CSMModule *CSMModuleTransactorSession) RecoverERC721(token common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.RecoverERC721(&_CSMModule.TransactOpts, token, tokenId)
}

// RecoverEther is a paid mutator transaction binding the contract method 0x52d8bfc2.
//
// Solidity: function recoverEther() returns()
func (_CSMModule *CSMModuleTransactor) RecoverEther(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "recoverEther")
}

// RecoverEther is a paid mutator transaction binding the contract method 0x52d8bfc2.
//
// Solidity: function recoverEther() returns()
func (_CSMModule *CSMModuleSession) RecoverEther() (*types.Transaction, error) {
	return _CSMModule.Contract.RecoverEther(&_CSMModule.TransactOpts)
}

// RecoverEther is a paid mutator transaction binding the contract method 0x52d8bfc2.
//
// Solidity: function recoverEther() returns()
func (_CSMModule *CSMModuleTransactorSession) RecoverEther() (*types.Transaction, error) {
	return _CSMModule.Contract.RecoverEther(&_CSMModule.TransactOpts)
}

// RemoveKeys is a paid mutator transaction binding the contract method 0x8b3ac71d.
//
// Solidity: function removeKeys(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) returns()
func (_CSMModule *CSMModuleTransactor) RemoveKeys(opts *bind.TransactOpts, nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "removeKeys", nodeOperatorId, startIndex, keysCount)
}

// RemoveKeys is a paid mutator transaction binding the contract method 0x8b3ac71d.
//
// Solidity: function removeKeys(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) returns()
func (_CSMModule *CSMModuleSession) RemoveKeys(nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.RemoveKeys(&_CSMModule.TransactOpts, nodeOperatorId, startIndex, keysCount)
}

// RemoveKeys is a paid mutator transaction binding the contract method 0x8b3ac71d.
//
// Solidity: function removeKeys(uint256 nodeOperatorId, uint256 startIndex, uint256 keysCount) returns()
func (_CSMModule *CSMModuleTransactorSession) RemoveKeys(nodeOperatorId *big.Int, startIndex *big.Int, keysCount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.RemoveKeys(&_CSMModule.TransactOpts, nodeOperatorId, startIndex, keysCount)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_CSMModule *CSMModuleTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_CSMModule *CSMModuleSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.RenounceRole(&_CSMModule.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_CSMModule *CSMModuleTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.RenounceRole(&_CSMModule.TransactOpts, role, callerConfirmation)
}

// ReportELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x388dd1d1.
//
// Solidity: function reportELRewardsStealingPenalty(uint256 nodeOperatorId, bytes32 blockHash, uint256 amount) returns()
func (_CSMModule *CSMModuleTransactor) ReportELRewardsStealingPenalty(opts *bind.TransactOpts, nodeOperatorId *big.Int, blockHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "reportELRewardsStealingPenalty", nodeOperatorId, blockHash, amount)
}

// ReportELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x388dd1d1.
//
// Solidity: function reportELRewardsStealingPenalty(uint256 nodeOperatorId, bytes32 blockHash, uint256 amount) returns()
func (_CSMModule *CSMModuleSession) ReportELRewardsStealingPenalty(nodeOperatorId *big.Int, blockHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.ReportELRewardsStealingPenalty(&_CSMModule.TransactOpts, nodeOperatorId, blockHash, amount)
}

// ReportELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x388dd1d1.
//
// Solidity: function reportELRewardsStealingPenalty(uint256 nodeOperatorId, bytes32 blockHash, uint256 amount) returns()
func (_CSMModule *CSMModuleTransactorSession) ReportELRewardsStealingPenalty(nodeOperatorId *big.Int, blockHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.ReportELRewardsStealingPenalty(&_CSMModule.TransactOpts, nodeOperatorId, blockHash, amount)
}

// ResetNodeOperatorManagerAddress is a paid mutator transaction binding the contract method 0x6a6304cc.
//
// Solidity: function resetNodeOperatorManagerAddress(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleTransactor) ResetNodeOperatorManagerAddress(opts *bind.TransactOpts, nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "resetNodeOperatorManagerAddress", nodeOperatorId)
}

// ResetNodeOperatorManagerAddress is a paid mutator transaction binding the contract method 0x6a6304cc.
//
// Solidity: function resetNodeOperatorManagerAddress(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleSession) ResetNodeOperatorManagerAddress(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.ResetNodeOperatorManagerAddress(&_CSMModule.TransactOpts, nodeOperatorId)
}

// ResetNodeOperatorManagerAddress is a paid mutator transaction binding the contract method 0x6a6304cc.
//
// Solidity: function resetNodeOperatorManagerAddress(uint256 nodeOperatorId) returns()
func (_CSMModule *CSMModuleTransactorSession) ResetNodeOperatorManagerAddress(nodeOperatorId *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.ResetNodeOperatorManagerAddress(&_CSMModule.TransactOpts, nodeOperatorId)
}

// Resume is a paid mutator transaction binding the contract method 0x046f7da2.
//
// Solidity: function resume() returns()
func (_CSMModule *CSMModuleTransactor) Resume(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "resume")
}

// Resume is a paid mutator transaction binding the contract method 0x046f7da2.
//
// Solidity: function resume() returns()
func (_CSMModule *CSMModuleSession) Resume() (*types.Transaction, error) {
	return _CSMModule.Contract.Resume(&_CSMModule.TransactOpts)
}

// Resume is a paid mutator transaction binding the contract method 0x046f7da2.
//
// Solidity: function resume() returns()
func (_CSMModule *CSMModuleTransactorSession) Resume() (*types.Transaction, error) {
	return _CSMModule.Contract.Resume(&_CSMModule.TransactOpts)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_CSMModule *CSMModuleTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_CSMModule *CSMModuleSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.RevokeRole(&_CSMModule.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_CSMModule *CSMModuleTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _CSMModule.Contract.RevokeRole(&_CSMModule.TransactOpts, role, account)
}

// SetKeyRemovalCharge is a paid mutator transaction binding the contract method 0xba1557ae.
//
// Solidity: function setKeyRemovalCharge(uint256 amount) returns()
func (_CSMModule *CSMModuleTransactor) SetKeyRemovalCharge(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "setKeyRemovalCharge", amount)
}

// SetKeyRemovalCharge is a paid mutator transaction binding the contract method 0xba1557ae.
//
// Solidity: function setKeyRemovalCharge(uint256 amount) returns()
func (_CSMModule *CSMModuleSession) SetKeyRemovalCharge(amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.SetKeyRemovalCharge(&_CSMModule.TransactOpts, amount)
}

// SetKeyRemovalCharge is a paid mutator transaction binding the contract method 0xba1557ae.
//
// Solidity: function setKeyRemovalCharge(uint256 amount) returns()
func (_CSMModule *CSMModuleTransactorSession) SetKeyRemovalCharge(amount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.SetKeyRemovalCharge(&_CSMModule.TransactOpts, amount)
}

// SettleELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x37b12b5f.
//
// Solidity: function settleELRewardsStealingPenalty(uint256[] nodeOperatorIds) returns()
func (_CSMModule *CSMModuleTransactor) SettleELRewardsStealingPenalty(opts *bind.TransactOpts, nodeOperatorIds []*big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "settleELRewardsStealingPenalty", nodeOperatorIds)
}

// SettleELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x37b12b5f.
//
// Solidity: function settleELRewardsStealingPenalty(uint256[] nodeOperatorIds) returns()
func (_CSMModule *CSMModuleSession) SettleELRewardsStealingPenalty(nodeOperatorIds []*big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.SettleELRewardsStealingPenalty(&_CSMModule.TransactOpts, nodeOperatorIds)
}

// SettleELRewardsStealingPenalty is a paid mutator transaction binding the contract method 0x37b12b5f.
//
// Solidity: function settleELRewardsStealingPenalty(uint256[] nodeOperatorIds) returns()
func (_CSMModule *CSMModuleTransactorSession) SettleELRewardsStealingPenalty(nodeOperatorIds []*big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.SettleELRewardsStealingPenalty(&_CSMModule.TransactOpts, nodeOperatorIds)
}

// SubmitInitialSlashing is a paid mutator transaction binding the contract method 0xf96d3952.
//
// Solidity: function submitInitialSlashing(uint256 nodeOperatorId, uint256 keyIndex) returns()
func (_CSMModule *CSMModuleTransactor) SubmitInitialSlashing(opts *bind.TransactOpts, nodeOperatorId *big.Int, keyIndex *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "submitInitialSlashing", nodeOperatorId, keyIndex)
}

// SubmitInitialSlashing is a paid mutator transaction binding the contract method 0xf96d3952.
//
// Solidity: function submitInitialSlashing(uint256 nodeOperatorId, uint256 keyIndex) returns()
func (_CSMModule *CSMModuleSession) SubmitInitialSlashing(nodeOperatorId *big.Int, keyIndex *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.SubmitInitialSlashing(&_CSMModule.TransactOpts, nodeOperatorId, keyIndex)
}

// SubmitInitialSlashing is a paid mutator transaction binding the contract method 0xf96d3952.
//
// Solidity: function submitInitialSlashing(uint256 nodeOperatorId, uint256 keyIndex) returns()
func (_CSMModule *CSMModuleTransactorSession) SubmitInitialSlashing(nodeOperatorId *big.Int, keyIndex *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.SubmitInitialSlashing(&_CSMModule.TransactOpts, nodeOperatorId, keyIndex)
}

// SubmitWithdrawal is a paid mutator transaction binding the contract method 0xf408b551.
//
// Solidity: function submitWithdrawal(uint256 nodeOperatorId, uint256 keyIndex, uint256 amount, bool isSlashed) returns()
func (_CSMModule *CSMModuleTransactor) SubmitWithdrawal(opts *bind.TransactOpts, nodeOperatorId *big.Int, keyIndex *big.Int, amount *big.Int, isSlashed bool) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "submitWithdrawal", nodeOperatorId, keyIndex, amount, isSlashed)
}

// SubmitWithdrawal is a paid mutator transaction binding the contract method 0xf408b551.
//
// Solidity: function submitWithdrawal(uint256 nodeOperatorId, uint256 keyIndex, uint256 amount, bool isSlashed) returns()
func (_CSMModule *CSMModuleSession) SubmitWithdrawal(nodeOperatorId *big.Int, keyIndex *big.Int, amount *big.Int, isSlashed bool) (*types.Transaction, error) {
	return _CSMModule.Contract.SubmitWithdrawal(&_CSMModule.TransactOpts, nodeOperatorId, keyIndex, amount, isSlashed)
}

// SubmitWithdrawal is a paid mutator transaction binding the contract method 0xf408b551.
//
// Solidity: function submitWithdrawal(uint256 nodeOperatorId, uint256 keyIndex, uint256 amount, bool isSlashed) returns()
func (_CSMModule *CSMModuleTransactorSession) SubmitWithdrawal(nodeOperatorId *big.Int, keyIndex *big.Int, amount *big.Int, isSlashed bool) (*types.Transaction, error) {
	return _CSMModule.Contract.SubmitWithdrawal(&_CSMModule.TransactOpts, nodeOperatorId, keyIndex, amount, isSlashed)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0xf2e2ca63.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 nodeOperatorId, uint256 exitedValidatorsKeysCount, uint256 stuckValidatorsKeysCount) returns()
func (_CSMModule *CSMModuleTransactor) UnsafeUpdateValidatorsCount(opts *bind.TransactOpts, nodeOperatorId *big.Int, exitedValidatorsKeysCount *big.Int, stuckValidatorsKeysCount *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "unsafeUpdateValidatorsCount", nodeOperatorId, exitedValidatorsKeysCount, stuckValidatorsKeysCount)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0xf2e2ca63.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 nodeOperatorId, uint256 exitedValidatorsKeysCount, uint256 stuckValidatorsKeysCount) returns()
func (_CSMModule *CSMModuleSession) UnsafeUpdateValidatorsCount(nodeOperatorId *big.Int, exitedValidatorsKeysCount *big.Int, stuckValidatorsKeysCount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.UnsafeUpdateValidatorsCount(&_CSMModule.TransactOpts, nodeOperatorId, exitedValidatorsKeysCount, stuckValidatorsKeysCount)
}

// UnsafeUpdateValidatorsCount is a paid mutator transaction binding the contract method 0xf2e2ca63.
//
// Solidity: function unsafeUpdateValidatorsCount(uint256 nodeOperatorId, uint256 exitedValidatorsKeysCount, uint256 stuckValidatorsKeysCount) returns()
func (_CSMModule *CSMModuleTransactorSession) UnsafeUpdateValidatorsCount(nodeOperatorId *big.Int, exitedValidatorsKeysCount *big.Int, stuckValidatorsKeysCount *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.UnsafeUpdateValidatorsCount(&_CSMModule.TransactOpts, nodeOperatorId, exitedValidatorsKeysCount, stuckValidatorsKeysCount)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes nodeOperatorIds, bytes exitedValidatorsCounts) returns()
func (_CSMModule *CSMModuleTransactor) UpdateExitedValidatorsCount(opts *bind.TransactOpts, nodeOperatorIds []byte, exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "updateExitedValidatorsCount", nodeOperatorIds, exitedValidatorsCounts)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes nodeOperatorIds, bytes exitedValidatorsCounts) returns()
func (_CSMModule *CSMModuleSession) UpdateExitedValidatorsCount(nodeOperatorIds []byte, exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.UpdateExitedValidatorsCount(&_CSMModule.TransactOpts, nodeOperatorIds, exitedValidatorsCounts)
}

// UpdateExitedValidatorsCount is a paid mutator transaction binding the contract method 0x9b00c146.
//
// Solidity: function updateExitedValidatorsCount(bytes nodeOperatorIds, bytes exitedValidatorsCounts) returns()
func (_CSMModule *CSMModuleTransactorSession) UpdateExitedValidatorsCount(nodeOperatorIds []byte, exitedValidatorsCounts []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.UpdateExitedValidatorsCount(&_CSMModule.TransactOpts, nodeOperatorIds, exitedValidatorsCounts)
}

// UpdateRefundedValidatorsCount is a paid mutator transaction binding the contract method 0xa2e080f1.
//
// Solidity: function updateRefundedValidatorsCount(uint256 , uint256 ) returns()
func (_CSMModule *CSMModuleTransactor) UpdateRefundedValidatorsCount(opts *bind.TransactOpts, arg0 *big.Int, arg1 *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "updateRefundedValidatorsCount", arg0, arg1)
}

// UpdateRefundedValidatorsCount is a paid mutator transaction binding the contract method 0xa2e080f1.
//
// Solidity: function updateRefundedValidatorsCount(uint256 , uint256 ) returns()
func (_CSMModule *CSMModuleSession) UpdateRefundedValidatorsCount(arg0 *big.Int, arg1 *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.UpdateRefundedValidatorsCount(&_CSMModule.TransactOpts, arg0, arg1)
}

// UpdateRefundedValidatorsCount is a paid mutator transaction binding the contract method 0xa2e080f1.
//
// Solidity: function updateRefundedValidatorsCount(uint256 , uint256 ) returns()
func (_CSMModule *CSMModuleTransactorSession) UpdateRefundedValidatorsCount(arg0 *big.Int, arg1 *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.UpdateRefundedValidatorsCount(&_CSMModule.TransactOpts, arg0, arg1)
}

// UpdateStuckValidatorsCount is a paid mutator transaction binding the contract method 0x9b3d1900.
//
// Solidity: function updateStuckValidatorsCount(bytes nodeOperatorIds, bytes stuckValidatorsCounts) returns()
func (_CSMModule *CSMModuleTransactor) UpdateStuckValidatorsCount(opts *bind.TransactOpts, nodeOperatorIds []byte, stuckValidatorsCounts []byte) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "updateStuckValidatorsCount", nodeOperatorIds, stuckValidatorsCounts)
}

// UpdateStuckValidatorsCount is a paid mutator transaction binding the contract method 0x9b3d1900.
//
// Solidity: function updateStuckValidatorsCount(bytes nodeOperatorIds, bytes stuckValidatorsCounts) returns()
func (_CSMModule *CSMModuleSession) UpdateStuckValidatorsCount(nodeOperatorIds []byte, stuckValidatorsCounts []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.UpdateStuckValidatorsCount(&_CSMModule.TransactOpts, nodeOperatorIds, stuckValidatorsCounts)
}

// UpdateStuckValidatorsCount is a paid mutator transaction binding the contract method 0x9b3d1900.
//
// Solidity: function updateStuckValidatorsCount(bytes nodeOperatorIds, bytes stuckValidatorsCounts) returns()
func (_CSMModule *CSMModuleTransactorSession) UpdateStuckValidatorsCount(nodeOperatorIds []byte, stuckValidatorsCounts []byte) (*types.Transaction, error) {
	return _CSMModule.Contract.UpdateStuckValidatorsCount(&_CSMModule.TransactOpts, nodeOperatorIds, stuckValidatorsCounts)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 nodeOperatorId, uint256 targetLimitMode, uint256 targetLimit) returns()
func (_CSMModule *CSMModuleTransactor) UpdateTargetValidatorsLimits(opts *bind.TransactOpts, nodeOperatorId *big.Int, targetLimitMode *big.Int, targetLimit *big.Int) (*types.Transaction, error) {
	return _CSMModule.contract.Transact(opts, "updateTargetValidatorsLimits", nodeOperatorId, targetLimitMode, targetLimit)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 nodeOperatorId, uint256 targetLimitMode, uint256 targetLimit) returns()
func (_CSMModule *CSMModuleSession) UpdateTargetValidatorsLimits(nodeOperatorId *big.Int, targetLimitMode *big.Int, targetLimit *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.UpdateTargetValidatorsLimits(&_CSMModule.TransactOpts, nodeOperatorId, targetLimitMode, targetLimit)
}

// UpdateTargetValidatorsLimits is a paid mutator transaction binding the contract method 0x08a679ad.
//
// Solidity: function updateTargetValidatorsLimits(uint256 nodeOperatorId, uint256 targetLimitMode, uint256 targetLimit) returns()
func (_CSMModule *CSMModuleTransactorSession) UpdateTargetValidatorsLimits(nodeOperatorId *big.Int, targetLimitMode *big.Int, targetLimit *big.Int) (*types.Transaction, error) {
	return _CSMModule.Contract.UpdateTargetValidatorsLimits(&_CSMModule.TransactOpts, nodeOperatorId, targetLimitMode, targetLimit)
}

// CSMModuleBatchEnqueuedIterator is returned from FilterBatchEnqueued and is used to iterate over the raw logs and unpacked data for BatchEnqueued events raised by the CSMModule contract.
type CSMModuleBatchEnqueuedIterator struct {
	Event *CSMModuleBatchEnqueued // Event containing the contract specifics and raw log

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
func (it *CSMModuleBatchEnqueuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleBatchEnqueued)
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
		it.Event = new(CSMModuleBatchEnqueued)
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
func (it *CSMModuleBatchEnqueuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleBatchEnqueuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleBatchEnqueued represents a BatchEnqueued event raised by the CSMModule contract.
type CSMModuleBatchEnqueued struct {
	NodeOperatorId *big.Int
	Count          *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterBatchEnqueued is a free log retrieval operation binding the contract event 0x162b3db9d9ca7d0abe51ad8229dc058550a74b769457fd055579b5bdc5492536.
//
// Solidity: event BatchEnqueued(uint256 indexed nodeOperatorId, uint256 count)
func (_CSMModule *CSMModuleFilterer) FilterBatchEnqueued(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleBatchEnqueuedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "BatchEnqueued", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleBatchEnqueuedIterator{contract: _CSMModule.contract, event: "BatchEnqueued", logs: logs, sub: sub}, nil
}

// WatchBatchEnqueued is a free log subscription operation binding the contract event 0x162b3db9d9ca7d0abe51ad8229dc058550a74b769457fd055579b5bdc5492536.
//
// Solidity: event BatchEnqueued(uint256 indexed nodeOperatorId, uint256 count)
func (_CSMModule *CSMModuleFilterer) WatchBatchEnqueued(opts *bind.WatchOpts, sink chan<- *CSMModuleBatchEnqueued, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "BatchEnqueued", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleBatchEnqueued)
				if err := _CSMModule.contract.UnpackLog(event, "BatchEnqueued", log); err != nil {
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

// ParseBatchEnqueued is a log parse operation binding the contract event 0x162b3db9d9ca7d0abe51ad8229dc058550a74b769457fd055579b5bdc5492536.
//
// Solidity: event BatchEnqueued(uint256 indexed nodeOperatorId, uint256 count)
func (_CSMModule *CSMModuleFilterer) ParseBatchEnqueued(log types.Log) (*CSMModuleBatchEnqueued, error) {
	event := new(CSMModuleBatchEnqueued)
	if err := _CSMModule.contract.UnpackLog(event, "BatchEnqueued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleDepositableSigningKeysCountChangedIterator is returned from FilterDepositableSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for DepositableSigningKeysCountChanged events raised by the CSMModule contract.
type CSMModuleDepositableSigningKeysCountChangedIterator struct {
	Event *CSMModuleDepositableSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleDepositableSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleDepositableSigningKeysCountChanged)
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
		it.Event = new(CSMModuleDepositableSigningKeysCountChanged)
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
func (it *CSMModuleDepositableSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleDepositableSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleDepositableSigningKeysCountChanged represents a DepositableSigningKeysCountChanged event raised by the CSMModule contract.
type CSMModuleDepositableSigningKeysCountChanged struct {
	NodeOperatorId       *big.Int
	DepositableKeysCount *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterDepositableSigningKeysCountChanged is a free log retrieval operation binding the contract event 0xf9109091b368cedad2edff45414eef892edd6b4fe80084bd590aa8f8def8ed33.
//
// Solidity: event DepositableSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositableKeysCount)
func (_CSMModule *CSMModuleFilterer) FilterDepositableSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleDepositableSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "DepositableSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleDepositableSigningKeysCountChangedIterator{contract: _CSMModule.contract, event: "DepositableSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchDepositableSigningKeysCountChanged is a free log subscription operation binding the contract event 0xf9109091b368cedad2edff45414eef892edd6b4fe80084bd590aa8f8def8ed33.
//
// Solidity: event DepositableSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositableKeysCount)
func (_CSMModule *CSMModuleFilterer) WatchDepositableSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleDepositableSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "DepositableSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleDepositableSigningKeysCountChanged)
				if err := _CSMModule.contract.UnpackLog(event, "DepositableSigningKeysCountChanged", log); err != nil {
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

// ParseDepositableSigningKeysCountChanged is a log parse operation binding the contract event 0xf9109091b368cedad2edff45414eef892edd6b4fe80084bd590aa8f8def8ed33.
//
// Solidity: event DepositableSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositableKeysCount)
func (_CSMModule *CSMModuleFilterer) ParseDepositableSigningKeysCountChanged(log types.Log) (*CSMModuleDepositableSigningKeysCountChanged, error) {
	event := new(CSMModuleDepositableSigningKeysCountChanged)
	if err := _CSMModule.contract.UnpackLog(event, "DepositableSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleDepositedSigningKeysCountChangedIterator is returned from FilterDepositedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for DepositedSigningKeysCountChanged events raised by the CSMModule contract.
type CSMModuleDepositedSigningKeysCountChangedIterator struct {
	Event *CSMModuleDepositedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleDepositedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleDepositedSigningKeysCountChanged)
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
		it.Event = new(CSMModuleDepositedSigningKeysCountChanged)
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
func (it *CSMModuleDepositedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleDepositedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleDepositedSigningKeysCountChanged represents a DepositedSigningKeysCountChanged event raised by the CSMModule contract.
type CSMModuleDepositedSigningKeysCountChanged struct {
	NodeOperatorId     *big.Int
	DepositedKeysCount *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterDepositedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x24eb1c9e765ba41accf9437300ea91ece5ed3f897ec3cdee0e9debd7fe309b78.
//
// Solidity: event DepositedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositedKeysCount)
func (_CSMModule *CSMModuleFilterer) FilterDepositedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleDepositedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "DepositedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleDepositedSigningKeysCountChangedIterator{contract: _CSMModule.contract, event: "DepositedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchDepositedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x24eb1c9e765ba41accf9437300ea91ece5ed3f897ec3cdee0e9debd7fe309b78.
//
// Solidity: event DepositedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositedKeysCount)
func (_CSMModule *CSMModuleFilterer) WatchDepositedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleDepositedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "DepositedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleDepositedSigningKeysCountChanged)
				if err := _CSMModule.contract.UnpackLog(event, "DepositedSigningKeysCountChanged", log); err != nil {
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
// Solidity: event DepositedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 depositedKeysCount)
func (_CSMModule *CSMModuleFilterer) ParseDepositedSigningKeysCountChanged(log types.Log) (*CSMModuleDepositedSigningKeysCountChanged, error) {
	event := new(CSMModuleDepositedSigningKeysCountChanged)
	if err := _CSMModule.contract.UnpackLog(event, "DepositedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleELRewardsStealingPenaltyCancelledIterator is returned from FilterELRewardsStealingPenaltyCancelled and is used to iterate over the raw logs and unpacked data for ELRewardsStealingPenaltyCancelled events raised by the CSMModule contract.
type CSMModuleELRewardsStealingPenaltyCancelledIterator struct {
	Event *CSMModuleELRewardsStealingPenaltyCancelled // Event containing the contract specifics and raw log

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
func (it *CSMModuleELRewardsStealingPenaltyCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleELRewardsStealingPenaltyCancelled)
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
		it.Event = new(CSMModuleELRewardsStealingPenaltyCancelled)
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
func (it *CSMModuleELRewardsStealingPenaltyCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleELRewardsStealingPenaltyCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleELRewardsStealingPenaltyCancelled represents a ELRewardsStealingPenaltyCancelled event raised by the CSMModule contract.
type CSMModuleELRewardsStealingPenaltyCancelled struct {
	NodeOperatorId *big.Int
	Amount         *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterELRewardsStealingPenaltyCancelled is a free log retrieval operation binding the contract event 0x1e7ebd3c5f4de9502000b6f7e6e7cf5d4ecb27d6fe1778e43fb9d1d0ca87d0e7.
//
// Solidity: event ELRewardsStealingPenaltyCancelled(uint256 indexed nodeOperatorId, uint256 amount)
func (_CSMModule *CSMModuleFilterer) FilterELRewardsStealingPenaltyCancelled(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleELRewardsStealingPenaltyCancelledIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ELRewardsStealingPenaltyCancelled", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleELRewardsStealingPenaltyCancelledIterator{contract: _CSMModule.contract, event: "ELRewardsStealingPenaltyCancelled", logs: logs, sub: sub}, nil
}

// WatchELRewardsStealingPenaltyCancelled is a free log subscription operation binding the contract event 0x1e7ebd3c5f4de9502000b6f7e6e7cf5d4ecb27d6fe1778e43fb9d1d0ca87d0e7.
//
// Solidity: event ELRewardsStealingPenaltyCancelled(uint256 indexed nodeOperatorId, uint256 amount)
func (_CSMModule *CSMModuleFilterer) WatchELRewardsStealingPenaltyCancelled(opts *bind.WatchOpts, sink chan<- *CSMModuleELRewardsStealingPenaltyCancelled, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ELRewardsStealingPenaltyCancelled", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleELRewardsStealingPenaltyCancelled)
				if err := _CSMModule.contract.UnpackLog(event, "ELRewardsStealingPenaltyCancelled", log); err != nil {
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

// ParseELRewardsStealingPenaltyCancelled is a log parse operation binding the contract event 0x1e7ebd3c5f4de9502000b6f7e6e7cf5d4ecb27d6fe1778e43fb9d1d0ca87d0e7.
//
// Solidity: event ELRewardsStealingPenaltyCancelled(uint256 indexed nodeOperatorId, uint256 amount)
func (_CSMModule *CSMModuleFilterer) ParseELRewardsStealingPenaltyCancelled(log types.Log) (*CSMModuleELRewardsStealingPenaltyCancelled, error) {
	event := new(CSMModuleELRewardsStealingPenaltyCancelled)
	if err := _CSMModule.contract.UnpackLog(event, "ELRewardsStealingPenaltyCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleELRewardsStealingPenaltyCompensatedIterator is returned from FilterELRewardsStealingPenaltyCompensated and is used to iterate over the raw logs and unpacked data for ELRewardsStealingPenaltyCompensated events raised by the CSMModule contract.
type CSMModuleELRewardsStealingPenaltyCompensatedIterator struct {
	Event *CSMModuleELRewardsStealingPenaltyCompensated // Event containing the contract specifics and raw log

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
func (it *CSMModuleELRewardsStealingPenaltyCompensatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleELRewardsStealingPenaltyCompensated)
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
		it.Event = new(CSMModuleELRewardsStealingPenaltyCompensated)
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
func (it *CSMModuleELRewardsStealingPenaltyCompensatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleELRewardsStealingPenaltyCompensatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleELRewardsStealingPenaltyCompensated represents a ELRewardsStealingPenaltyCompensated event raised by the CSMModule contract.
type CSMModuleELRewardsStealingPenaltyCompensated struct {
	NodeOperatorId *big.Int
	Amount         *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterELRewardsStealingPenaltyCompensated is a free log retrieval operation binding the contract event 0xb1858b4c2ab6242521725a8f7350a6cb22ad4ecae009c9b63ef114baffb054be.
//
// Solidity: event ELRewardsStealingPenaltyCompensated(uint256 indexed nodeOperatorId, uint256 amount)
func (_CSMModule *CSMModuleFilterer) FilterELRewardsStealingPenaltyCompensated(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleELRewardsStealingPenaltyCompensatedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ELRewardsStealingPenaltyCompensated", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleELRewardsStealingPenaltyCompensatedIterator{contract: _CSMModule.contract, event: "ELRewardsStealingPenaltyCompensated", logs: logs, sub: sub}, nil
}

// WatchELRewardsStealingPenaltyCompensated is a free log subscription operation binding the contract event 0xb1858b4c2ab6242521725a8f7350a6cb22ad4ecae009c9b63ef114baffb054be.
//
// Solidity: event ELRewardsStealingPenaltyCompensated(uint256 indexed nodeOperatorId, uint256 amount)
func (_CSMModule *CSMModuleFilterer) WatchELRewardsStealingPenaltyCompensated(opts *bind.WatchOpts, sink chan<- *CSMModuleELRewardsStealingPenaltyCompensated, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ELRewardsStealingPenaltyCompensated", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleELRewardsStealingPenaltyCompensated)
				if err := _CSMModule.contract.UnpackLog(event, "ELRewardsStealingPenaltyCompensated", log); err != nil {
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

// ParseELRewardsStealingPenaltyCompensated is a log parse operation binding the contract event 0xb1858b4c2ab6242521725a8f7350a6cb22ad4ecae009c9b63ef114baffb054be.
//
// Solidity: event ELRewardsStealingPenaltyCompensated(uint256 indexed nodeOperatorId, uint256 amount)
func (_CSMModule *CSMModuleFilterer) ParseELRewardsStealingPenaltyCompensated(log types.Log) (*CSMModuleELRewardsStealingPenaltyCompensated, error) {
	event := new(CSMModuleELRewardsStealingPenaltyCompensated)
	if err := _CSMModule.contract.UnpackLog(event, "ELRewardsStealingPenaltyCompensated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleELRewardsStealingPenaltyReportedIterator is returned from FilterELRewardsStealingPenaltyReported and is used to iterate over the raw logs and unpacked data for ELRewardsStealingPenaltyReported events raised by the CSMModule contract.
type CSMModuleELRewardsStealingPenaltyReportedIterator struct {
	Event *CSMModuleELRewardsStealingPenaltyReported // Event containing the contract specifics and raw log

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
func (it *CSMModuleELRewardsStealingPenaltyReportedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleELRewardsStealingPenaltyReported)
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
		it.Event = new(CSMModuleELRewardsStealingPenaltyReported)
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
func (it *CSMModuleELRewardsStealingPenaltyReportedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleELRewardsStealingPenaltyReportedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleELRewardsStealingPenaltyReported represents a ELRewardsStealingPenaltyReported event raised by the CSMModule contract.
type CSMModuleELRewardsStealingPenaltyReported struct {
	NodeOperatorId    *big.Int
	ProposedBlockHash [32]byte
	StolenAmount      *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterELRewardsStealingPenaltyReported is a free log retrieval operation binding the contract event 0xeec4d6dbe34149c6728a9638eca869d0e5a7fcd85c7a96178f7e9780b4b7fe4b.
//
// Solidity: event ELRewardsStealingPenaltyReported(uint256 indexed nodeOperatorId, bytes32 proposedBlockHash, uint256 stolenAmount)
func (_CSMModule *CSMModuleFilterer) FilterELRewardsStealingPenaltyReported(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleELRewardsStealingPenaltyReportedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ELRewardsStealingPenaltyReported", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleELRewardsStealingPenaltyReportedIterator{contract: _CSMModule.contract, event: "ELRewardsStealingPenaltyReported", logs: logs, sub: sub}, nil
}

// WatchELRewardsStealingPenaltyReported is a free log subscription operation binding the contract event 0xeec4d6dbe34149c6728a9638eca869d0e5a7fcd85c7a96178f7e9780b4b7fe4b.
//
// Solidity: event ELRewardsStealingPenaltyReported(uint256 indexed nodeOperatorId, bytes32 proposedBlockHash, uint256 stolenAmount)
func (_CSMModule *CSMModuleFilterer) WatchELRewardsStealingPenaltyReported(opts *bind.WatchOpts, sink chan<- *CSMModuleELRewardsStealingPenaltyReported, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ELRewardsStealingPenaltyReported", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleELRewardsStealingPenaltyReported)
				if err := _CSMModule.contract.UnpackLog(event, "ELRewardsStealingPenaltyReported", log); err != nil {
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

// ParseELRewardsStealingPenaltyReported is a log parse operation binding the contract event 0xeec4d6dbe34149c6728a9638eca869d0e5a7fcd85c7a96178f7e9780b4b7fe4b.
//
// Solidity: event ELRewardsStealingPenaltyReported(uint256 indexed nodeOperatorId, bytes32 proposedBlockHash, uint256 stolenAmount)
func (_CSMModule *CSMModuleFilterer) ParseELRewardsStealingPenaltyReported(log types.Log) (*CSMModuleELRewardsStealingPenaltyReported, error) {
	event := new(CSMModuleELRewardsStealingPenaltyReported)
	if err := _CSMModule.contract.UnpackLog(event, "ELRewardsStealingPenaltyReported", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleELRewardsStealingPenaltySettledIterator is returned from FilterELRewardsStealingPenaltySettled and is used to iterate over the raw logs and unpacked data for ELRewardsStealingPenaltySettled events raised by the CSMModule contract.
type CSMModuleELRewardsStealingPenaltySettledIterator struct {
	Event *CSMModuleELRewardsStealingPenaltySettled // Event containing the contract specifics and raw log

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
func (it *CSMModuleELRewardsStealingPenaltySettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleELRewardsStealingPenaltySettled)
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
		it.Event = new(CSMModuleELRewardsStealingPenaltySettled)
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
func (it *CSMModuleELRewardsStealingPenaltySettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleELRewardsStealingPenaltySettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleELRewardsStealingPenaltySettled represents a ELRewardsStealingPenaltySettled event raised by the CSMModule contract.
type CSMModuleELRewardsStealingPenaltySettled struct {
	NodeOperatorId *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterELRewardsStealingPenaltySettled is a free log retrieval operation binding the contract event 0x00f4fe19c0404d2fbb58da6f646c0a3ee5a6994a034213bbd22b072ed1ca5c27.
//
// Solidity: event ELRewardsStealingPenaltySettled(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) FilterELRewardsStealingPenaltySettled(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleELRewardsStealingPenaltySettledIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ELRewardsStealingPenaltySettled", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleELRewardsStealingPenaltySettledIterator{contract: _CSMModule.contract, event: "ELRewardsStealingPenaltySettled", logs: logs, sub: sub}, nil
}

// WatchELRewardsStealingPenaltySettled is a free log subscription operation binding the contract event 0x00f4fe19c0404d2fbb58da6f646c0a3ee5a6994a034213bbd22b072ed1ca5c27.
//
// Solidity: event ELRewardsStealingPenaltySettled(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) WatchELRewardsStealingPenaltySettled(opts *bind.WatchOpts, sink chan<- *CSMModuleELRewardsStealingPenaltySettled, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ELRewardsStealingPenaltySettled", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleELRewardsStealingPenaltySettled)
				if err := _CSMModule.contract.UnpackLog(event, "ELRewardsStealingPenaltySettled", log); err != nil {
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

// ParseELRewardsStealingPenaltySettled is a log parse operation binding the contract event 0x00f4fe19c0404d2fbb58da6f646c0a3ee5a6994a034213bbd22b072ed1ca5c27.
//
// Solidity: event ELRewardsStealingPenaltySettled(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) ParseELRewardsStealingPenaltySettled(log types.Log) (*CSMModuleELRewardsStealingPenaltySettled, error) {
	event := new(CSMModuleELRewardsStealingPenaltySettled)
	if err := _CSMModule.contract.UnpackLog(event, "ELRewardsStealingPenaltySettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleERC1155RecoveredIterator is returned from FilterERC1155Recovered and is used to iterate over the raw logs and unpacked data for ERC1155Recovered events raised by the CSMModule contract.
type CSMModuleERC1155RecoveredIterator struct {
	Event *CSMModuleERC1155Recovered // Event containing the contract specifics and raw log

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
func (it *CSMModuleERC1155RecoveredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleERC1155Recovered)
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
		it.Event = new(CSMModuleERC1155Recovered)
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
func (it *CSMModuleERC1155RecoveredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleERC1155RecoveredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleERC1155Recovered represents a ERC1155Recovered event raised by the CSMModule contract.
type CSMModuleERC1155Recovered struct {
	Token     common.Address
	TokenId   *big.Int
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterERC1155Recovered is a free log retrieval operation binding the contract event 0x5cf02e753b3eb0f4bee4460a72817d8e5e3c75cd4d65c1d0b06dca88b8032936.
//
// Solidity: event ERC1155Recovered(address indexed token, uint256 tokenId, address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) FilterERC1155Recovered(opts *bind.FilterOpts, token []common.Address, recipient []common.Address) (*CSMModuleERC1155RecoveredIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ERC1155Recovered", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleERC1155RecoveredIterator{contract: _CSMModule.contract, event: "ERC1155Recovered", logs: logs, sub: sub}, nil
}

// WatchERC1155Recovered is a free log subscription operation binding the contract event 0x5cf02e753b3eb0f4bee4460a72817d8e5e3c75cd4d65c1d0b06dca88b8032936.
//
// Solidity: event ERC1155Recovered(address indexed token, uint256 tokenId, address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) WatchERC1155Recovered(opts *bind.WatchOpts, sink chan<- *CSMModuleERC1155Recovered, token []common.Address, recipient []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ERC1155Recovered", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleERC1155Recovered)
				if err := _CSMModule.contract.UnpackLog(event, "ERC1155Recovered", log); err != nil {
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

// ParseERC1155Recovered is a log parse operation binding the contract event 0x5cf02e753b3eb0f4bee4460a72817d8e5e3c75cd4d65c1d0b06dca88b8032936.
//
// Solidity: event ERC1155Recovered(address indexed token, uint256 tokenId, address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) ParseERC1155Recovered(log types.Log) (*CSMModuleERC1155Recovered, error) {
	event := new(CSMModuleERC1155Recovered)
	if err := _CSMModule.contract.UnpackLog(event, "ERC1155Recovered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleERC20RecoveredIterator is returned from FilterERC20Recovered and is used to iterate over the raw logs and unpacked data for ERC20Recovered events raised by the CSMModule contract.
type CSMModuleERC20RecoveredIterator struct {
	Event *CSMModuleERC20Recovered // Event containing the contract specifics and raw log

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
func (it *CSMModuleERC20RecoveredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleERC20Recovered)
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
		it.Event = new(CSMModuleERC20Recovered)
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
func (it *CSMModuleERC20RecoveredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleERC20RecoveredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleERC20Recovered represents a ERC20Recovered event raised by the CSMModule contract.
type CSMModuleERC20Recovered struct {
	Token     common.Address
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterERC20Recovered is a free log retrieval operation binding the contract event 0xaca8fb252cde442184e5f10e0f2e6e4029e8cd7717cae63559079610702436aa.
//
// Solidity: event ERC20Recovered(address indexed token, address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) FilterERC20Recovered(opts *bind.FilterOpts, token []common.Address, recipient []common.Address) (*CSMModuleERC20RecoveredIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ERC20Recovered", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleERC20RecoveredIterator{contract: _CSMModule.contract, event: "ERC20Recovered", logs: logs, sub: sub}, nil
}

// WatchERC20Recovered is a free log subscription operation binding the contract event 0xaca8fb252cde442184e5f10e0f2e6e4029e8cd7717cae63559079610702436aa.
//
// Solidity: event ERC20Recovered(address indexed token, address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) WatchERC20Recovered(opts *bind.WatchOpts, sink chan<- *CSMModuleERC20Recovered, token []common.Address, recipient []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ERC20Recovered", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleERC20Recovered)
				if err := _CSMModule.contract.UnpackLog(event, "ERC20Recovered", log); err != nil {
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

// ParseERC20Recovered is a log parse operation binding the contract event 0xaca8fb252cde442184e5f10e0f2e6e4029e8cd7717cae63559079610702436aa.
//
// Solidity: event ERC20Recovered(address indexed token, address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) ParseERC20Recovered(log types.Log) (*CSMModuleERC20Recovered, error) {
	event := new(CSMModuleERC20Recovered)
	if err := _CSMModule.contract.UnpackLog(event, "ERC20Recovered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleERC721RecoveredIterator is returned from FilterERC721Recovered and is used to iterate over the raw logs and unpacked data for ERC721Recovered events raised by the CSMModule contract.
type CSMModuleERC721RecoveredIterator struct {
	Event *CSMModuleERC721Recovered // Event containing the contract specifics and raw log

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
func (it *CSMModuleERC721RecoveredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleERC721Recovered)
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
		it.Event = new(CSMModuleERC721Recovered)
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
func (it *CSMModuleERC721RecoveredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleERC721RecoveredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleERC721Recovered represents a ERC721Recovered event raised by the CSMModule contract.
type CSMModuleERC721Recovered struct {
	Token     common.Address
	TokenId   *big.Int
	Recipient common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterERC721Recovered is a free log retrieval operation binding the contract event 0x8166bf75d2ff2fa3c8f3c44410540bf42e9a5359b48409e8d660291dc9f788c8.
//
// Solidity: event ERC721Recovered(address indexed token, uint256 tokenId, address indexed recipient)
func (_CSMModule *CSMModuleFilterer) FilterERC721Recovered(opts *bind.FilterOpts, token []common.Address, recipient []common.Address) (*CSMModuleERC721RecoveredIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ERC721Recovered", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleERC721RecoveredIterator{contract: _CSMModule.contract, event: "ERC721Recovered", logs: logs, sub: sub}, nil
}

// WatchERC721Recovered is a free log subscription operation binding the contract event 0x8166bf75d2ff2fa3c8f3c44410540bf42e9a5359b48409e8d660291dc9f788c8.
//
// Solidity: event ERC721Recovered(address indexed token, uint256 tokenId, address indexed recipient)
func (_CSMModule *CSMModuleFilterer) WatchERC721Recovered(opts *bind.WatchOpts, sink chan<- *CSMModuleERC721Recovered, token []common.Address, recipient []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ERC721Recovered", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleERC721Recovered)
				if err := _CSMModule.contract.UnpackLog(event, "ERC721Recovered", log); err != nil {
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

// ParseERC721Recovered is a log parse operation binding the contract event 0x8166bf75d2ff2fa3c8f3c44410540bf42e9a5359b48409e8d660291dc9f788c8.
//
// Solidity: event ERC721Recovered(address indexed token, uint256 tokenId, address indexed recipient)
func (_CSMModule *CSMModuleFilterer) ParseERC721Recovered(log types.Log) (*CSMModuleERC721Recovered, error) {
	event := new(CSMModuleERC721Recovered)
	if err := _CSMModule.contract.UnpackLog(event, "ERC721Recovered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleEtherRecoveredIterator is returned from FilterEtherRecovered and is used to iterate over the raw logs and unpacked data for EtherRecovered events raised by the CSMModule contract.
type CSMModuleEtherRecoveredIterator struct {
	Event *CSMModuleEtherRecovered // Event containing the contract specifics and raw log

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
func (it *CSMModuleEtherRecoveredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleEtherRecovered)
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
		it.Event = new(CSMModuleEtherRecovered)
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
func (it *CSMModuleEtherRecoveredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleEtherRecoveredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleEtherRecovered represents a EtherRecovered event raised by the CSMModule contract.
type CSMModuleEtherRecovered struct {
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEtherRecovered is a free log retrieval operation binding the contract event 0x8e274e42262a7f013b700b35c2b4629ccce1702f8fe83f8dfb7eacbb26a4382c.
//
// Solidity: event EtherRecovered(address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) FilterEtherRecovered(opts *bind.FilterOpts, recipient []common.Address) (*CSMModuleEtherRecoveredIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "EtherRecovered", recipientRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleEtherRecoveredIterator{contract: _CSMModule.contract, event: "EtherRecovered", logs: logs, sub: sub}, nil
}

// WatchEtherRecovered is a free log subscription operation binding the contract event 0x8e274e42262a7f013b700b35c2b4629ccce1702f8fe83f8dfb7eacbb26a4382c.
//
// Solidity: event EtherRecovered(address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) WatchEtherRecovered(opts *bind.WatchOpts, sink chan<- *CSMModuleEtherRecovered, recipient []common.Address) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "EtherRecovered", recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleEtherRecovered)
				if err := _CSMModule.contract.UnpackLog(event, "EtherRecovered", log); err != nil {
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

// ParseEtherRecovered is a log parse operation binding the contract event 0x8e274e42262a7f013b700b35c2b4629ccce1702f8fe83f8dfb7eacbb26a4382c.
//
// Solidity: event EtherRecovered(address indexed recipient, uint256 amount)
func (_CSMModule *CSMModuleFilterer) ParseEtherRecovered(log types.Log) (*CSMModuleEtherRecovered, error) {
	event := new(CSMModuleEtherRecovered)
	if err := _CSMModule.contract.UnpackLog(event, "EtherRecovered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleExitedSigningKeysCountChangedIterator is returned from FilterExitedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for ExitedSigningKeysCountChanged events raised by the CSMModule contract.
type CSMModuleExitedSigningKeysCountChangedIterator struct {
	Event *CSMModuleExitedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleExitedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleExitedSigningKeysCountChanged)
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
		it.Event = new(CSMModuleExitedSigningKeysCountChanged)
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
func (it *CSMModuleExitedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleExitedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleExitedSigningKeysCountChanged represents a ExitedSigningKeysCountChanged event raised by the CSMModule contract.
type CSMModuleExitedSigningKeysCountChanged struct {
	NodeOperatorId  *big.Int
	ExitedKeysCount *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterExitedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x0f67960648751434ae86bf350db61194f387fda387e7f568b0ccd0ae0c220166.
//
// Solidity: event ExitedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 exitedKeysCount)
func (_CSMModule *CSMModuleFilterer) FilterExitedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleExitedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ExitedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleExitedSigningKeysCountChangedIterator{contract: _CSMModule.contract, event: "ExitedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchExitedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x0f67960648751434ae86bf350db61194f387fda387e7f568b0ccd0ae0c220166.
//
// Solidity: event ExitedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 exitedKeysCount)
func (_CSMModule *CSMModuleFilterer) WatchExitedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleExitedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ExitedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleExitedSigningKeysCountChanged)
				if err := _CSMModule.contract.UnpackLog(event, "ExitedSigningKeysCountChanged", log); err != nil {
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
// Solidity: event ExitedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 exitedKeysCount)
func (_CSMModule *CSMModuleFilterer) ParseExitedSigningKeysCountChanged(log types.Log) (*CSMModuleExitedSigningKeysCountChanged, error) {
	event := new(CSMModuleExitedSigningKeysCountChanged)
	if err := _CSMModule.contract.UnpackLog(event, "ExitedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleInitialSlashingSubmittedIterator is returned from FilterInitialSlashingSubmitted and is used to iterate over the raw logs and unpacked data for InitialSlashingSubmitted events raised by the CSMModule contract.
type CSMModuleInitialSlashingSubmittedIterator struct {
	Event *CSMModuleInitialSlashingSubmitted // Event containing the contract specifics and raw log

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
func (it *CSMModuleInitialSlashingSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleInitialSlashingSubmitted)
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
		it.Event = new(CSMModuleInitialSlashingSubmitted)
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
func (it *CSMModuleInitialSlashingSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleInitialSlashingSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleInitialSlashingSubmitted represents a InitialSlashingSubmitted event raised by the CSMModule contract.
type CSMModuleInitialSlashingSubmitted struct {
	NodeOperatorId *big.Int
	KeyIndex       *big.Int
	Pubkey         []byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterInitialSlashingSubmitted is a free log retrieval operation binding the contract event 0x0d541877c9d326d4c8ccfd72e6719f06dccb62a28292ae647e923441bcaad5c0.
//
// Solidity: event InitialSlashingSubmitted(uint256 indexed nodeOperatorId, uint256 keyIndex, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) FilterInitialSlashingSubmitted(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleInitialSlashingSubmittedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "InitialSlashingSubmitted", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleInitialSlashingSubmittedIterator{contract: _CSMModule.contract, event: "InitialSlashingSubmitted", logs: logs, sub: sub}, nil
}

// WatchInitialSlashingSubmitted is a free log subscription operation binding the contract event 0x0d541877c9d326d4c8ccfd72e6719f06dccb62a28292ae647e923441bcaad5c0.
//
// Solidity: event InitialSlashingSubmitted(uint256 indexed nodeOperatorId, uint256 keyIndex, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) WatchInitialSlashingSubmitted(opts *bind.WatchOpts, sink chan<- *CSMModuleInitialSlashingSubmitted, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "InitialSlashingSubmitted", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleInitialSlashingSubmitted)
				if err := _CSMModule.contract.UnpackLog(event, "InitialSlashingSubmitted", log); err != nil {
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

// ParseInitialSlashingSubmitted is a log parse operation binding the contract event 0x0d541877c9d326d4c8ccfd72e6719f06dccb62a28292ae647e923441bcaad5c0.
//
// Solidity: event InitialSlashingSubmitted(uint256 indexed nodeOperatorId, uint256 keyIndex, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) ParseInitialSlashingSubmitted(log types.Log) (*CSMModuleInitialSlashingSubmitted, error) {
	event := new(CSMModuleInitialSlashingSubmitted)
	if err := _CSMModule.contract.UnpackLog(event, "InitialSlashingSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the CSMModule contract.
type CSMModuleInitializedIterator struct {
	Event *CSMModuleInitialized // Event containing the contract specifics and raw log

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
func (it *CSMModuleInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleInitialized)
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
		it.Event = new(CSMModuleInitialized)
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
func (it *CSMModuleInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleInitialized represents a Initialized event raised by the CSMModule contract.
type CSMModuleInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_CSMModule *CSMModuleFilterer) FilterInitialized(opts *bind.FilterOpts) (*CSMModuleInitializedIterator, error) {

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &CSMModuleInitializedIterator{contract: _CSMModule.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_CSMModule *CSMModuleFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *CSMModuleInitialized) (event.Subscription, error) {

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleInitialized)
				if err := _CSMModule.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_CSMModule *CSMModuleFilterer) ParseInitialized(log types.Log) (*CSMModuleInitialized, error) {
	event := new(CSMModuleInitialized)
	if err := _CSMModule.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleKeyRemovalChargeAppliedIterator is returned from FilterKeyRemovalChargeApplied and is used to iterate over the raw logs and unpacked data for KeyRemovalChargeApplied events raised by the CSMModule contract.
type CSMModuleKeyRemovalChargeAppliedIterator struct {
	Event *CSMModuleKeyRemovalChargeApplied // Event containing the contract specifics and raw log

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
func (it *CSMModuleKeyRemovalChargeAppliedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleKeyRemovalChargeApplied)
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
		it.Event = new(CSMModuleKeyRemovalChargeApplied)
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
func (it *CSMModuleKeyRemovalChargeAppliedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleKeyRemovalChargeAppliedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleKeyRemovalChargeApplied represents a KeyRemovalChargeApplied event raised by the CSMModule contract.
type CSMModuleKeyRemovalChargeApplied struct {
	NodeOperatorId *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterKeyRemovalChargeApplied is a free log retrieval operation binding the contract event 0x1cbb8dafbedbdf4f813a8ed1f50d871def63e1104f8729b677af57905eda90f6.
//
// Solidity: event KeyRemovalChargeApplied(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) FilterKeyRemovalChargeApplied(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleKeyRemovalChargeAppliedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "KeyRemovalChargeApplied", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleKeyRemovalChargeAppliedIterator{contract: _CSMModule.contract, event: "KeyRemovalChargeApplied", logs: logs, sub: sub}, nil
}

// WatchKeyRemovalChargeApplied is a free log subscription operation binding the contract event 0x1cbb8dafbedbdf4f813a8ed1f50d871def63e1104f8729b677af57905eda90f6.
//
// Solidity: event KeyRemovalChargeApplied(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) WatchKeyRemovalChargeApplied(opts *bind.WatchOpts, sink chan<- *CSMModuleKeyRemovalChargeApplied, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "KeyRemovalChargeApplied", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleKeyRemovalChargeApplied)
				if err := _CSMModule.contract.UnpackLog(event, "KeyRemovalChargeApplied", log); err != nil {
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

// ParseKeyRemovalChargeApplied is a log parse operation binding the contract event 0x1cbb8dafbedbdf4f813a8ed1f50d871def63e1104f8729b677af57905eda90f6.
//
// Solidity: event KeyRemovalChargeApplied(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) ParseKeyRemovalChargeApplied(log types.Log) (*CSMModuleKeyRemovalChargeApplied, error) {
	event := new(CSMModuleKeyRemovalChargeApplied)
	if err := _CSMModule.contract.UnpackLog(event, "KeyRemovalChargeApplied", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleKeyRemovalChargeSetIterator is returned from FilterKeyRemovalChargeSet and is used to iterate over the raw logs and unpacked data for KeyRemovalChargeSet events raised by the CSMModule contract.
type CSMModuleKeyRemovalChargeSetIterator struct {
	Event *CSMModuleKeyRemovalChargeSet // Event containing the contract specifics and raw log

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
func (it *CSMModuleKeyRemovalChargeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleKeyRemovalChargeSet)
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
		it.Event = new(CSMModuleKeyRemovalChargeSet)
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
func (it *CSMModuleKeyRemovalChargeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleKeyRemovalChargeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleKeyRemovalChargeSet represents a KeyRemovalChargeSet event raised by the CSMModule contract.
type CSMModuleKeyRemovalChargeSet struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterKeyRemovalChargeSet is a free log retrieval operation binding the contract event 0x699ec9c671aad1f3dcc15e571375584a1d6fb7176afd545d14467fd31477e98e.
//
// Solidity: event KeyRemovalChargeSet(uint256 amount)
func (_CSMModule *CSMModuleFilterer) FilterKeyRemovalChargeSet(opts *bind.FilterOpts) (*CSMModuleKeyRemovalChargeSetIterator, error) {

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "KeyRemovalChargeSet")
	if err != nil {
		return nil, err
	}
	return &CSMModuleKeyRemovalChargeSetIterator{contract: _CSMModule.contract, event: "KeyRemovalChargeSet", logs: logs, sub: sub}, nil
}

// WatchKeyRemovalChargeSet is a free log subscription operation binding the contract event 0x699ec9c671aad1f3dcc15e571375584a1d6fb7176afd545d14467fd31477e98e.
//
// Solidity: event KeyRemovalChargeSet(uint256 amount)
func (_CSMModule *CSMModuleFilterer) WatchKeyRemovalChargeSet(opts *bind.WatchOpts, sink chan<- *CSMModuleKeyRemovalChargeSet) (event.Subscription, error) {

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "KeyRemovalChargeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleKeyRemovalChargeSet)
				if err := _CSMModule.contract.UnpackLog(event, "KeyRemovalChargeSet", log); err != nil {
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

// ParseKeyRemovalChargeSet is a log parse operation binding the contract event 0x699ec9c671aad1f3dcc15e571375584a1d6fb7176afd545d14467fd31477e98e.
//
// Solidity: event KeyRemovalChargeSet(uint256 amount)
func (_CSMModule *CSMModuleFilterer) ParseKeyRemovalChargeSet(log types.Log) (*CSMModuleKeyRemovalChargeSet, error) {
	event := new(CSMModuleKeyRemovalChargeSet)
	if err := _CSMModule.contract.UnpackLog(event, "KeyRemovalChargeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleNodeOperatorAddedIterator is returned from FilterNodeOperatorAdded and is used to iterate over the raw logs and unpacked data for NodeOperatorAdded events raised by the CSMModule contract.
type CSMModuleNodeOperatorAddedIterator struct {
	Event *CSMModuleNodeOperatorAdded // Event containing the contract specifics and raw log

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
func (it *CSMModuleNodeOperatorAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleNodeOperatorAdded)
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
		it.Event = new(CSMModuleNodeOperatorAdded)
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
func (it *CSMModuleNodeOperatorAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleNodeOperatorAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleNodeOperatorAdded represents a NodeOperatorAdded event raised by the CSMModule contract.
type CSMModuleNodeOperatorAdded struct {
	NodeOperatorId *big.Int
	ManagerAddress common.Address
	RewardAddress  common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorAdded is a free log retrieval operation binding the contract event 0xf35982c84fdc94f58d48e901c54c615804cf7d7939b9b8f76ce4d459354e6363.
//
// Solidity: event NodeOperatorAdded(uint256 indexed nodeOperatorId, address indexed managerAddress, address indexed rewardAddress)
func (_CSMModule *CSMModuleFilterer) FilterNodeOperatorAdded(opts *bind.FilterOpts, nodeOperatorId []*big.Int, managerAddress []common.Address, rewardAddress []common.Address) (*CSMModuleNodeOperatorAddedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var managerAddressRule []interface{}
	for _, managerAddressItem := range managerAddress {
		managerAddressRule = append(managerAddressRule, managerAddressItem)
	}
	var rewardAddressRule []interface{}
	for _, rewardAddressItem := range rewardAddress {
		rewardAddressRule = append(rewardAddressRule, rewardAddressItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "NodeOperatorAdded", nodeOperatorIdRule, managerAddressRule, rewardAddressRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleNodeOperatorAddedIterator{contract: _CSMModule.contract, event: "NodeOperatorAdded", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorAdded is a free log subscription operation binding the contract event 0xf35982c84fdc94f58d48e901c54c615804cf7d7939b9b8f76ce4d459354e6363.
//
// Solidity: event NodeOperatorAdded(uint256 indexed nodeOperatorId, address indexed managerAddress, address indexed rewardAddress)
func (_CSMModule *CSMModuleFilterer) WatchNodeOperatorAdded(opts *bind.WatchOpts, sink chan<- *CSMModuleNodeOperatorAdded, nodeOperatorId []*big.Int, managerAddress []common.Address, rewardAddress []common.Address) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var managerAddressRule []interface{}
	for _, managerAddressItem := range managerAddress {
		managerAddressRule = append(managerAddressRule, managerAddressItem)
	}
	var rewardAddressRule []interface{}
	for _, rewardAddressItem := range rewardAddress {
		rewardAddressRule = append(rewardAddressRule, rewardAddressItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "NodeOperatorAdded", nodeOperatorIdRule, managerAddressRule, rewardAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleNodeOperatorAdded)
				if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorAdded", log); err != nil {
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

// ParseNodeOperatorAdded is a log parse operation binding the contract event 0xf35982c84fdc94f58d48e901c54c615804cf7d7939b9b8f76ce4d459354e6363.
//
// Solidity: event NodeOperatorAdded(uint256 indexed nodeOperatorId, address indexed managerAddress, address indexed rewardAddress)
func (_CSMModule *CSMModuleFilterer) ParseNodeOperatorAdded(log types.Log) (*CSMModuleNodeOperatorAdded, error) {
	event := new(CSMModuleNodeOperatorAdded)
	if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleNodeOperatorManagerAddressChangeProposedIterator is returned from FilterNodeOperatorManagerAddressChangeProposed and is used to iterate over the raw logs and unpacked data for NodeOperatorManagerAddressChangeProposed events raised by the CSMModule contract.
type CSMModuleNodeOperatorManagerAddressChangeProposedIterator struct {
	Event *CSMModuleNodeOperatorManagerAddressChangeProposed // Event containing the contract specifics and raw log

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
func (it *CSMModuleNodeOperatorManagerAddressChangeProposedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleNodeOperatorManagerAddressChangeProposed)
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
		it.Event = new(CSMModuleNodeOperatorManagerAddressChangeProposed)
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
func (it *CSMModuleNodeOperatorManagerAddressChangeProposedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleNodeOperatorManagerAddressChangeProposedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleNodeOperatorManagerAddressChangeProposed represents a NodeOperatorManagerAddressChangeProposed event raised by the CSMModule contract.
type CSMModuleNodeOperatorManagerAddressChangeProposed struct {
	NodeOperatorId     *big.Int
	OldProposedAddress common.Address
	NewProposedAddress common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorManagerAddressChangeProposed is a free log retrieval operation binding the contract event 0x4048f15a706950765ca59f99d0fa6fe8edaaa3f3e3d0337417082e2131df82fb.
//
// Solidity: event NodeOperatorManagerAddressChangeProposed(uint256 indexed nodeOperatorId, address indexed oldProposedAddress, address indexed newProposedAddress)
func (_CSMModule *CSMModuleFilterer) FilterNodeOperatorManagerAddressChangeProposed(opts *bind.FilterOpts, nodeOperatorId []*big.Int, oldProposedAddress []common.Address, newProposedAddress []common.Address) (*CSMModuleNodeOperatorManagerAddressChangeProposedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var oldProposedAddressRule []interface{}
	for _, oldProposedAddressItem := range oldProposedAddress {
		oldProposedAddressRule = append(oldProposedAddressRule, oldProposedAddressItem)
	}
	var newProposedAddressRule []interface{}
	for _, newProposedAddressItem := range newProposedAddress {
		newProposedAddressRule = append(newProposedAddressRule, newProposedAddressItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "NodeOperatorManagerAddressChangeProposed", nodeOperatorIdRule, oldProposedAddressRule, newProposedAddressRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleNodeOperatorManagerAddressChangeProposedIterator{contract: _CSMModule.contract, event: "NodeOperatorManagerAddressChangeProposed", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorManagerAddressChangeProposed is a free log subscription operation binding the contract event 0x4048f15a706950765ca59f99d0fa6fe8edaaa3f3e3d0337417082e2131df82fb.
//
// Solidity: event NodeOperatorManagerAddressChangeProposed(uint256 indexed nodeOperatorId, address indexed oldProposedAddress, address indexed newProposedAddress)
func (_CSMModule *CSMModuleFilterer) WatchNodeOperatorManagerAddressChangeProposed(opts *bind.WatchOpts, sink chan<- *CSMModuleNodeOperatorManagerAddressChangeProposed, nodeOperatorId []*big.Int, oldProposedAddress []common.Address, newProposedAddress []common.Address) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var oldProposedAddressRule []interface{}
	for _, oldProposedAddressItem := range oldProposedAddress {
		oldProposedAddressRule = append(oldProposedAddressRule, oldProposedAddressItem)
	}
	var newProposedAddressRule []interface{}
	for _, newProposedAddressItem := range newProposedAddress {
		newProposedAddressRule = append(newProposedAddressRule, newProposedAddressItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "NodeOperatorManagerAddressChangeProposed", nodeOperatorIdRule, oldProposedAddressRule, newProposedAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleNodeOperatorManagerAddressChangeProposed)
				if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorManagerAddressChangeProposed", log); err != nil {
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

// ParseNodeOperatorManagerAddressChangeProposed is a log parse operation binding the contract event 0x4048f15a706950765ca59f99d0fa6fe8edaaa3f3e3d0337417082e2131df82fb.
//
// Solidity: event NodeOperatorManagerAddressChangeProposed(uint256 indexed nodeOperatorId, address indexed oldProposedAddress, address indexed newProposedAddress)
func (_CSMModule *CSMModuleFilterer) ParseNodeOperatorManagerAddressChangeProposed(log types.Log) (*CSMModuleNodeOperatorManagerAddressChangeProposed, error) {
	event := new(CSMModuleNodeOperatorManagerAddressChangeProposed)
	if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorManagerAddressChangeProposed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleNodeOperatorManagerAddressChangedIterator is returned from FilterNodeOperatorManagerAddressChanged and is used to iterate over the raw logs and unpacked data for NodeOperatorManagerAddressChanged events raised by the CSMModule contract.
type CSMModuleNodeOperatorManagerAddressChangedIterator struct {
	Event *CSMModuleNodeOperatorManagerAddressChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleNodeOperatorManagerAddressChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleNodeOperatorManagerAddressChanged)
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
		it.Event = new(CSMModuleNodeOperatorManagerAddressChanged)
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
func (it *CSMModuleNodeOperatorManagerAddressChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleNodeOperatorManagerAddressChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleNodeOperatorManagerAddressChanged represents a NodeOperatorManagerAddressChanged event raised by the CSMModule contract.
type CSMModuleNodeOperatorManagerAddressChanged struct {
	NodeOperatorId *big.Int
	OldAddress     common.Address
	NewAddress     common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorManagerAddressChanged is a free log retrieval operation binding the contract event 0x862021f23449d6e8516867bd839be15a3d8698a7561c5c2c35069074b7e91e61.
//
// Solidity: event NodeOperatorManagerAddressChanged(uint256 indexed nodeOperatorId, address indexed oldAddress, address indexed newAddress)
func (_CSMModule *CSMModuleFilterer) FilterNodeOperatorManagerAddressChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int, oldAddress []common.Address, newAddress []common.Address) (*CSMModuleNodeOperatorManagerAddressChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "NodeOperatorManagerAddressChanged", nodeOperatorIdRule, oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleNodeOperatorManagerAddressChangedIterator{contract: _CSMModule.contract, event: "NodeOperatorManagerAddressChanged", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorManagerAddressChanged is a free log subscription operation binding the contract event 0x862021f23449d6e8516867bd839be15a3d8698a7561c5c2c35069074b7e91e61.
//
// Solidity: event NodeOperatorManagerAddressChanged(uint256 indexed nodeOperatorId, address indexed oldAddress, address indexed newAddress)
func (_CSMModule *CSMModuleFilterer) WatchNodeOperatorManagerAddressChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleNodeOperatorManagerAddressChanged, nodeOperatorId []*big.Int, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "NodeOperatorManagerAddressChanged", nodeOperatorIdRule, oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleNodeOperatorManagerAddressChanged)
				if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorManagerAddressChanged", log); err != nil {
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

// ParseNodeOperatorManagerAddressChanged is a log parse operation binding the contract event 0x862021f23449d6e8516867bd839be15a3d8698a7561c5c2c35069074b7e91e61.
//
// Solidity: event NodeOperatorManagerAddressChanged(uint256 indexed nodeOperatorId, address indexed oldAddress, address indexed newAddress)
func (_CSMModule *CSMModuleFilterer) ParseNodeOperatorManagerAddressChanged(log types.Log) (*CSMModuleNodeOperatorManagerAddressChanged, error) {
	event := new(CSMModuleNodeOperatorManagerAddressChanged)
	if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorManagerAddressChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleNodeOperatorRewardAddressChangeProposedIterator is returned from FilterNodeOperatorRewardAddressChangeProposed and is used to iterate over the raw logs and unpacked data for NodeOperatorRewardAddressChangeProposed events raised by the CSMModule contract.
type CSMModuleNodeOperatorRewardAddressChangeProposedIterator struct {
	Event *CSMModuleNodeOperatorRewardAddressChangeProposed // Event containing the contract specifics and raw log

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
func (it *CSMModuleNodeOperatorRewardAddressChangeProposedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleNodeOperatorRewardAddressChangeProposed)
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
		it.Event = new(CSMModuleNodeOperatorRewardAddressChangeProposed)
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
func (it *CSMModuleNodeOperatorRewardAddressChangeProposedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleNodeOperatorRewardAddressChangeProposedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleNodeOperatorRewardAddressChangeProposed represents a NodeOperatorRewardAddressChangeProposed event raised by the CSMModule contract.
type CSMModuleNodeOperatorRewardAddressChangeProposed struct {
	NodeOperatorId     *big.Int
	OldProposedAddress common.Address
	NewProposedAddress common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorRewardAddressChangeProposed is a free log retrieval operation binding the contract event 0xb5878cdb1d66f971efe3b138a71c64bc5bc519314db2533e0e4cde954409ea5a.
//
// Solidity: event NodeOperatorRewardAddressChangeProposed(uint256 indexed nodeOperatorId, address indexed oldProposedAddress, address indexed newProposedAddress)
func (_CSMModule *CSMModuleFilterer) FilterNodeOperatorRewardAddressChangeProposed(opts *bind.FilterOpts, nodeOperatorId []*big.Int, oldProposedAddress []common.Address, newProposedAddress []common.Address) (*CSMModuleNodeOperatorRewardAddressChangeProposedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var oldProposedAddressRule []interface{}
	for _, oldProposedAddressItem := range oldProposedAddress {
		oldProposedAddressRule = append(oldProposedAddressRule, oldProposedAddressItem)
	}
	var newProposedAddressRule []interface{}
	for _, newProposedAddressItem := range newProposedAddress {
		newProposedAddressRule = append(newProposedAddressRule, newProposedAddressItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "NodeOperatorRewardAddressChangeProposed", nodeOperatorIdRule, oldProposedAddressRule, newProposedAddressRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleNodeOperatorRewardAddressChangeProposedIterator{contract: _CSMModule.contract, event: "NodeOperatorRewardAddressChangeProposed", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorRewardAddressChangeProposed is a free log subscription operation binding the contract event 0xb5878cdb1d66f971efe3b138a71c64bc5bc519314db2533e0e4cde954409ea5a.
//
// Solidity: event NodeOperatorRewardAddressChangeProposed(uint256 indexed nodeOperatorId, address indexed oldProposedAddress, address indexed newProposedAddress)
func (_CSMModule *CSMModuleFilterer) WatchNodeOperatorRewardAddressChangeProposed(opts *bind.WatchOpts, sink chan<- *CSMModuleNodeOperatorRewardAddressChangeProposed, nodeOperatorId []*big.Int, oldProposedAddress []common.Address, newProposedAddress []common.Address) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var oldProposedAddressRule []interface{}
	for _, oldProposedAddressItem := range oldProposedAddress {
		oldProposedAddressRule = append(oldProposedAddressRule, oldProposedAddressItem)
	}
	var newProposedAddressRule []interface{}
	for _, newProposedAddressItem := range newProposedAddress {
		newProposedAddressRule = append(newProposedAddressRule, newProposedAddressItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "NodeOperatorRewardAddressChangeProposed", nodeOperatorIdRule, oldProposedAddressRule, newProposedAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleNodeOperatorRewardAddressChangeProposed)
				if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorRewardAddressChangeProposed", log); err != nil {
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

// ParseNodeOperatorRewardAddressChangeProposed is a log parse operation binding the contract event 0xb5878cdb1d66f971efe3b138a71c64bc5bc519314db2533e0e4cde954409ea5a.
//
// Solidity: event NodeOperatorRewardAddressChangeProposed(uint256 indexed nodeOperatorId, address indexed oldProposedAddress, address indexed newProposedAddress)
func (_CSMModule *CSMModuleFilterer) ParseNodeOperatorRewardAddressChangeProposed(log types.Log) (*CSMModuleNodeOperatorRewardAddressChangeProposed, error) {
	event := new(CSMModuleNodeOperatorRewardAddressChangeProposed)
	if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorRewardAddressChangeProposed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleNodeOperatorRewardAddressChangedIterator is returned from FilterNodeOperatorRewardAddressChanged and is used to iterate over the raw logs and unpacked data for NodeOperatorRewardAddressChanged events raised by the CSMModule contract.
type CSMModuleNodeOperatorRewardAddressChangedIterator struct {
	Event *CSMModuleNodeOperatorRewardAddressChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleNodeOperatorRewardAddressChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleNodeOperatorRewardAddressChanged)
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
		it.Event = new(CSMModuleNodeOperatorRewardAddressChanged)
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
func (it *CSMModuleNodeOperatorRewardAddressChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleNodeOperatorRewardAddressChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleNodeOperatorRewardAddressChanged represents a NodeOperatorRewardAddressChanged event raised by the CSMModule contract.
type CSMModuleNodeOperatorRewardAddressChanged struct {
	NodeOperatorId *big.Int
	OldAddress     common.Address
	NewAddress     common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNodeOperatorRewardAddressChanged is a free log retrieval operation binding the contract event 0x069ac7cd8230db015b7250c8e5425149cf1a3e912d9569f497165e55b3b6b7b2.
//
// Solidity: event NodeOperatorRewardAddressChanged(uint256 indexed nodeOperatorId, address indexed oldAddress, address indexed newAddress)
func (_CSMModule *CSMModuleFilterer) FilterNodeOperatorRewardAddressChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int, oldAddress []common.Address, newAddress []common.Address) (*CSMModuleNodeOperatorRewardAddressChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "NodeOperatorRewardAddressChanged", nodeOperatorIdRule, oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleNodeOperatorRewardAddressChangedIterator{contract: _CSMModule.contract, event: "NodeOperatorRewardAddressChanged", logs: logs, sub: sub}, nil
}

// WatchNodeOperatorRewardAddressChanged is a free log subscription operation binding the contract event 0x069ac7cd8230db015b7250c8e5425149cf1a3e912d9569f497165e55b3b6b7b2.
//
// Solidity: event NodeOperatorRewardAddressChanged(uint256 indexed nodeOperatorId, address indexed oldAddress, address indexed newAddress)
func (_CSMModule *CSMModuleFilterer) WatchNodeOperatorRewardAddressChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleNodeOperatorRewardAddressChanged, nodeOperatorId []*big.Int, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "NodeOperatorRewardAddressChanged", nodeOperatorIdRule, oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleNodeOperatorRewardAddressChanged)
				if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorRewardAddressChanged", log); err != nil {
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

// ParseNodeOperatorRewardAddressChanged is a log parse operation binding the contract event 0x069ac7cd8230db015b7250c8e5425149cf1a3e912d9569f497165e55b3b6b7b2.
//
// Solidity: event NodeOperatorRewardAddressChanged(uint256 indexed nodeOperatorId, address indexed oldAddress, address indexed newAddress)
func (_CSMModule *CSMModuleFilterer) ParseNodeOperatorRewardAddressChanged(log types.Log) (*CSMModuleNodeOperatorRewardAddressChanged, error) {
	event := new(CSMModuleNodeOperatorRewardAddressChanged)
	if err := _CSMModule.contract.UnpackLog(event, "NodeOperatorRewardAddressChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleNonceChangedIterator is returned from FilterNonceChanged and is used to iterate over the raw logs and unpacked data for NonceChanged events raised by the CSMModule contract.
type CSMModuleNonceChangedIterator struct {
	Event *CSMModuleNonceChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleNonceChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleNonceChanged)
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
		it.Event = new(CSMModuleNonceChanged)
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
func (it *CSMModuleNonceChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleNonceChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleNonceChanged represents a NonceChanged event raised by the CSMModule contract.
type CSMModuleNonceChanged struct {
	Nonce *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNonceChanged is a free log retrieval operation binding the contract event 0x7220970e1f1f12864ecccd8942690a837c7a8dd45d158cb891eb45a8a69134aa.
//
// Solidity: event NonceChanged(uint256 nonce)
func (_CSMModule *CSMModuleFilterer) FilterNonceChanged(opts *bind.FilterOpts) (*CSMModuleNonceChangedIterator, error) {

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "NonceChanged")
	if err != nil {
		return nil, err
	}
	return &CSMModuleNonceChangedIterator{contract: _CSMModule.contract, event: "NonceChanged", logs: logs, sub: sub}, nil
}

// WatchNonceChanged is a free log subscription operation binding the contract event 0x7220970e1f1f12864ecccd8942690a837c7a8dd45d158cb891eb45a8a69134aa.
//
// Solidity: event NonceChanged(uint256 nonce)
func (_CSMModule *CSMModuleFilterer) WatchNonceChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleNonceChanged) (event.Subscription, error) {

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "NonceChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleNonceChanged)
				if err := _CSMModule.contract.UnpackLog(event, "NonceChanged", log); err != nil {
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
func (_CSMModule *CSMModuleFilterer) ParseNonceChanged(log types.Log) (*CSMModuleNonceChanged, error) {
	event := new(CSMModuleNonceChanged)
	if err := _CSMModule.contract.UnpackLog(event, "NonceChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModulePausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the CSMModule contract.
type CSMModulePausedIterator struct {
	Event *CSMModulePaused // Event containing the contract specifics and raw log

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
func (it *CSMModulePausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModulePaused)
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
		it.Event = new(CSMModulePaused)
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
func (it *CSMModulePausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModulePausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModulePaused represents a Paused event raised by the CSMModule contract.
type CSMModulePaused struct {
	Duration *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x32fb7c9891bc4f963c7de9f1186d2a7755c7d6e9f4604dabe1d8bb3027c2f49e.
//
// Solidity: event Paused(uint256 duration)
func (_CSMModule *CSMModuleFilterer) FilterPaused(opts *bind.FilterOpts) (*CSMModulePausedIterator, error) {

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &CSMModulePausedIterator{contract: _CSMModule.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x32fb7c9891bc4f963c7de9f1186d2a7755c7d6e9f4604dabe1d8bb3027c2f49e.
//
// Solidity: event Paused(uint256 duration)
func (_CSMModule *CSMModuleFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *CSMModulePaused) (event.Subscription, error) {

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModulePaused)
				if err := _CSMModule.contract.UnpackLog(event, "Paused", log); err != nil {
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

// ParsePaused is a log parse operation binding the contract event 0x32fb7c9891bc4f963c7de9f1186d2a7755c7d6e9f4604dabe1d8bb3027c2f49e.
//
// Solidity: event Paused(uint256 duration)
func (_CSMModule *CSMModuleFilterer) ParsePaused(log types.Log) (*CSMModulePaused, error) {
	event := new(CSMModulePaused)
	if err := _CSMModule.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModulePublicReleaseIterator is returned from FilterPublicRelease and is used to iterate over the raw logs and unpacked data for PublicRelease events raised by the CSMModule contract.
type CSMModulePublicReleaseIterator struct {
	Event *CSMModulePublicRelease // Event containing the contract specifics and raw log

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
func (it *CSMModulePublicReleaseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModulePublicRelease)
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
		it.Event = new(CSMModulePublicRelease)
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
func (it *CSMModulePublicReleaseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModulePublicReleaseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModulePublicRelease represents a PublicRelease event raised by the CSMModule contract.
type CSMModulePublicRelease struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterPublicRelease is a free log retrieval operation binding the contract event 0xe5eb57aa4d841adeece4ac87bd294965df4a894f0aa24db4a4a55a39ab101d6e.
//
// Solidity: event PublicRelease()
func (_CSMModule *CSMModuleFilterer) FilterPublicRelease(opts *bind.FilterOpts) (*CSMModulePublicReleaseIterator, error) {

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "PublicRelease")
	if err != nil {
		return nil, err
	}
	return &CSMModulePublicReleaseIterator{contract: _CSMModule.contract, event: "PublicRelease", logs: logs, sub: sub}, nil
}

// WatchPublicRelease is a free log subscription operation binding the contract event 0xe5eb57aa4d841adeece4ac87bd294965df4a894f0aa24db4a4a55a39ab101d6e.
//
// Solidity: event PublicRelease()
func (_CSMModule *CSMModuleFilterer) WatchPublicRelease(opts *bind.WatchOpts, sink chan<- *CSMModulePublicRelease) (event.Subscription, error) {

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "PublicRelease")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModulePublicRelease)
				if err := _CSMModule.contract.UnpackLog(event, "PublicRelease", log); err != nil {
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

// ParsePublicRelease is a log parse operation binding the contract event 0xe5eb57aa4d841adeece4ac87bd294965df4a894f0aa24db4a4a55a39ab101d6e.
//
// Solidity: event PublicRelease()
func (_CSMModule *CSMModuleFilterer) ParsePublicRelease(log types.Log) (*CSMModulePublicRelease, error) {
	event := new(CSMModulePublicRelease)
	if err := _CSMModule.contract.UnpackLog(event, "PublicRelease", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleReferrerSetIterator is returned from FilterReferrerSet and is used to iterate over the raw logs and unpacked data for ReferrerSet events raised by the CSMModule contract.
type CSMModuleReferrerSetIterator struct {
	Event *CSMModuleReferrerSet // Event containing the contract specifics and raw log

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
func (it *CSMModuleReferrerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleReferrerSet)
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
		it.Event = new(CSMModuleReferrerSet)
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
func (it *CSMModuleReferrerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleReferrerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleReferrerSet represents a ReferrerSet event raised by the CSMModule contract.
type CSMModuleReferrerSet struct {
	NodeOperatorId *big.Int
	Referrer       common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterReferrerSet is a free log retrieval operation binding the contract event 0x67334334c388385e5f244703f8a8b28b7f4ffe52909130aca69bc62a8e27f09a.
//
// Solidity: event ReferrerSet(uint256 indexed nodeOperatorId, address indexed referrer)
func (_CSMModule *CSMModuleFilterer) FilterReferrerSet(opts *bind.FilterOpts, nodeOperatorId []*big.Int, referrer []common.Address) (*CSMModuleReferrerSetIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var referrerRule []interface{}
	for _, referrerItem := range referrer {
		referrerRule = append(referrerRule, referrerItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "ReferrerSet", nodeOperatorIdRule, referrerRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleReferrerSetIterator{contract: _CSMModule.contract, event: "ReferrerSet", logs: logs, sub: sub}, nil
}

// WatchReferrerSet is a free log subscription operation binding the contract event 0x67334334c388385e5f244703f8a8b28b7f4ffe52909130aca69bc62a8e27f09a.
//
// Solidity: event ReferrerSet(uint256 indexed nodeOperatorId, address indexed referrer)
func (_CSMModule *CSMModuleFilterer) WatchReferrerSet(opts *bind.WatchOpts, sink chan<- *CSMModuleReferrerSet, nodeOperatorId []*big.Int, referrer []common.Address) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}
	var referrerRule []interface{}
	for _, referrerItem := range referrer {
		referrerRule = append(referrerRule, referrerItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "ReferrerSet", nodeOperatorIdRule, referrerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleReferrerSet)
				if err := _CSMModule.contract.UnpackLog(event, "ReferrerSet", log); err != nil {
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

// ParseReferrerSet is a log parse operation binding the contract event 0x67334334c388385e5f244703f8a8b28b7f4ffe52909130aca69bc62a8e27f09a.
//
// Solidity: event ReferrerSet(uint256 indexed nodeOperatorId, address indexed referrer)
func (_CSMModule *CSMModuleFilterer) ParseReferrerSet(log types.Log) (*CSMModuleReferrerSet, error) {
	event := new(CSMModuleReferrerSet)
	if err := _CSMModule.contract.UnpackLog(event, "ReferrerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleResumedIterator is returned from FilterResumed and is used to iterate over the raw logs and unpacked data for Resumed events raised by the CSMModule contract.
type CSMModuleResumedIterator struct {
	Event *CSMModuleResumed // Event containing the contract specifics and raw log

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
func (it *CSMModuleResumedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleResumed)
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
		it.Event = new(CSMModuleResumed)
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
func (it *CSMModuleResumedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleResumedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleResumed represents a Resumed event raised by the CSMModule contract.
type CSMModuleResumed struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterResumed is a free log retrieval operation binding the contract event 0x62451d457bc659158be6e6247f56ec1df424a5c7597f71c20c2bc44e0965c8f9.
//
// Solidity: event Resumed()
func (_CSMModule *CSMModuleFilterer) FilterResumed(opts *bind.FilterOpts) (*CSMModuleResumedIterator, error) {

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "Resumed")
	if err != nil {
		return nil, err
	}
	return &CSMModuleResumedIterator{contract: _CSMModule.contract, event: "Resumed", logs: logs, sub: sub}, nil
}

// WatchResumed is a free log subscription operation binding the contract event 0x62451d457bc659158be6e6247f56ec1df424a5c7597f71c20c2bc44e0965c8f9.
//
// Solidity: event Resumed()
func (_CSMModule *CSMModuleFilterer) WatchResumed(opts *bind.WatchOpts, sink chan<- *CSMModuleResumed) (event.Subscription, error) {

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "Resumed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleResumed)
				if err := _CSMModule.contract.UnpackLog(event, "Resumed", log); err != nil {
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

// ParseResumed is a log parse operation binding the contract event 0x62451d457bc659158be6e6247f56ec1df424a5c7597f71c20c2bc44e0965c8f9.
//
// Solidity: event Resumed()
func (_CSMModule *CSMModuleFilterer) ParseResumed(log types.Log) (*CSMModuleResumed, error) {
	event := new(CSMModuleResumed)
	if err := _CSMModule.contract.UnpackLog(event, "Resumed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the CSMModule contract.
type CSMModuleRoleAdminChangedIterator struct {
	Event *CSMModuleRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleRoleAdminChanged)
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
		it.Event = new(CSMModuleRoleAdminChanged)
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
func (it *CSMModuleRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleRoleAdminChanged represents a RoleAdminChanged event raised by the CSMModule contract.
type CSMModuleRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_CSMModule *CSMModuleFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*CSMModuleRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleRoleAdminChangedIterator{contract: _CSMModule.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_CSMModule *CSMModuleFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleRoleAdminChanged)
				if err := _CSMModule.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_CSMModule *CSMModuleFilterer) ParseRoleAdminChanged(log types.Log) (*CSMModuleRoleAdminChanged, error) {
	event := new(CSMModuleRoleAdminChanged)
	if err := _CSMModule.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the CSMModule contract.
type CSMModuleRoleGrantedIterator struct {
	Event *CSMModuleRoleGranted // Event containing the contract specifics and raw log

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
func (it *CSMModuleRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleRoleGranted)
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
		it.Event = new(CSMModuleRoleGranted)
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
func (it *CSMModuleRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleRoleGranted represents a RoleGranted event raised by the CSMModule contract.
type CSMModuleRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_CSMModule *CSMModuleFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*CSMModuleRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleRoleGrantedIterator{contract: _CSMModule.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_CSMModule *CSMModuleFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *CSMModuleRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleRoleGranted)
				if err := _CSMModule.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_CSMModule *CSMModuleFilterer) ParseRoleGranted(log types.Log) (*CSMModuleRoleGranted, error) {
	event := new(CSMModuleRoleGranted)
	if err := _CSMModule.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the CSMModule contract.
type CSMModuleRoleRevokedIterator struct {
	Event *CSMModuleRoleRevoked // Event containing the contract specifics and raw log

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
func (it *CSMModuleRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleRoleRevoked)
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
		it.Event = new(CSMModuleRoleRevoked)
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
func (it *CSMModuleRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleRoleRevoked represents a RoleRevoked event raised by the CSMModule contract.
type CSMModuleRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_CSMModule *CSMModuleFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*CSMModuleRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleRoleRevokedIterator{contract: _CSMModule.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_CSMModule *CSMModuleFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *CSMModuleRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleRoleRevoked)
				if err := _CSMModule.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_CSMModule *CSMModuleFilterer) ParseRoleRevoked(log types.Log) (*CSMModuleRoleRevoked, error) {
	event := new(CSMModuleRoleRevoked)
	if err := _CSMModule.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleSigningKeyAddedIterator is returned from FilterSigningKeyAdded and is used to iterate over the raw logs and unpacked data for SigningKeyAdded events raised by the CSMModule contract.
type CSMModuleSigningKeyAddedIterator struct {
	Event *CSMModuleSigningKeyAdded // Event containing the contract specifics and raw log

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
func (it *CSMModuleSigningKeyAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleSigningKeyAdded)
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
		it.Event = new(CSMModuleSigningKeyAdded)
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
func (it *CSMModuleSigningKeyAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleSigningKeyAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleSigningKeyAdded represents a SigningKeyAdded event raised by the CSMModule contract.
type CSMModuleSigningKeyAdded struct {
	NodeOperatorId *big.Int
	Pubkey         []byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSigningKeyAdded is a free log retrieval operation binding the contract event 0xc77a17d6b857abe6d6e6c37301621bc72c4dd52fa8830fb54dfa715c04911a89.
//
// Solidity: event SigningKeyAdded(uint256 indexed nodeOperatorId, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) FilterSigningKeyAdded(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleSigningKeyAddedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "SigningKeyAdded", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleSigningKeyAddedIterator{contract: _CSMModule.contract, event: "SigningKeyAdded", logs: logs, sub: sub}, nil
}

// WatchSigningKeyAdded is a free log subscription operation binding the contract event 0xc77a17d6b857abe6d6e6c37301621bc72c4dd52fa8830fb54dfa715c04911a89.
//
// Solidity: event SigningKeyAdded(uint256 indexed nodeOperatorId, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) WatchSigningKeyAdded(opts *bind.WatchOpts, sink chan<- *CSMModuleSigningKeyAdded, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "SigningKeyAdded", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleSigningKeyAdded)
				if err := _CSMModule.contract.UnpackLog(event, "SigningKeyAdded", log); err != nil {
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

// ParseSigningKeyAdded is a log parse operation binding the contract event 0xc77a17d6b857abe6d6e6c37301621bc72c4dd52fa8830fb54dfa715c04911a89.
//
// Solidity: event SigningKeyAdded(uint256 indexed nodeOperatorId, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) ParseSigningKeyAdded(log types.Log) (*CSMModuleSigningKeyAdded, error) {
	event := new(CSMModuleSigningKeyAdded)
	if err := _CSMModule.contract.UnpackLog(event, "SigningKeyAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleSigningKeyRemovedIterator is returned from FilterSigningKeyRemoved and is used to iterate over the raw logs and unpacked data for SigningKeyRemoved events raised by the CSMModule contract.
type CSMModuleSigningKeyRemovedIterator struct {
	Event *CSMModuleSigningKeyRemoved // Event containing the contract specifics and raw log

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
func (it *CSMModuleSigningKeyRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleSigningKeyRemoved)
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
		it.Event = new(CSMModuleSigningKeyRemoved)
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
func (it *CSMModuleSigningKeyRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleSigningKeyRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleSigningKeyRemoved represents a SigningKeyRemoved event raised by the CSMModule contract.
type CSMModuleSigningKeyRemoved struct {
	NodeOperatorId *big.Int
	Pubkey         []byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSigningKeyRemoved is a free log retrieval operation binding the contract event 0xea4b75aaf57196f73d338cadf79ecd0a437902e2dd0d2c4c2cf3ea71b8ab27b9.
//
// Solidity: event SigningKeyRemoved(uint256 indexed nodeOperatorId, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) FilterSigningKeyRemoved(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleSigningKeyRemovedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "SigningKeyRemoved", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleSigningKeyRemovedIterator{contract: _CSMModule.contract, event: "SigningKeyRemoved", logs: logs, sub: sub}, nil
}

// WatchSigningKeyRemoved is a free log subscription operation binding the contract event 0xea4b75aaf57196f73d338cadf79ecd0a437902e2dd0d2c4c2cf3ea71b8ab27b9.
//
// Solidity: event SigningKeyRemoved(uint256 indexed nodeOperatorId, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) WatchSigningKeyRemoved(opts *bind.WatchOpts, sink chan<- *CSMModuleSigningKeyRemoved, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "SigningKeyRemoved", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleSigningKeyRemoved)
				if err := _CSMModule.contract.UnpackLog(event, "SigningKeyRemoved", log); err != nil {
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

// ParseSigningKeyRemoved is a log parse operation binding the contract event 0xea4b75aaf57196f73d338cadf79ecd0a437902e2dd0d2c4c2cf3ea71b8ab27b9.
//
// Solidity: event SigningKeyRemoved(uint256 indexed nodeOperatorId, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) ParseSigningKeyRemoved(log types.Log) (*CSMModuleSigningKeyRemoved, error) {
	event := new(CSMModuleSigningKeyRemoved)
	if err := _CSMModule.contract.UnpackLog(event, "SigningKeyRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleStETHSharesRecoveredIterator is returned from FilterStETHSharesRecovered and is used to iterate over the raw logs and unpacked data for StETHSharesRecovered events raised by the CSMModule contract.
type CSMModuleStETHSharesRecoveredIterator struct {
	Event *CSMModuleStETHSharesRecovered // Event containing the contract specifics and raw log

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
func (it *CSMModuleStETHSharesRecoveredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleStETHSharesRecovered)
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
		it.Event = new(CSMModuleStETHSharesRecovered)
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
func (it *CSMModuleStETHSharesRecoveredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleStETHSharesRecoveredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleStETHSharesRecovered represents a StETHSharesRecovered event raised by the CSMModule contract.
type CSMModuleStETHSharesRecovered struct {
	Recipient common.Address
	Shares    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStETHSharesRecovered is a free log retrieval operation binding the contract event 0x426e7e0100db57255d4af4a46cd49552ef74f5f002bbdc8d4ebb6371c0070a02.
//
// Solidity: event StETHSharesRecovered(address indexed recipient, uint256 shares)
func (_CSMModule *CSMModuleFilterer) FilterStETHSharesRecovered(opts *bind.FilterOpts, recipient []common.Address) (*CSMModuleStETHSharesRecoveredIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "StETHSharesRecovered", recipientRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleStETHSharesRecoveredIterator{contract: _CSMModule.contract, event: "StETHSharesRecovered", logs: logs, sub: sub}, nil
}

// WatchStETHSharesRecovered is a free log subscription operation binding the contract event 0x426e7e0100db57255d4af4a46cd49552ef74f5f002bbdc8d4ebb6371c0070a02.
//
// Solidity: event StETHSharesRecovered(address indexed recipient, uint256 shares)
func (_CSMModule *CSMModuleFilterer) WatchStETHSharesRecovered(opts *bind.WatchOpts, sink chan<- *CSMModuleStETHSharesRecovered, recipient []common.Address) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "StETHSharesRecovered", recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleStETHSharesRecovered)
				if err := _CSMModule.contract.UnpackLog(event, "StETHSharesRecovered", log); err != nil {
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

// ParseStETHSharesRecovered is a log parse operation binding the contract event 0x426e7e0100db57255d4af4a46cd49552ef74f5f002bbdc8d4ebb6371c0070a02.
//
// Solidity: event StETHSharesRecovered(address indexed recipient, uint256 shares)
func (_CSMModule *CSMModuleFilterer) ParseStETHSharesRecovered(log types.Log) (*CSMModuleStETHSharesRecovered, error) {
	event := new(CSMModuleStETHSharesRecovered)
	if err := _CSMModule.contract.UnpackLog(event, "StETHSharesRecovered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleStuckSigningKeysCountChangedIterator is returned from FilterStuckSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for StuckSigningKeysCountChanged events raised by the CSMModule contract.
type CSMModuleStuckSigningKeysCountChangedIterator struct {
	Event *CSMModuleStuckSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleStuckSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleStuckSigningKeysCountChanged)
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
		it.Event = new(CSMModuleStuckSigningKeysCountChanged)
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
func (it *CSMModuleStuckSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleStuckSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleStuckSigningKeysCountChanged represents a StuckSigningKeysCountChanged event raised by the CSMModule contract.
type CSMModuleStuckSigningKeysCountChanged struct {
	NodeOperatorId *big.Int
	StuckKeysCount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterStuckSigningKeysCountChanged is a free log retrieval operation binding the contract event 0xb4f5879eca27b32881cec7907d1310378e9b4c79927062fb7d4a321434b5b04a.
//
// Solidity: event StuckSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 stuckKeysCount)
func (_CSMModule *CSMModuleFilterer) FilterStuckSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleStuckSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "StuckSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleStuckSigningKeysCountChangedIterator{contract: _CSMModule.contract, event: "StuckSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchStuckSigningKeysCountChanged is a free log subscription operation binding the contract event 0xb4f5879eca27b32881cec7907d1310378e9b4c79927062fb7d4a321434b5b04a.
//
// Solidity: event StuckSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 stuckKeysCount)
func (_CSMModule *CSMModuleFilterer) WatchStuckSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleStuckSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "StuckSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleStuckSigningKeysCountChanged)
				if err := _CSMModule.contract.UnpackLog(event, "StuckSigningKeysCountChanged", log); err != nil {
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

// ParseStuckSigningKeysCountChanged is a log parse operation binding the contract event 0xb4f5879eca27b32881cec7907d1310378e9b4c79927062fb7d4a321434b5b04a.
//
// Solidity: event StuckSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 stuckKeysCount)
func (_CSMModule *CSMModuleFilterer) ParseStuckSigningKeysCountChanged(log types.Log) (*CSMModuleStuckSigningKeysCountChanged, error) {
	event := new(CSMModuleStuckSigningKeysCountChanged)
	if err := _CSMModule.contract.UnpackLog(event, "StuckSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleTargetValidatorsCountChangedIterator is returned from FilterTargetValidatorsCountChanged and is used to iterate over the raw logs and unpacked data for TargetValidatorsCountChanged events raised by the CSMModule contract.
type CSMModuleTargetValidatorsCountChangedIterator struct {
	Event *CSMModuleTargetValidatorsCountChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleTargetValidatorsCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleTargetValidatorsCountChanged)
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
		it.Event = new(CSMModuleTargetValidatorsCountChanged)
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
func (it *CSMModuleTargetValidatorsCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleTargetValidatorsCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleTargetValidatorsCountChanged represents a TargetValidatorsCountChanged event raised by the CSMModule contract.
type CSMModuleTargetValidatorsCountChanged struct {
	NodeOperatorId        *big.Int
	TargetLimitMode       *big.Int
	TargetValidatorsCount *big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterTargetValidatorsCountChanged is a free log retrieval operation binding the contract event 0xf92eb109ce5b449e9b121c352c6aeb4319538a90738cb95d84f08e41274e92d2.
//
// Solidity: event TargetValidatorsCountChanged(uint256 indexed nodeOperatorId, uint256 targetLimitMode, uint256 targetValidatorsCount)
func (_CSMModule *CSMModuleFilterer) FilterTargetValidatorsCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleTargetValidatorsCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "TargetValidatorsCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleTargetValidatorsCountChangedIterator{contract: _CSMModule.contract, event: "TargetValidatorsCountChanged", logs: logs, sub: sub}, nil
}

// WatchTargetValidatorsCountChanged is a free log subscription operation binding the contract event 0xf92eb109ce5b449e9b121c352c6aeb4319538a90738cb95d84f08e41274e92d2.
//
// Solidity: event TargetValidatorsCountChanged(uint256 indexed nodeOperatorId, uint256 targetLimitMode, uint256 targetValidatorsCount)
func (_CSMModule *CSMModuleFilterer) WatchTargetValidatorsCountChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleTargetValidatorsCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "TargetValidatorsCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleTargetValidatorsCountChanged)
				if err := _CSMModule.contract.UnpackLog(event, "TargetValidatorsCountChanged", log); err != nil {
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
// Solidity: event TargetValidatorsCountChanged(uint256 indexed nodeOperatorId, uint256 targetLimitMode, uint256 targetValidatorsCount)
func (_CSMModule *CSMModuleFilterer) ParseTargetValidatorsCountChanged(log types.Log) (*CSMModuleTargetValidatorsCountChanged, error) {
	event := new(CSMModuleTargetValidatorsCountChanged)
	if err := _CSMModule.contract.UnpackLog(event, "TargetValidatorsCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleTotalSigningKeysCountChangedIterator is returned from FilterTotalSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for TotalSigningKeysCountChanged events raised by the CSMModule contract.
type CSMModuleTotalSigningKeysCountChangedIterator struct {
	Event *CSMModuleTotalSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleTotalSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleTotalSigningKeysCountChanged)
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
		it.Event = new(CSMModuleTotalSigningKeysCountChanged)
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
func (it *CSMModuleTotalSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleTotalSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleTotalSigningKeysCountChanged represents a TotalSigningKeysCountChanged event raised by the CSMModule contract.
type CSMModuleTotalSigningKeysCountChanged struct {
	NodeOperatorId *big.Int
	TotalKeysCount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterTotalSigningKeysCountChanged is a free log retrieval operation binding the contract event 0xdd01838a366ae4dc9a86e1922512c0716abebc9a440baae0e22d2dec578223f0.
//
// Solidity: event TotalSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 totalKeysCount)
func (_CSMModule *CSMModuleFilterer) FilterTotalSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleTotalSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "TotalSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleTotalSigningKeysCountChangedIterator{contract: _CSMModule.contract, event: "TotalSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchTotalSigningKeysCountChanged is a free log subscription operation binding the contract event 0xdd01838a366ae4dc9a86e1922512c0716abebc9a440baae0e22d2dec578223f0.
//
// Solidity: event TotalSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 totalKeysCount)
func (_CSMModule *CSMModuleFilterer) WatchTotalSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleTotalSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "TotalSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleTotalSigningKeysCountChanged)
				if err := _CSMModule.contract.UnpackLog(event, "TotalSigningKeysCountChanged", log); err != nil {
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
// Solidity: event TotalSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 totalKeysCount)
func (_CSMModule *CSMModuleFilterer) ParseTotalSigningKeysCountChanged(log types.Log) (*CSMModuleTotalSigningKeysCountChanged, error) {
	event := new(CSMModuleTotalSigningKeysCountChanged)
	if err := _CSMModule.contract.UnpackLog(event, "TotalSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleVettedSigningKeysCountChangedIterator is returned from FilterVettedSigningKeysCountChanged and is used to iterate over the raw logs and unpacked data for VettedSigningKeysCountChanged events raised by the CSMModule contract.
type CSMModuleVettedSigningKeysCountChangedIterator struct {
	Event *CSMModuleVettedSigningKeysCountChanged // Event containing the contract specifics and raw log

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
func (it *CSMModuleVettedSigningKeysCountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleVettedSigningKeysCountChanged)
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
		it.Event = new(CSMModuleVettedSigningKeysCountChanged)
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
func (it *CSMModuleVettedSigningKeysCountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleVettedSigningKeysCountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleVettedSigningKeysCountChanged represents a VettedSigningKeysCountChanged event raised by the CSMModule contract.
type CSMModuleVettedSigningKeysCountChanged struct {
	NodeOperatorId  *big.Int
	VettedKeysCount *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterVettedSigningKeysCountChanged is a free log retrieval operation binding the contract event 0x947f955eec7e1f626bee3afd2aa47b5de04ddcdd3fe78dc8838213015ef58dfd.
//
// Solidity: event VettedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 vettedKeysCount)
func (_CSMModule *CSMModuleFilterer) FilterVettedSigningKeysCountChanged(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleVettedSigningKeysCountChangedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "VettedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleVettedSigningKeysCountChangedIterator{contract: _CSMModule.contract, event: "VettedSigningKeysCountChanged", logs: logs, sub: sub}, nil
}

// WatchVettedSigningKeysCountChanged is a free log subscription operation binding the contract event 0x947f955eec7e1f626bee3afd2aa47b5de04ddcdd3fe78dc8838213015ef58dfd.
//
// Solidity: event VettedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 vettedKeysCount)
func (_CSMModule *CSMModuleFilterer) WatchVettedSigningKeysCountChanged(opts *bind.WatchOpts, sink chan<- *CSMModuleVettedSigningKeysCountChanged, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "VettedSigningKeysCountChanged", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleVettedSigningKeysCountChanged)
				if err := _CSMModule.contract.UnpackLog(event, "VettedSigningKeysCountChanged", log); err != nil {
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
// Solidity: event VettedSigningKeysCountChanged(uint256 indexed nodeOperatorId, uint256 vettedKeysCount)
func (_CSMModule *CSMModuleFilterer) ParseVettedSigningKeysCountChanged(log types.Log) (*CSMModuleVettedSigningKeysCountChanged, error) {
	event := new(CSMModuleVettedSigningKeysCountChanged)
	if err := _CSMModule.contract.UnpackLog(event, "VettedSigningKeysCountChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleVettedSigningKeysCountDecreasedIterator is returned from FilterVettedSigningKeysCountDecreased and is used to iterate over the raw logs and unpacked data for VettedSigningKeysCountDecreased events raised by the CSMModule contract.
type CSMModuleVettedSigningKeysCountDecreasedIterator struct {
	Event *CSMModuleVettedSigningKeysCountDecreased // Event containing the contract specifics and raw log

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
func (it *CSMModuleVettedSigningKeysCountDecreasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleVettedSigningKeysCountDecreased)
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
		it.Event = new(CSMModuleVettedSigningKeysCountDecreased)
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
func (it *CSMModuleVettedSigningKeysCountDecreasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleVettedSigningKeysCountDecreasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleVettedSigningKeysCountDecreased represents a VettedSigningKeysCountDecreased event raised by the CSMModule contract.
type CSMModuleVettedSigningKeysCountDecreased struct {
	NodeOperatorId *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterVettedSigningKeysCountDecreased is a free log retrieval operation binding the contract event 0xe5725d045d5c47bd1483feba445e395dc8647486963e6d54aad9ed03ff7d6ce6.
//
// Solidity: event VettedSigningKeysCountDecreased(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) FilterVettedSigningKeysCountDecreased(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleVettedSigningKeysCountDecreasedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "VettedSigningKeysCountDecreased", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleVettedSigningKeysCountDecreasedIterator{contract: _CSMModule.contract, event: "VettedSigningKeysCountDecreased", logs: logs, sub: sub}, nil
}

// WatchVettedSigningKeysCountDecreased is a free log subscription operation binding the contract event 0xe5725d045d5c47bd1483feba445e395dc8647486963e6d54aad9ed03ff7d6ce6.
//
// Solidity: event VettedSigningKeysCountDecreased(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) WatchVettedSigningKeysCountDecreased(opts *bind.WatchOpts, sink chan<- *CSMModuleVettedSigningKeysCountDecreased, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "VettedSigningKeysCountDecreased", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleVettedSigningKeysCountDecreased)
				if err := _CSMModule.contract.UnpackLog(event, "VettedSigningKeysCountDecreased", log); err != nil {
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

// ParseVettedSigningKeysCountDecreased is a log parse operation binding the contract event 0xe5725d045d5c47bd1483feba445e395dc8647486963e6d54aad9ed03ff7d6ce6.
//
// Solidity: event VettedSigningKeysCountDecreased(uint256 indexed nodeOperatorId)
func (_CSMModule *CSMModuleFilterer) ParseVettedSigningKeysCountDecreased(log types.Log) (*CSMModuleVettedSigningKeysCountDecreased, error) {
	event := new(CSMModuleVettedSigningKeysCountDecreased)
	if err := _CSMModule.contract.UnpackLog(event, "VettedSigningKeysCountDecreased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CSMModuleWithdrawalSubmittedIterator is returned from FilterWithdrawalSubmitted and is used to iterate over the raw logs and unpacked data for WithdrawalSubmitted events raised by the CSMModule contract.
type CSMModuleWithdrawalSubmittedIterator struct {
	Event *CSMModuleWithdrawalSubmitted // Event containing the contract specifics and raw log

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
func (it *CSMModuleWithdrawalSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CSMModuleWithdrawalSubmitted)
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
		it.Event = new(CSMModuleWithdrawalSubmitted)
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
func (it *CSMModuleWithdrawalSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CSMModuleWithdrawalSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CSMModuleWithdrawalSubmitted represents a WithdrawalSubmitted event raised by the CSMModule contract.
type CSMModuleWithdrawalSubmitted struct {
	NodeOperatorId *big.Int
	KeyIndex       *big.Int
	Amount         *big.Int
	Pubkey         []byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalSubmitted is a free log retrieval operation binding the contract event 0x9bc54857932b6f10bb750fdad91f736b82dd4de202ed5c2f9f076773bb5b3fb7.
//
// Solidity: event WithdrawalSubmitted(uint256 indexed nodeOperatorId, uint256 keyIndex, uint256 amount, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) FilterWithdrawalSubmitted(opts *bind.FilterOpts, nodeOperatorId []*big.Int) (*CSMModuleWithdrawalSubmittedIterator, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.FilterLogs(opts, "WithdrawalSubmitted", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return &CSMModuleWithdrawalSubmittedIterator{contract: _CSMModule.contract, event: "WithdrawalSubmitted", logs: logs, sub: sub}, nil
}

// WatchWithdrawalSubmitted is a free log subscription operation binding the contract event 0x9bc54857932b6f10bb750fdad91f736b82dd4de202ed5c2f9f076773bb5b3fb7.
//
// Solidity: event WithdrawalSubmitted(uint256 indexed nodeOperatorId, uint256 keyIndex, uint256 amount, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) WatchWithdrawalSubmitted(opts *bind.WatchOpts, sink chan<- *CSMModuleWithdrawalSubmitted, nodeOperatorId []*big.Int) (event.Subscription, error) {

	var nodeOperatorIdRule []interface{}
	for _, nodeOperatorIdItem := range nodeOperatorId {
		nodeOperatorIdRule = append(nodeOperatorIdRule, nodeOperatorIdItem)
	}

	logs, sub, err := _CSMModule.contract.WatchLogs(opts, "WithdrawalSubmitted", nodeOperatorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CSMModuleWithdrawalSubmitted)
				if err := _CSMModule.contract.UnpackLog(event, "WithdrawalSubmitted", log); err != nil {
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

// ParseWithdrawalSubmitted is a log parse operation binding the contract event 0x9bc54857932b6f10bb750fdad91f736b82dd4de202ed5c2f9f076773bb5b3fb7.
//
// Solidity: event WithdrawalSubmitted(uint256 indexed nodeOperatorId, uint256 keyIndex, uint256 amount, bytes pubkey)
func (_CSMModule *CSMModuleFilterer) ParseWithdrawalSubmitted(log types.Log) (*CSMModuleWithdrawalSubmitted, error) {
	event := new(CSMModuleWithdrawalSubmitted)
	if err := _CSMModule.contract.UnpackLog(event, "WithdrawalSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
