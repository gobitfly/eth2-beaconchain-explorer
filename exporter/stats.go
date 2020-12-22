package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func statsAggregator() {
	for {
		t0 := time.Now()
		logger.Info("aggregating stats")
		err := aggregateStats()
		if err != nil {
			logger.Errorf("DEBUG error aggregating stats: %v", err)
		} else {
			logger.WithField("duration", time.Since(t0)).Info("DEBUG aggregating stats completed")
		}
		time.Sleep(time.Hour)
	}
}

func aggregateStats() error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transaction: %w", err)
	}
	defer tx.Rollback()

	var currentEpoch uint64
	err = tx.Get(&currentEpoch, "SELECT MAX(epoch) FROM epochs")
	if err != nil {
		return fmt.Errorf("error retrieving latest epoch from epochs table: %w", err)
	}

	var lastAggregatedEpoch uint64
	err = tx.Get(&lastAggregatedEpoch, "SELECT COALESCE(MAX(start_epoch),0) FROM aggregated_validator_stats")
	if err != nil {
		return err
	}

	lastEpochOfFirstDay := uint64(utils.TimeToEpoch(utils.EpochToTime(0).Add(time.Hour * 24).Truncate(time.Hour * 24)))
	epochsPerDay := 3600 * 24 / (utils.Config.Chain.SecondsPerSlot * utils.Config.Chain.SlotsPerEpoch)
	lastStartEpoch := uint64(0)
	if currentEpoch > lastEpochOfFirstDay {
		lastStartEpoch = epochsPerDay * (currentEpoch - lastEpochOfFirstDay + 1) / epochsPerDay
	}

	fmt.Println("DEBUG lastStartEpoch", lastStartEpoch, utils.EpochToTime(lastStartEpoch))
	fmt.Println("DEBUG lastEpochOfFirstDay", lastEpochOfFirstDay, utils.EpochToTime(lastEpochOfFirstDay))
	fmt.Println("DEBUG lastAggregatedEpoch", lastAggregatedEpoch, utils.EpochToTime(lastAggregatedEpoch))

	startEpoch := lastAggregatedEpoch
	for startEpoch <= lastStartEpoch {
		endEpoch := startEpoch + epochsPerDay - 1
		if startEpoch == 0 {
			endEpoch = lastEpochOfFirstDay
		} else if startEpoch == lastStartEpoch {
			endEpoch = currentEpoch
		}
		t0 := time.Now()
		err = aggregateStatsEpochs(tx, startEpoch, endEpoch)
		if err != nil {
			return err
		}
		fmt.Printf("DEBUG aggregateValidatorStats %v %v %v\n", startEpoch, endEpoch, time.Since(t0))
		startEpoch = endEpoch + 1
	}

	return tx.Commit()
}

func aggregateStatsEpochs(tx *sqlx.Tx, startEpoch, endEpoch uint64) error {
	type validatorStats struct {
		StartEpoch uint64
		EndEpoch   uint64
		MinBalance uint64
		MaxBalance uint64
		AvgBalance uint64
	}
	validatorStatsMap := map[uint64]*validatorStats{}
	networkStats := struct {
		StartEpoch uint64
		EndEpoch   uint64
		MinBalance uint64
		MaxBalance uint64
		AvgBalance uint64

		MinInclusionDelay uint64
		MaxInclusionDelay uint64
		AvgInclusionDelay float64

		MinOptimalInclusionDistance float64
		MaxOptimalInclusionDistance float64
		AvgOptimalInclusionDistance float64

		MissedAttestations uint64
		MissedBlocks       uint64
		OrphanedBlocks     uint64
		AttesterSlashings  uint64
		ProposerSlashings  uint64
		VoluntaryExits     uint64
		Activations        uint64

		TotalIncome int64
		MinIncome   int64
		MaxIncome   int64
		AvgIncome   int64

		StartParticipationRate float64
		EndParticipationRate   float64
		MinParticipationRate   float64
		MaxParticipationRate   float64
		AvgParticipationRate   float64
	}{}
	networkStats.StartEpoch = startEpoch
	networkStats.EndEpoch = endEpoch

	t0 := time.Now()
	validators := []uint64{}
	err := db.DB.Select(&validators, "select validatorindex from validators where activationepoch <= $1 and exitepoch > $2", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error getting validators: %w", err)
	}
	for _, v := range validators {
		validatorStatsMap[v] = &validatorStats{}
	}

	t1 := time.Now()
	deposits := []struct {
		Validatorindex uint64
		Epoch          uint64
		Amount         uint64
	}{}
	err = db.DB.Select(&deposits, "select validators.validatorindex, (block_slot/32)::int as epoch, amount from blocks_deposits inner join validators on validators.pubkey = blocks_deposits.publickey where block_slot/32 >= $1 and block_slot/32 <= $2", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error getting deposits: %w", err)
	}
	// map[validator]map[epoch]depositamount
	depositsMap := map[uint64]map[uint64]uint64{}
	for _, d := range deposits {
		if _, exists := depositsMap[d.Validatorindex]; !exists {
			depositsMap[d.Validatorindex] = map[uint64]uint64{}
		}
		depositsMap[d.Validatorindex][d.Epoch] = d.Amount
	}

	t2 := time.Now()
	type aggregatedBalance struct {
		Validatorindex uint64
		Min            uint64
		Max            uint64
		Avg            uint64
	}
	aggregatedBalances := []aggregatedBalance{}
	err = db.DB.Select(&aggregatedBalances, `
		select vb.validatorindex, min(vb.balance), max(vb.balance), avg(vb.balance)::bigint as avg 
		from validators v 
		left join validator_balances vb on v.validatorindex = vb.validatorindex and (vb.epoch >= $1 or vb.epoch <= $2) 
		where v.activationepoch <= $1 and v.exitepoch > $2 
		group by vb.validatorindex`, startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error getting aggregatedBalance: %w", err)
	}
	aggregatedBalancesMap := map[uint64]aggregatedBalance{}
	networkStats.MinBalance = math.MaxUint64
	networkStats.MaxBalance = 0
	totalAvgBalance := uint64(0)
	for _, b := range aggregatedBalances {
		aggregatedBalancesMap[b.Validatorindex] = b
		if b.Min < networkStats.MinBalance {
			networkStats.MinBalance = b.Min
		}
		if b.Max > networkStats.MaxBalance {
			networkStats.MaxBalance = b.Max
		}
		totalAvgBalance += b.Avg
	}
	networkStats.AvgBalance = totalAvgBalance / uint64(len(aggregatedBalances))

	t3 := time.Now()
	startEndBalances := []struct {
		Validatorindex uint64
		Balance        uint64
		Epoch          uint64
	}{}
	err = db.DB.Select(&startEndBalances, "select validatorindex, balance, epoch from validator_balances where epoch = $1 or epoch = $2", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error getting startEndBalances: %w", err)
	}
	// map[validator]map[epoch]balance
	startEndBalancesMap := map[uint64]map[uint64]uint64{}
	for _, b := range startEndBalances {
		if _, exists := startEndBalancesMap[b.Validatorindex]; !exists {
			startEndBalancesMap[b.Validatorindex] = map[uint64]uint64{}
		}
		startEndBalancesMap[b.Validatorindex][b.Epoch] = b.Balance
	}

	t4 := time.Now()
	minIncome := int64(math.MaxInt64)
	maxIncome := int64(math.MinInt64)
	totalIncome := int64(0)
	// map[validator]income
	incomeMap := map[uint64]int64{}
	for _, i := range validators {
		if _, exists := startEndBalancesMap[i]; !exists {
			return fmt.Errorf("could not find any balance for validator %v (%v-%v)", i, startEpoch, endEpoch)
		}
		endBalance, exists := startEndBalancesMap[i][endEpoch]
		if !exists {
			return fmt.Errorf("could not find endBalance for validator %v (%v-%v)", i, startEpoch, endEpoch)
		}
		startBalance, exists := startEndBalancesMap[i][startEpoch]
		if !exists {
			startBalance = 0
		}
		depositsSum := uint64(0)
		if _, exists := depositsMap[i]; exists {
			for _, depositAmount := range depositsMap[i] {
				depositsSum += depositAmount
			}
		}
		income := int64(endBalance) - int64(startBalance) - int64(depositsSum)
		if income < minIncome {
			minIncome = income
		}
		if income > maxIncome {
			maxIncome = income
		}
		totalIncome += income
		incomeMap[i] = income
	}
	networkStats.AvgIncome = totalIncome / int64(len(incomeMap))
	networkStats.MinIncome = minIncome
	networkStats.MaxIncome = maxIncome
	networkStats.TotalIncome = totalIncome

	t5 := time.Now()
	attestations := []struct {
		Validatorindex uint64
		Attesterslot   uint64
		Inclusionslot  uint64
		Earliestslot   uint64
		Status         string
	}{}
	err = db.DB.Select(&attestations, `
		select validatorindex, attesterslot, inclusionslot, inclusionslot as earliestslot, status 
		from attestation_assignments 
		where attesterslot/32 >= $1 and attesterslot/32 <= $2`, startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error getting attestations: %w", err)
	}
	inclusionDelays := make([]uint64, len(attestations))
	totalInclusionDelay := uint64(0)
	totalOptimalInclusionDistance := uint64(0)
	optimalInclusionDistanceMap := map[uint64][]uint64{}
	networkStats.MinInclusionDelay = math.MaxUint64
	networkStats.MaxInclusionDelay = 0
	networkStats.MinOptimalInclusionDistance = math.MaxUint64
	networkStats.MaxOptimalInclusionDistance = 0
	for i, a := range attestations {
		inclusionDelays[i] = a.Inclusionslot - a.Attesterslot
		totalInclusionDelay += a.Inclusionslot - a.Attesterslot
		totalOptimalInclusionDistance += a.Inclusionslot - a.Earliestslot
		if _, exists := optimalInclusionDistanceMap[a.Validatorindex]; exists {
			optimalInclusionDistanceMap[a.Validatorindex] = []uint64{}
		} else {
			optimalInclusionDistanceMap[a.Validatorindex] = append(optimalInclusionDistanceMap[a.Validatorindex], a.Inclusionslot-a.Earliestslot)
		}
	}
	networkStats.AvgInclusionDelay = float64(totalInclusionDelay) / float64(len(attestations))
	networkStats.AvgOptimalInclusionDistance = float64(totalOptimalInclusionDistance) / float64(len(attestations))

	t6 := time.Now()
	blocks := []struct {
		Epoch                  uint64
		Slot                   uint64
		Proposer               uint64
		Attestationscount      uint64
		Depositscount          uint64
		Voluntaryexitscount    uint64
		Proposerslashingscount uint64
		Attesterslashingscount uint64
		Status                 string
	}{}
	err = db.DB.Select(&blocks, "select epoch, slot, proposer, attestationscount, depositscount, voluntaryexitscount, proposerslashingscount, attesterslashingscount, status from blocks where slot/32 >= $1 and slot/32 <= $2", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error getting blocks: %w", err)
	}
	for _, b := range blocks {
		networkStats.AttesterSlashings += b.Attesterslashingscount
		networkStats.ProposerSlashings += b.Proposerslashingscount
		if b.Status == "2" {
			networkStats.MissedBlocks += 1
		} else if b.Status == "3" {
			networkStats.OrphanedBlocks += 1
		}
	}

	t7 := time.Now()
	epochs := []struct {
		Epoch             uint64
		Participationrate float64
	}{}
	err = db.DB.Select(&epochs, "select epoch, globalparticipationrate as participationrate from epochs where epoch >= $1 and epoch <= $2", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error getting epochs: %w", err)
	}
	totalParticipationRate := float64(0)
	networkStats.MinParticipationRate = float64(1)
	networkStats.MaxParticipationRate = float64(0)
	for _, e := range epochs {
		totalParticipationRate += e.Participationrate
		if e.Participationrate < networkStats.MinParticipationRate {
			networkStats.MinParticipationRate = e.Participationrate
		}
		if e.Participationrate > networkStats.MaxParticipationRate {
			networkStats.MaxParticipationRate = e.Participationrate
		}
	}
	networkStats.StartParticipationRate = epochs[0].Participationrate
	networkStats.EndParticipationRate = epochs[len(epochs)-1].Participationrate
	networkStats.AvgParticipationRate = totalParticipationRate / float64(len(epochs))

	// insert stats per validator

	// insert stats for the global network

	t8 := time.Now()
	logger.WithFields(logrus.Fields{
		"t0": t1.Sub(t0),
		"t1": t2.Sub(t1),
		"t2": t3.Sub(t2),
		"t3": t4.Sub(t3),
		"t4": t5.Sub(t4),
		"t5": t6.Sub(t5),
		"t6": t7.Sub(t6),
		"t7": t8.Sub(t7),
	}).Infof("DEBUG aggregate start: %v, end: %v, stats: %+v\n", startEpoch, endEpoch, networkStats)
	return nil
}
