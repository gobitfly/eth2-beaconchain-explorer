package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"golang.org/x/exp/maps"
)

// ApiETH1ExecBlocks godoc
// @Summary Get Execution Blocks
// @Tags Execution
// @Description Get execution blocks by execution block number
// @Produce json
// @Param blockNumber path string true "Provide one or more execution block numbers. Coma separated up to max 100. "
// @Success 200 {object} string
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

	relaysData, err := getRelayDataForIndexedBlocks(blocks)
	if err != nil {
		logger.Errorf("can not load mev data %v", err)
		sendErrorResponse(w, r.URL.String(), "can not retrieve mev data")
		return
	}

	results := formatBlocksForApiResponse(blocks, relaysData, beaconDataMap)

	j := json.NewEncoder(w)
	sendOKResponse(j, r.URL.String(), []interface{}{results})
}

// ApiETH1AccountProposedBlocks godoc
// @Summary Get Proposed Blocks
// @Tags Execution
// @Description Get a list of proposed or mined blocks from a given fee recipient address, proposer index or proposer pubkey
// @Produce json
// @Param addressIndexOrPubkey path string true "Either the fee recipient address, the proposer index or proposer pubkey. You can provide multiple by seperating them with ',' up to max 20."
// @Param offset query int false "Offset"
// @Param limit query int false "Limit, amount of entries you wish to receive"
// @Success 200 {object} string
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/execution/{addressIndexOrPubkey}/proposed [get]
func ApiETH1AccountProducedBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	addresses, indices, err := getAddressesOrIndicesFromAddressIndexOrPubkey(vars["addressIndexOrPubkey"])
	if err != nil {
		sendErrorResponse(w, r.URL.String(), "invalid address, validator index or pubkey or exceeded max of 20 params")
		return
	}

	var offset uint64 = 0
	var limit uint64 = 10

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

	var blockList []uint64
	var beaconDataMap = map[uint64]types.ExecBlockProposer{}
	if len(addresses) > 0 {
		blockListSub, beaconDataMapSub, err := findExecBlockNumbersByFeeRecipient(addresses, offset, limit)
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
		blockListSub, beaconDataMapSub, err := findExecBlockNumbersByProposerIndex(indices, offset, limit)
		if err != nil {
			sendErrorResponse(w, r.URL.String(), "can not retrieve blocks from database")
			return
		}
		blockList = append(blockList, blockListSub...)
		for key, val := range beaconDataMapSub {
			beaconDataMap[key] = val
		}
	}

	blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, uint64(limit))
	if err != nil {
		logger.Errorf("Can not retrieve blocks from bigtable %v", err)
		sendErrorResponse(w, r.URL.String(), "can not retrieve blocks from bigtable")
		return
	}

	relaysData, err := getRelayDataForIndexedBlocks(blocks)
	if err != nil {
		logger.Errorf("can not load mev data %v", err)
		sendErrorResponse(w, r.URL.String(), "can not retrieve mev data")
		return
	}

	results := formatBlocksForApiResponse(blocks, relaysData, beaconDataMap)

	j := json.NewEncoder(w)
	sendOKResponse(j, r.URL.String(), []interface{}{results})
}

func getRelayDataForIndexedBlocks(blocks []*types.Eth1BlockIndexed) (map[common.Hash]types.RelaysData, error) {
	var execBlockHashes [][]byte
	var relaysData []types.RelaysData

	for _, block := range blocks {
		execBlockHashes = append(execBlockHashes, block.Hash)
	}
	// try to get mev rewards from relys_blocks table
	err := db.ReaderDb.Select(&relaysData,
		`SELECT proposer_fee_recipient, value, exec_block_hash, tag_id, builder_pubkey FROM relays_blocks WHERE relays_blocks.exec_block_hash = ANY($1)`,
		pq.ByteaArray(execBlockHashes),
	)
	if err != nil {
		return nil, err
	}
	var relaysDataMap = make(map[common.Hash]types.RelaysData)
	for _, relayData := range relaysData {
		relaysDataMap[common.BytesToHash(relayData.ExecBlockHash)] = relayData
	}

	return relaysDataMap, nil
}

func formatBlocksForApiResponse(blocks []*types.Eth1BlockIndexed, relaysData map[common.Hash]types.RelaysData, beaconDataMap map[uint64]types.ExecBlockProposer) []types.ExecutionBlockApiResponse {
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
			BaseFee:            baseFee,
			TxCount:            block.GetTransactionCount(),
			InternalTxCount:    block.GetInternalTransactionCount(),
			UncleCount:         block.GetUncleCount(),
			ParentHash:         fmt.Sprintf("0x%v", hex.EncodeToString(block.GetParentHash())),
			UncleHash:          fmt.Sprintf("0x%v", hex.EncodeToString(block.GetUncleHash())),
			Difficulty:         difficulty,
			PoSData:            posDataPt,
			RelayData:          relayDataResponse,
			ConsensusAlgorithm: consensusAlgorithm,
		})
	}

	return results
}

func getValidatorExecutionPerformance(queryIndices []uint64) ([]types.ExecutionPerformanceResponse, error) {
	latestEpoch := services.LatestEpoch()
	last30dTimestamp := time.Now().Add(-31 * 24 * time.Hour)
	last7dTimestamp := time.Now().Add(-7 * 24 * time.Hour)
	last1dTimestamp := time.Now().Add(-1 * 24 * time.Hour)

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
		pq.Array(queryIndices),
		latestEpoch-7200, // 32d range
	)
	if err != nil {
		logger.WithError(err).Error("can not load proposed blocks from db")
		return nil, err
	}

	blockList, blockToProposerMap := getBlockNumbersAndMapProposer(execBlocks)

	blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, 10000)
	if err != nil {
		logger.WithError(err).Errorf("can not load mined blocks by GetBlocksIndexedMultiple")
		return nil, err
	}

	resultPerProposer := make(map[uint64]types.ExecutionPerformanceResponse)

	relaysData, err := getRelayDataForIndexedBlocks(blocks)
	if err != nil {
		logger.WithError(err).Errorf("can not get relays data")
		return nil, err
	}

	for _, block := range blocks {
		proposer := blockToProposerMap[block.Number].Proposer
		result, ok := resultPerProposer[proposer]
		if !ok {
			result = types.ExecutionPerformanceResponse{
				Performance1d:  big.NewInt(0),
				Performance7d:  big.NewInt(0),
				Performance31d: big.NewInt(0),
				ValidatorIndex: proposer,
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

		if block.Time.AsTime().After(last30dTimestamp) {
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

func findExecBlockNumbersByProposerIndex(indices []uint64, offset, limit uint64) ([]uint64, map[uint64]types.ExecBlockProposer, error) {
	var blockListSub []types.ExecBlockProposer
	err := db.ReaderDb.Select(&blockListSub,
		`SELECT 
			exec_block_number,
			proposer,
			slot,
			epoch  
		FROM blocks 
		WHERE proposer = ANY($1)
		AND exec_block_number IS NOT NULL AND exec_block_number > 0 
		ORDER BY exec_block_number DESC
		OFFSET $2 LIMIT $3`,
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

func findExecBlockNumbersByFeeRecipient(addresses [][]byte, offset, limit uint64) ([]uint64, map[uint64]types.ExecBlockProposer, error) {
	var blockListSub []types.ExecBlockProposer
	err := db.ReaderDb.Select(&blockListSub,
		`SELECT 
				exec_block_number,
				proposer,
				slot,
				epoch  
			FROM blocks 
			WHERE exec_fee_recipient = ANY($1)
			AND exec_block_number IS NOT NULL AND exec_block_number > 0 
			ORDER BY exec_block_number DESC
			OFFSET $2 LIMIT $3`,
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

func getAddressesOrIndicesFromAddressIndexOrPubkey(search string) ([][]byte, []uint64, error) {
	individuals := strings.Split(search, ",")
	if len(individuals) > 20 {
		return nil, nil, fmt.Errorf("only a maximum of 20 query parameters are allowed")
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
		indicesFromPubkeys := []uint64{}
		err := db.ReaderDb.Select(&indicesFromPubkeys,
			"SELECT validatorindex FROM validators WHERE pubkey = ANY($1)",
			pq.ByteaArray(pubkeys),
		)
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
		pubkey, err := hex.DecodeString(search[2:])
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
