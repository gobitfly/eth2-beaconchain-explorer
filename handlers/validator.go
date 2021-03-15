package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
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

	"github.com/gorilla/mux"
	"github.com/juliangruber/go-intersect"
)

var validatorTemplate = template.Must(template.New("validator").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/validator/validator.html",
	"templates/validator/heading.html",
	"templates/validator/tables.html",
	"templates/validator/modals.html",
	"templates/validator/overview.html",
	"templates/validator/charts.html",
	"templates/validator/countdown.html",

	"templates/components/flashMessage.html",
	"templates/components/rocket.html",
))
var validatorNotFoundTemplate = template.Must(template.New("validatornotfound").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validator/validatornotfound.html"))
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

	data := InitPageData(w, r, "validators", "/validators", "")
	data.HeaderAd = true
	validatorPageData.NetworkStats = services.LatestIndexPageData()
	validatorPageData.User = data.User

	validatorPageData.FlashMessage, err = utils.GetFlash(w, r, validatorEditFlash)
	if err != nil {
		logger.Errorf("error retrieving flashes for validator %v: %v", vars["index"], err)
		http.Error(w, "Internal server error", 503)
		return
	}

	// Request came with a hash
	if strings.Contains(vars["index"], "0x") || len(vars["index"]) == 96 {
		pubKey, err := hex.DecodeString(strings.Replace(vars["index"], "0x", "", -1))
		if err != nil {
			logger.Errorf("error parsing validator public key %v: %v", vars["index"], err)
			http.Error(w, "Internal server error", 503)
			return
		}
		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			var name string
			err = db.DB.Get(&name, `SELECT name FROM validator_names WHERE publickey = $1`, pubKey)
			if err != nil {
				logger.Errorf("error getting validator-name from db: %v", err)
			} else {
				validatorPageData.Name = name
			}
			deposits, err := db.GetValidatorDeposits(pubKey)
			if err != nil {
				logger.Errorf("error getting validator-deposits from db: %v", err)
			}
			validatorPageData.DepositsCount = uint64(len(deposits.Eth1Deposits))
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
			validatorPageData.PublicKey = pubKey
			if deposits != nil && len(deposits.Eth1Deposits) > 0 {
				deposits.LastEth1DepositTs = deposits.Eth1Deposits[len(deposits.Eth1Deposits)-1].BlockTs
			}
			validatorPageData.Deposits = deposits

			churnRate, err := db.GetValidatorChurnLimit(services.LatestEpoch())
			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", 503)
				return
			}

			pendingCount, err := db.GetPendingValidatorCount()
			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", 503)
				return
			}
			validatorPageData.PendingCount = pendingCount

			validatorPageData.InclusionDelay = int64((utils.Config.Chain.Phase0.Eth1FollowDistance*utils.Config.Chain.Phase0.SecondsPerETH1Block+utils.Config.Chain.Phase0.SecondsPerSlot*utils.Config.Chain.Phase0.SlotsPerEpoch*utils.Config.Chain.Phase0.EpochsPerEth1VotingPeriod)/3600) + 1

			latestDeposit := time.Now().Unix()
			if len(deposits.Eth1Deposits) > 1 {
				latestDeposit = deposits.Eth1Deposits[len(deposits.Eth1Deposits)-1].BlockTs
			} else if time.Unix(latestDeposit, 0).Before(utils.SlotToTime(0)) {
				latestDeposit = utils.SlotToTime(0).Unix()
				validatorPageData.InclusionDelay = 0
			}

			if churnRate == 0 {
				churnRate = 4
				logger.Warning("Churn rate not set in config using 4 as default please set minPerEpochChurnLimit")
			}

			activationEstimate := (pendingCount/churnRate)*(utils.Config.Chain.Phase0.SecondsPerSlot*utils.Config.Chain.Phase0.SlotsPerEpoch) + uint64(latestDeposit)
			validatorPageData.EstimatedActivationTs = int64(activationEstimate)

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
			}
			watchlist, err := db.GetTaggedValidators(filter)
			if err != nil {
				logger.Errorf("error getting tagged validators from db: %v", err)
				http.Error(w, "Internal server error", 503)
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
				http.Error(w, "Internal server error", 503)
				return
			}

			return
		}
	} else {
		// Request came with a validator index number
		index, err = strconv.ParseUint(vars["index"], 10, 64)
		if err != nil {
			logger.Errorf("error parsing validator index: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
	}

	// GetAvgOptimalInclusionDistance(index)

	data.Meta.Title = fmt.Sprintf("%v - Validator %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, index, time.Now().Year())
	data.Meta.Path = fmt.Sprintf("/validator/%v", index)

	//logger.Infof("retrieving data, elapsed: %v", time.Since(start))
	//start = time.Now()

	err = db.DB.Get(&validatorPageData, `
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
			COALESCE(validators.balance, 0) AS balance,
			COALESCE(validator_performance.rank7d, 0) AS rank7d,
		    validators.status,
		    COALESCE(validators.balanceactivation, 0) AS balanceactivation,
		    COALESCE(validators.balance7d, 0) AS balance7d,
		    COALESCE(validators.balance31d, 0) AS balance31d
		FROM validators 
		LEFT JOIN validator_names 
			ON validators.pubkey = validator_names.publickey
		LEFT JOIN validator_performance 
			ON validators.validatorindex = validator_performance.validatorindex
		WHERE validators.validatorindex = $1`, index)
	if err != nil {
		//logger.Errorf("error retrieving validator page data: %v", err)

		err := validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	if validatorPageData.Rank7d > 0 {
		validatorPageData.RankPercentage = float64(validatorPageData.Rank7d) / float64(int64(validatorPageData.NetworkStats.ActiveValidators))
	}

	//logger.Infof("validator page data retrieved, elapsed: %v", time.Since(start))
	//start = time.Now()

	validatorPageData.Epoch = services.LatestEpoch()
	validatorPageData.Index = index
	if err != nil {
		logger.Errorf("error retrieving validator public key %v: %v", index, err)

		err := validatorNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	filter := db.WatchlistFilter{
		UserId:         data.User.UserID,
		Validators:     &pq.ByteaArray{validatorPageData.PublicKey},
		Tag:            types.ValidatorTagsWatchlist,
		JoinValidators: false,
	}

	watchlist, err := db.GetTaggedValidators(filter)
	if err != nil {
		logger.Errorf("error getting tagged validators from db: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	validatorPageData.Watchlist = watchlist

	//logger.Infof("watchlist data retrieved, elapsed: %v", time.Since(start))
	//start = time.Now()

	deposits, err := db.GetValidatorDeposits(validatorPageData.PublicKey)
	if err != nil {
		logger.Errorf("error getting validator-deposits from db: %v", err)
		http.Error(w, "Internal server error", 503)
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

	//logger.Infof("deposit data retrieved, elapsed: %v", time.Since(start))
	//start = time.Now()

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

	validatorPageData.ProposedBlocksCount = uint64(len(proposals))

	//logger.Infof("proposals data retrieved, elapsed: %v", time.Since(start))
	//start = time.Now()

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

	var incomeHistory []*types.ValidatorIncomeHistory
	err = db.DB.Select(&incomeHistory, "select day, start_balance, end_balance, COALESCE(deposits_amount, 0) as deposits_amount from validator_stats where validatorindex = $1 order by day;", index)
	if err != nil {
		logger.Errorf("error retrieving validator balance history: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	validatorPageData.IncomeHistoryChartData = make([]*types.ChartDataPoint, len(incomeHistory))

	if len(incomeHistory) > 0 {
		for i := 0; i < len(incomeHistory)-1; i++ {
			income := incomeHistory[i+1].StartBalance - incomeHistory[i].StartBalance
			if income >= incomeHistory[i].Deposits {
				income = income - incomeHistory[i].Deposits
			}
			color := "#7cb5ec"
			if income < 0 {
				color = "#f7a35c"
			}
			balanceTs := utils.DayToTime(incomeHistory[i+1].Day)
			validatorPageData.IncomeHistoryChartData[i] = &types.ChartDataPoint{X: float64(balanceTs.Unix() * 1000), Y: utils.ExchangeRateForCurrency(currency) * (float64(income) / 1000000000), Color: color}
		}

		lastDayBalance := incomeHistory[len(incomeHistory)-1].EndBalance
		lastDayIncome := int64(validatorPageData.CurrentBalance) - lastDayBalance
		lastDayIncomeColor := "#7cb5ec"
		if lastDayIncome < 0 {
			lastDayIncomeColor = "#f7a35c"
		}
		currentDay := validatorPageData.Epoch / ((24 * 60 * 60) / utils.Config.Chain.SlotsPerEpoch / utils.Config.Chain.SecondsPerSlot)

		validatorPageData.IncomeHistoryChartData[len(validatorPageData.IncomeHistoryChartData)-1] = &types.ChartDataPoint{X: float64(utils.DayToTime(currentDay).Unix() * 1000), Y: utils.ExchangeRateForCurrency(currency) * (float64(lastDayIncome) / 1000000000), Color: lastDayIncomeColor}
	}

	//logger.Infof("balance history retrieved, elapsed: %v", time.Since(start))
	//start = time.Now()

	earnings, err := GetValidatorEarnings([]uint64{index})
	if err != nil {
		logger.Errorf("error retrieving validator earnings: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	validatorPageData.Income1d = earnings.LastDay
	validatorPageData.Income7d = earnings.LastWeek
	validatorPageData.Income31d = earnings.LastMonth
	validatorPageData.Apr = earnings.APR

	//logger.Infof("income data retrieved, elapsed: %v", time.Since(start))
	//start = time.Now()

	if validatorPageData.Slashed {
		var slashingInfo struct {
			Slot    uint64
			Slasher uint64
			Reason  string
		}
		err = db.DB.Get(&slashingInfo,
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
			http.Error(w, "Internal server error", 503)
			return
		}
		validatorPageData.SlashedBy = slashingInfo.Slasher
		validatorPageData.SlashedAt = slashingInfo.Slot
		validatorPageData.SlashedFor = slashingInfo.Reason
	}

	err = db.DB.Get(&validatorPageData.SlashingsCount, `
		select COALESCE(sum(attesterslashingscount) + sum(proposerslashingscount), 0) from blocks where blocks.proposer = $1 and blocks.status = '1'
		`, index)
	if err != nil {
		logger.Errorf("error retrieving slashings-count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	//logger.Infof("slashing data retrieved, elapsed: %v", time.Since(start))
	//start = time.Now()

	err = db.DB.Get(&validatorPageData.AverageAttestationInclusionDistance, `
	SELECT COALESCE(
		AVG(1 + inclusionslot - COALESCE((
			SELECT MIN(slot)
			FROM blocks
			WHERE slot > aa.attesterslot AND blocks.status = '1'
		), 0)
	), 0)
	FROM attestation_assignments_p aa
	INNER JOIN blocks ON blocks.slot = aa.inclusionslot AND blocks.status <> '3'
	WHERE aa.week >= $1 / 1575 AND aa.epoch > $1 AND aa.validatorindex = $2 AND aa.inclusionslot > 0
	`, int64(validatorPageData.Epoch)-100, index)
	if err != nil {
		logger.Errorf("error retrieving AverageAttestationInclusionDistance: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if validatorPageData.AverageAttestationInclusionDistance > 0 {
		validatorPageData.AttestationInclusionEffectiveness = 1.0 / validatorPageData.AverageAttestationInclusionDistance * 100
	}

	var attestationStreaks []struct {
		Length uint64
	}
	err = db.DB.Select(&attestationStreaks, `select length from validator_attestation_streaks where validatorindex = $1 and status = 1 order by start desc`, index)
	if err != nil {
		logger.Errorf("error retrieving AttestationStreaks: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if len(attestationStreaks) > 1 {
		validatorPageData.CurrentAttestationStreak = attestationStreaks[0].Length
		validatorPageData.LongestAttestationStreak = attestationStreaks[1].Length
	} else if len(attestationStreaks) > 0 {
		validatorPageData.CurrentAttestationStreak = attestationStreaks[0].Length
		validatorPageData.LongestAttestationStreak = attestationStreaks[0].Length
	}

	//logger.Infof("effectiveness data retrieved, elapsed: %v", time.Since(start))
	//start = time.Now()

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

// ValidatorAttestationInclusionEffectiveness returns a validator's effectiveness in json
func ValidatorAttestationInclusionEffectiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		logger.Errorf("error parsing validator index: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var avgIncDistance float64

	err = db.DB.Get(&avgIncDistance, `
	SELECT COALESCE(
		AVG(1 + inclusionslot - COALESCE((
			SELECT MIN(slot)
			FROM blocks
			WHERE slot > aa.attesterslot AND blocks.status = '1'
		), 0)
	), 0)
	FROM attestation_assignments_p aa
	INNER JOIN blocks ON blocks.slot = aa.inclusionslot AND blocks.status <> '3'
	WHERE aa.week >= $1 / 1575 AND aa.epoch > $1 AND aa.validatorindex = $2 AND aa.inclusionslot > 0
	`, int64(services.LatestEpoch())-100, index)
	if err != nil {
		logger.Errorf("error retrieving AverageAttestationInclusionDistance: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var attestationInclusionEffectiveness float64

	if avgIncDistance > 0 {
		attestationInclusionEffectiveness = 1.0 / avgIncDistance * 100
	}

	type resp struct {
		Effectiveness float64 `json:"effectiveness"`
	}
	err = json.NewEncoder(w).Encode(resp{Effectiveness: attestationInclusionEffectiveness})
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
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
		ORDER BY `+orderBy+` `+orderDir+`
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
	start, err := strconv.ParseInt(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseInt(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "epoch",
		"2": "status",
		"4": "committeeindex",
		"6": "delay",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "epoch"
	}

	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	epoch := services.LatestEpoch()

	ae := struct {
		ActivationEpoch uint64
		ExitEpoch       uint64
	}{}

	err = db.DB.Get(&ae, "SELECT activationepoch, exitepoch FROM validators WHERE validatorindex = $1", index)
	if err != nil {
		logger.Errorf("error retrieving attestations count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	totalCount := epoch - ae.ActivationEpoch + 1
	if ae.ActivationEpoch > epoch {
		totalCount = 0
	}
	if ae.ExitEpoch != 9223372036854775807 {
		totalCount = ae.ExitEpoch - ae.ActivationEpoch
	}

	tableData := [][]interface{}{}

	if totalCount > 0 {
		var blocks []*types.ValidatorAttestation
		err = db.DB.Select(&blocks, `
			SELECT 
				aa.epoch, 
				aa.attesterslot, 
				aa.committeeindex, 
				CASE 
					WHEN blocks.status = '3' THEN '3'
					ELSE aa.status
				END AS status,
				CASE 
					WHEN blocks.status = '3' THEN 0
					ELSE aa.inclusionslot
				END AS inclusionslot,
				COALESCE(inclusionslot - (SELECT MIN(slot) FROM blocks WHERE slot > aa.attesterslot AND blocks.status = '1'), 0) as delay
			FROM attestation_assignments_p aa
			LEFT JOIN blocks on blocks.slot = aa.inclusionslot
			WHERE validatorindex = $1 AND aa.epoch > $2 AND aa.epoch <= $3
			ORDER BY `+orderBy+` `+orderDir, index, int64(epoch)-start-length, int64(epoch)-start)

		if err != nil {
			logger.Errorf("error retrieving validator attestations data: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}

		tableData = make([][]interface{}, len(blocks))

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
				utils.FormatAttestationInclusionSlot(b.InclusionSlot),
				utils.FormatInclusionDelay(b.InclusionSlot, b.Delay),
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
		http.Error(w, "Internal server error", 503)
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
		http.Redirect(w, r, "/validator/"+pubkey, 301)
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
		http.Redirect(w, r, "/validator/"+pubkey, 301)
		return
	}

	msgForHashing := "\x19Ethereum Signed Message:\n" + strconv.Itoa(len(signatureWrapper.Msg)) + signatureWrapper.Msg
	msgHash := crypto.Keccak256Hash([]byte(msgForHashing))

	signatureParsed, err := hex.DecodeString(strings.Replace(signatureWrapper.Sig, "0x", "", -1))
	if err != nil {
		logger.Errorf("error parsing submitted signature %v: %v", signatureWrapper.Sig, err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, 301)
		return
	}

	if len(signatureParsed) != 65 {
		logger.Errorf("signature must be 65 bytes long")
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, 301)
		return
	}

	if signatureParsed[64] == 27 || signatureParsed[64] == 28 {
		signatureParsed[64] -= 27
	}

	recoveredPubkey, err := crypto.SigToPub(msgHash.Bytes(), signatureParsed)
	if err != nil {
		logger.Errorf("error recovering pubkey: %v", err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, 301)
		return
	}
	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubkey)

	var depositedAddress string
	deposits, err := db.GetValidatorDeposits(pubkeyDecoded)
	if err != nil {
		logger.Errorf("error getting validator-deposits from db for signature verification: %v", err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, 301)
	}
	for _, deposit := range deposits.Eth1Deposits {
		if deposit.ValidSignature {
			depositedAddress = "0x" + fmt.Sprintf("%x", deposit.FromAddress)
			break
		}
	}

	if strings.ToLower(depositedAddress) == strings.ToLower(recoveredAddress.Hex()) {
		if applyNameToAll == "on" {
			res, err := db.DB.Exec(`
				INSERT INTO validator_names (publickey, name)
				SELECT publickey, $1 as name
				FROM (SELECT DISTINCT publickey FROM eth1_deposits WHERE from_address = $2 AND valid_signature) a
				ON CONFLICT (publickey) DO UPDATE SET name = excluded.name`, name, recoveredAddress.Bytes())
			if err != nil {
				logger.Errorf("error saving validator name (apply to all): %x: %v: %v", pubkeyDecoded, name, err)
				utils.SetFlash(w, r, validatorEditFlash, "Error: Db error while updating validator names")
				http.Redirect(w, r, "/validator/"+pubkey, 301)
				return
			}

			rowsAffected, _ := res.RowsAffected()
			utils.SetFlash(w, r, validatorEditFlash, fmt.Sprintf("Your custom name has been saved for %v validator(s).", rowsAffected))
			http.Redirect(w, r, "/validator/"+pubkey, 301)
		} else {
			_, err := db.DB.Exec(`
				INSERT INTO validator_names (publickey, name) 
				VALUES($2, $1) 
				ON CONFLICT (publickey) DO UPDATE SET name = excluded.name`, name, pubkeyDecoded)
			if err != nil {
				logger.Errorf("error saving validator name: %x: %v: %v", pubkeyDecoded, name, err)
				utils.SetFlash(w, r, validatorEditFlash, "Error: Db error while updating validator name")
				http.Redirect(w, r, "/validator/"+pubkey, 301)
				return
			}

			utils.SetFlash(w, r, validatorEditFlash, "Your custom name has been saved.")
			http.Redirect(w, r, "/validator/"+pubkey, 301)
		}

	} else {
		utils.SetFlash(w, r, validatorEditFlash, "Error: the provided signature is invalid")
		http.Redirect(w, r, "/validator/"+pubkey, 301)
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

	//length := 10

	var activationAndExitEpoch = struct {
		ActivationEpoch uint64 `db:"activationepoch"`
		ExitEpoch       uint64 `db:"exitepoch"`
	}{}
	err = db.DB.Get(&activationAndExitEpoch, "SELECT activationepoch, exitepoch FROM validators WHERE validatorindex = $1", index)
	if err != nil {
		logger.Errorf("error retrieving activationAndExitEpoch for validator-history: %v", err)
		http.Error(w, "Internal server error", 503)
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

	currentEpoch := services.LatestEpoch()

	var validatorHistory []*types.ValidatorHistory
	err = db.DB.Select(&validatorHistory, `
			SELECT 
				vbalance.epoch, 
				COALESCE(vbalance.balance - LAG(vbalance.balance) OVER (ORDER BY vbalance.epoch), 0) AS balancechange,
				COALESCE(assign.attesterslot, -1) AS attestatation_attesterslot,
				assign.inclusionslot AS attestation_inclusionslot,
				vblocks.status as proposal_status,
				vblocks.slot as proposal_slot
			FROM validator_balances_p vbalance
			LEFT JOIN attestation_assignments_p assign ON vbalance.validatorindex = assign.validatorindex AND vbalance.epoch = assign.epoch AND vbalance.week = assign.week
			LEFT JOIN blocks vblocks ON vbalance.validatorindex = vblocks.proposer AND vbalance.epoch = vblocks.epoch AND vbalance.week = vblocks.epoch / 1575
			WHERE vbalance.validatorindex = $1 AND vbalance.epoch >= $2 AND vbalance.epoch <= $3 AND vbalance.week >= $2 / 1575 AND vbalance.week <= $3 / 1575
			ORDER BY epoch DESC
			LIMIT 10
			`, index, currentEpoch-10-start, currentEpoch-start)

	if err != nil {
		logger.Errorf("error retrieving validator history: %v", err)
		http.Error(w, "Internal server error", 503)
		return
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
		http.Error(w, "Internal server error", 503)
		return
	}
}

var validatorStatsTableTemplate = template.Must(template.New("validator_stats").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validator_stats_table.html"))

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
			http.Error(w, "Internal server error", 503)
			return
		}
		index, err = db.GetValidatorIndex(pubKey)
		if err != nil {
			logger.Errorf("error parsing validator pubkey: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
	} else {
		// Request came with a validator index number
		index, err = strconv.ParseUint(vars["index"], 10, 64)
		if err != nil {
			logger.Errorf("error parsing validator index: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
	}

	data.Meta.Title = fmt.Sprintf("%v - Daily Validator Statistics %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, index, time.Now().Year())
	data.Meta.Path = fmt.Sprintf("/validator/%v/stats", index)

	validatorStatsTablePageData := &types.ValidatorStatsTablePageData{
		ValidatorIndex: index,
		Rows:           make([]*types.ValidatorStatsTableRow, 0),
	}

	err = db.DB.Select(&validatorStatsTablePageData.Rows, "SELECT * FROM validator_stats WHERE validatorindex = $1 ORDER BY day DESC", index)

	if err != nil {
		logger.Errorf("error retrieving validator stats history: %v", err)
		http.Error(w, "Validator not found", 404)
		return
	}

	for i := len(validatorStatsTablePageData.Rows) - 1; i > 0; i-- {
		validatorStatsTablePageData.Rows[i].EndBalance = validatorStatsTablePageData.Rows[i-1].StartBalance
		if validatorStatsTablePageData.Rows[i].EndBalance.Valid && validatorStatsTablePageData.Rows[i].StartBalance.Valid {

			validatorStatsTablePageData.Rows[i].Income = validatorStatsTablePageData.Rows[i].EndBalance.Int64 - validatorStatsTablePageData.Rows[i].StartBalance.Int64

			if validatorStatsTablePageData.Rows[i].DepositsAmount.Valid && validatorStatsTablePageData.Rows[i].Income >= validatorStatsTablePageData.Rows[i].DepositsAmount.Int64 {
				validatorStatsTablePageData.Rows[i].Income -= validatorStatsTablePageData.Rows[i].DepositsAmount.Int64
			}
		}
	}

	if len(validatorStatsTablePageData.Rows) > 0 {
		if validatorStatsTablePageData.Rows[0].EndBalance.Valid && validatorStatsTablePageData.Rows[0].StartBalance.Valid {
			validatorStatsTablePageData.Rows[0].Income = validatorStatsTablePageData.Rows[0].EndBalance.Int64 - validatorStatsTablePageData.Rows[0].StartBalance.Int64
		}

		if validatorStatsTablePageData.Rows[0].DepositsAmount.Valid {
			validatorStatsTablePageData.Rows[0].Income -= validatorStatsTablePageData.Rows[0].DepositsAmount.Int64
		}
	}

	data.Data = validatorStatsTablePageData
	err = validatorStatsTableTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

}
