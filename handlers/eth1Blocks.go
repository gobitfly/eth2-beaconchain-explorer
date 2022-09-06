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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func GetEth1BlocksTableData(draw, start, length uint64) (*types.DataTableResponse, error) {
	latestBlockNumber := services.LatestEth1BlockNumber()

	if start > latestBlockNumber {
		start = 1
	} else {
		start = latestBlockNumber - start
	}

	if length > start {
		length = start
	}

	blocks, err := db.BigtableClient.GetBlocksDescending(start, length)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		ts := uint64(b.GetTime().AsTime().Unix())
		slotText := "-"
		epochText := "-"
		if ts >= utils.Config.Chain.GenesisTimestamp {
			// slot
			slot := (ts - utils.Config.Chain.GenesisTimestamp) / utils.Config.Chain.Config.SecondsPerSlot
			slotText = fmt.Sprintf(`<A href="block/%d">%s</A>`, b.GetNumber(), utils.FormatAddCommas(slot))
			// epoch
			{
				epoch := slot / utils.Config.Chain.Config.SlotsPerEpoch
				epochText = fmt.Sprintf(`<A href="epoch/%d">%s</A>`, epoch, utils.FormatAddCommas(epoch))
			}
		}

		reward := new(big.Int).Add(utils.BlockReward(b.GetNumber()), new(big.Int).SetBytes(b.GetTxReward()))
		tableData[i] = []interface{}{
			epochText, // Epoch
			fmt.Sprintf(`%s<BR /><font style="font-size: .63rem; color: grey;">%v</font>`, slotText, utils.FormatTimestamp(b.GetTime().AsTime().Unix())), // Slot
			fmt.Sprintf(`<A href="block/%d">%v</A>`, b.GetNumber(), utils.FormatAddCommas(b.GetNumber())),                                                // Block // utils.FormatBlockNumber(b.GetNumber()),
			"-",                     // Status
			"-",                     // Proposer
			b.GetTransactionCount(), // Transactions
			fmt.Sprintf(`%v<BR /><font style="font-size: .63rem; color: grey;">%.2f%% + -%%</font>`, utils.FormatAddCommas(b.GetGasUsed()), float64(b.GetGasUsed())/float64(b.GetGasLimit())*100.0), // Gas Used
			utils.FormatAddCommas(b.GetGasLimit()),                               // Gas Limit
			utils.FormatAmount(new(big.Int).SetBytes(b.GetBaseFee()), "GWei", 2), // Base Fee
			utils.FormatAmount(reward, "ETH", 5),                                 // Reward
			fmt.Sprintf(`%v (%.2f%%)`, utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetBytes(b.GetBaseFee()), big.NewInt(int64(b.GetGasUsed()))), "ETH", 5), 0.0), // Burned Fees
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
