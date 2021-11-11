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
	stmt := fmt.Sprintf(`insert into sync_committees (period, validator) values %s`, strings.Join(valueStrings, ","))
	_, err = db.DB.Exec(stmt, valueArgs...)
	if err != nil {
		return err
	}
	return nil
}
