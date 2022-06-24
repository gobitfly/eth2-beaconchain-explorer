package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func syncCommitteesExporter(rpcClient rpc.Client) {
	for {
		t0 := time.Now()
		err := exportSyncCommittees(rpcClient)
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err, "duration": time.Since(t0)}).Errorf("error exporting sync_committees")
		}
		time.Sleep(time.Second * 12)
	}
}

func exportSyncCommittees(rpcClient rpc.Client) error {
	var dbPeriods []uint64
	err := db.WriterDb.Select(&dbPeriods, `select period from sync_committees group by period`)
	if err != nil {
		return err
	}
	dbPeriodsMap := make(map[uint64]bool, len(dbPeriods))
	for _, p := range dbPeriods {
		dbPeriodsMap[p] = true
	}
	currEpoch := utils.TimeToEpoch(time.Now())
	lastPeriod := utils.SyncPeriodOfEpoch(uint64(currEpoch)) + 1 // we can look into the future
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.Config.AltairForkEpoch)
	for p := firstPeriod; p <= lastPeriod; p++ {
		_, exists := dbPeriodsMap[p]
		if !exists {
			t0 := time.Now()
			err = exportSyncCommitteeAtPeriod(rpcClient, p)
			if err != nil {
				return fmt.Errorf("error exporting snyc-committee at period %v: %w", p, err)
			}
			logrus.WithFields(logrus.Fields{
				"period":   p,
				"epoch":    utils.FirstEpochOfSyncPeriod(p),
				"duration": time.Since(t0),
			}).Infof("exported sync_committee")
		}
	}
	return nil
}

func exportSyncCommitteeAtPeriod(rpcClient rpc.Client, p uint64) error {
	stateID := uint64(0)
	if p > 0 {
		stateID = utils.FirstEpochOfSyncPeriod(p-1) * utils.Config.Chain.Config.SlotsPerEpoch
	}
	epoch := utils.FirstEpochOfSyncPeriod(p)
	if stateID/utils.Config.Chain.Config.SlotsPerEpoch <= utils.Config.Chain.Config.AltairForkEpoch {
		stateID = utils.Config.Chain.Config.AltairForkEpoch * utils.Config.Chain.Config.SlotsPerEpoch
		epoch = utils.Config.Chain.Config.AltairForkEpoch
	}

	firstEpoch := utils.FirstEpochOfSyncPeriod(p)
	lastEpoch := firstEpoch + utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod
	firstWeek := firstEpoch / 1575
	lastWeek := lastEpoch / 1575
	for w := firstWeek; w <= lastWeek; w++ {
		var one int
		err := db.WriterDb.Get(&one, fmt.Sprintf("SELECT 1 FROM information_schema.tables WHERE table_name = 'sync_assignments_%v'", w))
		if err != nil {
			logger.Infof("creating partition sync_assignments_%v", w)
			_, err := db.WriterDb.Exec(fmt.Sprintf("CREATE TABLE sync_assignments_%v PARTITION OF sync_assignments_p FOR VALUES IN (%v);", w, w))
			if err != nil {
				logger.Fatalf("unable to create partition sync_assignments_%v: %v", w, err)
			}
		}
	}

	c, err := rpcClient.GetSyncCommittee(fmt.Sprintf("%d", stateID), epoch)
	if err != nil {
		return err
	}

	validatorsU64 := make([]uint64, len(c.Validators))
	for i, idxStr := range c.Validators {
		idxU64, err := strconv.ParseUint(idxStr, 10, 64)
		if err != nil {
			return err
		}
		validatorsU64[i] = idxU64
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nArgs := 3
	valueArgs := make([]interface{}, len(c.Validators)*nArgs)
	valueIds := make([]string, len(c.Validators))
	for i, idxU64 := range validatorsU64 {
		valueArgs[i*nArgs+0] = p
		valueArgs[i*nArgs+1] = idxU64
		valueArgs[i*nArgs+2] = i
		valueIds[i] = fmt.Sprintf("($%d,$%d,$%d)", i*nArgs+1, i*nArgs+2, i*nArgs+3)
	}
	_, err = tx.Exec(
		fmt.Sprintf(`
			INSERT INTO sync_committees (period, validatorindex, committeeindex) 
			VALUES %s ON CONFLICT (period, validatorindex, committeeindex) DO NOTHING`,
			strings.Join(valueIds, ",")),
		valueArgs...)
	if err != nil {
		return err
	}

	slotsPerSyncPeriod := utils.Config.Chain.Config.EpochsPerSyncCommitteePeriod * utils.Config.Chain.Config.SlotsPerEpoch
	firstSlot := utils.FirstEpochOfSyncPeriod(p) * utils.Config.Chain.Config.SlotsPerEpoch
	nArgs = 4
	valueArgs = make([]interface{}, int(slotsPerSyncPeriod)*nArgs)
	valueIds = make([]string, slotsPerSyncPeriod)
	for _, idxU64 := range validatorsU64 {
		for i := 0; i < int(slotsPerSyncPeriod); i++ {
			slot := firstSlot + uint64(i)
			valueArgs[i*nArgs+0] = slot
			valueArgs[i*nArgs+1] = idxU64
			valueArgs[i*nArgs+2] = 0 // status = scheduled
			valueArgs[i*nArgs+3] = utils.WeekOfSlot(slot)
			valueIds[i] = fmt.Sprintf("($%d,$%d,$%d,$%d)", i*nArgs+1, i*nArgs+2, i*nArgs+3, i*nArgs+4)
		}
		_, err = tx.Exec(
			fmt.Sprintf(`
				INSERT INTO sync_assignments_p (slot, validatorindex, status, week)
				VALUES %s ON CONFLICT (slot, validatorindex, week) DO NOTHING`,
				strings.Join(valueIds, ",")),
			valueArgs...)
	}

	return tx.Commit()
}
