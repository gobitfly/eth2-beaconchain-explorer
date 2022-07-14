package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
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

var blockTemplate = template.Must(template.New("block").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/block/block.html",
	"templates/block/transactions.html",
	"templates/block/attestations.html",
	"templates/block/deposits.html",
	"templates/block/votes.html",
	"templates/block/attesterSlashing.html",
	"templates/block/proposerSlashing.html",
	"templates/block/exits.html",
	"templates/block/overview.html",
))
var blockFutureTemplate = template.Must(template.New("blockFuture").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/block/blockFuture.html",
))
var blockNotFoundTemplate = template.Must(template.New("blocknotfound").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/blocknotfound.html"))

// Block will return the data for a block
func Block(w http.ResponseWriter, r *http.Request) {
	const MaxSlotValue = 137438953503 // we only render a page for blocks up to this slot
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

	data := InitPageData(w, r, "blocks", "/blocks", "")

	if blockSlot == -1 {
		err = db.ReaderDb.Get(&blockSlot, `SELECT slot FROM blocks WHERE blockroot = $1 OR stateroot = $1 LIMIT 1`, blockRootHash)
		if blockSlot == -1 {
			err := searchNotFoundTemplate.ExecuteTemplate(w, "layout", data)
			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
			return
		}
		if err != nil {
			logger.Errorf("error retrieving entry count of given block or state data: %v", err)
			http.Error(w, "Internal server error", http.StatusServiceUnavailable)
			return
		}
	}

	if err != nil {
		data.Meta.Title = fmt.Sprintf("%v - Slot %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, slotOrHash, time.Now().Year())
		data.Meta.Path = "/block/" + slotOrHash
		logger.Errorf("error retrieving block data: %v", err)
		err = blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	blockPageData := types.BlockPageData{}
	blockPageData.Mainnet = utils.Config.Chain.Config.ConfigName == "mainnet"
	err = db.ReaderDb.Get(&blockPageData, `
		SELECT
			blocks.epoch,
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
			blocks.voluntaryexitscount,
			blocks.proposer,
			blocks.status,
			exec_parent_hash,
			exec_fee_recipient,
			exec_state_root,
		    exec_receipts_root,
		    exec_logs_bloom,
		    exec_random,
			exec_block_number,
			exec_gas_limit,
			exec_gas_used,
			exec_timestamp,
		    exec_extra_data,
		    exec_base_fee_per_gas,
		    exec_block_hash,
			exec_transactions_count,
			COALESCE(validator_names.name, '') AS name
		FROM blocks 
		LEFT JOIN validators ON blocks.proposer = validators.validatorindex
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		WHERE blocks.slot = $1 ORDER BY blocks.status LIMIT 1`,
		blockSlot)

	blockPageData.Slot = uint64(blockSlot)
	if err != nil {
		//Slot not in database -> Show future block
		data.Meta.Title = fmt.Sprintf("%v - Slot %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, slotOrHash, time.Now().Year())
		data.Meta.Path = "/block/" + slotOrHash

		if blockPageData.Slot > MaxSlotValue {
			logger.Errorf("error retrieving blockPageData: %v", err)
			err = blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)

			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		blockPageData = types.BlockPageData{
			Slot:         blockPageData.Slot,
			Epoch:        utils.EpochOfSlot(blockPageData.Slot),
			Ts:           utils.SlotToTime(blockPageData.Slot),
			NextSlot:     blockPageData.Slot + 1,
			PreviousSlot: blockPageData.Slot - 1,
			Status:       4,
		}
		data.Data = blockPageData

		err = blockFutureTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	data.Meta.Title = fmt.Sprintf("%v - Slot %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, blockPageData.Slot, time.Now().Year())
	data.Meta.Path = fmt.Sprintf("/block/%v", blockPageData.Slot)

	blockPageData.Ts = utils.SlotToTime(blockPageData.Slot)
	if blockPageData.ExecTimestamp.Valid {
		blockPageData.ExecTime = time.Unix(int64(blockPageData.ExecTimestamp.Int64), 0)
	}
	blockPageData.SlashingsCount = blockPageData.AttesterSlashingsCount + blockPageData.ProposerSlashingsCount

	err = db.ReaderDb.Get(&blockPageData.NextSlot, "SELECT slot FROM blocks WHERE slot > $1 ORDER BY slot LIMIT 1", blockPageData.Slot)
	if err == sql.ErrNoRows {
		blockPageData.NextSlot = 0
	} else if err != nil {
		logger.Errorf("error retrieving next slot for block %v: %v", blockPageData.Slot, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = db.ReaderDb.Get(&blockPageData.PreviousSlot, "SELECT slot FROM blocks WHERE slot < $1 ORDER BY slot DESC LIMIT 1", blockPageData.Slot)
	if err != nil {
		logger.Errorf("error retrieving previous slot for block %v: %v", blockPageData.Slot, err)
		blockPageData.PreviousSlot = 0
	}

	var transactions []*types.BlockPageTransaction
	rows, err := db.ReaderDb.Query(`
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
		logger.Errorf("error retrieving block transaction data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
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
			logger.Errorf("error scanning block transaction data: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
		var amount, price big.Int
		amount.SetBytes(tx.Amount)
		price.SetBytes(tx.Price)
		tx.AmountPretty = ToEth(&amount)
		tx.PricePretty = ToGWei(&amount)
		transactions = append(transactions, tx)
	}
	blockPageData.Transactions = transactions

	var attestations []*types.BlockPageAttestation
	rows, err = db.ReaderDb.Query(`
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
		ORDER BY block_index`,
		blockPageData.Slot)
	if err != nil {
		logger.Errorf("error retrieving block attestation data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
			logger.Errorf("error scanning block attestation data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		attestations = append(attestations, attestation)
	}
	blockPageData.Attestations = attestations

	rows, err = db.ReaderDb.Query(`
		SELECT validators
		FROM blocks_attestations
		WHERE beaconblockroot = $1`,
		blockPageData.BlockRoot)
	if err != nil {
		logger.Errorf("error retrieving block votes data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	votesCount := 0
	votesPerValidator := map[int64]int{}
	for rows.Next() {
		validators := pq.Int64Array{}
		err := rows.Scan(&validators)
		if err != nil {
			logger.Errorf("error scanning votes validators data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
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
		logger.Errorf("error retrieving block deposit data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
		logger.Errorf("error retrieving block attester slashings data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(blockPageData.AttesterSlashings) > 0 {
		for _, slashing := range blockPageData.AttesterSlashings {
			inter := intersect.Simple(slashing.Attestation1Indices, slashing.Attestation2Indices)

			for _, i := range inter {
				slashing.SlashedValidators = append(slashing.SlashedValidators, i.(int64))
			}
		}
	}

	err = db.ReaderDb.Select(&blockPageData.ProposerSlashings, "SELECT * FROM blocks_proposerslashings WHERE block_slot = $1", blockPageData.Slot)
	if err != nil {
		logger.Errorf("error retrieving block proposer slashings data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: fix blockPageData data type to include SyncCommittee
	err = db.ReaderDb.Select(&blockPageData.SyncCommittee, "SELECT validatorindex FROM sync_committees WHERE period = $1 ORDER BY committeeindex", utils.SyncPeriodOfEpoch(blockPageData.Epoch))
	if err != nil {
		logger.Errorf("error retrieving sync-committee of block %v: %v", blockPageData.Slot, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data.Data = blockPageData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		w.Header().Set("Content-Type", "text/html")
		err = blockTemplate.ExecuteTemplate(w, "layout", data)
	}

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// BlockDepositData returns the deposits for a specific slot
func BlockDepositData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slotOrHash := strings.Replace(vars["slotOrHash"], "0x", "", -1)
	blockSlot := int64(-1)
	blockRootHash, err := hex.DecodeString(slotOrHash)
	if err != nil || len(slotOrHash) != 64 {
		blockRootHash = []byte{}
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

	search := q.Get("search[value]")
	search = strings.Replace(search, "0x", "", -1)

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

// BlockVoteData returns the votes for a specific slot
func BlockVoteData(w http.ResponseWriter, r *http.Request) {
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
		BlockSlot      uint64        `db:"block_slot"`
		Validators     pq.Int64Array `db:"validators"`
		CommitteeIndex uint64        `db:"committeeindex"`
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
				block_slot,
				validators,
				committeeindex
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
				block_slot,
				validators,
				committeeindex
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
			vote.BlockSlot,
			vote.CommitteeIndex,
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
