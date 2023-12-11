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
	"net/http"
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
	ensData, err := GetEnsDomain(address)
	if err != nil && utils.IsValidEnsDomain(address) {
		handleNotFoundHtml(w, r)
		return
	}
	if len(ensData.Address) > 0 {
		address = ensData.Address
	}

	isValid := utils.IsEth1Address(address)
	if !isValid {
		handleNotFoundHtml(w, r)
		return
	}

	address = strings.Replace(address, "0x", "", -1)
	address = strings.ToLower(address)

	currency := GetCurrency(r)

	addressBytes := common.FromHex(address)
	data := InitPageData(w, r, "blockchain", "/address", fmt.Sprintf("Address 0x%x", addressBytes), templateFiles)

	metadata, err := db.BigtableClient.GetMetadataForAddress(addressBytes, 0, db.ECR20TokensPerAddressLimit)
	if err != nil {
		logger.Errorf("error retrieving balances for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	g := new(errgroup.Group)
	g.SetLimit(11)

	isContract := false
	txns := &types.DataTableResponse{}
	blobs := &types.DataTableResponse{}
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

		var err error
		isContract, err = eth1data.IsContract(ctx, common.BytesToAddress(addressBytes))
		if err != nil {
			return fmt.Errorf("IsContract: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		txns, err = db.BigtableClient.GetAddressTransactionsTableData(addressBytes, "")
		if err != nil {
			return fmt.Errorf("GetAddressTransactionsTableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		blobs, err = db.BigtableClient.GetAddressBlobTableData(addressBytes, "")
		if err != nil {
			return fmt.Errorf("GetAddressBlobTableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		internal, err = db.BigtableClient.GetAddressInternalTableData(addressBytes, "")
		if err != nil {
			return fmt.Errorf("GetAddressInternalTableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc20, err = db.BigtableClient.GetAddressErc20TableData(addressBytes, "")
		if err != nil {
			return fmt.Errorf("GetAddressErc20TableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc721, err = db.BigtableClient.GetAddressErc721TableData(addressBytes, "")
		if err != nil {
			return fmt.Errorf("GetAddressErc721TableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc1155, err = db.BigtableClient.GetAddressErc1155TableData(addressBytes, "")
		if err != nil {
			return fmt.Errorf("GetAddressErc1155TableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		blocksMined, err = db.BigtableClient.GetAddressBlocksMinedTableData(address, "")
		if err != nil {
			return fmt.Errorf("GetAddressBlocksMinedTableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		unclesMined, err = db.BigtableClient.GetAddressUnclesMinedTableData(address, "")
		if err != nil {
			return fmt.Errorf("GetAddressUnclesMinedTableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		withdrawals, err = db.GetAddressWithdrawalTableData(addressBytes, "", currency)
		if err != nil {
			return fmt.Errorf("GetAddressWithdrawalTableData: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		sumWithdrawals, err := db.GetAddressWithdrawalsTotal(addressBytes)
		if err != nil {
			return fmt.Errorf("GetAddressWithdrawalsTotal: %w", err)
		}
		withdrawalSummary = utils.FormatClCurrency(sumWithdrawals, currency, 6, true, false, false, true)
		return nil
	})

	if err = g.Wait(); err != nil {
		if handleTemplateError(w, r, "eth1Account.go", "Eth1Address", "g.Wait()", err) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	pngStr, pngStrInverse, err := utils.GenerateQRCodeForAddress(addressBytes)
	if err != nil {
		logger.WithError(err).Errorf("error generating qr code for address %v", address)
	}

	tabs := []types.Eth1AddressPageTabs{}

	if blobs != nil && len(blobs.Data) != 0 {
		tabs = append(tabs, types.Eth1AddressPageTabs{
			Id:   "blobTxns",
			Href: "#blobTxns",
			Text: "Blob Txns",
			Data: blobs,
		})
	}
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
		EnsName:            ensData.Domain,
		IsContract:         isContract,
		QRCode:             pngStr,
		QRCodeInverse:      pngStrInverse,
		Metadata:           metadata,
		WithdrawalsSummary: withdrawalSummary,
		TransactionsTable:  txns,
		BlobTxnsTable:      blobs,
		InternalTxnsTable:  internal,
		Erc20Table:         erc20,
		Erc721Table:        erc721,
		Erc1155Table:       erc1155,
		WithdrawalsTable:   withdrawals,
		BlocksMinedTable:   blocksMined,
		UnclesMinedTable:   unclesMined,
		EtherValue:         utils.FormatPricedValue(utils.WeiBytesToEther(metadata.EthBalance.Balance), utils.Config.Frontend.ElCurrency, currency),
		Tabs:               tabs,
	}

	if handleTemplateError(w, r, "eth1Account.go", "Eth1Address", "Done", eth1AddressTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func Eth1AddressTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}
	addressBytes := common.FromHex(address)

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	pageToken := q.Get("pageToken")

	data, err := db.BigtableClient.GetAddressTransactionsTableData(addressBytes, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 tx table data", 0, errFields)
	}

	err = json.NewEncoder(w).Encode(data)

	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func Eth1AddressBlocksMined(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	pageToken := q.Get("pageToken")

	data, err := db.BigtableClient.GetAddressBlocksMinedTableData(address, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 blocks mined table data", 0, errFields)
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func Eth1AddressUnclesMined(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	pageToken := q.Get("pageToken")

	data, err := db.BigtableClient.GetAddressUnclesMinedTableData(address, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 uncles mined data", 0, errFields)
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func Eth1AddressWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	currency := GetCurrency(r)
	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	data, err := db.GetAddressWithdrawalTableData(common.HexToAddress(address).Bytes(), q.Get("pageToken"), currency)
	if err != nil {
		utils.LogError(err, "error getting address withdrawals data", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func Eth1AddressBlobTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}
	addressBytes := common.FromHex(address)

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	pageToken := q.Get("pageToken")
	data, err := db.BigtableClient.GetAddressBlobTableData(addressBytes, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 blob table data", 0, errFields)
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func Eth1AddressInternalTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}
	addressBytes := common.FromHex(address)

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	pageToken := q.Get("pageToken")
	data, err := db.BigtableClient.GetAddressInternalTableData(addressBytes, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 internal tx table data", 0, errFields)
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func Eth1AddressErc20Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}
	addressBytes := common.FromHex(address)

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	pageToken := q.Get("pageToken")
	data, err := db.BigtableClient.GetAddressErc20TableData(addressBytes, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 ERC20 transactions table data", 0, errFields)
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func Eth1AddressErc721Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}
	addressBytes := common.FromHex(address)

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	pageToken := q.Get("pageToken")
	data, err := db.BigtableClient.GetAddressErc721TableData(addressBytes, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 ERC721 transactions table data", 0, errFields)
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func Eth1AddressErc1155Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	address, err := lowerAddressFromRequest(w, r)
	if err != nil {
		return
	}
	addressBytes := common.FromHex(address)

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	pageToken := q.Get("pageToken")
	data, err := db.BigtableClient.GetAddressErc1155TableData(addressBytes, pageToken)
	if err != nil {
		utils.LogError(err, "error getting eth1 ERC1155 transactions table data", 0, errFields)
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// takes the "address" parameter from the request and transforms it to lower case. The ENS name can be used instead of the address
func lowerAddressFromRequest(w http.ResponseWriter, r *http.Request) (string, error) {
	vars := mux.Vars(r)
	address := vars["address"]
	if utils.IsValidEnsDomain(address) {
		ensData, err := GetEnsDomain(address)
		if err != nil {
			handleNotFoundJson(address, w, r, err)
			return "", err
		}
		if len(ensData.Address) > 0 {
			address = ensData.Address
		}
	}
	return strings.ToLower(strings.Replace(address, "0x", "", -1)), nil
}

func handleNotFoundJson(address string, w http.ResponseWriter, r *http.Request, err error) {
	logger.Errorf("error getting address for ENS name [%v] not found for %v route: %v", address, r.URL.String(), err)
	http.Error(w, "Invalid ENS name", http.StatusInternalServerError)
}

func handleNotFoundHtml(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "sprites.html", "execution/addressNotFound.html")
	data := InitPageData(w, r, "blockchain", "/address", "not found", templateFiles)

	if handleTemplateError(w, r, "eth1Account.go", "Eth1Address", "not valid", templates.GetTemplate(templateFiles...).ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
