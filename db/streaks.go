package db

import (
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

	lastFinalizedEpoch := 0
	err = DB.Get(&lastFinalizedEpoch, `select coalesce(max(epoch),0) from epochs where finalized = 't'`)
	if err != nil {
		return false, fmt.Errorf("Error getting latestEpoch: %w", err)
	}

	if lastFinalizedEpoch == 0 {
		return false, nil
	}

	lastStreaksEpoch := 0
	err = DB.Get(&lastStreaksEpoch, `select coalesce(max(start+length)-1,0) from validator_attestation_streaks`)
	if err != nil {
		return false, fmt.Errorf("Error getting lastStreaksEpoch: %w", err)
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

	lastStatsDay := 0
	err = DB.Get(&lastStatsDay, `select coalesce(max(day),0) from validator_stats`)
	if err != nil {
		return false, fmt.Errorf("Error getting lastStatsDay: %w", err)
	}

	logger.WithFields(logrus.Fields{"day": day, "lastStreaksEpoch": lastStreaksEpoch, "startEpoch": startEpoch, "endEpoch": endEpoch, "lastFinalizedEpoch": lastFinalizedEpoch, "lastStatsDay": lastStatsDay}).Infof("updating streaks")

	var streaks []struct {
		Validatorindex int
		Status         int
		Start          int
		Length         int
	}

	boundingsQry := ``
	if startEpoch == endEpoch {
		// if we are only looking at 1 epoch there is no way to limit the search-space
		boundingsQry = `boundings as (select validatorindex, $2+1 as epoch, status from attestation_assignments_p where week = $2/255/7 and epoch = $2),`
	} else {
		// use validator_stats table to limit search-space
		nomissesQry := "select validatorindex, $2+1 as epoch, 1 as status from validator_stats where day = $1/225 and (missed_attestations = 0 or missed_attestations is null) and validatorindex != 2147483647"
		if lastStatsDay < day {
			// if the validator_stats table has no entry for this day we find validators with only misses or no misses
			nomissesQry = "select validatorindex, $2+1 as epoch, status from attestation_assignments_p where week = $1/225/7 and epoch >= $1 and epoch <= $2 group by validatorindex, status having count(*) = $2-$1+1"
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
			-- consider validator-activation, extra-step for performance-reasons
			fixedstreaks as (
				select streaks.validatorindex, streaks.status, 
					coalesce(v.activationepoch, streaks.start) as start,
					coalesce(streaks.end - v.activationepoch, streaks.length) as length
				from streaks
				left join (select validatorindex, activationepoch from validators where activationepoch > $1) v
					on v.validatorindex = streaks.validatorindex and v.activationepoch > streaks.start
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
		select validatorindex, status, start, length
		from rankedstreaks where r = 1 or start+length = $2+1 order by validatorindex, start`, boundingsQry)

	// fmt.Println(strings.ReplaceAll(strings.ReplaceAll(qry, "$1", fmt.Sprintf("%d", startEpoch)), "$2", fmt.Sprintf("%d", endEpoch)))

	err = DB.Select(&streaks, qry, uint64(startEpoch), uint64(endEpoch))
	if err != nil {
		return false, fmt.Errorf("Error getting streaks: %w", err)
	}

	t1 := time.Now()

	tx, err := DB.Beginx()
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
		n := 4
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*n)
		for i, d := range streaks[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*n+1, i*n+2, i*n+3, i*n+4))
			valueArgs = append(valueArgs, d.Validatorindex)
			valueArgs = append(valueArgs, d.Status)
			valueArgs = append(valueArgs, d.Start)
			valueArgs = append(valueArgs, d.Length)
		}
		stmt := fmt.Sprintf(`insert into validator_attestation_streaks (validatorindex, status, start, length) values %s`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return false, fmt.Errorf("Error inserting streaks %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return false, fmt.Errorf("Error committing validator_attestation_streaks: %w", err)
	}

	t2 := time.Now()
	logger.WithFields(logrus.Fields{"day": day, "lastStreaksEpoch": lastStreaksEpoch, "startEpoch": startEpoch, "endEpoch": endEpoch, "lastFinalizedEpoch": lastFinalizedEpoch, "lastStatsDay": lastStatsDay, "calculate": t1.Sub(t0), "save": t2.Sub(t1), "all": t2.Sub(t0), "count": len(streaks)}).Info("updating streaks completed")

	return lastFinalizedEpoch == endEpoch, nil
}
