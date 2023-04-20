package handlers

import (
	"context"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/eth1data"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
)

func Eth1Address(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "sprites.html", "execution/address.html")
	var eth1AddressTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	address := template.HTMLEscapeString(vars["address"])
	isValid := utils.IsEth1Address(address)
	if !isValid {
		templateFiles = append(layoutTemplateFiles, "sprites.html", "execution/addressNotFound.html")
		data := InitPageData(w, r, "blockchain", "/address", "not found", templateFiles)

		if handleTemplateError(w, r, "eth1Account.go", "Eth1Address", "not valid", templates.GetTemplate(templateFiles...).ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	address = strings.Replace(address, "0x", "", -1)
	address = strings.ToLower(address)

	// currency := GetCurrency(r)
	price := GetCurrentPrice(r)
	symbol := GetCurrencySymbol(r)

	addressBytes := common.FromHex(address)
	data := InitPageData(w, r, "blockchain", "/address", fmt.Sprintf("Address 0x%x", addressBytes), templateFiles)

	metadata, err := db.BigtableClient.GetMetadataForAddress(addressBytes)
	if err != nil {
		logger.Errorf("error retieving balances for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	g := new(errgroup.Group)
	g.SetLimit(9)

	isContract := false
	txns := &types.DataTableResponse{}
	internal := &types.DataTableResponse{}
	erc20 := &types.DataTableResponse{}
	erc721 := &types.DataTableResponse{}
	erc1155 := &types.DataTableResponse{}
	blocksMined := &types.DataTableResponse{}
	unclesMined := &types.DataTableResponse{}
	withdrawals := &types.DataTableResponse{}
	withdrawalSummary := template.HTML("0")

	g.Go(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		isContract, err = eth1data.IsContract(ctx, common.BytesToAddress(addressBytes))
		return err
	})
	g.Go(func() error {
		var err error
		txns, err = db.BigtableClient.GetAddressTransactionsTableData(addressBytes, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	// if !utils.Config.Frontend.Debug {
	g.Go(func() error {
		var err error
		internal, err = db.BigtableClient.GetAddressInternalTableData(addressBytes, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc20, err = db.BigtableClient.GetAddressErc20TableData(addressBytes, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc721, err = db.BigtableClient.GetAddressErc721TableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc1155, err = db.BigtableClient.GetAddressErc1155TableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		blocksMined, err = db.BigtableClient.GetAddressBlocksMinedTableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		unclesMined, err = db.BigtableClient.GetAddressUnclesMinedTableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		addressWithdrawals, err := db.GetAddressWithdrawals(addressBytes, 25, 0)
		if err != nil {
			return err
		}

		withdrawalsData := make([][]interface{}, 0, len(addressWithdrawals))
		for _, w := range addressWithdrawals {
			withdrawalsData = append(withdrawalsData, []interface{}{
				template.HTML(fmt.Sprintf("%v", utils.FormatEpoch(utils.EpochOfSlot(w.Slot)))),
				template.HTML(fmt.Sprintf("%v", utils.FormatBlockSlot(w.Slot))),
				template.HTML(fmt.Sprintf("%v", utils.FormatTimeFromNow(utils.SlotToTime(w.Slot)))),
				template.HTML(fmt.Sprintf("%v", utils.FormatValidator(w.ValidatorIndex))),
				template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), "Ether", 6))),
			})
		}

		withdrawals = &types.DataTableResponse{
			Draw:         1,
			RecordsTotal: uint64(len(withdrawalsData)),
			// RecordsFiltered: uint64(len(withdrawals)),
			Data:        withdrawalsData,
			PagingToken: fmt.Sprintf("%v", 25),
		}

		return nil
	})
	g.Go(func() error {
		sumWithdrawals, err := db.GetAddressWithdrawalsTotal(addressBytes)
		if err != nil {
			return err
		}
		withdrawalSummary = template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(sumWithdrawals), big.NewInt(1e9)), "Ether", 6)))
		return nil
	})
	// }

	if err := g.Wait(); err != nil {
		if handleTemplateError(w, r, "eth1Account.go", "Eth1Address", "g.Wait()", err) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	pngStr, pngStrInverse, err := utils.GenerateQRCodeForAddress(addressBytes)
	if err != nil {
		logger.WithError(err).Errorf("error generating qr code for address %v", address)
	}

	ef := new(big.Float).SetInt(new(big.Int).SetBytes(metadata.EthBalance.Balance))
	etherBalance := new(big.Float).Quo(ef, big.NewFloat(1e18))
	ethPrice := new(big.Float).Mul(etherBalance, big.NewFloat(float64(price)))
	tabs := []types.Eth1AddressPageTabs{}

	// if txns != nil && len(txns.Data) != 0 {
	// 	tabs = append(tabs, types.Eth1AddressPageTabs{
	// 		Id:   "transactions",
	// 		Href: "#transactions",
	// 		Text: "Transactions",
	// 	})
	// }
	if internal != nil && len(internal.Data) != 0 {
		tabs = append(tabs, types.Eth1AddressPageTabs{
			Id:   "internalTxns",
			Href: "#internalTxns",
			Text: "Internal Txns",
			Data: internal,
		})
	}
	if erc20 != nil && len(erc20.Data) != 0 {
		tabs = append(tabs, types.Eth1AddressPageTabs{
			Id:   "erc20Txns",
			Href: "#erc20Txns",
			Text: "Erc20 Token Txns",
			Data: erc20,
		})
	}
	if erc721 != nil && len(erc721.Data) != 0 {
		tabs = append(tabs, types.Eth1AddressPageTabs{
			Id:   "erc721Txns",
			Href: "#erc721Txns",
			Text: "Erc721 Token Txns",
			Data: erc721,
		})
	}
	if blocksMined != nil && len(blocksMined.Data) != 0 {
		tabs = append(tabs, types.Eth1AddressPageTabs{
			Id:   "blocks",
			Href: "#blocks",
			Text: "Produced Blocks",
			Data: blocksMined,
		})
	}
	if unclesMined != nil && len(unclesMined.Data) != 0 {
		tabs = append(tabs, types.Eth1AddressPageTabs{
			Id:   "uncles",
			Href: "#uncles",
			Text: "Produced Uncles",
			Data: unclesMined,
		})
	}
	if erc1155 != nil && len(erc1155.Data) != 0 {
		tabs = append(tabs, types.Eth1AddressPageTabs{
			Id:   "erc1155Txns",
			Href: "#erc1155Txns",
			Text: "Erc1155 Token Txns",
			Data: erc1155,
		})
	}

	if withdrawals != nil && len(withdrawals.Data) != 0 {
		tabs = append(tabs, types.Eth1AddressPageTabs{
			Id:   "withdrawals",
			Href: "#withdrawals",
			Text: "Withdrawals",
			Data: withdrawals,
		})
	}

	data.Data = types.Eth1AddressPageData{
		Address:            address,
		IsContract:         isContract,
		QRCode:             pngStr,
		QRCodeInverse:      pngStrInverse,
		Metadata:           metadata,
		WithdrawalsSummary: withdrawalSummary,
		TransactionsTable:  txns,
		InternalTxnsTable:  internal,
		Erc20Table:         erc20,
		Erc721Table:        erc721,
		Erc1155Table:       erc1155,
		WithdrawalsTable:   withdrawals,
		BlocksMinedTable:   blocksMined,
		UnclesMinedTable:   unclesMined,
		EtherValue:         utils.FormatEtherValue(symbol, ethPrice, GetCurrentPriceFormatted(r)),
		Tabs:               tabs,
	}

	if handleTemplateError(w, r, "eth1Account.go", "Eth1Address", "Done", eth1AddressTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func Eth1AddressTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)
	addressBytes := common.FromHex(address)

	pageToken := q.Get("pageToken")

	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressTransactionsTableData(addressBytes, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressBlocksMined(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	pageToken := q.Get("pageToken")

	search := ""
	data, err := db.BigtableClient.GetAddressBlocksMinedTableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressUnclesMined(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	pageToken := q.Get("pageToken")

	search := ""
	data, err := db.BigtableClient.GetAddressUnclesMinedTableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	pageToken, err := strconv.ParseUint(q.Get("pageToken"), 10, 64)
	if err != nil {
		logger.WithError(err).Errorf("error parsing page token")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	withdrawals, err := db.GetAddressWithdrawals(common.HexToAddress(address).Bytes(), 25, uint64(pageToken))
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	tableData := make([][]interface{}, len(withdrawals))
	for i, w := range withdrawals {
		tableData[i] = []interface{}{
			template.HTML(fmt.Sprintf("%v", utils.FormatEpoch(utils.EpochOfSlot(w.Slot)))),
			template.HTML(fmt.Sprintf("%v", utils.FormatBlockSlot(w.Slot))),
			template.HTML(fmt.Sprintf("%v", utils.FormatTimeFromNow(utils.SlotToTime(w.Slot)))),
			template.HTML(fmt.Sprintf("%v", utils.FormatValidator(w.ValidatorIndex))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), "Ether", 6))),
		}
	}

	nextPageToken := pageToken + 25
	if len(withdrawals) < 25 {
		nextPageToken = 0
	}

	next := ""
	if nextPageToken != 0 {
		next = fmt.Sprintf("%d", nextPageToken)
	}

	data := &types.DataTableResponse{
		// Draw: draw,
		// RecordsTotal:    ,
		// RecordsFiltered: ,
		Data:        tableData,
		PagingToken: next,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressInternalTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)
	addressBytes := common.FromHex(address)

	pageToken := q.Get("pageToken")

	search := ""

	data, err := db.BigtableClient.GetAddressInternalTableData(addressBytes, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressErc20Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	addressBytes := common.FromHex(address)
	pageToken := q.Get("pageToken")

	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressErc20TableData(addressBytes, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 internal transactions table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressErc721Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	pageToken := q.Get("pageToken")
	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressErc721TableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressErc1155Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)
	pageToken := q.Get("pageToken")

	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressErc1155TableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 internal transactions table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
