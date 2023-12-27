package handlers

import (
	"bytes"
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
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lib/pq"
	protomath "github.com/protolambda/zrnt/eth2/util/math"
	"golang.org/x/sync/errgroup"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/juliangruber/go-intersect"

	itypes "github.com/gobitfly/eth-rewards/types"
)

var validatorEditFlash = "edit_validator_flash"

// Validator returns validator data using a go template
func Validator(w http.ResponseWriter, r *http.Request) {
	validatorTemplateFiles := append(layoutTemplateFiles,
		"validator/validator.html",
		"validator/heading.html",
		"validator/tables.html",
		"validator/modals.html",
		"modals.html",
		"validator/overview.html",
		"validator/charts.html",
		"validator/countdown.html",
		"components/flashMessage.html",
		"components/rocket.html")
	var validatorTemplate = templates.GetTemplate(validatorTemplateFiles...)

	currency := GetCurrency(r)

	timings := struct {
		Start         time.Time
		BasicInfo     time.Duration
		Earnings      time.Duration
		Deposits      time.Duration
		Proposals     time.Duration
		Charts        time.Duration
		Effectiveness time.Duration
		Statistics    time.Duration
		SyncStats     time.Duration
		Rocketpool    time.Duration
	}{
		Start: time.Now(),
	}

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	var index uint64
	var err error
	errFields := map[string]interface{}{
		"route": r.URL.String()}

	latestEpoch := services.LatestEpoch()
	latestProposedSlot := services.LatestProposedSlot()
	lastFinalizedEpoch := services.LatestFinalizedEpoch()
	isPreGenesis := false
	if latestEpoch == 0 {
		latestEpoch = 1
		latestProposedSlot = 1
		lastFinalizedEpoch = 1
		isPreGenesis = true
	}

	validatorPageData := types.ValidatorPageData{}

	validatorPageData.CappellaHasHappened = latestEpoch >= (utils.Config.Chain.ClConfig.CappellaForkEpoch)
	futureProposalEpoch := uint64(0)
	futureSyncDutyEpoch := uint64(0)

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
	validatorPageData.InclusionDelay = int64((utils.Config.Chain.ClConfig.Eth1FollowDistance*utils.Config.Chain.ClConfig.SecondsPerEth1Block+utils.Config.Chain.ClConfig.SecondsPerSlot*utils.Config.Chain.ClConfig.SlotsPerEpoch*utils.Config.Chain.ClConfig.EpochsPerEth1VotingPeriod)/3600) + 1

	data := InitPageData(w, r, "validators", "/validators", "", validatorTemplateFiles)
	validatorPageData.NetworkStats = services.LatestIndexPageData()
	validatorPageData.User = data.User

	validatorPageData.FlashMessage, err = utils.GetFlash(w, r, validatorEditFlash)
	if err != nil {
		utils.LogError(err, "error getting flash message", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if strings.Contains(vars["index"], "0x") || len(vars["index"]) == 96 {
		// Request came with a hash
		pubKey, err := hex.DecodeString(strings.Replace(vars["index"], "0x", "", -1))
		if err != nil {
			validatorNotFound(data, w, r, vars, "")
			return
		}
		errFields["pubKey"] = pubKey
		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			if err != sql.ErrNoRows {
				utils.LogError(err, "error getting index for validator based on pubkey", 0, errFields)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// the validator might only have a public key but no index yet
			var name string
			err := db.ReaderDb.Get(&name, `SELECT name FROM validator_names WHERE publickey = $1`, pubKey)
			if err != nil && err != sql.ErrNoRows {
				utils.LogError(err, "error getting validator-name from db for pubKey", 0, errFields)
				validatorNotFound(data, w, r, vars, "")
				return
				// err == sql.ErrNoRows -> unnamed
			} else {
				validatorPageData.Name = name
			}

			var pool string
			err = db.ReaderDb.Get(&pool, `SELECT pool FROM validator_pool WHERE publickey = $1`, pubKey)
			if err != nil && err != sql.ErrNoRows {
				utils.LogError(err, "error getting validator-pool from db for pubkey", 0, errFields)
				validatorNotFound(data, w, r, vars, "")
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
				utils.LogError(err, "error getting validator-deposits from db for pubkey", 0, errFields)
			}
			validatorPageData.DepositsCount = uint64(len(deposits.Eth1Deposits))
			validatorPageData.ShowMultipleWithdrawalCredentialsWarning = hasMultipleWithdrawalCredentials(deposits)
			if err != nil || len(deposits.Eth1Deposits) == 0 {
				validatorNotFound(data, w, r, vars, "")
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
			} else if time.Unix(latestDeposit, 0).Before(utils.SlotToTime(0)) {
				validatorPageData.InclusionDelay = 0
			}

			for _, deposit := range deposits.Eth1Deposits {
				if deposit.ValidSignature {
					validatorPageData.Eth1DepositAddress = deposit.FromAddress
					break
				}
			}

			// check if an invalid deposit exists
			for _, deposit := range deposits.Eth1Deposits {
				if !deposit.ValidSignature {
					validatorPageData.Status = "deposited_invalid"
					break
				}
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
				errFields["userID"] = data.User.UserID
				utils.LogError(err, "error getting tagged validators from db", 0, errFields)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			validatorPageData.Watchlist = watchlist

			if data.User.Authenticated {
				events := make([]types.EventNameCheckbox, 0)
				for _, ev := range types.AddWatchlistEvents {
					events = append(events, types.EventNameCheckbox{
						EventLabel: ev.Desc,
						EventName:  ev.Event,
						Active:     false,
						Warning:    ev.Warning,
						Info:       ev.Info,
					})
				}
				validatorPageData.AddValidatorWatchlistModal = &types.AddValidatorWatchlistModal{
					Events:         events,
					ValidatorIndex: validatorPageData.Index,
					CsrfField:      csrf.TemplateField(r),
				}
			}

			data.Data = validatorPageData
			if utils.IsApiRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				err = json.NewEncoder(w).Encode(data.Data)
			} else {
				err = validatorTemplate.ExecuteTemplate(w, "layout", data)
			}

			if handleTemplateError(w, r, "validator.go", "Validator", "Done (no index)", err) != nil {
				return // an error has occurred and was processed
			}
			return
		}
	} else {
		// Request came with a validator index number
		index, err = strconv.ParseUint(vars["index"], 10, 64)
		if err != nil || index > math.MaxInt32 { // index in postgres is limited to int
			validatorNotFound(data, w, r, vars, "")
			return
		}
	}

	errFields["index"] = index

	SetPageDataTitle(data, fmt.Sprintf("Validator %v", index))
	data.Meta.Path = fmt.Sprintf("/validator/%v", index)

	// we use MAX(validatorindex)+1 instead of COUNT(*) for querying the rank_count for performance-reasons
	err = db.ReaderDb.Get(&validatorPageData, `
		SELECT
			validators.pubkey,
			validators.validatorindex,
			validators.withdrawableepoch,
			validators.slashed,
			validators.activationeligibilityepoch,
			validators.activationepoch,
			validators.exitepoch,
			validators.withdrawalcredentials,
			COALESCE(validator_names.name, '') AS name,
			COALESCE(validator_pool.pool, '') AS pool,
			COALESCE(validator_performance.rank7d, 0) AS rank7d,
			COALESCE(validator_performance_count.total_count, 0) AS rank_count,
			validators.status,
			COALESCE(validators.balanceactivation, 0) AS balanceactivation,
			COALESCE((SELECT ARRAY_AGG(tag) FROM validator_tags WHERE publickey = validators.pubkey),'{}') AS tags
		FROM validators
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		LEFT JOIN validator_pool ON validators.pubkey = validator_pool.publickey
		LEFT JOIN validator_performance ON validators.validatorindex = validator_performance.validatorindex
		LEFT JOIN (SELECT MAX(validatorindex)+1 FROM validator_performance WHERE validatorindex < 2147483647 AND validatorindex >= 0) validator_performance_count(total_count) ON true
		WHERE validators.validatorindex = $1`, index)

	if err == sql.ErrNoRows {
		validatorNotFound(data, w, r, vars, "")
		return
	} else if err != nil {
		utils.LogError(err, "error getting validator page data info from db", 0, errFields)
		validatorNotFound(data, w, r, vars, "")
		return
	}

	lastAttestationSlots, err := db.BigtableClient.GetLastAttestationSlots([]uint64{index})
	if err != nil {
		utils.LogError(err, "error getting last attestation slots from bigtable", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	validatorPageData.LastAttestationSlot = lastAttestationSlots[index]

	lastStatsDay, lastStatsDayErr := services.LatestExportedStatisticDay()

	timings.BasicInfo = time.Since(timings.Start)

	timings.Start = time.Now()

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

	validatorPageData.Epoch = latestEpoch
	validatorPageData.Index = index

	if data.User.Authenticated {
		events := make([]types.EventNameCheckbox, 0)
		for _, ev := range types.AddWatchlistEvents {
			events = append(events, types.EventNameCheckbox{
				EventLabel: ev.Desc,
				EventName:  ev.Event,
				Active:     false,
				Warning:    ev.Warning,
				Info:       ev.Info,
			})
		}
		validatorPageData.AddValidatorWatchlistModal = &types.AddValidatorWatchlistModal{
			Events:         events,
			ValidatorIndex: validatorPageData.Index,
			CsrfField:      csrf.TemplateField(r),
		}
	}

	validatorPageData.ActivationEligibilityTs = utils.EpochToTime(validatorPageData.ActivationEligibilityEpoch)
	validatorPageData.ActivationTs = utils.EpochToTime(validatorPageData.ActivationEpoch)
	validatorPageData.ExitTs = utils.EpochToTime(validatorPageData.ExitEpoch)
	validatorPageData.WithdrawableTs = utils.EpochToTime(validatorPageData.WithdrawableEpoch)

	avgSyncInterval := uint64(getAvgSyncCommitteeInterval(1))
	avgSyncIntervalAsDuration := time.Duration(
		utils.Config.Chain.ClConfig.SecondsPerSlot*
			utils.SlotsPerSyncCommittee()*
			avgSyncInterval) * time.Second
	validatorPageData.AvgSyncInterval = &avgSyncIntervalAsDuration

	var lowerBoundDay uint64
	if lastStatsDay > 30 {
		lowerBoundDay = lastStatsDay - 30
	}
	g := errgroup.Group{}
	g.Go(func() error {
		start := time.Now()
		defer func() {
			timings.Charts = time.Since(start)
		}()

		incomeHistoryChartData, err := db.GetValidatorIncomeHistoryChart([]uint64{index}, currency, lastFinalizedEpoch, lowerBoundDay)
		if err != nil {
			return fmt.Errorf("error calling db.GetValidatorIncomeHistoryChart: %w", err)
		}

		if isPreGenesis {
			incomeHistoryChartData = make([]*types.ChartDataPoint, 0)
		}

		validatorPageData.IncomeHistoryChartData = incomeHistoryChartData
		return nil
	})

	g.Go(func() error {
		start := time.Now()
		defer func() {
			timings.Charts = time.Since(start)
		}()

		executionIncomeHistoryData, err := getExecutionChartData([]uint64{index}, currency, lowerBoundDay)
		if err != nil {
			return fmt.Errorf("error calling getExecutionChartData: %w", err)
		}

		validatorPageData.ExecutionIncomeHistoryData = executionIncomeHistoryData
		return nil
	})

	g.Go(func() error {
		// those functions need to be executed sequentially as both require the CurrentBalance value
		start := time.Now()
		defer func() {
			timings.Earnings = time.Since(start)
		}()
		earnings, balances, err := GetValidatorEarnings([]uint64{index}, currency)
		if err != nil {
			return fmt.Errorf("error getting validator earnings: %w", err)
		}
		validatorPageData.Income = earnings

		validatorPageData.ValidatorProposalData = earnings.ProposalData

		if latestEpoch < earnings.ProposalData.LastScheduledSlot/data.ChainConfig.SlotsPerEpoch {
			futureProposalEpoch = earnings.ProposalData.LastScheduledSlot / data.ChainConfig.SlotsPerEpoch
		}

		vbalance, ok := balances[validatorPageData.ValidatorIndex]
		if ok {
			validatorPageData.CurrentBalance = vbalance.Balance
			validatorPageData.EffectiveBalance = vbalance.EffectiveBalance
		}

		if bytes.Equal(validatorPageData.WithdrawCredentials[:1], []byte{0x01}) {
			// validators can have 0x01 credentials even before the cappella fork
			validatorPageData.IsWithdrawableAddress = true
		}

		if validatorPageData.CappellaHasHappened {
			// if we are currently past the cappella fork epoch, we can calculate the withdrawal information

			validatorSlice := []uint64{index}
			withdrawalsCount, err := db.GetTotalWithdrawalsCount(validatorSlice)
			if err != nil {
				return fmt.Errorf("error getting validator withdrawals count from db: %w", err)
			}
			validatorPageData.WithdrawalCount = withdrawalsCount
			lastWithdrawalsEpochs, err := db.GetLastWithdrawalEpoch(validatorSlice)
			if err != nil {
				return fmt.Errorf("error getting validator last withdrawal epoch from db: %w", err)
			}
			lastWithdrawalsEpoch := lastWithdrawalsEpochs[index]

			blsChange, err := db.GetValidatorBLSChange(validatorPageData.Index)
			if err != nil {
				return fmt.Errorf("error getting validator bls change from db: %w", err)
			}
			validatorPageData.BLSChange = blsChange

			if bytes.Equal(validatorPageData.WithdrawCredentials[:1], []byte{0x00}) && blsChange != nil {
				// blsChanges are only possible afters cappeala
				validatorPageData.IsWithdrawableAddress = true
			}

			// only calculate the expected next withdrawal if the validator is eligible
			isFullWithdrawal := validatorPageData.CurrentBalance > 0 && validatorPageData.WithdrawableEpoch <= validatorPageData.Epoch
			isPartialWithdrawal := validatorPageData.EffectiveBalance == utils.Config.Chain.ClConfig.MaxEffectiveBalance && validatorPageData.CurrentBalance > utils.Config.Chain.ClConfig.MaxEffectiveBalance
			if stats != nil && stats.LatestValidatorWithdrawalIndex != nil && stats.TotalValidatorCount != nil && validatorPageData.IsWithdrawableAddress && (isFullWithdrawal || isPartialWithdrawal) {
				distance, err := GetWithdrawableCountFromCursor(validatorPageData.Epoch, validatorPageData.Index, *stats.LatestValidatorWithdrawalIndex)
				if err != nil {
					return fmt.Errorf("error getting withdrawable validator count from cursor: %w", err)
				}

				timeToWithdrawal := utils.GetTimeToNextWithdrawal(distance)

				// it normally takes two epochs to finalize
				if timeToWithdrawal.After(utils.EpochToTime(latestEpoch + (latestEpoch - lastFinalizedEpoch))) {
					address, err := utils.WithdrawalCredentialsToAddress(validatorPageData.WithdrawCredentials)
					if err != nil {
						// warning only as "N/A" will be displayed
						logger.Warn("invalid withdrawal credentials")
					}

					// create the table data
					tableData := make([][]interface{}, 0, 1)
					var withdrawalCredentialsTemplate template.HTML
					if address != nil {
						withdrawalCredentialsTemplate = template.HTML(fmt.Sprintf(`<a href="/address/0x%x"><span class="text-muted">%s</span></a>`, address, utils.FormatAddress(address, nil, "", false, false, true)))
					} else {
						withdrawalCredentialsTemplate = `<span class="text-muted">N/A</span>`
					}

					var withdrawalAmount uint64
					if isFullWithdrawal {
						withdrawalAmount = validatorPageData.CurrentBalance
					} else {
						withdrawalAmount = validatorPageData.CurrentBalance - utils.Config.Chain.ClConfig.MaxEffectiveBalance
					}

					if latestEpoch == lastWithdrawalsEpoch {
						withdrawalAmount = 0
					}
					tableData = append(tableData, []interface{}{
						template.HTML(fmt.Sprintf(`<span class="text-muted">~ %s</span>`, utils.FormatEpoch(uint64(utils.TimeToEpoch(timeToWithdrawal))))),
						template.HTML(fmt.Sprintf(`<span class="text-muted">~ %s</span>`, utils.FormatBlockSlot(utils.TimeToSlot(uint64(timeToWithdrawal.Unix()))))),
						template.HTML(fmt.Sprintf(`<span class="">~ %s</span>`, utils.FormatTimestamp(timeToWithdrawal.Unix()))),
						withdrawalCredentialsTemplate,
						template.HTML(fmt.Sprintf(`<span class="text-muted"><span data-toggle="tooltip" title="If the withdrawal were to be processed at this very moment, this amount would be withdrawn"><i class="far ml-1 fa-question-circle" style="margin-left: 0px !important;"></i></span> %s</span>`, utils.FormatClCurrency(withdrawalAmount, currency, 6, true, false, false, true))),
					})

					validatorPageData.NextWithdrawalRow = tableData
				}
			}
		}
		return nil
	})

	g.Go(func() error {
		filter := db.WatchlistFilter{
			UserId:         data.User.UserID,
			Validators:     &pq.ByteaArray{validatorPageData.PublicKey},
			Tag:            types.ValidatorTagsWatchlist,
			JoinValidators: false,
			Network:        utils.GetNetwork(),
		}

		watchlist, err := db.GetTaggedValidators(filter)
		if err != nil {
			return fmt.Errorf("error getting tagged validators from db: %w", err)
		}

		validatorPageData.Watchlist = watchlist
		return nil
	})

	g.Go(func() error {
		start := time.Now()
		defer func() {
			timings.Deposits = time.Since(start)
		}()
		deposits, err := db.GetValidatorDeposits(validatorPageData.PublicKey)
		if err != nil {
			return fmt.Errorf("error getting validator-deposits from db: %w", err)
		}
		validatorPageData.Deposits = deposits
		validatorPageData.DepositsCount = uint64(len(deposits.Eth1Deposits))

		for _, deposit := range validatorPageData.Deposits.Eth1Deposits {
			if deposit.ValidSignature {
				validatorPageData.Eth1DepositAddress = deposit.FromAddress
				break
			}
		}

		validatorPageData.ShowMultipleWithdrawalCredentialsWarning = hasMultipleWithdrawalCredentials(validatorPageData.Deposits)

		return nil
	})

	g.Go(func() error {
		// we only need to get the queue information if we don't have an activation epoch but we have an eligibility epoch
		if validatorPageData.ActivationEpoch > 100_000_000 && validatorPageData.ActivationEligibilityEpoch < 100_000_000 {
			queueAhead, err := db.GetQueueAheadOfValidator(validatorPageData.Index)
			if err != nil {
				return fmt.Errorf("failed to retrieve queue ahead of validator %v: %w", validatorPageData.ValidatorIndex, err)
			}
			validatorPageData.QueuePosition = queueAhead + 1
			epochsToWait := queueAhead / *churnRate
			// calculate dequeue epoch
			estimatedActivationEpoch := validatorPageData.Epoch + epochsToWait + 1
			// add activation offset
			estimatedActivationEpoch += utils.Config.Chain.ClConfig.MaxSeedLookahead + 1
			validatorPageData.EstimatedActivationEpoch = estimatedActivationEpoch
			estimatedDequeueTs := utils.EpochToTime(estimatedActivationEpoch)
			validatorPageData.EstimatedActivationTs = estimatedDequeueTs
		}
		return nil
	})

	g.Go(func() error {
		// Every validator is scheduled to issue an attestation once per epoch
		// Hence we can calculate the number of attestations using the current epoch and the activation epoch
		// Special care needs to be take for exited and pending validators
		if validatorPageData.ExitEpoch != 9223372036854775807 && validatorPageData.ExitEpoch <= validatorPageData.Epoch {
			validatorPageData.AttestationsCount = validatorPageData.ExitEpoch - validatorPageData.ActivationEpoch
		} else if validatorPageData.ActivationEpoch > validatorPageData.Epoch {
			validatorPageData.AttestationsCount = 0

			return nil
		} else if isPreGenesis {
			validatorPageData.AttestationsCount = 1
			validatorPageData.MissedAttestationsCount = 0
			validatorPageData.ExecutedAttestationsCount = 0
			validatorPageData.UnmissedAttestationsPercentage = 1

			return nil
		} else {
			validatorPageData.AttestationsCount = validatorPageData.Epoch - validatorPageData.ActivationEpoch + 1

			// Check if the latest epoch still needs to be attested (scheduled) and if so do not count it
			attestationData, err := db.BigtableClient.GetValidatorAttestationHistory([]uint64{index}, validatorPageData.Epoch, validatorPageData.Epoch)
			if err != nil {
				return fmt.Errorf("error getting validator attestations data for epoch [%v]: %w", validatorPageData.Epoch, err)
			}

			if len(attestationData[index]) > 0 && attestationData[index][0].Status == 0 {
				validatorPageData.AttestationsCount--
			}
		}

		if validatorPageData.AttestationsCount > 0 {
			// get attestationStats from validator_stats
			attestationStats := struct {
				MissedAttestations uint64 `db:"missed_attestations"`
			}{}
			if lastStatsDay > 0 {
				err := db.ReaderDb.Get(&attestationStats, "SELECT missed_attestations_total AS missed_attestations FROM validator_stats WHERE validatorindex = $1 AND day = $2", index, lastStatsDay)
				if err == sql.ErrNoRows {
					logger.Warningf("no entry in validator_stats for validator index %v while lastStatsDay = %v", index, lastStatsDay)
				} else if err != nil {
					return fmt.Errorf("error getting validator attestationStats while lastStatsDay = %v: %w", lastStatsDay, err)
				}
			}

			// add attestationStats that are not yet in validator_stats (if any)
			nextStatsDayFirstEpoch, _ := utils.GetFirstAndLastEpochForDay(lastStatsDay + 1)
			if validatorPageData.Epoch > nextStatsDayFirstEpoch {
				lookback := validatorPageData.Epoch - nextStatsDayFirstEpoch
				missedAttestations, err := db.BigtableClient.GetValidatorMissedAttestationHistory([]uint64{index}, validatorPageData.Epoch-lookback, validatorPageData.Epoch-1)
				if err != nil {
					return fmt.Errorf("error getting validator attestations not in stats from bigtable: %w", err)
				}
				attestationStats.MissedAttestations += uint64(len(missedAttestations[index]))
			}

			if attestationStats.MissedAttestations > validatorPageData.AttestationsCount {
				// save guard against negative values (should never happen but happened once because of wrong data)
				attestationStats.MissedAttestations = validatorPageData.AttestationsCount
			}
			validatorPageData.MissedAttestationsCount = attestationStats.MissedAttestations
			validatorPageData.ExecutedAttestationsCount = validatorPageData.AttestationsCount - validatorPageData.MissedAttestationsCount
			validatorPageData.UnmissedAttestationsPercentage = float64(validatorPageData.ExecutedAttestationsCount) / float64(validatorPageData.AttestationsCount)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		if validatorPageData.Slashed {
			var slashingInfo struct {
				Slot    uint64
				Slasher uint64
				Reason  string
			}
			err = db.ReaderDb.Get(&slashingInfo,
				`SELECT block_slot AS slot, proposer AS slasher, 'Attestation Violation' AS reason
					FROM blocks_attesterslashings a1 LEFT JOIN blocks b1 ON b1.slot = a1.block_slot
					WHERE b1.status = '1' AND $1 = ANY(a1.attestation1_indices) AND $1 = ANY(a1.attestation2_indices)
				UNION ALL
				SELECT block_slot AS slot, proposer AS slasher, 'Proposer Violation' AS reason
					FROM blocks_proposerslashings a2 LEFT JOIN blocks b2 ON b2.slot = a2.block_slot
					WHERE b2.status = '1' AND a2.proposerindex = $1
				LIMIT 1`,
				index)
			if err != nil {
				return fmt.Errorf("error getting validator slashing info: %w", err)
			}
			validatorPageData.SlashedBy = slashingInfo.Slasher
			validatorPageData.SlashedAt = slashingInfo.Slot
			validatorPageData.SlashedFor = slashingInfo.Reason
		}

		err = db.ReaderDb.Get(&validatorPageData.SlashingsCount, `SELECT COALESCE(SUM(attesterslashingscount) + SUM(proposerslashingscount), 0) FROM blocks WHERE blocks.proposer = $1 AND blocks.status = '1'`, index)
		if err != nil {
			return fmt.Errorf("error getting slashings-count: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		eff, err := db.BigtableClient.GetValidatorEffectiveness([]uint64{index}, validatorPageData.Epoch-1)
		if err != nil {
			return fmt.Errorf("error getting validator effectiveness: %w", err)
		}
		if len(eff) > 1 {
			return fmt.Errorf("error getting validator effectiveness: invalid length %v", len(eff))
		} else if len(eff) == 0 {
			validatorPageData.AttestationInclusionEffectiveness = 0
		} else {
			validatorPageData.AttestationInclusionEffectiveness = eff[0].AttestationEfficiency
		}
		return nil
	})

	g.Go(func() error {
		validatorPageData.SlotsPerSyncCommittee = utils.SlotsPerSyncCommittee()

		// sync participation
		// get all sync periods this validator has been part of
		var actualSyncPeriods []struct {
			Period     uint64 `db:"period"`
			FirstEpoch uint64 `db:"firstepoch"`
			LastEpoch  uint64 `db:"lastepoch"`
		}
		allSyncPeriods := actualSyncPeriods

		err := db.ReaderDb.Select(&allSyncPeriods, `
		SELECT period, GREATEST(period*$1, $2) AS firstepoch, ((period+1)*$1)-1 AS lastepoch
		FROM sync_committees 
		WHERE validatorindex = $3
		ORDER BY period desc`, utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod, utils.Config.Chain.ClConfig.AltairForkEpoch, index)
		if err != nil {
			return fmt.Errorf("error getting sync participation count data of sync-assignments: %w", err)
		}

		if len(allSyncPeriods) > 0 && allSyncPeriods[0].LastEpoch > latestEpoch {
			futureSyncDutyEpoch = allSyncPeriods[0].LastEpoch
		}

		// remove scheduled committees
		for i, syncPeriod := range allSyncPeriods {
			if syncPeriod.FirstEpoch <= latestEpoch {
				actualSyncPeriods = allSyncPeriods[i:]
				break
			}
		}

		if len(actualSyncPeriods) > 0 {
			// get sync stats from validator_stats
			syncStats := types.SyncCommitteesStats{}
			if lastStatsDay > 0 {
				err = db.ReaderDb.Get(&syncStats, `
					SELECT
						COALESCE(participated_sync_total, 0) AS participated_sync,
						COALESCE(missed_sync_total, 0) AS missed_sync,
						COALESCE(orphaned_sync_total, 0) AS orphaned_sync
					FROM validator_stats
					WHERE validatorindex = $1 AND day = $2`, index, lastStatsDay)
				if err != nil {
					return fmt.Errorf("error getting validator syncStats: %w", err)
				}
			}

			// if sync duties of last period haven't fully been exported yet, fetch remaining duties from bigtable
			lastExportedEpoch := (lastStatsDay+1)*utils.EpochsPerDay() - 1

			if lastStatsDayErr == db.ErrNoStats {
				lastExportedEpoch = 0
			}
			lastSyncPeriod := actualSyncPeriods[0]
			if lastSyncPeriod.LastEpoch > lastExportedEpoch {
				res, err := db.BigtableClient.GetValidatorSyncDutiesHistory([]uint64{index}, (lastExportedEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch, latestProposedSlot)
				if err != nil {
					return fmt.Errorf("error getting validator sync participations data from bigtable: %w", err)
				}
				syncStatsBt := utils.AddSyncStats([]uint64{index}, res, nil)
				// if last sync period is the current one, add remaining scheduled slots
				if lastSyncPeriod.LastEpoch >= latestEpoch {
					syncStatsBt.ScheduledSlots += utils.GetRemainingScheduledSyncDuties(1, syncStatsBt, lastExportedEpoch, lastSyncPeriod.FirstEpoch)
				}

				syncStats.MissedSlots += syncStatsBt.MissedSlots
				syncStats.ParticipatedSlots += syncStatsBt.ParticipatedSlots
				syncStats.ScheduledSlots += syncStatsBt.ScheduledSlots
			}
			validatorPageData.SlotsDoneInCurrentSyncCommittee = validatorPageData.SlotsPerSyncCommittee - syncStats.ScheduledSlots

			validatorPageData.ParticipatedSyncCountSlots = syncStats.ParticipatedSlots
			validatorPageData.MissedSyncCountSlots = syncStats.MissedSlots
			validatorPageData.OrphanedSyncCountSlots = syncStats.OrphanedSlots
			validatorPageData.ScheduledSyncCountSlots = syncStats.ScheduledSlots
			// actual sync duty count and percentage
			validatorPageData.SyncCount = uint64(len(actualSyncPeriods))
			validatorPageData.UnmissedSyncPercentage = float64(validatorPageData.ParticipatedSyncCountSlots) / float64(validatorPageData.ParticipatedSyncCountSlots+validatorPageData.MissedSyncCountSlots+validatorPageData.OrphanedSyncCountSlots)
		}
		// sync luck
		if len(allSyncPeriods) > 0 {
			maxPeriod := allSyncPeriods[0].Period
			expectedSyncCount, err := getExpectedSyncCommitteeSlots([]uint64{index}, lastFinalizedEpoch)
			if err != nil {
				return fmt.Errorf("error getting expected sync committee slots: %w", err)
			}
			if expectedSyncCount != 0 {
				validatorPageData.SyncLuck = float64(validatorPageData.ParticipatedSyncCountSlots+validatorPageData.MissedSyncCountSlots) / float64(expectedSyncCount)
			}
			nextEstimate := utils.EpochToTime(utils.FirstEpochOfSyncPeriod(maxPeriod + avgSyncInterval))
			validatorPageData.SyncEstimate = &nextEstimate
		}
		return nil
	})

	g.Go(func() error {
		// add rocketpool-data if available
		validatorPageData.Rocketpool = &types.RocketpoolValidatorPageData{}
		err := db.ReaderDb.Get(validatorPageData.Rocketpool, `
		SELECT
			rplm.node_address      					AS node_address,
			rplm.address           					AS minipool_address,
			rplm.node_fee          					AS minipool_node_fee,
			rplm.deposit_type      					AS minipool_deposit_type,
			rplm.status            					AS minipool_status,
			rplm.status_time       					AS minipool_status_time,
			COALESCE(rplm.penalty_count,0) 			AS penalty_count,
			rpln.timezone_location 					AS node_timezone_location,
			rpln.rpl_stake 							AS node_rpl_stake,
			rpln.max_rpl_stake 						AS node_max_rpl_stake,
			rpln.min_rpl_stake 						AS node_min_rpl_stake,
			rpln.rpl_cumulative_rewards 			AS rpl_cumulative_rewards,
			rpln.claimed_smoothing_pool 			AS claimed_smoothing_pool,
			rpln.unclaimed_smoothing_pool 			AS unclaimed_smoothing_pool,
			rpln.unclaimed_rpl_rewards 				AS unclaimed_rpl_rewards,
			COALESCE(node_deposit_balance, 0) 		AS node_deposit_balance,
			COALESCE(node_refund_balance, 0) 		AS node_refund_balance,
			COALESCE(user_deposit_balance, 0) 		AS user_deposit_balance,
			COALESCE(rpln.effective_rpl_stake, 0) 	AS effective_rpl_stake,
			COALESCE(deposit_credit, 0) 			AS deposit_credit,
			COALESCE(is_vacant, false) 				AS is_vacant,
			version,
			COALESCE(rpln.smoothing_pool_opted_in, false) AS smoothing_pool_opted_in 
		FROM validators
		LEFT JOIN rocketpool_minipools rplm ON rplm.pubkey = validators.pubkey
		LEFT JOIN rocketpool_nodes rpln ON rplm.node_address = rpln.address
		WHERE validators.validatorindex = $1
		ORDER BY rplm.status_time DESC 
		LIMIT 1`, index)
		if err == nil && (validatorPageData.Rocketpool.MinipoolAddress != nil || validatorPageData.Rocketpool.NodeAddress != nil) {
			validatorPageData.IsRocketpool = true
			if utils.Config.Chain.ClConfig.DepositChainID == 1 {
				validatorPageData.Rocketpool.RocketscanUrl = "rocketscan.io"
			} else if utils.Config.Chain.ClConfig.DepositChainID == 5 {
				validatorPageData.Rocketpool.RocketscanUrl = "prater.rocketscan.io"
			}
		} else if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error getting rocketpool-data for validator for %v route: %w", r.URL.String(), err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		utils.LogError(err, "error getting validator data", 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	validatorPageData.FutureDutiesEpoch = protomath.MaxU64(futureProposalEpoch, futureSyncDutyEpoch)

	data.Data = validatorPageData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		err = validatorTemplate.ExecuteTemplate(w, "layout", data)
	}

	if handleTemplateError(w, r, "validator.go", "Validator", "Done", err) != nil {
		return // an error has occurred and was processed
	}
}

// Returns true if there are more than one different withdrawal credentials within both Eth1Deposits and Eth2Deposits
func hasMultipleWithdrawalCredentials(deposits *types.ValidatorDeposits) bool {
	if deposits == nil {
		return false
	}

	credential := make([]byte, 0)

	if deposits == nil {
		return false
	}

	// check Eth1Deposits
	for _, deposit := range deposits.Eth1Deposits {
		if len(credential) == 0 {
			credential = deposit.WithdrawalCredentials
		} else if !bytes.Equal(credential, deposit.WithdrawalCredentials) {
			return true
		}
	}

	// check Eth2Deposits
	for _, deposit := range deposits.Eth2Deposits {
		if len(credential) == 0 {
			credential = deposit.Withdrawalcredentials
		} else if !bytes.Equal(credential, deposit.Withdrawalcredentials) {
			return true
		}
	}

	return false
}

// ValidatorDeposits returns a validator's deposits in json
func ValidatorDeposits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	pubkey, err := hex.DecodeString(strings.Replace(vars["pubkey"], "0x", "", -1))
	if err != nil {
		logger.Warnf("error parsing validator public key %v: %v", vars["pubkey"], err)
		http.Error(w, "Error: Invalid parameter public key.", http.StatusBadRequest)
		return
	}

	errFields := map[string]interface{}{
		"route":  r.URL.String(),
		"pubkey": pubkey}

	deposits, err := db.GetValidatorDeposits(pubkey)
	if err != nil {
		utils.LogError(err, "error getting validator-deposits from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(deposits)
	if err != nil {
		utils.LogError(err, "error encoding validator-deposits", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ValidatorAttestationInclusionEffectiveness returns a validator's effectiveness in json
func ValidatorAttestationInclusionEffectiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil || index > math.MaxInt32 { // index in postgres is limited to int
		logger.Warnf("error parsing validator index: %v", err)
		http.Error(w, "Error: Invalid parameter validator index.", http.StatusBadRequest)
		return
	}
	epoch := services.LatestEpoch()
	if epoch > 0 {
		epoch = epoch - 1
	}

	errFields := map[string]interface{}{
		"route": r.URL.String(),
		"index": index,
		"epoch": epoch}

	eff, err := db.BigtableClient.GetValidatorEffectiveness([]uint64{index}, epoch)
	if err != nil {
		utils.LogError(err, "error getting validator effectiveness from bigtable", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type resp struct {
		Effectiveness float64 `json:"effectiveness"`
	}

	errFields["effectiveness length"] = len(eff)
	if len(eff) > 1 {
		utils.LogError(err, "error getting validator effectiveness because of invalid length", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if len(eff) == 0 {
		err = json.NewEncoder(w).Encode(resp{Effectiveness: 0})
		if err != nil {
			utils.LogError(err, "error encoding json response", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = json.NewEncoder(w).Encode(resp{Effectiveness: eff[0].AttestationEfficiency})
		if err != nil {
			utils.LogError(err, "error encoding json response", 0, errFields)
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
	if err != nil || index > math.MaxInt32 { // index in postgres is limited to int
		logger.Warnf("error parsing validator index: %v", err)
		http.Error(w, "Error: Invalid parameter validator index.", http.StatusBadRequest)
		return
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

	errFields := map[string]interface{}{
		"route":  r.URL.String(),
		"index":  index,
		"draw":   draw,
		"start":  start,
		"length": length}

	var totalCount uint64

	err = db.ReaderDb.Get(&totalCount, "SELECT COUNT(*) FROM blocks WHERE proposer = $1", index)
	if err != nil {
		utils.LogError(err, "error getting proposed blocks count from db", 0, errFields)
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
			epoch, 
			slot, 
			proposer, 
			blockroot, 
			parentroot, 
			attestationscount, 
			depositscount,
			withdrawalcount, 
			voluntaryexitscount, 
			proposerslashingscount, 
			attesterslashingscount, 
			status, 
			graffiti 
		FROM blocks 
		WHERE proposer = $1
		ORDER BY `+orderBy+` `+orderDir+`
		LIMIT $2 OFFSET $3`, index, length, start)

	if err != nil {
		utils.LogError(err, "error getting proposed blocks data from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		tableData[i] = []interface{}{
			utils.FormatEpoch(b.Epoch),
			utils.FormatBlockSlot(b.Slot),
			utils.FormatBlockStatus(b.Status, b.Slot),
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
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ValidatorAttestations returns a validators attestations in json
func ValidatorAttestations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil || index > math.MaxInt32 { // index in postgres is limited to int
		logger.Warnf("error parsing validator index: %v", err)
		http.Error(w, "Error: Invalid parameter validator index.", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseInt(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}

	errFields := map[string]interface{}{
		"route": r.URL.String(),
		"index": index,
		"draw":  draw,
		"start": start}

	length := 10

	epoch := services.LatestEpoch()

	ae := struct {
		ActivationEpoch uint64
		ExitEpoch       uint64
	}{}

	err = db.ReaderDb.Get(&ae, "SELECT activationepoch, exitepoch FROM validators WHERE validatorindex = $1", index)
	if err != nil {
		utils.LogError(err, "error getting attestations count from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalCount := epoch - ae.ActivationEpoch + 1
	if ae.ActivationEpoch > epoch {
		totalCount = 0
	}
	lastAttestationEpoch := epoch
	if ae.ExitEpoch != 9223372036854775807 && ae.ExitEpoch <= epoch {
		lastAttestationEpoch = ae.ExitEpoch - 1
		totalCount = ae.ExitEpoch - ae.ActivationEpoch
	}

	tableData := [][]interface{}{}

	if totalCount > 0 {
		endEpoch := int64(lastAttestationEpoch) - start
		if endEpoch < 0 {
			endEpoch = 0
		}

		startEpoch := endEpoch - int64(length) + 1
		if startEpoch < 0 {
			startEpoch = 0
		}

		attestationData, err := db.BigtableClient.GetValidatorAttestationHistory([]uint64{index}, uint64(startEpoch), uint64(endEpoch))
		if err != nil {
			errFields["startEpoch"] = startEpoch
			errFields["endEpoch"] = endEpoch
			utils.LogError(err, "error getting validator attestations data from bigtable", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		tableData = make([][]interface{}, len(attestationData[index]))

		for i, history := range attestationData[index] {

			if history.Status == 0 && int64(history.Epoch) < int64(epoch)-1 {
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
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ValidatorWithdrawals returns a validators withdrawals in json
func ValidatorWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqCurrency := GetCurrency(r)

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil || index > math.MaxInt32 { // index in postgres is limited to int
		logger.Warnf("error parsing validator index: %v", err)
		http.Error(w, "Error: Invalid parameter validator index.", http.StatusBadRequest)
		return
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

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "block_slot",
		"1": "block_slot",
		"2": "block_slot",
		"3": "address",
		"4": "amount",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "block_slot"
	}
	orderDir := q.Get("order[0][dir]")
	if orderDir != "asc" {
		orderDir = "desc"
	}

	errFields := map[string]interface{}{
		"route":       r.URL.String(),
		"index":       index,
		"draw":        draw,
		"start":       start,
		"orderColumn": orderColumn,
		"orderBy":     orderBy,
		"orderDir":    orderDir}

	length := uint64(10)

	withdrawalCount, err := db.GetTotalWithdrawalsCount([]uint64{index})
	if err != nil {
		utils.LogError(err, "error getting validator withdrawals count from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	withdrawals, err := db.GetValidatorWithdrawals(index, length, start, orderBy, orderDir)
	if err != nil {
		utils.LogError(err, "error getting validator withdrawals from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, 0, len(withdrawals))

	for _, w := range withdrawals {
		tableData = append(tableData, []interface{}{
			utils.FormatEpoch(utils.EpochOfSlot(w.Slot)),
			utils.FormatBlockSlot(w.Slot),
			utils.FormatTimestamp(utils.SlotToTime(w.Slot).Unix()),
			utils.FormatAddress(w.Address, nil, "", false, false, true),
			utils.FormatClCurrency(w.Amount, reqCurrency, 6, true, false, false, true),
		})
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    withdrawalCount,
		RecordsFiltered: withdrawalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ValidatorSlashings returns a validators slashings in json
func ValidatorSlashings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil || index > math.MaxInt32 { // index in postgres is limited to int
		logger.Warnf("error parsing validator index: %v", err)
		http.Error(w, "Error: Invalid parameter validator index.", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}

	errFields := map[string]interface{}{
		"route": r.URL.String(),
		"index": index,
		"draw":  draw}

	var totalCount uint64
	err = db.ReaderDb.Get(&totalCount, `
		SELECT
			(
				SELECT COUNT(*) FROM blocks_attesterslashings a
				INNER JOIN blocks b ON b.slot = a.block_slot AND b.proposer = $1
				WHERE attestation1_indices IS NOT null AND attestation2_indices IS NOT null
			) + (
				SELECT COUNT(*) FROM blocks_proposerslashings c
				INNER JOIN blocks d ON d.slot = c.block_slot AND d.proposer = $1
			)`, index)
	if err != nil {
		utils.LogError(err, "error getting totalCount of validator-slashings from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		utils.LogError(err, "error getting validator attestations data from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var proposerSlashings []*types.ValidatorProposerSlashing
	err = db.ReaderDb.Select(&proposerSlashings, `
		SELECT blocks.slot, blocks.epoch, blocks.proposer, blocks_proposerslashings.proposerindex 
		FROM blocks_proposerslashings 
		INNER JOIN blocks ON blocks.proposer = $1 AND blocks_proposerslashings.block_slot = blocks.slot`, index)
	if err != nil {
		utils.LogError(err, "error getting validator proposer slashings data from db", 0, errFields)
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
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Function checks if the generated ECDSA signature has correct lentgth and if needed sets recovery byte to 0 or 1
func sanitizeSignature(sig string) ([]byte, error) {
	sig = strings.Replace(sig, "0x", "", -1)
	decodedSig, _ := hex.DecodeString(sig)
	if len(decodedSig) != 65 {
		return nil, fmt.Errorf("signature is less than 65 bytes (len = %v)", len(decodedSig))
	}
	if decodedSig[crypto.RecoveryIDOffset] == 27 || decodedSig[crypto.RecoveryIDOffset] == 28 {
		decodedSig[crypto.RecoveryIDOffset] -= 27
	}
	return []byte(decodedSig), nil
}

// Function tries to find the substring.
//
// If successful it turns string into []byte value and returns it
//
// If it fails, it will try to decode `msg`value from Hexadecimal to string and retry search again
func sanitizeMessage(msg string) ([]byte, error) {
	subString := "beaconcha.in"

	if strings.Contains(msg, subString) {
		return []byte(msg), nil
	} else {
		decoded := strings.Replace(msg, "0x", "", -1)
		dec, _ := hex.DecodeString(decoded)
		decodedString := (string(dec))
		if strings.Contains(decodedString, subString) {
			return []byte(decodedString), nil
		}
		return nil, fmt.Errorf("%v was not found", subString)
	}
}

func SaveValidatorName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pubkey := vars["pubkey"]
	pubkey = strings.ToLower(pubkey)
	pubkey = strings.Replace(pubkey, "0x", "", -1)

	pubkeyDecoded, err := hex.DecodeString(pubkey)
	if err != nil {
		logger.Warnf("error parsing submitted pubkey %v: %v", pubkey, err)
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
		logger.Warnf("error decoding submitted signature %v: %v", signature, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	msg, err := sanitizeMessage(signatureWrapper.Msg)
	if err != nil {
		logger.Warnf("Message is invalid %v: %v", signatureWrapper.Msg, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided message is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}
	msgHash := accounts.TextHash(msg)

	sig, err := sanitizeSignature(signatureWrapper.Sig)
	if err != nil {
		logger.Warnf("error parsing submitted signature %v: %v", signatureWrapper.Sig, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	recoveredPubkey, err := crypto.SigToPub(msgHash, sig)
	if err != nil {
		logger.Warnf("error recovering pubkey: %v", err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubkey)

	errFields := map[string]interface{}{
		"route":            r.URL.String(),
		"pubkey":           pubkey,
		"name":             name,
		"applyNameToAll":   applyNameToAll,
		"recoveredAddress": recoveredAddress}

	var depositedAddress string
	deposits, err := db.GetValidatorDeposits(pubkeyDecoded)
	if err != nil {
		utils.LogError(err, "error getting validator-deposits from db for signature verification", 0, errFields)
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
				utils.LogError(err, "error saving validator name", 0, errFields)
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
				utils.LogError(err, "error saving validator name", 0, errFields)
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
	pageLength := 10
	maxPages := 10

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 31)
	if err != nil {
		logger.Warnf("error parsing validator index: %v", err)
		http.Error(w, "Error: Invalid parameter validator index.", http.StatusBadRequest)
		return
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

	errFields := map[string]interface{}{
		"route": r.URL.String(),
		"index": index,
		"draw":  draw,
		"start": start}

	var activationAndExitEpoch = struct {
		ActivationEpoch uint64 `db:"activationepoch"`
		ExitEpoch       uint64 `db:"exitepoch"`
	}{}
	err = db.ReaderDb.Get(&activationAndExitEpoch, "SELECT activationepoch, exitepoch FROM validators WHERE validatorindex = $1", index)
	if err != nil {
		utils.LogError(err, "error getting activationAndExitEpoch for validator-history from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalCount := uint64(0)

	// Every validator is scheduled to issue an attestation once per epoch
	// Hence we can calculate the number of attestations using the current epoch and the activation epoch
	// Special care needs to be take for exited and pending validators
	if activationAndExitEpoch.ExitEpoch != 9223372036854775807 && activationAndExitEpoch.ExitEpoch <= services.LatestFinalizedEpoch() {
		totalCount += activationAndExitEpoch.ExitEpoch - activationAndExitEpoch.ActivationEpoch
	} else {
		totalCount += services.LatestFinalizedEpoch() - activationAndExitEpoch.ActivationEpoch + 1
	}

	if start > uint64((maxPages-1)*pageLength) {
		start = uint64((maxPages - 1) * pageLength)
	}

	currentEpoch := services.LatestEpoch()

	if currentEpoch != 0 {
		currentEpoch = currentEpoch - 1
	}
	var postExitEpochs uint64 = 0
	// for an exited validator we show the history until his exit or (in rare cases) until his last sync / propose duties are finished
	if activationAndExitEpoch.ExitEpoch != 9223372036854775807 && currentEpoch > (activationAndExitEpoch.ExitEpoch-1) {
		currentEpoch = activationAndExitEpoch.ExitEpoch - 1

		var lastActionDay uint64
		// let's get the last day where validator had duties which can be after the exit epoch of the validator
		err = db.ReaderDb.Get(&lastActionDay, `
			SELECT COALESCE(MAX(day), 0) 
			FROM validator_stats 
			WHERE 
				validatorindex = $1 AND 
				( missed_sync > 0 
					OR orphaned_sync > 0 
					OR participated_sync > 0 
					OR proposed_blocks > 0 
					OR missed_blocks > 0 
					OR orphaned_blocks > 0 
			);`, index)
		if err != nil {
			utils.LogError(err, "error getting lastActionDay for validator-history from db", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		lastActionEpoch := (lastActionDay + 1) * utils.EpochsPerDay()
		// if the validator had some duties after the exit epoch we calculate how many epochs we have to check after the exit epoch
		if lastActionEpoch > currentEpoch {
			postExitEpochs = protomath.MinU64(lastActionEpoch, services.LatestEpoch()-1) - currentEpoch
		}
	}

	tableData := make([][]interface{}, 0)
	if postExitEpochs > 0 {
		startEpoch := currentEpoch + 1
		endEpoch := startEpoch + postExitEpochs
		withdrawalMap, incomeDetails, err := getWithdrawalAndIncome(index, startEpoch, endEpoch)
		if err != nil {
			return
		}

		// if there are additional epochs with duties we have to go through all of them as there can be gaps (after the exit before the duty)
		for i := endEpoch; i >= startEpoch; i-- {
			if i > endEpoch {
				break
			}

			if incomeDetails[index] == nil || incomeDetails[index][i] == nil {
				continue
			}
			totalCount++
			// for paging we skip the first X epochs with duties
			if start > 0 {
				start--
				continue
			} else if len(tableData) >= pageLength {
				continue
			}
			tableData = append(tableData, incomeToTableData(i, incomeDetails[index][i], withdrawalMap[i], currency))
		}
	}

	postExitItemLength := len(tableData)
	// if we have already found enough items we can skip the 'normal' step.
	if postExitItemLength < pageLength {
		startEpoch := currentEpoch - start - uint64(pageLength-1-postExitItemLength) // we only get the exact number of epochs to get to the page length
		endEpoch := currentEpoch - start
		if startEpoch > endEpoch { // handle underflows of startEpoch
			startEpoch = 0
		}

		withdrawalMap, incomeDetails, err := getWithdrawalAndIncome(index, startEpoch, endEpoch)
		if err != nil {
			return
		}

		for epoch := endEpoch; epoch >= startEpoch && len(tableData) < pageLength; epoch-- {

			if epoch > endEpoch {
				break
			}

			if incomeDetails[index] == nil || incomeDetails[index][epoch] == nil {
				if epoch <= endEpoch {
					rewardsStr := "pending..."
					eventStr := template.HTML("")
					if epoch < activationAndExitEpoch.ActivationEpoch {
						rewardsStr = ""
						eventStr = utils.FormatAttestationStatusShort(5)
					}
					tableData = append(tableData, []interface{}{
						utils.FormatEpoch(epoch),
						rewardsStr,
						template.HTML(""),
						eventStr,
					})
				}
				continue
			}
			tableData = append(tableData, incomeToTableData(epoch, incomeDetails[index][epoch], withdrawalMap[epoch], currency))

		}
	}

	if len(tableData) == 0 {
		tableData = append(tableData, []interface{}{
			template.HTML("Validator no longer active"),
		})
	}

	totalCount = protomath.MinU64(totalCount, uint64(pageLength*maxPages))

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func getWithdrawalAndIncome(index uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*types.ValidatorWithdrawal, map[uint64]map[uint64]*itypes.ValidatorEpochIncome, error) {
	g := new(errgroup.Group)

	errFields := map[string]interface{}{
		"index":      index,
		"startEpoch": startEpoch,
		"endEpoch":   endEpoch}

	var withdrawals []*types.WithdrawalsByEpoch
	g.Go(func() error {
		var err error
		withdrawals, err = db.GetValidatorsWithdrawalsByEpoch([]uint64{index}, startEpoch, endEpoch)
		if err != nil {
			utils.LogError(err, "error getting validator withdrawals by epoch", 0, errFields)
			return err
		}
		return nil
	})

	var incomeDetails map[uint64]map[uint64]*itypes.ValidatorEpochIncome
	g.Go(func() error {
		var err error
		incomeDetails, err = db.BigtableClient.GetValidatorIncomeDetailsHistory([]uint64{index}, startEpoch, endEpoch)
		if err != nil {
			utils.LogError(err, "error getting validator income details history from bigtable", 0, errFields)
			return err
		}
		return nil
	})

	err := g.Wait()
	withdrawalMap := make(map[uint64]*types.ValidatorWithdrawal)
	for _, withdrawals := range withdrawals {
		withdrawalMap[withdrawals.Epoch] = &types.ValidatorWithdrawal{
			Index:  withdrawals.ValidatorIndex,
			Epoch:  withdrawals.Epoch,
			Amount: withdrawals.Amount,
			Slot:   withdrawals.Epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch,
		}
	}
	return withdrawalMap, incomeDetails, err
}

func incomeToTableData(epoch uint64, income *itypes.ValidatorEpochIncome, withdrawal *types.ValidatorWithdrawal, currency string) []interface{} {
	events := template.HTML("")
	if income.AttestationSourcePenalty > 0 && income.AttestationTargetPenalty > 0 {
		events += utils.FormatAttestationStatusShort(2)
	} else {
		events += utils.FormatAttestationStatusShort(1)
	}

	if income.ProposerAttestationInclusionReward > 0 {
		block := utils.FormatBlockStatusShort(1, 0)
		events += block
	} else if income.ProposalsMissed > 0 {
		block := utils.FormatBlockStatusShort(2, 0)
		events += block
	}

	if withdrawal != nil {
		withdrawal := utils.FormatWithdrawalShort(uint64(withdrawal.Slot), withdrawal.Amount)
		events += withdrawal
	}

	rewards := income.TotalClRewards()
	return []interface{}{
		utils.FormatEpoch(epoch),
		utils.FormatBalanceChangeFormatted(&rewards, currency, income),
		template.HTML(""),
		events,
	}
}

// Validator returns validator data using a go template
func ValidatorStatsTable(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "validator_stats_table.html")
	var validatorStatsTableTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	var index uint64
	var err error

	errFields := map[string]interface{}{
		"route": r.URL.String()}

	data := InitPageData(w, r, "validators", "/validators", "", templateFiles)

	// Request came with a hash
	if utils.IsHash(vars["index"]) {
		pubKey, err := hex.DecodeString(strings.TrimPrefix(vars["index"], "0x"))
		if err != nil {
			utils.LogError(err, "error decoding validator pubkey", 0, errFields)
			validatorNotFound(data, w, r, vars, "/stats")
			return
		}
		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			if err != sql.ErrNoRows {
				errFields["pubkey"] = pubKey
				utils.LogError(err, "error getting validator index from db", 0, errFields)
			}
			validatorNotFound(data, w, r, vars, "/stats")
			return
		}
	} else {
		// Request came with a validator index number
		index, err = strconv.ParseUint(vars["index"], 10, 31)

		if err != nil {
			// Request is not a valid index number
			logger.Warnf("error parsing validator index: %v", err)
			validatorNotFound(data, w, r, vars, "/stats")
			return
		}
	}

	errFields["index"] = index

	// Check if validator index is available
	var doesIndexExist bool
	err = db.ReaderDb.Get(&doesIndexExist, "SELECT EXISTS (SELECT validatorindex FROM validators WHERE validatorindex = $1)", index)
	if err != nil || !doesIndexExist {
		if err != nil {
			utils.LogError(err, "error checking for index in validators table", 0, errFields)
		}
		validatorNotFound(data, w, r, vars, "/stats")
		return
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
		COALESCE(proposed_blocks, 0) AS proposed_blocks,
		COALESCE(missed_blocks, 0) AS missed_blocks,
		COALESCE(orphaned_blocks, 0) AS orphaned_blocks,
		COALESCE(attester_slashings, 0) AS attester_slashings,
		COALESCE(proposer_slashings, 0) AS proposer_slashings,
		COALESCE(deposits, 0) AS deposits,
		COALESCE(deposits_amount, 0) AS deposits_amount,
		COALESCE(participated_sync, 0) AS participated_sync,
		COALESCE(missed_sync, 0) AS missed_sync,
		COALESCE(orphaned_sync, 0) AS orphaned_sync,
		COALESCE(cl_rewards_gwei, 0) AS cl_rewards_gwei
	FROM validator_stats
	WHERE validatorindex = $1
	ORDER BY day DESC`, index)

	if err != nil {
		utils.LogError(err, "error getting validator stats history from db", 0, errFields)
		validatorNotFound(data, w, r, vars, "/stats")
		return
	}

	data.Data = validatorStatsTablePageData
	if handleTemplateError(w, r, "validator.go", "ValidatorStatsTable", "", validatorStatsTableTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// ValidatorSync retrieves one page of sync duties of a specific validator for DataTable.
func ValidatorSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	validatorIndex, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil || validatorIndex > math.MaxInt32 { // index in postgres is limited to int
		logger.Warnf("error parsing validator index: %v", err)
		http.Error(w, "Error: Invalid parameter validator index.", http.StatusBadRequest)
		return
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

	errFields := map[string]interface{}{
		"route":  r.URL.String(),
		"index":  validatorIndex,
		"draw":   draw,
		"start":  start,
		"length": length}

	// retrieve all sync periods for this validator
	var syncPeriods []uint64 = []uint64{}
	err = db.ReaderDb.Select(&syncPeriods, `
		SELECT distinct period
		FROM sync_committees 
		WHERE validatorindex = $1
		ORDER BY period desc`, validatorIndex)

	if err != nil {
		utils.LogError(err, "error getting sync tab count data of sync-assignments from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := [][]interface{}{}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    0,
		RecordsFiltered: 0,
		Data:            tableData,
	}

	if len(syncPeriods) == 0 {
		// no sync periods for this validator => early exit with empty tableData
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			utils.LogError(err, "error encoding json response", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		return
	}

	totalCount := uint64(0) // total count of sync duties for this validator
	latestProposedSlot := services.LatestProposedSlot()
	slots := make([]uint64, 0, utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod*utils.Config.Chain.ClConfig.SlotsPerEpoch*uint64(len(syncPeriods)))

	for _, period := range syncPeriods {
		firstEpoch := utils.FirstEpochOfSyncPeriod(period)
		lastEpoch := firstEpoch + utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod - 1

		firstSlot := firstEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		lastSlot := (lastEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1

		for slot := lastSlot; slot >= firstSlot && (slot <= lastSlot /* guards against underflows */); slot-- {
			if slot > latestProposedSlot || utils.EpochOfSlot(slot) < utils.Config.Chain.ClConfig.AltairForkEpoch {
				continue
			}
			slots = append(slots, slot)
		}
	}

	totalCount = uint64(len(slots))
	if start >= totalCount {
		start = totalCount - 1
	}

	startIndex := start + length - 1
	if startIndex >= totalCount {
		startIndex = totalCount - 1
	}
	endIndex := start

	// retrieve sync duties and sync participations
	syncDuties := make(map[uint64]*types.ValidatorSyncParticipation, length)
	participations := make(map[uint64]uint64, length)

	// the slot range for the given table page might contain multiple sync periods and therefore we may need to split the queries to avoid fetching potentially thousands of duties at once
	type SlotRange struct {
		StartSlot uint64
		EndSlot   uint64
	}
	var consecutiveSlotRanges []SlotRange

	// slots are sorted in descending order
	nextSlotRange := SlotRange{StartSlot: slots[endIndex], EndSlot: slots[endIndex]}
	for i := endIndex + 1; i <= startIndex; i++ {
		if slots[i] == nextSlotRange.StartSlot-1 {
			nextSlotRange.StartSlot = slots[i]
		} else {
			consecutiveSlotRanges = append(consecutiveSlotRanges, nextSlotRange)
			nextSlotRange = SlotRange{StartSlot: slots[i], EndSlot: slots[i]}
		}
	}
	consecutiveSlotRanges = append(consecutiveSlotRanges, nextSlotRange)

	// make individual queries for each consecutive slot range and accumulate results
	for _, slotRange := range consecutiveSlotRanges {
		sdh, err := db.BigtableClient.GetValidatorSyncDutiesHistory([]uint64{validatorIndex}, slotRange.StartSlot, slotRange.EndSlot)
		if err != nil {
			errFields["slotRange"] = slotRange
			utils.LogError(err, "error getting validator sync duties data from bigtable", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		for slot, duty := range sdh[validatorIndex] {
			syncDuties[slot] = duty
		}

		par, err := db.GetSyncParticipationBySlotRange(slotRange.StartSlot, slotRange.EndSlot)
		if err != nil {
			errFields["slotRange"] = slotRange
			utils.LogError(err, "error getting validator sync participation data from db", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		for slot, participation := range par {
			participations[slot] = participation
		}
	}

	// search for the missed slots (status = 2), to see if it was only our validator that missed the slot or if the block was missed
	slotsRange := slots[endIndex : startIndex+1]

	missedSlots := []uint64{}
	err = db.ReaderDb.Select(&missedSlots, `SELECT slot FROM blocks WHERE slot = ANY($1) AND status = '2'`, slotsRange)
	if err != nil {
		utils.LogError(err, "error getting missed slots data from db", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	missedSlotsMap := make(map[uint64]bool, len(missedSlots))
	for _, slot := range missedSlots {
		missedSlotsMap[slot] = true
	}

	// extract correct slots
	tableData = make([][]interface{}, 0, length)
	for index := endIndex; index <= startIndex; index++ {
		slot := slots[index]

		epoch := utils.EpochOfSlot(slot)
		participation := participations[slot]

		status := uint64(0)
		if syncDuties[slot] != nil {
			status = syncDuties[slot].Status
		}
		if _, ok := missedSlotsMap[slot]; ok {
			status = 3
		}

		tableData = append(tableData, []interface{}{
			fmt.Sprintf("%d", utils.SyncPeriodOfEpoch(epoch)),
			utils.FormatEpoch(epoch),
			utils.FormatBlockSlot(slot),
			utils.FormatSyncParticipationStatus(status, slot),
			utils.FormatSyncParticipations(participation),
		})
	}

	// fill and send data
	data.RecordsTotal = totalCount
	data.RecordsFiltered = totalCount
	data.Data = tableData

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// validatorNotFound will print the appropriate error message for when the requested validator cannot be found
func validatorNotFound(data *types.PageData, w http.ResponseWriter, r *http.Request, vars map[string]string, page string) {
	validatorNotFoundTemplateFiles := append(layoutTemplateFiles, "validator/validatornotfound.html")
	var validatorNotFoundTemplate = templates.GetTemplate(validatorNotFoundTemplateFiles...)

	SetPageDataTitle(data, "Validator not found")
	d := InitPageData(w, r, "validators", fmt.Sprintf("/validator/%v%v", vars["index"], page), "", validatorNotFoundTemplateFiles)

	err := handleTemplateError(w, r, "validator.go", "Validator", "GetValidatorDeposits", validatorNotFoundTemplate.ExecuteTemplate(w, "layout", d))
	if err != nil {
		return // an error has occurred and was processed
	}
}
