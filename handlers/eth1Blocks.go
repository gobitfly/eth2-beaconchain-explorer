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
	"math/rand"
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
		/* // #RECY #TODO add to GetBlocksDescending?
		fullBlockData, err := db.BigtableClient.GetBlockFromBlocksTable(b.GetNumber())
		if err != nil {
			return nil, err
		}

		logrus.Infof("%v %v", b.GetTransactionCount(), len(fullBlockData.Transactions))
		tTypes := make([]int, 100)
		for _, v := range fullBlockData.Transactions {
			tTypes[v.Type] += 1
		}
		for k, v := range tTypes {
			if v != 0 {
				logrus.Infof(">> %d %d", k, v)
			}
		} /**/

		posActive := true
		slotText := "-"
		epochText := "-"
		{
			// Difficulty == 0 represent active staking, so we will show the slot
			for _, v := range b.GetDifficulty() {
				if v != 0 {
					posActive = false
					break
				}
			}

			if posActive {
				ts := uint64(b.GetTime().AsTime().Unix())
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
			}
		}

		// #RECY #RANDOM
		randomPropserName := ""
		switch os := rand.Intn(20); os {
		case 0:
			randomPropserName = "gabuwhale"
		case 1:
			randomPropserName = "Uma-70"
		case 2:
			randomPropserName = "Untitled"
		}

		gasHalf := float64(b.GetGasLimit()) / 2.0
		blockReward := utils.BlockReward(b.GetNumber())
		txReward := new(big.Int).Sub(new(big.Int).SetBytes(b.GetTxReward()), new(big.Int).Mul(new(big.Int).SetBytes(b.GetBaseFee()), big.NewInt(int64(b.GetTransactionCount()))))
		totalReward := new(big.Int).Add(blockReward, new(big.Int).Add(txReward, new(big.Int).SetBytes(b.GetUncleReward())))
		burned := new(big.Int).Mul(new(big.Int).SetBytes(b.GetBaseFee()), big.NewInt(int64(b.GetGasUsed())))

		burnedPercentage := float64(0.0)
		if len(txReward.Bits()) != 0 {
			txBurnedBig := new(big.Float).SetInt(burned)
			txBurnedBig.Quo(txBurnedBig, new(big.Float).SetInt(txReward))
			burnedPercentage, _ = txBurnedBig.Float64()
		}

		tableData[i] = []interface{}{
			epochText, // Epoch
			fmt.Sprintf(`%s<BR /><font style="font-size: .63rem; color: grey;">%v</font>`, slotText, utils.FormatTimestamp(b.GetTime().AsTime().Unix())), // Slot
			fmt.Sprintf(`<A href="block/%d">%v</A>`, b.GetNumber(), utils.FormatAddCommas(b.GetNumber())),                                                // Block
			utils.FormatBlockStatus(uint64(rand.Intn(4))),                       // Status #RECY #RANDOM
			fmt.Sprintf("%x", b.GetCoinbase()),                                  // Coinbase
			utils.FormatValidatorWithName(rand.Intn(400000), randomPropserName), // Proposer #RECY #RANDOM
			b.GetTransactionCount(),                                             // Transactions
			fmt.Sprintf(`%v<BR /><span data-toggle="tooltip" data-placement="top" title="Gas Used %%" style="font-size: .63rem; color: grey;">%.2f%%</span>&nbsp;<span data-toggle="tooltip" data-placement="top" title="%% of Gas Target" style="font-size: .63rem; color: grey;">(%+.2f%%)</span>`, utils.FormatAddCommas(b.GetGasUsed()), float64(b.GetGasUsed())/float64(b.GetGasLimit())*100.0, ((float64(b.GetGasUsed())-gasHalf)/gasHalf)*100.0), // Gas Used
			utils.FormatAddCommas(b.GetGasLimit()),                               // Gas Limit
			utils.FormatAmount(new(big.Int).SetBytes(b.GetBaseFee()), "GWei", 2), // Base Fee
			utils.FormatAmount(totalReward, "ETH", 5),                            // Reward
			fmt.Sprintf(`%v<BR /><font style="font-size: .63rem; color: grey;">%.2f%%</font>`, utils.FormatAmount(burned, "ETH", 5), burnedPercentage*100.0), // Burned Fees
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
