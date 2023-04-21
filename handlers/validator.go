package handlers

import (
	"bytes"
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
	"math"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
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
	validatorNotFoundTemplateFiles := append(layoutTemplateFiles, "validator/validatornotfound.html")
	var validatorTemplate = templates.GetTemplate(validatorTemplateFiles...)
	var validatorNotFoundTemplate = templates.GetTemplate(validatorNotFoundTemplateFiles...)

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

	// Request came with a hash
	if strings.Contains(vars["index"], "0x") || len(vars["index"]) == 96 {
		pubKey, err := hex.DecodeString(strings.Replace(vars["index"], "0x", "", -1))
		if err != nil {
			// logger.Errorf("error parsing validator public key %v: %v", vars["index"], err)
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
			validatorPageData.ShowMultipleWithdrawalCredentialsWarning = hasMultipleWithdrawalCredentials(deposits)
			if err != nil || len(deposits.Eth1Deposits) == 0 {
				SetPageDataTitle(data, fmt.Sprintf("Validator %x", pubKey))
				data := InitPageData(w, r, "validators", fmt.Sprintf("/validator/%v", index), "", validatorNotFoundTemplateFiles)

				if handleTemplateError(w, r, "validator.go", "Validator", "GetValidatorDeposits", validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
					return // an error has occurred and was processed
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
			} else if time.Unix(latestDeposit, 0).Before(utils.SlotToTime(0)) {
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
			validators.slashed,
			validators.activationeligibilityepoch,
			validators.activationepoch,
			validators.exitepoch,
			validators.lastattestationslot,
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
		data := InitPageData(w, r, "validators", fmt.Sprintf("/validator/%v", index), "", validatorNotFoundTemplateFiles)
		if handleTemplateError(w, r, "validator.go", "Validator", "no rows", validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	} else if err != nil {
		logger.Errorf("error getting validator for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var lastStatsDay uint64
	err = db.ReaderDb.Get(&lastStatsDay, "SELECT COALESCE(MAX(day),0) FROM validator_stats_status WHERE status")
	if err != nil {
		logger.Errorf("error getting lastStatsDay for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

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
			utils.Config.Chain.Config.SlotsPerEpoch*
			utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod*
			avgSyncInterval) * time.Second
	validatorPageData.AvgSyncInterval = &avgSyncIntervalAsDuration

	g := errgroup.Group{}
	g.Go(func() error {
		start := time.Now()
		defer func() {
			timings.Charts = time.Since(start)
		}()

		validatorPageData.IncomeHistoryChartData, validatorPageData.IncomeToday.Cl, err = db.GetValidatorIncomeHistoryChart([]uint64{index}, currency)

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
		validatorPageData.ExecutionIncomeHistoryData, err = getExecutionChartData([]uint64{index}, currency)

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
		if utils.Config.Frontend.Validator.ShowProposerRewards {
			validatorPageData.IncomeProposerFormatted = &earnings.ProposerTotalFormatted
		}

		vbalance, ok := balances[validatorPageData.ValidatorIndex]
		if !ok {
			return fmt.Errorf("error retrieving validator balances: %v", err)
		}
		validatorPageData.CurrentBalance = vbalance.Balance
		validatorPageData.EffectiveBalance = vbalance.EffectiveBalance

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
						template.HTML(fmt.Sprintf(`<span class="">~ %s</span>`, utils.FormatTimeFromNow(timeToWithdrawal))),
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
		if validatorPageData.ActivationEpoch > 100_000_000 {
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
		proposals := []struct {
			Slot            uint64 `db:"slot"`
			Status          uint64 `db:"status"`
			ExecBlockNumber uint64 `db:"exec_block_number"`
		}{}

		err = db.ReaderDb.Select(&proposals, `
			SELECT 
				slot, 
				status, 
				COALESCE(exec_block_number, 0) as exec_block_number
			FROM blocks 
			WHERE proposer = $1
			ORDER BY slot ASC`, index)
		if err != nil {
			return fmt.Errorf("error retrieving block-proposals: %v", err)
		}

		proposedToday := []uint64{}
		todayStartEpoch := uint64(lastStatsDay+1) * utils.EpochsPerDay()
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
				// add to list of blocks proposed today if epoch hasn't been exported into stats yet
				if utils.EpochOfSlot(b.Slot) >= todayStartEpoch && b.ExecBlockNumber > 0 {
					proposedToday = append(proposedToday, b.ExecBlockNumber)
				}
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

		var slots []uint64
		for _, p := range proposals {
			if p.ExecBlockNumber > 0 {
				slots = append(slots, p.Slot)
			}
		}

		lookbackAmount := getProposalLuckBlockLookbackAmount(1)
		startPeriod := len(slots) - lookbackAmount
		if startPeriod < 0 {
			startPeriod = 0
		}

		validatorPageData.ProposalLuck = getProposalLuck(slots[startPeriod:], 1)
		avgSlotInterval := uint64(getAvgSlotInterval(1))
		avgSlotIntervalAsDuration := time.Duration(utils.Config.Chain.Config.SecondsPerSlot*avgSlotInterval) * time.Second
		validatorPageData.AvgSlotInterval = &avgSlotIntervalAsDuration
		if len(slots) > 0 {
			nextSlotEstimate := utils.SlotToTime(slots[len(slots)-1] + avgSlotInterval)
			validatorPageData.ProposalEstimate = &nextSlotEstimate
		}

		if len(proposedToday) > 0 {
			// get el data
			execBlocks, err := db.BigtableClient.GetBlocksIndexedMultiple(proposedToday, 10000)
			if err != nil {
				return fmt.Errorf("error retrieving execution blocks data from bigtable: %v", err)
			}

			// get mev data
			relaysData, err := db.GetRelayDataForIndexedBlocks(execBlocks)
			if err != nil {
				return fmt.Errorf("error retrieving mev bribe data: %v", err)
			}

			incomeTodayEl := new(big.Int)
			for _, execBlock := range execBlocks {

				blockEpoch := utils.TimeToEpoch(execBlock.Time.AsTime())
				if blockEpoch > int64(lastFinalizedEpoch) {
					continue
				}
				// add mev bribe if present
				if relaysDatum, hasMevBribes := relaysData[common.BytesToHash(execBlock.Hash)]; hasMevBribes {
					incomeTodayEl = new(big.Int).Add(incomeTodayEl, relaysDatum.MevBribe.Int)
				} else {
					incomeTodayEl = new(big.Int).Add(incomeTodayEl, new(big.Int).SetBytes(execBlock.GetTxReward()))
				}
			}
			validatorPageData.IncomeToday.El = incomeTodayEl.Int64() / 1e9
		}

		return nil
	})

	g.Go(func() error {
		if validatorPageData.AttestationsCount > 0 {
			// get attestationStats from validator_stats
			attestationStats := struct {
				MissedAttestations   uint64 `db:"missed_attestations"`
				OrphanedAttestations uint64 `db:"orphaned_attestations"`
			}{}
			if lastStatsDay > 0 {
				err = db.ReaderDb.Get(&attestationStats, "select coalesce(sum(missed_attestations), 0) as missed_attestations, coalesce(sum(orphaned_attestations), 0) as orphaned_attestations from validator_stats where validatorindex = $1", index)
				if err != nil {
					return fmt.Errorf("error retrieving validator attestationStats: %v", err)
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
			validatorPageData.OrphanedAttestationsCount = attestationStats.OrphanedAttestations
			validatorPageData.ExecutedAttestationsCount = validatorPageData.AttestationsCount - validatorPageData.MissedAttestationsCount - validatorPageData.OrphanedAttestationsCount
			validatorPageData.UnmissedAttestationsPercentage = float64(validatorPageData.AttestationsCount-validatorPageData.MissedAttestationsCount) / float64(validatorPageData.AttestationsCount)
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
		validatorPageData.SlotsPerSyncCommittee = utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod * utils.Config.Chain.Config.SlotsPerEpoch

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

		// remove scheduled committees
		for i, syncPeriod := range allSyncPeriods {
			if syncPeriod.FirstEpoch <= latestEpoch {
				actualSyncPeriods = allSyncPeriods[i:]
				break
			}
		}

		if len(actualSyncPeriods) > 0 {
			// get sync stats from validator_stats
			syncStats := struct {
				ParticipatedSync uint64 `db:"participated_sync"`
				MissedSync       uint64 `db:"missed_sync"`
				OrphanedSync     uint64 `db:"orphaned_sync"`
				ScheduledSync    uint64
			}{}
			if lastStatsDay > 0 {
				err = db.ReaderDb.Get(&syncStats, "select coalesce(sum(participated_sync), 0) as participated_sync, coalesce(sum(missed_sync), 0) as missed_sync, coalesce(sum(orphaned_sync), 0) as orphaned_sync from validator_stats where validatorindex = $1", index)
				if err != nil {
					return fmt.Errorf("error retrieving validator syncStats: %v", err)
				}
			}

			// if sync duties of last period haven't fully been exported yet, fetch remaining duties from bigtable
			lastExportedEpoch := (lastStatsDay+1)*utils.EpochsPerDay() - 1
			if actualSyncPeriods[0].LastEpoch > lastExportedEpoch {
				lookback := int64(latestEpoch - lastExportedEpoch)
				syncStatsBt, err := db.BigtableClient.GetValidatorSyncCommitteesStats([]uint64{index}, latestEpoch-uint64(lookback), latestEpoch)
				if err != nil {
					return fmt.Errorf("error retrieving validator sync participations data from bigtable: %v", err)
				}
				syncStats.MissedSync += syncStatsBt.MissedSlots
				syncStats.ParticipatedSync += syncStatsBt.ParticipatedSlots
				syncStats.ScheduledSync += syncStatsBt.ScheduledSlots
			}

			validatorPageData.SlotsDoneInCurrentSyncCommittee = validatorPageData.SlotsPerSyncCommittee - syncStats.ScheduledSync

			validatorPageData.ParticipatedSyncCountSlots = syncStats.ParticipatedSync
			validatorPageData.MissedSyncCountSlots = syncStats.MissedSync
			validatorPageData.OrphanedSyncCountSlots = syncStats.OrphanedSync
			validatorPageData.ScheduledSyncCountSlots = syncStats.ScheduledSync
			// actual sync duty count and percentage
			syncCountSlotsIncludingScheduled := validatorPageData.ParticipatedSyncCountSlots + validatorPageData.MissedSyncCountSlots + validatorPageData.OrphanedSyncCountSlots + validatorPageData.ScheduledSyncCountSlots
			validatorPageData.SyncCount = uint64(math.Ceil(float64(syncCountSlotsIncludingScheduled) / float64(validatorPageData.SlotsPerSyncCommittee)))
			validatorPageData.UnmissedSyncPercentage = float64(validatorPageData.ParticipatedSyncCountSlots) / float64(validatorPageData.ParticipatedSyncCountSlots+validatorPageData.MissedSyncCountSlots)
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
		attestationData, err := db.BigtableClient.GetValidatorAttestationHistory([]uint64{index}, uint64(int64(lastAttestationEpoch)-start)-uint64(length), uint64(int64(lastAttestationEpoch)-start))
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
			template.HTML(fmt.Sprintf("%v", utils.FormatTimeFromNow(utils.SlotToTime(w.Slot)))),
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
		logger.Errorf("error decoding submitted signature %v: %v", signature, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	msg, err := sanitizeMessage(signatureWrapper.Msg)
	if err != nil {
		logger.Errorf("Message is invalid %v: %v", signatureWrapper.Msg, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided message is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}
	msgHash := accounts.TextHash(msg)

	sig, err := sanitizeSignature(signatureWrapper.Sig)
	if err != nil {
		logger.Errorf("error parsing submitted signature %v: %v", signatureWrapper.Sig, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, http.StatusMovedPermanently)
		return
	}

	recoveredPubkey, err := crypto.SigToPub(msgHash, sig)
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
	if activationAndExitEpoch.ExitEpoch != 9223372036854775807 && activationAndExitEpoch.ExitEpoch <= services.LatestFinalizedEpoch() {
		totalCount += activationAndExitEpoch.ExitEpoch - activationAndExitEpoch.ActivationEpoch
	} else {
		totalCount += services.LatestFinalizedEpoch() - activationAndExitEpoch.ActivationEpoch + 1
	}

	if totalCount > 100 {
		totalCount = 100
	}

	if start > 90 {
		start = 90
	}

	currentEpoch := services.LatestEpoch() - 1
	// for an exited validator we show the history until his exit
	if activationAndExitEpoch.ExitEpoch != 9223372036854775807 && currentEpoch > (activationAndExitEpoch.ExitEpoch-1) {
		currentEpoch = activationAndExitEpoch.ExitEpoch - 1
	}

	var validatorHistory []*types.ValidatorHistory

	g := new(errgroup.Group)

	startEpoch := currentEpoch - start - 9
	endEpoch := currentEpoch - start
	if startEpoch > endEpoch { // handle underflows of startEpoch
		startEpoch = 0
	}

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

	err = g.Wait()

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	withdrawalMap := make(map[uint64]*types.ValidatorWithdrawal)
	for _, withdrawals := range withdrawals {
		withdrawalMap[withdrawals.Epoch] = &types.ValidatorWithdrawal{
			Index:  withdrawals.ValidatorIndex,
			Epoch:  withdrawals.Epoch,
			Amount: withdrawals.Amount,
			Slot:   withdrawals.Epoch * utils.Config.Chain.Config.SlotsPerEpoch,
		}
	}

	tableData := make([][]interface{}, 0, len(validatorHistory))

	for i := endEpoch; i >= startEpoch; i-- {
		if incomeDetails[index] == nil || incomeDetails[index][i] == nil {
			tableData = append(tableData, []interface{}{
				utils.FormatEpoch(i),
				"pending...",
				template.HTML(""),
				template.HTML(""),
			})
			continue
		}
		events := template.HTML("")
		if incomeDetails[index][i].AttestationSourcePenalty > 0 && incomeDetails[index][i].AttestationTargetPenalty > 0 {
			events += utils.FormatAttestationStatusShort(2)
		} else {
			events += utils.FormatAttestationStatusShort(1)
		}

		if incomeDetails[index][i].ProposerAttestationInclusionReward > 0 {
			block := utils.FormatBlockStatusShort(1)
			events += block
		} else if incomeDetails[index][i].ProposalsMissed > 0 {
			block := utils.FormatBlockStatusShort(2)
			events += block
		}

		if withdrawalMap[i] != nil {
			withdrawal := utils.FormatWithdrawalShort(uint64(withdrawalMap[i].Slot), withdrawalMap[i].Amount)
			events += withdrawal
		}

		rewards := incomeDetails[index][i].TotalClRewards()
		tableData = append(tableData, []interface{}{
			utils.FormatEpoch(i),
			utils.FormatBalanceChangeFormated(&rewards, currency, incomeDetails[index][i]),
			template.HTML(""),
			template.HTML(events),
		})
	}

	if len(tableData) == 0 {
		tableData = append(tableData, []interface{}{
			template.HTML("Validator no longer active"),
		})
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
		// Request is not a valid index number
		if err != nil {
			logger.Errorf("error parsing validator index: %v", err)
			http.Error(w, "Validator not found", http.StatusNotFound)
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
	COALESCE(orphaned_sync, 0) AS orphaned_sync,
	COALESCE(cl_rewards_gwei, 0) AS cl_rewards_gwei
	FROM validator_stats WHERE validatorindex = $1 ORDER BY day DESC`, index)

	if err != nil {
		logger.Errorf("error retrieving validator stats history: %v", err)
		http.Error(w, "Validator not found", http.StatusNotFound)
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
	ascOrdering := q.Get("order[0][dir]") == "asc"

	// retrieve all sync periods for this validator
	// ordering is descending for now
	var syncPeriods []struct {
		Period     uint64 `db:"period"`
		StartEpoch uint64 `db:"startepoch"`
		EndEpoch   uint64 `db:"endepoch"`
	}
	tempSyncPeriods := syncPeriods

	err = db.ReaderDb.Select(&tempSyncPeriods, `
		SELECT period as period, (period*$1) as endepoch, ((period+1)*$1)-1 as startepoch
		FROM sync_committees 
		WHERE validatorindex = $2
		ORDER BY period desc`, utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod, validatorIndex)
	if err != nil {
		logger.WithError(err).Errorf("error getting sync tab count data of sync-assignments")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	latestEpoch := services.LatestEpoch()

	//remove scheduled committees
	for i, syncPeriod := range tempSyncPeriods {
		if syncPeriod.EndEpoch <= latestEpoch {
			syncPeriods = tempSyncPeriods[i:]
			break
		}
	}

	// set latest epoch of this validators latest sync period to current epoch if latest sync epoch has yet to happen
	var diffToLatestEpoch uint64 = 0
	if latestEpoch < syncPeriods[0].StartEpoch {
		diffToLatestEpoch = syncPeriods[0].StartEpoch - latestEpoch
		syncPeriods[0].StartEpoch = latestEpoch
	}

	// total count of sync duties for this validator
	totalCount := (uint64(len(syncPeriods))*utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod - diffToLatestEpoch) * utils.Config.Chain.Config.SlotsPerEpoch

	tableData := [][]interface{}{}

	if totalCount > 0 && start <= totalCount {
		// if ordering is ascending, reverse sync period slice & swap start and end epoch of each period
		if ascOrdering {
			utils.ReverseSlice(syncPeriods)
			for i := range syncPeriods {
				syncPeriods[i].StartEpoch, syncPeriods[i].EndEpoch = syncPeriods[i].EndEpoch, syncPeriods[i].StartEpoch
			}
		}
		// syncPeriods[0].startEpoch will always be the epoch shown on page 1, regardless of the ordering
		// meaning that for descending ordering, syncPeriods[0].startEpoch will be the chronologically latest epoch of this validators lastest sync period
		// and for ascending ordering, syncPeriods[0].startEpoch will be the chronologically earliest epoch of this validators first sync period

		// set functions for moving away from start and back to start (with start being page 1)
		// depending on the ordering, this means either going up or down in epoch number
		var moveAway func(uint64, uint64) uint64
		var moveBack func(uint64, uint64) uint64
		var IsFurtherAway func(uint64, uint64) bool
		if ascOrdering {
			moveAway = func(a uint64, b uint64) uint64 {
				return a + b
			}
			moveBack = func(a uint64, b uint64) uint64 {
				return a - b
			}
			IsFurtherAway = func(a uint64, b uint64) bool {
				return a > b
			}
		} else {
			moveAway = func(a uint64, b uint64) uint64 {
				return a - b
			}
			moveBack = func(a uint64, b uint64) uint64 {
				return a + b
			}
			IsFurtherAway = func(a uint64, b uint64) bool {
				return a < b
			}
		}

		// amount of epochs moved away from start epoch
		epochOffset := (start / utils.Config.Chain.Config.SlotsPerEpoch)
		// amount of distinct consecutive epochs shown on this page
		epochsDiff := ((start + length) / utils.Config.Chain.Config.SlotsPerEpoch) - epochOffset

		shownEpochIndex := 0
		// first epoch containing the duties shown on this page
		firstShownEpoch := moveAway(syncPeriods[shownEpochIndex].StartEpoch, epochOffset)

		// handle first shown epoch being in the next sync period
		for IsFurtherAway(firstShownEpoch, syncPeriods[shownEpochIndex].EndEpoch) {
			overshoot := firstShownEpoch - syncPeriods[shownEpochIndex].EndEpoch
			shownEpochIndex++
			firstShownEpoch = syncPeriods[shownEpochIndex].StartEpoch + moveBack(overshoot, 1)
		}

		// last epoch containing the duties shown on this page
		lastShownEpoch := moveAway(firstShownEpoch, epochsDiff)
		// amount of epochs fetched by bigtable
		limit := moveBack(firstShownEpoch-lastShownEpoch, 1)

		var nextPeriodLimit int64 = 0
		if IsFurtherAway(lastShownEpoch, syncPeriods[shownEpochIndex].EndEpoch) {
			if IsFurtherAway(lastShownEpoch, syncPeriods[len(syncPeriods)-1].EndEpoch) {
				// handle showing the last page, which may hold less than 'length' amount of rows
				length = utils.Config.Chain.Config.SlotsPerEpoch - (start % utils.Config.Chain.Config.SlotsPerEpoch)
			} else {
				// handle crossing sync periods on the same page (i.e. including the earliest and latest slot of two sync periods from this validator)
				overshoot := lastShownEpoch - syncPeriods[shownEpochIndex].EndEpoch
				lastShownEpoch = syncPeriods[shownEpochIndex+1].StartEpoch + moveBack(overshoot, 1)
				limit += overshoot
				nextPeriodLimit = int64(overshoot)
			}
		}

		// retrieve sync duties from bigtable
		// note that the limit may be negative for either call, which results in the function fetching epochs for the absolute limit value in ascending ordering
		syncDuties, err := db.BigtableClient.GetValidatorSyncDutiesHistoryOrdered(validatorIndex, firstShownEpoch-limit, firstShownEpoch, ascOrdering)
		if err != nil {
			logger.Errorf("error retrieving validator sync duty data from bigtable: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if nextPeriodLimit != 0 {
			nextPeriodSyncDuties, err := db.BigtableClient.GetValidatorSyncDutiesHistoryOrdered(validatorIndex, lastShownEpoch-uint64(nextPeriodLimit), lastShownEpoch, ascOrdering)
			if err != nil {
				logger.Errorf("error retrieving second validator sync duty data from bigtable: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			syncDuties = append(syncDuties, nextPeriodSyncDuties...)
		}

		// sanity check for right amount of slots in response
		if uint64(len(syncDuties))%utils.Config.Chain.Config.SlotsPerEpoch == 0 {
			// extract correct slots
			tableData = make([][]interface{}, length)
			for dataIndex, slotIndex := 0, start%utils.Config.Chain.Config.SlotsPerEpoch; slotIndex < protomath.MinU64((start%utils.Config.Chain.Config.SlotsPerEpoch)+length, uint64(len(syncDuties))); dataIndex, slotIndex = dataIndex+1, slotIndex+1 {
				epoch := utils.EpochOfSlot(syncDuties[slotIndex].Slot)

				slotTime := utils.SlotToTime(syncDuties[slotIndex].Slot)

				if syncDuties[slotIndex].Status == 0 && time.Since(slotTime) <= time.Minute {
					syncDuties[slotIndex].Status = 2 // scheduled
				}
				tableData[dataIndex] = []interface{}{
					fmt.Sprintf("%d", utils.SyncPeriodOfEpoch(epoch)),
					utils.FormatEpoch(epoch),
					utils.FormatBlockSlot(syncDuties[slotIndex].Slot),
					utils.FormatSyncParticipationStatus(syncDuties[slotIndex].Status),
				}
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
