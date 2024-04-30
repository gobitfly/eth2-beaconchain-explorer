package ens

import (
	"eth2-exporter/utils"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var ENSRegistryParsedABI, ENSBaseRegistrarParsedABI *abi.ABI

func init() {
	var err error
	ENSRegistryParsedABI, err = ENSRegistryMetaData.GetAbi()
	if err != nil {
		utils.LogFatal(err, "error getting ens-registry-abi", 0)
	}
	ENSBaseRegistrarParsedABI, err = ENSBaseRegistrarMetaData.GetAbi()
	if err != nil {
		utils.LogFatal(err, "error getting ens-registry-abi", 0)
	}
}
