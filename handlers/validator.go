package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/juliangruber/go-intersect"
)

var validatorTemplate = template.Must(template.New("validator").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validator.html"))
var validatorNotFoundTemplate = template.Must(template.New("validatornotfound").ParseFiles("templates/layout.html", "templates/validatornotfound.html"))

// Validator returns validator data using a go template
func Validator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	var index uint64
	var err error

	validatorPageData := types.ValidatorPageData{}

	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "validators",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

	if strings.Contains(vars["index"], "0x") || len(vars["index"]) == 96 {
		pubKey, err := hex.DecodeString(strings.Replace(vars["index"], "0x", "", -1))
		if err != nil {
			logger.Errorf("error parsing validator public key %v: %v", vars["index"], err)
			http.Error(w, "Internal server error", 503)
			return
		}

		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			deposits, err := db.GetValidatorDeposits(pubKey)
			if err != nil {
				logger.Errorf("error getting validator-deposits from db: %v", err)
			}
			if err != nil || len(deposits.Eth1Deposits) == 0 {
				data.Meta.Title = fmt.Sprintf("%v - Validator %x - beaconcha.in - %v", utils.Config.Frontend.SiteName, pubKey, time.Now().Year())
				data.Meta.Path = fmt.Sprintf("/validator/%v", index)
				err := validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)

				if err != nil {
					logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
					http.Error(w, "Internal server error", 503)
					return
				}
				return
			}
			// there is no validator-index but there are eth1-deposits for the publickey
			// which means the validator is in DEPOSITED state
			// in this state there is nothing to display but the eth1-deposits
			validatorPageData.Status = "deposited"
			validatorPageData.PublicKey = []byte(pubKey)
			validatorPageData.Deposits = deposits

			sumValid := uint64(0)
			// check if a valid deposit exists
			for _, d := range deposits.Eth1Deposits {
				if d.ValidSignature {
					sumValid += d.Amount
				} else {
					validatorPageData.Status = "deposited_invalid"
				}
			}

			if sumValid >= 32e9 {
				validatorPageData.Status = "deposited_valid"
			}

			data.Data = validatorPageData
			if utils.IsApiRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				err = json.NewEncoder(w).Encode(data.Data)
			} else {
				err = validatorTemplate.ExecuteTemplate(w, "layout", data)
			}
			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", 503)
				return
			}
			return
		}
	} else {
		index, err = strconv.ParseUint(vars["index"], 10, 64)
		if err != nil {
			logger.Errorf("error parsing validator index: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
	}

	data.Meta.Title = fmt.Sprintf("%v - Validator %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, index, time.Now().Year())
	data.Meta.Path = fmt.Sprintf("/validator/%v", index)

	err = db.DB.Get(&validatorPageData, `
		SELECT 
			validators.validatorindex, 
			validators.withdrawableepoch, 
			validators.effectivebalance, 
			validators.slashed, 
			validators.activationeligibilityepoch, 
			validators.activationepoch, 
			validators.exitepoch, 
			validators.lastattestationslot, 
			COALESCE(validator_balances.balance, 0) AS balance 
		FROM validators 
		LEFT JOIN validator_balances 
			ON validators.validatorindex = validator_balances.validatorindex 
			AND validator_balances.epoch = $1 
		WHERE validators.validatorindex = $2 
		LIMIT 1`, services.LatestEpoch(), index)
	if err != nil {
		logger.Printf("Error retrieving validator page data: %v", err)

		err := validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	validatorPageData.Epoch = services.LatestEpoch()
	validatorPageData.Index = index
	validatorPageData.PublicKey, err = db.GetValidatorPublicKey(index)
	if err != nil {
		logger.Printf("Error retrieving validator public key %v: %v", index, err)

		err := validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	deposits, err := db.GetValidatorDeposits(validatorPageData.PublicKey)
	if err != nil {
		logger.Errorf("error getting validator-deposits from db: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	validatorPageData.Deposits = deposits

	validatorPageData.ActivationEligibilityTs = utils.EpochToTime(validatorPageData.ActivationEligibilityEpoch)
	validatorPageData.ActivationTs = utils.EpochToTime(validatorPageData.ActivationEpoch)
	validatorPageData.ExitTs = utils.EpochToTime(validatorPageData.ExitEpoch)
	validatorPageData.WithdrawableTs = utils.EpochToTime(validatorPageData.WithdrawableEpoch)

	proposals := []struct {
		Slot   uint64
		Status uint64
	}{}

	err = db.DB.Select(&proposals, "SELECT slot, status FROM blocks WHERE proposer = $1 ORDER BY slot", index)
	if err != nil {
		logger.Errorf("error retrieving block-proposals: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	validatorPageData.Proposals = make([][]uint64, len(proposals))
	for i, b := range proposals {
		validatorPageData.Proposals[i] = []uint64{
			uint64(utils.SlotToTime(b.Slot).Unix()),
			b.Status,
		}
	}

	err = db.DB.Get(&validatorPageData.ProposedBlocksCount, "SELECT COUNT(*) FROM blocks WHERE proposer = $1", index)
	if err != nil {
		logger.Errorf("error retrieving proposed blocks count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	err = db.DB.Get(&validatorPageData.AttestationsCount, "SELECT LEAST(COUNT(*), 10000) FROM attestation_assignments WHERE validatorindex = $1", index)
	if err != nil {
		logger.Errorf("error retrieving attestation count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var balanceHistory []*types.ValidatorBalanceHistory
	err = db.DB.Select(&balanceHistory, "SELECT epoch, balance FROM validator_balances WHERE validatorindex = $1 ORDER BY epoch", index)
	if err != nil {
		logger.Errorf("error retrieving validator balance history: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	validatorPageData.BalanceHistoryChartData = make([][]float64, len(balanceHistory))
	for i, balance := range balanceHistory {
		balanceTs := utils.EpochToTime(balance.Epoch)
		validatorPageData.BalanceHistoryChartData[i] = []float64{float64(balanceTs.Unix() * 1000), float64(balance.Balance) / 1000000000}
	}

	earnings, err := GetValidatorEarnings([]uint64{index})
	if err != nil {
		logger.Errorf("error retrieving validator effective balance history: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	validatorPageData.Income1d = earnings.LastDay
	validatorPageData.Income7d = earnings.LastWeek
	validatorPageData.Income31d = earnings.LastMonth

	var effectiveBalanceHistory []*types.ValidatorBalanceHistory
	err = db.DB.Select(&effectiveBalanceHistory, "SELECT epoch, COALESCE(effectivebalance, 0) as balance FROM validator_balances WHERE validatorindex = $1 ORDER BY epoch", index)
	if err != nil {
		logger.Errorf("error retrieving validator effective balance history: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	validatorPageData.EffectiveBalanceHistoryChartData = make([][]float64, len(effectiveBalanceHistory))
	for i, balance := range effectiveBalanceHistory {
		validatorPageData.EffectiveBalanceHistoryChartData[i] = []float64{float64(utils.EpochToTime(balance.Epoch).Unix() * 1000), float64(balance.Balance) / 1000000000}
	}

	validatorOnlineThresholdSlot := GetValidatorOnlineThresholdSlot()

	if validatorPageData.Epoch > validatorPageData.ExitEpoch {
		if validatorPageData.Slashed {
			validatorPageData.Status = "slashed"
		} else {
			validatorPageData.Status = "exited"
		}
	} else if validatorPageData.Epoch < validatorPageData.ActivationEpoch {
		validatorPageData.Status = "pending"
	} else if validatorPageData.Slashed {
		if validatorPageData.ActivationEpoch < services.LatestEpoch() && (validatorPageData.LastAttestationSlot == nil || *validatorPageData.LastAttestationSlot < validatorOnlineThresholdSlot) {
			validatorPageData.Status = "slashing_offline"
		} else {
			validatorPageData.Status = "slashing_online"
		}
	} else {
		if validatorPageData.ActivationEpoch < services.LatestEpoch() && (validatorPageData.LastAttestationSlot == nil || *validatorPageData.LastAttestationSlot < validatorOnlineThresholdSlot) {
			validatorPageData.Status = "active_offline"
		} else {
			validatorPageData.Status = "active_online"
		}
	}

	if validatorPageData.Slashed {
		var slashingInfo struct {
			Slot    uint64
			Slasher uint64
			Reason  string
		}
		err = db.DB.Get(&slashingInfo,
			`select block_slot as slot, proposer as slasher, 'Attestation Violation' as reason
				from blocks_attesterslashings a1 left join blocks b1 on b1.slot = a1.block_slot
				where $1 = ANY(a1.attestation1_indices) and $1 = ANY(a1.attestation2_indices)
			union all
			select block_slot as slot, proposer as slasher, 'Proposer Violation' as reason
				from blocks_proposerslashings a2 left join blocks b2 on b2.slot = a2.block_slot
				where a2.proposerindex = $1
			limit 1`,
			index)
		if err != nil {
			logger.Errorf("error retrieving validator slashing info: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
		validatorPageData.SlashedBy = slashingInfo.Slasher
		validatorPageData.SlashedAt = slashingInfo.Slot
		validatorPageData.SlashedFor = slashingInfo.Reason
	}

	err = db.DB.Get(&validatorPageData.SlashingsCount, `
		select
			(
				select count(*) from blocks_attesterslashings a
				inner join blocks b on b.slot = a.block_slot and b.proposer = $1
				where attestation1_indices is not null and attestation2_indices is not null
			) + (
				select count(*) from blocks_proposerslashings c
				inner join blocks d on d.slot = c.block_slot and d.proposer = $1
			)`, index)
	if err != nil {
		logger.Errorf("error retrieving slashings-count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data.Data = validatorPageData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		err = validatorTemplate.ExecuteTemplate(w, "layout", data)
	}

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorDeposits returns a validator's deposits in json
func ValidatorDeposits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	pubkey, err := hex.DecodeString(strings.Replace(vars["pubkey"], "0x", "", -1))
	if err != nil {
		logger.Errorf("error parsing validator public key %v: %v", vars["pubkey"], err)
		http.Error(w, "Internal server error", 503)
		return
	}

	deposits, err := db.GetValidatorDeposits(pubkey)
	if err != nil {
		logger.Errorf("error getting validator-deposits for %v: %v", vars["pubkey"], err)
		http.Error(w, "Internal server error", 503)
		return
	}

	err = json.NewEncoder(w).Encode(deposits)
	if err != nil {
		logger.Errorf("error encoding validator-deposits for %v: %v", vars["pubkey"], err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorProposedBlocks returns a validator's proposed blocks in json
func ValidatorProposedBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing validator index: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM blocks WHERE proposer = $1", index)
	if err != nil {
		logger.Errorf("error retrieving proposed blocks count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var blocks []*types.IndexPageDataBlocks
	err = db.DB.Select(&blocks, `
		SELECT 
			blocks.epoch, 
			blocks.slot, 
			blocks.proposer, 
			blocks.blockroot, 
			blocks.parentroot, 
			blocks.attestationscount, 
			blocks.depositscount, 
			blocks.voluntaryexitscount, 
			blocks.proposerslashingscount, 
			blocks.attesterslashingscount, 
			blocks.status, 
			blocks.graffiti 
		FROM blocks 
		WHERE blocks.proposer = $1
		ORDER BY blocks.slot DESC
		LIMIT $2 OFFSET $3`, index, length, start)

	if err != nil {
		logger.Errorf("error retrieving proposed blocks data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		tableData[i] = []interface{}{
			utils.FormatEpoch(b.Epoch),
			utils.FormatBlockSlot(b.Slot),
			utils.FormatBlockStatus(b.Status),
			utils.FormatTimestamp(utils.SlotToTime(b.Slot).Unix()),
			utils.FormatBlockRoot(b.BlockRoot),
			b.Attestations,
			b.Deposits,
			fmt.Sprintf("%v / %v", b.Proposerslashings, b.Attesterslashings),
			b.Exits,
			utils.FormatGraffiti(b.Graffiti),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorAttestations returns a validators attestations in json
func ValidatorAttestations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing validator index: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT LEAST(COUNT(*), 10000) FROM attestation_assignments WHERE validatorindex = $1", index)
	if err != nil {
		logger.Errorf("error retrieving proposed blocks count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var blocks []*types.ValidatorAttestation
	err = db.DB.Select(&blocks, `
		SELECT 
			attestation_assignments.epoch, 
			attestation_assignments.attesterslot, 
			attestation_assignments.committeeindex, 
			attestation_assignments.status 
		FROM attestation_assignments 
		WHERE validatorindex = $1
		ORDER BY epoch desc, attesterslot DESC
		LIMIT $2 OFFSET $3`, index, length, start)

	if err != nil {
		logger.Errorf("error retrieving validator attestations data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {

		if utils.SlotToTime(b.AttesterSlot).Before(time.Now().Add(time.Minute*-1)) && b.Status == 0 {
			b.Status = 2
		}
		tableData[i] = []interface{}{
			utils.FormatEpoch(b.Epoch),
			utils.FormatBlockSlot(b.AttesterSlot),
			utils.FormatAttestationStatus(b.Status),
			utils.FormatTimestamp(utils.SlotToTime(b.AttesterSlot).Unix()),
			b.CommitteeIndex,
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorSlashings returns a validators slashings in json
func ValidatorSlashings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing validator index: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var totalCount uint64
	err = db.DB.Get(&totalCount, `
		select
			(
				select count(*) from blocks_attesterslashings a
				inner join blocks b on b.slot = a.block_slot and b.proposer = $1
				where attestation1_indices is not null and attestation2_indices is not null
			) + (
				select count(*) from blocks_proposerslashings c
				inner join blocks d on d.slot = c.block_slot and d.proposer = $1
			)`, index)
	if err != nil {
		logger.Errorf("error retrieving totalCount of validator-slashings: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var attesterSlashings []*types.ValidatorAttestationSlashing
	err = db.DB.Select(&attesterSlashings, `
		SELECT 
			blocks.slot, 
			blocks.epoch, 
			blocks.proposer, 
			blocks_attesterslashings.attestation1_indices, 
			blocks_attesterslashings.attestation2_indices 
		FROM blocks_attesterslashings 
		INNER JOIN blocks ON blocks.proposer = $1 and blocks_attesterslashings.block_slot = blocks.slot 
		WHERE attestation1_indices IS NOT NULL AND attestation2_indices IS NOT NULL`, index)

	if err != nil {
		logger.Errorf("error retrieving validator attestations data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var proposerSlashings []*types.ValidatorProposerSlashing
	err = db.DB.Select(&proposerSlashings, `
		SELECT blocks.slot, blocks.epoch, blocks.proposer, blocks_proposerslashings.proposerindex 
		FROM blocks_proposerslashings 
		INNER JOIN blocks ON blocks.proposer = $1 AND blocks_proposerslashings.block_slot = blocks.slot`, index)
	if err != nil {
		logger.Errorf("error retrieving block proposer slashings data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, 0, len(attesterSlashings)+len(proposerSlashings))
	for _, b := range attesterSlashings {

		inter := intersect.Simple(b.Attestestation1Indices, b.Attestestation2Indices)

		slashedValidator := uint64(0)
		if len(inter) > 0 {
			slashedValidator = uint64(inter[0].(int64))
		}

		tableData = append(tableData, []interface{}{
			utils.FormatSlashedValidator(slashedValidator),
			utils.SlotToTime(b.Slot).Unix(),
			"Attestation Violation",
			utils.FormatBlockSlot(b.Slot),
			utils.FormatEpoch(b.Epoch),
		})
	}

	for _, b := range proposerSlashings {
		tableData = append(tableData, []interface{}{
			utils.FormatSlashedValidator(b.ProposerIndex),
			utils.SlotToTime(b.Slot).Unix(),
			"Proposer Violation",
			utils.FormatBlockSlot(b.Slot),
			utils.FormatEpoch(b.Epoch),
		})
	}

	sort.Slice(tableData, func(i, j int) bool {
		return tableData[i][1].(int64) > tableData[j][1].(int64)
	})

	for _, b := range tableData {
		b[1] = utils.FormatTimestamp(b[1].(int64))
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
