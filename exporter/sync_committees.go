package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/utils"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func syncCommitteesExporter(rpcClient rpc.Client) {
	for {
		t0 := time.Now()
		err := exportSyncCommittees(rpcClient)
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err, "duration": time.Since(t0)}).Errorf("error exporting sync_committees")
		} else {
			logrus.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("exported sync_committees")
		}
		time.Sleep(time.Second * 12)
	}
}

func exportSyncCommittees(rpcClient rpc.Client) error {
	var dbPeriods []uint64
	err := db.DB.Select(&dbPeriods, `select period from sync_committees`)
	if err != nil {
		return err
	}
	dbPeriodsMap := make(map[uint64]bool, len(dbPeriods))
	for _, p := range dbPeriods {
		dbPeriodsMap[p] = true
	}
	currEpoch := utils.TimeToEpoch(time.Now())
	nextPeriod := utils.EpochToSyncPeriod(uint64(currEpoch) + 1)
	for p := uint64(0); p <= nextPeriod; p++ {
		_, exists := dbPeriodsMap[p]
		if !exists {
			c, err := rpcClient.GetSyncCommittee("head", p*utils.Config.Chain.Altair.EpochsPerSyncCommitteePeriod)
			if err != nil {
				return err
			}
			validatorsI64 := make([]int64, 0)
			for _, i := range c.Validators {
				parsed, err := strconv.ParseInt(i, 10, 64)
				if err != nil {
					return err
				}
				validatorsI64 = append(validatorsI64, parsed)
			}
			pqValidators := pq.Int64Array(validatorsI64)
			_, err = db.DB.Exec(
				`insert into sync_committees (period, validators) values ($1, $2)`,
				p,
				pqValidators,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
