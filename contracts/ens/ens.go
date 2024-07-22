package ens

import (
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var ENSCrontractAddressesEthereum = map[string]string{
	"0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e": "Registry",
	"0x253553366Da8546fC250F225fe3d25d0C782303b": "ETHRegistrarController",
	"0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5": "OldEnsRegistrarController",
}

var ENSCrontractAddressesHolesky = map[string]string{
	"0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e": "Registry",
	"0x179Be112b24Ad4cFC392eF8924DfA08C20Ad8583": "ETHRegistrarController",
	"0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5": "OldEnsRegistrarController",
}

var ENSCrontractAddressesSepolia = map[string]string{
	"0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e": "Registry",
	"0xFED6a969AaA60E4961FCD3EBF1A2e8913ac65B72": "ETHRegistrarController",
	"0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5": "OldEnsRegistrarController",
}

var ENSRegistryParsedABI, ENSBaseRegistrarParsedABI, ENSOldRegistrarControllerParsedABI, ENSPublicResolverParsedABI, ENSETHRegistrarControllerParsedABI *abi.ABI

var ENSRegistryContract, ENSBaseRegistrarContract, ENSOldRegistrarControllerContract, ENSPublicResolverContract, ENSETHRegistrarControllerContract *bind.BoundContract

func init() {
	var err error

	ENSRegistryParsedABI, err = ENSRegistryMetaData.GetAbi()
	if err != nil {
		utils.LogFatal(err, "error getting ens-registry-abi", 0)
	}
	ENSRegistryParsedABI, err = ENSRegistryMetaData.GetAbi()
	if err != nil {
		utils.LogFatal(err, "error getting ens-registry-abi", 0)
	}
	ENSBaseRegistrarParsedABI, err = ENSBaseRegistrarMetaData.GetAbi()
	if err != nil {
		utils.LogFatal(err, "error getting ens-base-regsitrar-abi", 0)
	}
	ENSOldRegistrarControllerParsedABI, err = ENSOldRegistrarControllerMetaData.GetAbi()
	if err != nil {
		utils.LogFatal(err, "error getting ens-old-registrar-controller-abi", 0)
	}
	ENSPublicResolverParsedABI, err = ENSPublicResolverMetaData.GetAbi()
	if err != nil {
		utils.LogFatal(err, "error getting ens-public-resolver-abi", 0)
	}
	ENSETHRegistrarControllerParsedABI, err = ENSETHRegistrarControllerMetaData.GetAbi()
	if err != nil {
		utils.LogFatal(err, "error getting ens-eth-registrar-controller-abi", 0)
	}

	ENSRegistryContract = bind.NewBoundContract(common.Address{}, *ENSRegistryParsedABI, nil, nil, nil)
	if err != nil {
		utils.LogFatal(err, "error creating ens-registry-contract", 0)
	}
	ENSBaseRegistrarContract = bind.NewBoundContract(common.Address{}, *ENSBaseRegistrarParsedABI, nil, nil, nil)
	if err != nil {
		utils.LogFatal(err, "error creating ens-base-registrar-contract", 0)
	}
	ENSOldRegistrarControllerContract = bind.NewBoundContract(common.Address{}, *ENSOldRegistrarControllerParsedABI, nil, nil, nil)
	if err != nil {
		utils.LogFatal(err, "error creating ens-old-registrar-controller-contract", 0)
	}
	ENSPublicResolverContract = bind.NewBoundContract(common.Address{}, *ENSPublicResolverParsedABI, nil, nil, nil)
	if err != nil {
		utils.LogFatal(err, "error creating ens-public-resolver-contract", 0)
	}
	ENSETHRegistrarControllerContract = bind.NewBoundContract(common.Address{}, *ENSETHRegistrarControllerParsedABI, nil, nil, nil)
	if err != nil {
		utils.LogFatal(err, "error creating ens-eth-registrar-controller-contract", 0)
	}
}
