package db

import (
	"eth2-exporter/metrics"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// UpdateAttestationStreaks updates the table `validator_attestation_streaks` which holds the current, longest streaks for consecutive missed and executed attestations.
// It will use `validator_stats.missed_attestations` to optimize the query if possible.
// It will update streaks for the day after the last updated streak or up to the last finalized epoch.
// It will return `true` for `updatedToLastFinalizedEpoch` if all attesations up to the last finalized epoch have been considered and false otherwise.
func UpdateAttestationStreaks() (updatedToLastFinalizedEpoch bool, err error) {
	t0 := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_attestation_streaks").Observe(time.Since(t0).Seconds())
	}()

	lastFinalizedEpoch := 0
	err = WriterDb.Get(&lastFinalizedEpoch, `SELECT COALESCE(MAX(epoch), 0) FROM epochs where epoch <= (select finalizedepoch from network_liveness order by headepoch desc limit 1)`)
	if err != nil {
		return false, fmt.Errorf("error getting latestEpoch: %w", err)
	}

	if lastFinalizedEpoch == 0 {
		return false, nil
	}

	lastStreaksEpoch := 0
	err = WriterDb.Get(&lastStreaksEpoch, `select coalesce(max(start+length)-1,0) from validator_attestation_streaks`)
	if err != nil {
		return false, fmt.Errorf("error getting lastStreaksEpoch: %w", err)
	}

	if lastStreaksEpoch >= lastFinalizedEpoch {
		return true, nil
	}

	startEpoch := lastStreaksEpoch + 1
	if lastStreaksEpoch == 0 {
		startEpoch = 0
	}
	endEpoch := lastFinalizedEpoch

	day := int(startEpoch / 225)

	if int(endEpoch/225) > day {
		endEpoch = (day+1)*225 - 1
	}

	if startEpoch > endEpoch {
		return false, nil
	}

	var statsExist bool
	err = WriterDb.Get(&statsExist, `select coalesce((select status from validator_stats_status where day = $1),false)`, day)
	if err != nil {
		return false, fmt.Errorf("error getting statsExist: %w", err)
	}

	if !statsExist {
		// only do streaks once per day for now
		logger.WithFields(logrus.Fields{"day": day, "lastStreaksEpoch": lastStreaksEpoch, "startEpoch": startEpoch, "endEpoch": endEpoch, "lastFinalizedEpoch": lastFinalizedEpoch, "statsExist": statsExist}).Infof("skipping streaks")
		return true, nil
	}

	logger.WithFields(logrus.Fields{"day": day, "lastStreaksEpoch": lastStreaksEpoch, "startEpoch": startEpoch, "endEpoch": endEpoch, "lastFinalizedEpoch": lastFinalizedEpoch, "statsExist": statsExist}).Infof("updating streaks")

	var streaks []struct {
		Validatorindex int
		Status         int
		Start          int
		Length         int
		Longest        bool
		Current        bool
	}

	boundingsQry := ``
	if startEpoch == endEpoch {
		// if we are only looking at 1 epoch there is no way to limit the search-space
		boundingsQry = `boundings as (select validatorindex, $2+1 as epoch, status from attestation_assignments_p where week = $2/255/7 and epoch = $2),`
	} else {
		// use validator_stats table to limit search-space
		nomissesQry := `select validatorindex, $2+1 as epoch, 1 as status from validator_stats where day = $1/225 and (missed_attestations = 0 or missed_attestations is null) and validatorindex != 2147483647`
		if !statsExist {
			// if the validator_stats table has no entry for this day we find validators with only misses or no misses
			nomissesQry = `select validatorindex, $2+1 as epoch, status from attestation_assignments_p where week = $1/225/7 and epoch >= $1 and epoch <= $2 group by validatorindex, status having count(*) = $2-$1+1`
		}
		boundingsQry = fmt.Sprintf(`
			-- limit search-space
			nomisses as (%s),
			aa as (
				select validatorindex, epoch, status from attestation_assignments_p 
				where week = $1/225/7 and epoch >= $1 and epoch <= $2 and validatorindex not in (select validatorindex from nomisses)
			),
			-- find boundings
			boundings as (
				select aa1.validatorindex, aa1.epoch+1 as epoch, aa1.status
				from aa aa1 
				left join aa aa2 on aa1.validatorindex = aa2.validatorindex and aa1.epoch+1 = aa2.epoch
				where aa1.status != aa2.status or aa1.epoch = $2
				union (select * from nomisses)
			),`, nomissesQry)
	}

	qry := fmt.Sprintf(`
		with
			%s
			-- calculate streaklengths
			streaks as (
				select 
					b1.validatorindex, 
					b1.status,
					coalesce(lag(epoch) over (partition by b1.validatorindex order by epoch), coalesce(vas.start, $1)) as start,
					b1.epoch as end,
					b1.epoch - coalesce(lag(epoch) over (partition by b1.validatorindex order by epoch), coalesce(vas.start, $1)) as length
				from boundings b1
					left join validator_attestation_streaks vas on 
						vas.validatorindex = b1.validatorindex 
						and vas.status = b1.status
						and vas.start+vas.length = $1
			),
			-- consider validator-activation and validator-exit, extra-step for performance-reasons
			fixedstreaks as (
				select streaks.validatorindex, streaks.status, 
					case
						when v.activationepoch > streaks.start then v.activationepoch
						else streaks.start
					end as start,
					case
						when v.exitepoch < streaks.end and v.activationepoch > streaks.start then v.exitepoch - v.activationepoch
						when v.exitepoch < streaks.end then v.exitepoch - streaks.start
						when v.activationepoch > streaks.start then streaks.end - v.activationepoch
						else streaks.length
					end as length
				from streaks
					inner join (
						select validatorindex, activationepoch, exitepoch
						from validators
						where exitepoch > $1 and activationepoch <= $2
					) v on v.validatorindex = streaks.validatorindex
			),
			-- rank by validator and status (we only save current and longest streaks (missed and executed))
			rankedstreaks as (
				select *, rank() over (partition by validatorindex, status order by length desc, start desc) as r
				from (
					select validatorindex, status, start, length from fixedstreaks 
					union 
					select validatorindex, status, start, length from validator_attestation_streaks
				) a
			)
		select 
			validatorindex, status, start, length, 
			case when r = 1 then true else false end as longest, 
			case when start+length = $2+1 then true else false end as current
		from rankedstreaks where r = 1 or start+length = $2+1 
		order by validatorindex, start`, boundingsQry)

	// fmt.Println(strings.ReplaceAll(strings.ReplaceAll(qry, "$1", fmt.Sprintf("%d", startEpoch)), "$2", fmt.Sprintf("%d", endEpoch)))

	err = WriterDb.Select(&streaks, qry, uint64(startEpoch), uint64(endEpoch))
	if err != nil {
		return false, fmt.Errorf("error getting streaks: %w", err)
	}

	t1 := time.Now()

	tx, err := WriterDb.Beginx()
	if err != nil {
		return false, fmt.Errorf("error starting db transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("truncate validator_attestation_streaks")
	if err != nil {
		return false, fmt.Errorf("error truncating validator performance table: %w", err)
	}

	batchSize := 5000
	for b := 0; b < len(streaks); b += batchSize {
		start := b
		end := b + batchSize
		if len(streaks) < end {
			end = len(streaks)
		}
		n := 6
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*n)
		for i, d := range streaks[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*n+1, i*n+2, i*n+3, i*n+4, i*n+5, i*n+6))
			valueArgs = append(valueArgs, d.Validatorindex)
			valueArgs = append(valueArgs, d.Status)
			valueArgs = append(valueArgs, d.Start)
			valueArgs = append(valueArgs, d.Length)
			valueArgs = append(valueArgs, d.Longest)
			valueArgs = append(valueArgs, d.Current)
		}
		stmt := fmt.Sprintf(`insert into validator_attestation_streaks (validatorindex, status, start, length, longest, current) values %s`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return false, fmt.Errorf("error inserting streaks %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return false, fmt.Errorf("error committing validator_attestation_streaks: %w", err)
	}

	t2 := time.Now()
	logger.WithFields(logrus.Fields{"day": day, "lastStreaksEpoch": lastStreaksEpoch, "startEpoch": startEpoch, "endEpoch": endEpoch, "lastFinalizedEpoch": lastFinalizedEpoch, "statsExist": statsExist, "calculate": t1.Sub(t0), "save": t2.Sub(t1), "all": t2.Sub(t0), "count": len(streaks)}).Info("updating streaks completed")

	return lastFinalizedEpoch == endEpoch, nil
}
