package erc20

import (
	"encoding/json"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

var ERC20Abi, _ = abi.JSON(strings.NewReader(Erc20ABI))

var TransferTopic []byte = []byte{0xdd, 0xf2, 0x52, 0xad, 0x1b, 0xe2, 0xc8, 0x9b, 0x69, 0xc2, 0xb0, 0x68, 0xfc, 0x37, 0x8d, 0xaa, 0x95, 0x2b, 0xa7, 0xf1, 0x63, 0xc4, 0xa1, 0x16, 0x28, 0xf5, 0x5a, 0x4d, 0xf5, 0x23, 0xb3, 0xef}

var tokenMap = make(map[string]*ERC20TokenDetail)

var logger = logrus.StandardLogger().WithField("module", "erc20")

func InitTokenList(path string) {
	body, err := os.ReadFile(path)
	if err != nil {
		utils.LogFatal(err, "unable to retrieve erc20 token list", 0)
	}
	TokenList := &ERC20TokenList{}

	err = json.Unmarshal(body, TokenList)
	if err != nil {
		logger.Fatalf("unable to parse erc20 token list: %v", err)
	}

	for _, token := range TokenList.Tokens {
		address := strings.Replace(token.Address, "0x", "", -1)
		address = strings.ToLower(address)
		tokenMap[address] = token
		// logger.Info(address)
	}

}

func GetTokenDetail(address string) *ERC20TokenDetail {
	return tokenMap[address]
}

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
	Address  string `json:"address"`
	Owner    string `json:"-"`
	ChainID  int64  `json:"chainId"`
	Decimals int64  `json:"decimals"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Divider  *big.Int
	Contract *Erc20
}

func (td *ERC20TokenDetail) FormatAmount(in *big.Int) string {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(td.Decimals)))
	num := decimal.NewFromBigInt(in, 0)
	result := num.Div(mul)

	return fmt.Sprintf("%v", result)
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
