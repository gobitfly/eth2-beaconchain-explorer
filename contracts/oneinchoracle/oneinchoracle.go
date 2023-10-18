package oneinchoracle

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

//go:generate abigen -abi abi.json -out bindings.go -pkg oneinchoracle -type OneinchOracle

// see: https://docs.1inch.io/docs/spot-price-aggregator/introduction/

var OracleAddressesByChainID = map[string]common.Address{
	"1":   common.HexToAddress("0x3E1Fe1Bd5a5560972bFa2D393b9aC18aF279fF56"),
	"100": common.HexToAddress("0x3Ce81621e674Db129033548CbB9FF31AEDCc1BF6"),
}

func SupportedChainId(chainIDif interface{}) bool {
	_, exists := OracleAddressesByChainID[fmt.Sprintf("%v", chainIDif)]
	return exists
}

func NewOneInchOracleByChainID(chainIDif interface{}, backend bind.ContractBackend) (*OneinchOracle, error) {
	chainID := fmt.Sprintf("%v", chainIDif)
	addr, exists := OracleAddressesByChainID[chainID]
	if !exists {
		return nil, fmt.Errorf("unsupported chainID: %v", chainID)
	}
	return NewOneinchOracle(addr, backend)
}
