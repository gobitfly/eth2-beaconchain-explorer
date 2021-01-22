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

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/mssola/user_agent"
)

// @title Beaconcha.in ETH2 API
// @version 1.0
// @description High performance API for querying information from the Ethereum 2.0 beacon chain
// @description The API is currently free to use. A fair use policy applies. Calls are rate limited to
// @description 10 requests / 1 minute / IP. All API results are cached for 1 minute.
// @description If you required a higher usage plan please checkout https://beaconcha.in/pricing.
// @securitydefinitions.oauth2.accessCode OAuthAccessCode
// @tokenurl https://beaconcha.in/user/token
// @authorizationurl https://beaconcha.in/user/authorize
// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

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

	if 18446744073709551615 == utils.Config.Chain.GenesisTimestamp {
		fmt.Fprint(w, "OK. No GENESIS_TIMESTAMP defined yet")
		return
	}

	genesisTime := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)
	if genesisTime.After(time.Now()) {
		fmt.Fprintf(w, "OK. Genesis in %v (%v)", time.Until(genesisTime), genesisTime)
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

	rows, err := db.DB.Query("SELECT * FROM blocks_attestations WHERE block_slot = $1 ORDER BY block_index", slot)
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
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{indexOrPubkey} [get]
func ApiValidator(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"])
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.DB.Query("SELECT validatorindex, pubkey, withdrawableepoch, withdrawalcredentials, balance, effectivebalance, slashed, activationeligibilityepoch, activationepoch, exitepoch, lastattestationslot, name, status FROM validators WHERE validatorindex = ANY($1) OR pubkey = ANY($2) ORDER BY validatorindex", pq.Array(queryIndices), queryPubkeys)
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
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{indexOrPubkey}/balancehistory [get]
func ApiValidatorBalanceHistory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"])
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.DB.Query("SELECT validator_balances_p.* FROM validator_balances_p LEFT JOIN validators ON validators.validatorindex = validator_balances_p.validatorindex WHERE validators.validatorindex = ANY($1) OR validators.pubkey = ANY($2) ORDER BY validatorindex, week desc, epoch DESC LIMIT 100", pq.Array(queryIndices), queryPubkeys)
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
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{indexOrPubkey}/performance [get]
func ApiValidatorPerformance(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"])
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.DB.Query("SELECT validator_performance.* FROM validator_performance LEFT JOIN validators ON validators.validatorindex = validator_performance.validatorindex WHERE validator_performance.validatorindex = ANY($1) OR validators.pubkey = ANY($2) ORDER BY validatorindex", pq.Array(queryIndices), queryPubkeys)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorAttestationEfficiency godoc
// @Summary Get the current performance of up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  index path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{indexOrPubkey}/attestationefficiency [get]
func ApiValidatorAttestationEfficiency(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	epoch := int64(services.LatestEpoch()) - 100
	if epoch < 0 {
		epoch = 0
	}

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"])
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.DB.Query(`
	SELECT aa.validatorindex, validators.pubkey, COALESCE(
		AVG(1 + inclusionslot - COALESCE((
			SELECT MIN(slot)
			FROM blocks
			WHERE slot > aa.attesterslot AND blocks.status = '1'
		), 0)
	), 0)::float AS attestation_efficiency
	FROM attestation_assignments_p aa
	INNER JOIN blocks ON blocks.slot = aa.inclusionslot AND blocks.status <> '3'
	INNER JOIN validators ON validators.validatorindex = aa.validatorindex
	WHERE aa.week >= $1 / 1575 AND aa.epoch > $1 AND (validators.validatorindex = ANY($2) OR validators.pubkey = ANY($3)) AND aa.inclusionslot > 0
	GROUP BY aa.validatorindex, validators.pubkey
	ORDER BY aa.validatorindex
	`, epoch, pq.Array(queryIndices), pq.Array(queryPubkeys))
	if err != nil {
		logger.Error(err)
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
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{indexOrPubkey}/deposits [get]
func ApiValidatorDeposits(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"])
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.DB.Query("SELECT eth1_deposits.* FROM eth1_deposits LEFT JOIN validators ON validators.pubkey = eth1_deposits.publickey WHERE validators.validatorindex = ANY($1) or eth1_deposits.publickey = ANY($2)", pq.Array(queryIndices), queryPubkeys)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorAttestations godoc
// @Summary Get all attestations during the last 10 epochs for up to 100 validators
// @Tags Validator
// @Produce  json
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{indexOrPubkey}/attestations [get]
func ApiValidatorAttestations(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"])
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.DB.Query("SELECT attestation_assignments_p.* FROM attestation_assignments_p LEFT JOIN validators ON validators.validatorindex = attestation_assignments_p.validatorindex WHERE (validators.validatorindex = ANY($1) OR validators.pubkey = ANY($2)) AND week >= $3 / 1575 AND epoch > $3 ORDER BY validatorindex, epoch desc LIMIT 100", pq.Array(queryIndices), queryPubkeys, services.LatestEpoch()-10)
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
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{indexOrPubkey}/proposals [get]
func ApiValidatorProposals(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"])
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.DB.Query("SELECT blocks.* FROM blocks LEFT JOIN validators on validators.validatorindex = blocks.proposer WHERE (proposer = ANY($1) OR validators.pubkey = ANY($2)) AND epoch > $3 ORDER BY proposer, epoch desc, slot desc LIMIT 100", pq.Array(queryIndices), queryPubkeys, services.LatestEpoch()-100)
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
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.InvalidGrant, "grant type must be authroization_code or refresh_token")
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

	// Create access token
	token, expiresIn, err := utils.CreateAccessToken(codeAuthData.UserID, codeAuthData.AppID, deviceID)
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
		logger.Errorf("Error refreshtoken check: %v | %v | %v", unsafeClaims.UserID, refreshTokenHashed, err)
		w.WriteHeader(http.StatusUnauthorized)
		utils.SendOAuthErrorResponse(j, r.URL.String(), utils.UnauthorizedClient, "invalid token credentials")
		return
	}

	// Create access token
	token, expiresIn, err := utils.CreateAccessToken(userID, unsafeClaims.AppID, unsafeClaims.DeviceID)
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
	j := json.NewEncoder(w)

	notifyToken := FormValueOrJSON(r, "token")

	claims := getAuthClaims(r)

	err2 := db.MobileNotificatonTokenUpdate(claims.UserID, claims.DeviceID, notifyToken)
	if err2 != nil {
		sendErrorResponse(j, r.URL.String(), "Can not save notify token")
		return
	}

	OKResponse(w, r)
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
	j := json.NewEncoder(w)

	claims := getAuthClaims(r)

	rows, err := db.MobileDeviceSettingsSelect(claims.UserID, claims.DeviceID)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}

	defer rows.Close()

	returnQueryResults(rows, j, r)
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
	j := json.NewEncoder(w)

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
			sendErrorResponse(j, r.URL.String(), "could not parse id")
			return
		}
		userDeviceID = temp
		sessionUser := getUser(w, r)
		if !sessionUser.Authenticated {
			sendErrorResponse(j, r.URL.String(), "not authenticated")
		}
		userID = sessionUser.UserID
	} else {
		userDeviceID = claims.DeviceID
		userID = claims.UserID
	}

	rows, err2 := db.MobileDeviceSettingsUpdate(userID, userDeviceID, notifyEnabled, active)
	if err2 != nil {
		logger.Errorf("could not retrieve db results err: %v", err2)
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}

	defer rows.Close()

	returnQueryResults(rows, j, r)
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
	}

	validators, err2 := db.GetTaggedValidators(filter)
	if err2 != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}

	data := make([]interface{}, len(validators))
	for i, v := range validators {
		temp := types.MinimalTaggedValidators{}
		temp.PubKey = fmt.Sprintf("0x%v", hex.EncodeToString(v.PublicKey))
		temp.Index = v.Index
		data[i] = temp
	}

	sendOKResponse(j, r.URL.String(), data)
}

func getAuthClaims(r *http.Request) *utils.CustomClaims {
	claims := context.Get(r, utils.ClaimsContextKey)
	if claims == nil {
		return nil
	}
	return claims.(*utils.CustomClaims)
}

func returnQueryResults(rows *sql.Rows, j *json.Encoder, r *http.Request) {
	data, err := utils.SqlRowsToJSON(rows)

	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results")
		return
	}

	sendOKResponse(j, r.URL.String(), data)
}

// SendErrorResponse exposes sendErrorResponse
func SendErrorResponse(j *json.Encoder, route, message string) {
	sendErrorResponse(j, route, message)
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

// SendOKResponse exposes sendOKResponse
func SendOKResponse(j *json.Encoder, route string, data []interface{}) {
	sendOKResponse(j, route, data)
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

func parseApiValidatorParam(origParam string) (indices []uint64, pubkeys pq.ByteaArray, err error) {
	params := strings.Split(origParam, ",")
	if len(params) > 100 {
		return nil, nil, fmt.Errorf("only a maximum of 100 query parameters are allowed")
	}
	for _, param := range params {
		if strings.Contains(param, "0x") || len(param) == 96 {
			pubkey, err := hex.DecodeString(strings.Replace(param, "0x", "", -1))
			if err != nil {
				return nil, nil, fmt.Errorf("invalid validator-parameter")
			}
			pubkeys = append(pubkeys, pubkey)
		} else {
			index, err := strconv.ParseUint(param, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid validator-parameter: %v", param)
			}
			indices = append(indices, index)
		}
	}
	return indices, pubkeys, nil
}
