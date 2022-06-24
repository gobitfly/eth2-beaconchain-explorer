package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/exporter"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	gorillacontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
	"github.com/mssola/user_agent"
	"golang.org/x/sync/errgroup"
)

// @title Beaconcha.in ETH2 API
// @version 1.0
// @description High performance API for querying information about the Ethereum 2.0 beacon chain
// @description The API is currently free to use. A fair use policy applies. Calls are rate limited to
// @description 10 requests / 1 minute / IP. All API results are cached for 1 minute.
// @description If you required a higher usage plan please checkout https://beaconcha.in/pricing.
// @description The API key can be provided in the Header or as a query string parameter.
// @description
// @description Key as a query string parameter: `curl https://beaconcha.in/api/v1/block/1?apikey=<your_key>`
// @description
// @description Key in a request header:  `curl -H 'apikey: <your_key>' https://beaconcha.in/api/v1/block/1`
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
		http.Error(w, "Internal server error: could not retrieve latest epoch from the db", http.StatusServiceUnavailable)
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
		http.Error(w, "Internal server error: last epoch in db is more than 13 minutes old", http.StatusServiceUnavailable)
		return
	}

	fmt.Fprintf(w, "OK. Last epoch is from %v ago", time.Since(epochTime))
}

// ApiHealthzLoadbalancer godoc
// @Summary Health of the explorer-api regarding having a healthy connection to the database
// @Tags Health
// @Description Health endpoint for montitoring if the explorer-api
// @Produce  text/plain
// @Success 200 {object} string
// @Router /api/healthz-loadbalancer [get]
func ApiHealthzLoadbalancer(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	lastEpoch, err := db.GetLatestEpoch()

	if err != nil {
		http.Error(w, "Internal server error: could not retrieve latest epoch from the db", http.StatusServiceUnavailable)
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

	fmt.Fprintf(w, "OK. Last epoch is from %v ago", time.Since(utils.EpochToTime(lastEpoch)))
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

	rows, err := db.ReaderDb.Query(`SELECT *, 
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

	rows, err := db.ReaderDb.Query("SELECT * FROM blocks WHERE epoch = $1 ORDER BY slot", epoch)
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

	rows, err := db.ReaderDb.Query("SELECT * FROM blocks WHERE slot = $1 OR blockroot = $2", blockSlot, blockRootHash)
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

	rows, err := db.ReaderDb.Query("SELECT * FROM blocks_attestations WHERE block_slot = $1 ORDER BY block_index", slot)
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

	rows, err := db.ReaderDb.Query("SELECT * FROM blocks_deposits WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorQueue godoc
// @Summary Get the current validator queue
// @Tags Block
// @Description Returns the current number of validators entering and exiting the beacon chain
// @Produce  json
// @Success 200 {object} string
// @Router /api/v1/validators/queue [get]
func ApiValidatorQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)

	rows, err := db.ReaderDb.Query("SELECT entering_validators_count as beaconchain_entering, exiting_validators_count as beaconchain_exiting FROM queue ORDER BY ts DESC LIMIT 1")
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

	rows, err := db.ReaderDb.Query("SELECT * FROM blocks_attesterslashings WHERE block_slot = $1 ORDER BY block_index DESC", slot)
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

	rows, err := db.ReaderDb.Query("SELECT * FROM blocks_proposerslashings WHERE block_slot = $1 ORDER BY block_index DESC", slot)
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

	rows, err := db.ReaderDb.Query("SELECT * FROM blocks_voluntaryexits WHERE block_slot = $1 ORDER BY block_index DESC", slot)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiBlockVoluntaryExits godoc
// @Summary Get the sync-committee for a sync-period
// @Tags SyncCommittee
// @Description Returns the sync-committee for a sync-period. Validators are sorted by sync-committee-index.
// @Produce json
// @Param period path string true "Period ('latest' for latest period or 'next' for next period in the future)"
// @Success 200 {object} string
// @Router /api/v1/sync_committee/{period} [get]
func ApiSyncCommittee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	period, err := strconv.ParseUint(vars["period"], 10, 64)
	if err != nil && vars["period"] != "latest" && vars["period"] != "next" {
		sendErrorResponse(j, r.URL.String(), "invalid epoch provided")
		return
	}

	if vars["period"] == "latest" {
		period = utils.SyncPeriodOfEpoch(services.LatestEpoch())
	} else if vars["period"] == "next" {
		period = utils.SyncPeriodOfEpoch(services.LatestEpoch()) + 1
	}

	rows, err := db.ReaderDb.Query(`SELECT period, period*$2 AS start_epoch, (period+1)*$2-1 AS end_epoch, ARRAY_AGG(validatorindex ORDER BY committeeindex) AS validators FROM sync_committees WHERE period = $1 GROUP BY period`, period, utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod)
	if err != nil {
		logger.WithError(err).WithField("url", r.URL.String()).Errorf("error querying db")
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

	rows, err := db.ReaderDb.Query("SELECT * FROM eth1_deposits WHERE tx_hash = $1", eth1TxHash)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiRocketpoolStats godoc
// @Summary Get global rocketpool network statistics
// @Tags Rocketpool
// @Produce  json
// @Success 200 {object} string
// @Router /api/v1/rocketpool/stats [get]
func ApiRocketpoolStats(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)

	stats, err := getRocketpoolStats()

	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results")
		return
	}

	sendOKResponse(j, r.URL.String(), stats)
}

// ApiRocketpoolValidators godoc
// @Summary Get rocketpool specific data for given validators
// @Tags Rocketpool
// @Param  indexOrPubkey path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Produce  json
// @Success 200 {object} string
// @Router /api/v1/rocketpool/validator/{indexOrPubkey} [get]
func ApiRocketpoolValidators(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}
	if len(queryPubkeys) > 0 {
		err := db.ReaderDb.Select(&queryIndices, "SELECT validatorindex FROM validators WHERE pubkey = ANY($1) ORDER BY validatorindex", queryPubkeys)
		if err != nil {
			logger.Errorf("dashboard could not resolve pubkeys to indices err: %v", err)
			sendErrorResponse(j, r.URL.String(), err.Error())
			return
		}
	}

	stats, err := getRocketpoolValidators(queryIndices)

	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results")
		return
	}

	sendOKResponse(j, r.URL.String(), stats)
}

/*
	Combined validator get, performance, attestationefficency, epoch, historic epoch and rpl
	Not public documented
*/
func ApiDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body | err: %v", err)
		sendErrorResponse(j, r.URL.String(), "could not read body")
		return
	}

	var getValidators bool = true
	var parsedBody types.DashboardRequest
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		getValidators = false
	}

	maxValidators := getUserPremium(r).MaxValidators

	epoch := int64(services.LatestEpoch())

	g, _ := errgroup.WithContext(context.Background())
	var validatorsData []interface{}
	var validatorEffectivenessData []interface{}
	var rocketpoolData []interface{}
	var rocketpoolStats []interface{}
	var currentEpochData []interface{}
	var olderEpochData []interface{}

	if getValidators {
		queryIndices, queryPubkeys, err := parseApiValidatorParam(parsedBody.IndicesOrPubKey, maxValidators)
		if err != nil {
			sendErrorResponse(j, r.URL.String(), err.Error())
			return
		}

		if len(queryPubkeys) > 0 {
			err := db.ReaderDb.Select(&queryIndices, "SELECT validatorindex FROM validators WHERE pubkey = ANY($1) ORDER BY validatorindex", queryPubkeys)
			if err != nil {
				logger.Errorf("dashboard could not resolve pubkeys to indices err: %v", err)
				sendErrorResponse(j, r.URL.String(), err.Error())
				return
			}
		}

		if len(queryIndices) > 0 {
			g.Go(func() error {
				validatorsData, err = validators(queryIndices)
				return err
			})

			g.Go(func() error {
				validatorEffectivenessData, err = validatorEffectiveness(epoch, queryIndices)
				return err
			})
			g.Go(func() error {
				rocketpoolData, err = getRocketpoolValidators(queryIndices)
				return err
			})

		}
	}

	g.Go(func() error {
		currentEpochData, err = getEpoch(epoch)
		return err
	})

	g.Go(func() error {
		olderEpochData, err = getEpoch(epoch - 10)
		return err
	})

	g.Go(func() error {
		rocketpoolStats, err = getRocketpoolStats()
		return err
	})

	err = g.Wait()
	if err != nil {
		logger.Errorf("dashboard %v", err)
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	data := &DashboardResponse{
		Validators:      validatorsData,
		Effectiveness:   validatorEffectivenessData,
		CurrentEpoch:    currentEpochData,
		OlderEpoch:      olderEpochData,
		Rocketpool:      rocketpoolData,
		RocketpoolStats: rocketpoolStats,
	}

	sendOKResponse(j, r.URL.String(), []interface{}{data})
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
			rplm.status_time       AS minipool_status_time,
			rpln.timezone_location AS node_timezone_location,
			rpln.rpl_stake         AS node_rpl_stake,
			rpln.max_rpl_stake     AS node_max_rpl_stake,
			rpln.min_rpl_stake     AS node_min_rpl_stake,
			validators.validatorindex AS index 
		FROM rocketpool_minipools rplm 
		LEFT JOIN validators validators ON rplm.pubkey = validators.pubkey 
		LEFT JOIN rocketpool_nodes rpln ON rplm.node_address = rpln.address
		WHERE validatorindex = ANY($1)`, pq.Array(queryIndices))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return utils.SqlRowsToJSON(rows)
}

func validators(queryIndices []uint64) ([]interface{}, error) {
	rows, err := db.ReaderDb.Query("SELECT validators.validatorindex, pubkey, withdrawableepoch, withdrawalcredentials, validators.balance, effectivebalance, slashed, activationeligibilityepoch, activationepoch, exitepoch, lastattestationslot, status, validator_names.name, performance1d, performance7d, performance31d, performance365d, rank7d FROM validators LEFT JOIN validator_performance ON validators.validatorindex = validator_performance.validatorindex LEFT JOIN validator_names ON validator_names.publickey = validators.pubkey WHERE validators.validatorindex = ANY($1) ORDER BY validators.validatorindex", pq.Array(queryIndices))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return utils.SqlRowsToJSON(rows)
}

func validatorEffectiveness(epoch int64, indices []uint64) ([]interface{}, error) {
	effectivenessEpochRange := epoch - 100
	if epoch < 0 {
		effectivenessEpochRange = 0
	}

	rows, err := db.ReaderDb.Query(`
	SELECT aa.validatorindex, validators.pubkey, TRUNC(COALESCE(
		AVG(1 + inclusionslot - COALESCE((
			SELECT MIN(slot)
			FROM blocks
			WHERE slot > aa.attesterslot AND blocks.status = '1'
		), 0)
	), 0)::decimal, 10)::float AS attestation_efficiency
	FROM attestation_assignments_p aa
	INNER JOIN blocks ON blocks.slot = aa.inclusionslot AND blocks.status <> '3'
	INNER JOIN validators ON validators.validatorindex = aa.validatorindex
	WHERE aa.week >= $1 / 1575 AND aa.epoch > $1 AND (validators.validatorindex = ANY($2)) AND aa.inclusionslot > 0
	GROUP BY aa.validatorindex, validators.pubkey
	ORDER BY aa.validatorindex
	`, effectivenessEpochRange, pq.Array(indices))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return utils.SqlRowsToJSON(rows)
}

type DashboardResponse struct {
	Validators      interface{} `json:"validators"`
	Effectiveness   interface{} `json:"effectiveness"`
	CurrentEpoch    interface{} `json:"currentEpoch"`
	OlderEpoch      interface{} `json:"olderEpoch"`
	Rocketpool      interface{} `json:"rocketpool_validators"`
	RocketpoolStats interface{} `json:"rocketpool_network_stats"`
}

func getEpoch(epoch int64) ([]interface{}, error) {
	rows, err := db.ReaderDb.Query(`SELECT attestationscount, attesterslashingscount, averagevalidatorbalance,
	blockscount, depositscount, eligibleether, epoch, finalized, TRUNC(globalparticipationrate::decimal, 10)::float as globalparticipationrate, proposerslashingscount,
	totalvalidatorbalance, validatorscount, voluntaryexitscount, votedether
	FROM epochs WHERE epoch = $1`, epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return utils.SqlRowsToJSON(rows)
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
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query("SELECT validatorindex, pubkey, withdrawableepoch, withdrawalcredentials, balance, effectivebalance, slashed, activationeligibilityepoch, activationepoch, exitepoch, lastattestationslot, status, validator_names.name FROM validators LEFT JOIN validator_names ON validator_names.publickey = validators.pubkey WHERE validatorindex = ANY($1) OR pubkey = ANY($2) ORDER BY validatorindex", pq.Array(queryIndices), queryPubkeys)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorDailyStats godoc
// @Summary Get the daily validator stats by the validator index
// @Tags Validator
// @Produce  json
// @Param  index path string true "Validator index"
// @Success 200 {object} string
// @Router /api/v1/validator/stats/{index} [get]
func ApiValidatorDailyStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	index := vars["index"]

	rows, err := db.ReaderDb.Query("SELECT * FROM validator_stats WHERE validatorindex = $1 ORDER BY day DESC", index)
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

	rows, err := db.ReaderDb.Query("SELECT publickey, validatorindex, valid_signature FROM eth1_deposits LEFT JOIN validators ON eth1_deposits.publickey = validators.pubkey WHERE from_address = $1 GROUP BY publickey, validatorindex, valid_signature ORDER BY validatorindex;", eth1Address)
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
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query("SELECT validator_balances_p.* FROM validator_balances_p LEFT JOIN validators ON validators.validatorindex = validator_balances_p.validatorindex WHERE week >= ((SELECT MAX(epoch) FROM epochs)-100)/(225*7) AND (validators.validatorindex = ANY($1) OR validators.pubkey = ANY($2)) ORDER BY epoch DESC, validatorindex LIMIT 100", pq.Array(queryIndices), queryPubkeys)
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
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query("SELECT validator_performance.* FROM validator_performance LEFT JOIN validators ON validators.validatorindex = validator_performance.validatorindex WHERE validator_performance.validatorindex = ANY($1) OR validators.pubkey = ANY($2) ORDER BY validatorindex", pq.Array(queryIndices), queryPubkeys)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiValidatorAttestationEffectiveness godoc
// @Summary Get the current attestation-effectiveness of up to 100 validators. 1 = all attestations are included in the next possible block, < 1 some attestations have been included after the next possible block.
// @Tags Validator
// @Produce  json
// @Param  index path string true "Up to 100 validator indicesOrPubkeys, comma separated"
// @Success 200 {object} string
// @Router /api/v1/validator/{indexOrPubkey}/attestationeffectiveness [get]
func ApiValidatorAttestationEffectiveness(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	epoch := int64(services.LatestEpoch()) - 100
	if epoch < 0 {
		epoch = 0
	}

	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query(`
		SELECT aa.validatorindex, validators.pubkey, COALESCE(
			1 / AVG(1 + inclusionslot - COALESCE((
				SELECT MIN(slot)
				FROM blocks
				WHERE slot > aa.attesterslot AND blocks.status = '1'
			), 0)
		), 0)::float AS attestation_effectiveness
		FROM attestation_assignments_p aa
		INNER JOIN blocks ON blocks.slot = aa.inclusionslot AND blocks.status <> '3'
		INNER JOIN validators ON validators.validatorindex = aa.validatorindex
		WHERE aa.week >= $1 / 1575 AND aa.epoch > $1 AND (validators.validatorindex = ANY($2) OR validators.pubkey = ANY($3)) AND aa.inclusionslot > 0
		GROUP BY aa.validatorindex, validators.pubkey
		ORDER BY aa.validatorindex`,
		epoch, pq.Array(queryIndices), queryPubkeys)

	if err != nil {
		logger.Error(err)
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

	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := getAttestationEfficiencyQuery(epoch, queryIndices, queryPubkeys)
	if err != nil {
		logger.Error(err)
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

func getAttestationEfficiencyQuery(epoch int64, queryIndices []uint64, queryPubkeys pq.ByteaArray) (*sql.Rows, error) {
	return db.ReaderDb.Query(`
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
	`, epoch, pq.Array(queryIndices), queryPubkeys)
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

	rows, err := db.ReaderDb.Query(`
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
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query("SELECT eth1_deposits.* FROM eth1_deposits LEFT JOIN validators ON validators.pubkey = eth1_deposits.publickey WHERE validators.validatorindex = ANY($1) or eth1_deposits.publickey = ANY($2)", pq.Array(queryIndices), queryPubkeys)
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
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query("SELECT attestation_assignments_p.* FROM attestation_assignments_p LEFT JOIN validators ON validators.validatorindex = attestation_assignments_p.validatorindex WHERE (validators.validatorindex = ANY($1) OR validators.pubkey = ANY($2)) AND week >= $3 / 1575 AND epoch > $3 ORDER BY validatorindex, epoch desc LIMIT 100", pq.Array(queryIndices), queryPubkeys, services.LatestEpoch()-10)
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
	maxValidators := getUserPremium(r).MaxValidators

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], maxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query("SELECT blocks.* FROM blocks LEFT JOIN validators on validators.validatorindex = blocks.proposer WHERE (proposer = ANY($1) OR validators.pubkey = ANY($2)) AND epoch > $3 ORDER BY proposer, epoch desc, slot desc LIMIT 100", pq.Array(queryIndices), queryPubkeys, services.LatestEpoch()-100)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, j, r)
}

// ApiGraffitiwall godoc
// @Summary Get all pixels that have been painted until now on the graffitiwall
// @Tags Graffitiwall
// @Produce  json
// @Success 200 {object} string
// @Router /api/v1/graffitiwall [get]
func ApiGraffitiwall(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	j := json.NewEncoder(w)

	rows, err := db.ReaderDb.Query("SELECT x, y, color, slot, validator FROM graffitiwall ORDER BY x, y LIMIT 1000000")
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
	err := db.ReaderDb.Get(&image, "SELECT image FROM chart_images WHERE name = $1", chartName)
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

	logger.Info("access token:", accessToken, "refreshToken: ", refreshToken)

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

func RegisterEthpoolSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)

	pkg := FormValueOrJSON(r, "package")
	ethpoolUserID := FormValueOrJSON(r, "user_id")
	signature := FormValueOrJSON(r, "signature")

	localSignature := hmacSign(fmt.Sprintf("ETHPOOL %v %v", pkg, ethpoolUserID))
	if signature != localSignature {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Errorf("signature missmatch %v | %v", signature, localSignature)
		sendErrorResponse(j, r.URL.String(), "Unauthorized: signature not valid")
		return
	}

	claims := getAuthClaims(r)

	subscriptionCount, err := db.GetAppSubscriptionCount(claims.UserID)
	if err != nil || subscriptionCount >= 5 {
		w.WriteHeader(http.StatusInternalServerError)
		sendErrorResponse(j, r.URL.String(), "reached max subscription limit")
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
		sendErrorResponse(j, r.URL.String(), "Can not save subscription data")
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
	j := json.NewEncoder(w)

	var parsedBase types.MobileSubscription
	err := json.Unmarshal(gorillacontext.Get(r, utils.JsonBodyNakedKey).([]byte), &parsedBase)

	if err != nil {
		logger.Errorf("error parsing body | err: %v %v", err)
		sendErrorResponse(j, r.URL.String(), "could not parse body")
		return
	}

	claims := getAuthClaims(r)

	subscriptionCount, err := db.GetAppSubscriptionCount(claims.UserID)
	if err != nil || subscriptionCount >= 5 {
		sendErrorResponse(j, r.URL.String(), "reached max subscription limit")
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
		sendErrorResponse(j, r.URL.String(), "Can not save subscription data")
		return
	}

	if parsedBase.Valid == false {
		logger.Errorf("receipt is not valid %v", validationResult.RejectReason)
		sendErrorResponse(j, r.URL.String(), "receipt is not valid")
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

func GetMobileWidgetStats(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)
	vars := mux.Vars(r)

	epoch := int64(services.LatestEpoch()) - 100
	if epoch < 0 {
		epoch = 0
	}
	prime := getUserPremium(r)
	if !prime.WidgetSupport {
		sendErrorResponse(j, r.URL.String(), "feature only available for premium users")
		return
	}

	queryIndices, queryPubkeys, err := parseApiValidatorParam(vars["indexOrPubkey"], prime.MaxValidators)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), err.Error())
		return
	}

	rows, err := db.ReaderDb.Query(
		"SELECT pubkey, effectivebalance, slashed, activationeligibilityepoch, "+
			"activationepoch, exitepoch, lastattestationslot, status, validator_performance.* FROM validators "+
			"LEFT JOIN validator_performance ON validators.validatorindex = validator_performance.validatorindex "+
			" WHERE validator_performance.validatorindex = ANY($1) OR pubkey = ANY($2) ORDER BY validator_performance.validatorindex",
		pq.Array(queryIndices), queryPubkeys,
	)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	efficiencyRows, err := getAttestationEfficiencyQuery(epoch, queryIndices, queryPubkeys)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve efficiency db results")
		return
	}
	defer efficiencyRows.Close()

	generalData, err := utils.SqlRowsToJSON(rows)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results")
		return
	}

	efficiencyData, err := utils.SqlRowsToJSON(efficiencyRows)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results")
		return
	}

	data := &types.WidgetResponse{
		Eff:       efficiencyData,
		Validator: generalData,
		Epoch:     epoch,
	}

	sendOKResponse(j, r.URL.String(), []interface{}{data})
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
		sessionUser := getUser(r)
		if !sessionUser.Authenticated {
			sendErrorResponse(j, r.URL.String(), "not authenticated")
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
		Network:        utils.GetNetwork(),
	}

	validators, err2 := db.GetTaggedValidators(filter)
	if err2 != nil {
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}

	data := make([]interface{}, len(validators))
	for i, v := range validators {
		temp := types.MinimalTaggedValidators{}
		temp.PubKey = fmt.Sprintf("0x%v", hex.EncodeToString(v.ValidatorPublickey))
		temp.Index = v.Validator.Index
		data[i] = temp
	}

	sendOKResponse(j, r.URL.String(), data)
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

	validator, err := db.GetStatsValidator(claims.UserID, limit, offset)
	if err != nil {
		logger.Errorf("validator stat error : %v", err)
		sendErrorResponse(j, r.URL.String(), "could not retrieve validator stats from db")
		return
	}

	node, err := db.GetStatsNode(claims.UserID, limit, offset)
	if err != nil {
		logger.Errorf("node stat error : %v", err)
		sendErrorResponse(j, r.URL.String(), "could not retrieve beaconnode stats from db")
		return
	}

	system, err := db.GetStatsSystem(claims.UserID, limit, offset)
	if err != nil {
		logger.Errorf("system stat error : %v", err)
		sendErrorResponse(j, r.URL.String(), "could not retrieve system stats from db")
		return
	}

	dataValidator, err := utils.SqlRowsToJSON(validator)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results for validator stats")
		return
	}

	dataNode, err := utils.SqlRowsToJSON(node)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results for beaconnode stats")
		return
	}

	dataSystem, err := utils.SqlRowsToJSON(system)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "could not parse db results for system stats")
		return
	}

	data := &types.StatsDataStruct{
		Validator: dataValidator,
		Node:      dataNode,
		System:    dataSystem,
	}

	sendOKResponse(j, r.URL.String(), []interface{}{data})
}

func ClientStatsPostNew(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	apiKey := q.Get("apikey")
	machine := q.Get("machine")

	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}

	clientStatsPost(w, r, apiKey, machine)
}

// ClientStatsPost godoc
// @Summary Used in eth2 clients to submit stats to your beaconcha.in account. This data can be accessed by the app or the user stats api call.
// @Tags User
// @Produce json
// @Param apiKey query string true "User API key, can be found on https://beaconcha.in/user/settings"
// @Param machine query string false "Name your device if you have multiple devices you want to monitor"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/client/metrics [POST]
func ClientStatsPostOld(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	clientStatsPost(w, r, vars["apiKey"], vars["machine"])
}

func clientStatsPost(w http.ResponseWriter, r *http.Request, apiKey, machine string) {
	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)

	if utils.Config.Frontend.DisableStatsInserts {
		sendErrorResponse(j, r.URL.String(), "service temporarily unavailable")
		return
	}

	userData, err := db.GetUserIdByApiKey(apiKey)
	if err != nil {
		sendErrorResponse(j, r.URL.String(), "no user found with api key")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body | err: %v", err)
		sendErrorResponse(j, r.URL.String(), "could not read body")
		return
	}

	var jsonObjects []map[string]interface{}
	err = json.Unmarshal(body, &jsonObjects)
	if err != nil {
		var jsonObject map[string]interface{}
		err = json.Unmarshal(body, &jsonObject)
		if err != nil {
			logger.Errorf("Could not parse stats (meta stats) general | %v ", err)
			sendErrorResponse(j, r.URL.String(), "capi rate limit reached, one process per machine per user each minute is allowed.")
			return
		}
		jsonObjects = []map[string]interface{}{jsonObject}
	}

	if len(jsonObjects) >= 10 {
		logger.Errorf("Max number of stat entries are 10", err)
		sendErrorResponse(j, r.URL.String(), "Max number of stat entries are 10")
		return
	}

	var result bool = false
	for i := 0; i < len(jsonObjects); i++ {
		result = insertStats(userData, machine, &jsonObjects[i], j, r)
		if !result {
			break
		}
	}

	if result {
		OKResponse(w, r)
		return
	}
}

func insertStats(userData *types.UserWithPremium, machine string, body *map[string]interface{}, j *json.Encoder, r *http.Request) bool {

	var parsedMeta *types.StatsMeta
	err := mapstructure.Decode(body, &parsedMeta)
	if err != nil {
		logger.Errorf("Could not parse stats (meta stats) | %v ", err)
		sendErrorResponse(j, r.URL.String(), "could not parse meta")
		return false
	}

	parsedMeta.Machine = machine

	if parsedMeta.Version > 2 || parsedMeta.Version <= 0 {
		sendErrorResponse(j, r.URL.String(), "this version is not supported")
		return false
	}

	if parsedMeta.Process != "validator" && parsedMeta.Process != "beaconnode" && parsedMeta.Process != "slasher" && parsedMeta.Process != "system" {
		sendErrorResponse(j, r.URL.String(), "unknown process")
		return false
	}

	maxNodes := GetUserPremiumByPackage(userData.Product.String).MaxNodes

	count, err := db.GetStatsMachineCount(userData.ID)
	if err != nil {
		logger.Errorf("Could not get max machine count| %v", err)
		sendErrorResponse(j, r.URL.String(), "could not get machine count")
		return false
	}

	if count > maxNodes {
		logger.Errorf("User has reached max machine count | %v", err)
		sendErrorResponse(j, r.URL.String(), "reached max machine count")
		return false
	}

	tx, err := db.NewTransaction()
	if err != nil {
		logger.Errorf("Could not transact | %v", err)
		sendErrorResponse(j, r.URL.String(), "could not store")
		return false
	}
	defer tx.Rollback()

	id, err := db.InsertStatsMeta(tx, userData.ID, parsedMeta)
	if err != nil {
		if strings.Contains(err.Error(), "no partition of relation") {
			db.CreateNewStatsMetaPartition()
			tx.Rollback()
			tx, err = db.NewTransaction()
			id, err = db.InsertStatsMeta(tx, userData.ID, parsedMeta)
		}
		if err != nil {
			if !strings.HasPrefix(err.Error(), "ERROR: duplicate key value violates unique constraint") {
				logger.Errorf("Could not store stats (meta stats) | %v", err)
			}
			sendErrorResponse(j, r.URL.String(), "could not store meta")
			return false
		}
	}

	// Special case for system
	if parsedMeta.Process == "system" {
		var parsedResponse *types.StatsSystem
		err = mapstructure.Decode(body, &parsedResponse)

		if err != nil {
			logger.Errorf("Could not parse stats (system stats) | %v", err)
			sendErrorResponse(j, r.URL.String(), "could not parse system")
			return false
		}
		_, err := db.InsertStatsSystem(
			tx,
			id,
			parsedResponse,
		)
		if err != nil {
			logger.Errorf("Could not store stats (system stats) | %v", err)
			sendErrorResponse(j, r.URL.String(), "could not store system")
			return false
		}

		err = tx.Commit()
		if err != nil {
			logger.Errorf("Could not store (tx commit) | %v", err)
			sendErrorResponse(j, r.URL.String(), "could not store")
			return false
		}
		return true
	}

	var parsedGeneral *types.StatsProcess
	err = mapstructure.Decode(body, &parsedGeneral)

	if err != nil {
		logger.Errorf("Could not parse stats (process stats) | %v", err)
		sendErrorResponse(j, r.URL.String(), "could not parse process")
		return false
	}

	processGeneralID, err := db.InsertStatsProcessGeneral(
		tx,
		id,
		parsedGeneral,
	)
	if err != nil {
		logger.Errorf("Could not store stats (global process stats) | %v", err)
		sendErrorResponse(j, r.URL.String(), "could not store global process")
		return false
	}

	if parsedMeta.Process == "validator" {
		var parsedValidator *types.StatsAdditionalsValidator
		err = mapstructure.Decode(body, &parsedValidator)

		if err != nil {
			logger.Errorf("Could not parse stats (validator stats) | %v", err)
			sendErrorResponse(j, r.URL.String(), "could not parse validator")
			return false
		}

		_, err := db.InsertStatsValidator(
			tx,
			processGeneralID,
			parsedValidator,
		)
		if err != nil {
			logger.Errorf("Could not store stats (validatorstats) | %v", err)
			sendErrorResponse(j, r.URL.String(), "could not store validator")
			return false
		}

	} else if parsedMeta.Process == "beaconnode" {
		var parsedNode *types.StatsAdditionalsBeaconnode
		err = mapstructure.Decode(body, &parsedNode)

		if err != nil {
			logger.Errorf("Could not parse stats (node stats) | %v", err)
			sendErrorResponse(j, r.URL.String(), "could not parse node")
			return false
		}

		_, err := db.InsertStatsBeaconnode(
			tx,
			processGeneralID,
			parsedNode,
		)
		if err != nil {
			logger.Errorf("Could not store stats (beaconnode) | %v", err)
			sendErrorResponse(j, r.URL.String(), "could not store beaconnode")
			return false
		}
	}

	err = tx.Commit()
	if err != nil {
		logger.Errorf("Could not store (tx commit) | %v", err)
		sendErrorResponse(j, r.URL.String(), "could not store")
		return false
	}
	return true
}

// TODO Replace app code to work with new income balance dashboard
// Meanwhile keep old code from Feb 2021 to be app compatible
func APIDashboardDataBalance(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"), 100)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	if len(queryValidators) < 1 {
		http.Error(w, "Invalid query", 400)
		return
	}
	queryValidatorsArr := pq.Array(queryValidators)

	// get data from one week before latest epoch
	latestEpoch := services.LatestEpoch()
	oneWeekEpochs := uint64(3600 * 24 * 7 / float64(utils.Config.Chain.Config.SecondsPerSlot*utils.Config.Chain.Config.SlotsPerEpoch))
	queryOffsetEpoch := uint64(0)
	if latestEpoch > oneWeekEpochs {
		queryOffsetEpoch = latestEpoch - oneWeekEpochs
	}

	query := `
		SELECT
			epoch,
			COALESCE(SUM(effectivebalance),0) AS effectivebalance,
			COALESCE(SUM(balance),0) AS balance,
			COUNT(*) AS validatorcount
		FROM validator_balances_p
		WHERE validatorindex = ANY($1) AND epoch > $2 AND week >= $2 / 1575
		GROUP BY epoch
		ORDER BY epoch ASC`

	data := []*types.DashboardValidatorBalanceHistory{}
	err = db.ReaderDb.Select(&data, query, queryValidatorsArr, queryOffsetEpoch)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator balance history")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	balanceHistoryChartData := make([][4]float64, len(data))
	for i, item := range data {
		balanceHistoryChartData[i][0] = float64(utils.EpochToTime(item.Epoch).Unix() * 1000)
		balanceHistoryChartData[i][1] = item.ValidatorCount
		balanceHistoryChartData[i][2] = float64(item.Balance) / 1e9 * price.GetEthPrice(currency)
		balanceHistoryChartData[i][3] = float64(item.EffectiveBalance) / 1e9 * price.GetEthPrice(currency)
	}

	err = json.NewEncoder(w).Encode(balanceHistoryChartData)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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

func parseApiValidatorParam(origParam string, limit int) (indices []uint64, pubkeys pq.ByteaArray, err error) {
	params := strings.Split(origParam, ",")
	if len(params) > limit {
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
