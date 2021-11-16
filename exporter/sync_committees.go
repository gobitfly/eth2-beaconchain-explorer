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
	err := db.DB.Select(&dbPeriods, `select period from sync_committees group by period`)
	if err != nil {
		return err
	}
	dbPeriodsMap := make(map[uint64]bool, len(dbPeriods))
	for _, p := range dbPeriods {
		dbPeriodsMap[p] = true
	}
	currEpoch := utils.TimeToEpoch(time.Now())
	lastPeriod := utils.SyncPeriodOfEpoch(uint64(currEpoch) + 1) // we can look into the future
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.AltairForkEpoch)
	for p := firstPeriod; p <= lastPeriod; p++ {
		_, exists := dbPeriodsMap[p]
		if !exists {
			t0 := time.Now()
			err = exportSyncCommitteeAtPeriod(rpcClient, p)
			if err != nil {
				return err
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
		stateID = utils.FirstEpochOfSyncPeriod(p-1) * utils.Config.Chain.SlotsPerEpoch
	}
	epoch := utils.FirstEpochOfSyncPeriod(p)
	if stateID/utils.Config.Chain.SlotsPerEpoch <= utils.Config.Chain.AltairForkEpoch {
		stateID = utils.Config.Chain.AltairForkEpoch * utils.Config.Chain.SlotsPerEpoch
		epoch = utils.Config.Chain.AltairForkEpoch
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

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	valueArgs := make([]interface{}, len(c.Validators)*2)
	valueStrings := make([]string, len(c.Validators))
	for i, idxStr := range c.Validators {
		idxU64, err := strconv.ParseUint(idxStr, 10, 64)
		if err != nil {
			return err
		}
		valueArgs[i*2+0] = p
		valueArgs[i*2+1] = idxU64
		valueStrings[i] = fmt.Sprintf("($%d,$%d)", i*2+1, i*2+2)
	}
	_, err = tx.Exec(
		fmt.Sprintf(`
			INSERT INTO sync_committees (period, validatorindex) 
			VALUES %s ON CONFLICT (period, validatorindex) NO NOTHING`,
			strings.Join(valueStrings, ",")),
		valueArgs...)
	if err != nil {
		return err
	}

	firstSlot := utils.FirstEpochOfSyncPeriod(p) * utils.Config.Chain.SlotsPerEpoch
	nArgs := 4
	valueArgs = make([]interface{}, int(utils.Config.Chain.EpochsPerSyncCommitteePeriod)*nArgs)
	valueStrings = make([]string, utils.Config.Chain.EpochsPerSyncCommitteePeriod)
	for _, idx := range validatorsU64 {
		for i := 0; i < int(utils.Config.Chain.EpochsPerSyncCommitteePeriod*utils.Config.Chain.SlotsPerEpoch); i++ {
			slot := firstSlot + uint64(i)
			valueArgs[i*nArgs+0] = slot
			valueArgs[i*nArgs+1] = idx
			valueArgs[i*nArgs+2] = 0 // status = scheduled
			valueArgs[i*nArgs+3] = utils.WeekOfSlot(slot)
			valueStrings[i] = fmt.Sprintf("($%d,$%d,$%d,$%d)", i*nArgs+1, i*nArgs+2, i*nArgs+3, i*nArgs+4)
		}
		_, err = tx.Exec(
			fmt.Sprintf(`
				INSERT INTO sync_assignments_p (slot, validatorindex, status, week)
				VALUES %s ON CONFLICT (slot, validatorindex, week) DO NOTHING`,
				strings.Join(valueStrings, ",")),
			valueArgs...)
	}

	return tx.Commit()
}
