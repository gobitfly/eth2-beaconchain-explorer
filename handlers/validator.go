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
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/gorilla/mux"
	"github.com/juliangruber/go-intersect"
)

var validatorTemplate = template.Must(template.New("validator").Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files,
	"layout.html",
	"validator/validator.html",
	"validator/heading.html",
	"validator/tables.html",
	"validator/modals.html",
	"validator/overview.html",
	"validator/charts.html",
	"validator/countdown.html",

	"components/flashMessage.html",
	"components/rocket.html",
))
var validatorNotFoundTemplate = template.Must(template.New("validatornotfound").Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files, "layout.html", "validator/validatornotfound.html"))
var validatorEditFlash = "edit_validator_flash"

// Validator returns validator data using a go template
func Validator(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	//start := time.Now()

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	var index uint64
	var err error

	validatorPageData := types.ValidatorPageData{}

	stats := services.GetLatestStats()
	churnRate := stats.ValidatorChurnLimit
	if churnRate == nil {
		churnRate = new(uint64)
	}

	if *churnRate == 0 {
		*churnRate = 4
		logger.Warning("Churn rate not set in config using 4 as default please set minPerEpochChurnLimit")
	}
	validatorPageData.ChurnRate = *churnRate

	pendingCount := stats.PendingValidatorCount
	if pendingCount == nil {
		pendingCount = new(uint64)
	}

	validatorPageData.PendingCount = *pendingCount
	validatorPageData.InclusionDelay = int64((utils.Config.Chain.Config.Eth1FollowDistance*utils.Config.Chain.Config.SecondsPerEth1Block+utils.Config.Chain.Config.SecondsPerSlot*utils.Config.Chain.Config.SlotsPerEpoch*utils.Config.Chain.Config.EpochsPerEth1VotingPeriod)/3600) + 1

	data := InitPageData(w, r, "validators", "/validators", "")
	data.HeaderAd = true
	validatorPageData.NetworkStats = services.LatestIndexPageData()
	validatorPageData.User = data.User

	validatorPageData.FlashMessage, err = utils.GetFlash(w, r, validatorEditFlash)
	if err != nil {
		logger.Errorf("error retrieving flashes for validator %v: %v", vars["index"], err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Request came with a hash
	if strings.Contains(vars["index"], "0x") || len(vars["index"]) == 96 {
		pubKey, err := hex.DecodeString(strings.Replace(vars["index"], "0x", "", -1))
		if err != nil {
			logger.Errorf("error parsing validator public key %v: %v", vars["index"], err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			// the validator might only have a public key but no index yet
			var name string
			err = db.ReaderDb.Get(&name, `SELECT name FROM validator_names WHERE publickey = $1`, pubKey)
			if err != nil && err != sql.ErrNoRows {
				logger.Errorf("error getting validator-name from db for pubKey %v: %v", pubKey, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
				// err == sql.ErrNoRows -> unnamed
			} else {
				validatorPageData.Name = name
			}

			var pool string
			err = db.ReaderDb.Get(&pool, `SELECT pool FROM validator_pool WHERE publickey = $1`, pubKey)
			if err != nil && err != sql.ErrNoRows {
				logger.Errorf("error getting validator-pool from db for pubKey %v: %v", pubKey, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
				// err == sql.ErrNoRows -> (no pool set)
			} else {
				if validatorPageData.Name == "" {
					validatorPageData.Name = fmt.Sprintf("Pool: %s", pool)
				} else {
					validatorPageData.Name += fmt.Sprintf(" / Pool: %s", pool)
				}
			}
			deposits, err := db.GetValidatorDeposits(pubKey)
			if err != nil {
				logger.Errorf("error getting validator-deposits from db: %v", err)
			}
			validatorPageData.DepositsCount = uint64(len(deposits.Eth1Deposits))
			if err != nil || len(deposits.Eth1Deposits) == 0 {

				SetPageDataTitle(data, fmt.Sprintf("Validator %x", pubKey))
				data.Meta.Path = fmt.Sprintf("/validator/%v", index)
				err := validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)
				if err != nil {
					logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				return
			}

			// there is no validator-index but there are eth1-deposits for the publickey
			// which means the validator is in DEPOSITED state
			// in this state there is nothing to display but the eth1-deposits
			validatorPageData.Status = "deposited"
			validatorPageData.PublicKey = pubKey
			if deposits != nil && len(deposits.Eth1Deposits) > 0 {
				deposits.LastEth1DepositTs = deposits.Eth1Deposits[len(deposits.Eth1Deposits)-1].BlockTs
			}
			validatorPageData.Deposits = deposits

			latestDeposit := time.Now().Unix()
			if len(deposits.Eth1Deposits) > 1 {
				latestDeposit = deposits.Eth1Deposits[len(deposits.Eth1Deposits)-1].BlockTs
			} else if time.Unix(latestDeposit, 0).Before(utils.SlotToTime(0)) {
				latestDeposit = utils.SlotToTime(0).Unix()
				validatorPageData.InclusionDelay = 0
			}

			for _, deposit := range validatorPageData.Deposits.Eth1Deposits {
				if deposit.ValidSignature {
					validatorPageData.Eth1DepositAddress = deposit.FromAddress
					break
				}
			}

			sumValid := uint64(0)
			// check if a valid deposit exists
			for _, d := range deposits.Eth1Deposits {
				if d.ValidSignature {
					sumValid += d.Amount
				} else {
					validatorPageData.Status = "deposited_invalid"
				}
			}

			// enough deposited for the validator to be activated
			if sumValid >= 32e9 {
				validatorPageData.Status = "deposited_valid"
			}

			filter := db.WatchlistFilter{
				UserId:         data.User.UserID,
				Validators:     &pq.ByteaArray{validatorPageData.PublicKey},
				Tag:            types.ValidatorTagsWatchlist,
				JoinValidators: false,
				Network:        utils.GetNetwork(),
			}
			watchlist, err := db.GetTaggedValidators(filter)
			if err != nil {
				logger.Errorf("error getting tagged validators from db: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			validatorPageData.Watchlist = watchlist
			data.Data = validatorPageData
			if utils.IsApiRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				err = json.NewEncoder(w).Encode(data.Data)
			} else {
				err = validatorTemplate.ExecuteTemplate(w, "layout", data)
			}
			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			return
		}
	} else {
		// Request came with a validator index number
		index, err = strconv.ParseUint(vars["index"], 10, 64)
		if err != nil {
			http.Error(w, "Validator not found", 404)
			return
		}
	}

	// GetAvgOptimalInclusionDistance(index)

	SetPageDataTitle(data, fmt.Sprintf("Validator %v", index))
	data.Meta.Path = fmt.Sprintf("/validator/%v", index)

	// logger.Infof("retrieving data, elapsed: %v", time.Since(start))
	// start = time.Now()

	// we use MAX(validatorindex)+1 instead of COUNT(*) for querying the rank_count for performance-reasons
	err = db.ReaderDb.Get(&validatorPageData, `
		SELECT
			validators.pubkey,
			validators.validatorindex,
			validators.withdrawableepoch,
			validators.effectivebalance,
			validators.slashed,
			validators.activationeligibilityepoch,
			validators.activationepoch,
			validators.exitepoch,
			validators.lastattestationslot,
			COALESCE(validator_names.name, '') AS name,
			COALESCE(validator_pool.pool, '') AS pool,
			COALESCE(validators.balance, 0) AS balance,
			COALESCE(validator_performance.rank7d, 0) AS rank7d,
			COALESCE(validator_performance_count.total_count, 0) AS rank_count,
			validators.status,
			COALESCE(validators.balanceactivation, 0) AS balanceactivation,
			COALESCE(validators.balance7d, 0) AS balance7d,
			COALESCE(validators.balance31d, 0) AS balance31d,
			COALESCE((SELECT ARRAY_AGG(tag) FROM validator_tags WHERE publickey = validators.pubkey),'{}') AS tags
		FROM validators
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		LEFT JOIN validator_pool ON validators.pubkey = validator_pool.publickey
		LEFT JOIN validator_performance ON validators.validatorindex = validator_performance.validatorindex
		LEFT JOIN (SELECT MAX(validatorindex)+1 FROM validator_performance) validator_performance_count(total_count) ON true
		WHERE validators.validatorindex = $1`, index)

	if err == sql.ErrNoRows {
		err = validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	} else if err != nil {
		logger.Errorf("error getting validator for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if validatorPageData.Pool != "" {
		if validatorPageData.Name == "" {
			validatorPageData.Name = fmt.Sprintf("Pool: %s", validatorPageData.Pool)
		} else {
			validatorPageData.Name += fmt.Sprintf(" / Pool: %s", validatorPageData.Pool)
		}
	}

	if validatorPageData.Rank7d > 0 && validatorPageData.RankCount > 0 {
		validatorPageData.RankPercentage = float64(validatorPageData.Rank7d) / float64(validatorPageData.RankCount)
	}

	validatorPageData.Epoch = services.LatestEpoch()
	validatorPageData.Index = index
	if err != nil {
		logger.Errorf("error retrieving validator public key %v: %v", index, err)

		err := validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	filter := db.WatchlistFilter{
		UserId:         data.User.UserID,
		Validators:     &pq.ByteaArray{validatorPageData.PublicKey},
		Tag:            types.ValidatorTagsWatchlist,
		JoinValidators: false,
		Network:        utils.GetNetwork(),
	}

	watchlist, err := db.GetTaggedValidators(filter)
	if err != nil {
		logger.Errorf("error getting tagged validators from db: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	validatorPageData.Watchlist = watchlist

	deposits, err := db.GetValidatorDeposits(validatorPageData.PublicKey)
	if err != nil {
		logger.Errorf("error getting validator-deposits from db: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	validatorPageData.Deposits = deposits
	validatorPageData.DepositsCount = uint64(len(deposits.Eth1Deposits))

	for _, deposit := range validatorPageData.Deposits.Eth1Deposits {
		if deposit.ValidSignature {
			validatorPageData.Eth1DepositAddress = deposit.FromAddress
			break
		}
	}

	validatorPageData.ActivationEligibilityTs = utils.EpochToTime(validatorPageData.ActivationEligibilityEpoch)
	validatorPageData.ActivationTs = utils.EpochToTime(validatorPageData.ActivationEpoch)
	validatorPageData.ExitTs = utils.EpochToTime(validatorPageData.ExitEpoch)
	validatorPageData.WithdrawableTs = utils.EpochToTime(validatorPageData.WithdrawableEpoch)

	if validatorPageData.ActivationEpoch > 100_000_000 {
		queueAhead, err := db.GetQueueAheadOfValidator(validatorPageData.Index)
		if err != nil {
			logger.WithError(err).Warnf("failed to retrieve queue ahead of validator %v: %v", validatorPageData.ValidatorIndex, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		validatorPageData.QueuePosition = queueAhead + 1
		epochsToWait := queueAhead / *churnRate
		// calculate dequeue epoch
		estimatedActivationEpoch := validatorPageData.Epoch + epochsToWait + 1
		// add activation offset
		estimatedActivationEpoch += utils.Config.Chain.Config.MaxSeedLookahead + 1
		validatorPageData.EstimatedActivationEpoch = estimatedActivationEpoch
		estimatedDequeueTs := utils.EpochToTime(estimatedActivationEpoch)
		validatorPageData.EstimatedActivationTs = estimatedDequeueTs
	}

	proposals := []struct {
		Slot   uint64
		Status uint64
	}{}

	err = db.ReaderDb.Select(&proposals, "SELECT slot, status FROM blocks WHERE proposer = $1 ORDER BY slot", index)
	if err != nil {
		logger.Errorf("error retrieving block-proposals: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	validatorPageData.Proposals = make([][]uint64, len(proposals))
	for i, b := range proposals {
		validatorPageData.Proposals[i] = []uint64{
			uint64(utils.SlotToTime(b.Slot).Unix()),
			b.Status,
		}
		if b.Status == 0 {
			validatorPageData.ScheduledBlocksCount++
		} else if b.Status == 1 {
			validatorPageData.ProposedBlocksCount++
		} else if b.Status == 2 {
			validatorPageData.MissedBlocksCount++
		} else if b.Status == 3 {
			validatorPageData.OrphanedBlocksCount++
		}
	}

	validatorPageData.BlocksCount = uint64(len(proposals))
	if validatorPageData.BlocksCount > 0 {
		validatorPageData.UnmissedBlocksPercentage = float64(validatorPageData.BlocksCount-validatorPageData.MissedBlocksCount) / float64(len(proposals))
	} else {
		validatorPageData.UnmissedBlocksPercentage = 1.0
	}

	// logger.Infof("propoals data retrieved, elapsed: %v", time.Since(start))
	// start = time.Now()

	// Every validator is scheduled to issue an attestation once per epoch
	// Hence we can calculate the number of attestations using the current epoch and the activation epoch
	// Special care needs to be take for exited and pending validators
	validatorPageData.AttestationsCount = validatorPageData.Epoch - validatorPageData.ActivationEpoch + 1
	if validatorPageData.ActivationEpoch > validatorPageData.Epoch {
		validatorPageData.AttestationsCount = 0
	}
	if validatorPageData.ExitEpoch != 9223372036854775807 {
		validatorPageData.AttestationsCount = validatorPageData.ExitEpoch - validatorPageData.ActivationEpoch
	}

	var lastStatsDay uint64
	err = db.ReaderDb.Get(&lastStatsDay, "select coalesce(max(day),0) from validator_stats")
	if err != nil {
		logger.Errorf("error retrieving lastStatsDay: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if validatorPageData.AttestationsCount > 0 {
		// get attestationStats from validator_stats
		attestationStats := struct {
			MissedAttestations   uint64 `db:"missed_attestations"`
			OrphanedAttestations uint64 `db:"orphaned_attestations"`
		}{}
		if lastStatsDay > 0 {
			err = db.ReaderDb.Get(&attestationStats, "select coalesce(sum(missed_attestations), 0) as missed_attestations, coalesce(sum(orphaned_attestations), 0) as orphaned_attestations from validator_stats where validatorindex = $1", index)
			if err != nil {
				logger.Errorf("error retrieving validator attestationStats: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		// add attestationStats that are not yet in validator_stats TODO: Reimplement for bigtable
		validatorPageData.MissedAttestationsCount = attestationStats.MissedAttestations
		validatorPageData.OrphanedAttestationsCount = attestationStats.OrphanedAttestations
		validatorPageData.ExecutedAttestationsCount = validatorPageData.AttestationsCount - validatorPageData.MissedAttestationsCount - validatorPageData.OrphanedAttestationsCount
		validatorPageData.UnmissedAttestationsPercentage = float64(validatorPageData.AttestationsCount-validatorPageData.MissedAttestationsCount) / float64(validatorPageData.AttestationsCount)
	}

	// logger.Infof("attestations data retrieved, elapsed: %v", time.Since(start))
	// start = time.Now()

	validatorPageData.IncomeHistoryChartData, err = db.GetValidatorIncomeHistoryChart([]uint64{index}, currency)
	if err != nil {
		logger.Errorf("failed to generate income history chart data for validator view: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// logger.Infof("balance history retrieved, elapsed: %v", time.Since(start))
	// start = time.Now()

	earnings, err := GetValidatorEarnings([]uint64{index}, GetCurrency(r))
	if err != nil {
		logger.Errorf("error retrieving validator earnings: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	validatorPageData.Income1d = earnings.LastDay
	validatorPageData.Income7d = earnings.LastWeek
	validatorPageData.Income31d = earnings.LastMonth
	validatorPageData.Apr = earnings.APR

	// logger.Infof("income data retrieved, elapsed: %v", time.Since(start))
	// start = time.Now()

	if validatorPageData.Slashed {
		var slashingInfo struct {
			Slot    uint64
			Slasher uint64
			Reason  string
		}
		err = db.ReaderDb.Get(&slashingInfo,
			`select block_slot as slot, proposer as slasher, 'Attestation Violation' as reason
				from blocks_attesterslashings a1 left join blocks b1 on b1.slot = a1.block_slot
				where b1.status = '1' and $1 = ANY(a1.attestation1_indices) and $1 = ANY(a1.attestation2_indices)
			union all
			select block_slot as slot, proposer as slasher, 'Proposer Violation' as reason
				from blocks_proposerslashings a2 left join blocks b2 on b2.slot = a2.block_slot
				where b2.status = '1' and a2.proposerindex = $1
			limit 1`,
			index)
		if err != nil {
			logger.Errorf("error retrieving validator slashing info: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		validatorPageData.SlashedBy = slashingInfo.Slasher
		validatorPageData.SlashedAt = slashingInfo.Slot
		validatorPageData.SlashedFor = slashingInfo.Reason
	}

	err = db.ReaderDb.Get(&validatorPageData.SlashingsCount, `select COALESCE(sum(attesterslashingscount) + sum(proposerslashingscount), 0) from blocks where blocks.proposer = $1 and blocks.status = '1'`, index)
	if err != nil {
		logger.Errorf("error retrieving slashings-count: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// logger.Infof("slashing data retrieved, elapsed: %v", time.Since(start))
	// start = time.Now()

	eff, err := db.BigtableClient.GetValidatorEffectiveness([]uint64{index}, validatorPageData.Epoch-1)
	if err != nil {
		logger.Errorf("error retrieving validator effectiveness: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(eff) > 1 {
		logger.Errorf("error retrieving validator effectiveness: invalid length %v", len(eff))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if len(eff) == 0 {
		validatorPageData.AttestationInclusionEffectiveness = 0
	} else {
		validatorPageData.AttestationInclusionEffectiveness = eff[0].AttestationEfficiency
	}

	// logger.Infof("effectiveness data retrieved, elapsed: %v", time.Since(start))
	// start = time.Now()

	err = db.ReaderDb.Get(&validatorPageData.SyncCount, `SELECT count(*)*$1 FROM sync_committees WHERE validatorindex = $2`, utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod*utils.Config.Chain.Config.SlotsPerEpoch, index)
	if err != nil {
		logger.Errorf("error retrieving syncCount for validator %v: %v", index, err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	if validatorPageData.SyncCount > 0 {
		// get syncStats from validator_stats
		syncStats := struct {
			ParticipatedSync uint64 `db:"participated_sync"`
			MissedSync       uint64 `db:"missed_sync"`
			OrphanedSync     uint64 `db:"orphaned_sync"`
		}{}
		if lastStatsDay > 0 {
			err = db.ReaderDb.Get(&syncStats, "select coalesce(sum(participated_sync), 0) as participated_sync, coalesce(sum(missed_sync), 0) as missed_sync, coalesce(sum(orphaned_sync), 0) as orphaned_sync from validator_stats where validatorindex = $1", index)
			if err != nil {
				logger.Errorf("error retrieving validator syncStats: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		validatorPageData.UnmissedSyncPercentage = float64(validatorPageData.SyncCount-validatorPageData.MissedSyncCount) / float64(validatorPageData.SyncCount)
	}

	// add rocketpool-data if available
	validatorPageData.Rocketpool = &types.RocketpoolValidatorPageData{}
	err = db.ReaderDb.Get(validatorPageData.Rocketpool, `
		SELECT
			rplm.node_address      AS node_address,
			rplm.address           AS minipool_address,
			rplm.node_fee          AS minipool_node_fee,
			rplm.deposit_type      AS minipool_deposit_type,
			rplm.status            AS minipool_status,
			rplm.status_time       AS minipool_status_time,
			rplm.penalty_count     AS penalty_count,
			rpln.timezone_location AS node_timezone_location,
			rpln.rpl_stake         AS node_rpl_stake,
			rpln.max_rpl_stake     AS node_max_rpl_stake,
			rpln.min_rpl_stake     AS node_min_rpl_stake,
			rpln.rpl_cumulative_rewards     AS rpl_cumulative_rewards,
			rpln.claimed_smoothing_pool     AS claimed_smoothing_pool,
			rpln.unclaimed_smoothing_pool   AS unclaimed_smoothing_pool,
			rpln.unclaimed_rpl_rewards      AS unclaimed_rpl_rewards,
			COALESCE(rpln.smoothing_pool_opted_in, false)    AS smoothing_pool_opted_in 
		FROM validators
		LEFT JOIN rocketpool_minipools rplm ON rplm.pubkey = validators.pubkey
		LEFT JOIN rocketpool_nodes rpln ON rplm.node_address = rpln.address
		WHERE validators.validatorindex = $1`, index)
	if err == nil && (validatorPageData.Rocketpool.MinipoolAddress != nil || validatorPageData.Rocketpool.NodeAddress != nil) {
		validatorPageData.IsRocketpool = true
		if utils.Config.Chain.Config.DepositChainID == 1 {
			validatorPageData.Rocketpool.RocketscanUrl = "rocketscan.io"
		} else if utils.Config.Chain.Config.DepositChainID == 5 {
			validatorPageData.Rocketpool.RocketscanUrl = "prater.rocketscan.io"
		}
	} else if err != nil && err != sql.ErrNoRows {
		logger.Errorf("error getting rocketpool-data for validator for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	deposits, err := db.GetValidatorDeposits(pubkey)
	if err != nil {
		logger.Errorf("error getting validator-deposits for %v: %v", vars["pubkey"], err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(deposits)
	if err != nil {
		logger.Errorf("error encoding validator-deposits for %v: %v", vars["pubkey"], err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ValidatorAttestationInclusionEffectiveness returns a validator's effectiveness in json
func ValidatorAttestationInclusionEffectiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing validator index: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	eff, err := db.BigtableClient.GetValidatorEffectiveness([]uint64{index}, services.LatestEpoch()-1)
	if err != nil {
		logger.Errorf("error retrieving validator effectiveness: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type resp struct {
		Effectiveness float64 `json:"effectiveness"`
	}

	if len(eff) > 1 {
		logger.Errorf("error retrieving validator effectiveness: invalid length %v", len(eff))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if len(eff) == 0 {
		err = json.NewEncoder(w).Encode(resp{Effectiveness: 0})
		if err != nil {
			logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = json.NewEncoder(w).Encode(resp{Effectiveness: eff[0].AttestationEfficiency})
		if err != nil {
			logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

}

// ValidatorProposedBlocks returns a validator's proposed blocks in json
func ValidatorProposedBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing validator index: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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

	var totalCount uint64

	err = db.ReaderDb.Get(&totalCount, "SELECT COUNT(*) FROM blocks WHERE proposer = $1", index)
	if err != nil {
		logger.Errorf("error retrieving proposed blocks count: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "epoch",
		"2": "status",
		"5": "attestationscount",
		"6": "depositscount",
		"8": "voluntaryexitscount",
		"9": "graffiti",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "epoch"
	}
	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	var blocks []*types.IndexPageDataBlocks
	err = db.ReaderDb.Select(&blocks, `
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
		ORDER BY `+orderBy+` `+orderDir+`
		LIMIT $2 OFFSET $3`, index, length, start)

	if err != nil {
		logger.Errorf("error retrieving proposed blocks data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	start, err := strconv.ParseInt(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	length := 10

	epoch := services.LatestEpoch()

	ae := struct {
		ActivationEpoch uint64
		ExitEpoch       uint64
	}{}

	err = db.ReaderDb.Get(&ae, "SELECT activationepoch, exitepoch FROM validators WHERE validatorindex = $1", index)
	if err != nil {
		logger.Errorf("error retrieving attestations count: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalCount := epoch - ae.ActivationEpoch + 1
	if ae.ActivationEpoch > epoch {
		totalCount = 0
	}
	lastAttestationEpoch := epoch
	if ae.ExitEpoch != 9223372036854775807 {
		lastAttestationEpoch = ae.ExitEpoch
		totalCount = ae.ExitEpoch - ae.ActivationEpoch
	}

	tableData := [][]interface{}{}

	if totalCount > 0 {
		attestationData, err := db.BigtableClient.GetValidatorAttestationHistory([]uint64{index}, uint64(int64(lastAttestationEpoch)-start), int64(length))
		if err != nil {
			logger.Errorf("error retrieving validator attestations data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		tableData = make([][]interface{}, len(attestationData[index]))

		for i, history := range attestationData[index] {

			if history.Status == 0 && history.Epoch < epoch-1 {
				history.Status = 2
			}
			tableData[i] = []interface{}{
				utils.FormatEpoch(history.Epoch),
				utils.FormatBlockSlot(history.AttesterSlot),
				utils.FormatAttestationStatus(history.Status),
				utils.FormatTimestamp(utils.SlotToTime(history.AttesterSlot).Unix()),
				utils.FormatAttestationInclusionSlot(history.InclusionSlot),
				utils.FormatInclusionDelay(history.InclusionSlot, history.Delay),
			}
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var totalCount uint64
	err = db.ReaderDb.Get(&totalCount, `
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	var attesterSlashings []*types.ValidatorAttestationSlashing
	err = db.ReaderDb.Select(&attesterSlashings, `
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var proposerSlashings []*types.ValidatorProposerSlashing
	err = db.ReaderDb.Select(&proposerSlashings, `
		SELECT blocks.slot, blocks.epoch, blocks.proposer, blocks_proposerslashings.proposerindex 
		FROM blocks_proposerslashings 
		INNER JOIN blocks ON blocks.proposer = $1 AND blocks_proposerslashings.block_slot = blocks.slot`, index)
	if err != nil {
		logger.Errorf("error retrieving block proposer slashings data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, 0, len(attesterSlashings)+len(proposerSlashings))
	for _, b := range attesterSlashings {

		inter := intersect.Simple(b.Attestestation1Indices, b.Attestestation2Indices)
		slashedValidators := []uint64{}
		if len(inter) == 0 {
			logger.Warning("No intersection found for attestation violation")
		}
		for _, v := range inter {
			slashedValidators = append(slashedValidators, uint64(v.(int64)))
		}

		tableData = append(tableData, []interface{}{
			utils.FormatSlashedValidators(slashedValidators),
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func ValidatorSave(w http.ResponseWriter, r *http.Request) {

	pubkey := r.FormValue("pubkey")
	pubkey = strings.ToLower(pubkey)
	pubkey = strings.Replace(pubkey, "0x", "", -1)

	pubkeyDecoded, err := hex.DecodeString(pubkey)
	if err != nil {
		logger.Errorf("error parsing submitted pubkey %v: %v", pubkey, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	name := r.FormValue("name")
	if len(name) > 40 {
		name = name[:40]
	}

	applyNameToAll := r.FormValue("apply-to-all")

	signature := r.FormValue("signature")
	signatureWrapper := &types.MyCryptoSignature{}
	err = json.Unmarshal([]byte(signature), signatureWrapper)
	if err != nil {
		logger.Errorf("error decoding submitted signature %v: %v", signature, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	msgForHashing := "\x19Ethereum Signed Message:\n" + strconv.Itoa(len(signatureWrapper.Msg)) + signatureWrapper.Msg
	msgHash := crypto.Keccak256Hash([]byte(msgForHashing))

	signatureParsed, err := hex.DecodeString(strings.Replace(signatureWrapper.Sig, "0x", "", -1))
	if err != nil {
		logger.Errorf("error parsing submitted signature %v: %v", signatureWrapper.Sig, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	if len(signatureParsed) != 65 {
		logger.Errorf("signature must be 65 bytes long")
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	if signatureParsed[64] == 27 || signatureParsed[64] == 28 {
		signatureParsed[64] -= 27
	}

	recoveredPubkey, err := crypto.SigToPub(msgHash.Bytes(), signatureParsed)
	if err != nil {
		logger.Errorf("error recovering pubkey: %v", err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}
	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubkey)

	var depositedAddress string
	deposits, err := db.GetValidatorDeposits(pubkeyDecoded)
	if err != nil {
		logger.Errorf("error getting validator-deposits from db for signature verification: %v", err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
	}
	for _, deposit := range deposits.Eth1Deposits {
		if deposit.ValidSignature {
			depositedAddress = "0x" + fmt.Sprintf("%x", deposit.FromAddress)
			break
		}
	}

	if strings.EqualFold(depositedAddress, recoveredAddress.Hex()) {
		if applyNameToAll == "on" {
			res, err := db.WriterDb.Exec(`
				INSERT INTO validator_names (publickey, name)
				SELECT publickey, $1 as name
				FROM (SELECT DISTINCT publickey FROM eth1_deposits WHERE from_address = $2 AND valid_signature) a
				ON CONFLICT (publickey) DO UPDATE SET name = excluded.name`, name, recoveredAddress.Bytes())
			if err != nil {
				logger.Errorf("error saving validator name (apply to all): %x: %v: %v", pubkeyDecoded, name, err)
				utils.SetFlash(w, r, validatorEditFlash, "Error: Db error while updating validator names")
				http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
				return
			}

			rowsAffected, _ := res.RowsAffected()
			utils.SetFlash(w, r, validatorEditFlash, fmt.Sprintf("Your custom name has been saved for %v validator(s).", rowsAffected))
			http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		} else {
			_, err := db.WriterDb.Exec(`
				INSERT INTO validator_names (publickey, name) 
				VALUES($2, $1) 
				ON CONFLICT (publickey) DO UPDATE SET name = excluded.name`, name, pubkeyDecoded)
			if err != nil {
				logger.Errorf("error saving validator name: %x: %v: %v", pubkeyDecoded, name, err)
				utils.SetFlash(w, r, validatorEditFlash, "Error: Db error while updating validator name")
				http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
				return
			}

			utils.SetFlash(w, r, validatorEditFlash, "Your custom name has been saved.")
			http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		}

	} else {
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
	}

}

// ValidatorHistory returns a validators history in json
func ValidatorHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	currency := GetCurrency(r)

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing validator index: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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

	//length := 10

	var activationAndExitEpoch = struct {
		ActivationEpoch uint64 `db:"activationepoch"`
		ExitEpoch       uint64 `db:"exitepoch"`
	}{}
	err = db.ReaderDb.Get(&activationAndExitEpoch, "SELECT activationepoch, exitepoch FROM validators WHERE validatorindex = $1", index)
	if err != nil {
		logger.Errorf("error retrieving activationAndExitEpoch for validator-history: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalCount := uint64(0)

	// Every validator is scheduled to issue an attestation once per epoch
	// Hence we can calculate the number of attestations using the current epoch and the activation epoch
	// Special care needs to be take for exited and pending validators
	if activationAndExitEpoch.ExitEpoch != 9223372036854775807 {
		totalCount += activationAndExitEpoch.ExitEpoch - activationAndExitEpoch.ActivationEpoch
	} else {
		totalCount += services.LatestEpoch() - activationAndExitEpoch.ActivationEpoch + 1
	}

	if start > 90 {
		start = 90
	}

	currentEpoch := services.LatestEpoch() - 1

	lookBack := uint64(0)

	if currentEpoch >= 10 {
		lookBack = currentEpoch - 10
	}

	if lookBack >= start {
		lookBack = lookBack - start
	}

	var validatorHistory []*types.ValidatorHistory

	g := new(errgroup.Group)

	var balanceHistory map[uint64][]*types.ValidatorBalance
	g.Go(func() error {
		var err error
		balanceHistory, err = db.BigtableClient.GetValidatorBalanceHistory([]uint64{index}, currentEpoch-start, 12)
		if err != nil {
			logger.Errorf("error retrieving validator balance history from bigtable: %v", err)
			return err
		}
		return nil
	})

	var attestationHistory map[uint64][]*types.ValidatorAttestation
	g.Go(func() error {
		var err error
		attestationHistory, err = db.BigtableClient.GetValidatorAttestationHistory([]uint64{index}, currentEpoch-start, 12)
		if err != nil {
			logger.Errorf("error retrieving validator attestation history from bigtable: %v", err)
			return err
		}
		return nil
	})

	var proposalHistory map[uint64][]*types.ValidatorProposal
	g.Go(func() error {
		var err error
		proposalHistory, err = db.BigtableClient.GetValidatorProposalHistory([]uint64{index}, currentEpoch-start, 12)
		if err != nil {
			logger.Errorf("error retrieving validator proposal history from bigtable: %v", err)
			return err
		}
		return nil
	})
	err = g.Wait()

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	proposalMap := make(map[uint64]*types.ValidatorProposal)
	for _, proposal := range proposalHistory[index] {
		proposalMap[proposal.Slot/32] = &types.ValidatorProposal{
			Index:  index,
			Slot:   proposal.Slot,
			Status: proposal.Status,
		}
	}

	attestationsMap := make(map[uint64]*types.ValidatorAttestation)
	for _, attestation := range attestationHistory[index] {
		attestationsMap[attestation.Epoch] = &types.ValidatorAttestation{
			Index:          index,
			Epoch:          attestation.Epoch,
			AttesterSlot:   attestation.AttesterSlot,
			CommitteeIndex: 0,
			Status:         attestation.Status,
			InclusionSlot:  attestation.InclusionSlot,
			Delay:          attestation.Delay,
		}
	}

	for i := 0; i < len(balanceHistory[index])-2; i++ {
		balanceChange := int64(balanceHistory[index][i].Balance) - int64(balanceHistory[index][i+1].Balance)

		h := &types.ValidatorHistory{
			Epoch:          balanceHistory[index][i].Epoch,
			BalanceChange:  sql.NullInt64{Int64: balanceChange, Valid: true},
			AttesterSlot:   sql.NullInt64{Int64: 0, Valid: false},
			InclusionSlot:  sql.NullInt64{Int64: 0, Valid: false},
			ProposalStatus: sql.NullInt64{Int64: 0, Valid: false},
			ProposalSlot:   sql.NullInt64{Int64: 0, Valid: false},
		}

		if attestationsMap[balanceHistory[index][i].Epoch] != nil {
			h.AttesterSlot = sql.NullInt64{Int64: int64(attestationsMap[balanceHistory[index][i].Epoch].AttesterSlot), Valid: true}
			h.InclusionSlot = sql.NullInt64{Int64: int64(attestationsMap[balanceHistory[index][i].Epoch].InclusionSlot), Valid: true}
		}

		if proposalMap[balanceHistory[index][i].Epoch] != nil {
			h.ProposalStatus = sql.NullInt64{Int64: int64(proposalMap[balanceHistory[index][i].Epoch].Status), Valid: true}
			h.ProposalSlot = sql.NullInt64{Int64: int64(proposalMap[balanceHistory[index][i].Epoch].Slot), Valid: true}
		}

		validatorHistory = append(validatorHistory, h)
	}

	tableData := make([][]interface{}, 0, len(validatorHistory))
	for _, b := range validatorHistory {
		if b.AttesterSlot.Int64 == -1 && b.BalanceChange.Int64 < 0 {
			b.AttestationStatus = 4
		}

		if b.AttesterSlot.Int64 == -1 && b.BalanceChange.Int64 >= 0 {
			b.AttestationStatus = 5
		}

		if b.AttesterSlot.Int64 != -1 && b.AttesterSlot.Valid && utils.SlotToTime(uint64(b.AttesterSlot.Int64)).Before(time.Now().Add(time.Minute*-1)) && b.InclusionSlot.Int64 == 0 {
			b.AttestationStatus = 2
		}

		if b.InclusionSlot.Valid && b.InclusionSlot.Int64 != 0 && b.AttestationStatus == 0 {
			b.AttestationStatus = 1
		}

		events := utils.FormatAttestationStatusShort(b.AttestationStatus)

		if b.ProposalSlot.Valid {
			block := utils.FormatBlockStatusShort(uint64(b.ProposalStatus.Int64))
			events += " & " + block
		}

		if b.BalanceChange.Valid {
			tableData = append(tableData, []interface{}{
				utils.FormatEpoch(b.Epoch),
				utils.FormatBalanceChange(&b.BalanceChange.Int64, currency),
				template.HTML(events),
			})
		}
	}

	if totalCount > 100 {
		totalCount = 100
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

var validatorStatsTableTemplate = template.Must(template.New("validator_stats").Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files, "layout.html", "validator_stats_table.html"))

// Validator returns validator data using a go template
func ValidatorStatsTable(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	var index uint64
	var err error

	data := InitPageData(w, r, "validators", "/validators", "")
	data.HeaderAd = true

	// Request came with a hash
	if strings.Contains(vars["index"], "0x") || len(vars["index"]) == 96 {
		pubKey, err := hex.DecodeString(strings.Replace(vars["index"], "0x", "", -1))
		if err != nil {
			logger.Errorf("error parsing validator public key %v: %v", vars["index"], err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			logger.Errorf("error parsing validator pubkey: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		// Request came with a validator index number
		index, err = strconv.ParseUint(vars["index"], 10, 64)
		if err != nil {
			logger.Errorf("error parsing validator index: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	SetPageDataTitle(data, fmt.Sprintf("Validator %v Daily Statistics", index))
	data.Meta.Path = fmt.Sprintf("/validator/%v/stats", index)

	validatorStatsTablePageData := &types.ValidatorStatsTablePageData{
		ValidatorIndex: index,
		Rows:           make([]*types.ValidatorStatsTableRow, 0),
	}

	err = db.ReaderDb.Select(&validatorStatsTablePageData.Rows, `
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
	COALESCE(orphaned_attestations, 0) AS orphaned_attestations,
	COALESCE(proposed_blocks, 0) AS proposed_blocks,
	COALESCE(missed_blocks, 0) AS missed_blocks,
	COALESCE(orphaned_blocks, 0) AS orphaned_blocks,
	COALESCE(attester_slashings, 0) AS attester_slashings,
	COALESCE(proposer_slashings, 0) AS proposer_slashings,
	COALESCE(deposits, 0) AS deposits,
	COALESCE(deposits_amount, 0) AS deposits_amount,
	COALESCE(participated_sync, 0) AS participated_sync,
	COALESCE(missed_sync, 0) AS missed_sync,
	COALESCE(orphaned_sync, 0) AS orphaned_sync
	FROM validator_stats WHERE validatorindex = $1 ORDER BY day DESC`, index)

	if err != nil {
		logger.Errorf("error retrieving validator stats history: %v", err)
		http.Error(w, "Validator not found", http.StatusNotFound)
		return
	}

	balanceData, err := db.GetValidatorIncomeHistory([]uint64{index}, 0, 0)

	if err != nil {
		logger.Errorf("error retrieving validator income history: %v", err)
		http.Error(w, "Validator not found", http.StatusNotFound)
		return
	}
	// day => index mapping
	dayMapping := make(map[int64]int)
	for i := 0; i < len(validatorStatsTablePageData.Rows); i++ {
		dayMapping[validatorStatsTablePageData.Rows[i].Day] = i

	}

	for i := 0; i < len(balanceData); i++ {
		j, found := dayMapping[balanceData[i].Day]
		if !found {
			continue
		}
		validatorStatsTablePageData.Rows[j].StartBalance = balanceData[i].StartBalance
		validatorStatsTablePageData.Rows[j].EndBalance = balanceData[i].EndBalance
		validatorStatsTablePageData.Rows[j].Income = balanceData[i].Income
		validatorStatsTablePageData.Rows[j].Deposits = balanceData[i].DepositAmount
	}

	// if validatorStatsTablePageData.Rows[len(validatorStatsTablePageData.Rows)-1].Day == -1 {
	// 	validatorStatsTablePageData.Rows = validatorStatsTablePageData.Rows[:len(validatorStatsTablePageData.Rows)-1]
	// }

	data.Data = validatorStatsTablePageData
	err = validatorStatsTableTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

func ValidatorSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing validator index: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data draw-parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start start-parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	length, err := strconv.ParseInt(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length length-parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if length > 100 {
		length = 100
	}

	var countData []struct {
		TotalCount uint64 `db:"totalcount"`
		MaxPeriod  uint64 `db:"maxperiod"`
	}
	err = db.ReaderDb.Select(&countData, `
		SELECT count(*)*$1 AS totalcount, max(period) AS maxperiod 
		FROM sync_committees 
		WHERE validatorindex = $2`, utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod*utils.Config.Chain.Config.SlotsPerEpoch, index)
	if err != nil {
		logger.WithError(err).Errorf("error getting countData of sync-assignments")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalCount := uint64(0)
	tableData := [][]interface{}{}
	currentEpoch := services.LatestEpoch()

	if len(countData) > 0 {
		// only show 1 scheduled slot in the sync-table
		totalCount = countData[0].TotalCount

		res, err := db.BigtableClient.GetValidatorSyncDutiesHistory([]uint64{index}, currentEpoch-start, 10)
		if err != nil {
			logger.Errorf("error retrieving validator sync participations data from bigtable: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		tableData = make([][]interface{}, len(res[index]))
		for i, r := range res[index] {
			epoch := utils.EpochOfSlot(r.Slot)

			slotTime := utils.SlotToTime(r.Slot)

			// if slotTime.After(time.Now().Add(time.Minute)) {
			// 	continue
			// }
			if r.Status == 0 && time.Since(slotTime) > time.Minute {
				r.Status = 2
			}
			tableData[i] = []interface{}{
				fmt.Sprintf("%d", utils.SyncPeriodOfEpoch(epoch)),
				utils.FormatEpoch(epoch),
				utils.FormatBlockSlot(r.Slot),
				utils.FormatSyncParticipationStatus(r.Status),
			}
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
