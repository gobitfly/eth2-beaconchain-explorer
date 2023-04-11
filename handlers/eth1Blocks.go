package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
)

func Eth1Blocks(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "execution/blocks.html")
	var eth1BlocksTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "blockchain", "/eth1blocks", "Ethereum Blocks", templateFiles)
	if handleTemplateError(w, r, "eth1Blocks.go", "Eth1Blocks", "", eth1BlocksTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func Eth1BlocksData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	recordsTotal, err := strconv.ParseUint(q.Get("recordsTotal"), 10, 64)
	if err != nil {
		recordsTotal = 0
	}
	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int for route %v: %v", r.URL.String(), err)
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

	data, err := getEth1BlocksTableData(draw, start, length, recordsTotal)
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

type additionalSlotData struct {
	Epoch        uint64 `db:"epoch"`
	Slot         uint64 `db:"slot"`
	Proposer     uint64 `db:"proposer"`
	Status       uint64 `db:"status"`
	ProposerName string `db:"name"`
}

func getSlotByTimestamp(t *timestamp.Timestamp) uint64 {
	ts := uint64(t.AsTime().Unix())
	if ts >= utils.Config.Chain.GenesisTimestamp {
		return (ts - utils.Config.Chain.GenesisTimestamp) / utils.Config.Chain.Config.SecondsPerSlot
	}
	return 0
}

func getProposerAndStatusFromSlot(startSlot uint64, endSlot uint64) (map[uint64]*additionalSlotData, error) {
	data := make(map[uint64]*additionalSlotData)

	var blocks []*additionalSlotData
	err := db.ReaderDb.Select(&blocks, `
		SELECT 
			blocks.epoch, 
			blocks.slot,
			blocks.proposer,
			blocks.status,
			COALESCE(validator_names.name, '') AS name
		FROM blocks 
		LEFT JOIN validators ON blocks.proposer = validators.validatorindex
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		WHERE blocks.slot >= $1 AND blocks.slot <= $2
		ORDER BY blocks.slot DESC`, startSlot, endSlot)
	if err != nil {
		return nil, err
	}

	for _, v := range blocks {
		data[v.Slot] = v
	}
	return data, nil
}

func Eth1BlocksHighest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/text")
	w.Write([]byte(fmt.Sprintf("%d", services.LatestEth1BlockNumber())))
}

func getEth1BlocksTableData(draw, start, length, recordsTotal uint64) (*types.DataTableResponse, error) {
	if recordsTotal == 0 {
		recordsTotal = services.LatestEth1BlockNumber()
	}

	if start > recordsTotal {
		start = 1
	} else {
		start = recordsTotal - start
	}

	if length > start {
		length = start
	}

	blocks, err := db.BigtableClient.GetBlocksDescending(start, length)
	if err != nil {
		return nil, err
	}

	var slotData map[uint64]*additionalSlotData
	{
		foundAtLeastOneValidSlot := false
		startSlot := ^uint64(0)
		endSlot := uint64(0)
		for _, b := range blocks {
			s := getSlotByTimestamp(b.GetTime())
			if s > 0 {
				foundAtLeastOneValidSlot = true
				if s < startSlot {
					startSlot = s
				}
				if s > endSlot {
					endSlot = s
				}
			}
		}

		if foundAtLeastOneValidSlot {
			slotData, err = getProposerAndStatusFromSlot(startSlot, endSlot)
			if err != nil {
				return nil, err
			}
		}
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		var sData *additionalSlotData
		if slotData != nil {
			ts := uint64(b.GetTime().AsTime().Unix())
			if ts >= utils.Config.Chain.GenesisTimestamp {
				slot := (ts - utils.Config.Chain.GenesisTimestamp) / utils.Config.Chain.Config.SecondsPerSlot
				if val, ok := slotData[slot]; ok {
					sData = val
				} else {
					// return nil, fmt.Errorf("slot %d doesn't exists in ReaderDb", slot)
					logrus.Infof("slot %d doesn't exists in ReaderDb", slot)
				}
			}
		}

		slotText := "-"
		epochText := "-"
		status := template.HTML("-")
		proposer := template.HTML("-")
		if sData != nil {
			status = utils.FormatBlockStatus(sData.Status)
			proposer = utils.FormatValidatorWithName(sData.Proposer, sData.ProposerName)

			posActive := true
			for _, v := range b.GetDifficulty() {
				if v != 0 {
					posActive = false
					break
				}
			}

			if posActive && sData != nil {
				slotText = fmt.Sprintf(`<A href="slot/%d">%s</A>`, sData.Slot, utils.FormatAddCommas(sData.Slot))
				epochText = fmt.Sprintf(`<A href="epoch/%d">%s</A>`, sData.Epoch, utils.FormatAddCommas(sData.Epoch))
			}
		}

		blockNumber := b.GetNumber()
		baseFee := new(big.Int).SetBytes(b.GetBaseFee())
		gasHalf := float64(b.GetGasLimit()) / 2.0
		txReward := new(big.Int).SetBytes(b.GetTxReward())

		burned := new(big.Int).Mul(baseFee, big.NewInt(int64(b.GetGasUsed())))
		burnedPercentage := float64(0.0)
		if len(txReward.Bits()) != 0 {
			txBurnedBig := new(big.Float).SetInt(burned)
			txBurnedBig.Quo(txBurnedBig, new(big.Float).SetInt(txReward))
			burnedPercentage, _ = txBurnedBig.Float64()
		}

		tableData[i] = []interface{}{
			epochText, // Epoch
			fmt.Sprintf(`%s<BR /><span style="font-size: .63rem; color: grey;">%v</span>`, slotText, utils.FormatTimestamp(b.GetTime().AsTime().Unix())), // Slot
			fmt.Sprintf(`<A href="block/%d">%v</A>`, blockNumber, utils.FormatAddCommas(blockNumber)),                                                    // Block
			status,                             // Status
			fmt.Sprintf("%x", b.GetCoinbase()), // Recipient
			proposer,                           // Proposer
			fmt.Sprintf(`<span data-toggle="tooltip" data-placement="top" title="%d transactions (%d internal transactions)">%d<BR /><span style="font-size: .63rem; color: grey;">%d</span></span>`, b.GetTransactionCount(), b.GetInternalTransactionCount(), b.GetTransactionCount(), b.GetInternalTransactionCount()),                                                                                                                                                                               // Transactions
			fmt.Sprintf(`%v<BR /><span data-toggle="tooltip" data-placement="top" title="Gas Used %%" style="font-size: .63rem; color: grey;">%.2f%%</span>&nbsp;<span data-toggle="tooltip" data-placement="top" title="%% of Gas Target" style="font-size: .63rem; color: grey;">(%+.2f%%)</span>`, utils.FormatAddCommas(b.GetGasUsed()), float64(int64(float64(b.GetGasUsed())/float64(b.GetGasLimit())*10000.0))/100.0, float64(int64(((float64(b.GetGasUsed())-gasHalf)/gasHalf)*10000.0))/100.0), // Gas Used
			utils.FormatAddCommas(b.GetGasLimit()),                               // Gas Limit
			utils.FormatAmountFormatted(baseFee, "GWei", 5, 4, true, true, true), // Base Fee
			utils.FormatAmountFormatted(new(big.Int).Add(utils.Eth1BlockReward(blockNumber, b.GetDifficulty()), new(big.Int).Add(txReward, new(big.Int).SetBytes(b.GetUncleReward()))), "Ether", 5, 4, true, true, true),                                                                         // Reward
			fmt.Sprintf(`%v<BR /><span data-toggle="tooltip" data-placement="top" title="%% of Transactions Fees" style="font-size: .63rem; color: grey;">%.2f%%</span>`, utils.FormatAmountFormatted(burned, "Ether", 5, 4, true, true, false), float64(int64(burnedPercentage*10000.0))/100.0), // Burned Fees
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsTotal,
		Data:            tableData,
	}

	return data, nil
}
