package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
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

	//start := time.Now()
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

	latestEpoch := services.LatestEpoch()
	lastFinalizedEpoch := services.LatestFinalizedEpoch()

	validatorPageData := types.ValidatorPageData{}

	validatorPageData.CappellaHasHappened = latestEpoch >= (utils.Config.Chain.Config.CappellaForkEpoch)
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
	validatorPageData.InclusionDelay = int64((utils.Config.Chain.Config.Eth1FollowDistance*utils.Config.Chain.Config.SecondsPerEth1Block+utils.Config.Chain.Config.SecondsPerSlot*utils.Config.Chain.Config.SlotsPerEpoch*utils.Config.Chain.Config.EpochsPerEth1VotingPeriod)/3600) + 1

	data := InitPageData(w, r, "validators", "/validators", "", validatorTemplateFiles)
	validatorPageData.NetworkStats = services.LatestIndexPageData()
	validatorPageData.User = data.User

	validatorPageData.FlashMessage, err = utils.GetFlash(w, r, validatorEditFlash)
	if err != nil {
		logger.Errorf("error retrieving flashes for validator %v: %v", vars["index"], err)
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
		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			// the validator might only have a public key but no index yet
			var name string
			err := db.ReaderDb.Get(&name, `SELECT name FROM validator_names WHERE publickey = $1`, pubKey)
			if err != nil && err != sql.ErrNoRows {
				logger.Errorf("error getting validator-name from db for pubKey %v: %v", pubKey, err)
				validatorNotFound(data, w, r, vars, "")
				return
				// err == sql.ErrNoRows -> unnamed
			} else {
				validatorPageData.Name = name
			}

			var pool string
			err = db.ReaderDb.Get(&pool, `SELECT pool FROM validator_pool WHERE publickey = $1`, pubKey)
			if err != nil && err != sql.ErrNoRows {
				logger.Errorf("error getting validator-pool from db for pubKey %v: %v", pubKey, err)
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
				logger.Errorf("error getting validator-deposits from db: %v", err)
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
				logger.Errorf("error getting tagged validators from db: %v", err)
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
		if err != nil {
			validatorNotFound(data, w, r, vars, "")
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
		LEFT JOIN (SELECT MAX(validatorindex)+1 FROM validator_performance WHERE validatorindex >= 0) validator_performance_count(total_count) ON true
		WHERE validators.validatorindex = $1`, index)

	if err == sql.ErrNoRows {
		validatorNotFound(data, w, r, vars, "")
		return
	} else if err != nil {
		logger.Errorf("error getting validator for %v route: %v", r.URL.String(), err)
		validatorNotFound(data, w, r, vars, "")
		return
	}

	lastAttestationSlots, err := db.BigtableClient.GetLastAttestationSlots([]uint64{index})
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator last attestation slot data")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	validatorPageData.LastAttestationSlot = lastAttestationSlots[index]

	lastStatsDay := services.LatestExportedStatisticDay()

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

	// Every validator is scheduled to issue an attestation once per epoch
	// Hence we can calculate the number of attestations using the current epoch and the activation epoch
	// Special care needs to be take for exited and pending validators
	validatorPageData.AttestationsCount = validatorPageData.Epoch - validatorPageData.ActivationEpoch + 1
	if validatorPageData.ActivationEpoch > validatorPageData.Epoch {
		validatorPageData.AttestationsCount = 0
	}

	if validatorPageData.ExitEpoch != 9223372036854775807 && validatorPageData.ExitEpoch <= validatorPageData.Epoch {
		validatorPageData.AttestationsCount = validatorPageData.ExitEpoch - validatorPageData.ActivationEpoch
	}

	avgSyncInterval := uint64(getAvgSyncCommitteeInterval(1))
	avgSyncIntervalAsDuration := time.Duration(
		utils.Config.Chain.Config.SecondsPerSlot*
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

		validatorPageData.IncomeHistoryChartData, err = db.GetValidatorIncomeHistoryChart([]uint64{index}, currency, lastFinalizedEpoch, lowerBoundDay)

		if err != nil {
			return fmt.Errorf("error calling db.GetValidatorIncomeHistoryChart: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		start := time.Now()
		defer func() {
			timings.Charts = time.Since(start)
		}()
		validatorPageData.ExecutionIncomeHistoryData, err = getExecutionChartData([]uint64{index}, currency, lowerBoundDay)

		if err != nil {
			return fmt.Errorf("error calling getExecutionChartData: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		// those functions need to be executed sequentially as both require the CurrentBalance value
		start := time.Now()
		defer func() {
			timings.Earnings = time.Since(start)
		}()
		earnings, balances, err := GetValidatorEarnings([]uint64{index}, GetCurrency(r))
		if err != nil {
			return fmt.Errorf("error retrieving validator earnings: %v", err)
		}
		// each income and apr variable is a struct of 3 fields: cl, el and total
		validatorPageData.Income1d = earnings.Income1d
		validatorPageData.Income7d = earnings.Income7d
		validatorPageData.Income31d = earnings.Income31d
		validatorPageData.Apr7d = earnings.Apr7d
		validatorPageData.Apr31d = earnings.Apr31d
		validatorPageData.Apr365d = earnings.Apr365d
		validatorPageData.IncomeTotal = earnings.IncomeTotal
		validatorPageData.IncomeTotalFormatted = earnings.TotalFormatted
		validatorPageData.IncomeToday = earnings.IncomeToday
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

			// get validator withdrawals
			withdrawalsCount, lastWithdrawalsEpoch, err := db.GetValidatorWithdrawalsCount(validatorPageData.Index)
			if err != nil {
				return fmt.Errorf("error getting validator withdrawals count from db: %v", err)
			}
			validatorPageData.WithdrawalCount = withdrawalsCount

			blsChange, err := db.GetValidatorBLSChange(validatorPageData.Index)
			if err != nil {
				return fmt.Errorf("error getting validator bls change from db: %v", err)
			}
			validatorPageData.BLSChange = blsChange

			if bytes.Equal(validatorPageData.WithdrawCredentials[:1], []byte{0x00}) && blsChange != nil {
				// blsChanges are only possible afters cappeala
				validatorPageData.IsWithdrawableAddress = true
			}

			// only calculate the expected next withdrawal if the validator is eligible
			isFullWithdrawal := validatorPageData.CurrentBalance > 0 && validatorPageData.WithdrawableEpoch <= validatorPageData.Epoch
			isPartialWithdrawal := validatorPageData.EffectiveBalance == utils.Config.Chain.Config.MaxEffectiveBalance && validatorPageData.CurrentBalance > utils.Config.Chain.Config.MaxEffectiveBalance
			if stats != nil && stats.LatestValidatorWithdrawalIndex != nil && stats.TotalValidatorCount != nil && validatorPageData.IsWithdrawableAddress && (isFullWithdrawal || isPartialWithdrawal) {
				distance, err := GetWithdrawableCountFromCursor(validatorPageData.Epoch, validatorPageData.Index, *stats.LatestValidatorWithdrawalIndex)
				if err != nil {
					return fmt.Errorf("error getting withdrawable validator count from cursor: %v", err)
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

					var withdrawalAmont uint64
					if isFullWithdrawal {
						withdrawalAmont = validatorPageData.CurrentBalance
					} else {
						withdrawalAmont = validatorPageData.CurrentBalance - utils.Config.Chain.Config.MaxEffectiveBalance
					}

					if latestEpoch == lastWithdrawalsEpoch {
						withdrawalAmont = 0
					}
					tableData = append(tableData, []interface{}{
						template.HTML(fmt.Sprintf(`<span class="text-muted">~ %s</span>`, utils.FormatEpoch(uint64(utils.TimeToEpoch(timeToWithdrawal))))),
						template.HTML(fmt.Sprintf(`<span class="text-muted">~ %s</span>`, utils.FormatBlockSlot(utils.TimeToSlot(uint64(timeToWithdrawal.Unix()))))),
						template.HTML(fmt.Sprintf(`<span class="">~ %s</span>`, utils.FormatTimestamp(timeToWithdrawal.Unix()))),
						withdrawalCredentialsTemplate,
						template.HTML(fmt.Sprintf(`<span class="text-muted"><span data-toggle="tooltip" title="If the withdrawal were to be processed at this very moment, this amount would be withdrawn"><i class="far ml-1 fa-question-circle" style="margin-left: 0px !important;"></i></span> %s</span>`, utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(withdrawalAmont), big.NewInt(1e9)), "Ether", 6))),
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
			return fmt.Errorf("error getting tagged validators from db: %v", err)
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
			return fmt.Errorf("error getting validator-deposits from db: %v", err)
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
				return fmt.Errorf("failed to retrieve queue ahead of validator %v: %v", validatorPageData.ValidatorIndex, err)
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
		return nil
	})

	g.Go(func() error {
		if validatorPageData.AttestationsCount > 0 {
			// get attestationStats from validator_stats
			attestationStats := struct {
				MissedAttestations uint64 `db:"missed_attestations"`
			}{}
			if lastStatsDay > 0 {
				err = db.ReaderDb.Get(&attestationStats, "SELECT missed_attestations_total AS missed_attestations FROM validator_stats WHERE validatorindex = $1 AND day = $2", index, lastStatsDay)
				if err != nil {
					return fmt.Errorf("error retrieving validator attestationStats: %w", err)
				}
			}

			// add attestationStats that are not yet in validator_stats
			lookback := int64(lastFinalizedEpoch - (lastStatsDay+1)*utils.EpochsPerDay())
			if lookback > 0 {
				// logger.Infof("retrieving attestations not yet in stats, lookback is %v", lookback)
				missedAttestations, err := db.BigtableClient.GetValidatorMissedAttestationHistory([]uint64{index}, lastFinalizedEpoch-uint64(lookback), lastFinalizedEpoch)
				if err != nil {
					return fmt.Errorf("error retrieving validator attestations not in stats from bigtable: %v", err)
				}
				attestationStats.MissedAttestations += uint64(len(missedAttestations[index]))
			}

			validatorPageData.MissedAttestationsCount = attestationStats.MissedAttestations
			validatorPageData.ExecutedAttestationsCount = validatorPageData.AttestationsCount - validatorPageData.MissedAttestationsCount
			validatorPageData.UnmissedAttestationsPercentage = float64(validatorPageData.ExecutedAttestationsCount) / float64(validatorPageData.AttestationsCount)
		}
		return nil
	})

	g.Go(func() error {
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
				return fmt.Errorf("error retrieving validator slashing info: %v", err)
			}
			validatorPageData.SlashedBy = slashingInfo.Slasher
			validatorPageData.SlashedAt = slashingInfo.Slot
			validatorPageData.SlashedFor = slashingInfo.Reason
		}

		err = db.ReaderDb.Get(&validatorPageData.SlashingsCount, `select COALESCE(sum(attesterslashingscount) + sum(proposerslashingscount), 0) from blocks where blocks.proposer = $1 and blocks.status = '1'`, index)
		if err != nil {
			return fmt.Errorf("error retrieving slashings-count: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		eff, err := db.BigtableClient.GetValidatorEffectiveness([]uint64{index}, validatorPageData.Epoch-1)
		if err != nil {
			return fmt.Errorf("error retrieving validator effectiveness: %v", err)
		}
		if len(eff) > 1 {
			return fmt.Errorf("error retrieving validator effectiveness: invalid length %v", len(eff))
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

		err = db.ReaderDb.Select(&allSyncPeriods, `
		SELECT period as period, (period*$1) as firstepoch, ((period+1)*$1)-1 as lastepoch
		FROM sync_committees 
		WHERE validatorindex = $2
		ORDER BY period desc`, utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod, index)
		if err != nil {
			return fmt.Errorf("error getting sync participation count data of sync-assignments: %v", err)
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
						COALESCE(SUM(participated_sync), 0) as participated_sync,
						COALESCE(SUM(missed_sync), 0) as missed_sync,
						COALESCE(SUM(orphaned_sync), 0) as orphaned_sync
					FROM validator_stats
					WHERE validatorindex = $1`, index)
				if err != nil {
					return fmt.Errorf("error retrieving validator syncStats: %v", err)
				}
			}

			// if sync duties of last period haven't fully been exported yet, fetch remaining duties from bigtable
			lastExportedEpoch := (lastStatsDay+1)*utils.EpochsPerDay() - 1
			lastSyncPeriod := actualSyncPeriods[0]
			if lastSyncPeriod.LastEpoch > lastExportedEpoch {
				res, err := db.BigtableClient.GetValidatorSyncDutiesHistory([]uint64{index}, lastExportedEpoch+1, latestEpoch)
				if err != nil {
					return fmt.Errorf("error retrieving validator sync participations data from bigtable: %v", err)
				}
				syncStatsBt := utils.AddSyncStats([]uint64{index}, res, nil)
				// if last sync period is the current one, add remaining scheduled slots
				if lastSyncPeriod.LastEpoch >= latestEpoch {
					syncStatsBt.ScheduledSlots += utils.GetRemainingScheduledSync(1, syncStatsBt, lastExportedEpoch, lastSyncPeriod.FirstEpoch)
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
			expectedSyncCount, err := getExpectedSyncCommitteeSlots([]uint64{index}, latestEpoch)
			if err != nil {
				return fmt.Errorf("error retrieving expected sync committee slots: %v", err)
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
		err = db.ReaderDb.Get(validatorPageData.Rocketpool, `
		SELECT
			rplm.node_address      AS node_address,
			rplm.address           AS minipool_address,
			rplm.node_fee          AS minipool_node_fee,
			rplm.deposit_type      AS minipool_deposit_type,
			rplm.status            AS minipool_status,
			rplm.status_time       AS minipool_status_time,
			COALESCE(rplm.penalty_count,0)     AS penalty_count,
			rpln.timezone_location AS node_timezone_location,
			rpln.rpl_stake         AS node_rpl_stake,
			rpln.max_rpl_stake     AS node_max_rpl_stake,
			rpln.min_rpl_stake     AS node_min_rpl_stake,
			rpln.rpl_cumulative_rewards     AS rpl_cumulative_rewards,
			rpln.claimed_smoothing_pool     AS claimed_smoothing_pool,
			rpln.unclaimed_smoothing_pool   AS unclaimed_smoothing_pool,
			rpln.unclaimed_rpl_rewards      AS unclaimed_rpl_rewards,
			COALESCE(node_deposit_balance, 0) AS node_deposit_balance,
			COALESCE(node_refund_balance, 0) AS node_refund_balance,
			COALESCE(user_deposit_balance, 0) AS user_deposit_balance,
			COALESCE(rpln.effective_rpl_stake, 0) as effective_rpl_stake,
			COALESCE(deposit_credit, 0) AS deposit_credit,
			COALESCE(is_vacant, false) AS is_vacant,
			version,
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
			return fmt.Errorf("error getting rocketpool-data for validator for %v route: %v", r.URL.String(), err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		logger.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	validatorPageData.FutureDutiesEpoch = protomath.MaxU64(futureProposalEpoch, futureSyncDutyEpoch)
	validatorPageData.IncomeToday.Total = validatorPageData.IncomeToday.Cl + validatorPageData.IncomeToday.El

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
			blocks.withdrawalcount, 
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
	if ae.ExitEpoch != 9223372036854775807 && ae.ExitEpoch <= epoch {
		lastAttestationEpoch = ae.ExitEpoch - 1
		totalCount = ae.ExitEpoch - ae.ActivationEpoch
	}

	tableData := [][]interface{}{}

	if totalCount > 0 {
		endEpoch := uint64(int64(lastAttestationEpoch) - start)
		attestationData, err := db.BigtableClient.GetValidatorAttestationHistory([]uint64{index}, endEpoch-uint64(length)+1, endEpoch)
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

// ValidatorWithdrawals returns a validators withdrawals in json
func ValidatorWithdrawals(w http.ResponseWriter, r *http.Request) {
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

	length := uint64(10)

	withdrawalCount, _, err := db.GetValidatorWithdrawalsCount(index)
	if err != nil {
		logger.Errorf("error retrieving validator withdrawals count: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	withdrawals, err := db.GetValidatorWithdrawals(index, length, start, orderBy, orderDir)
	if err != nil {
		logger.Errorf("error retrieving validator withdrawals: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, 0, len(withdrawals))

	for _, w := range withdrawals {
		tableData = append(tableData, []interface{}{
			template.HTML(fmt.Sprintf("%v", utils.FormatEpoch(utils.EpochOfSlot(w.Slot)))),
			template.HTML(fmt.Sprintf("%v", utils.FormatBlockSlot(w.Slot))),
			template.HTML(fmt.Sprintf("%v", utils.FormatTimestamp(utils.SlotToTime(w.Slot).Unix()))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAddress(w.Address, nil, "", false, false, true))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), "Ether", 6))),
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

/*
Function checks if the generated ECDSA signature has correct lentgth and if needed sets recovery byte to 0 or 1
*/
func sanitizeSignature(sig string) ([]byte, error) {
	sig = strings.Replace(sig, "0x", "", -1)
	decodedSig, _ := hex.DecodeString(sig)
	if len(decodedSig) != 65 {
		return nil, errors.New("signature is less then 65 bytes")
	}
	if decodedSig[crypto.RecoveryIDOffset] == 27 || decodedSig[crypto.RecoveryIDOffset] == 28 {
		decodedSig[crypto.RecoveryIDOffset] -= 27
	}
	return []byte(decodedSig), nil
}

/*
Function tries to find the substring.
If successful it turns string into []byte value and returns it
If it fails, it will try to decode `msg`value from Hexadecimal to string and retry search again
*/
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
		return nil, errors.New("beachoncha.in was not found")

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
	pageLength := 10
	maxPages := 10

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
	if activationAndExitEpoch.ExitEpoch != 9223372036854775807 && activationAndExitEpoch.ExitEpoch <= services.LatestFinalizedEpoch() {
		totalCount += activationAndExitEpoch.ExitEpoch - activationAndExitEpoch.ActivationEpoch
	} else {
		totalCount += services.LatestFinalizedEpoch() - activationAndExitEpoch.ActivationEpoch + 1
	}

	if start > uint64((maxPages-1)*pageLength) {
		start = uint64((maxPages - 1) * pageLength)
	}

	currentEpoch := services.LatestEpoch() - 1
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
			logger.Errorf("error retrieving lastActionDay for validator-history: %v", err)
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
			tableData = append(tableData, icomeToTableData(i, incomeDetails[index][i], withdrawalMap[i], currency))
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
			tableData = append(tableData, icomeToTableData(epoch, incomeDetails[index][epoch], withdrawalMap[epoch], currency))
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
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func getWithdrawalAndIncome(index uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*types.ValidatorWithdrawal, map[uint64]map[uint64]*itypes.ValidatorEpochIncome, error) {
	g := new(errgroup.Group)

	var withdrawals []*types.WithdrawalsByEpoch
	g.Go(func() error {
		var err error
		withdrawals, err = db.GetValidatorsWithdrawalsByEpoch([]uint64{index}, startEpoch, endEpoch)
		if err != nil {
			logger.Errorf("error retrieving validator withdrawals by epoch: %v", err)
			return err
		}
		return nil
	})

	var incomeDetails map[uint64]map[uint64]*itypes.ValidatorEpochIncome
	g.Go(func() error {
		var err error
		incomeDetails, err = db.BigtableClient.GetValidatorIncomeDetailsHistory([]uint64{index}, startEpoch, endEpoch)
		if err != nil {
			logger.Errorf("error retrieving validator income details history from bigtable: %v", err)
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
			Slot:   withdrawals.Epoch * utils.Config.Chain.Config.SlotsPerEpoch,
		}
	}
	return withdrawalMap, incomeDetails, err
}

func icomeToTableData(epoch uint64, income *itypes.ValidatorEpochIncome, withdrawal *types.ValidatorWithdrawal, currency string) []interface{} {
	events := template.HTML("")
	if income.AttestationSourcePenalty > 0 && income.AttestationTargetPenalty > 0 {
		events += utils.FormatAttestationStatusShort(2)
	} else {
		events += utils.FormatAttestationStatusShort(1)
	}

	if income.ProposerAttestationInclusionReward > 0 {
		block := utils.FormatBlockStatusShort(1)
		events += block
	} else if income.ProposalsMissed > 0 {
		block := utils.FormatBlockStatusShort(2)
		events += block
	}

	if withdrawal != nil {
		withdrawal := utils.FormatWithdrawalShort(uint64(withdrawal.Slot), withdrawal.Amount)
		events += withdrawal
	}

	rewards := income.TotalClRewards()
	return []interface{}{
		utils.FormatEpoch(epoch),
		utils.FormatBalanceChangeFormated(&rewards, currency, income),
		template.HTML(""),
		template.HTML(events),
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

	data := InitPageData(w, r, "validators", "/validators", "", templateFiles)

	// Request came with a hash
	if strings.Contains(vars["index"], "0x") || len(vars["index"]) == 96 {
		pubKey, err := hex.DecodeString(strings.Replace(vars["index"], "0x", "", -1))
		if err != nil {
			logger.Errorf("error parsing validator public key %v: %v", vars["index"], err)

			validatorNotFound(data, w, r, vars, "/stats")

			return
		}
		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			logger.Errorf("error parsing validator pubkey: %v", err)
			validatorNotFound(data, w, r, vars, "/stats")
			return
		}
	} else {
		// Request came with a validator index number
		index, err = strconv.ParseUint(vars["index"], 10, 64)
		// Request is not a valid index number
		if err != nil {
			logger.Errorf("error parsing validator index: %v", err)
			validatorNotFound(data, w, r, vars, "/stats")
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
	FROM validator_stats WHERE validatorindex = $1 ORDER BY day DESC`, index)

	if err != nil {
		logger.Errorf("error retrieving validator stats history: %v", err)
		validatorNotFound(data, w, r, vars, "/stats")
		return
	}

	// if validatorStatsTablePageData.Rows[len(validatorStatsTablePageData.Rows)-1].Day == -1 {
	// 	validatorStatsTablePageData.Rows = validatorStatsTablePageData.Rows[:len(validatorStatsTablePageData.Rows)-1]
	// }

	data.Data = validatorStatsTablePageData
	if handleTemplateError(w, r, "validator.go", "ValidatorStatsTable", "", validatorStatsTableTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// ValidatorSync retrieves one page of sync duties of a specific validator for DataTable.
func ValidatorSync(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	validatorIndex, err := strconv.ParseUint(vars["index"], 10, 64)
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
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length length-parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if length > 100 {
		length = 100
	}
	descOrdering := q.Get("order[0][dir]") == "desc"

	// retrieve all sync periods for this validator
	var syncPeriods []uint64 = []uint64{}

	err = db.ReaderDb.Select(&syncPeriods, `
		SELECT period
		FROM sync_committees 
		WHERE validatorindex = $1
		ORDER BY period asc`, validatorIndex)
	if err != nil {
		logger.WithError(err).Errorf("error getting sync tab count data of sync-assignments")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	latestEpoch := services.LatestEpoch()

	var syncSlots []uint64 = []uint64{}
	// getting all sync slots until the latestEpoch (so we exclude scheduled)
	for _, syncPeriod := range syncPeriods {
		startEpoch := syncPeriod * utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod
		endEpoch := (syncPeriod + 1) * utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod
		for epoch := startEpoch; epoch < endEpoch; epoch++ {
			if epoch <= latestEpoch {
				for slot := epoch * utils.Config.Chain.Config.SlotsPerEpoch; slot < (epoch+1)*utils.Config.Chain.Config.SlotsPerEpoch; slot++ {
					syncSlots = append(syncSlots, slot)
				}
			}
		}
	}

	// total count of sync duties for this validator
	totalCount := uint64(len(syncSlots))

	tableData := [][]interface{}{}

	if totalCount > 0 && start <= totalCount {

		// if ordering is desc, reverse sync slots
		if descOrdering {
			utils.ReverseSlice(syncSlots)
		}

		// let's pick the slots we want to display on the requested page
		pageSlots := syncSlots[start:protomath.MinU64(start+length, totalCount)]
		length = uint64(len(pageSlots)) // Last page might be shorter
		epochs := []uint64{}
		last := uint64(0)
		// we gather all epochs for that page (the table currently displays 10 items per page, so it displays max 2 epoches. But we want to be future proof, so we support bigger page sizes)
		for _, slot := range pageSlots {
			epoch := slot / utils.Config.Chain.Config.SlotsPerEpoch
			if len(epochs) == 0 || epoch != last {
				epochs = append(epochs, epoch)
				last = epoch
			}
		}

		type PageSlot struct {
			Status        uint64
			Participation uint64
		}
		pageSlotsMap := make(map[uint64]PageSlot)
		for i := 0; i < len(pageSlots); i++ {
			pageSlotsMap[pageSlots[i]] = PageSlot{}
		}
		var missedSyncSlots []uint64

		mux := sync.Mutex{}
		g, gCtx := errgroup.WithContext(ctx)

		// load the sync duties for all epochs involved from bt
		for i := 0; i < len(epochs); i++ {
			epoch := epochs[i]
			g.Go(func() error {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
				}

				syncDuties, err := db.BigtableClient.GetValidatorSyncDutiesHistory([]uint64{}, epoch, epoch)
				if err != nil {
					return fmt.Errorf("error retrieving validator [%v] sync duty data from bigtable for epoch [%v]: %w", validatorIndex, epoch, err)
				} else if uint64(len(syncDuties[validatorIndex]))%utils.Config.Chain.Config.SlotsPerEpoch > 0 {
					return fmt.Errorf("wrong number [%v] of syncDuties for validator [%v] received from from bigtable for epoch [%v]", len(syncDuties[validatorIndex]), validatorIndex, epoch)
				}

				for validator, duties := range syncDuties {
					for _, duty := range duties {
						mux.Lock()
						pageSlot, ok := pageSlotsMap[duty.Slot]
						if !ok {
							mux.Unlock()
							continue // this is not the slot we are looking for
						}

						// We want to know how many validators successfully participated in this duty
						if duty.Status == 1 {
							pageSlot.Participation++
						}

						// Get the status if the duty belongs to our validator
						if validator == validatorIndex {
							slotTime := utils.SlotToTime(duty.Slot)
							if duty.Status == 0 && time.Since(slotTime) <= time.Minute {
								duty.Status = 2 // scheduled
							}

							if duty.Status == 0 {
								missedSyncSlots = append(missedSyncSlots, duty.Slot)
							}
							pageSlot.Status = duty.Status
						}
						pageSlotsMap[duty.Slot] = pageSlot
						mux.Unlock()
					}
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			utils.LogError(err, "", 0)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Search for the missed slots (status = 2), to see if it was only our validator that missed the slot or if the block was missed
		missedSlotsMap, err := db.GetMissedSlotsMap(missedSyncSlots)
		if err != nil {
			logger.WithError(err).Errorf("error getting missed slots data")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// extract correct slots
		tableData = make([][]interface{}, length)
		for i := 0; i < len(pageSlots); i++ {
			slot := pageSlots[i]
			pageSlot := pageSlotsMap[slot]
			participation := pageSlot.Participation

			epoch := utils.EpochOfSlot(slot)
			status := pageSlot.Status
			if _, ok := missedSlotsMap[slot]; ok {
				status = 3
			}

			tableData[i] = []interface{}{
				fmt.Sprintf("%d", utils.SyncPeriodOfEpoch(epoch)),
				utils.FormatEpoch(epoch),
				utils.FormatBlockSlot(slot),
				utils.FormatSyncParticipationStatus(status, slot),
				utils.FormatSyncParticipations(participation),
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
