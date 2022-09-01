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

var eth1BlocksTemplate = template.Must(template.New("blocks").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/blocks.html"))

func Eth1Blocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "eth1blocks", "/eth1blocks", "eth1blocks")

	if utils.Config.Frontend.Debug {
		eth1BlocksTemplate = template.Must(template.New("blocks").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/blocks.html"))
	}

	err := eth1BlocksTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		return
	}
}

func Eth1BlocksData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		//logger.Errorf("error converting datatables data parameter from string to int for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	data, err := GetEth1BlocksTableData(draw, start, length)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GetEth1BlocksTableData(draw, start, length uint64) (*types.DataTableResponse, error) {
	latestBlockNumber := services.LatestEth1BlockNumber()

	blocks, err := db.BigtableClient.GetBlocksDescending(start, length)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		reward := new(big.Int).Add(utils.BlockReward(b.GetNumber()), new(big.Int).SetBytes(b.TxReward))
		tableData[i] = []interface{}{
			b.GetNumber(), // utils.FormatBlockNumber(b.GetNumber()),
			utils.FormatHash(b.GetHash(), true),
			utils.FormatDifficulty(float64(new(big.Int).SetBytes(b.GetDifficulty()).Int64())),
			utils.FormatHash(b.GetCoinbase(), true), // utils.FormatAddressAsLink(b.MinerAddress, b.MinerName, b.MinerNameVerified, b.MinerIsContract, 6),
			utils.FormatTimeFromNow(b.GetTime().AsTime()),
			b.GetTransactionCount(), // b.TxCount,
			b.GetUncleCount(),       // b.UncleCount,
			// fmt.Sprintf("%ds", b.GetDuration()),
			utils.FormatAmount(float64(reward.Int64()), "ETH", 5),
			utils.FormatAmount(float64(new(big.Int).SetBytes(b.GetMev()).Int64()), "ETH", 5),
			utils.FormatAmount(float64(new(big.Int).SetBytes(b.GetBaseFee()).Int64()), "GWei", 5),
			utils.FormatAmount(float64(new(big.Int).Mul(new(big.Int).SetBytes(b.GetBaseFee()), big.NewInt(int64(b.GetGasUsed()))).Int64()), "ETH", 5),
			// utils.FormatBlockUsage(b.GetGasUsed(), b.GasLimit),
			fmt.Sprintf("%.1f%%", 100*float64(b.GetGasUsed())/float64(b.GetGasLimit())),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    latestBlockNumber,
		RecordsFiltered: latestBlockNumber,
		Data:            tableData,
	}

	return data, nil

}
