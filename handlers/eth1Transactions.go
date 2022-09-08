package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"
)

const (
	minimumTransactionsPerUpdate = 25
)

var eth1TransactionsTemplate = template.Must(template.New("transactions").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/transactions.html"))

func Eth1Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "eth1transactions", "/eth1transactions", "eth1transactions")
	data.Data = getTransactionDataStartingWithPageToken("")

	// Internal Transx
	// Mindesthöhe von den Dingern?!?
	// Mindest Update sonst gehts iwie ned gut

	if utils.Config.Frontend.Debug {
		eth1TransactionsTemplate = template.Must(template.New("transactions").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/transactions.html"))
	}

	err := eth1TransactionsTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1TransactionsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(getTransactionDataStartingWithPageToken(r.URL.Query().Get("pageToken")))
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func getTransactionDataStartingWithPageToken(pageToken string) *types.DataTableResponse {
	pageTokenId := uint64(0)
	{
		if len(pageToken) > 0 {
			v, err := strconv.ParseUint(pageToken, 10, 64)
			if err == nil && v > 0 {
				pageTokenId = v
			}
		}
	}
	if pageTokenId == 0 {
		pageTokenId = services.LatestEth1BlockNumber()
	}

	tableData := make([][]interface{}, 0, minimumTransactionsPerUpdate)
	for len(tableData) < minimumTransactionsPerUpdate && pageTokenId != 0 {
		t, n, err := getEth1TransactionFromBlock(pageTokenId)
		if err != nil {
			logger.Errorf("error getting transaction from block %v", err)
			return nil
		}

		// retrieve metadata
		names := make(map[string]string)
		{
			for _, v := range *t {
				names[string(v.GetFrom())] = ""
				names[string(v.GetTo())] = ""
			}
			/**
			g := new(errgroup.Group)
			g.SetLimit(25)
			mux := sync.Mutex{}
			for address := range names {
				address := address
				g.Go(func() error {
					name, err := db.BigtableClient.GetAddressName([]byte(address))
					if err != nil {
						logger.Errorf("error getting name for address '%s': %v", address, err)
						return nil
					}
					mux.Lock()
					names[address] = name
					mux.Unlock()
					return nil
				})
			}

			err := g.Wait()
			if err != nil {
				logger.Errorf("error waiting for threads collecting address names: %v", err)
				return nil
			}/**/
		}

		for _, v := range *t {
			method := "Transfer #missing" // #RECY #TODO // "Transfer"
			/* if len(v.MethodId) > 0 {

				if v.InvokesContract {
					method = fmt.Sprintf("0x%x", v.MethodId)
				} else {
					method = "Transfer*"
				}
			}/**/

			tableData = append(tableData, []interface{}{
				utils.FormatAddressWithLimits(v.GetHash(), "", "tx", 15, 18, true),
				utils.FormatMethod(method),
				"block #todo",
				"time #todo",
				utils.FormatAddressWithLimits(v.From, names[string(v.GetFrom())], "address", 15, 18, true),
				utils.FormatAddressWithLimits(v.To, names[string(v.GetTo())], "address", 15, 18, true),
				utils.FormatAmountFormated(new(big.Int).SetBytes(v.GetValue()), "ETH", 8, 3, true, true, true),
				"fee #missing",
			})
		}

		pageTokenId = n
	}

	return &types.DataTableResponse{
		Data:        tableData,
		PagingToken: fmt.Sprintf("%d", pageTokenId),
	}
}

// Return transactions from given block, next block number and error
// If block doesn't exists nil, 0, nil is returned
func getEth1TransactionFromBlock(number uint64) (*[]*types.Eth1Transaction, uint64, error) {
	block, err := db.BigtableClient.GetBlockFromBlocksTable(number)
	if err != nil {
		return nil, 0, err
	}
	if block == nil {
		return nil, 0, fmt.Errorf("Block %d not found", number)
	}

	nextBlock := uint64(0)
	{
		blocks, err := db.BigtableClient.GetBlocksDescending(number, 2)
		if err != nil {
			return nil, 0, err
		}
		if len(blocks) > 1 {
			nextBlock = blocks[1].GetNumber()
		}
	}

	v := block.GetTransactions()
	return &v, nextBlock, nil
}
