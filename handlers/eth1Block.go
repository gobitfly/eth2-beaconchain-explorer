package handlers

import (
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var preMergeBlockTemplate = template.Must(template.New("executionBlock").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/block.html", "templates/block/execTransactions.html"))
var eth1BlockNotFoundTemplate = template.Must(template.New("executionBlockNotFound").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/blocknotfound.html"))

// Faq will return the data from the frequently asked questions (FAQ) using a go template
func Eth1Block(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	data := InitPageData(w, r, "blockchain", "/block", "Execution Block")
	data.HeaderAd = true

	// parse block number from url
	numberString := vars["block"]
	number, err := strconv.ParseUint(numberString, 10, 64)
	if err != nil {
		err = eth1BlockNotFoundTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("a error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	// retrieve block from bigtable
	block, err := db.BigtableClient.GetBlockFromBlocksTable(number)
	if err != nil {
		err = eth1BlockNotFoundTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("b error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	// code taken from GetTokenTransactionsTableData() in bigtable.go produces race condition when retrieving address names which are already cached
	names := make(map[string]string)
	/* names[string(block.Coinbase)] = ""
	for _, t := range block.Transactions {
		names[string(t.From)] = ""
		names[string(t.To)] = ""
	}
	for _, u := range block.Uncles {
		names[string(u.Coinbase)] = ""
	}
	g := new(errgroup.Group)
	g.SetLimit(25)
	mux := sync.RWMutex{}
	for address := range names {
		address := address
		g.Go(func() error {
			name, err := db.BigtableClient.GetAddressName([]byte(address))
			if err != nil {
				return err
			}
			mux.Lock()
			names[address] = name
			mux.Unlock()
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		logger.Errorf("b error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} */

	// calculate total block reward and set lowest gas price
	txs := []types.Eth1BlockPageTransaction{}
	txFees := new(big.Int)
	lowestGasPrice := big.NewInt(1 << 32)

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

		if gasPrice := new(big.Int).SetBytes(tx.GasPrice); gasPrice.Cmp(lowestGasPrice) < 0 {
			lowestGasPrice = gasPrice
		}
		if tx.To == nil {
			tx.To = tx.Itx[0].From
		}

		method := make([]byte, 0)
		if len(tx.GetData()) > 3 {
			method = tx.GetData()[:4]
		}
		txs = append(txs, types.Eth1BlockPageTransaction{
			Hash:     fmt.Sprintf("%#x", tx.Hash),
			From:     fmt.Sprintf("%#x", tx.From),
			FromName: names[string(tx.From)],
			To:       fmt.Sprintf("%#x", tx.To),
			ToName:   names[string(tx.To)],
			Value:    new(big.Int).SetBytes(tx.Value),
			Fee:      txFee,
			GasPrice: new(big.Int).SetBytes(tx.GasPrice),
			Method:   fmt.Sprintf("%#x", method),
		})
	}

	blockReward := utils.BlockReward(block.Number)
	uncleInclusionRewards := new(big.Int)
	uncleInclusionRewards.Div(blockReward, big.NewInt(32)).Mul(uncleInclusionRewards, big.NewInt(int64(len(block.Uncles))))
	uncles := []types.Eth1BlockPageData{}
	for _, uncle := range block.Uncles {
		reward := big.NewInt(int64(uncle.Number - block.Number + 8))
		reward.Mul(reward, blockReward).Div(reward, big.NewInt(8))
		uncles = append(uncles, types.Eth1BlockPageData{
			Number:       uncle.Number,
			MinerAddress: fmt.Sprintf("%#x", uncle.Coinbase),
			MinerName:    names[string(uncle.Coinbase)],
			Reward:       reward,
			Extra:        string(uncle.Extra),
		})
	}

	burnedEth := new(big.Int).Mul(new(big.Int).SetBytes(block.BaseFee), big.NewInt(int64(block.GasUsed)))
	blockReward.Add(blockReward, txFees).Add(blockReward, uncleInclusionRewards).Sub(blockReward, burnedEth)
	eth1BlockPageData := types.Eth1BlockPageData{
		Number:         number,
		PreviousBlock:  number - 1,
		NextBlock:      number + 1,
		TxCount:        uint64(len(block.Transactions)),
		UncleCount:     uint64(len(block.Uncles)),
		Hash:           fmt.Sprintf("%#x", block.Hash),
		ParentHash:     fmt.Sprintf("%#x", block.ParentHash),
		MinerAddress:   fmt.Sprintf("%#x", block.Coinbase),
		MinerName:      names[string(block.Coinbase)],
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

	// execute template based on whether block is pre or post merge
	if eth1BlockPageData.Difficulty.Cmp(big.NewInt(0)) == 0 /* || eth1BlockPageData.Number >= 14477303 */ {
		// Post Merge PoS Block

		// calculate PoS slot number based on block timestamp
		blockSlot := (uint64(block.Time.Seconds) - utils.Config.Chain.GenesisTimestamp) / utils.Config.Chain.Config.SecondsPerSlot

		// retrieve consensus data
		// execution data is set in GetSlotPageData
		blockPageData, err := GetSlotPageData(blockSlot, false)
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
		blockPageData.ExecutionData = &eth1BlockPageData
		blockPageData.ExecFeeRecipient = block.Coinbase
		data.Data = blockPageData

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
