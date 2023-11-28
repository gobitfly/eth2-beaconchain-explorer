package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/exporter"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gorillacontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
	"github.com/mssola/user_agent"
	utilMath "github.com/protolambda/zrnt/eth2/util/math"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

// @title beaconcha.in Ethereum API Documentation
// @version 1.1
// @description High performance API for querying information about Ethereum
// @description The API is currently free to use. A fair use policy applies. Calls are rate limited to
// @description 10 requests / 1 minute / IP. All API results are cached for 1 minute.
// @description If you required a higher usage plan please checkout https://beaconcha.in/pricing.
// @description The API key can be provided in the Header or as a query string parameter.
// @description
// @description Key as a query string parameter: `curl https://beaconcha.in/api/v1/slot/1?apikey=<your_key>`
// @description
// @description Key in a request header:  `curl -H 'apikey: <your_key>' https://beaconcha.in/api/v1/slot/1`
// @tag.name Epoch
// @tag.description Consensus layer information about epochs
// @tag.docs.url https://example.com
// @tag.name Slot
// @tag.description Consensus layer information about slots
// @tag.name Validator
// @tag.description Consensus layer information about validators
// @tag.name SyncCommittee
// @tag.name Execution
// @tag.description layer information about addresses, blocks and transactions
// @tag.name ETH.STORE®
// @tag.description is the transparent Ethereum staking reward reference rate.
// @tag.docs.url https://staking.ethermine.org/statistics
// @tag.docs.description More info
// @tag.name Rocketpool
// @tag.description validator statistics
// @tag.docs.url https://rocketpool.net
// @tag.docs.description More info
// @tag.name Misc
// @tag.name User
// @tag.description provided for Oauth applications (public OAuth support is a work in progress).
// @securitydefinitions.oauth2.accessCode OAuthAccessCode
// @tokenurl https://beaconcha.in/user/token
// @authorizationurl https://beaconcha.in/user/authorize
// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// ApiHealthz godoc
// @Summary Health of the explorer
// @Tags Misc
// @Description Health endpoint for monitoring if the explorer is in sync
// @Produce  text/plain
// @Success 200 {object} types.ApiResponse
// @Router /api/healthz [get]
func ApiHealthz(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	modules := []string{
		"monitoring_app",
		"monitoring_el_data",
		"monitoring_services",
		"monitoring_cl_data",
		"monitoring_api",
		"monitoring_redis",
	}

	res := []struct {
		Name   string
		Status string
	}{}
	err := db.WriterDb.Select(&res, "SELECT name, status FROM service_status WHERE name = ANY($1) AND last_update > NOW() - INTERVAL '5 MINUTES' ORDER BY last_update DESC", pq.Array(modules))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No monitoring data available", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	modulesMap := make(map[string]string)
	for _, module := range modules {
		modulesMap[module] = ""
	}

	hasError := false
	response := strings.Builder{}
	for _, status := range res {

		if modulesMap[status.Name] == "" {
			modulesMap[status.Name] = status.Status

			if status.Status != "OK" {
				hasError = true
			}

			response.WriteString(fmt.Sprintf("module %s: %s\n", status.Name, status.Status))
		}
	}

	for _, module := range modules {
		if modulesMap[module] == "" {
			hasError = true
			response.WriteString(fmt.Sprintf("module %s: %s\n", module, "No monitoring data available"))
		}
	}

	if !hasError {
		_, err = fmt.Fprint(w, response.String())

		if err != nil {
			logger.Debugf("error writing status: %v", err)
		}
	} else {
		http.Error(w, response.String(), http.StatusInternalServerError)
		return
	}
}

// ApiHealthzLoadbalancer godoc
// @Summary Health of the explorer-api regarding having a healthy connection to the database
// @Tags Misc
// @Description Health endpoint for montitoring if the explorer-api
// @Produce  text/plain
// @Success 200 {object} types.ApiResponse
// @Router /api/healthz-loadbalancer [get]
func ApiHealthzLoadbalancer(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	lastEpoch, err := db.GetLatestEpoch()

	if err != nil {
		http.Error(w, "Internal server error: could not retrieve latest epoch from the db", http.StatusInternalServerError)
		return
	}

	if utils.Config.Chain.GenesisTimestamp == 18446744073709551615 {
		fmt.Fprint(w, "OK. No GENESIS_TIMESTAMP defined yet")
		return
	}

	genesisTime := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)
	if genesisTime.After(time.Now()) {
		fmt.Fprintf(w, "OK. Genesis in %v (%v)", time.Until(genesisTime), genesisTime)
		return
	}

	fmt.Fprintf(w, "OK. Last epoch is from %v ago", time.Since(utils.EpochToTime(lastEpoch)))
}

// ApiEthStoreDay godoc
// @Summary Get ETH.STORE® reference rate for a specified beaconchain-day or the latest day
// @Tags ETH.STORE®
// @Description ETH.STORE® represents the average financial return validators on the Ethereum network have achieved in a 24-hour period.
// @Description For each 24-hour period the datapoint is denoted by the number of days that have passed since genesis for that period (= beaconchain-day)
// @Description See https://github.com/gobitfly/eth.store for further information.
// @Produce json
// @Param day path string true "The beaconchain-day (periods of <(24 * 60 * 60) // SlotsPerEpoch // SecondsPerSlot> epochs) to get the the ETH.STORE® for. Must be a number or the string 'latest'."
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/ethstore/{day} [get]
func ApiEthStoreDay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var rows *sql.Rows
	query := `
		SELECT 
			day, 
			effective_balances_sum_wei, 
			start_balances_sum_wei, 
			end_balances_sum_wei, 
			deposits_sum_wei, 
			tx_fees_sum_wei, 
			consensus_rewards_sum_wei,
			total_rewards_wei,
			apr,
			CAST (ROUND((365 * consensus_rewards_sum_wei) / effective_balances_sum_wei, 16) AS double precision) as cl_apr,
			CAST (ROUND((365 * tx_fees_sum_wei) / effective_balances_sum_wei, 16) AS double precision) as el_apr,
			(select avg(apr) from eth_store_stats as e1 where e1.validator = -1 AND e1.day > e.day - 7 AND e1.day <= e.day) as avgapr7d,
			(select avg(CAST (ROUND((365 * consensus_rewards_sum_wei) / effective_balances_sum_wei, 16) AS double precision)) from eth_store_stats as e1 where e1.validator = -1 AND e1.day > e.day - 7 AND e1.day <= e.day) as cl_avgapr7d,
			(select avg(CAST (ROUND((365 * tx_fees_sum_wei) / effective_balances_sum_wei, 16) AS double precision)) from eth_store_stats as e1 where e1.validator = -1 AND e1.day > e.day - 7 AND e1.day <= e.day) as el_avgapr7d,
			(select avg(consensus_rewards_sum_wei) from eth_store_stats as e1 where e1.validator = -1 AND e1.day > e.day - 7 AND e1.day <= e.day) as avgconsensus_rewards7d_wei,
			(select avg(tx_fees_sum_wei) from eth_store_stats as e1 where e1.validator = -1 AND e1.day > e.day - 7 AND e1.day <= e.day) as avgtx_fees7d_wei,
			(select avg(apr) from eth_store_stats as e2 where e2.validator = -1 AND e2.day > e.day - 31 AND e2.day <= e.day) as avgapr31d,
			(select avg(CAST (ROUND((365 * consensus_rewards_sum_wei) / effective_balances_sum_wei, 16) AS double precision)) from eth_store_stats as e2 where e2.validator = -1 AND e2.day > e.day - 31 AND e2.day <= e.day) as cl_avgapr31d,
			(select avg(CAST (ROUND((365 * tx_fees_sum_wei) / effective_balances_sum_wei, 16) AS double precision)) from eth_store_stats as e2 where e2.validator = -1 AND e2.day > e.day - 31 AND e2.day <= e.day) as el_avgapr31d,
			(select avg(consensus_rewards_sum_wei) from eth_store_stats as e2 where e2.validator = -1 AND e2.day > e.day - 31 AND e2.day <= e.day) as avgconsensus_rewards31d_wei,
			(select avg(tx_fees_sum_wei) from eth_store_stats as e2 where e2.validator = -1 AND e2.day > e.day - 31 AND e2.day <= e.day) as avgtx_fees31d_wei
		FROM eth_store_stats e
		WHERE validator = -1 `

	vars := mux.Vars(r)
	if vars["day"] == "latest" {
		rows, err = db.ReaderDb.Query(query + ` ORDER BY day DESC LIMIT 1;`)
	} else {
		day, e := strconv.ParseInt(vars["day"], 10, 64)
		if e != nil {
			SendBadRequestResponse(w, r.URL.String(), "invalid day provided")
			return
		}
		rows, err = db.ReaderDb.Query(query+` AND day = $1;`, day)
	}

	if err != nil {
		logger.Errorf("error retrieving eth.store data: %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	addDayTime := func(dataEntryMap map[string]interface{}) error {
		day, ok := dataEntryMap["day"].(int64)
		if !ok {
			return fmt.Errorf("error type asserting day as an int")
		} else {
			dataEntryMap["day_start"] = utils.DayToTime(day)
			dataEntryMap["day_end"] = utils.DayToTime(day + 1)
		}
		return nil
	}

	returnQueryResults(rows, w, r, addDayTime)
}

// ApiEpoch godoc
// @Summary Get epoch by number, latest, finalized
// @Tags Epoch
// @Description Returns information for a specified epoch by the epoch number or an epoch tag (can be latest or finalized)
// @Produce  json
// @Param  epoch path string true "Epoch number, the string latest or the string finalized"
// @Success 200 {object} types.ApiResponse{data=types.APIEpochResponse} "Success"
// @Failure 400 {object} types.ApiResponse "Failure"
// @Failure 500 {object} types.ApiResponse "Server Error"
// @Router /api/v1/epoch/{epoch} [get]
func ApiEpoch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	epoch, err := strconv.ParseInt(vars["epoch"], 10, 64)
	if err != nil && vars["epoch"] != "latest" && vars["epoch"] != "finalized" {
		SendBadRequestResponse(w, r.URL.String(), "invalid epoch provided")
		return
	}

	if vars["epoch"] == "latest" {
		epoch = int64(services.LatestEpoch())
	}

	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	if vars["epoch"] == "finalized" {
		epoch = int64(services.LatestFinalizedEpoch())
	}

	if epoch > int64(services.LatestEpoch()) {
		SendBadRequestResponse(w, r.URL.String(), fmt.Sprintf("epoch is in the future. The latest epoch is %v", services.LatestEpoch()))
		return
	}

	if epoch < 0 {
		SendBadRequestResponse(w, r.URL.String(), "epoch must be a positive number")
		return
	}

	rows, err := db.ReaderDb.Query(`SELECT attestationscount, attesterslashingscount, averagevalidatorbalance, blockscount, depositscount, eligibleether, epoch, (epoch <= $2) AS finalized, globalparticipationrate, proposerslashingscount, rewards_exported, totalvalidatorbalance, validatorscount, voluntaryexitscount, votedether, withdrawalcount, 
		(SELECT COUNT(*) FROM blocks WHERE epoch = $1 AND status = '0') as scheduledblocks,
		(SELECT COUNT(*) FROM blocks WHERE epoch = $1 AND status = '1') as proposedblocks,
		(SELECT COUNT(*) FROM blocks WHERE epoch = $1 AND status = '2') as missedblocks,
		(SELECT COUNT(*) FROM blocks WHERE epoch = $1 AND status = '3') as orphanedblocks
		FROM epochs WHERE epoch = $1`, epoch, latestFinalizedEpoch)
	if err != nil {
		logger.WithError(err).Error("error retrieving epoch data")
		sendServerErrorResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	addEpochTime := func(dataEntryMap map[string]interface{}) error {
		dataEntryMap["ts"] = utils.EpochToTime(uint64(epoch))
		return nil
	}

	returnQueryResults(rows, w, r, addEpochTime)
}

// ApiEpochSlots godoc
// @Summary Get epoch blocks by epoch number, latest or finalized
// @Tags Epoch
// @Description Returns all slots for a specified epoch
// @Produce  json
// @Param  epoch path string true "Epoch number, the string latest or string finalized"
// @Success 200 {object} types.ApiResponse{data=[]types.APISlotResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/epoch/{epoch}/slots [get]
func ApiEpochSlots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	epoch, err := strconv.ParseInt(vars["epoch"], 10, 64)
	if err != nil && vars["epoch"] != "latest" && vars["epoch"] != "finalized" {
		SendBadRequestResponse(w, r.URL.String(), "invalid epoch provided")
		return
	}

	if vars["epoch"] == "latest" {
		epoch = int64(services.LatestEpoch())
	}

	if vars["epoch"] == "finalized" {
		epoch = int64(services.LatestFinalizedEpoch())
	}

	if epoch > int64(services.LatestEpoch()) {
		SendBadRequestResponse(w, r.URL.String(), fmt.Sprintf("epoch is in the future. The latest epoch is %v", services.LatestEpoch()))
		return
	}

	if epoch < 0 {
		SendBadRequestResponse(w, r.URL.String(), "epoch must be a positive number")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT attestationscount, attesterslashingscount, blockroot, depositscount, epoch, eth1data_blockhash, eth1data_depositcount, eth1data_depositroot, exec_base_fee_per_gas, exec_block_hash, exec_block_number, exec_extra_data, exec_fee_recipient, exec_gas_limit, exec_gas_used, exec_logs_bloom, exec_parent_hash, exec_random, exec_receipts_root, exec_state_root, exec_timestamp, exec_transactions_count, graffiti, graffiti_text, parentroot, proposer, proposerslashingscount, randaoreveal, signature, slot, stateroot, status, syncaggregate_bits, syncaggregate_participation, syncaggregate_signature, voluntaryexitscount, withdrawalcount FROM blocks WHERE epoch = $1 ORDER BY slot", epoch)
	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiSlots godoc
// @Summary Get a slot by its slot number or root hash. Alternatively get the latest slot or the slot containing the head block.
// @Tags Slot
// @Description Returns a slot by its slot number or root hash, the latest slot with string latest or the slot containing the head block with string head
// @Produce  json
// @Param  slotOrHash path string true "Slot or root hash or the string latest or head"
// @Success 200 {object} types.ApiResponse{data=types.APISlotResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/slot/{slotOrHash} [get]
func ApiSlots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slotOrHash := strings.Replace(vars["slotOrHash"], "0x", "", -1)

	blockSlot := int64(-1)
	blockRootHash := []byte{}

	if slotOrHash == "latest" {
		// simply check the latest slot (might be empty which causes an error)
		blockSlot = int64(services.LatestSlot())
	} else if slotOrHash == "head" {
		// retrieve the slot containing the head block of the chain
		blockRootHash = services.Eth1HeadBlockRootHash()
		if len(blockRootHash) != 32 {
			SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
			return
		}
	} else {
		var err error
		blockRootHash, err = hex.DecodeString(slotOrHash)
		if err != nil || len(slotOrHash) != 64 {
			// not a valid root hash, try to parse as slot number instead
			blockRootHash = []byte{}
			blockSlot, err = strconv.ParseInt(vars["slotOrHash"], 10, 64)
			if err != nil {
				SendBadRequestResponse(w, r.URL.String(), "could not parse slot number")
				return
			}
		}
	}

	if len(blockRootHash) != 32 {
		err := db.ReaderDb.Get(&blockRootHash, `SELECT blockroot FROM blocks WHERE slot = $1`, blockSlot)

		if err != nil || len(blockRootHash) != 32 {
			SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
			return
		}
	}

	rows, err := db.ReaderDb.Query(`
	SELECT
		blocks.epoch,
		blocks.slot,
		blocks.blockroot,
		blocks.parentroot,
		blocks.stateroot,
		blocks.signature,
		blocks.randaoreveal,
		blocks.graffiti,
		blocks.graffiti_text,
		blocks.eth1data_depositroot,
		blocks.eth1data_depositcount,
		blocks.eth1data_blockhash,
		blocks.proposerslashingscount,
		blocks.attesterslashingscount,
		blocks.attestationscount,
		blocks.depositscount,
		blocks.withdrawalcount, 
		blocks.voluntaryexitscount,
		blocks.proposer,
		blocks.status,
		blocks.syncaggregate_bits,
		blocks.syncaggregate_signature,
		blocks.syncaggregate_participation,
		blocks.exec_parent_hash,
		blocks.exec_fee_recipient,
		blocks.exec_state_root,
		blocks.exec_receipts_root,
		blocks.exec_logs_bloom,
		blocks.exec_random,
		blocks.exec_block_number,
		blocks.exec_gas_limit,
		blocks.exec_gas_used,
		blocks.exec_timestamp,
		blocks.exec_extra_data,
		blocks.exec_base_fee_per_gas,
		blocks.exec_block_hash,     
		blocks.exec_transactions_count,
		ba.votes
	FROM
		blocks
	LEFT JOIN
		(SELECT beaconblockroot, sum(array_length(validators, 1)) AS votes FROM blocks_attestations GROUP BY beaconblockroot) ba ON (blocks.blockroot = ba.beaconblockroot)
	WHERE
		blocks.blockroot = $1`, blockRootHash)

	if err != nil {
		logger.WithError(err).Error("could not retrieve db results")
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, w, r)
}

// ApiSlotAttestations godoc
// @Summary Get the attestations included in a specific slot
// @Tags Slot
// @Description Returns the attestations included in a specific slot
// @Produce  json
// @Param  slot path string true "Slot"
// @Success 200 {object} types.ApiResponse{data=[]types.APIAttestationResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/slot/{slot}/attestations [get]
func ApiSlotAttestations(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil && vars["slot"] != "latest" {
		SendBadRequestResponse(w, r.URL.String(), "invalid block slot provided")
		return
	}

	if vars["slot"] == "latest" {
		slot = int64(services.LatestSlot())
	}

	if slot > int64(services.LatestSlot()) {
		SendBadRequestResponse(w, r.URL.String(), fmt.Sprintf("slot is in the future. The latest slot is %v", services.LatestSlot()))
		return
	}

	if slot < 0 {
		SendBadRequestResponse(w, r.URL.String(), "slot must be a positive number")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT aggregationbits, beaconblockroot, block_index, block_root, block_slot, committeeindex, signature, slot, source_epoch, source_root, target_epoch, target_root, validators FROM blocks_attestations WHERE block_slot = $1 ORDER BY block_index", slot)
	if err != nil {
		logger.WithError(err).Error("could not retrieve db results")
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiSlotAttesterSlashings godoc
// @Summary Get the attester slashings included in a specific slot
// @Tags Slot
// @Description Returns the attester slashings included in a specific slot
// @Produce  json
// @Param  slot path string true "Slot"
// @Success 200 {object} types.ApiResponse{data=[]types.APIAttesterSlashingResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/slot/{slot}/attesterslashings [get]
func ApiSlotAttesterSlashings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT attestation1_beaconblockroot, attestation1_index, attestation1_indices, attestation1_signature, attestation1_slot, attestation1_source_epoch, attestation1_source_root, attestation1_target_epoch, attestation1_target_root, attestation2_beaconblockroot, attestation2_index, attestation2_indices, attestation2_signature, attestation2_slot, attestation2_source_epoch, attestation2_source_root, attestation2_target_epoch, attestation2_target_root, block_index, block_root, block_slot FROM blocks_attesterslashings WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiSlotDeposits godoc
// @Summary Get the deposits included in a specific block
// @Tags Slot
// @Description Returns the deposits included in a specific block
// @Produce  json
// @Param  slot path string true "Block slot"
// @Param  limit query string false "Limit the number of results"
// @Param offset query string false "Offset the number of results"
// @Success 200 {object} types.ApiResponse{[]APIAttestationResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/slot/{slot}/deposits [get]
func ApiSlotDeposits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	q := r.URL.Query()

	limitQuery := q.Get("limit")
	offsetQuery := q.Get("offset")

	offset, err := strconv.ParseInt(offsetQuery, 10, 64)
	if err != nil {
		offset = 0
	}

	limit, err := strconv.ParseInt(limitQuery, 10, 64)
	if err != nil {
		limit = 100 + offset
	}

	if offset < 0 {
		offset = 0
	}

	if limit > (100+offset) || limit <= 0 || limit <= offset {
		limit = 100 + offset
	}

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT amount, block_index, block_root, block_slot, proof, publickey, signature, withdrawalcredentials FROM blocks_deposits WHERE block_slot = $1 ORDER BY block_index DESC limit $2 offset $3", slot, limit, offset)
	if err != nil {
		logger.WithError(err).Error("could not retrieve db results")
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiSlotProposerSlashings godoc
// @Summary Get the proposer slashings included in a specific slot
// @Tags Slot
// @Description Returns the proposer slashings included in a specific slot
// @Produce  json
// @Param  slot path string true "Slot"
// @Success 200 {object} types.ApiResponse{data=[]types.APIProposerSlashingResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/slot/{slot}/proposerslashings [get]
func ApiSlotProposerSlashings(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT block_index, block_root, block_slot, header1_bodyroot, header1_parentroot, header1_signature, header1_slot, header1_stateroot, header2_bodyroot, header2_parentroot, header2_signature, header2_slot, header2_stateroot, proposerindex FROM blocks_proposerslashings WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		logger.WithError(err).Error("could not retrieve db results")
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiSlotVoluntaryExits godoc
// @Summary Get the voluntary exits included in a specific slot
// @Tags Slot
// @Description Returns the voluntary exits included in a specific slot
// @Produce  json
// @Param  slot path string true "Slot"
// @Success 200 {object} types.ApiResponse{data=[]types.APIVoluntaryExitResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/slot/{slot}/voluntaryexits [get]
func ApiSlotVoluntaryExits(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT block_slot, block_index, block_root, epoch, validatorindex, signature FROM blocks_voluntaryexits WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		logger.WithError(err).Error("could not retrieve db results")
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiSlotWithdrawals godoc
// @Summary Get the withdrawals included in a specific slot
// @Tags Slot
// @Description Returns the withdrawals included in a specific slot
// @Produce json
// @Param slot path string true "Block slot"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/slot/{slot}/withdrawals [get]
func ApiSlotWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT block_slot, withdrawalindex, validatorindex, address, amount FROM blocks_withdrawals WHERE block_slot = $1 ORDER BY withdrawalindex", slot)
	if err != nil {
		logger.WithError(err).Error("error getting blocks_withdrawals")
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()
	returnQueryResults(rows, w, r)
}

// ApiBlockVoluntaryExits godoc
// ApiSyncCommittee godoc
// @Summary Get the sync-committee for a sync-period
// @Tags SyncCommittee
// @Description Returns the sync-committee for a sync-period. Validators are sorted by sync-committee-index.
// @Description Sync committees where introduced in the Altair hardfork. Peroids before the hardfork do not contain sync-committees.
// @Description For mainnet sync-committes first started after epoch 74240 (period 290) and each sync-committee is active for 256 epochs.
// @Produce json
// @Param period path string true "Period ('latest' for latest period or 'next' for next period in the future)"
// @Success 200 {object} types.ApiResponse{data=types.APISyncCommitteeResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/sync_committee/{period} [get]
func ApiSyncCommittee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	period, err := strconv.ParseUint(vars["period"], 10, 64)
	if err != nil && vars["period"] != "latest" && vars["period"] != "next" {
		SendBadRequestResponse(w, r.URL.String(), "invalid epoch provided")
		return
	}

	if vars["period"] == "latest" {
		period = utils.SyncPeriodOfEpoch(services.LatestEpoch())
	} else if vars["period"] == "next" {
		period = utils.SyncPeriodOfEpoch(services.LatestEpoch()) + 1
	}

	// Beware that we do not deduplicate here since a validator can be part multiple times of the same sync committee period
	// and the order of the committeeindex is important, deduplicating it would mess up the order
	rows, err := db.ReaderDb.Query(`SELECT period, GREATEST(period*$2, $3) AS start_epoch, ((period+1)*$2)-1 AS end_epoch, ARRAY_AGG(validatorindex ORDER BY committeeindex) AS validators FROM sync_committees WHERE period = $1 GROUP BY period`, period, utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod, utils.Config.Chain.ClConfig.AltairForkEpoch)
	if err != nil {
		logger.WithError(err).WithField("url", r.URL.String()).Errorf("error querying db")
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, w, r)
}

// ApiValidatorQueue godoc
// @Summary Get the current validator queue
// @Tags Validator
// @Description Returns the current number of validators entering and exiting the beacon chain
// @Produce  json
// @Success 200 {object} types.ApiResponse{data=types.ApiValidatorQueueResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validators/queue [get]
func ApiValidatorQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.ReaderDb.Query("SELECT e.validatorscount, q.entering_validators_count as beaconchain_entering, q.exiting_validators_count as beaconchain_exiting FROM epochs e, queue q ORDER BY e.epoch DESC, q.ts DESC LIMIT 1")
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, w, r)
}

// ApiRocketpoolStats godoc
// @Summary Get global rocketpool network statistics
// @Tags Rocketpool
// @Produce  json
// @Success 200 {object} types.ApiResponse{data=types.APIRocketpoolStatsResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/rocketpool/stats [get]
func ApiRocketpoolStats(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)

	stats, err := getRocketpoolStats()

	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not parse db results")
		return
	}

	SendOKResponse(j, r.URL.String(), stats)
}

// ApiRocketpoolValidators godoc
// @Summary Get rocketpool specific data for given validators
// @Tags Rocketpool
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Produce  json
// @Success 200 {object} types.ApiResponse{data=types.ApiRocketpoolValidatorResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/rocketpool/validator/{indexOrPubkey} [get]
func ApiRocketpoolValidators(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	stats, err := getRocketpoolValidators(queryIndices)

	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not parse db results")
		return
	}

	SendOKResponse(j, r.URL.String(), stats)
}

/*
Combined validator get, performance, attestation efficency, sync committee statistics, epoch, historic epoch and rpl
Not public documented
*/
func ApiDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.LogError(err, "reading body", 0)
		SendBadRequestResponse(w, r.URL.String(), "could not read body")
		return
	}

	var getValidators bool = true
	var parsedBody types.DashboardRequest
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		utils.LogError(err, "unmarshal json body error", 0)
		getValidators = false
	}

	maxValidators := getUserPremium(r).MaxValidators

	epoch := services.LatestEpoch()

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(5) // limit concurrency
	var validatorsData []interface{}
	var validatorEffectivenessData []*types.ValidatorEffectiveness
	var rocketpoolData []interface{}
	var rocketpoolStats []interface{}
	var currentEpochData []interface{}
	var executionPerformance []types.ExecutionPerformanceResponse
	var olderEpochData []interface{}
	var currentSyncCommittee []interface{}
	var nextSyncCommittee []interface{}
	var syncCommitteeStats *SyncCommitteesInfo
	var proposalLuckStats *types.ApiProposalLuckResponse

	if getValidators {
		queryIndices, err := parseApiValidatorParamToIndices(parsedBody.IndicesOrPubKey, maxValidators)
		if err != nil {
			SendBadRequestResponse(w, r.URL.String(), err.Error())
			return
		}

		if len(queryIndices) > 0 {
			g.Go(func() error {
				start := time.Now()
				var err error
				validatorsData, err = getGeneralValidatorInfoForAppDashboard(queryIndices)
				elapsed := time.Since(start)
				if elapsed > 10*time.Second {
					logger.Warnf("getGeneralValidatorInfoForAppDashboard(%v) took longer than 10 sec", queryIndices)
				}
				return err
			})

			g.Go(func() error {
				start := time.Now()
				var err error
				validatorEffectivenessData, err = getValidatorEffectiveness(epoch-1, queryIndices)
				elapsed := time.Since(start)
				if elapsed > 10*time.Second {
					logger.Warnf("getValidatorEffectiveness(%v, %v) took longer than 10 sec", epoch-1, queryIndices)
				}
				return err
			})

			g.Go(func() error {
				start := time.Now()
				var err error
				rocketpoolData, err = getRocketpoolValidators(queryIndices)
				elapsed := time.Since(start)
				if elapsed > 10*time.Second {
					logger.Warnf("getRocketpoolValidators(%v) took longer than 10 sec", queryIndices)
				}
				return err
			})

			g.Go(func() error {
				start := time.Now()
				var err error
				executionPerformance, err = getValidatorExecutionPerformance(queryIndices)
				elapsed := time.Since(start)
				if elapsed > 10*time.Second {
					logger.Warnf("getValidatorExecutionPerformance(%v) took longer than 10 sec", queryIndices)
				}
				return err
			})

			g.Go(func() error {
				start := time.Now()
				var err error
				period := utils.SyncPeriodOfEpoch(epoch)
				currentSyncCommittee, err = getSyncCommitteeInfoForValidators(queryIndices, period)
				elapsed := time.Since(start)
				if elapsed > 10*time.Second {
					logger.Warnf("getSyncCommitteeInfoForValidators(%v, %v) took longer than 10 sec", queryIndices, period)
				}
				return err
			})

			g.Go(func() error {
				start := time.Now()
				var err error
				period := utils.SyncPeriodOfEpoch(epoch) + 1
				nextSyncCommittee, err = getSyncCommitteeInfoForValidators(queryIndices, period)
				elapsed := time.Since(start)
				if elapsed > 10*time.Second {
					logger.Warnf("SyncPeriodOfEpoch(%v) + 1 took longer than 10 sec", epoch)
					logger.Warnf("getSyncCommitteeInfoForValidators(%v, %v) took longer than 10 sec", queryIndices, period)
				}
				return err
			})

			g.Go(func() error {
				start := time.Now()
				var err error
				syncCommitteeStats, err = getSyncCommitteeStatistics(queryIndices, epoch)
				elapsed := time.Since(start)
				if elapsed > 10*time.Second {
					logger.Warnf("getSyncCommitteeStatistics(%v, %v) took longer than 10 sec", queryIndices, epoch)
				}
				return err
			})

			g.Go(func() error {
				start := time.Now()
				var err error
				proposalLuckStats, err = getProposalLuckStats(queryIndices)
				elapsed := time.Since(start)
				if elapsed > 10*time.Second {
					logger.Warnf("getProposalLuck(%v, %v) took longer than 10 sec", queryIndices, epoch)
				}
				return err
			})
		}
	}

	g.Go(func() error {
		start := time.Now()
		var err error
		currentEpochData, err = getEpoch(int64(epoch) - 1)
		elapsed := time.Since(start)
		if elapsed > 10*time.Second {
			logger.Warnf("getEpoch(%v) took longer than 10 sec", int64(epoch)-1)
		}
		return err
	})

	g.Go(func() error {
		start := time.Now()
		var err error
		olderEpochData, err = getEpoch(int64(epoch) - 10)
		elapsed := time.Since(start)
		if elapsed > 10*time.Second {
			logger.Warnf("getEpoch(%v) took longer than 10 sec", int64(epoch)-10)
		}
		return err
	})

	g.Go(func() error {
		start := time.Now()
		var err error
		rocketpoolStats, err = getRocketpoolStats()
		elapsed := time.Since(start)
		if elapsed > 10*time.Second {
			logger.Warnf("getRocketpoolStats() took longer than 10 sec")
		}
		return err
	})

	err = g.Wait()
	if err != nil {
		utils.LogError(err, "dashboard", 0)
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	data := &DashboardResponse{
		Validators:           validatorsData,
		Effectiveness:        validatorEffectivenessData,
		CurrentEpoch:         currentEpochData,
		OlderEpoch:           olderEpochData,
		Rocketpool:           rocketpoolData,
		RocketpoolStats:      rocketpoolStats,
		ExecutionPerformance: executionPerformance,
		CurrentSyncCommittee: currentSyncCommittee,
		NextSyncCommittee:    nextSyncCommittee,
		SyncCommitteesStats:  syncCommitteeStats,
		ProposalLuckStats:    proposalLuckStats,
	}

	SendOKResponse(j, r.URL.String(), []interface{}{data})
}

func getSyncCommitteeInfoForValidators(validators []uint64, period uint64) ([]interface{}, error) {
	rows, err := db.ReaderDb.Query(`
			WITH
				data as (
					SELECT 
						period,
						validatorindex,
						max(committeeindex) as committeeindex
					FROM sync_committees 
					WHERE period = $1 AND validatorindex = ANY($2)
					group by period, validatorindex
				)	
			SELECT 
				period, 
				GREATEST(period*$3, $4) AS start_epoch, 
				((period+1)*$3)-1 AS end_epoch, 
				ARRAY_AGG(validatorindex ORDER BY committeeindex) AS validators 
			FROM data 	
			group by period;`,
		period, pq.Array(validators),
		utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod, utils.Config.Chain.ClConfig.AltairForkEpoch,
	)
	if err != nil {
		return nil, fmt.Errorf("could not get sync committee for period %d: %w", period, err)
	}
	defer rows.Close()
	return utils.SqlRowsToJSON(rows)
}

func getSyncCommitteeStatistics(validators []uint64, epoch uint64) (*SyncCommitteesInfo, error) {
	if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
		// no sync committee duties before altair fork
		return &SyncCommitteesInfo{}, nil
	}

	if len(validators) == 0 {
		// no validators mean no sync committee duties either
		return &SyncCommitteesInfo{}, nil
	}

	expectedSlots, err := getExpectedSyncCommitteeSlots(validators, epoch)
	if err != nil {
		return nil, err
	}

	stats, err := getSyncCommitteeSlotsStatistics(validators, epoch)
	if err != nil {
		return nil, err
	}

	return &SyncCommitteesInfo{SyncCommitteesStats: stats, ExpectedSlots: expectedSlots}, nil
}

func getExpectedSyncCommitteeSlots(validators []uint64, epoch uint64) (expectedSlots uint64, err error) {
	if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
		// no sync committee duties before altair fork
		return 0, nil
	}

	lastFinalizedEpoch := services.LatestFinalizedEpoch()
	if epoch > lastFinalizedEpoch {
		epoch = lastFinalizedEpoch
	}

	// retrieve activation and exit epochs from database per validator
	type ValidatorInfo struct {
		Id                         int64  `db:"validatorindex"`
		ActivationEpoch            uint64 `db:"activationepoch"`
		ExitEpoch                  uint64 `db:"exitepoch"`
		FirstPossibleSyncCommittee uint64 // calculated
		LastPossibleSyncCommittee  uint64 // calculated
	}

	var validatorsInfoFromDb = []ValidatorInfo{}
	query, args, err := sqlx.In(`SELECT validatorindex, activationepoch, exitepoch FROM validators WHERE validatorindex IN (?) ORDER BY validatorindex ASC`, validators)
	if err != nil {
		return 0, err
	}

	err = db.ReaderDb.Select(&validatorsInfoFromDb, db.ReaderDb.Rebind(query), args...)
	if err != nil {
		return 0, err
	}

	// only check validators that are/have been active and that did not exit before altair
	const noEpoch = uint64(9223372036854775807)
	var validatorsInfo = make([]ValidatorInfo, 0, len(validatorsInfoFromDb))
	for _, v := range validatorsInfoFromDb {
		if v.ActivationEpoch != noEpoch && v.ActivationEpoch < epoch && (v.ExitEpoch == noEpoch || v.ExitEpoch >= utils.Config.Chain.ClConfig.AltairForkEpoch) {
			validatorsInfo = append(validatorsInfo, v)
		}
	}

	if len(validatorsInfo) == 0 {
		// no validators relevant for sync duties
		return 0, nil
	}

	// we need all related and unique timeframes (activation and exit sync period) for all validators
	uniquePeriods := make(map[uint64]bool)
	for i := range validatorsInfo {
		// first epoch (activation epoch or Altair if Altair was later as there were no sync committees pre Altair)
		firstSyncEpoch := validatorsInfo[i].ActivationEpoch
		if validatorsInfo[i].ActivationEpoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
			firstSyncEpoch = utils.Config.Chain.ClConfig.AltairForkEpoch
		}
		validatorsInfo[i].FirstPossibleSyncCommittee = utils.SyncPeriodOfEpoch(firstSyncEpoch)
		uniquePeriods[validatorsInfo[i].FirstPossibleSyncCommittee] = true

		// last epoch (exit epoch or current epoch if not exited yet)
		lastSyncEpoch := epoch
		if validatorsInfo[i].ExitEpoch != noEpoch && validatorsInfo[i].ExitEpoch <= epoch {
			lastSyncEpoch = validatorsInfo[i].ExitEpoch
		}
		validatorsInfo[i].LastPossibleSyncCommittee = utils.SyncPeriodOfEpoch(lastSyncEpoch)
		uniquePeriods[validatorsInfo[i].LastPossibleSyncCommittee] = true
	}

	// transform map to slice; this will be used to query sync_committees_count_per_validator
	periodSlice := make([]uint64, 0, len(uniquePeriods))
	for period := range uniquePeriods {
		periodSlice = append(periodSlice, period)
	}

	// get aggregated count for all relevant committees from sync_committees_count_per_validator
	var countStatistics []struct {
		Period     uint64  `db:"period"`
		CountSoFar float64 `db:"count_so_far"`
	}

	query, args, errs := sqlx.In(`SELECT period, count_so_far FROM sync_committees_count_per_validator WHERE period IN (?) ORDER BY period ASC`, periodSlice)
	if errs != nil {
		return 0, errs
	}
	err = db.ReaderDb.Select(&countStatistics, db.ReaderDb.Rebind(query), args...)
	if err != nil {
		return 0, err
	}
	if len(countStatistics) != len(periodSlice) {
		return 0, fmt.Errorf("unable to retrieve all sync committee count statistics, required %v entries but got %v entries (epoch: %v)", len(periodSlice), len(countStatistics), epoch)
	}

	// transform query result to map for easy access
	periodInfoMap := make(map[uint64]float64)
	for _, pl := range countStatistics {
		periodInfoMap[pl.Period] = pl.CountSoFar
	}

	// calculate expected committies for every single validator and aggregate them
	expectedCommitties := 0.0
	for _, vi := range validatorsInfo {
		expectedCommitties += periodInfoMap[vi.LastPossibleSyncCommittee] - periodInfoMap[vi.FirstPossibleSyncCommittee]
	}

	// transform committees to slots
	expectedSlots = uint64(expectedCommitties * float64(utils.SlotsPerSyncCommittee()))

	return expectedSlots, nil
}

func getSyncCommitteeSlotsStatistics(validators []uint64, epoch uint64) (types.SyncCommitteesStats, error) {
	if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
		// no sync committee duties before altair fork
		return types.SyncCommitteesStats{}, nil
	}

	// collect aggregated sync committee stats from validator_stats table for all validators
	var syncStats struct {
		Participated int64 `db:"participated"`
		Missed       int64 `db:"missed"`
	}

	// validator_stats is updated only once a day, everything missing has to be collected from bigtable (which is slower than validator_stats)
	// check when the last update to validator_stats was
	epochsPerDay := utils.EpochsPerDay()
	lastExportedEpoch := uint64(0)
	lastExportedDay, err := services.LatestExportedStatisticDay()
	if err != nil && err != db.ErrNoStats {
		return types.SyncCommitteesStats{}, fmt.Errorf("error retrieving latest exported statistics day: %v", err)
	} else if err == nil {
		lastExportedEpoch = ((lastExportedDay + 1) * epochsPerDay) - 1
	}

	err = db.ReaderDb.Get(&syncStats, `SELECT SUM(COALESCE(participated_sync_total, 0)) AS participated, SUM(COALESCE(missed_sync_total, 0)) AS missed FROM validator_stats WHERE day = $1 AND validatorindex = ANY($2)`, lastExportedDay, pq.Array(validators))
	if err != nil {
		return types.SyncCommitteesStats{}, err
	}

	retv := types.SyncCommitteesStats{}
	retv.ParticipatedSlots = uint64(syncStats.Participated)
	retv.MissedSlots = uint64(syncStats.Missed)

	// if epoch is not yet exported, we may need to collect the data from bigtable
	if lastExportedEpoch < epoch {
		// get relevant period
		periodOfEpoch := utils.SyncPeriodOfEpoch(epoch)
		periods := []int64{int64(periodOfEpoch)}
		// if the committee period before the relevant one is also not yet fully exported, add it to the query
		if periods[0] > 0 && lastExportedEpoch < utils.FirstEpochOfSyncPeriod(periodOfEpoch)-1 {
			periods = append(periods, periods[0]-1)
		}

		// get all validators part of the relevant committees, grouped by period
		var syncCommitteeValidators []struct {
			Period     uint64        `db:"period"`
			Validators pq.Int64Array `db:"validators"`
		}
		query, args, err := sqlx.In(`
			SELECT period, COALESCE(ARRAY_AGG(distinct validatorindex), '{}') AS validators
			FROM sync_committees
			WHERE period IN (?) AND validatorindex IN (?)
			GROUP BY period
			ORDER BY period DESC
		`, periods, validators)
		if err != nil {
			return types.SyncCommitteesStats{}, err
		}
		err = db.ReaderDb.Select(&syncCommitteeValidators, db.ReaderDb.Rebind(query), args...)
		if err != nil {
			return types.SyncCommitteesStats{}, err
		}

		// if there validators are present in relevant periods, query bigtable
		if len(syncCommitteeValidators) > 0 {
			// flatten validator list
			vs := []uint64{}
			for _, scv := range syncCommitteeValidators {
				for _, v := range scv.Validators {
					vs = append(vs, uint64(v))
				}
			}

			// get sync stats from bigtable
			startSlot := (lastExportedEpoch + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch
			endSlot := epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch + utils.Config.Chain.ClConfig.SlotsPerEpoch - 1

			res, err := db.BigtableClient.GetValidatorSyncDutiesHistory(vs, startSlot, endSlot)
			if err != nil {
				return retv, fmt.Errorf("error retrieving validator sync participations data from bigtable: %v", err)
			}
			// add sync stats for validators in latest returned period
			latestPeriodCount := len(syncCommitteeValidators[0].Validators)
			syncStats := utils.AddSyncStats(vs[:latestPeriodCount], res, nil)
			// if latest returned period is the active one, add remaining scheduled slots
			firstEpochOfPeriod := utils.FirstEpochOfSyncPeriod(syncCommitteeValidators[0].Period)
			lastEpochOfPeriod := firstEpochOfPeriod + utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod - 1
			if firstEpochOfPeriod < utils.Config.Chain.ClConfig.AltairForkEpoch {
				// the first actual sync period starts at the altair fork epoch and might be shorter than all others
				// https://eth2book.info/capella/annotated-spec/#sync-committee-updates
				firstEpochOfPeriod = utils.Config.Chain.ClConfig.AltairForkEpoch
			}
			if lastEpochOfPeriod >= services.LatestEpoch() {
				syncStats.ScheduledSlots += utils.GetRemainingScheduledSyncDuties(latestPeriodCount, syncStats, lastExportedEpoch, firstEpochOfPeriod)
			}
			// add sync stats for validators in previous returned period
			utils.AddSyncStats(vs[latestPeriodCount:], res, &syncStats)
			retv.MissedSlots += syncStats.MissedSlots
			retv.ParticipatedSlots += syncStats.ParticipatedSlots
			retv.ScheduledSlots += syncStats.ScheduledSlots
		}
	}

	return retv, nil
}

type Cached struct {
	Data interface{}
	Ts   int64
}

var rocketpoolStats atomic.Value

func getRocketpoolStats() ([]interface{}, error) {
	cached := rocketpoolStats.Load()
	if cached != nil {
		cachedObj := cached.(*Cached)
		if cachedObj.Ts+10*60 > time.Now().Unix() { // cache for 30min
			return cachedObj.Data.([]interface{}), nil
		}
	}
	rows, err := db.ReaderDb.Query(`
		SELECT claim_interval_time, claim_interval_time_start, 
		current_node_demand, TRUNC(current_node_fee::decimal, 10)::float as current_node_fee, effective_rpl_staked,
		node_operator_rewards, TRUNC(reth_exchange_rate::decimal, 10)::float as reth_exchange_rate, reth_supply, rpl_price, total_eth_balance, total_eth_staking, 
		minipool_count, node_count, odao_member_count, 
		(SELECT TRUNC(((1 - (min(history.reth_exchange_rate) / max(history.reth_exchange_rate))) * 52.14)::decimal , 10) FROM (SELECT ts, reth_exchange_rate FROM rocketpool_network_stats LIMIT 168) history)::float as reth_apr  
		from rocketpool_network_stats ORDER BY ts desc LIMIT 1;
			`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data, err := utils.SqlRowsToJSON(rows)
	if err != nil {
		return nil, err
	}

	rocketpoolStats.Store(&Cached{
		Data: data,
		Ts:   time.Now().Unix(),
	})

	return data, nil
}

func getRocketpoolValidators(queryIndices []uint64) ([]interface{}, error) {
	rows, err := db.ReaderDb.Query(`
		SELECT
			rplm.node_address      AS node_address,
			rplm.address           AS minipool_address,
			TRUNC(rplm.node_fee::decimal, 10)::float          AS minipool_node_fee,
			rplm.deposit_type      AS minipool_deposit_type,
			rplm.status            AS minipool_status,
			rplm.penalty_count     AS penalty_count,
			rplm.status_time       AS minipool_status_time,
			rpln.timezone_location AS node_timezone_location,
			rpln.rpl_stake         AS node_rpl_stake,
			rpln.max_rpl_stake     AS node_max_rpl_stake,
			rpln.min_rpl_stake     AS node_min_rpl_stake,
			rpln.rpl_cumulative_rewards     AS rpl_cumulative_rewards,
			validators.validatorindex AS index,
			rpln.claimed_smoothing_pool     AS claimed_smoothing_pool,
			rpln.unclaimed_smoothing_pool   AS unclaimed_smoothing_pool,
			rpln.unclaimed_rpl_rewards      AS unclaimed_rpl_rewards,
			COALESCE(rpln.smoothing_pool_opted_in, false)    AS smoothing_pool_opted_in,
			COALESCE(rpln.deposit_credit, 0) as node_deposit_credit,
			COALESCE(rplm.node_deposit_balance, 0) AS node_deposit_balance,
			COALESCE(rplm.node_refund_balance, 0) AS node_refund_balance,
			COALESCE(rplm.user_deposit_balance, 0) AS user_deposit_balance,
			COALESCE(rplm.is_vacant, false) AS is_vacant,
			COALESCE(rpln.effective_rpl_stake, 0) as effective_rpl_stake,
			COALESCE(rplm.version, 0) AS version
		FROM rocketpool_minipools rplm 
		LEFT JOIN validators validators ON rplm.pubkey = validators.pubkey 
		LEFT JOIN rocketpool_nodes rpln ON rplm.node_address = rpln.address
		WHERE validatorindex = ANY($1)`, pq.Array(queryIndices))

	if err != nil {
		return nil, fmt.Errorf("error querying rocketpool minipools: %w", err)
	}
	defer rows.Close()

	return utils.SqlRowsToJSON(rows)
}

func getGeneralValidatorInfoForAppDashboard(queryIndices []uint64) ([]interface{}, error) {
	// we use MAX(validatorindex)+1 instead of COUNT(*) for querying the rank_count for performance-reasons
	rows, err := db.ReaderDb.Query(`
	WITH maxValidatorIndex AS (
		SELECT MAX(validatorindex)+1 as total_count
		FROM validator_performance
		WHERE validatorindex >= 0 AND validatorindex < 2147483647
	)
	SELECT 
		validators.validatorindex,
		pubkey,
		withdrawableepoch,
		withdrawalcredentials,
		slashed,
		activationeligibilityepoch,
		activationepoch,
		exitepoch,
		status,
		COALESCE(validator_names.name, '') AS name,
		COALESCE(validator_performance.cl_performance_1d, 0) AS performance1d,
		COALESCE(validator_performance.cl_performance_7d, 0) AS performance7d,
		COALESCE(validator_performance.cl_performance_31d, 0) AS performance31d,
		COALESCE(validator_performance.cl_performance_365d, 0) AS performance365d,
		COALESCE(validator_performance.cl_performance_total, 0) AS performanceTotal,
		COALESCE(validator_performance.rank7d, 0) AS rank7d,
		((validator_performance.rank7d::float * 100) / COALESCE((SELECT total_count FROM maxValidatorIndex), 1)) as rankpercentage
	FROM validators
	LEFT JOIN validator_performance ON validators.validatorindex = validator_performance.validatorindex
	LEFT JOIN validator_names ON validator_names.publickey = validators.pubkey
	WHERE validators.validatorindex = ANY($1)
	ORDER BY validators.validatorindex`, pq.Array(queryIndices))
	if err != nil {
		return nil, fmt.Errorf("error querying validators: %w", err)
	}
	defer rows.Close()

	data, err := utils.SqlRowsToJSON(rows)
	if err != nil {
		return nil, fmt.Errorf("error converting validators to json: %w", err)
	}

	g := new(errgroup.Group)

	var balances map[uint64][]*types.ValidatorBalance
	g.Go(func() error {
		var err error
		balances, err = db.BigtableClient.GetValidatorBalanceHistory(queryIndices, services.LatestEpoch(), services.LatestEpoch())
		if err != nil {
			return fmt.Errorf("error in GetValidatorBalanceHistory: %w", err)
		}
		return nil
	})

	var currentDayIncome map[uint64]int64
	g.Go(func() error {
		var err error
		currentDayIncome, err = db.GetCurrentDayClIncome(queryIndices)
		if err != nil {
			return fmt.Errorf("error in GetCurrentDayClIncome: %w", err)
		}
		return nil
	})

	var lastAttestationSlots map[uint64]uint64
	g.Go(func() error {
		var err error
		lastAttestationSlots, err = db.BigtableClient.GetLastAttestationSlots(queryIndices)
		if err != nil {
			return fmt.Errorf("error in GetLastAttestationSlots: %w", err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		return nil, fmt.Errorf("error in validator errgroup: %w", err)
	}

	for _, entry := range data {
		eMap, ok := entry.(map[string]interface{})
		if !ok {
			logger.Errorf("error converting validator data to map[string]interface{}")
			continue
		}

		validatorIndex, ok := eMap["validatorindex"].(int64)
		if !ok {
			logger.Errorf("error converting validatorindex to int64")
			continue
		}
		eMap["lastattestationslot"] = lastAttestationSlots[uint64(validatorIndex)]

		for balanceIndex, balance := range balances {
			if len(balance) == 0 {
				continue
			}

			if !ok {
				logger.Errorf("error converting validatorindex to int64")
				continue
			}
			if int64(balanceIndex) == validatorIndex {
				eMap["balance"] = balance[0].Balance
				eMap["effectivebalance"] = balance[0].EffectiveBalance
				eMap["performance1d"] = currentDayIncome[uint64(validatorIndex)]
				eMap["performancetotal"] = eMap["performancetotal"].(int64) + currentDayIncome[uint64(validatorIndex)]
			}
		}
	}

	return data, nil
}

func getValidatorEffectiveness(epoch uint64, indices []uint64) ([]*types.ValidatorEffectiveness, error) {
	data, err := db.BigtableClient.GetValidatorEffectiveness(indices, epoch)
	if err != nil {
		return nil, fmt.Errorf("error getting validator effectiveness from bigtable: %w", err)
	}
	for i := 0; i < len(data); i++ {
		// convert value to old api schema
		data[i].AttestationEfficiency = 1 + (1 - data[i].AttestationEfficiency/100)
	}
	return data, nil
}

type SyncCommitteesInfo struct {
	types.SyncCommitteesStats
	ExpectedSlots uint64 `json:"expectedSlots"`
}

type DashboardResponse struct {
	Validators           interface{}                          `json:"validators"`
	Effectiveness        interface{}                          `json:"effectiveness"`
	CurrentEpoch         interface{}                          `json:"currentEpoch"`
	OlderEpoch           interface{}                          `json:"olderEpoch"`
	Rocketpool           interface{}                          `json:"rocketpool_validators"`
	RocketpoolStats      interface{}                          `json:"rocketpool_network_stats"`
	ExecutionPerformance []types.ExecutionPerformanceResponse `json:"execution_performance"`
	CurrentSyncCommittee interface{}                          `json:"current_sync_committee"`
	NextSyncCommittee    interface{}                          `json:"next_sync_committee"`
	SyncCommitteesStats  *SyncCommitteesInfo                  `json:"sync_committees_stats"`
	ProposalLuckStats    *types.ApiProposalLuckResponse       `json:"proposal_luck_stats"`
}

func getEpoch(epoch int64) ([]interface{}, error) {
	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	rows, err := db.ReaderDb.Query(`SELECT attestationscount, attesterslashingscount, averagevalidatorbalance,
	blockscount, depositscount, eligibleether, epoch, (epoch <= $2) AS finalized, TRUNC(globalparticipationrate::decimal, 10)::float as globalparticipationrate, proposerslashingscount,
	totalvalidatorbalance, validatorscount, voluntaryexitscount, votedether
	FROM epochs WHERE epoch = $1`, epoch, latestFinalizedEpoch)
	if err != nil {
		return nil, fmt.Errorf("error querying epoch: %w", err)
	}
	defer rows.Close()
	return utils.SqlRowsToJSON(rows)
}

// ApiValidator godoc
// @Summary Get up to 100 validators
// @Tags Validator
// @Description Searching for too many validators based on their pubkeys will lead to a "URI too long" error
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse{data=[]types.APIValidatorResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey} [get]
func ApiValidatorGet(w http.ResponseWriter, r *http.Request) {
	getApiValidator(w, r)
}

// ApiValidator godoc
// @Summary Get up to 100 validators
// @Tags Validator
// @Description This POST endpoint exists because the GET endpoint can lead to a "URI too long" error when searching for too many validators based on their pubkeys.
// @Produce  json
// @Param  indexOrPubkey body types.DashboardRequest true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse{data=[]types.APIValidatorResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator [post]
func ApiValidatorPost(w http.ResponseWriter, r *http.Request) {
	getApiValidator(w, r)
}

// This endpoint supports both GET and POST but requires different swagger descriptions based on the type
func getApiValidator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	maxValidators := getUserPremium(r).MaxValidators

	var param string
	if r.Method == http.MethodGet {
		// Get the validators from the URL
		param = vars["indexOrPubkey"]
	} else {
		// Get the validators from the request body
		decoder := json.NewDecoder(r.Body)
		req := &types.DashboardRequest{}

		err := decoder.Decode(req)
		if err != nil {
			SendBadRequestResponse(w, r.URL.String(), "error decoding request body")
			return
		}
		param = req.IndicesOrPubKey
	}

	queryIndices, err := parseApiValidatorParamToIndices(param, maxValidators)

	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	lastExportedDay, err := services.LatestExportedStatisticDay()
	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "error retrieving data, please try again later")
		return
	}
	_, lastEpochOfDay := utils.GetFirstAndLastEpochForDay(lastExportedDay)
	cutoffSlot := (lastEpochOfDay * utils.Config.Chain.ClConfig.SlotsPerEpoch) + 1

	data := make([]*ApiValidatorResponse, 0)

	err = db.ReaderDb.Select(&data, `
		WITH today AS (
			SELECT
				w.validatorindex,
				COALESCE(SUM(w.amount), 0) as amount
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
			WHERE w.validatorindex = ANY($1) AND w.block_slot >= $2
			GROUP BY w.validatorindex
		),
		stats AS (
			SELECT
				vs.validatorindex,
				COALESCE(vs.withdrawals_amount_total, 0) as amount
			FROM validator_stats vs
			WHERE vs.validatorindex = ANY($1) AND vs.day = $3
		),
		withdrawals_summary AS (
			SELECT
				COALESCE(t.validatorindex, s.validatorindex) as validatorindex,
				COALESCE(t.amount, 0) + COALESCE(s.amount, 0) as total
			FROM today t
			FULL JOIN stats s ON t.validatorindex = s.validatorindex
		)
		SELECT
			v.validatorindex, '0x' || encode(pubkey, 'hex') as  pubkey, withdrawableepoch,
			'0x' || encode(withdrawalcredentials, 'hex') as withdrawalcredentials,
			slashed,
			activationeligibilityepoch,
			activationepoch,
			exitepoch,
			status,
			COALESCE(n.name, '') AS name,
			COALESCE(ws.total, 0) as total_withdrawals
		FROM validators v
		LEFT JOIN validator_names n ON n.publickey = v.pubkey
		LEFT JOIN withdrawals_summary ws ON ws.validatorindex = v.validatorindex
		WHERE v.validatorindex = ANY($1)
		ORDER BY v.validatorindex
	`, pq.Array(queryIndices), cutoffSlot, lastExportedDay)
	if err != nil {
		logger.Warnf("error retrieving validator data from db: %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	balances, err := db.BigtableClient.GetValidatorBalanceHistory(queryIndices, services.LatestEpoch(), services.LatestEpoch())
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve validator balance data")
		return
	}

	for _, validator := range data {
		for balanceIndex, balance := range balances {
			if len(balance) == 0 {
				continue
			}
			if validator.Validatorindex == int64(balanceIndex) {
				validator.Balance = int64(balance[0].Balance)
				validator.Effectivebalance = int64(balance[0].EffectiveBalance)
			}
		}
	}

	lastAttestationSlots, err := db.BigtableClient.GetLastAttestationSlots(queryIndices)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), fmt.Sprintf("error getting validator last attestation slots from bigtable: %v", err))
		return
	}

	for _, validator := range data {
		validator.Lastattestationslot = int64(lastAttestationSlots[uint64(validator.Validatorindex)])
	}

	j := json.NewEncoder(w)
	response := &types.ApiResponse{}
	response.Status = "OK"

	if len(data) == 1 {
		response.Data = data[0]
	} else {
		response.Data = data
	}
	err = j.Encode(response)

	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		logger.Errorf("error serializing json data for API %v route: %v", r.URL, err)
	}
}

type ApiValidatorResponse struct {
	Activationeligibilityepoch int64  `json:"activationeligibilityepoch"`
	Activationepoch            int64  `json:"activationepoch"`
	Balance                    int64  `json:"balance"`
	Effectivebalance           int64  `json:"effectivebalance"`
	Exitepoch                  int64  `json:"exitepoch"`
	Lastattestationslot        int64  `json:"lastattestationslot"`
	Name                       string `json:"name"`
	Pubkey                     string `json:"pubkey"`
	Slashed                    bool   `json:"slashed"`
	Status                     string `json:"status"`
	Validatorindex             int64  `json:"validatorindex"`
	Withdrawableepoch          int64  `json:"withdrawableepoch"`
	Withdrawalcredentials      string `json:"withdrawalcredentials"`
	TotalWithdrawals           uint64 `json:"total_withdrawals" db:"total_withdrawals"`
}

// ApiValidatorDailyStats godoc
// @Summary Get the daily validator stats by the validator index
// @Tags Validator
// @Produce  json
// @Param  index path string true "Validator index"
// @Param  end_day query string false "End day (default: latest day)"
// @Param  start_day query string false "Start day (default: 0)"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorDailyStatsResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/stats/{index} [get]
func ApiValidatorDailyStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	q := r.URL.Query()

	latestEpoch := services.LatestEpoch()

	latestDay := latestEpoch / utils.EpochsPerDay()

	startDay := int64(-1)
	endDay := int64(latestDay)

	if q.Get("end_day") != "" {
		end, err := strconv.ParseInt(q.Get("end_day"), 10, 64)
		if err != nil {
			SendBadRequestResponse(w, r.URL.String(), "invalid end_day parameter")
			return
		}
		if end < endDay {
			endDay = end
		}
	}

	if q.Get("start_day") != "" {
		start, err := strconv.ParseInt(q.Get("start_day"), 10, 64)
		if err != nil {
			SendBadRequestResponse(w, r.URL.String(), "invalid start_day parameter")
			return
		}
		if start > endDay {
			SendBadRequestResponse(w, r.URL.String(), "start_day must be less than end_day")
			return
		}
		if start > startDay {
			startDay = start
		}
	}

	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "invalid validator index")
		return
	}

	rows, err := db.ReaderDb.Query(`
		SELECT 
		validatorindex,
		day,
		start_balance,
		end_balance,
		min_balance,
		max_balance,
		start_effective_balance,
		end_effective_balance,
		min_effective_balance,
		max_effective_balance,
		COALESCE(missed_attestations, 0) AS missed_attestations,
		0 AS orphaned_attestations,
		COALESCE(proposed_blocks, 0) AS proposed_blocks,
		COALESCE(missed_blocks, 0) AS missed_blocks,
		COALESCE(orphaned_blocks, 0) AS orphaned_blocks,
		COALESCE(attester_slashings, 0) AS attester_slashings,
		COALESCE(proposer_slashings, 0) AS proposer_slashings,
		COALESCE(deposits, 0) AS deposits,
		COALESCE(deposits_amount, 0) AS deposits_amount,
		COALESCE(withdrawals, 0) AS withdrawals,
		COALESCE(withdrawals_amount, 0) AS withdrawals_amount,
		COALESCE(participated_sync, 0) AS participated_sync,
		COALESCE(missed_sync, 0) AS missed_sync,
		COALESCE(orphaned_sync, 0) AS orphaned_sync
	FROM validator_stats WHERE validatorindex = $1 and day <= $2 and day >= $3 ORDER BY day DESC`, index, endDay, startDay)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	addDayTime := func(dataEntryMap map[string]interface{}) error {
		day, ok := dataEntryMap["day"].(int64)
		if !ok {
			return fmt.Errorf("error type asserting day as an int")
		} else {
			dataEntryMap["day_start"] = utils.DayToTime(day)
			dataEntryMap["day_end"] = utils.DayToTime(day + 1)
		}
		return nil
	}

	returnQueryResultsAsArray(rows, w, r, addDayTime)
}

// ApiValidatorByEth1Address godoc
// @Summary Get all validators that belong to an eth1 address
// @Tags Validator
// @Produce  json
// @Param  eth1address path string true "Eth1 address from which the validator deposits were sent". It can also be a valid ENS name.
// @Param limit query string false "Limit the number of results (default: 2000)"
// @Param offset query string false "Offset the results (default: 0)"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorEth1Response}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/eth1/{eth1address} [get]
func ApiValidatorByEth1Address(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
	limitQuery := q.Get("limit")
	offsetQuery := q.Get("offset")

	limit, err := strconv.ParseInt(limitQuery, 10, 64)
	if err != nil {
		limit = 2000
	}

	offset, err := strconv.ParseInt(offsetQuery, 10, 64)
	if err != nil {
		offset = 0
	}

	if offset < 0 {
		offset = 0
	}

	if limit > (2000+offset) || limit <= 0 || limit <= offset {
		limit = 2000 + offset
	}

	vars := mux.Vars(r)
	search := ReplaceEnsNameWithAddress(vars["address"])
	eth1Address, err := hex.DecodeString(strings.Replace(search, "0x", "", -1))
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "invalid eth1 address provided")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT publickey, validatorindex, valid_signature FROM eth1_deposits LEFT JOIN validators ON eth1_deposits.publickey = validators.pubkey WHERE from_address = $1 GROUP BY publickey, validatorindex, valid_signature ORDER BY validatorindex OFFSET $2 LIMIT $3;", eth1Address, offset, limit)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiValidator godoc
// @Summary Get the income detail history of up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Param  latest_epoch query int false "The latest epoch to consider in the query"
// @Param  offset query int false "Number of items to skip"
// @Param  limit query int false "Maximum number of items to return, up to 100"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorIncomeHistoryResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/incomedetailhistory [get]
func ApiValidatorIncomeDetailsHistory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	latestEpoch, limit, err := getIncomeDetailsHistoryQueryParameters(r.URL.Query())
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	if len(queryIndices) == 0 {
		SendBadRequestResponse(w, r.URL.String(), "no validators provided")
		return
	}

	history, err := db.BigtableClient.GetValidatorIncomeDetailsHistory(queryIndices, latestEpoch-(limit-1), latestEpoch)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	responseData := make([]*types.ApiValidatorIncomeHistoryResponse, 0, uint64(len(history))*limit)

	epochsPerWeek := utils.EpochsPerDay() * 7
	for validatorIndex, epochs := range history {
		for epoch, income := range epochs {
			epochAtStartOfTheWeek := (epoch / epochsPerWeek) * epochsPerWeek

			txFeeRewardWei := ""
			if len(income.TxFeeRewardWei) > 0 {
				txFeeRewardWei = new(big.Int).SetBytes(income.TxFeeRewardWei).String()
			}

			responseIncome := &types.ApiValidatorIncomeHistory{
				AttestationSourceReward:            income.AttestationSourceReward,
				AttestationSourcePenalty:           income.AttestationSourcePenalty,
				AttestationTargetReward:            income.AttestationTargetReward,
				AttestationTargetPenalty:           income.AttestationTargetPenalty,
				AttestationHeadReward:              income.AttestationHeadReward,
				FinalityDelayPenalty:               income.FinalityDelayPenalty,
				ProposerSlashingInclusionReward:    income.ProposerSlashingInclusionReward,
				ProposerAttestationInclusionReward: income.ProposerAttestationInclusionReward,
				ProposerSyncInclusionReward:        income.ProposerSyncInclusionReward,
				SyncCommitteeReward:                income.SyncCommitteeReward,
				SyncCommitteePenalty:               income.SyncCommitteePenalty,
				SlashingReward:                     income.SlashingReward,
				SlashingPenalty:                    income.SlashingPenalty,
				TxFeeRewardWei:                     txFeeRewardWei,
				ProposalsMissed:                    income.ProposalsMissed}

			responseData = append(responseData, &types.ApiValidatorIncomeHistoryResponse{
				Income:         responseIncome,
				Epoch:          epoch,
				ValidatorIndex: validatorIndex,
				Week:           epoch / epochsPerWeek,
				WeekStart:      utils.EpochToTime(epochAtStartOfTheWeek),
				WeekEnd:        utils.EpochToTime(epochAtStartOfTheWeek + epochsPerWeek),
			})
		}
	}

	sort.Slice(responseData, func(i, j int) bool {
		if responseData[i].Epoch != responseData[j].Epoch {
			return responseData[i].Epoch > responseData[j].Epoch
		}
		return responseData[i].ValidatorIndex < responseData[j].ValidatorIndex
	})

	response := &types.ApiResponse{}
	response.Status = "OK"

	response.Data = responseData

	err = j.Encode(response)

	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		return
	}
}

func getIncomeDetailsHistoryQueryParameters(q url.Values) (uint64, uint64, error) {
	onChainLatestEpoch := services.LatestFinalizedEpoch()
	defaultLimit := uint64(100)

	latestEpoch := onChainLatestEpoch
	if q.Has("latest_epoch") {
		var err error
		latestEpoch, err = strconv.ParseUint(q.Get("latest_epoch"), 10, 64)
		if err != nil || latestEpoch > onChainLatestEpoch {
			return 0, 0, fmt.Errorf("invalid latest epoch parameter")
		}
	}

	if q.Has("offset") {
		offset, err := strconv.ParseUint(q.Get("offset"), 10, 64)
		if err != nil || offset > latestEpoch {
			return 0, 0, fmt.Errorf("invalid offset parameter")
		}
		latestEpoch -= offset
	}

	limit := defaultLimit
	if q.Has("limit") {
		var err error
		limit, err = strconv.ParseUint(q.Get("limit"), 10, 64)
		if err != nil || limit > defaultLimit || limit < 1 {
			return 0, 0, fmt.Errorf("invalid limit parameter")
		}
	}

	return latestEpoch, limit, nil
}

// ApiValidatorWithdrawals godoc
// @Summary Get the withdrawal history of up to 100 validators for the last 100 epochs. To receive older withdrawals modify the epoch paraum
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Param  epoch query int false "the start epoch for the withdrawal history (default: latest epoch)"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorWithdrawalResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/withdrawals [get]
func ApiValidatorWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	if len(queryIndices) == 0 {
		SendBadRequestResponse(w, r.URL.String(), "no or invalid validator indicies provided")
	}

	q := r.URL.Query()

	epoch, err := strconv.ParseUint(q.Get("epoch"), 10, 64)
	if err != nil {
		epoch = services.LatestEpoch()
	}

	// startEpoch and endEpoch are both inclusive, so substracting 99 here will result in a limit of 100 epochs
	endEpoch := epoch - 99
	if epoch < 99 {
		endEpoch = 0
	}

	data, err := db.GetValidatorsWithdrawals(queryIndices, endEpoch, epoch)
	if err != nil {
		logger.Errorf("error retrieving withdrawals for %v route: %v", r.URL.String(), err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	dataFormatted := make([]*types.ApiValidatorWithdrawalResponse, 0, len(data))
	for _, w := range data {
		dataFormatted = append(dataFormatted, &types.ApiValidatorWithdrawalResponse{
			Epoch:          w.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch,
			Slot:           w.Slot,
			Index:          w.Index,
			ValidatorIndex: w.ValidatorIndex,
			Amount:         w.Amount,
			BlockRoot:      fmt.Sprintf("0x%x", w.BlockRoot),
			Address:        fmt.Sprintf("0x%x", w.Address),
		})
	}

	response := &types.ApiResponse{}
	response.Status = "OK"

	response.Data = dataFormatted

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		return
	}
}

// ApiValidatorBlsChange godoc
// @Summary Gets the BLS withdrawal address change for up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorBlsChangeResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/blsChange [get]
func ApiValidatorBlsChange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	if len(queryIndices) == 0 {
		SendBadRequestResponse(w, r.URL.String(), "no or invalid validator indicies provided")
	}

	data, err := db.GetValidatorsBLSChange(queryIndices)
	if err != nil {
		logger.Errorf("error retrieving validators bls change for %v route: %v", r.URL.String(), err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	dataFormatted := make([]*types.ApiValidatorBlsChangeResponse, 0, len(data))

	for _, d := range data {
		dataFormatted = append(dataFormatted, &types.ApiValidatorBlsChangeResponse{
			Epoch:                    d.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch,
			Slot:                     d.Slot,
			BlockRoot:                fmt.Sprintf("0x%x", d.BlockRoot),
			Validatorindex:           d.Validatorindex,
			BlsPubkey:                fmt.Sprintf("0x%x", d.BlsPubkey),
			Address:                  fmt.Sprintf("0x%x", d.Address),
			Signature:                fmt.Sprintf("0x%x", d.Signature),
			WithdrawalCredentialsOld: fmt.Sprintf("0x%x", d.WithdrawalCredentialsOld),
			WithdrawalCredentialsNew: fmt.Sprintf("0x010000000000000000000000%x", d.Address),
		})
	}

	response := &types.ApiResponse{}
	response.Status = "OK"

	response.Data = dataFormatted

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		return
	}
}

// ApiValidator godoc
// @Summary Get the balance history of up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Param  latest_epoch query int false "The latest epoch to consider in the query"
// @Param  offset query int false "Number of items to skip"
// @Param  limit query int false "Maximum number of items to return, up to 100"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorBalanceHistoryResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/balancehistory [get]
func ApiValidatorBalanceHistory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	latestEpoch, limit, err := getBalanceHistoryQueryParameters(r.URL.Query())
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	if len(queryIndices) == 0 {
		SendBadRequestResponse(w, r.URL.String(), "no or invalid validator indicies provided")
	}

	history, err := db.BigtableClient.GetValidatorBalanceHistory(queryIndices, latestEpoch-(limit-1), latestEpoch)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	responseData := make([]*types.ApiValidatorBalanceHistoryResponse, 0, len(history)*101)

	epochsPerWeek := utils.EpochsPerDay() * 7
	for validatorIndex, balances := range history {
		for _, balance := range balances {
			epochAtStartOfTheWeek := (balance.Epoch / epochsPerWeek) * epochsPerWeek
			responseData = append(responseData, &types.ApiValidatorBalanceHistoryResponse{
				Balance:          balance.Balance,
				EffectiveBalance: balance.EffectiveBalance,
				Epoch:            balance.Epoch,
				Validatorindex:   validatorIndex,
				Week:             balance.Epoch / epochsPerWeek,
				WeekStart:        utils.EpochToTime(epochAtStartOfTheWeek),
				WeekEnd:          utils.EpochToTime(epochAtStartOfTheWeek + epochsPerWeek),
			})
		}
	}

	sort.Slice(responseData, func(i, j int) bool {
		if responseData[i].Epoch != responseData[j].Epoch {
			return responseData[i].Epoch > responseData[j].Epoch
		}
		return responseData[i].Validatorindex < responseData[j].Validatorindex
	})

	response := &types.ApiResponse{}
	response.Status = "OK"

	response.Data = responseData

	err = j.Encode(response)

	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		return
	}
}

func getBalanceHistoryQueryParameters(q url.Values) (uint64, uint64, error) {
	onChainLatestEpoch := services.LatestEpoch()
	defaultLimit := uint64(100)

	latestEpoch := onChainLatestEpoch
	if q.Has("latest_epoch") {
		var err error
		latestEpoch, err = strconv.ParseUint(q.Get("latest_epoch"), 10, 64)
		if err != nil || latestEpoch > onChainLatestEpoch {
			return 0, 0, fmt.Errorf("invalid latest epoch parameter")
		}
	}

	if q.Has("offset") {
		offset, err := strconv.ParseUint(q.Get("offset"), 10, 64)
		if err != nil || offset > latestEpoch {
			return 0, 0, fmt.Errorf("invalid offset parameter")
		}
		latestEpoch -= offset
	}

	limit := defaultLimit
	if q.Has("limit") {
		var err error
		limit, err = strconv.ParseUint(q.Get("limit"), 10, 64)
		if err != nil || limit > defaultLimit || limit < 1 {
			return 0, 0, fmt.Errorf("invalid limit parameter")
		}
	}

	return latestEpoch, limit, nil
}

// ApiValidatorPerformance godoc
// @Summary Get the current consensus reward performance of up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorPerformanceResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/performance [get]
func ApiValidatorPerformance(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query(`
	SELECT 
		validators.validatorindex, 
		COALESCE(validator_performance.cl_performance_1d, 0) AS performance1d, 
		COALESCE(validator_performance.cl_performance_7d, 0) AS performance7d, 
		COALESCE(validator_performance.cl_performance_31d, 0) AS performance31d, 
		COALESCE(validator_performance.cl_performance_365d, 0) AS performance365d, 
		COALESCE(validator_performance.cl_performance_total, 0) AS performanceTotal, 
		COALESCE(validator_performance.rank7d, 0) AS rank7d
	FROM validators 
	LEFT JOIN validator_performance ON 
		validators.validatorindex = validator_performance.validatorindex 
	WHERE validators.validatorindex = ANY($1) 
	ORDER BY validatorindex`, pq.Array(queryIndices))
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	data, err := utils.SqlRowsToJSON(rows)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not parse db results")
		return
	}

	currentDayIncome, err := db.GetCurrentDayClIncome(queryIndices)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "error retrieving current day income")
		return
	}

	latestEpoch := int64(services.LatestFinalizedEpoch())
	latestBalances, err := db.BigtableClient.GetValidatorBalanceHistory(queryIndices, uint64(latestEpoch), uint64(latestEpoch))
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "error retrieving balances")
		return
	}

	// create a map to easily check if a validator is part of data
	validatorIndexMap := make(map[uint64]bool)
	for _, entry := range data {
		eMap, ok := entry.(map[string]interface{})
		if !ok {
			logger.Errorf("error converting validator data to map[string]interface{}")
			continue
		}

		validatorIndex, ok := eMap["validatorindex"].(int64)
		if !ok {
			logger.Errorf("error converting validatorindex to int64")
			continue
		}

		validatorIndexMap[uint64(validatorIndex)] = true
	}

	// check for recently activated validators that have no performance data yet but already generate income
	for incomeValidatorIndex := range currentDayIncome {
		_, ok := validatorIndexMap[incomeValidatorIndex]
		if !ok {
			// validator not found in data, add minimum set of data
			data = append(data, map[string]interface{}{
				"validatorindex":   int64(incomeValidatorIndex),
				"performancetotal": int64(0), // has to exist and will be updated below
			})
		}
	}

	for _, entry := range data {
		eMap, ok := entry.(map[string]interface{})
		if !ok {
			logger.Errorf("error converting validator data to map[string]interface{}")
			continue
		}

		validatorIndex, ok := eMap["validatorindex"].(int64)
		if !ok {
			logger.Errorf("error converting validatorindex to int64")
			continue
		}

		eMap["balance"] = latestBalances[uint64(validatorIndex)][0].Balance
		eMap["performancetoday"] = currentDayIncome[uint64(validatorIndex)]
		eMap["performancetotal"] = eMap["performancetotal"].(int64) + currentDayIncome[uint64(validatorIndex)]
	}

	j := json.NewEncoder(w)
	SendOKResponse(j, r.URL.String(), []any{data})
}

// ApiValidatorExecutionPerformance godoc
// @Summary Get the current execution reward performance of up to 100 validators. If block was produced via mev relayer, this endpoint will use the relayer data as block reward instead of the normal block reward.
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorExecutionPerformanceResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/execution/performance [get]
func ApiValidatorExecutionPerformance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	result, err := getValidatorExecutionPerformance(queryIndices)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		logger.WithError(err).Error("can not getValidatorExecutionPerformance")
		return
	}

	SendOKResponse(j, r.URL.String(), []any{result})
}

// ApiValidatorAttestationEffectiveness godoc
// @Summary DEPRECIATED - USE /attestationefficiency (Get the current performance of up to 100 validators)
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/attestationeffectiveness [get]
func ApiValidatorAttestationEffectiveness(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	data, err := getValidatorEffectiveness(services.LatestEpoch()-1, queryIndices)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	response := &types.ApiResponse{}
	response.Status = "OK"

	response.Data = data

	err = j.Encode(response)

	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		return
	}
}

// ApiValidatorAttestationEfficiency godoc
// @Summary Get the current performance of up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/attestationefficiency [get]
func ApiValidatorAttestationEfficiency(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	data, err := getValidatorEffectiveness(services.LatestEpoch()-1, queryIndices)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	response := &types.ApiResponse{}
	response.Status = "OK"

	response.Data = data

	err = j.Encode(response)

	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		return
	}
}

// func getAttestationEfficiencyQuery(epoch int64, queryIndices []uint64) (*sql.Rows, error) {
// 	return db.ReaderDb.Query(`
// 	SELECT aa.validatorindex, validators.pubkey, COALESCE(
// 		AVG(1 + inclusionslot - COALESCE((
// 			SELECT MIN(slot)
// 			FROM blocks
// 			WHERE slot > aa.attesterslot AND blocks.status = '1'
// 		), 0)
// 	), 0)::float AS attestation_efficiency
// 	FROM attestation_assignments_p aa
// 	INNER JOIN blocks ON blocks.slot = aa.inclusionslot AND blocks.status <> '3'
// 	INNER JOIN validators ON validators.validatorindex = aa.validatorindex
// 	WHERE aa.week >= $1 / 1575 AND aa.epoch > $1 AND (validators.validatorindex = ANY($2)) AND aa.inclusionslot > 0
// 	GROUP BY aa.validatorindex, validators.pubkey
// 	ORDER BY aa.validatorindex
// 	`, epoch, pq.Array(queryIndices))
// }

// ApiValidatorLeaderboard godoc
// @Summary Get the current top 100 performing validators (using the income over the last 7 days)
// @Tags Validator
// @Produce  json
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorPerformanceResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/leaderboard [get]
func ApiValidatorLeaderboard(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	rows, err := db.ReaderDb.Query(`
			SELECT 
				balance, 
				COALESCE(validator_performance.cl_performance_1d, 0) AS performance1d, 
				COALESCE(validator_performance.cl_performance_7d, 0) AS performance7d, 
				COALESCE(validator_performance.cl_performance_31d, 0) AS performance31d, 
				COALESCE(validator_performance.cl_performance_365d, 0) AS performance365d, 
				COALESCE(validator_performance.cl_performance_total, 0) AS performanceTotal, 
				rank7d, 
				validatorindex
			FROM validator_performance 
			ORDER BY rank7d ASC LIMIT 100`)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiValidatorDeposits godoc
// @Summary Get all eth1 deposits for up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorDepositsResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/deposits [get]
func ApiValidatorDeposits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	pubkeys, err := parseApiValidatorParamToPubkeys(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query(
		`SELECT amount, block_number, block_ts, from_address, merkletree_index, publickey, removed, signature, tx_hash, tx_index, tx_input, valid_signature, withdrawal_credentials FROM eth1_deposits 
		WHERE publickey = ANY($1)`, pubkeys,
	)
	if err != nil {
		logger.WithError(err).Error("could not retrieve db results")
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiValidatorAttestations godoc
// @Summary Get all attestations during the last 100 epochs for up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} types.ApiResponse{[]types.ApiValidatorAttestationsResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/attestations [get]
func ApiValidatorAttestations(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	history, err := db.BigtableClient.GetValidatorAttestationHistory(queryIndices, services.LatestEpoch()-99, services.LatestEpoch())
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	responseData := make([]*types.ApiValidatorAttestationsResponse, 0, len(history)*100)

	epochsPerWeek := utils.EpochsPerDay() * 7
	for validatorIndex, balances := range history {
		for _, attestation := range balances {
			epochAtStartOfTheWeek := (attestation.Epoch / epochsPerWeek) * epochsPerWeek
			responseData = append(responseData, &types.ApiValidatorAttestationsResponse{
				AttesterSlot:   attestation.AttesterSlot,
				CommitteeIndex: 0,
				Epoch:          attestation.Epoch,
				InclusionSlot:  attestation.InclusionSlot,
				Status:         attestation.Status,
				ValidatorIndex: validatorIndex,
				Week:           attestation.Epoch / epochsPerWeek,
				WeekStart:      utils.EpochToTime(epochAtStartOfTheWeek),
				WeekEnd:        utils.EpochToTime(epochAtStartOfTheWeek + epochsPerWeek),
			})
		}
	}

	sort.Slice(responseData, func(i, j int) bool {
		if responseData[i].Epoch != responseData[j].Epoch {
			return responseData[i].Epoch > responseData[j].Epoch
		}
		return responseData[i].ValidatorIndex < responseData[j].ValidatorIndex
	})

	response := &types.ApiResponse{}
	response.Status = "OK"

	response.Data = responseData

	err = j.Encode(response)

	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		return
	}
}

// ApiValidatorProposals godoc
// @Summary Get all proposed blocks during the last 100 epochs for up to 100 validators. Optionally set the epoch query parameter to look back further.
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Param  epoch query string false "Page the result by epoch"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiValidatorProposalsResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/{indexOrPubkey}/proposals [get]
func ApiValidatorProposals(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators
	q := r.URL.Query()

	epochQuery := uint64(0)
	if q.Get("epoch") == "" {
		epochQuery = services.LatestEpoch()
	} else {
		var err error
		epochQuery, err = strconv.ParseUint(q.Get("epoch"), 10, 64)
		if err != nil {
			SendBadRequestResponse(w, r.URL.String(), err.Error())
			return
		}
	}

	queryIndices, err := parseApiValidatorParamToIndices(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}
	if epochQuery < 100 {
		epochQuery = 100
	}

	rows, err := db.ReaderDb.Query(`
	SELECT 
		b.epoch,
		b.slot,
		b.blockroot,
		b.parentroot,
		b.stateroot,
		b.signature,
		b.attestationscount,
		b.attesterslashingscount,
		b.depositscount,
		b.eth1data_blockhash,
		b.eth1data_depositcount,
		b.eth1data_depositroot,
		b.exec_base_fee_per_gas,
		b.exec_block_hash,
		b.exec_block_number,
		b.exec_extra_data,
		b.exec_fee_recipient,
		b.exec_gas_limit,
		b.exec_gas_used,
		b.exec_logs_bloom,
		b.exec_parent_hash,
		b.exec_random,
		b.exec_receipts_root,
		b.exec_state_root,
		b.exec_timestamp,
		b.exec_transactions_count,
		b.graffiti,
		b.graffiti_text,
		b.proposer,
		b.proposerslashingscount,
		b.randaoreveal,
		b.status,
		b.syncaggregate_bits,
		b.syncaggregate_participation,
		b.syncaggregate_signature,
		b.voluntaryexitscount
	FROM blocks as b 
	LEFT JOIN validators ON validators.validatorindex = b.proposer 
	WHERE (proposer = ANY($1)) and epoch <= $2 AND epoch >= $3 
	ORDER BY proposer, epoch desc, slot desc`, pq.Array(queryIndices), epochQuery, epochQuery-100)
	if err != nil {
		logger.Errorf("could not retrieve db results: %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	returnQueryResultsAsArray(rows, w, r)
}

// ApiGraffitiwall godoc
// @Summary Get the most recent pixels that have been painted.
// @Tags Misc
// @Description Returns the most recent pixels that have been painted during the last 10000 slots.
// @Description Optionally set the slot query parameter to look back further.
// @Description Boundary coordinates are included.
// @Description X = 0 and Y = 0 start at the upper left corner.
// @Description Returns an error if an invalid area is provided by the coordinates.
// @Produce  json
// @Param startx query int false "Start X offset" default(0)
// @Param starty query int false "Start Y offset" default(0)
// @Param endx query int false "End X limit" default(999)
// @Param endy query int false "End Y limit" default(999)
// @Param startSlot query string false "Start slot to query (end slot - 10000 if empty)"
// @Param slot query string false "End slot to query"
// @Param summarize query bool false "Only return end state of each pixel" default(true)
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/graffitiwall [get]
func ApiGraffitiwall(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
	startSlot := uint64(0)
	endSlot := uint64(0)
	if q.Get("slot") == "" {
		endSlot = services.LatestSlot()
	} else {
		var err error
		endSlot, err = strconv.ParseUint(q.Get("slot"), 10, 64)
		if err != nil {
			// logger.WithError(err).Errorf("invalid slot provided: %v", err)
			SendBadRequestResponse(w, r.URL.String(), "invalid slot provided")
			return
		}
	}
	endSlot = utilMath.MaxU64(endSlot, 10000)

	if q.Get("startSlot") == "" {
		startSlot = endSlot - 10000
	} else {
		var err error
		startSlot, err = strconv.ParseUint(q.Get("startSlot"), 10, 64)
		if err != nil {
			// logger.WithError(err).Errorf("invalid startSlot provided: %v", err)
			SendBadRequestResponse(w, r.URL.String(), "invalid startSlot provided")
			return
		}
		if startSlot > endSlot {
			// logger.Errorf("start slot greater than end slot")
			SendBadRequestResponse(w, r.URL.String(), "start slot greater than end slot")
			return
		}
	}

	defaultStartPxl := uint64(0)
	defaultEndPxl := uint64(999)

	startX := utilMath.MinU64(parseUintWithDefault(q.Get("startx"), defaultStartPxl), defaultEndPxl)
	startY := utilMath.MinU64(parseUintWithDefault(q.Get("starty"), defaultStartPxl), defaultEndPxl)
	endX := utilMath.MinU64(parseUintWithDefault(q.Get("endx"), defaultEndPxl), defaultEndPxl)
	endY := utilMath.MinU64(parseUintWithDefault(q.Get("endy"), defaultEndPxl), defaultEndPxl)

	if startX > endX || startY > endY {
		// logger.Error("invalid area provided by the coordinates")
		SendBadRequestResponse(w, r.URL.String(), "invalid area provided by the coordinates")
		return
	}

	summarize := true
	summarizeParam := q.Get("summarize")
	if summarizeParam != "" {
		var err error
		summarize, err = strconv.ParseBool(summarizeParam)
		if err != nil {
			// logger.WithError(err).Errorf("invalid value for summarize provided: %v", err)
			SendBadRequestResponse(w, r.URL.String(), "invalid value for summarize provided")
			return
		}
	}

	summarize_query := ""
	if summarize {
		// only pick latest pixel update
		summarize_query = "DISTINCT ON (x, y) "
	}

	rows, err := db.ReaderDb.Query(`
	SELECT `+summarize_query+`
		x,
		y,
		color,
		slot,
		validator
	FROM graffitiwall
	WHERE slot BETWEEN $1 AND $2 AND x BETWEEN $3 AND $4 AND y BETWEEN $5 AND $6
	ORDER BY x, y, slot DESC`, startSlot, endSlot, startX, endX, startY, endY)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.WithError(err).Error("could not retrieve db results")
		}
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResultsAsArray(rows, w, r)
}

// ApiChart godoc
// @Summary Returns charts from the page https://beaconcha.in/charts as PNG
// @Tags Misc
// @Produce  json
// @Param  chart path string true "Chart name (see https://github.com/gobitfly/eth2-beaconchain-explorer/blob/master/services/charts_updater.go#L20 for all available names)"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/chart/{chart} [get]
func ApiChart(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	chartName := vars["chart"]

	var image []byte
	err := db.ReaderDb.Get(&image, "SELECT image FROM chart_images WHERE name = $1", chartName)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "no data available for the requested chart")
		return
	}

	w.Header().Set("Content-Type", "image/png")

	_, err = w.Write(image)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "error writing chart data")
		return
	}
}

// APIGetToken godoc
// @Summary Exchange your oauth code for an access token or refresh your access token
// @Tags User
// @Produce  json
// @Param grant_type formData string true "grant_type use authorization_code for oauth code or refresh_token if you wish to refresh an token"
// @Param code formData string false "Only required when using authorization_code grant type. Code received via oauth redirect_uri"
// @Param redirect_uri formData string false "Only required when using authorization_code grant type. Must match the redirect_uri from your oauth flow."
// @Param refresh_token formData string false "Only required when using refresh_token grant type. The refresh_token you received during authorization_code flow."
// @Header 200 jwt Authorization "Authorization Only required when using refresh_token grant type. Use any access token that is linked with your refresh_token."
// @Success 200 {object} utils.OAuthResponse
// @Failure 400 {object} utils.OAuthErrorResponse
// @Failure 500 {object} utils.OAuthErrorResponse
// @Security OAuthAccessCode
// @Router /api/v1/user/token [post]
func APIGetToken(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		getTokenByCode(w, r)
	case "refresh_token":
		getTokenByRefresh(w, r)
	default:
		j := json.NewEncoder(w)
		w.WriteHeader(http.StatusBadRequest)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.InvalidGrant, "grant type must be authorization_code or refresh_token")
	}
}

func getTokenByCode(w http.ResponseWriter, r *http.Request) {
	j := json.NewEncoder(w)

	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	deviceName := getDeviceNameFromUA(r.Header.Get("User-Agent"))

	// Check if redirect URI is correct
	_, err := db.GetAppDataFromRedirectUri(redirectURI)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.InvalidRequest, "redirect_uri do not match")
		return
	}

	// Hash code, we only store codes as sha256 hash in db
	codeHashed := utils.HashAndEncode(code)

	// Check if code entry exists and isn't expired (codes expire after 5 minutes)
	codeAuthData, err := db.GetUserAuthDataByAuthorizationCode(codeHashed)
	if err != nil {
		logger.Errorf("Error hashed code can not be found in table: %v | Error: %v", codeHashed, err)
		w.WriteHeader(http.StatusUnauthorized)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.AccessDenied, "access_token or refresh_token invalid")
		return
	}

	// Create refresh token
	refreshTokenBytes, err := utils.GenerateRandomBytesSecure(32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.ServerError, "can not generate refresh_token")
		return
	}

	refreshToken := hex.EncodeToString(refreshTokenBytes)   // return to user
	refreshTokenHashed := utils.HashAndEncode(refreshToken) // save hashed in db

	// save refreshtoken hashed in db
	deviceID, errDb := db.InsertUserDevice(codeAuthData.UserID, refreshTokenHashed, deviceName, codeAuthData.AppID)
	if errDb != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.ServerError, "can not store auth info")
		return
	}

	pkg, err := db.GetUserPremiumPackage(codeAuthData.UserID)
	if err != nil {
		pkg.Package = "standard"
	}

	var theme string = ""
	if pkg.Store == "ethpool" {
		theme = "ethpool"
	}

	// Create access token
	token, expiresIn, err := utils.CreateAccessToken(codeAuthData.UserID, codeAuthData.AppID, deviceID, pkg.Package, theme)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.ServerError, "can not create access_token")
		return
	}

	// finally creating the oauth message
	utils.SendOAuthResponse(j, r.URL.String(), token, refreshToken, expiresIn)
}

func getTokenByRefresh(w http.ResponseWriter, r *http.Request) {
	j := json.NewEncoder(w)

	refreshToken := r.FormValue("refresh_token")
	accessToken := r.Header.Get("Authorization")

	// hash refreshtoken
	refreshTokenHashed := utils.HashAndEncode(refreshToken)

	// Extract userId from JWT. Note that this is just an unvalidated claim!
	// Do not use userIDClaim as userID until confirmed by refreshToken validation
	unsafeClaims, err := utils.UnsafeGetClaims(accessToken)
	if err != nil {
		logger.Errorf("Error access_token claim: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.InvalidRequest, "access_token validation failed")
		return
	}

	// confirm all claims via db lookup and refreshtoken check
	userID, err := db.GetByRefreshToken(unsafeClaims.UserID, unsafeClaims.AppID, unsafeClaims.DeviceID, refreshTokenHashed)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warnf("No refresh token found for user: %v | %v", unsafeClaims.UserID, refreshTokenHashed)
		} else {
			logger.Errorf("Error refreshtoken check: %v | %v | %v", unsafeClaims.UserID, refreshTokenHashed, err)
		}
		w.WriteHeader(http.StatusUnauthorized)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.UnauthorizedClient, "invalid token credentials")
		return
	}

	pkg, err := db.GetUserPremiumPackage(userID)
	if err != nil {
		pkg.Package = "standard"
	}

	var theme string = ""
	if pkg.Store == "ethpool" {
		theme = "ethpool"
	}

	// Create access token
	token, expiresIn, err := utils.CreateAccessToken(userID, unsafeClaims.AppID, unsafeClaims.DeviceID, pkg.Package, theme)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.ServerError, "can not create access_token")
		return
	}

	// finally creating the oauth message
	utils.SendOAuthResponse(j, r.URL.String(), token, "", expiresIn)
}

// Device name is limited to 20 chars
func getDeviceNameFromUA(userAgent string) string {
	ua := user_agent.New(userAgent)
	platformLen := len(ua.Platform())
	osLen := len(ua.OS())

	if platformLen+osLen > 19 {
		if osLen <= 20 && osLen > 0 {
			return ua.OS()
		} else if platformLen <= 20 && platformLen > 0 {
			return ua.Platform()
		} else {
			return "Unknown"
		}
	} else if platformLen+osLen > 0 {
		return ua.Platform() + " " + ua.OS()
	} else {
		return "Unknown"
	}
}

// MobileNotificationUpdatePOST godoc
// @Summary Register or update your mobile notification token
// @Tags User
// @Produce  json
// @Param token body string true "Your device`s firebase notification token"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/mobile/notify/register [post]
func MobileNotificationUpdatePOST(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	notifyToken := FormValueOrJSON(r, "token")

	claims := getAuthClaims(r)

	err2 := db.MobileNotificatonTokenUpdate(claims.UserID, claims.DeviceID, notifyToken)
	if err2 != nil {
		SendBadRequestResponse(w, r.URL.String(), "Can not save notify token")
		return
	}

	OKResponse(w, r)
}

func RegisterEthpoolSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pkg := FormValueOrJSON(r, "package")
	ethpoolUserID := FormValueOrJSON(r, "user_id")
	signature := FormValueOrJSON(r, "signature")

	localSignature := hmacSign(fmt.Sprintf("ETHPOOL %v %v", pkg, ethpoolUserID))
	if signature != localSignature {
		w.WriteHeader(http.StatusBadRequest)
		logger.Errorf("signature mismatch %v | %v", signature, localSignature)
		SendBadRequestResponse(w, r.URL.String(), "Unauthorized: signature not valid")
		return
	}

	claims := getAuthClaims(r)

	subscriptionCount, err := db.GetAppSubscriptionCount(claims.UserID)
	if err != nil {
		utils.LogError(err, "could not get subscription count", 0)
		sendServerErrorResponse(w, r.URL.String(), "Internal Server Error")
		return
	}
	if subscriptionCount >= USER_SUBSCRIPTION_LIMIT {
		sendErrorWithCodeResponse(w, r.URL.String(), "Conflicting Request: reached max subscription limit", http.StatusConflict)
		return
	}

	parsedBase := types.MobileSubscription{
		ProductID:   pkg,
		Valid:       true,
		PriceMicros: 0,
		Currency:    "USD",
		Transaction: types.MobileSubscriptionTransactionGeneric{
			Type:    "ethpool",
			Receipt: hmacSign(fmt.Sprintf("BEACONCHAIN %v", ethpoolUserID)), // use own signed message that excludes pkg to mitigate 2x free (goldfish and whale) keys
			ID:      pkg,
		},
	}

	err = db.InsertMobileSubscription(nil, claims.UserID, parsedBase, parsedBase.Transaction.Type, parsedBase.Transaction.Receipt, 0, "", "")
	if err != nil {
		logger.Errorf("could not save subscription data %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		SendBadRequestResponse(w, r.URL.String(), "Can not save subscription data")
		return
	}

	OKResponse(w, r)
}

func hmacSign(data string) string {
	h := hmac.New(sha256.New, []byte(utils.Config.Frontend.BeaconchainETHPoolBridgeSecret))
	h.Write([]byte(data))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

func RegisterMobileSubscriptions(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var parsedBase types.MobileSubscription
	err := json.Unmarshal(gorillacontext.Get(r, utils.JsonBodyNakedKey).([]byte), &parsedBase)

	if err != nil {
		logger.Errorf("error parsing body | err: %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not parse body")
		return
	}

	claims := getAuthClaims(r)

	subscriptionCount, err := db.GetAppSubscriptionCount(claims.UserID)
	if err != nil {
		utils.LogError(err, "could not get subscription count", 0)
		sendServerErrorResponse(w, r.URL.String(), "Internal Server Error")
		return
	}
	if subscriptionCount >= USER_SUBSCRIPTION_LIMIT {
		sendErrorWithCodeResponse(w, r.URL.String(), "Conflicting Request: reached max subscription limit", http.StatusConflict)
		return
	}

	// Verify subscription with apple/google
	verifyPackage := &types.PremiumData{
		ID:        0,
		Receipt:   parsedBase.Transaction.Receipt,
		Store:     parsedBase.Transaction.Type,
		Active:    false,
		ProductID: parsedBase.ProductID,
		ExpiresAt: time.Now(),
	}

	// we can ignore this error since it always returns a result object and err
	// case is not needed on receipt insert
	validationResult, _ := exporter.VerifyReceipt(nil, verifyPackage)
	parsedBase.Valid = validationResult.Valid

	err = db.InsertMobileSubscription(nil, claims.UserID, parsedBase, parsedBase.Transaction.Type, parsedBase.Transaction.Receipt, validationResult.ExpirationDate, validationResult.RejectReason, "")
	if err != nil {
		logger.Errorf("could not save subscription data %v", err)
		SendBadRequestResponse(w, r.URL.String(), "Can not save subscription data")
		return
	}

	if !parsedBase.Valid {
		logger.Errorf("receipt is not valid %v", validationResult.RejectReason)
		SendBadRequestResponse(w, r.URL.String(), "receipt is not valid")
		return
	}

	OKResponse(w, r)
}

type PremiumUser struct {
	Package                string
	MaxValidators          int
	MaxStats               uint64
	MaxNodes               uint64
	WidgetSupport          bool
	NotificationThresholds bool
	NoAds                  bool
}

func getUserPremium(r *http.Request) PremiumUser {
	var pkg string = ""

	if strings.HasPrefix(r.URL.Path, "/api/") {
		claims := getAuthClaims(r)
		if claims != nil {
			pkg = claims.Package
		}
	} else {
		sessionUser := getUser(r)
		if sessionUser.Authenticated {
			pkg = sessionUser.Subscription
		}
	}

	return GetUserPremiumByPackage(pkg)
}

func GetUserPremiumByPackage(pkg string) PremiumUser {
	result := PremiumUser{
		Package:                "standard",
		MaxValidators:          100,
		MaxStats:               180,
		MaxNodes:               1,
		WidgetSupport:          false,
		NotificationThresholds: false,
		NoAds:                  false,
	}

	if pkg == "" || pkg == "standard" {
		return result
	}

	result.Package = pkg
	result.MaxStats = 43200
	result.NotificationThresholds = true
	result.NoAds = true

	if result.Package != "plankton" {
		result.WidgetSupport = true
	}

	if result.Package == "goldfish" {
		result.MaxNodes = 2
	}
	if result.Package == "whale" {
		result.MaxValidators = 300
		result.MaxNodes = 10
	}

	return result
}

func GetMobileWidgetStatsPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)

	var parsedBody types.DashboardRequest
	err := decoder.Decode(&parsedBody)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not read body")
		return
	}

	GetMobileWidgetStats(w, r, parsedBody.IndicesOrPubKey)
}

func GetMobileWidgetStatsGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	GetMobileWidgetStats(w, r, vars["indexOrPubkey"])
}

func GetMobileWidgetStats(w http.ResponseWriter, r *http.Request, indexOrPubkey string) {
	epoch := int64(services.LatestEpoch())
	if epoch < 0 {
		epoch = 0
	}
	prime := getUserPremium(r)

	queryIndices, err := parseApiValidatorParamToIndices(indexOrPubkey, prime.MaxValidators)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	g, _ := errgroup.WithContext(context.Background())
	var rocketpoolStats []any
	var efficiencyRows *sql.Rows
	var validatorRows *sql.Rows

	g.Go(func() error {
		validatorRows, err = db.ReaderDb.Query(
			`SELECT 
					validators.pubkey, 
					slashed, 
					activationeligibilityepoch, 
					activationepoch, 
					exitepoch, 
					validators.status, 
					validator_performance.balance, 
					COALESCE(validator_performance.cl_performance_1d, 0) AS performance1d, 
					COALESCE(validator_performance.cl_performance_7d, 0) AS performance7d, 
					COALESCE(validator_performance.cl_performance_31d, 0) AS performance31d, 
					COALESCE(validator_performance.cl_performance_365d, 0) AS performance365d, 
					COALESCE(validator_performance.cl_performance_total, 0) AS performanceTotal, 
					validator_performance.rank7d, 
					validator_performance.validatorindex,
					TRUNC(rplm.node_fee::decimal, 10)::float  AS minipool_node_fee,
					TRUNC(rplm.node_deposit_balance::decimal / 1e9, 10)::float  AS minipool_deposit  
				FROM validators 
				LEFT JOIN validator_performance ON validators.validatorindex = validator_performance.validatorindex 
				LEFT JOIN rocketpool_minipools rplm ON rplm.pubkey = validators.pubkey
				WHERE validator_performance.validatorindex = ANY($1) ORDER BY validator_performance.validatorindex`,
			pq.Array(queryIndices),
		)
		return err
	})

	g.Go(func() error {
		rocketpoolStats, err = getRocketpoolStats()
		return err
	})

	err = g.Wait()
	if validatorRows != nil {
		defer validatorRows.Close()
	}
	if efficiencyRows != nil {
		defer efficiencyRows.Close()
	}
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	generalData, err := utils.SqlRowsToJSON(validatorRows)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not parse db results")
		return
	}

	balances, err := db.BigtableClient.GetValidatorBalanceHistory(queryIndices, uint64(epoch), uint64(epoch))
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "error retrieving validator balance data")
		return
	}

	lastAttestationSlots, err := db.BigtableClient.GetLastAttestationSlots(queryIndices)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "error retrieving validator balance data")
		return
	}

	currentDayIncome, err := db.GetCurrentDayClIncome(queryIndices)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "error retrieving current day income")
		return
	}

	for _, entry := range generalData {
		eMap, ok := entry.(map[string]interface{})
		if !ok {
			logger.Errorf("error converting validator data to map[string]interface{}")
			continue
		}

		validatorIndex, ok := eMap["validatorindex"].(int64)

		if !ok {
			logger.Errorf("error converting validatorindex to int64")
			continue
		}

		eMap["lastattestationslot"] = lastAttestationSlots[uint64(validatorIndex)]

		for balanceIndex, balance := range balances {
			if len(balance) == 0 {
				continue
			}

			if int64(balanceIndex) == validatorIndex {
				eMap["effectivebalance"] = balance[0].EffectiveBalance
				eMap["performance1d"] = currentDayIncome[uint64(validatorIndex)]
				eMap["performancetotal"] = eMap["performancetotal"].(int64) + currentDayIncome[uint64(validatorIndex)]
			}
		}
	}

	efficiencyData, err := getValidatorEffectiveness(services.LatestEpoch()-1, queryIndices)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not parse db results")
		return
	}

	data := &types.WidgetResponse{
		Eff:             efficiencyData,
		Validator:       generalData,
		Epoch:           epoch,
		RocketpoolStats: rocketpoolStats,
	}

	j := json.NewEncoder(w)
	SendOKResponse(j, r.URL.String(), []any{data})
}

// MobileDeviceSettings godoc
// @Summary Get your device settings, currently only whether to enable mobile notifcations or not
// @Tags User
// @Produce json
// @Success 200 {object} types.ApiResponse{data=types.MobileSettingsData}
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/user/mobile/settings [get]
func MobileDeviceSettings(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	claims := getAuthClaims(r)

	rows, err := db.MobileDeviceSettingsSelect(claims.UserID, claims.DeviceID)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	defer rows.Close()

	returnQueryResults(rows, w, r)
}

// MobileDeviceSettingsPOST godoc
// @Summary Changing your devices mobile settings
// @Tags User
// @Produce json
// @Param notify_enabled body bool true "Whether to enable mobile notifications for this device or not"
// @Success 200 {object} types.ApiResponse{data=types.MobileSettingsData}
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/user/mobile/settings [post]
func MobileDeviceSettingsPOST(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	notifyEnabled := FormValueOrJSON(r, "notify_enabled")
	active := FormValueOrJSON(r, "active")

	claims := getAuthClaims(r)
	var userDeviceID uint64
	var userID uint64

	if claims == nil {
		customDeviceID := FormValueOrJSON(r, "id")
		temp, err := strconv.ParseUint(customDeviceID, 10, 64)
		if err != nil {
			logger.Errorf("error parsing id %v | err: %v", customDeviceID, err)
			SendBadRequestResponse(w, r.URL.String(), "could not parse id")
			return
		}
		userDeviceID = temp
		sessionUser := getUser(r)
		if !sessionUser.Authenticated {
			SendBadRequestResponse(w, r.URL.String(), "not authenticated")
			return
		}
		userID = sessionUser.UserID
	} else {
		userDeviceID = claims.DeviceID
		userID = claims.UserID
	}

	rows, err := db.MobileDeviceSettingsUpdate(userID, userDeviceID, notifyEnabled, active)
	if err != nil {
		logger.Errorf("could not retrieve db results err: %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	defer rows.Close()

	returnQueryResults(rows, w, r)
}

// MobileTagedValidators godoc
// @Summary Get all your tagged validators
// @Tags User
// @Produce json
// @Success 200 {object} types.ApiResponse{data=[]types.MinimalTaggedValidators}
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/user/validator/saved [get]
func MobileTagedValidators(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)

	claims := getAuthClaims(r)

	filter := db.WatchlistFilter{
		UserId:         claims.UserID,
		Validators:     nil,
		Tag:            types.ValidatorTagsWatchlist,
		JoinValidators: true,
		Network:        utils.GetNetwork(),
	}

	validators, err2 := db.GetTaggedValidators(filter)
	if err2 != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	data := make([]interface{}, len(validators))
	for i, v := range validators {
		temp := types.MinimalTaggedValidators{}
		temp.PubKey = fmt.Sprintf("0x%v", hex.EncodeToString(v.ValidatorPublickey))
		temp.Index = v.Validator.Index
		data[i] = temp
	}

	SendOKResponse(j, r.URL.String(), data)
}

func parseUintWithDefault(input string, defaultValue uint64) uint64 {
	result, error := strconv.ParseUint(input, 10, 64)
	if error != nil {
		return defaultValue
	}
	return result
}

// ClientStats godoc
// @Summary Get your client submitted stats
// @Tags User
// @Produce json
// @Param offset path int false "Data offset, default 0" default(0)
// @Param limit path int false "Data limit, default 180 (~3h)." default(180)
// @Success 200 {object} types.ApiResponse{data=[]types.StatsDataStruct}
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/user/stats/{offset}/{limit} [get]
func ClientStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)
	claims := getAuthClaims(r)

	maxStats := getUserPremium(r).MaxStats

	vars := mux.Vars(r)
	offset := parseUintWithDefault(vars["offset"], 0)
	limit := parseUintWithDefault(vars["limit"], 180)
	timeframe := offset + limit
	if timeframe > maxStats {
		limit = maxStats
		offset = 0
	}

	system, err := db.BigtableClient.GetMachineMetricsSystem(claims.UserID, int(limit), int(offset))
	if err != nil {
		logger.Errorf("sytem stat error : %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve system stats from db")
		return
	}

	validator, err := db.BigtableClient.GetMachineMetricsValidator(claims.UserID, int(limit), int(offset))
	if err != nil {
		logger.Errorf("validator stat error : %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve validator stats from db")
		return
	}

	node, err := db.BigtableClient.GetMachineMetricsNode(claims.UserID, int(limit), int(offset))
	if err != nil {
		logger.Errorf("node stat error : %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve beaconnode stats from db")
		return
	}

	data := &types.StatsDataStruct{
		Validator: validator,
		Node:      node,
		System:    system,
	}

	SendOKResponse(j, r.URL.String(), []interface{}{data})
}

// ClientStatsPost godoc
// @Summary Used in eth2 clients to submit stats to your beaconcha.in account. This data can be accessed by the app or the user stats api call.
// @Tags User
// @Produce json
// @Param apikey query string true "User API key, can be found on https://beaconcha.in/user/settings"
// @Param machine query string false "Name your device if you have multiple devices you want to monitor"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Router /api/v1/client/metrics [POST]
func ClientStatsPostNew(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	apiKey := q.Get("apikey")
	machine := q.Get("machine")

	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}

	clientStatsPost(w, r, apiKey, machine)
}

func ClientStatsPostOld(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	clientStatsPost(w, r, vars["apiKey"], vars["machine"])
}

func clientStatsPost(w http.ResponseWriter, r *http.Request, apiKey, machine string) {
	w.Header().Set("Content-Type", "application/json")

	if utils.Config.Frontend.DisableStatsInserts {
		SendBadRequestResponse(w, r.URL.String(), "service temporarily unavailable")
		return
	}

	userData, err := db.GetUserIdByApiKey(apiKey)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "no user found with api key")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warnf("error reading body | err: %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not read body")
		return
	}

	var jsonObjects []map[string]interface{}
	err = json.Unmarshal(body, &jsonObjects)
	if err != nil {
		var jsonObject map[string]interface{}
		err = json.Unmarshal(body, &jsonObject)
		if err != nil {
			logger.Warnf("Could not parse stats (meta stats) general | %v ", err)
			SendBadRequestResponse(w, r.URL.String(), "Invalid JSON format in request body")
			return
		}
		jsonObjects = []map[string]interface{}{jsonObject}
	}

	if len(jsonObjects) >= 10 {
		logger.Info("Max number of stat entries are 10", err)
		SendBadRequestResponse(w, r.URL.String(), "Max number of stat entries are 10")
		return
	}

	var rateLimitErrs = 0
	var result bool = false
	for i := 0; i < len(jsonObjects); i++ {
		err = insertStats(userData, machine, &jsonObjects[i], w, r)
		result = err == nil
		if err != nil {
			// ignore rate limit errors unless all are rate limit errors
			if strings.HasPrefix(err.Error(), "rate limit") {
				result = true
				rateLimitErrs++
				continue
			}
			break
		}
	}

	if rateLimitErrs >= len(jsonObjects) {
		sendErrorWithCodeResponse(w, r.URL.String(), "rate limit too many metric requests, max 1 per user per machine per process", 429)
		return
	}

	if result {
		OKResponse(w, r)
		return
	}
}

func insertStats(userData *types.UserWithPremium, machine string, body *map[string]interface{}, w http.ResponseWriter, r *http.Request) error {

	var parsedMeta *types.StatsMeta
	err := mapstructure.Decode(body, &parsedMeta)
	if err != nil {
		logger.Warnf("Could not parse stats (meta stats) | %v ", err)
		SendBadRequestResponse(w, r.URL.String(), "could not parse meta")
		return err
	}

	parsedMeta.Machine = machine

	if parsedMeta.Version > 2 || parsedMeta.Version <= 0 {
		SendBadRequestResponse(w, r.URL.String(), "this version is not supported")
		return fmt.Errorf("this version is not supported")
	}

	if parsedMeta.Process != "validator" && parsedMeta.Process != "beaconnode" && parsedMeta.Process != "slasher" && parsedMeta.Process != "system" {
		SendBadRequestResponse(w, r.URL.String(), "unknown process")
		return fmt.Errorf("unknown process")
	}

	maxNodes := GetUserPremiumByPackage(userData.Product.String).MaxNodes

	count, err := db.BigtableClient.GetMachineMetricsMachineCount(userData.ID)
	if err != nil {
		logger.Errorf("Could not get max machine count| %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not get machine count")
		return err
	}

	if count > maxNodes {
		sendErrorWithCodeResponse(w, r.URL.String(), "reached max machine count", 402)
		return fmt.Errorf("user has reached max machine count")
	}

	var data []byte
	if parsedMeta.Process == "system" {
		var parsedResponse *types.MachineMetricSystem
		err = DecodeMapStructure(body, &parsedResponse)
		if err != nil {
			logger.Warnf("Could not parse stats (system stats) | %v", err)
			SendBadRequestResponse(w, r.URL.String(), "could not parse system")
			return err
		}
		data, err = proto.Marshal(parsedResponse)
		if err != nil {
			logger.Errorf("Could not parse stats (system stats) | %v", err)
			SendBadRequestResponse(w, r.URL.String(), "could marshal system")
			return err
		}
	} else if parsedMeta.Process == "validator" {
		var parsedResponse *types.MachineMetricValidator
		err = DecodeMapStructure(body, &parsedResponse)
		if err != nil {
			logger.Warnf("Could not parse stats (validator stats) | %v", err)
			SendBadRequestResponse(w, r.URL.String(), "could marshal validator")
			return err
		}
		data, err = proto.Marshal(parsedResponse)
		if err != nil {
			logger.Errorf("Could not parse stats (validator stats) | %v", err)
			SendBadRequestResponse(w, r.URL.String(), "could marshal validator")
			return err
		}
	} else if parsedMeta.Process == "beaconnode" {
		var parsedResponse *types.MachineMetricNode
		err = DecodeMapStructure(body, &parsedResponse)
		if err != nil {
			logger.Warnf("Could not parse stats (beaconnode stats) | %v", err)
			SendBadRequestResponse(w, r.URL.String(), "could not parse beaconnode")
			return err
		}
		data, err = proto.Marshal(parsedResponse)
		if err != nil {
			logger.Errorf("Could not parse stats (beaconnode stats) | %v", err)
			SendBadRequestResponse(w, r.URL.String(), "could not parse beaconnode")
			return err
		}
	}

	err = db.BigtableClient.SaveMachineMetric(parsedMeta.Process, userData.ID, machine, data)
	if err != nil {
		if strings.HasPrefix(err.Error(), "rate limit") {
			return err
		}
		logger.Errorf("Could not store stats | %v", err)
		SendBadRequestResponse(w, r.URL.String(), fmt.Sprintf("could not store stats: %v", err))
		return err
	}
	return nil
}

// ApiWithdrawalCredentialsValidators godoc
// @Summary Get validator indexes and pubkeys of a withdrawal credential or eth1 address
// @Tags Validator
// @Description Returns the validator indexes and pubkeys of a withdrawal credential or eth1 address
// @Produce json
// @Param withdrawalCredentialsOrEth1address path string true "Provide a withdrawal credential or an eth1 address with an optional 0x prefix". It can also be a valid ENS name.
// @Param  limit query int false "Limit the number of results, maximum: 200" default(10)
// @Param offset query int false "Offset the number of results" default(0)
// @Success 200 {object} types.ApiResponse{data=[]types.ApiWithdrawalCredentialsResponse}
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/validator/withdrawalCredentials/{withdrawalCredentialsOrEth1address} [get]
func ApiWithdrawalCredentialsValidators(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	q := r.URL.Query()

	credentialsOrAddressString := ReplaceEnsNameWithAddress(vars["withdrawalCredentialsOrEth1address"])
	credentialsOrAddressString = strings.ToLower(credentialsOrAddressString)

	if !utils.IsValidEth1Address(credentialsOrAddressString) &&
		!utils.IsValidWithdrawalCredentials(credentialsOrAddressString) {
		SendBadRequestResponse(w, r.URL.String(), "invalid withdrawal credentials or eth1 address provided")
		return
	}

	credentialsOrAddress := common.FromHex(credentialsOrAddressString)

	credentials, err := utils.AddressToWithdrawalCredentials(credentialsOrAddress)
	if err != nil {
		// Input is not an address so it must already be withdrawal credentials
		credentials = credentialsOrAddress
	}

	limitQuery := q.Get("limit")
	offsetQuery := q.Get("offset")

	offset := parseUintWithDefault(offsetQuery, 0)
	limit := parseUintWithDefault(limitQuery, 10)

	// We set a max limit to limit the request call time.
	var maxLimit uint64 = utilMath.MaxU64(200, uint64(getUserPremium(r).MaxValidators))

	limit = utilMath.MinU64(limit, maxLimit)

	result := []struct {
		Index  uint64 `db:"validatorindex"`
		Pubkey []byte `db:"pubkey"`
	}{}

	err = db.ReaderDb.Select(&result, `
	SELECT
		validatorindex,
		pubkey
	FROM validators
	WHERE withdrawalcredentials = $1
	ORDER BY validatorindex ASC
	LIMIT $2
	OFFSET $3
	`, credentials, limit, offset)

	if err != nil {
		logger.Warnf("error retrieving validator data from db: %v", err)
		SendBadRequestResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	response := make([]*types.ApiWithdrawalCredentialsResponse, 0, len(result))
	for _, validator := range result {
		response = append(response, &types.ApiWithdrawalCredentialsResponse{
			Publickey:      fmt.Sprintf("%#x", validator.Pubkey),
			ValidatorIndex: validator.Index,
		})
	}

	SendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{response})
}

// ApiProposalLuck godoc
// @Summary Get the proposal luck of a validator or a list of validators
// @Tags Validator
// @Description Returns the proposal luck of a validator or a list of validators
// @Produce json
// @Param validators query string true "Provide a comma separated list of validator indices or pubkeys"
// @Success 200 {object} types.ApiResponse{data=[]types.ApiProposalLuckResponse}
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Router /api/v1/validators/proposalLuck [get]
func ApiProposalLuck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
	response := &types.ApiResponse{}
	response.Status = "OK"

	indices, pubkeys, err := parseValidatorsFromQueryString(q.Get("validators"), 100)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not parse validators")
		return
	}
	if len(pubkeys) > 0 {
		indicesFromPubkeys, err := resolveIndices(pubkeys)
		if err != nil {
			SendBadRequestResponse(w, r.URL.String(), "could not resolve pubkeys to indices")
			return
		}
		indices = append(indices, indicesFromPubkeys...)
	}

	if len(indices) == 0 {
		SendBadRequestResponse(w, r.URL.String(), "no validators provided")
		return
	}

	// dedup indices
	allKeys := make(map[uint64]bool)
	list := []uint64{}
	for _, item := range indices {
		if _, ok := allKeys[item]; !ok {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	indices = list
	data, err := getProposalLuckStats(indices)
	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "error processing request, please try again later")
		utils.LogError(err, "error retrieving data from db for proposal luck", 0, map[string]interface{}{"request": r.Method + " " + r.URL.String()})
	}

	response.Data = data
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		utils.LogError(err, "error serializing json data for API", 0, map[string]interface{}{"request": r.Method + " " + r.URL.String()})
	}
}

func getProposalLuckStats(indices []uint64) (*types.ApiProposalLuckResponse, error) {
	data := types.ApiProposalLuckResponse{}
	g := errgroup.Group{}

	var firstActivationEpoch uint64
	g.Go(func() error {
		return db.GetFirstActivationEpoch(indices, &firstActivationEpoch)
	})

	var slots []uint64
	g.Go(func() error {
		return db.ReaderDb.Select(&slots, `
			SELECT
				slot
			FROM blocks
			WHERE proposer = ANY($1)
			AND exec_block_number IS NOT NULL
			ORDER BY slot ASC`, pq.Array(indices))
	})

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	proposalLuck, proposalTimeFrame := getProposalLuck(slots, len(indices), firstActivationEpoch)
	if proposalLuck > 0 {
		data.ProposalLuck = &proposalLuck
		timeframeName := getProposalTimeframeName(proposalTimeFrame)
		data.TimeFrameName = &timeframeName
	}

	avgProposalInterval := getAvgSlotInterval(len(indices))
	data.AverageProposalInterval = avgProposalInterval

	var estimateLowerBoundSlot *uint64
	if len(slots) > 0 {
		estimateLowerBoundSlot = &slots[len(slots)-1]
	} else if len(indices) == 1 {
		activationSlot := firstActivationEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		estimateLowerBoundSlot = &activationSlot
	}

	if estimateLowerBoundSlot != nil {
		nextProposalEstimate := utils.SlotToTime(*estimateLowerBoundSlot + uint64(avgProposalInterval)).Unix()
		data.NextProposalEstimateTs = &nextProposalEstimate
	}
	return &data, nil
}

func DecodeMapStructure(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   output,
		TagName:  "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// TODO Replace app code to work with new income balance dashboard
// Meanwhile keep old code from Feb 2021 to be app compatible
func APIDashboardDataBalance(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	queryValidatorIndices, queryValidatorPubkeys, err := parseValidatorsFromQueryString(q.Get("validators"), 100)
	if err != nil || len(queryValidatorPubkeys) > 0 {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}
	if len(queryValidatorIndices) < 1 {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}
	// queryValidatorsArr := pq.Array(queryValidators)

	// get data from one week before latest epoch
	latestEpoch := services.LatestEpoch()
	oneWeekEpochs := uint64(3600 * 24 * 7 / float64(utils.Config.Chain.ClConfig.SecondsPerSlot*utils.Config.Chain.ClConfig.SlotsPerEpoch))
	queryOffsetEpoch := uint64(0)
	if latestEpoch > oneWeekEpochs {
		queryOffsetEpoch = latestEpoch - oneWeekEpochs
	}

	if len(queryValidatorIndices) == 0 {
		SendBadRequestResponse(w, r.URL.String(), "no or invalid validator indicies provided")
	}

	balances, err := db.BigtableClient.GetValidatorBalanceHistory(queryValidatorIndices, latestEpoch-queryOffsetEpoch, latestEpoch)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator balance history")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	dataMap := make(map[uint64]*types.DashboardValidatorBalanceHistory)

	for _, balanceHistory := range balances {
		for _, history := range balanceHistory {
			if dataMap[history.Epoch] == nil {
				dataMap[history.Epoch] = &types.DashboardValidatorBalanceHistory{}
			}
			dataMap[history.Epoch].Balance += history.Balance
			dataMap[history.Epoch].EffectiveBalance += history.EffectiveBalance
			dataMap[history.Epoch].Epoch = history.Epoch
			dataMap[history.Epoch].ValidatorCount++
		}
	}

	data := make([]*types.DashboardValidatorBalanceHistory, 0, len(dataMap))

	for _, e := range dataMap {
		data = append(data, e)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Epoch < data[j].Epoch
	})

	balanceHistoryChartData := make([][4]float64, len(data))
	clPrice := price.GetPrice(utils.Config.Frontend.ClCurrency, currency)
	for i, item := range data {
		balanceHistoryChartData[i][0] = float64(utils.EpochToTime(item.Epoch).Unix() * 1000)
		balanceHistoryChartData[i][1] = item.ValidatorCount
		balanceHistoryChartData[i][2] = float64(item.Balance) / 1e9 * clPrice
		balanceHistoryChartData[i][3] = float64(item.EffectiveBalance) / 1e9 * clPrice
	}

	err = json.NewEncoder(w).Encode(balanceHistoryChartData)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		return
	}
}

func getAuthClaims(r *http.Request) *utils.CustomClaims {
	middleWare := gorillacontext.Get(r, utils.MobileAuthorizedKey)
	if middleWare == nil {
		return utils.GetAuthorizationClaims(r)
	}

	claims := gorillacontext.Get(r, utils.ClaimsContextKey)
	if claims == nil {
		return nil
	}
	return claims.(*utils.CustomClaims)
}

// Saves the result of a query converted to JSON in the response writer.
// An arbitrary amount of functions adjustQueryEntriesFuncs can be added to adjust the JSON response.
func returnQueryResults(rows *sql.Rows, w http.ResponseWriter, r *http.Request, adjustQueryEntriesFuncs ...func(map[string]interface{}) error) {
	j := json.NewEncoder(w)
	data, err := utils.SqlRowsToJSON(rows)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not parse db results")
		return
	}

	err = adjustQueryResults(data, adjustQueryEntriesFuncs...)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not adjust query results")
		return
	}

	SendOKResponse(j, r.URL.String(), data)
}

// Saves the result of a query converted to JSON in the response writer as an array.
// An arbitrary amount of functions adjustQueryEntriesFuncs can be added to adjust the JSON response.
func returnQueryResultsAsArray(rows *sql.Rows, w http.ResponseWriter, r *http.Request, adjustQueryEntriesFuncs ...func(map[string]interface{}) error) {
	data, err := utils.SqlRowsToJSON(rows)

	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not parse db results")
		return
	}

	err = adjustQueryResults(data, adjustQueryEntriesFuncs...)
	if err != nil {
		SendBadRequestResponse(w, r.URL.String(), "could not adjust query results")
		return
	}

	response := &types.ApiResponse{
		Status: "OK",
		Data:   data,
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		sendServerErrorResponse(w, r.URL.String(), "could not serialize data results")
		logger.Errorf("error serializing json data for API %v route: %v", r.URL.String(), err)
	}
}

func adjustQueryResults(data []interface{}, adjustQueryEntriesFuncs ...func(map[string]interface{}) error) error {
	for _, dataEntry := range data {
		dataEntryMap, ok := dataEntry.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error type asserting query results as a map")
		} else {
			for _, f := range adjustQueryEntriesFuncs {
				if err := f(dataEntryMap); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func SendBadRequestResponse(w http.ResponseWriter, route, message string) {
	sendErrorWithCodeResponse(w, route, message, http.StatusBadRequest)
}

func sendServerErrorResponse(w http.ResponseWriter, route, message string) {
	sendErrorWithCodeResponse(w, route, message, http.StatusInternalServerError)
}

func sendErrorWithCodeResponse(w http.ResponseWriter, route, message string, errorcode int) {
	w.WriteHeader(errorcode)
	j := json.NewEncoder(w)
	response := &types.ApiResponse{}
	response.Status = "ERROR: " + message
	err := j.Encode(response)

	if err != nil {
		logger.Errorf("error serializing json error for API %v route: %v", route, err)
	}
}

func SendOKResponse(j *json.Encoder, route string, data []interface{}) {
	response := &types.ApiResponse{}
	response.Status = "OK"

	if len(data) == 1 {
		response.Data = data[0]
	} else {
		response.Data = data
	}
	err := j.Encode(response)

	if err != nil {
		logger.Errorf("error serializing json data for API %v route: %v", route, err)
	}
}

func parseApiValidatorParamToIndices(origParam string, limit int) (indices []uint64, err error) {
	var pubkeys pq.ByteaArray
	params := strings.Split(origParam, ",")
	if len(params) > limit {
		return nil, fmt.Errorf("only a maximum of %d query parameters are allowed", limit)
	}
	for _, param := range params {
		if strings.Contains(param, "0x") || len(param) == 96 {
			pubkey, err := hex.DecodeString(strings.Replace(param, "0x", "", -1))
			if err != nil {
				return nil, fmt.Errorf("invalid validator-parameter")
			}
			pubkeys = append(pubkeys, pubkey)
		} else {
			index, err := strconv.ParseUint(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid validator-parameter: %v", param)
			}
			if index < db.MaxSqlInteger {
				indices = append(indices, index)
			}
		}
	}

	var queryIndicesDeduped []uint64
	queryIndicesDeduped = append(queryIndicesDeduped, indices...)
	if len(pubkeys) != 0 {
		indicesFromPubkeys := []uint64{}
		err = db.ReaderDb.Select(&indicesFromPubkeys, "SELECT validatorindex FROM validators WHERE pubkey = ANY($1)", pubkeys)

		if err != nil {
			return nil, err
		}

		indices = append(indices, indicesFromPubkeys...)

		m := make(map[uint64]uint64)
		for _, x := range indices {
			m[x] = x
		}
		for x := range m {
			queryIndicesDeduped = append(queryIndicesDeduped, x)
		}
	}

	if len(queryIndicesDeduped) == 0 {
		return nil, fmt.Errorf("invalid validator argument, pubkey(s) did not resolve to a validator index")
	}

	return queryIndicesDeduped, nil
}

func parseApiValidatorParamToPubkeys(origParam string, limit int) (pubkeys pq.ByteaArray, err error) {
	var indices pq.Int64Array
	params := strings.Split(origParam, ",")
	if len(params) > limit {
		return nil, fmt.Errorf("only a maximum of 100 query parameters are allowed")
	}
	for _, param := range params {
		if strings.Contains(param, "0x") || len(param) == 96 {
			pubkey, err := hex.DecodeString(strings.Replace(param, "0x", "", -1))
			if err != nil {
				return nil, fmt.Errorf("invalid validator-parameter")
			}
			pubkeys = append(pubkeys, pubkey)
		} else {
			index, err := strconv.ParseUint(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid validator-parameter: %v", param)
			}
			indices = append(indices, int64(index))
		}
	}

	var queryIndicesDeduped pq.ByteaArray
	queryIndicesDeduped = append(queryIndicesDeduped, pubkeys...)
	if len(indices) != 0 {
		var pubkeysFromIndices pq.ByteaArray
		err = db.ReaderDb.Select(&pubkeysFromIndices, "SELECT pubkey FROM validators WHERE validatorindex = ANY($1)", indices)

		if err != nil {
			return nil, err
		}

		pubkeys = append(pubkeys, pubkeysFromIndices...)

		m := make(map[string][]byte)
		for _, x := range pubkeys {
			m[string(x)] = x
		}
		for _, x := range m {
			queryIndicesDeduped = append(queryIndicesDeduped, x)
		}
	}

	if len(queryIndicesDeduped) == 0 {
		return nil, fmt.Errorf("invalid validator argument, pubkey(s) did not resolve to a validator index")
	}

	return queryIndicesDeduped, nil
}
