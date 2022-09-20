package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var preMergeBlockTemplate = template.Must(template.New("executionBlock").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/block.html", "templates/block/execTransactions.html"))

func Eth1Block(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

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
		data := InitPageData(w, r, "blockchain", "/block", fmt.Sprintf("Block %d", 0))
		err = blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("a error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	data := InitPageData(w, r, "blockchain", "/block", fmt.Sprintf("Block %d", number))
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
		if err != nil {
			logger.Errorf("error retrieving slot page data: %v", err)

			err = blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)
			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			return
		}
		blockPageData.ExecutionData = eth1BlockPageData

		data.HeaderAd = true
		data.Data = blockPageData

		err = blockTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		// Pre  Merge PoW Block
		data := InitPageData(w, r, "block", "/block", fmt.Sprintf("Block %d", eth1BlockPageData.Number))
		data.HeaderAd = true
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
	block, err := db.BigtableClient.GetBlockFromBlocksTable(number)
	if diffToHead := int64(services.LatestEth1BlockNumber()) - int64(number); err != nil && diffToHead < 0 && diffToHead >= -5 {
		block, _, err = rpc.CurrentErigonClient.GetBlock(int64(number))
	}
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
		// sum txFees
		txFee := db.CalculateTxFeeFromTransaction(tx, new(big.Int).SetBytes(block.BaseFee))
		txFees.Add(txFees, txFee)

		effectiveGasPrice := big.NewInt(0)
		if gasUsed := new(big.Int).SetUint64(tx.GasUsed); gasUsed.Cmp(big.NewInt(0)) != 0 {
			// calculate effective gas price
			effectiveGasPrice = new(big.Int).Div(txFee, gasUsed)
			if effectiveGasPrice.Cmp(lowestGasPrice) < 0 {
				lowestGasPrice = effectiveGasPrice
			}
		}

		// set tx to if tx is contract creation
		if tx.To == nil && len(tx.Itx) >= 1 {
			tx.To = tx.Itx[0].To
			names[string(tx.To)] = "Contract Creation"
		}

		method := make([]byte, 0)
		if len(tx.GetData()) > 3 && (len(tx.GetItx()) > 0 || tx.GetGasUsed() > 21000 || tx.GetErrorMsg() != "") {
			method = tx.GetData()[:4]
		}
		txs = append(txs, types.Eth1BlockPageTransaction{
			Hash:          fmt.Sprintf("%#x", tx.Hash),
			HashFormatted: utils.FormatAddressWithLimits(tx.Hash, "", false, "tx", 15, 18, true),
			From:          fmt.Sprintf("%#x", tx.From),
			FromFormatted: utils.FormatAddressWithLimits(tx.From, names[string(tx.From)], false, "address", 15, 20, true),
			To:            fmt.Sprintf("%#x", tx.To),
			ToFormatted:   utils.FormatAddressWithLimits(tx.To, names[string(tx.To)], names[string(tx.To)] == "Contract Creation" || len(method) > 0, "address", 15, 20, true),
			Value:         new(big.Int).SetBytes(tx.Value),
			Fee:           txFee,
			GasPrice:      effectiveGasPrice,
			Method:        fmt.Sprintf("%#x", method),
		})
	}

	blockReward := utils.Eth1BlockReward(block.Number, block.Difficulty)

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
	nextBlock := number + 1
	if nextBlock > services.LatestEth1BlockNumber() {
		nextBlock = 0
	}
	eth1BlockPageData := types.Eth1BlockPageData{
		Number:        number,
		PreviousBlock: number - 1,
		NextBlock:     nextBlock,
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
