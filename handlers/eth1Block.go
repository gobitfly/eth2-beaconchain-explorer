package handlers

import (
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var preMergeBlockTemplate = template.Must(template.New("executionBlock").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/block.html", "templates/block/execTransactions.html"))

func Eth1Block(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	data := InitPageData(w, r, "block", "/block", "Execution Block")
	data.HeaderAd = true

	// parse block number from url
	numberString := strings.Replace(vars["block"], "0x", "", -1)
	var number uint64
	var err error
	if len(numberString) == 64 {
		number, err = rpc.CurrentErigonClient.GetBlockNumberByHash(numberString)
	} else {
		number, err = strconv.ParseUint(numberString, 10, 64)
	}
	if err != nil {
		err = blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("a error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	eth1BlockPageData, err := GetExecutionBlockPageData(number)
	if err != nil {
		err = blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("a error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	// execute template based on whether block is pre or post merge
	if eth1BlockPageData.Difficulty.Cmp(big.NewInt(0)) == 0 /* || eth1BlockPageData.Number >= 14477303 */ {
		// Post Merge PoS Block

		// calculate PoS slot number based on block timestamp
		blockSlot := (uint64(eth1BlockPageData.Ts.Unix()) - utils.Config.Chain.GenesisTimestamp) / utils.Config.Chain.Config.SecondsPerSlot

		// retrieve consensus data
		// execution data is set in GetSlotPageData
		blockPageData, err := GetSlotPageData(blockSlot)
		if err == sql.ErrNoRows {
			//Slot not in database -> Show future block
			slot := uint64(blockSlot)

			if slot > MaxSlotValue {
				logger.Errorf("error retrieving blockPageData: %v", err)
				err = blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)

				if err != nil {
					logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}

			futurePageData := types.BlockPageData{
				Slot:         slot,
				Epoch:        utils.EpochOfSlot(slot),
				Ts:           utils.SlotToTime(slot),
				NextSlot:     slot + 1,
				PreviousSlot: slot - 1,
			}
			data.Data = futurePageData

			err = blockFutureTemplate.ExecuteTemplate(w, "layout", data)
			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			return
		} else if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		blockPageData.ExecutionData = eth1BlockPageData
		data.Data = blockPageData
		data.Meta.Title = fmt.Sprintf("%v - Slot %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, eth1BlockPageData.Number, time.Now().Year())
		data.Meta.Path = fmt.Sprintf("/block/%v", blockPageData.Slot)

		err = blockTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		// Pre  Merge PoW Block
		data.Data = eth1BlockPageData
		err = preMergeBlockTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("c error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func GetExecutionBlockPageData(number uint64) (*types.Eth1BlockPageData, error) {
	// retrieve block from bigtable
	block, err := db.BigtableClient.GetBlockFromBlocksTable(number)
	if err != nil {
		return nil, err
	}

	// retrieve address names from bigtable
	names := make(map[string]string)
	names[string(block.Coinbase)] = ""
	for _, tx := range block.Transactions {
		names[string(tx.From)] = ""
		names[string(tx.To)] = ""
	}
	for _, uncle := range block.Uncles {
		names[string(uncle.Coinbase)] = ""
	}
	names, _, err = db.BigtableClient.GetAddressesNamesArMetadata(&names, nil)
	if err != nil {
		return nil, err
	}

	// calculate total block reward and set lowest gas price
	txs := []types.Eth1BlockPageTransaction{}
	txFees := new(big.Int)
	lowestGasPrice := big.NewInt(1 << 62)
	for _, tx := range block.Transactions {
		// calculate tx fee depending on tx type
		txFee := new(big.Int).SetUint64(tx.GasUsed)

		if tx.Type == uint32(2) {
			// multiply gasused with min(baseFee + maxpriorityfee, maxfee)
			if normalGasPrice, maxGasPrice := new(big.Int).Add(new(big.Int).SetBytes(block.BaseFee), new(big.Int).SetBytes(tx.MaxPriorityFeePerGas)), new(big.Int).SetBytes(tx.MaxFeePerGas); normalGasPrice.Cmp(maxGasPrice) <= 0 {
				txFee.Mul(txFee, normalGasPrice)
			} else {
				txFee.Mul(txFee, maxGasPrice)
			}
		} else {
			txFee.Mul(txFee, new(big.Int).SetBytes(tx.GasPrice))
		}
		txFees.Add(txFees, txFee)
		effectiveGasPrice := new(big.Int).Div(txFee, new(big.Int).SetUint64(tx.GasUsed))
		if effectiveGasPrice.Cmp(lowestGasPrice) < 0 {
			lowestGasPrice = effectiveGasPrice
		}
		if tx.To == nil && len(tx.Itx) >= 1 {
			tx.To = tx.Itx[0].To
			names[string(tx.To)] = "Contract Creation"
		}

		method := make([]byte, 0)
		if len(tx.GetData()) > 3 {
			method = tx.GetData()[:4]
		}
		txs = append(txs, types.Eth1BlockPageTransaction{
			Hash:          fmt.Sprintf("%#x", tx.Hash),
			HashFormatted: utils.FormatAddressWithLimits(tx.Hash, "", false, "tx", 15, 18, true),
			From:          fmt.Sprintf("%#x", tx.From),
			FromFormatted: utils.FormatAddressWithLimits(tx.From, names[string(tx.From)], false, "address", 15, 20, true),
			To:            fmt.Sprintf("%#x", tx.To),
			ToFormatted:   utils.FormatAddressWithLimits(tx.To, names[string(tx.To)], names[string(tx.To)] == "Contract Creation", "address", 15, 20, true),
			Value:         new(big.Int).SetBytes(tx.Value),
			Fee:           txFee,
			GasPrice:      effectiveGasPrice,
			Method:        fmt.Sprintf("%#x", method),
		})
	}

	var blockReward *big.Int
	if new(big.Int).SetBytes(block.Difficulty).Cmp(big.NewInt(0)) == 0 {
		blockReward = big.NewInt(0)
	} else {
		blockReward = utils.BlockReward(block.Number)
	}
	uncleInclusionRewards := new(big.Int)
	uncleInclusionRewards.Div(blockReward, big.NewInt(32)).Mul(uncleInclusionRewards, big.NewInt(int64(len(block.Uncles))))
	uncles := []types.Eth1BlockPageData{}
	for _, uncle := range block.Uncles {
		reward := big.NewInt(int64(uncle.Number - block.Number + 8))
		reward.Mul(reward, blockReward).Div(reward, big.NewInt(8))
		uncles = append(uncles, types.Eth1BlockPageData{
			Number:       uncle.Number,
			MinerAddress: fmt.Sprintf("%#x", uncle.Coinbase),
			//MinerFormatted: utils.FormatAddress(uncle.Coinbase, nil, names[string(uncle.Coinbase)], false, false, false),
			MinerFormatted: utils.FormatAddressWithLimits(uncle.Coinbase, names[string(uncle.Coinbase)], false, "address", 42, 42, true),
			Reward:         reward,
			Extra:          string(uncle.Extra),
		})
	}

	burnedEth := new(big.Int).Mul(new(big.Int).SetBytes(block.BaseFee), big.NewInt(int64(block.GasUsed)))
	blockReward.Add(blockReward, txFees).Add(blockReward, uncleInclusionRewards).Sub(blockReward, burnedEth)
	eth1BlockPageData := types.Eth1BlockPageData{
		Number:        number,
		PreviousBlock: number - 1,
		NextBlock:     number + 1,
		TxCount:       uint64(len(block.Transactions)),
		UncleCount:    uint64(len(block.Uncles)),
		Hash:          fmt.Sprintf("%#x", block.Hash),
		ParentHash:    fmt.Sprintf("%#x", block.ParentHash),
		MinerAddress:  fmt.Sprintf("%#x", block.Coinbase),
		//MinerFormatted: utils.FormatAddress(block.Coinbase, nil, names[string(block.Coinbase)], false, false, false),
		MinerFormatted: utils.FormatAddressWithLimits(block.Coinbase, names[string(block.Coinbase)], false, "address", 42, 42, true),
		Reward:         blockReward,
		MevReward:      db.CalculateMevFromBlock(block),
		TxFees:         txFees,
		GasUsage:       block.GasUsed,
		GasLimit:       block.GasLimit,
		LowestGasPrice: lowestGasPrice,
		Ts:             block.GetTime().AsTime(),
		Difficulty:     new(big.Int).SetBytes(block.Difficulty),
		BaseFeePerGas:  new(big.Int).SetBytes(block.BaseFee),
		BurnedFees:     burnedEth,
		Extra:          fmt.Sprintf("%#x", block.Extra),
		Txs:            txs,
		Uncles:         uncles,
	}
	return &eth1BlockPageData, nil
}
