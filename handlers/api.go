package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// @title Beaconcha.in ETH2 API
// @version 1.0
// @description High performance API for querying information from the Ethereum 2.0 beacon chain
// @description The API is currently free to use. A fair use policy applies. Calls are rate limited to
// @description 10 requests / 1 minute / IP. All API results are cached for 1 minute.
// @description If you required a higher usage plan please get in touch with us at support@beaconcha.in.

// ApiHealthz godoc
// @Summary Health of the explorer
// @Tags Health
// @Description Health endpoint for montitoring if the explorer is in sync
// @Produce  text/plain
// @Success 200 {object} string
// @Router /api/healthz [get]
func ApiHealthz(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	lastEpoch, err := db.GetLatestEpoch()

	if err != nil {
		http.Error(w, "Internal server error: could not retrieve latest epoch from the db", 503)
		return
	}

	epochTime := utils.EpochToTime(lastEpoch)
	if epochTime.Before(time.Now().Add(time.Minute * -13)) {
		http.Error(w, "Internal server error: last epoch in db is more than 13 minutes old", 503)
		return
	}

	fmt.Fprintf(w, "OK. Last epoch is from %v ago", time.Since(epochTime))
}

// ApiEpoch godoc
// @Summary Get epoch by number
// @Tags Epoch
// @Description Returns information for a specified epoch by the epoch number or the latest epoch
// @Produce  json
// @Param  epoch path string true "Epoch number or the string latest"
// @Success 200 {object} string
// @Router /api/v1/epoch/{epoch} [get]
func ApiEpoch(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	epoch, err := strconv.ParseInt(vars["epoch"], 10, 64)
	if err != nil && vars["epoch"] != "latest" {
		sendErrorResponse(j, r.URL.String(), "invalid epoch provided")
		return
	}

	if vars["epoch"] == "latest" {
		epoch = int64(services.LatestEpoch())
	}

	rows, err := db.DB.Query(`SELECT *, 
		(SELECT COUNT(*) FROM blocks WHERE epoch = $1 AND status = '0') as scheduledblocks,
		(SELECT COUNT(*) FROM blocks WHERE epoch = $1 AND status = '1') as proposedblocks,
		(SELECT COUNT(*) FROM blocks WHERE epoch = $1 AND status = '2') as missedblocks,
		(SELECT COUNT(*) FROM blocks WHERE epoch = $1 AND status = '3') as orphanedblocks
		FROM epochs WHERE epoch = $1`, epoch)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiEpochBlocks godoc
// @Summary Get epoch blocks by epoch number
// @Tags Epoch
// @Description Returns all blocks for a specified epoch
// @Produce  json
// @Param  epoch path string true "Epoch number or the string latest"
// @Success 200 {object} string
// @Router /api/v1/epoch/{epoch}/blocks [get]
func ApiEpochBlocks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	epoch, err := strconv.ParseInt(vars["epoch"], 10, 64)
	if err != nil && vars["epoch"] != "latest" {
		sendErrorResponse(j, r.URL.String(), "invalid epoch provided")
		return
	}

	if vars["epoch"] == "latest" {
		epoch = int64(services.LatestEpoch())
	}

	rows, err := db.DB.Query("SELECT * FROM blocks WHERE epoch = $1 ORDER BY slot", epoch)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiBlock godoc
// @Summary Get block
// @Tags Block
// @Description Returns a block by its slot or root hash
// @Produce  json
// @Param  slotOrHash path string true "Block slot or root hash or the string latest"
// @Success 200 {object} string
// @Router /api/v1/block/{slotOrHash} [get]
func ApiBlock(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	slotOrHash := strings.Replace(vars["slotOrHash"], "0x", "", -1)
	blockSlot := int64(-1)
	blockRootHash, err := hex.DecodeString(slotOrHash)
	if err != nil || len(slotOrHash) != 64 {
		blockRootHash = []byte{}
		blockSlot, err = strconv.ParseInt(vars["slotOrHash"], 10, 64)
	}
	if slotOrHash == "latest" {
		blockSlot = int64(services.LatestSlot())
	}

	rows, err := db.DB.Query("SELECT * FROM blocks WHERE slot = $1 OR blockroot = $2", blockSlot, blockRootHash)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiBlockAttestations godoc
// @Summary Get the attestations included in a specific block
// @Tags Block
// @Description Returns the attestations included in a specific block
// @Produce  json
// @Param  slot path string true "Block slot"
// @Success 200 {object} string
// @Router /api/v1/block/{slot}/attestations [get]
func ApiBlockAttestations(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "invalid block slot provided")
		return
	}

	var root []byte
	err = db.DB.Get(&root, "SELECT blockroot FROM blocks WHERE slot = $1 AND status = '1'", slot)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), fmt.Sprintf("no block available at slot %v", slot))
		return
	}

	rows, err := db.DB.Query("SELECT * FROM blocks_attestations WHERE beaconblockroot = $1 ORDER BY block_index DESC", root)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiBlockDeposits godoc
// @Summary Get the deposits included in a specific block
// @Tags Block
// @Description Returns the deposits included in a specific block
// @Produce  json
// @Param  slot path string true "Block slot"
// @Success 200 {object} string
// @Router /api/v1/block/{slot}/deposits [get]
func ApiBlockDeposits(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.DB.Query("SELECT * FROM blocks_deposits WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiBlockAttesterSlashings godoc
// @Summary Get the attester slashings included in a specific block
// @Tags Block
// @Description Returns the attester slashings included in a specific block
// @Produce  json
// @Param  slot path string true "Block slot"
// @Success 200 {object} string
// @Router /api/v1/block/{slot}/attesterslashings [get]
func ApiBlockAttesterSlashings(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.DB.Query("SELECT * FROM blocks_attesterslashings WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiBlockProposerSlashings godoc
// @Summary Get the proposer slashings included in a specific block
// @Tags Block
// @Description Returns the proposer slashings included in a specific block
// @Produce  json
// @Param  slot path string true "Block slot"
// @Success 200 {object} string
// @Router /api/v1/block/{slot}/proposerslashings [get]
func ApiBlockProposerSlashings(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.DB.Query("SELECT * FROM blocks_proposerslashings WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiBlockVoluntaryExits godoc
// @Summary Get the voluntary exits included in a specific block
// @Tags Block
// @Description Returns the voluntary exits included in a specific block
// @Produce  json
// @Param  slot path string true "Block slot"
// @Success 200 {object} string
// @Router /api/v1/block/{slot}/voluntaryexits [get]
func ApiBlockVoluntaryExits(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	slot, err := strconv.ParseInt(vars["slot"], 10, 64)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "invalid block slot provided")
		return
	}

	rows, err := db.DB.Query("SELECT * FROM blocks_voluntaryexits WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiEth1Deposit godoc
// @Summary Get an eth1 deposit by its eth1 transaction hash
// @Tags Eth1
// @Produce  json
// @Param  txhash path string true "Eth1 transaction hash"
// @Success 200 {object} string
// @Router /api/v1/eth1deposit/{txhash} [get]
func ApiEth1Deposit(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	eth1TxHash, err := hex.DecodeString(strings.Replace(vars["txhash"], "0x", "", -1))
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "invalid eth1 tx hash provided")
		return
	}

	rows, err := db.DB.Query("SELECT * FROM eth1_deposits WHERE tx_hash = $1", eth1TxHash)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidator godoc
// @Summary Get up to 100 validators by their index
// @Tags Validator
// @Produce  json
// @Param  index path string true "Up to 100 validator indices, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{index} [get]
func ApiValidator(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryStr := vars["index"]
	query := []string{}
	if strings.Contains(queryStr, ",") {
		query = strings.Split(queryStr, ",")
	} else {
		query = append(query, queryStr)
	}

	if len(query) > 100 {
		sendErrorResponse(j, r.URL.String(), "only a maximum of 100 query parameters are allowed")
		return
	}

	rows, err := db.DB.Query("SELECT * FROM validators WHERE validatorindex = ANY($1) ORDER BY validatorindex", pq.Array(query))
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorByEth1Address godoc
// @Summary Get all validators that belong to an eth1 address
// @Tags Validator
// @Produce  json
// @Param  eth1address path string true "Eth1 address from which the validator deposits were sent"
// @Success 200 {object} string
// @Router /api/v1/validator/eth1/{address} [get]
func ApiValidatorByEth1Address(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	eth1Address, err := hex.DecodeString(strings.Replace(vars["address"], "0x", "", -1))

	if err != nil {
		sendErrorResponse(j, r.URL.String(), "invalid eth1 address provided")
		return
	}

	rows, err := db.DB.Query("SELECT publickey, validatorindex, valid_signature FROM eth1_deposits LEFT JOIN validators ON eth1_deposits.publickey = validators.pubkey WHERE from_address = $1 ORDER BY validatorindex;", eth1Address)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidator godoc
// @Summary Get the balance history (last 100 epochs) of up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  index path string true "Up to 100 validator indices, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{index}/balancehistory [get]
func ApiValidatorBalanceHistory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryStr := vars["index"]
	query := []string{}
	if strings.Contains(queryStr, ",") {
		query = strings.Split(queryStr, ",")
	} else {
		query = append(query, queryStr)
	}

	if len(query) > 100 {
		sendErrorResponse(j, r.URL.String(), "only a maximum of 100 query parameters are allowed")
		return
	}

	rows, err := db.DB.Query("SELECT * FROM validator_balances WHERE validatorindex = ANY($1) ORDER BY validatorindex, epoch DESC LIMIT 100", pq.Array(query))
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorPerformance godoc
// @Summary Get the current performance of up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  index path string true "Up to 100 validator indices, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{index}/performance [get]
func ApiValidatorPerformance(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryStr := vars["index"]
	query := []string{}
	if strings.Contains(queryStr, ",") {
		query = strings.Split(queryStr, ",")
	} else {
		query = append(query, queryStr)
	}

	if len(query) > 100 {
		sendErrorResponse(j, r.URL.String(), "only a maximum of 100 query parameters are allowed")
		return
	}

	rows, err := db.DB.Query("SELECT * FROM validator_performance WHERE validatorindex = ANY($1) ORDER BY validatorindex", pq.Array(query))
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorLeaderboard godoc
// @Summary Get the current top 100 performing validators (using the income over the last 7 days)
// @Tags Validator
// @Produce  json
// @Success 200 {object} string
// @Router /api/v1/validator/leaderboard [get]
func ApiValidatorLeaderboard(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)

	rows, err := db.DB.Query(`
			SELECT 
				validator_performance.*
			FROM validator_performance 
			ORDER BY performance7d DESC LIMIT 100`)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorDeposits godoc
// @Summary Get all eth1 deposits for up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  index path string true "Up to 100 validator indices, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{index}/deposits [get]
func ApiValidatorDeposits(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryStr := vars["index"]
	query := []string{}
	if strings.Contains(queryStr, ",") {
		query = strings.Split(queryStr, ",")
	} else {
		query = append(query, queryStr)
	}

	if len(query) > 100 {
		sendErrorResponse(j, r.URL.String(), "only a maximum of 100 query parameters are allowed")
		return
	}

	rows, err := db.DB.Query("SELECT eth1_deposits.* FROM validators LEFT JOIN eth1_deposits ON validators.pubkey = eth1_deposits.publickey WHERE validatorindex = ANY($1)", pq.Array(query))
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorAttestations godoc
// @Summary Get all attestations during the last 100 epochs for up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  index path string true "Up to 100 validator indices, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{index}/attestations [get]
func ApiValidatorAttestations(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryStr := vars["index"]
	query := []string{}
	if strings.Contains(queryStr, ",") {
		query = strings.Split(queryStr, ",")
	} else {
		query = append(query, queryStr)
	}

	if len(query) > 100 {
		sendErrorResponse(j, r.URL.String(), "only a maximum of 100 query parameters are allowed")
		return
	}

	currentEpoch := services.LatestEpoch()
	rows, err := db.DB.Query("SELECT * FROM attestation_assignments WHERE validatorindex = ANY($1) AND epoch > $2 ORDER BY validatorindex, epoch desc LIMIT 100", pq.Array(query), currentEpoch-100)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorProposals godoc
// @Summary Get all proposed blocks during the last 100 epochs for up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  index path string true "Up to 100 validator indices, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{index}/proposals [get]
func ApiValidatorProposals(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryStr := vars["index"]
	query := []string{}
	if strings.Contains(queryStr, ",") {
		query = strings.Split(queryStr, ",")
	} else {
		query = append(query, queryStr)
	}

	if len(query) > 100 {
		sendErrorResponse(j, r.URL.String(), "only a maximum of 100 query parameters are allowed")
		return
	}

	currentEpoch := services.LatestEpoch()
	rows, err := db.DB.Query("SELECT * FROM blocks WHERE proposer = ANY($1) AND epoch > $2 ORDER BY proposer, epoch desc, slot desc LIMIT 100", pq.Array(query), currentEpoch-100)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiChart godoc
// @Summary Returns charts from the page https://beaconcha.in/charts as PNG
// @Tags Charts
// @Produce  json
// @Param  chart path string true "Chart name (see https://github.com/gobitfly/eth2-beaconchain-explorer/blob/master/services/charts_updater.go#L20 for all available names)"
// @Success 200 {object} string
// @Router /api/v1/chart/{chart} [get]
func ApiChart(w http.ResponseWriter, r *http.Request) {

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	chartName := vars["chart"]

	var image []byte
	err := db.DB.Get(&image, "SELECT image FROM chart_images WHERE name = $1", chartName)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "no data available for the requested chart")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	_, err = w.Write(image)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "error writing chart data")
		return
	}
}

func returnQueryResults(rows *sql.Rows, j *json.Encoder, r *http.Request) {
	data, err := utils.SqlRowsToJSON(rows)

	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results")
		return
	}

	sendOKResponse(j, r.URL.String(), data)
}
func sendErrorResponse(j *json.Encoder, route, message string) {
	response := &types.ApiResponse{}
	response.Status = "ERROR: " + message
	err := j.Encode(response)

	if err != nil {
		logger.Errorf("error serializing json error for API %v route: %v", route, err)
	}
	return
}

func sendOKResponse(j *json.Encoder, route string, data []interface{}) {
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
	return
}
