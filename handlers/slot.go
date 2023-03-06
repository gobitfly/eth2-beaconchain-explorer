package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/juliangruber/go-intersect"
	"github.com/lib/pq"

	"github.com/gorilla/mux"
)

// Hacky AF way to ensure intersect module is imported and not optimised away, unsure why its being optimised away
var _ = intersect.Simple

const MaxSlotValue = 137438953503 // we only render a page for blocks up to this slot

// Slot will return the data for a block contained in the slot
func Slot(w http.ResponseWriter, r *http.Request) {

	var slotTemplate = templates.GetTemplate(
		"layout.html",
		"slot/slot.html",
		"slot/transactions.html",
		"slot/withdrawals.html",
		"slot/attestations.html",
		"slot/deposits.html",
		"slot/votes.html",
		"slot/attesterSlashing.html",
		"slot/proposerSlashing.html",
		"slot/exits.html",
		"slot/overview.html",
		"slot/execTransactions.html",
	)

	var slotFutureTemplate = templates.GetTemplate(
		"layout.html",
		"slot/slotFuture.html",
	)

	var blockNotFoundTemplate = templates.GetTemplate("layout.html", "slotnotfound.html")

	w.Header().Set("Content-Type", "text/html")

	vars := mux.Vars(r)

	slotOrHash := strings.Replace(vars["slotOrHash"], "0x", "", -1)
	blockSlot := int64(-1)
	blockRootHash, err := hex.DecodeString(slotOrHash)
	if err != nil || len(slotOrHash) != 64 {
		blockRootHash = []byte{}
		blockSlot, err = strconv.ParseInt(vars["slotOrHash"], 10, 64)
		if err != nil {
			logger.Errorf("error parsing blockslot to int: %v", err)
			http.Error(w, "Internal server error", http.StatusServiceUnavailable)
			return
		}
	}

	data := InitPageData(w, r, "blockchain", "/slots", fmt.Sprintf("Slot %v", slotOrHash))

	if blockSlot == -1 {
		err = db.ReaderDb.Get(&blockSlot, `SELECT slot FROM blocks WHERE blockroot = $1 OR stateroot = $1 LIMIT 1`, blockRootHash)
		if blockSlot == -1 {
			if handleTemplateError(w, r, "slot.go", "Slot", "blockSlot", blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
				return // an error has occurred and was processed
			}
			return
		}
		if err != nil {
			logger.Errorf("error retrieving entry count of given block or state data: %v", err)
			http.Error(w, "Internal server error", http.StatusServiceUnavailable)
			return
		}
	}

	blockPageData, err := GetSlotPageData(uint64(blockSlot))
	if err == sql.ErrNoRows {
		slot := uint64(blockSlot)
		//Slot not in database -> Show future block
		data.Meta.Path = "/slot/" + slotOrHash

		if slot > MaxSlotValue {
			logger.Errorf("error retrieving blockPageData: %v", err)
			if handleTemplateError(w, r, "slot.go", "Slot", "MaxSlotValue", blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
				return // an error has occurred and was processed
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

		if handleTemplateError(w, r, "slot.go", "Slot", "ErrNoRows", slotFutureTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	} else if err != nil {
		if handleTemplateError(w, r, "slot.go", "Slot", "GetSlotPageData", err) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	if blockPageData.ExecBlockNumber.Int64 != 0 && blockPageData.Status == 1 {
		// slot has corresponding execution block, fetch execution data
		eth1BlockPageData, err := GetExecutionBlockPageData(uint64(blockPageData.ExecBlockNumber.Int64), 10)
		// if err != nil, simply show slot view without block
		if err == nil {
			blockPageData.ExecutionData = eth1BlockPageData
			blockPageData.ExecutionData.IsValidMev = blockPageData.IsValidMev
		}
	}
	data.Meta.Path = fmt.Sprintf("/slot/%v", blockPageData.Slot)
	data.Data = blockPageData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		w.Header().Set("Content-Type", "text/html")
		err = slotTemplate.ExecuteTemplate(w, "layout", data)
	}

	if handleTemplateError(w, r, "slot.go", "Slot", "ApiRequest", err) != nil {
		return // an error has occurred and was processed
	}
}

func getAttestationsData(slot uint64, onlyFirst bool) ([]*types.BlockPageAttestation, error) {
	limit := ";"
	if onlyFirst {
		limit = " LIMIT 1;"
	}

	var attestations []*types.BlockPageAttestation
	rows, err := db.ReaderDb.Query(`
		SELECT
			block_slot,
			block_index,
			aggregationbits,
			validators,
			signature,
			slot,
			committeeindex,
			beaconblockroot,
			source_epoch,
			source_root,
			target_epoch,
			target_root
		FROM blocks_attestations
		WHERE block_slot = $1
		ORDER BY block_index`+limit,
		slot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block attestation data: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		attestation := &types.BlockPageAttestation{}

		err := rows.Scan(
			&attestation.BlockSlot,
			&attestation.BlockIndex,
			&attestation.AggregationBits,
			&attestation.Validators,
			&attestation.Signature,
			&attestation.Slot,
			&attestation.CommitteeIndex,
			&attestation.BeaconBlockRoot,
			&attestation.SourceEpoch,
			&attestation.SourceRoot,
			&attestation.TargetEpoch,
			&attestation.TargetRoot)
		if err != nil {
			return nil, fmt.Errorf("error scanning block attestation data: %v", err)
		}
		attestations = append(attestations, attestation)
	}
	return attestations, nil
}

func GetSlotPageData(blockSlot uint64) (*types.BlockPageData, error) {
	blockPageData := types.BlockPageData{}
	blockPageData.Mainnet = utils.Config.Chain.Config.ConfigName == "mainnet"
	err := db.ReaderDb.Get(&blockPageData, `
		SELECT
			blocks.epoch,
			COALESCE(epochs.finalized, false) AS epoch_finalized,
			COALESCE(epochs.globalparticipationrate, 0) AS epoch_participation_rate,
			blocks.slot,
			blocks.blockroot,
			blocks.parentroot,
			blocks.stateroot,
			blocks.signature,
			blocks.randaoreveal,
			blocks.graffiti,
			blocks.eth1data_depositroot,
			blocks.eth1data_depositcount,
			blocks.eth1data_blockhash,
			blocks.syncaggregate_bits,
			blocks.syncaggregate_signature,
			blocks.syncaggregate_participation,
			blocks.proposerslashingscount,
			blocks.attesterslashingscount,
			blocks.attestationscount,
			blocks.depositscount,
			blocks.withdrawalcount,
			blocks.voluntaryexitscount,
			blocks.proposer,
			blocks.status,
			exec_block_number,
			jsonb_agg(tags.metadata) as tags,
			COALESCE(not 'invalid-relay-reward'=ANY(array_agg(tags.id)), true) as is_valid_mev,
			COALESCE(validator_names.name, '') AS name,
			(SELECT count(*) from blocks_bls_change where block_slot = $1) as bls_change_count
		FROM blocks 
		LEFT JOIN validators ON blocks.proposer = validators.validatorindex
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		LEFT JOIN blocks_tags ON blocks.slot = blocks_tags.slot and blocks.blockroot = blocks_tags.blockroot
		LEFT JOIN tags ON blocks_tags.tag_id = tags.id
		LEFT JOIN epochs ON blocks.epoch = epochs.epoch
		WHERE blocks.slot = $1 
		group by
			blocks.epoch,
			blocks.slot,
			blocks.blockroot,
			validator_names."name",
			epoch_finalized,
			epoch_participation_rate
		ORDER BY blocks.blockroot DESC, blocks.status ASC limit 1
		`,
		blockSlot)
	if err != nil {
		return nil, err
	}
	blockPageData.Slot = uint64(blockSlot)

	blockPageData.Ts = utils.SlotToTime(blockPageData.Slot)
	if blockPageData.ExecTimestamp.Valid {
		blockPageData.ExecTime = time.Unix(int64(blockPageData.ExecTimestamp.Int64), 0)
	}
	blockPageData.SlashingsCount = blockPageData.AttesterSlashingsCount + blockPageData.ProposerSlashingsCount

	blockPageData.NextSlot = blockPageData.Slot + 1
	blockPageData.PreviousSlot = blockPageData.Slot - 1

	blockPageData.Attestations, err = getAttestationsData(blockPageData.Slot, true)
	if err != nil {
		return nil, err
	}

	rows, err := db.ReaderDb.Query(`
		SELECT validators
		FROM blocks_attestations
		WHERE beaconblockroot = $1`,
		blockPageData.BlockRoot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block votes data: %v", err)
	}
	defer rows.Close()

	votesCount := 0
	votesPerValidator := map[int64]int{}
	for rows.Next() {
		validators := pq.Int64Array{}
		err := rows.Scan(&validators)
		if err != nil {
			return nil, fmt.Errorf("error scanning votes validators data: %v", err)
		}
		for _, validator := range validators {
			votesCount++
			_, exists := votesPerValidator[validator]
			if !exists {
				votesPerValidator[validator] = 1
			} else {
				votesPerValidator[validator]++
			}
		}
	}
	blockPageData.VotingValidatorsCount = uint64(len(votesPerValidator))
	blockPageData.VotesCount = uint64(votesCount)

	err = db.ReaderDb.Select(&blockPageData.VoluntaryExits, "SELECT validatorindex, signature FROM blocks_voluntaryexits WHERE block_slot = $1", blockPageData.Slot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block deposit data: %v", err)
	}

	err = db.ReaderDb.Select(&blockPageData.AttesterSlashings, `
		SELECT
			block_slot,
			block_index,
			attestation1_indices,
			attestation1_signature,
			attestation1_slot,
			attestation1_index,
			attestation1_beaconblockroot,
			attestation1_source_epoch,
			attestation1_source_root,
			attestation1_target_epoch,
			attestation1_target_root,
			attestation2_indices,
			attestation2_signature,
			attestation2_slot,
			attestation2_index,
			attestation2_beaconblockroot,
			attestation2_source_epoch,
			attestation2_source_root,
			attestation2_target_epoch,
			attestation2_target_root
		FROM blocks_attesterslashings
		WHERE block_slot = $1`, blockPageData.Slot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block attester slashings data: %v", err)
	}
	if len(blockPageData.AttesterSlashings) > 0 {
		for _, slashing := range blockPageData.AttesterSlashings {
			inter := intersect.Simple(slashing.Attestation1Indices, slashing.Attestation2Indices)

			for _, i := range inter {
				slashing.SlashedValidators = append(slashing.SlashedValidators, i.(int64))
			}
		}
	}

	err = db.ReaderDb.Select(&blockPageData.ProposerSlashings, "SELECT block_slot, block_index, block_root, proposerindex, header1_slot, header1_parentroot, header1_stateroot, header1_bodyroot, header1_signature, header2_slot, header2_parentroot, header2_stateroot, header2_bodyroot, header2_signature FROM blocks_proposerslashings WHERE block_slot = $1", blockPageData.Slot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block proposer slashings data: %v", err)
	}

	// TODO: fix blockPageData data type to include SyncCommittee
	err = db.ReaderDb.Select(&blockPageData.SyncCommittee, "SELECT validatorindex FROM sync_committees WHERE period = $1 ORDER BY committeeindex", utils.SyncPeriodOfEpoch(blockPageData.Epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving sync-committee of block %v: %v", blockPageData.Slot, err)
	}

	// old code retrieving txs from postgres db
	/* if retrieveTxsFromDb {
			// retrieve transactions from db
			var transactions []*types.BlockPageTransaction
			rows, err = db.ReaderDb.Query(`
				SELECT
	    		block_slot,
	    		block_index,
	    		txhash,
	    		nonce,
	    		gas_price,
	    		gas_limit,
	    		sender,
	    		recipient,
	    		amount,
	    		payload
				FROM blocks_transactions
				WHERE block_slot = $1
				ORDER BY block_index`,
				blockPageData.Slot)
			if err != nil {
				return nil, fmt.Errorf("error retrieving block transaction data: %v", err)
			}
			defer rows.Close()

			for rows.Next() {
				tx := &types.BlockPageTransaction{}

				err := rows.Scan(
					&tx.BlockSlot,
					&tx.BlockIndex,
					&tx.TxHash,
					&tx.AccountNonce,
					&tx.Price,
					&tx.GasLimit,
					&tx.Sender,
					&tx.Recipient,
					&tx.Amount,
					&tx.Payload,
				)
				if err != nil {
					return nil, fmt.Errorf("error scanning block transaction data: %v", err)
				}
				var amount, price big.Int
				amount.SetBytes(tx.Amount)
				price.SetBytes(tx.Price)
				tx.AmountPretty = ToEth(&amount)
				tx.PricePretty = ToGWei(&amount)
				transactions = append(transactions, tx)
			}
			blockPageData.Transactions = transactions
		}
	*/
	return &blockPageData, nil
}

// SlotDepositData returns the deposits for a specific slot
func SlotDepositData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slotOrHash := strings.Replace(vars["slotOrHash"], "0x", "", -1)
	blockSlot := int64(-1)
	blockRootHash, err := hex.DecodeString(slotOrHash)
	if err != nil || len(slotOrHash) != 64 {
		blockSlot, err = strconv.ParseInt(vars["slotOrHash"], 10, 64)
		if err != nil {
			logger.Errorf("error parsing slotOrHash url parameter %v, err: %v", vars["slotOrHash"], err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.ReaderDb.Get(&blockSlot, `
		SELECT
			blocks.slot
		FROM blocks
		WHERE blocks.blockroot = $1
		`, blockRootHash)
		if err != nil {
			logger.Errorf("error querying for block slot with block root hash %v err: %v", blockRootHash, err)
			http.Error(w, "Interal server error", http.StatusInternalServerError)
			return
		}
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if length > 100 {
		length = 100
	}

	var count uint64

	err = db.ReaderDb.Get(&count, `
	SELECT 
		count(*)
	FROM
		blocks_deposits
	WHERE
	 block_slot = $1
	GROUP BY
	 block_slot
	`, blockSlot)
	if err != nil {
		logger.Errorf("error retrieving deposit count for slot %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var deposits []*types.BlockPageDeposit

	err = db.ReaderDb.Select(&deposits, `
		SELECT
			publickey,
			withdrawalcredentials,
			amount,
			signature
		FROM blocks_deposits
		WHERE block_slot = $1
		ORDER BY block_index
		LIMIT $2
		OFFSET $3`,
		blockSlot, length, start)
	if err != nil {
		logger.Errorf("error retrieving block deposit data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, 0, len(deposits))

	for i, deposit := range deposits {
		tableData = append(tableData, []interface{}{
			i + 1 + int(start),
			utils.FormatPublicKey(deposit.PublicKey),
			utils.FormatBalance(deposit.Amount, currency),
			deposit.WithdrawalCredentials,
			deposit.Signature,
		})
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    count,
		RecordsFiltered: count,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error encoding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

// SlotVoteData returns the votes for a specific slot
func SlotVoteData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slotOrHash := strings.Replace(vars["slotOrHash"], "0x", "", -1)
	blockSlot := int64(-1)
	blockRootHash, err := hex.DecodeString(slotOrHash)
	if err != nil || len(slotOrHash) != 64 {
		blockSlot, err = strconv.ParseInt(vars["slotOrHash"], 10, 64)
		if err != nil {
			logger.Errorf("error parsing slotOrHash url parameter %v, err: %v", vars["slotOrHash"], err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		err = db.ReaderDb.Get(&blockRootHash, "select blocks.blockroot from blocks where blocks.slot = $1", blockSlot)
		if err != nil {
			logger.Errorf("error getting blockRootHash for slot %v: %v", blockSlot, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.ReaderDb.Get(&blockSlot, `SELECT blocks.slot FROM blocks WHERE blocks.blockroot = $1 OR blocks.stateroot = $1`, blockRootHash)
		if err != nil {
			logger.Errorf("error querying for block slot with block root hash %v err: %v", blockRootHash, err)
			http.Error(w, "Interal server error", http.StatusInternalServerError)
			return
		}
	}

	q := r.URL.Query()

	search := q.Get("search[value]")
	searchIsUint64 := false
	searchUint64, err := strconv.ParseUint(search, 10, 64)
	if err == nil {
		searchIsUint64 = true
	}

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if length > 100 {
		length = 100
	}

	var count uint64
	var votes []struct {
		AllocatedSlot  uint64        `db:"slot"` // "the slot during which the validators were allocated to vote for a block"
		CommitteeIndex uint64        `db:"committeeindex"`
		BlockSlot      uint64        `db:"block_slot"` // "the block where the attestation was included in"
		Validators     pq.Int64Array `db:"validators"`
	}
	if search == "" {
		err = db.ReaderDb.Get(&count, `SELECT count(*) FROM blocks_attestations WHERE beaconblockroot = $1`, blockRootHash)
		if err != nil {
			logger.Errorf("error retrieving deposit count for slot %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		err = db.ReaderDb.Select(&votes, `
			SELECT
				slot,
				committeeindex,
				block_slot,
				validators
			FROM blocks_attestations
			WHERE beaconblockroot = $1
			ORDER BY committeeindex
			LIMIT $2
			OFFSET $3`,
			blockRootHash, length, start)
		if err != nil {
			logger.Errorf("error retrieving block vote data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else if searchIsUint64 {
		err = db.ReaderDb.Get(&count, `SELECT count(*) FROM blocks_attestations WHERE beaconblockroot = $1 AND $2 = ANY(validators)`, blockRootHash, searchUint64)
		if err != nil {
			logger.Errorf("error retrieving deposit count for slot %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		err = db.ReaderDb.Select(&votes, `
			SELECT
				slot,
				committeeindex,
				block_slot,
				validators
			FROM blocks_attestations
			WHERE beaconblockroot = $1 AND $2 = ANY(validators)
			ORDER BY committeeindex
			LIMIT $3
			OFFSET $4`,
			blockRootHash, searchUint64, length, start)
		if err != nil {
			logger.Errorf("error retrieving block vote data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	tableData := make([][]interface{}, 0, len(votes))

	for _, vote := range votes {
		formatedValidators := make([]string, len(vote.Validators))
		for i, v := range vote.Validators {
			formatedValidators[i] = fmt.Sprintf("<a href='/validator/%[1]d'>%[1]d</a>", v)
		}
		tableData = append(tableData, []interface{}{
			vote.AllocatedSlot,
			vote.CommitteeIndex,
			vote.BlockSlot,
			strings.Join(formatedValidators, ", "),
		})
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    count,
		RecordsFiltered: count,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error encoding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

type transactionsData struct {
	HashFormatted template.HTML `json:"HashFormatted"`
	Method        string        `json:"Method"`
	FromFormatted template.HTML `json:"FromFormatted"`
	ToFormatted   template.HTML `json:"ToFormatted"`
	Value         template.HTML `json:"Value"`
	Fee           template.HTML `json:"Fee"`
	GasPrice      template.HTML `json:"GasPrice"`
}

// BlockTransactionsData returns the transactions for a specific block
func BlockTransactionsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slot, err := strconv.ParseUint(vars["block"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing slot url parameter %v, err: %v", vars["slot"], err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	transactions, err := GetExecutionBlockPageData(slot, 0)
	if err != nil || transactions == nil {
		logger.Errorf("error retrieving transactions data for slot %v, err: %v", slot, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := make([]*transactionsData, len(transactions.Txs))
	for i, v := range transactions.Txs {
		if len(v.Method) == 0 {
			v.Method = "Transfer"
		}
		data[i] = &transactionsData{
			HashFormatted: v.HashFormatted,
			Method:        `<span class="badge badge-light">` + v.Method + `</span>`,
			FromFormatted: v.FromFormatted,
			ToFormatted:   v.ToFormatted,
			Value:         utils.FormatAmountFormated(v.Value, "ETH", 5, 0, true, true, false),
			Fee:           utils.FormatAmountFormated(v.Fee, "ETH", 5, 0, true, true, false),
			GasPrice:      utils.FormatAmountFormated(v.GasPrice, "GWei", 5, 0, true, true, false),
		}
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error encoding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

type attestationsData struct {
	BlockIndex      uint64        `json:"BlockIndex"`
	Slot            uint64        `json:"Slot"`
	CommitteeIndex  uint64        `json:"CommitteeIndex"`
	AggregationBits template.HTML `json:"AggregationBits"`
	Validators      template.HTML `json:"Validators"`
	BeaconBlockRoot string        `json:"BeaconBlockRoot"`
	SourceEpoch     uint64        `json:"SourceEpoch"`
	SourceRoot      string        `json:"SourceRoot"`
	TargetEpoch     uint64        `json:"TargetEpoch"`
	TargetRoot      string        `json:"TargetRoot"`
	Signature       string        `json:"Signature"`
}

// SlotAttestationsData returns the attestations for a specific slot
func SlotAttestationsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slot, err := strconv.ParseUint(vars["slot"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing slot url parameter %v, err: %v", vars["slot"], err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	attestations, err := getAttestationsData(slot, false)
	if err != nil {
		logger.Errorf("error retrieving attestations data for slot %v, err: %v", slot, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := make([]*attestationsData, len(attestations))
	for i, v := range attestations {
		var validators template.HTML
		for _, val := range v.Validators {
			validators += utils.FormatValidatorInt64(val) + " "
		}
		data[i] = &attestationsData{
			BlockIndex:      v.BlockIndex,
			Slot:            v.Slot,
			CommitteeIndex:  v.CommitteeIndex,
			AggregationBits: utils.FormatBitlist(v.AggregationBits),
			Validators:      validators,
			BeaconBlockRoot: fmt.Sprintf("%x", v.BeaconBlockRoot),
			SourceEpoch:     v.SourceEpoch,
			SourceRoot:      fmt.Sprintf("%x", v.SourceRoot),
			TargetEpoch:     v.TargetEpoch,
			TargetRoot:      fmt.Sprintf("%x", v.TargetRoot),
			Signature:       fmt.Sprintf("%x", v.Signature),
		}
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error encoding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func SlotWithdrawalData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	slot, err := strconv.ParseUint(vars["slot"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing slot url parameter %v, err: %v", vars["slot"], err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	withdrawals, err := db.GetSlotWithdrawals(slot)
	if err != nil {
		logger.Errorf("error retrieving withdrawals data for slot %v, err: %v", slot, err)
	}

	tableData := make([][]interface{}, 0, len(withdrawals))
	for _, w := range withdrawals {
		// logger.Infof("w: %+v", w)
		tableData = append(tableData, []interface{}{
			template.HTML(fmt.Sprintf("%v", w.Index)),
			template.HTML(fmt.Sprintf("%v", utils.FormatValidator(w.ValidatorIndex))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAddress(w.Address, nil, "", false, false, true))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), "ETH", 6))),
		})
	}

	data := &types.DataTableResponse{
		Draw:         1,
		RecordsTotal: uint64(len(withdrawals)),
		// RecordsFiltered: uint64(len(withdrawals)),
		Data: tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error encoding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func SlotBlsChangeData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	slot, err := strconv.ParseUint(vars["slot"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing slot url parameter %v, err: %v", vars["slot"], err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	blsChange, err := db.GetSlotBLSChange(slot)
	if err != nil {
		logger.Errorf("error retrieving blsChange data for slot %v, err: %v", slot, err)
	}

	tableData := make([][]interface{}, 0, len(blsChange))
	for _, c := range blsChange {
		tableData = append(tableData, []interface{}{
			template.HTML(fmt.Sprintf("%v", utils.FormatValidator(c.Validatorindex))),
			template.HTML(fmt.Sprintf("%v", utils.FormatHashWithCopy(c.Signature))),
			template.HTML(fmt.Sprintf("%v", utils.FormatHashWithCopy(c.BlsPubkey))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAddress(c.Address, nil, "", false, false, true))),
		})
	}

	data := &types.DataTableResponse{
		Draw:         1,
		RecordsTotal: uint64(len(blsChange)),
		// RecordsFiltered: uint64(len(withdrawals)),
		Data: tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error encoding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ToWei converts the big.Int wei to its gwei string representation.
func ToWei(wei *big.Int) string {
	return wei.String()
}

// ToGWei converts the big.Int wei to its gwei string representation.
func ToGWei(wei *big.Int) string {
	return ToEth(new(big.Int).Mul(wei, big.NewInt(1e9)))
}

// ToEth converts the big.Int wei to its ether string representation.
func ToEth(wei *big.Int) string {
	z, m := new(big.Int).DivMod(wei, big.NewInt(1e18), new(big.Int))
	if m.Cmp(new(big.Int)) == 0 {
		return z.String()
	}
	s := strings.TrimRight(fmt.Sprintf("%018s", m.String()), "0")
	return z.String() + "." + s
}
