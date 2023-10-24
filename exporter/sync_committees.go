package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
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
	err := db.WriterDb.Select(&dbPeriods, `SELECT period FROM sync_committees GROUP BY period`)
	if err != nil {
		return err
	}
	dbPeriodsMap := make(map[uint64]bool, len(dbPeriods))
	for _, p := range dbPeriods {
		dbPeriodsMap[p] = true
	}
	currEpoch := services.LatestFinalizedEpoch()
	if currEpoch > 0 { // guard against underflows
		currEpoch = currEpoch - 1
	}
	lastPeriod := utils.SyncPeriodOfEpoch(uint64(currEpoch)) + 1 // we can look into the future
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.ClConfig.AltairForkEpoch)
	for p := firstPeriod; p <= lastPeriod; p++ {
		_, exists := dbPeriodsMap[p]
		if !exists {
			t0 := time.Now()
			err = exportSyncCommitteeAtPeriod(rpcClient, p)
			if err != nil {
				return fmt.Errorf("error exporting sync-committee at period %v: %w", p, err)
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
		stateID = utils.FirstEpochOfSyncPeriod(p-1) * utils.Config.Chain.ClConfig.SlotsPerEpoch
	}
	epoch := utils.FirstEpochOfSyncPeriod(p)
	if stateID/utils.Config.Chain.ClConfig.SlotsPerEpoch <= utils.Config.Chain.ClConfig.AltairForkEpoch {
		stateID = utils.Config.Chain.ClConfig.AltairForkEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		epoch = utils.Config.Chain.ClConfig.AltairForkEpoch
	}

	firstEpoch := utils.FirstEpochOfSyncPeriod(p)
	lastEpoch := firstEpoch + utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod - 1

	logger.Infof("exporting sync committee assignments for period %v (epoch %v to %v)", p, firstEpoch, lastEpoch)

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

	dedupMap := make(map[uint64]bool)

	for _, validator := range validatorsU64 {
		dedupMap[validator] = true
	}

	validatorsU64 = make([]uint64, 0, len(dedupMap))
	for validator := range dedupMap {
		validatorsU64 = append(validatorsU64, validator)
	}

	// start := time.Now()
	//
	// firstSlot := firstEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
	// lastSlot := lastEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch + utils.Config.Chain.ClConfig.SlotsPerEpoch - 1
	// logger.Infof("exporting sync committee assignments for period %v (epoch %v to %v, slot %v to %v) to bigtable", p, firstEpoch, lastEpoch, firstSlot, lastSlot)

	// err = db.BigtableClient.SaveSyncCommitteesAssignments(firstSlot, lastSlot, validatorsU64)
	// if err != nil {
	// 	return fmt.Errorf("error saving sync committee assignments: %v", err)
	// }
	// logger.Infof("exported sync committee assignments for period %v to bigtable in %v", p, time.Since(start))
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nArgs := 3
	valueArgs := make([]interface{}, len(validatorsU64)*nArgs)
	valueIds := make([]string, len(validatorsU64))
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

	return tx.Commit()
}
