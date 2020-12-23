package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func statsAggregator() {
	for {
		var start time.Time
		var err error

		start = time.Now()
		logger.Info("aggregating stats")
		err = aggregateStats()
		if err != nil {
			logger.WithError(err).Error("error aggregating stats")
		} else {
			logger.WithField("duration", time.Since(start)).Info("aggregating stats completed")
		}

		start = time.Now()
		logger.Info("collecting historical data")
		err = collectHistorical()
		if err != nil {
			logger.WithError(err).Error("error collecting historical data")
		} else {
			logger.WithField("duration", time.Since(start)).Info("collecting historical data completed")
		}

		time.Sleep(time.Second * 10)
	}
}

func aggregateStats() error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transaction: %w", err)
	}
	defer tx.Rollback()

	epochsPerDay := 3600 * 24 / (utils.Config.Chain.SecondsPerSlot * utils.Config.Chain.SlotsPerEpoch)
	lastEpochOfFirstDay := uint64(utils.TimeToEpoch(utils.EpochToTime(0).Add(time.Hour * 24).Truncate(time.Hour * 24)))

	var currentEpoch uint64
	err = tx.Get(&currentEpoch, "SELECT CAST(COALESCE(MAX(epoch),0) AS BIGINT) FROM epochs")
	if err != nil {
		return fmt.Errorf("error retrieving latest epoch from epochs table: %w", err)
	}

	if currentEpoch < lastEpochOfFirstDay {
		return nil
	}

	var lastAggregatedEpoch uint64
	err = tx.Get(&lastAggregatedEpoch, "SELECT CAST(COALESCE(MAX(end_epoch),0) AS BIGINT) FROM validator_stats")
	if err != nil {
		return err
	}

	lastStartEpoch := 1 + lastEpochOfFirstDay + epochsPerDay*((currentEpoch-lastEpochOfFirstDay)/epochsPerDay)
	if lastAggregatedEpoch != 0 && lastStartEpoch+epochsPerDay <= lastAggregatedEpoch {
		return nil
	}

	startEpoch := lastAggregatedEpoch
	batchSize := 2
	for i := 0; startEpoch <= lastStartEpoch && i < batchSize; i++ {
		endEpoch := startEpoch + epochsPerDay - 1
		if startEpoch == 0 {
			endEpoch = lastEpochOfFirstDay
		}
		err = aggregateStatsEpochs(tx, startEpoch, endEpoch)
		if err != nil {
			return err
		}
		startEpoch = endEpoch + 1
	}

	return tx.Commit()
}

func aggregateStatsEpochs(tx *sqlx.Tx, startEpoch, endEpoch uint64) error {
	logger.WithFields(logrus.Fields{"startEpoch": startEpoch, "endEpoch": endEpoch}).Info("aggregating stats")

	validatorStatsMap := map[uint64]*types.ValidatorStats{}

	networkStats := types.NetworkStats{}
	networkStats.StartEpoch = startEpoch
	networkStats.EndEpoch = endEpoch

	var err error

	t0 := time.Now()
	validators := []uint64{}
	err = tx.Select(&validators, "select validatorindex from validators where withdrawableepoch > $1 and activationepoch <= $2", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error getting validators: %w", err)
	}
	for _, v := range validators {
		s := &types.ValidatorStats{}
		s.ValidatorIndex = v
		s.StartEpoch = startEpoch
		s.EndEpoch = endEpoch
		s.MinBalance = math.MaxUint64
		s.MaxBalance = 0
		s.MinEffectiveBalance = math.MaxUint64
		s.MaxEffectiveBalance = 0
		validatorStatsMap[v] = s
	}

	t1 := time.Now()
	deposits := []struct {
		Validatorindex uint64
		Epoch          uint64
		Amount         uint64
	}{}
	err = tx.Select(&deposits, "select validators.validatorindex, (block_slot/32)::int as epoch, amount from blocks_deposits inner join validators on validators.pubkey = blocks_deposits.publickey where block_slot/32 >= $1 and block_slot/32 <= $2", startEpoch, endEpoch-1)
	if err != nil {
		return fmt.Errorf("error getting deposits: %w", err)
	}
	for _, d := range deposits {
		if _, exists := validatorStatsMap[d.Validatorindex]; !exists {
			// return fmt.Errorf("error aggregating deposits: no entry in validatorStatsMap for %v", d.Validatorindex)
			continue
		}
		validatorStatsMap[d.Validatorindex].Deposits++
		validatorStatsMap[d.Validatorindex].DepositsAmount += d.Amount
	}

	t2 := time.Now()
	// check if data is in hot-table, otherwise we need to use the historical-table
	epochBoundaries := struct {
		Firstepoch uint64
		Lastepoch  uint64
	}{}
	err = tx.Get(&epochBoundaries, "select min(epoch) as firstepoch, max(epoch) as lastepoch from validator_balances")
	if err != nil {
		return err
	}
	validatorBalancesTable := "validator_balances"
	if epochBoundaries.Firstepoch > startEpoch || epochBoundaries.Lastepoch < endEpoch {
		validatorBalancesTable = "validator_balances_historical"
	}
	type balance struct {
		Validatorindex   uint64
		Epoch            uint64
		Balance          uint64
		Effectivebalance uint64
	}
	balances := []*balance{}
	err = tx.Select(&balances, "select validatorindex, epoch, balance, effectivebalance from "+validatorBalancesTable+" where epoch >= $1 and epoch <= $2 order by validatorindex, epoch", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error aggregating balance: %w", err)
	}
	networkStats.MinBalance = math.MaxUint64
	networkStats.MaxBalance = 0
	firstBalanceMap := map[uint64]*balance{}
	for _, b := range balances {
		v, exists := validatorStatsMap[b.Validatorindex]
		if !exists {
			// return fmt.Errorf("error aggregating balances: no entry for %v in validatorStatsMap", b.Validatorindex)
			continue
		}
		if _, exists := firstBalanceMap[b.Validatorindex]; !exists {
			firstBalanceMap[b.Validatorindex] = b
		}
		if b.Epoch == startEpoch {
			v.StartBalance = b.Balance
			v.StartEffectiveBalance = b.Effectivebalance
		} else if b.Epoch == endEpoch {
			v.EndBalance = b.Balance
			v.EndEffectiveBalance = b.Effectivebalance
		}
		if b.Balance < v.MinBalance {
			v.MinBalance = b.Balance
		}
		if b.Balance > v.MaxBalance {
			v.MaxBalance = b.Balance
		}
		if b.Effectivebalance < v.MinEffectiveBalance {
			v.MinEffectiveBalance = b.Effectivebalance
		}
		if b.Effectivebalance > v.MaxEffectiveBalance {
			v.MaxEffectiveBalance = b.Effectivebalance
		}
		if b.Balance < networkStats.MinBalance {
			networkStats.MinBalance = b.Balance
		}
		if b.Balance > networkStats.MaxBalance {
			networkStats.MaxBalance = b.Balance
		}
	}

	t3 := time.Now()
	for v, s := range validatorStatsMap {
		firstBalance, exists := firstBalanceMap[v]
		if !exists {
			return fmt.Errorf("error aggregating incomes: no entry for %v in firstBalanceMap", v)
		}
		income := int64(s.EndBalance) - int64(firstBalance.Balance) - int64(s.DepositsAmount)
		if income < networkStats.MinIncome {
			networkStats.MinIncome = income
		}
		if income > networkStats.MaxIncome {
			networkStats.MaxIncome = income
		}
		networkStats.TotalIncome += income
	}
	networkStats.AvgIncome = networkStats.TotalIncome / int64(len(validatorStatsMap))

	t4 := time.Now()
	/*
		err = tx.Get(&epochBoundaries, "select min(epoch) as firstepoch, max(epoch) as lastepoch from attestation_assignments")
		if err != nil {
			return err
		}
		attestationAssignmentsTable := "attestation_assignments"
		if epochBoundaries.Firstepoch > startEpoch || epochBoundaries.LastEpoch < endEpoch {
			attestationAssignmentsTable = "attestation_assignments_historical"
		}
		logger.Info("DEBUG aggregating attestations")
		attestations := []struct {
			Validatorindex        uint64
			Attesterslot          uint64
			Inclusionslot         uint64
			Earliestinclusionslot uint64
			Status                string
		}{}
		err = tx.Select(&attestations, fmt.Sprintf(`
			with earliestinclusionslots as (
				select
					b1.slot,
					(select min(b2.slot) from blocks b2 where b2.slot > b1.slot and status = '1') as earliestinclusionslot
				from blocks b1 where slot/32 >= $1 and slot/32 <= $2
			)
			select
				validatorindex,
				attesterslot,
				inclusionslot,
				eis.earliestinclusionslot,
				status
			from %v aa
			left join earliestinclusionslots eis on eis.slot = aa.attesterslot
			where attesterslot/32 >= $1 and attesterslot/32 <= $2`, attestationAssignmentsTable), startEpoch, endEpoch)
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
			if a.Earliestinclusionslot == 0 {
				// return fmt.Errorf("error getting earliestinclusionslot for validator %v at slot %v", a.Validatorindex, a.Attesterslot)
				continue
			}
			inclusionDelays[i] = a.Inclusionslot - a.Attesterslot
			totalInclusionDelay += a.Inclusionslot - a.Attesterslot
			optimalInclusionDistance := a.Inclusionslot - a.Earliestinclusionslot
			totalOptimalInclusionDistance += optimalInclusionDistance
			if _, exists := optimalInclusionDistanceMap[a.Validatorindex]; exists {
				optimalInclusionDistanceMap[a.Validatorindex] = []uint64{optimalInclusionDistance}
			} else {
				optimalInclusionDistanceMap[a.Validatorindex] = append(optimalInclusionDistanceMap[a.Validatorindex], optimalInclusionDistance)
			}
		}
		networkStats.AvgInclusionDelay = float64(totalInclusionDelay) / float64(len(attestations))
		networkStats.AvgOptimalInclusionDistance = float64(totalOptimalInclusionDistance) / float64(len(attestations))
	*/

	t5 := time.Now()
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
	err = tx.Select(&blocks, "select epoch, slot, proposer, attestationscount, depositscount, voluntaryexitscount, proposerslashingscount, attesterslashingscount, status from blocks where slot != 0 and slot/32 >= $1 and slot/32 <= $2", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error aggregating blocks: %w", err)
	}
	for _, b := range blocks {
		networkStats.AttesterSlashings += b.Attesterslashingscount
		networkStats.ProposerSlashings += b.Proposerslashingscount
		if b.Status == "2" {
			networkStats.MissedBlocks += 1
		} else if b.Status == "3" {
			networkStats.OrphanedBlocks += 1
		}
		v, exists := validatorStatsMap[b.Proposer]
		if !exists {
			// continue
			return fmt.Errorf("error aggregating blocks: no entry in validatorStatsMap for %v (slot: %v, epoch: %v)", b.Proposer, b.Slot, b.Epoch)
		}
		v.AttesterSlashings += b.Attesterslashingscount
		v.ProposerSlashings += b.Proposerslashingscount
		if b.Status == "2" {
			v.MissedBlocks += 1
		} else if b.Status == "3" {
			v.OrphanedBlocks += 1
		} else {
			v.ProposedBlocks += 1
		}
	}

	t6 := time.Now()
	epochs := []struct {
		Epoch             uint64
		Participationrate float64
	}{}
	err = tx.Select(&epochs, "select epoch, globalparticipationrate as participationrate from epochs where epoch >= $1 and epoch <= $2", startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error aggregating epochs: %w", err)
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

	t7 := time.Now()
	stmtInsertValidatorStats, err := tx.Prepare(`
		insert into validator_stats (
			validatorindex         ,
			start_epoch            ,
			end_epoch              ,
			start_balance          ,
			end_balance            ,
			min_balance            ,
			max_balance            ,
			start_effective_balance,
			end_effective_balance  ,
			min_effective_balance  ,
			max_effective_balance  ,
			missed_attestations    ,
			orphaned_attestations  ,
			proposed_blocks        ,
			missed_blocks          ,
			orphaned_blocks        ,
			attester_slashings     ,
			proposer_slashings     ,
			income                 ,
			deposits               ,
			deposits_amount
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`)

	for _, v := range validatorStatsMap {
		_, err = stmtInsertValidatorStats.Exec(
			v.ValidatorIndex,
			v.StartEpoch,
			v.EndEpoch,
			v.StartBalance,
			v.EndBalance,
			v.MinBalance,
			v.MaxBalance,
			v.StartEffectiveBalance,
			v.EndEffectiveBalance,
			v.MinEffectiveBalance,
			v.MaxEffectiveBalance,
			v.MissedAttestations,
			v.OrphanedAttestations,
			v.ProposedBlocks,
			v.MissedBlocks,
			v.OrphanedBlocks,
			v.AttesterSlashings,
			v.ProposerSlashings,
			v.Income,
			v.Deposits,
			v.DepositsAmount,
		)
		if err != nil {
			return err
		}
	}

	t8 := time.Now()
	logger.WithFields(logrus.Fields{
		"t0":           t1.Sub(t0),
		"t1":           t2.Sub(t1),
		"t2":           t3.Sub(t2),
		"t3":           t4.Sub(t3),
		"t4":           t5.Sub(t4),
		"t5":           t6.Sub(t5),
		"t6":           t7.Sub(t6),
		"t7":           t8.Sub(t7),
		"tTotal":       t8.Sub(t0),
		"startEpoch":   startEpoch,
		"endEpoch":     endEpoch,
		"networkStats": fmt.Sprintf("%+v", networkStats),
	}).Info("aggregated stats")
	return nil
}
