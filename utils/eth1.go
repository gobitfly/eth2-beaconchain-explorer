package utils

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var Erc20TransferEventHash = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

func Eth1BlockReward(blockNumber uint64) *big.Int {
	if blockNumber < 4370000 {
		return big.NewInt(5e+18)
	} else if blockNumber < 7280000 {
		return big.NewInt(3e+18)
	} else {
		return big.NewInt(2e+18)
	}
}

func StripPrefix(hexStr string) string {
	return strings.Replace(hexStr, "0x", "", 1)
}

func EthBytesToFloat(b []byte) float64 {
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(b)), new(big.Float).SetInt(big.NewInt(params.Ether))).Float64()
	return f
}
