package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math"
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
	slotTemplateFiles := append(layoutTemplateFiles,
		"slot/slot.html",
		"slot/transactions.html",
		"slot/withdrawals.html",
		"slot/attestations.html",
		"slot/deposits.html",
		"slot/votes.html",
		"slot/attesterSlashing.html",
		"slot/proposerSlashing.html",
		"slot/exits.html",
		"slot/blobs.html",
		"components/timestamp.html",
		"slot/overview.html",
		"slot/execTransactions.html")
	slotFutureTemplateFiles := append(layoutTemplateFiles,
		"slot/slotFuture.html",
		"components/timestamp.html")
	blockNotFoundTemplateFiles := append(layoutTemplateFiles, "slotnotfound.html")
	var slotTemplate = templates.GetTemplate(slotTemplateFiles...)
	var slotFutureTemplate = templates.GetTemplate(slotFutureTemplateFiles...)
	var blockNotFoundTemplate = templates.GetTemplate(blockNotFoundTemplateFiles...)

	w.Header().Set("Content-Type", "text/html")

	vars := mux.Vars(r)

	slotOrHash := strings.Replace(vars["slotOrHash"], "0x", "", -1)
	blockSlot := int64(-1)
	blockRootHash, err := hex.DecodeString(slotOrHash)
	if err != nil || len(slotOrHash) != 64 {
		blockRootHash = []byte{}
		blockSlot, err = strconv.ParseInt(vars["slotOrHash"], 10, 64)
		if err != nil || blockSlot > math.MaxInt32 { // block slot must be lower than max int4
			data := InitPageData(w, r, "blockchain", "/slots", fmt.Sprintf("Slot %v", slotOrHash), blockNotFoundTemplateFiles)
			data.Data = "slot"
			if handleTemplateError(w, r, "slot.go", "Slot", "blockSlot", blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
				return // an error has occurred and was processed
			}
			return
		}
	}

	if blockSlot == -1 {
		err = db.ReaderDb.Get(&blockSlot, `SELECT slot FROM blocks WHERE blockroot = $1 OR stateroot = $1 LIMIT 1`, blockRootHash)
		if blockSlot == -1 {
			data := InitPageData(w, r, "blockchain", "/slots", fmt.Sprintf("Slot %v", slotOrHash), blockNotFoundTemplateFiles)
			data.Data = "slot"
			if handleTemplateError(w, r, "slot.go", "Slot", "blockSlot", blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
				return // an error has occurred and was processed
			}
			return
		}
		if err != nil {
			logger.Errorf("error retrieving entry count of given block or state data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	slotPageData, err := GetSlotPageData(uint64(blockSlot))
	if err == sql.ErrNoRows {
		slot := uint64(blockSlot)
		//Slot not in database -> Show future block

		if slot > MaxSlotValue {
			logger.Errorf("error retrieving blockPageData: %v", err)

			data := InitPageData(w, r, "blockchain", "/slots", fmt.Sprintf("Slot %v", slotOrHash), blockNotFoundTemplateFiles)
			if handleTemplateError(w, r, "slot.go", "Slot", "MaxSlotValue", blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
				return // an error has occurred and was processed
			}
		}

		data := InitPageData(w, r, "blockchain", "/slots", fmt.Sprintf("Slot %v", slotOrHash), slotFutureTemplateFiles)
		data.Meta.Path = "/slot/" + slotOrHash
		futurePageData := types.BlockPageData{
			ValidatorProposalInfo: types.ValidatorProposalInfo{
				Slot: slot,
			},
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

	// if the network started with PoS, slot 0 will contain block 0; checking for blockPageData.ExecBlockNumber.Int64 > 0 does not work in this case
	isMergedSlot0 := slotPageData.Slot == 0 && slotPageData.Epoch >= utils.Config.Chain.ClConfig.BellatrixForkEpoch

	if slotPageData.Status == 1 && (slotPageData.ExecBlockNumber.Int64 > 0 || isMergedSlot0) {
		// slot has corresponding execution block, fetch execution data
		eth1BlockPageData, err := GetExecutionBlockPageData(uint64(slotPageData.ExecBlockNumber.Int64), 10)
		// if err != nil, simply show slot view without block
		if err == nil {
			slotPageData.ExecutionData = eth1BlockPageData
			slotPageData.ExecutionData.IsValidMev = slotPageData.IsValidMev
		}
	}
	data := InitPageData(w, r, "blockchain", fmt.Sprintf("/slot/%v", slotPageData.Slot), fmt.Sprintf("Slot %v", slotOrHash), slotTemplateFiles)
	data.Data = slotPageData

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
	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	slotPageData := types.BlockPageData{}
	slotPageData.Mainnet = utils.Config.Chain.ClConfig.ConfigName == "mainnet"
	// for the first slot in an epoch the previous epoch defines the finalized state
	err := db.ReaderDb.Get(&slotPageData, `
		SELECT
			blocks.epoch,
			(COALESCE(epochs.epoch, 0) <= $3) AS epoch_finalized,
			(GREATEST((blocks.slot-1)/$2-1,0) <= $3) AS prev_epoch_finalized,
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
		LEFT JOIN epochs ON GREATEST((blocks.slot-1)/$2,0) = epochs.epoch
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
		blockSlot, utils.Config.Chain.ClConfig.SlotsPerEpoch, latestFinalizedEpoch)
	if err != nil {
		return nil, err
	}
	slotPageData.Slot = uint64(blockSlot)

	slotPageData.Ts = utils.SlotToTime(slotPageData.Slot)
	if slotPageData.ExecTimestamp.Valid {
		slotPageData.ExecTime = time.Unix(int64(slotPageData.ExecTimestamp.Int64), 0)
	}
	slotPageData.SlashingsCount = slotPageData.AttesterSlashingsCount + slotPageData.ProposerSlashingsCount

	slotPageData.NextSlot = slotPageData.Slot + 1
	slotPageData.PreviousSlot = slotPageData.Slot - 1

	slotPageData.Attestations, err = getAttestationsData(slotPageData.Slot, true)
	if err != nil {
		return nil, err
	}

	rows, err := db.ReaderDb.Query(`
		SELECT validators
		FROM blocks_attestations
		WHERE beaconblockroot = $1`,
		slotPageData.BlockRoot)
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
	slotPageData.VotingValidatorsCount = uint64(len(votesPerValidator))
	slotPageData.VotesCount = uint64(votesCount)

	err = db.ReaderDb.Select(&slotPageData.VoluntaryExits, "SELECT validatorindex, signature FROM blocks_voluntaryexits WHERE block_slot = $1", slotPageData.Slot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block deposit data: %v", err)
	}

	err = db.ReaderDb.Select(&slotPageData.AttesterSlashings, `
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
		WHERE block_slot = $1`, slotPageData.Slot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block attester slashings data: %v", err)
	}
	if len(slotPageData.AttesterSlashings) > 0 {
		for _, slashing := range slotPageData.AttesterSlashings {
			inter := intersect.Simple(slashing.Attestation1Indices, slashing.Attestation2Indices)

			for _, i := range inter {
				slashing.SlashedValidators = append(slashing.SlashedValidators, i.(int64))
			}
		}
	}

	err = db.ReaderDb.Select(&slotPageData.BlobSidecars, `SELECT block_slot, block_root, index, kzg_commitment, kzg_proof, blob_versioned_hash FROM blocks_blob_sidecars WHERE block_root = $1`, slotPageData.BlockRoot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block blob sidecars (slot: %d, blockroot: %#x): %w", slotPageData.Slot, slotPageData.BlockRoot, err)
	}

	err = db.ReaderDb.Select(&slotPageData.ProposerSlashings, "SELECT block_slot, block_index, block_root, proposerindex, header1_slot, header1_parentroot, header1_stateroot, header1_bodyroot, header1_signature, header2_slot, header2_parentroot, header2_stateroot, header2_bodyroot, header2_signature FROM blocks_proposerslashings WHERE block_slot = $1", slotPageData.Slot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block proposer slashings data: %v", err)
	}

	err = db.ReaderDb.Select(&slotPageData.SyncCommittee, "SELECT validatorindex FROM sync_committees WHERE period = $1 ORDER BY committeeindex", utils.SyncPeriodOfEpoch(slotPageData.Epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving sync-committee of block %v: %v", slotPageData.Slot, err)
	}

	return &slotPageData, nil
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
			logger.Warnf("error parsing slotOrHash url parameter %v, err: %v", vars["slotOrHash"], err)
			http.Error(w, "Error: Invalid parameter slotOrHash.", http.StatusBadRequest)
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
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}

	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
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
			utils.FormatWithdawalCredentials(deposit.WithdrawalCredentials, true),
			fmt.Sprintf("0x%v", hex.EncodeToString(deposit.Signature)),
			utils.FormatHash(deposit.Signature, true),
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
			logger.Warnf("error parsing slotOrHash url parameter %v, err: %v", vars["slotOrHash"], err)
			http.Error(w, "Error: Invalid parameter slotOrHash.", http.StatusBadRequest)
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
	searchIsInt32 := false
	searchInt32, err := strconv.ParseInt(search, 10, 32)
	if err == nil && searchInt32 >= 0 {
		searchIsInt32 = true
	}

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}

	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
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
	} else if searchIsInt32 {
		err = db.ReaderDb.Get(&count, `SELECT count(*) FROM blocks_attestations WHERE beaconblockroot = $1 AND $2 = ANY(validators)`, blockRootHash, searchInt32)
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
			blockRootHash, searchInt32, length, start)
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
		logger.Warnf("error parsing slot url parameter %v: %v", vars["slot"], err)
		http.Error(w, "Error: Invalid parameter slot.", http.StatusBadRequest)
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
		methodFormatted := `<span class="badge badge-light">Transfer</span>`
		if len(v.Method) > 0 && v.Method != "Transfer" {
			methodFormatted = fmt.Sprintf(`<span class="badge badge-light text-truncate mw-100" truncate-tooltip="%v">%v</span>`, v.Method, v.Method)
		}
		data[i] = &transactionsData{
			HashFormatted: v.HashFormatted,
			Method:        methodFormatted,
			FromFormatted: v.FromFormatted,
			ToFormatted:   v.ToFormatted,
			Value:         utils.FormatAmountFormatted(v.Value, utils.Config.Frontend.ElCurrency, 5, 0, true, true, false),
			Fee:           utils.FormatAmountFormatted(v.Fee, utils.Config.Frontend.ElCurrency, 5, 0, true, true, false),
			GasPrice:      utils.FormatAmountFormatted(v.GasPrice, "GWei", 5, 0, true, true, false),
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
	if err != nil || slot > math.MaxInt32 {
		logger.Warnf("error parsing slot url parameter %v: %v", vars["slot"], err)
		http.Error(w, "Error: Invalid parameter slot.", http.StatusBadRequest)
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
	currency := GetCurrency(r)
	slot, err := strconv.ParseUint(vars["slot"], 10, 64)
	if err != nil || slot > math.MaxInt32 {
		logger.Warnf("error parsing slot url parameter %v: %v", vars["slot"], err)
		http.Error(w, "Error: Invalid parameter slot.", http.StatusBadRequest)
		return
	}
	withdrawals, err := db.GetSlotWithdrawals(slot)
	if err != nil {
		logger.Errorf("error retrieving withdrawals data for slot %v, err: %v", slot, err)
	}

	tableData := make([][]interface{}, 0, len(withdrawals))
	for _, w := range withdrawals {
		tableData = append(tableData, []interface{}{
			template.HTML(fmt.Sprintf("%v", w.Index)),
			utils.FormatValidator(w.ValidatorIndex),
			utils.FormatAddress(w.Address, nil, "", false, false, true),
			utils.FormatClCurrency(w.Amount, currency, 6, true, false, false, true),
		})
	}

	data := &types.DataTableResponse{
		Draw:         1,
		RecordsTotal: uint64(len(withdrawals)),
		Data:         tableData,
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
	if err != nil || slot > math.MaxInt32 {
		logger.Warnf("error parsing slot url parameter %v: %v", vars["slot"], err)
		http.Error(w, "Error: Invalid parameter slot.", http.StatusBadRequest)
		return
	}
	blsChange, err := db.GetSlotBLSChange(slot)
	if err != nil {
		logger.Errorf("error retrieving blsChange data for slot %v, err: %v", slot, err)
	}

	tableData := make([][]interface{}, 0, len(blsChange))
	for _, c := range blsChange {
		tableData = append(tableData, []interface{}{
			utils.FormatValidator(c.Validatorindex),
			utils.FormatHashWithCopy(c.Signature),
			utils.FormatHashWithCopy(c.BlsPubkey),
			utils.FormatAddress(c.Address, nil, "", false, false, true),
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
