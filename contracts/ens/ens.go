package ens

import (
	"eth2-exporter/utils"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

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
