package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func Eth1Token(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "execution/token.html")
	var eth1TokenTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	token := common.FromHex(strings.TrimPrefix(vars["token"], "0x"))

	address := common.FromHex(strings.TrimPrefix(r.URL.Query().Get("a"), "0x"))

	// priceEth := GetCurrentPrice(r)
	// symbol := GetCurrencySymbol(r)

	g := new(errgroup.Group)
	g.SetLimit(3)

	var txns *types.DataTableResponse
	var metadata *types.ERC20Metadata
	var balance *types.Eth1AddressBalance
	// var holders *types.DataTableResponse

	g.Go(func() error {
		var err error
		txns, err = db.BigtableClient.GetTokenTransactionsTableData(token, address, "")
		return err
	})

	g.Go(func() error {
		var err error
		metadata, err = db.BigtableClient.GetERC20MetadataForAddress(token)
		return err
	})

	if address != nil {
		g.Go(func() error {
			var err error
			balance, err = db.BigtableClient.GetBalanceForAddress(address, token)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		if handleTemplateError(w, r, "eth1Token.go", "Eth1Token", "g.Wait()", err) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	pngStr, pngStrInverse, err := utils.GenerateQRCodeForAddress(token)
	if err != nil {
		logger.WithError(err).Errorf("error generating qr code for address %v", token)
	}

	data := InitPageData(w, r, "blockchain", "/token", fmt.Sprintf("Token 0x%x", token), templateFiles)

	tokenDecimals := decimal.NewFromBigInt(new(big.Int).SetBytes(metadata.Decimals), 0)

	ethDiv := decimal.NewFromInt(utils.Config.Frontend.ElCurrencyDivisor)
	tokenDiv := decimal.NewFromInt(10).Pow(tokenDecimals)

	_ = ethDiv
	_ = tokenDiv

	ethPriceUsd := decimal.NewFromFloat(price.GetPrice(utils.Config.Frontend.ElCurrency, "USD"))
	tokenPriceEth := decimal.NewFromBigInt(new(big.Int).SetBytes(metadata.Price), 0).DivRound(ethDiv, 18)
	tokenPriceUsd := ethPriceUsd.Mul(tokenPriceEth).Mul(tokenDiv).DivRound(ethDiv, 18)
	tokenSupply := decimal.NewFromBigInt(new(big.Int).SetBytes(metadata.TotalSupply), 0).DivRound(tokenDiv, 18)
	tokenMarketCapUsd := tokenPriceUsd.Mul(tokenSupply)

	data.Data = types.Eth1TokenPageData{
		Token:          fmt.Sprintf("%x", token),
		Address:        fmt.Sprintf("%x", address),
		TransfersTable: txns,
		Metadata:       metadata,
		Balance:        balance,
		QRCode:         pngStr,
		QRCodeInverse:  pngStrInverse,
		MarketCap:      template.HTML("$" + utils.FormatThousandsEnglish(tokenMarketCapUsd.StringFixed(2))),
		Supply:         template.HTML(utils.FormatThousandsEnglish(tokenSupply.StringFixed(6))),
		Price:          template.HTML("$" + utils.FormatThousandsEnglish(tokenPriceUsd.StringFixed(6))),
	}

	if handleTemplateError(w, r, "eth1Token.go", "Eth1Token", "Done", eth1TokenTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func Eth1TokenTransfers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)

	token := common.FromHex(strings.TrimPrefix(vars["token"], "0x"))
	address := common.FromHex(strings.TrimPrefix(q.Get("a"), "0x"))
	pageToken := q.Get("pageToken")

	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetTokenTransactionsTableData(token, address, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 block table data", 0)
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
