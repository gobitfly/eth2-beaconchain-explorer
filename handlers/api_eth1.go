package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/mr-tron/base58/base58"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/maps"
)

// ApiEth1Deposit godoc
// @Summary Get an eth1 deposit by its eth1 transaction hash
// @Tags Execution
// @Produce  json
// @Param  txhash path string true "Eth1 transaction hash"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/eth1deposit/{txhash} [get]
func ApiEth1Deposit(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	eth1TxHash, err := hex.DecodeString(strings.Replace(vars["txhash"], "0x", "", -1))
	if err != nil {
		sendErrorResponse(w, r.URL.String(), "invalid eth1 tx hash provided")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT amount, block_number, block_ts, from_address, merkletree_index, publickey, removed, signature, tx_hash, tx_index, tx_input, valid_signature, withdrawal_credentials FROM eth1_deposits WHERE tx_hash = $1", eth1TxHash)
	if err != nil {
		sendErrorResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, w, r)
}

// ApiETH1ExecBlocks godoc
// @Summary Get execution blocks
// @Tags Execution
// @Description Get execution blocks by execution block number
// @Produce json
// @Param blockNumber path string true "Provide one or more execution block numbers. Coma separated up to max 100. "
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/execution/block/{blockNumber} [get]
func ApiETH1ExecBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	var blockList []uint64
	splits := strings.Split(vars["blockNumber"], ",")
	for _, split := range splits {
		temp, err := strconv.ParseUint(split, 10, 64)
		if err != nil {
			sendErrorResponse(w, r.URL.String(), "invalid block number")
			return
		}
		blockList = append(blockList, temp)
	}

	blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, uint64(100))
	if err != nil {
		logger.Errorf("Can not retrieve blocks from bigtable %v", err)
		sendErrorResponse(w, r.URL.String(), "can not retrieve blocks from bigtable")
		return
	}

	_, beaconDataMap, err := findExecBlockNumbersByExecBlockNumber(blockList, 0, 100)
	if err != nil {
		sendErrorResponse(w, r.URL.String(), "can not retrieve proposer information")
		return
	}

	relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
	if err != nil {
		logger.Errorf("can not load mev data %v", err)
		sendErrorResponse(w, r.URL.String(), "can not retrieve mev data")
		return
	}

	results := formatBlocksForApiResponse(blocks, relaysData, beaconDataMap, nil)

	j := json.NewEncoder(w)
	sendOKResponse(j, r.URL.String(), []interface{}{results})
}

// ApiETH1AccountProposedBlocks godoc
// @Summary Get proposed or mined blocks
// @Tags Execution
// @Description Get a list of proposed or mined blocks from a given fee recipient address, proposer index or proposer pubkey.
// @Description Mixed use of recipient addresses and proposer indexes or proposer pubkeys with an offset is discouraged as it can lead to skipped entries.
// @Produce json
// @Param addressIndexOrPubkey path string true "Either the fee recipient address, the proposer index or proposer pubkey. You can provide multiple by separating them with ','. Max allowed index or pubkeys are 100, max allowed user addresses are 20."
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit, amount of entries you wish to receive" default(10)
// @Param sort query string false "Sort via the block number either by 'asc' or 'desc'" default(desc)
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/execution/{addressIndexOrPubkey}/produced [get]
func ApiETH1AccountProducedBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	maxValidators := getUserPremium(r).MaxValidators
	addresses, indices, err := getAddressesOrIndicesFromAddressIndexOrPubkey(vars["addressIndexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(
			w,
			r.URL.String(),
			fmt.Sprintf("invalid address, validator index or pubkey or exceeded max of %v params", maxValidators),
		)
		return
	}

	if len(addresses) > 20 {
		sendErrorResponse(
			w,
			r.URL.String(),
			"you are only allowed to query up to max 20 addresses",
		)
		return
	}

	var offset uint64 = 0
	var limit uint64 = 10
	var isSortAsc bool = false

	offsetString := r.URL.Query().Get("offset")
	offset, err = strconv.ParseUint(offsetString, 10, 64)
	if err != nil {
		offset = 0
	}

	limitString := r.URL.Query().Get("limit")
	limit, err = strconv.ParseUint(limitString, 10, 64)
	if err != nil {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	sortString := r.URL.Query().Get("sort")
	if sortString == "asc" {
		isSortAsc = true
	}

	var blockList []uint64
	var beaconDataMap = map[uint64]types.ExecBlockProposer{}
	if len(addresses) > 0 {
		blockListSub, beaconDataMapSub, err := findExecBlockNumbersByFeeRecipient(addresses, offset, limit, isSortAsc)
		if err != nil {
			sendErrorResponse(w, r.URL.String(), "can not retrieve blocks from database")
			return
		}
		blockList = append(blockList, blockListSub...)
		for key, val := range beaconDataMapSub {
			beaconDataMap[key] = val
		}
	}

	if len(indices) > 0 {
		blockListSub, beaconDataMapSub, err := findExecBlockNumbersByProposerIndex(indices, offset, limit, isSortAsc, false)
		if err != nil {
			sendErrorResponse(w, r.URL.String(), "can not retrieve blocks from database")
			return
		}
		blockList = append(blockList, blockListSub...)
		for key, val := range beaconDataMapSub {
			beaconDataMap[key] = val
		}
	}

	// Remove duplicates from the block list
	allKeys := make(map[uint64]bool)
	list := []uint64{}
	for _, item := range blockList {
		if _, ok := allKeys[item]; !ok {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	blockList = list

	// Trim to the blocks that are within the limit range
	if isSortAsc {
		sort.Slice(blockList, func(i, j int) bool { return blockList[i] < blockList[j] })
	} else {
		sort.Slice(blockList, func(i, j int) bool { return blockList[i] > blockList[j] })
	}
	if len(blockList) > int(limit) {
		blockList = blockList[:limit]
	}

	blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, uint64(limit))
	if err != nil {
		logger.Errorf("Can not retrieve blocks from bigtable %v", err)
		sendErrorResponse(w, r.URL.String(), "can not retrieve blocks from bigtable")
		return
	}

	relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
	if err != nil {
		logger.Errorf("can not load mev data %v", err)
		sendErrorResponse(w, r.URL.String(), "can not retrieve mev data")
		return
	}

	var sortFunc func(i, j types.ExecutionBlockApiResponse) bool
	if isSortAsc {
		sortFunc = func(i, j types.ExecutionBlockApiResponse) bool { return i.BlockNumber < j.BlockNumber }
	}

	results := formatBlocksForApiResponse(blocks, relaysData, beaconDataMap, sortFunc)

	j := json.NewEncoder(w)
	sendOKResponse(j, r.URL.String(), []interface{}{results})
}

// ApiETH1GasNowData godoc
// @Summary Gets the current estimation for gas prices in GWei.
// @Tags Execution
// @Description The response is split into four estimated inclusion speeds rapid (15 seconds), fast (1 minute), standard (3 minutes) and slow (> 10 minutes).
// @Produce json
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/execution/gasnow [get]
func ApiEth1GasNowData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	gasnowData := services.LatestGasNowData()

	if gasnowData == nil {
		logger.Errorf("error gasnow data is not defined. The frontend updater might not be running.")
		sendErrorResponse(w, r.URL.String(), "error gasnow data is currently not available.")
		return
	}

	gasnowData.Data.PriceUSD = price.GetEthPrice("USD")
	gasnowData.Data.Currency = ""

	err := json.NewEncoder(w).Encode(gasnowData)
	if err != nil {
		logger.Errorf("error gasnow data is not defined. The frontend updater might not be running.")
		sendErrorResponse(w, r.URL.String(), "error gasnow data is currently not available.")
		return
	}
}

// ApiEth1Address godoc
// @Summary Gets information about an ethereum address.
// @Tags Execution
// @Description Returns the ether balance and any token balances for a given ethereum address.
// @Produce json
// @Param address path string true "provide an ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters"
// @Param token query string false "filter for a specific token by providing a ethereum token contract address"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/execution/address/{address} [get]
func ApiEth1Address(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	vars := mux.Vars(r)
	address := vars["address"]
	q := r.URL.Query()

	address = strings.Replace(address, "0x", "", -1)
	address = strings.ToLower(address)

	if !utils.IsEth1Address(address) {
		sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
		return
	}
	token := q.Get("token")

	if len(token) > 0 {
		token = strings.Replace(token, "0x", "", -1)
		token = strings.ToLower(token)
		if !utils.IsEth1Address(token) {
			sendErrorResponse(w, r.URL.String(), "error invalid token query param. A token address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
			return
		}
	}

	response := types.ApiEth1AddressResponse{}

	metadata, err := db.BigtableClient.GetMetadataForAddress(common.FromHex(address))
	if err != nil {
		logger.Errorf("error retrieving metadata for address: %v route: %v err: %v", address, r.URL.String(), err)
		sendErrorResponse(w, r.URL.String(), "error could not get metadata for address")
		return
	}

	response.Ether = decimal.NewFromBigInt(new(big.Int).SetBytes(metadata.EthBalance.Balance), 0).DivRound(decimal.NewFromInt(1e18), 18).String()
	response.Address = fmt.Sprintf("0x%x", metadata.EthBalance.Address)
	response.Tokens = []struct {
		Address  string  `json:"address"`
		Balance  string  `json:"balance"`
		Symbol   string  `json:"symbol"`
		Decimals string  `json:"decimals,omitempty"`
		Price    float64 `json:"price,omitempty"`
		Currency string  `json:"currency,omitempty"`
	}{}
	for _, m := range metadata.Balances {
		// var price float64
		// if len(m.Metadata.Price) > 0 {
		// 	price, err = strconv.ParseFloat(string(m.Metadata.Price), 64)
		// 	if err != nil {
		// 		logger.Errorf("error parsing price to float for address: %v route: %v err: %v", address, r.URL.String(), err)
		// 		sendErrorResponse(w, r.URL.String(), "error could not get metadata for address")
		// 		return
		// 	}
		// }

		// if there is a token filter and we are currently not on the right value, skip to the next loop iteration
		if len(token) > 0 && token != fmt.Sprintf("%x", m.Token) {
			continue
		}

		response.Tokens = append(response.Tokens, struct {
			Address  string  `json:"address"`
			Balance  string  `json:"balance"`
			Symbol   string  `json:"symbol"`
			Decimals string  `json:"decimals,omitempty"`
			Price    float64 `json:"price,omitempty"`
			Currency string  `json:"currency,omitempty"`
		}{
			Address: fmt.Sprintf("0x%x", m.Token),
			Balance: decimal.NewFromBigInt(new(big.Int).SetBytes(m.Balance), 0).Div(decimal.NewFromBigInt(big.NewInt(1), int32(new(big.Int).SetBytes(m.Metadata.Decimals).Int64()))).String(),
			Symbol:  m.Metadata.Symbol,
			// Decimals: decimals.String(),
			// Price:   price,
			// Currency: "USD",
		})
	}

	sendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{response})
}

func ApiEth1AddressTx(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	vars := mux.Vars(r)
	address := vars["address"]
	q := r.URL.Query()

	address = strings.Replace(address, "0x", "", -1)
	address = strings.ToLower(address)

	if !utils.IsEth1Address(address) {
		sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
		return
	}

	response := types.APIEth1AddressTxResponse{}

	filter := q.Get("filter")
	filters := map[string]string{
		"":         string(db.FILTER_TIME),
		"time":     string(db.FILTER_TIME),
		"received": string(db.FILTER_FROM),
		"sent":     string(db.FILTER_TO),
		// "method":   string(db.FILTER_METHOD),
		// "contract": string(db.FILTER_CONTRACT),
	}

	filter, ok := filters[filter]
	if !ok {
		sendErrorResponse(w, r.URL.String(), "error invalid filter provided. Please provide a valid filter: time (default), received, sent")
		return
	}

	pageToken := q.Get("page")
	if len(pageToken) > 0 {
		token, err := base58.FastBase58Decoding(pageToken)
		if err != nil {
			logger.Errorf("error invalid page token provided: %v err: %v", q.Get("page"), err)
			sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
			return
		}
		pageToken = fmt.Sprintf("%d:I:TX:%s:%s:%s", utils.Config.Chain.Config.DepositChainID, address, filter, token)
	}

	if len(pageToken) == 0 {
		pageToken = fmt.Sprintf("%d:I:TX:%s:%s:", utils.Config.Chain.Config.DepositChainID, address, filter)
	}

	transactions, lastKey, err := db.BigtableClient.GetEth1TxForAddress(pageToken, 25)
	if err != nil {
		logger.Errorf("error getting transactions for address: %v route: %v err: %v", address, r.URL.String(), err)
		sendErrorResponse(w, r.URL.String(), "error getting transactions for address")
		return
	}
	response.Page = base58.FastBase58Encoding([]byte(strings.TrimPrefix(lastKey, fmt.Sprintf("%d:I:TX:%s:%s:", utils.Config.Chain.Config.DepositChainID, address, filter))))

	txsParsed := make([]types.Eth1TransactionParsed, 0, len(transactions))

	for _, tx := range transactions {
		txsParsed = append(txsParsed, types.Eth1TransactionParsed{
			Hash:               fmt.Sprintf("0x%x", tx.Hash),
			BlockNumber:        tx.BlockNumber,
			Time:               tx.Time.AsTime(),
			From:               utils.FixAddressCasing(fmt.Sprintf("%x", tx.From)),
			To:                 utils.FixAddressCasing(fmt.Sprintf("%x", tx.To)),
			MethodId:           fmt.Sprintf("0x%x", tx.MethodId),
			Value:              new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(tx.Value)), big.NewFloat(1e18)).String(),   //new(big.Int).Div(new(big.Int).SetBytes(tx.Value), big.NewInt(1e18)).String(),
			GasPrice:           new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(tx.GasPrice)), big.NewFloat(1e9)).String(), //new(big.Int).Div(new(big.Int).SetBytes(tx.GasPrice), new(big.Int).SetInt64(1e18)).String(),
			IsContractCreation: tx.IsContractCreation,
			InvokesContract:    tx.InvokesContract,
		})
	}

	response.Transactions = txsParsed

	sendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{response})
}

func ApiEth1AddressItx(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	vars := mux.Vars(r)
	address := vars["address"]
	q := r.URL.Query()

	address = strings.Replace(address, "0x", "", -1)
	address = strings.ToLower(address)

	if !utils.IsEth1Address(address) {
		sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
		return
	}

	response := types.APIEth1AddressItxResponse{}

	filter := q.Get("filter")
	filters := map[string]string{
		"":         string(db.FILTER_TIME),
		"time":     string(db.FILTER_TIME),
		"received": string(db.FILTER_FROM),
		"sent":     string(db.FILTER_TO),
		// "method":   string(db.FILTER_METHOD),
		// "contract": string(db.FILTER_CONTRACT),
	}

	filter, ok := filters[filter]
	if !ok {
		sendErrorResponse(w, r.URL.String(), "error invalid filter provided. Please provide a valid filter: time (default), received, sent")
		return
	}

	prefixFormat := "%d:I:ITX:%s:%s:"

	pageToken := q.Get("page")
	if len(pageToken) > 0 {
		token, err := base58.FastBase58Decoding(pageToken)
		if err != nil {
			logger.Errorf("error invalid page token provided: %v err: %v", q.Get("page"), err)
			sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
			return
		}
		pageToken = fmt.Sprintf(prefixFormat+"%s", utils.Config.Chain.Config.DepositChainID, address, filter, token)
	}

	if len(pageToken) == 0 {
		pageToken = fmt.Sprintf(prefixFormat, utils.Config.Chain.Config.DepositChainID, address, filter)
	}

	internalTransactions, lastKey, err := db.BigtableClient.GetEth1ItxForAddress(pageToken, 25)
	if err != nil {
		logger.Errorf("error getting transactions for address: %v route: %v err: %v", address, r.URL.String(), err)
		sendErrorResponse(w, r.URL.String(), "error getting transactions for address")
		return
	}
	response.Page = base58.FastBase58Encoding([]byte(strings.TrimPrefix(lastKey, fmt.Sprintf(prefixFormat, utils.Config.Chain.Config.DepositChainID, address, filter))))

	itxParsed := make([]types.Eth1InternalTransactionParsed, 0, len(internalTransactions))

	for _, itx := range internalTransactions {
		itxParsed = append(itxParsed, types.Eth1InternalTransactionParsed{
			ParentHash:  fmt.Sprintf("0x%x", itx.ParentHash),
			BlockNumber: itx.BlockNumber,
			Time:        itx.Time.AsTime(),
			Type:        itx.Type,
			From:        utils.FixAddressCasing(fmt.Sprintf("%x", itx.From)),
			To:          utils.FixAddressCasing(fmt.Sprintf("%x", itx.To)),
			Value:       new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(itx.Value)), big.NewFloat(1e18)).String(),
		})
	}

	response.InternalTransactions = itxParsed
	sendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{response})
}

func ApiEth1AddressBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	vars := mux.Vars(r)
	address := vars["address"]
	q := r.URL.Query()

	address = strings.Replace(address, "0x", "", -1)
	address = strings.ToLower(address)

	if !utils.IsEth1Address(address) {
		sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
		return
	}

	response := types.APIEth1AddressBlockResponse{}

	prefixFormat := "%d:I:B:%s:"

	pageToken := q.Get("page")
	if len(pageToken) > 0 {
		token, err := base58.FastBase58Decoding(pageToken)
		if err != nil {
			logger.Errorf("error invalid page token provided: %v err: %v", q.Get("page"), err)
			sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
			return
		}
		pageToken = fmt.Sprintf(prefixFormat+"%s", utils.Config.Chain.Config.DepositChainID, address, token)
	}

	if len(pageToken) == 0 {
		pageToken = fmt.Sprintf(prefixFormat, utils.Config.Chain.Config.DepositChainID, address)
	}

	producedBlocks, lastKey, err := db.BigtableClient.GetEth1BlocksForAddress(pageToken, 25)
	if err != nil {
		logger.Errorf("error getting transactions for address: %v route: %v err: %v", address, r.URL.String(), err)
		sendErrorResponse(w, r.URL.String(), "error getting transactions for address")
		return
	}
	response.Page = base58.FastBase58Encoding([]byte(strings.TrimPrefix(lastKey, fmt.Sprintf(prefixFormat, utils.Config.Chain.Config.DepositChainID, address))))

	blocksParsed := make([]types.Eth1BlockParsed, 0, len(producedBlocks))

	for _, blk := range producedBlocks {
		txReward := new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(blk.TxReward)), big.NewFloat(1e18)).String()
		if txReward == "0" {
			txReward = ""
		}

		uncleHash := fmt.Sprintf("0x%x", blk.UncleHash)
		uncleReward := new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(blk.UncleReward)), big.NewFloat(1e18)).String()
		if uncleReward == "0" {
			uncleReward = ""
			uncleHash = ""
		}
		// mev := new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(blk.Mev)), big.NewFloat(1e18)).String()
		// if mev == "0" {
		// 	mev = ""
		// }
		difficulty := new(big.Int).SetBytes(blk.Difficulty).String()
		if difficulty == "0" {
			difficulty = ""
		}

		// blkReward := utils.Eth1BlockReward(blk.Number, blk.Difficulty)

		blocksParsed = append(blocksParsed, types.Eth1BlockParsed{
			Hash:                     fmt.Sprintf("0x%x", blk.Hash),
			ParentHash:               fmt.Sprintf("0x%x", blk.ParentHash),
			UncleHash:                uncleHash,
			Coinbase:                 fmt.Sprintf("0x%x", blk.Coinbase), //new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(blk.Coinbase)), big.NewFloat(1e18)).String(),
			Difficulty:               difficulty,
			Number:                   blk.Number,
			GasLimit:                 blk.GasLimit,
			GasUsed:                  blk.GasUsed,
			Time:                     blk.Time.AsTime(),
			BaseFee:                  new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(blk.BaseFee)), big.NewFloat(1e9)).String(),
			UncleCount:               blk.UncleCount,
			TransactionCount:         blk.TransactionCount,
			InternalTransactionCount: blk.InternalTransactionCount,
			TxReward:                 txReward,
			UncleReward:              uncleReward,
			// Mev:                      mev,
			// LowestGasPrice:           new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(blk.LowestGasPrice)), big.NewFloat(1e9)).String(),
			// HighestGasPrice:          new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(blk.HighestGasPrice)), big.NewFloat(1e9)).String(),
		})
	}

	response.ProducedBlocks = blocksParsed
	sendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{response})
}

func ApiEth1AddressUncles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	vars := mux.Vars(r)
	address := vars["address"]
	q := r.URL.Query()

	address = strings.Replace(address, "0x", "", -1)
	address = strings.ToLower(address)

	if !utils.IsEth1Address(address) {
		sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
		return
	}

	response := types.APIEth1AddressUncleResponse{}

	prefixFormat := "%d:I:B:%s:"

	pageToken := q.Get("page")
	if len(pageToken) > 0 {
		token, err := base58.FastBase58Decoding(pageToken)
		if err != nil {
			logger.Errorf("error invalid page token provided: %v err: %v", q.Get("page"), err)
			sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
			return
		}
		pageToken = fmt.Sprintf(prefixFormat+"%s", utils.Config.Chain.Config.DepositChainID, address, token)
	}

	if len(pageToken) == 0 {
		pageToken = fmt.Sprintf(prefixFormat, utils.Config.Chain.Config.DepositChainID, address)
	}

	producedUncle, lastKey, err := db.BigtableClient.GetEth1UnclesForAddress(pageToken, 25)
	if err != nil {
		logger.Errorf("error getting transactions for address: %v route: %v err: %v", address, r.URL.String(), err)
		sendErrorResponse(w, r.URL.String(), "error getting transactions for address")
		return
	}
	response.Page = base58.FastBase58Encoding([]byte(strings.TrimPrefix(lastKey, fmt.Sprintf(prefixFormat, utils.Config.Chain.Config.DepositChainID, address))))

	unclesParsed := make([]types.Eth1UncleParsed, 0, len(producedUncle))

	for _, uncl := range producedUncle {

		unclesParsed = append(unclesParsed, types.Eth1UncleParsed{
			BlockNumber: uncl.BlockNumber,
			Number:      uncl.Number,
			GasLimit:    uncl.GasLimit,
			GasUsed:     uncl.GasUsed,
			BaseFee:     new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(uncl.BaseFee)), big.NewFloat(1e9)).String(),
			Difficulty:  new(big.Int).SetBytes(uncl.Difficulty).String(),
			Time:        uncl.Time.AsTime(),
			Reward:      new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(uncl.Reward)), big.NewFloat(1e18)).String(),
		})
	}

	response.ProducedUncles = unclesParsed
	sendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{response})
}

func ApiEth1AddressTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	vars := mux.Vars(r)
	address := vars["address"]
	q := r.URL.Query()

	address = strings.Replace(address, "0x", "", -1)
	address = strings.ToLower(address)

	if !utils.IsEth1Address(address) {
		sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
		return
	}

	token := q.Get("token")
	token = strings.ToLower(token)
	tokens := map[string]string{
		"":        "ERC20",
		"token":   "ERC20",
		"erc20":   "ERC20",
		"erc721":  "ERC721",
		"nft":     "ERC721",
		"erc1155": "ERC1155",
	}

	selectedToken, ok := tokens[token]
	if !ok {
		validTokens := make([]string, 0, len(tokens))
		for k, v := range tokens {
			if k == "" {
				k = fmt.Sprintf("(default %v)", v)
			}
			validTokens = append(validTokens, k)
		}
		sort.Slice(validTokens, func(i, j int) bool {
			return strings.Contains("default", validTokens[i]) || validTokens[i] < validTokens[j]
		})
		sendErrorResponse(w, r.URL.String(), fmt.Sprintf("error invalid token provided. Please provide a valid tokens: %s", strings.Join(validTokens, ", ")))
		return
	}

	response := types.APIEth1TokenResponse{}

	prefixFormat := fmt.Sprintf("%%d:I:%s:%%s:%%s:", selectedToken)

	pageToken := q.Get("page")
	if len(pageToken) > 0 {
		token, err := base58.FastBase58Decoding(pageToken)
		if err != nil {
			logger.Errorf("error invalid page token provided: %v err: %v", q.Get("page"), err)
			sendErrorResponse(w, r.URL.String(), "error invalid address. A ethereum address consists of an optional 0x prefix followed by 40 hexadecimal characters.")
			return
		}
		pageToken = fmt.Sprintf(prefixFormat+"%s", utils.Config.Chain.Config.DepositChainID, address, token)
	}

	if len(pageToken) == 0 {
		pageToken = fmt.Sprintf(prefixFormat, utils.Config.Chain.Config.DepositChainID, address)
	}
	pageSize := 25
	transactions := make([]*types.Eth1TokenTxParsed, 0, pageSize)
	pageKey := ""
	switch selectedToken {
	case "erc721":
		txs, lastKey, err := db.BigtableClient.GetEth1ERC721ForAddress(pageToken, 25)
		if err != nil {
			logger.Errorf("error getting token: %v transactions for address: %v route: %v err: %v", selectedToken, address, r.URL.String(), err)
			sendErrorResponse(w, r.URL.String(), "error getting transactions for address")
			return
		}
		pageKey = lastKey

		for _, tx := range txs {
			transactions = append(transactions, &types.Eth1TokenTxParsed{
				ParentHash:   fmt.Sprintf("0x%x", tx.ParentHash),
				BlockNumber:  tx.BlockNumber,
				TokenAddress: fmt.Sprintf("0x%x", tx.TokenAddress),
				Time:         tx.Time.AsTime(),
				From:         utils.FixAddressCasing(fmt.Sprintf("%x", tx.From)),
				To:           utils.FixAddressCasing(fmt.Sprintf("%x", tx.To)),
				TokenId:      new(big.Int).SetBytes(tx.TokenId).String(),
			})
		}

	case "erc1155":
		txs, lastKey, err := db.BigtableClient.GetEth1ERC1155ForAddress(pageToken, 25)
		if err != nil {
			logger.Errorf("error getting token: %v transactions for address: %v route: %v err: %v", selectedToken, address, r.URL.String(), err)
			sendErrorResponse(w, r.URL.String(), "error getting transactions for address")
			return
		}
		pageKey = lastKey

		for _, tx := range txs {
			transactions = append(transactions, &types.Eth1TokenTxParsed{
				ParentHash:   fmt.Sprintf("0x%x", tx.ParentHash),
				BlockNumber:  tx.BlockNumber,
				TokenAddress: fmt.Sprintf("0x%x", tx.TokenAddress),
				Time:         tx.Time.AsTime(),
				From:         utils.FixAddressCasing(fmt.Sprintf("%x", tx.From)),
				To:           utils.FixAddressCasing(fmt.Sprintf("%x", tx.To)),
				TokenId:      new(big.Int).SetBytes(tx.TokenId).String(),
				Value:        new(big.Int).SetBytes(tx.Value).String(),
				Operator:     new(big.Int).SetBytes(tx.Operator).String(),
			})
		}

	default:
		txs, lastKey, err := db.BigtableClient.GetEth1ERC20ForAddress(pageToken, 25)
		if err != nil {
			logger.Errorf("error getting token: %v transactions for address: %v route: %v err: %v", selectedToken, address, r.URL.String(), err)
			sendErrorResponse(w, r.URL.String(), "error getting transactions for token")
			return
		}
		pageKey = lastKey
		tokenMeta := make(map[string]*types.ERC20Metadata, 25)
		for _, tx := range txs {
			_, ok := tokenMeta[string(tx.TokenAddress)]
			if !ok {
				metadata, err := db.BigtableClient.GetERC20MetadataForAddress([]byte(address))
				if err != nil {
					logger.Errorf("error getting token: %v metadata for address: %v route: %v err: %v", selectedToken, address, r.URL.String(), err)
					sendErrorResponse(w, r.URL.String(), "error getting transactions for token")
					return
				}
				tokenMeta[string(tx.TokenAddress)] = metadata
			}
			value := new(big.Int).SetBytes(tx.Value).String()
			m, ok := tokenMeta[string(tx.TokenAddress)]
			if ok {
				value = utils.FormatErc20Decimals(tx.Value, m).String()
			}

			transactions = append(transactions, &types.Eth1TokenTxParsed{
				ParentHash:   fmt.Sprintf("0x%x", tx.ParentHash),
				BlockNumber:  tx.BlockNumber,
				TokenAddress: fmt.Sprintf("0x%x", tx.TokenAddress),
				Time:         tx.Time.AsTime(),
				From:         utils.FixAddressCasing(fmt.Sprintf("%x", tx.From)),
				To:           utils.FixAddressCasing(fmt.Sprintf("%x", tx.To)),
				Value:        value,
			})
		}
	}

	response.Page = base58.FastBase58Encoding([]byte(strings.TrimPrefix(pageKey, fmt.Sprintf(prefixFormat, utils.Config.Chain.Config.DepositChainID, address))))

	response.TokenTxs = transactions
	sendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{response})
}

func formatBlocksForApiResponse(blocks []*types.Eth1BlockIndexed, relaysData map[common.Hash]types.RelaysData, beaconDataMap map[uint64]types.ExecBlockProposer, sortFunc func(i, j types.ExecutionBlockApiResponse) bool) []types.ExecutionBlockApiResponse {
	results := []types.ExecutionBlockApiResponse{}

	latestFinalized := services.LatestFinalizedEpoch()

	for _, block := range blocks {

		totalReward := utils.Eth1TotalReward(block)

		baseFee := big.NewInt(0).SetBytes(block.GetBaseFee())
		difficulty := big.NewInt(0).SetBytes(block.GetDifficulty())

		posData, ok := beaconDataMap[block.GetNumber()]
		var posDataPt *types.ExecBlockProposer = nil
		if ok {
			posData.Finalized = latestFinalized >= posData.Epoch
			posDataPt = &posData
		}

		consensusAlgorithm := "pos"
		if len(block.GetDifficulty()) != 0 {
			consensusAlgorithm = "pow"
		}

		var mevBribe *big.Int = big.NewInt(0)
		relayData, ok := relaysData[common.BytesToHash(block.Hash)]
		var relayDataResponse *types.RelayDataApiResponse = nil
		if ok {
			mevBribe = relayData.MevBribe.BigInt()
			relayDataResponse = &types.RelayDataApiResponse{
				TagID:                relayData.TagID,
				BuilderPubKey:        fmt.Sprintf("0x%v", hex.EncodeToString(relayData.BuilderPubKey)),
				ProposerFeeRecipient: fmt.Sprintf("0x%v", hex.EncodeToString(relayData.MevRecipient)),
			}
		}

		var producerReward *big.Int
		if mevBribe.Int64() == 0 {
			producerReward = totalReward
		} else {
			producerReward = mevBribe
		}

		results = append(results, types.ExecutionBlockApiResponse{
			Hash:               fmt.Sprintf("0x%v", hex.EncodeToString(block.GetHash())),
			BlockNumber:        block.GetNumber(),
			Timestamp:          uint64(block.GetTime().AsTime().Unix()),
			BlockReward:        totalReward,
			BlockMevReward:     mevBribe,
			FeeRecipientReward: producerReward,
			FeeRecipient:       fmt.Sprintf("0x%v", hex.EncodeToString(block.GetCoinbase())),
			GasLimit:           block.GetGasLimit(),
			GasUsed:            block.GetGasUsed(),
			UncleCount:         block.GetUncleCount(),
			BaseFee:            baseFee,
			TxCount:            block.GetTransactionCount(),
			InternalTxCount:    block.GetInternalTransactionCount(),
			ParentHash:         fmt.Sprintf("0x%v", hex.EncodeToString(block.GetParentHash())),
			UncleHash:          fmt.Sprintf("0x%v", hex.EncodeToString(block.GetUncleHash())),
			Difficulty:         difficulty,
			PoSData:            posDataPt,
			RelayData:          relayDataResponse,
			ConsensusAlgorithm: consensusAlgorithm,
		})
	}

	if sortFunc != nil {
		sort.SliceStable(results, func(i, j int) bool { return sortFunc(results[i], results[j]) })
	}

	return results
}

func getValidatorExecutionPerformance(queryIndices []uint64) ([]types.ExecutionPerformanceResponse, error) {
	latestEpoch := services.LatestEpoch()
	last31dTimestamp := time.Now().Add(-31 * 24 * time.Hour)
	last7dTimestamp := time.Now().Add(-7 * 24 * time.Hour)
	last1dTimestamp := time.Now().Add(-1 * 24 * time.Hour)

	monthRange := latestEpoch - 7200
	if latestEpoch < 7200 {
		monthRange = 0
	}
	validatorsPQArray := pq.Array(queryIndices)

	var execBlocks []types.ExecBlockProposer
	err := db.ReaderDb.Select(&execBlocks,
		`SELECT 
			exec_block_number, 
			proposer 
			FROM blocks 
		WHERE proposer = ANY($1) 
		AND exec_block_number IS NOT NULL 
		AND exec_block_number > 0 
		AND epoch > $2`,
		validatorsPQArray,
		monthRange, // 32d range
	)
	if err != nil {
		return nil, fmt.Errorf("error cannot get proposed blocks from db with indicies: %+v and epoch: %v, err: %w", queryIndices, latestEpoch, err)
	}

	blockList, blockToProposerMap := getBlockNumbersAndMapProposer(execBlocks)

	blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, 10000)
	if err != nil {
		return nil, fmt.Errorf("error cannot get blocks from bigtable using GetBlocksIndexedMultiple: %w", err)
	}

	resultPerProposer := make(map[uint64]types.ExecutionPerformanceResponse)

	relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
	if err != nil {
		return nil, fmt.Errorf("error can not get relays data: %w", err)
	}

	type LongPerformanceResponse struct {
		Performance365d  string `db:"el_performance_365d" json:"performance365d"`
		PerformanceTotal string `db:"el_performance_total" json:"performanceTotal"`
		ValidatorIndex   uint64 `db:"validatorindex" json:"validatorindex"`
	}

	performanceList := []LongPerformanceResponse{}

	err = db.ReaderDb.Select(&performanceList, `
		SELECT 
		validatorindex,
		CAST(COALESCE(mev_performance_365d, 0) AS text) AS el_performance_365d,
		CAST(COALESCE(mev_performance_total, 0) AS text) AS el_performance_total
		FROM validator_performance WHERE validatorindex = ANY($1)`, validatorsPQArray)
	if err != nil {
		return nil, fmt.Errorf("error can cl performance from db: %w", err)
	}
	for _, val := range performanceList {
		performance365d, _ := new(big.Int).SetString(val.Performance365d, 10)
		performanceTotal, _ := new(big.Int).SetString(val.PerformanceTotal, 10)
		resultPerProposer[val.ValidatorIndex] = types.ExecutionPerformanceResponse{
			Performance1d:    big.NewInt(0),
			Performance7d:    big.NewInt(0),
			Performance31d:   big.NewInt(0),
			Performance365d:  performance365d,
			PerformanceTotal: performanceTotal,
			ValidatorIndex:   val.ValidatorIndex,
		}
	}

	lastStatsDay, err := db.GetLastExportedStatisticDay()
	if err != nil {
		return nil, fmt.Errorf("error getting last statistic day: %w", err)
	}
	firstEpochTime := utils.EpochToTime((lastStatsDay + 1) * utils.EpochsPerDay())

	for _, block := range blocks {
		proposer := blockToProposerMap[block.Number].Proposer
		result, ok := resultPerProposer[proposer]
		if !ok {
			result = types.ExecutionPerformanceResponse{
				Performance1d:    big.NewInt(0),
				Performance7d:    big.NewInt(0),
				Performance31d:   big.NewInt(0),
				Performance365d:  big.NewInt(0),
				PerformanceTotal: big.NewInt(0),
				ValidatorIndex:   proposer,
			}
		}

		txFees := big.NewInt(0).SetBytes(block.TxReward)
		mev := big.NewInt(0).SetBytes(block.Mev)
		income := big.NewInt(0).Add(txFees, mev)

		var mevBribe *big.Int = big.NewInt(0)
		relayData, ok := relaysData[common.BytesToHash(block.Hash)]
		if ok {
			mevBribe = relayData.MevBribe.BigInt()
		}

		var producerReward *big.Int
		if mevBribe.Int64() == 0 {
			producerReward = income
		} else {
			producerReward = mevBribe
		}

		if block.Time.AsTime().Equal(firstEpochTime) || block.Time.AsTime().After(firstEpochTime) {
			result.PerformanceTotal = result.PerformanceTotal.Add(result.PerformanceTotal, producerReward)
			result.Performance365d = result.Performance365d.Add(result.Performance365d, producerReward)
		}
		if block.Time.AsTime().After(last31dTimestamp) {
			result.Performance31d = result.Performance31d.Add(result.Performance31d, producerReward)
		}
		if block.Time.AsTime().After(last7dTimestamp) {
			result.Performance7d = result.Performance7d.Add(result.Performance7d, producerReward)
		}
		if block.Time.AsTime().After(last1dTimestamp) {
			result.Performance1d = result.Performance1d.Add(result.Performance1d, producerReward)
		}

		resultPerProposer[proposer] = result
	}

	return maps.Values(resultPerProposer), nil
}

func findExecBlockNumbersByProposerIndex(indices []uint64, offset, limit uint64, isSortAsc bool, onlyFinalized bool) ([]uint64, map[uint64]types.ExecBlockProposer, error) {
	var blockListSub []types.ExecBlockProposer

	order := "DESC"
	if isSortAsc {
		order = "ASC"
	}

	status := ""
	if onlyFinalized {
		status = `and status = '1'`
	}

	query := fmt.Sprintf(`SELECT 
			exec_block_number,
			proposer,
			slot,
			epoch  
		FROM blocks 
		WHERE proposer = ANY($1)
		AND exec_block_number IS NOT NULL AND exec_block_number > 0 %s
		ORDER BY exec_block_number %s
		OFFSET $2 LIMIT $3`, status, order)

	err := db.ReaderDb.Select(&blockListSub,
		query,
		pq.Array(indices),
		offset,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	blockList, blockProposerMap := getBlockNumbersAndMapProposer(blockListSub)
	return blockList, blockProposerMap, nil
}

func findExecBlockNumbersByFeeRecipient(addresses [][]byte, offset, limit uint64, isSortAsc bool) ([]uint64, map[uint64]types.ExecBlockProposer, error) {
	var blockListSub []types.ExecBlockProposer

	order := "DESC"
	if isSortAsc {
		order = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT 
			exec_block_number,
			proposer,
			slot,
			epoch  
		FROM blocks 
		WHERE exec_fee_recipient = ANY($1)
		AND exec_block_number IS NOT NULL AND exec_block_number > 0 
		ORDER BY exec_block_number %s
		OFFSET $2 LIMIT $3`, order)

	err := db.ReaderDb.Select(&blockListSub,
		query,
		pq.ByteaArray(addresses),
		offset,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	blockList, blockProposerMap := getBlockNumbersAndMapProposer(blockListSub)
	return blockList, blockProposerMap, nil
}

func findExecBlockNumbersByExecBlockNumber(execBlocks []uint64, offset, limit uint64) ([]uint64, map[uint64]types.ExecBlockProposer, error) {
	var blockListSub []types.ExecBlockProposer
	err := db.ReaderDb.Select(&blockListSub,
		`SELECT 
			exec_block_number,
			proposer,
			slot,
			epoch  
		FROM blocks 
		WHERE exec_block_number = ANY($1)
		AND exec_block_number IS NOT NULL AND exec_block_number > 0 
		ORDER BY exec_block_number DESC
		OFFSET $2 LIMIT $3`,
		pq.Array(execBlocks),
		offset,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	blockList, blockProposerMap := getBlockNumbersAndMapProposer(blockListSub)
	return blockList, blockProposerMap, nil
}

func getBlockNumbersAndMapProposer(data []types.ExecBlockProposer) ([]uint64, map[uint64]types.ExecBlockProposer) {
	blockList := []uint64{}
	blockToProposerMap := make(map[uint64]types.ExecBlockProposer)
	for _, execBlock := range data {
		blockList = append(blockList, execBlock.ExecBlock)
		blockToProposerMap[execBlock.ExecBlock] = execBlock
	}
	return blockList, blockToProposerMap
}

func resolveIndices(pubkeys [][]byte) ([]uint64, error) {
	indicesFromPubkeys := []uint64{}
	err := db.ReaderDb.Select(&indicesFromPubkeys,
		"SELECT validatorindex FROM validators WHERE pubkey = ANY($1)",
		pq.ByteaArray(pubkeys),
	)
	return indicesFromPubkeys, err
}

func getAddressesOrIndicesFromAddressIndexOrPubkey(search string, max int) ([][]byte, []uint64, error) {
	individuals := strings.Split(search, ",")
	if len(individuals) > max {
		return nil, nil, fmt.Errorf("only a maximum of %v query parameters are allowed", max)
	}
	var resultAddresses [][]byte

	var indices []uint64
	var pubkeys [][]byte
	for _, individual := range individuals {
		addInPub, err := parseFromAddressIndexOrPubkey(individual)
		if err != nil {
			return nil, nil, err
		}
		if len(addInPub.Address) > 0 {
			resultAddresses = append(resultAddresses, addInPub.Address)
		} else if len(addInPub.Pubkey) > 0 {
			pubkeys = append(pubkeys, addInPub.Pubkey)
		} else if addInPub.Index > 0 {
			indices = append(indices, addInPub.Index)
		}
	}

	// resolve pubkeys to index
	if len(pubkeys) > 0 {
		indicesFromPubkeys, err := resolveIndices(pubkeys)
		if err != nil {
			return nil, nil, err
		}
		indices = append(indices, indicesFromPubkeys...)
	}

	if len(indices) > 0 {
		return resultAddresses, indices, nil
	}

	return resultAddresses, nil, nil
}

func parseFromAddressIndexOrPubkey(search string) (types.AddressIndexOrPubkey, error) {
	if strings.Contains(search, "0x") && len(search) == 42 {
		address, err := hex.DecodeString(search[2:])
		if err != nil {
			return types.AddressIndexOrPubkey{}, err
		}
		return types.AddressIndexOrPubkey{
			Address: address,
		}, nil
	} else if strings.Contains(search, "0x") || len(search) == 96 {
		if len(search) < 94 {
			return types.AddressIndexOrPubkey{}, fmt.Errorf("invalid pubkey")
		}
		start := 2
		if len(search) == 96 {
			start = 0
		}
		pubkey, err := hex.DecodeString(search[start:])
		if err != nil {
			return types.AddressIndexOrPubkey{}, err
		}
		return types.AddressIndexOrPubkey{
			Pubkey: pubkey,
		}, nil
	} else {
		index, err := strconv.ParseUint(search, 10, 64)
		if err != nil {
			return types.AddressIndexOrPubkey{}, err
		}
		return types.AddressIndexOrPubkey{
			Index: index,
		}, nil
	}
}
