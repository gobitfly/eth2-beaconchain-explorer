package erc20

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

var ERC20Abi, _ = abi.JSON(strings.NewReader(Erc20ABI))

type ERC20TokenList struct {
	Keywords  []string            `json:"keywords"`
	LogoURI   string              `json:"logoURI"`
	Name      string              `json:"name"`
	Timestamp string              `json:"timestamp"`
	Tokens    []*ERC20TokenDetail `json:"tokens"`
	Version   struct {
		Major int64 `json:"major"`
		Minor int64 `json:"minor"`
		Patch int64 `json:"patch"`
	} `json:"version"`
}

type ERC20TokenDetail struct {
	AddressStr string         `json:"address"`
	Address    common.Address `json:"-"`
	Owner      string         `json:"-"`
	ChainID    int64          `json:"chainId"`
	Decimals   int64          `json:"decimals"`
	Name       string         `json:"name"`
	Symbol     string         `json:"symbol"`
	Divider    *big.Int
	Contract   *Erc20
}

func (td *ERC20TokenDetail) FormatAmount(in *big.Int) string {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.New(td.Decimals, 1))
	num := decimal.NewFromBigInt(in, 0)
	result := num.Div(mul)

	return fmt.Sprintf("%v %v", result, td.Symbol)
}

func (td *ERC20TokenDetail) FormatAmountFloat(in *big.Int) float64 {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.New((td.Decimals), 1))
	num := decimal.NewFromBigInt(in, 0)
	result := num.Div(mul)

	f, _ := result.Float64()
	return f
}

func (td *ERC20TokenDetail) ToScaled(in *big.Int) decimal.Decimal {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.New((td.Decimals), 1))
	num := decimal.NewFromBigInt(in, 0)
	result := num.Div(mul)

	return result
}

func (td *ERC20TokenDetail) RawAmount(in float64) *big.Int {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.New((td.Decimals), 1))
	res, _ := new(big.Int).SetString(decimal.NewFromFloat(in).Mul(mul).String(), 10)
	return res
}
