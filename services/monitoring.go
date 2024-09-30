package services

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/go-redis/redis/v8"
)

func startMonitoringService(wg *sync.WaitGroup) {
	defer wg.Done()

	go startClDataMonitoringService()
	go startElDataMonitoringService()
	go startRedisMonitoringService()
	go startApiMonitoringService()
	go startAppMonitoringService()
	go startServicesMonitoringService()
}

// The cl data monitoring service will check that the data in the validators, blocks & epochs tables is up to date
func startClDataMonitoringService() {

	name := "monitoring_cl_data"
	firstRun := true
	for {
		if !firstRun {
			time.Sleep(time.Minute)
		}
		firstRun = false

		// retrieve the max attestationslot from the validators table and check that it is not older than 15 minutes
		var maxAttestationSlot uint64
		lastAttestationSlots, err := db.BigtableClient.GetLastAttestationSlots([]uint64{})
		if err != nil {
			logger.Errorf("error retrieving max attestation slot data from bigtable: %v", err)
			continue
		}

		for _, lastAttestationSlot := range lastAttestationSlots {
			if lastAttestationSlot > maxAttestationSlot {
				maxAttestationSlot = lastAttestationSlot
			}
		}

		if time.Since(utils.SlotToTime(maxAttestationSlot)) > time.Minute*15 {
			errorMsg := fmt.Errorf("error: max attestation slot is older than 15 minutes: %v", time.Since(utils.SlotToTime(maxAttestationSlot)))
			utils.LogError(nil, errorMsg, 0)
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}

		// retrieve the max slot from the blocks table and check tat it is not older than 15 minutes
		var maxSlot uint64
		err = db.WriterDb.Get(&maxSlot, "SELECT MAX(slot) FROM blocks;")
		if err != nil {
			logger.Errorf("error retrieving max slot from blocks table: %v", err)
			continue
		}

		if time.Since(utils.SlotToTime(maxSlot)) > time.Minute*15 {
			errorMsg := fmt.Errorf("error: max slot in blocks table is older than 15 minutes: %v", time.Since(utils.SlotToTime(maxAttestationSlot)))
			utils.LogError(nil, errorMsg, 0)
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}

		// retrieve the max epoch from the epochs table and check tat it is not older than 15 minutes
		var maxEpoch uint64
		err = db.WriterDb.Get(&maxEpoch, "SELECT MAX(epoch) FROM epochs;")
		if err != nil {
			logger.Errorf("error retrieving max slot from blocks table: %v", err)
			continue
		}

		if time.Since(utils.EpochToTime(maxEpoch)) > time.Minute*15 {
			errorMsg := fmt.Errorf("error: max epoch in epochs table is older than 15 minutes: %v", time.Since(utils.SlotToTime(maxAttestationSlot)))
			utils.LogError(nil, errorMsg, 0)
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}

		ReportStatus(name, "OK", nil)
	}
}

func startElDataMonitoringService() {

	name := "monitoring_el_data"
	firstRun := true
	for {
		if !firstRun {
			time.Sleep(time.Minute)
		}
		firstRun = false

		// check latest eth1 indexed block
		numberBlocksTable, err := db.BigtableClient.GetLastBlockInBlocksTable()
		if err != nil {
			errorMsg := fmt.Errorf("error: could not retrieve latest block number from the blocks table: %v", err)
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}
		blockBlocksTable, err := db.BigtableClient.GetBlockFromBlocksTable(uint64(numberBlocksTable))
		if err != nil {
			errorMsg := fmt.Errorf("error: could not retrieve latest block (%d) from the blocks table: %v", numberBlocksTable, err)
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}
		if blockBlocksTable.Time.AsTime().Before(time.Now().Add(time.Minute * -13)) {
			errorMsg := fmt.Errorf("error: last block in blocks table is more than 13 minutes old (check eth1 indexer)")
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}

		// check if eth1 indices are up to date
		numberDataTable, err := db.BigtableClient.GetLastBlockInDataTable()
		if err != nil {
			errorMsg := fmt.Errorf("error: could not retrieve latest block number from the data table: %v", err)
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}

		if numberDataTable < numberBlocksTable-32 {
			errorMsg := fmt.Errorf("error: data table is lagging behind the blocks table (check eth1 indexer)")
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}
		ReportStatus(name, "OK", nil)
	}
}

func startRedisMonitoringService() {

	name := "monitoring_redis"
	firstRun := true
	for {
		if !firstRun {
			time.Sleep(time.Minute)
		}
		firstRun = false

		rdc := redis.NewClient(&redis.Options{
			Addr: utils.Config.RedisCacheEndpoint,
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		if err := rdc.Ping(ctx).Err(); err != nil {
			cancel()
			ReportStatus(name, err.Error(), nil)
			rdc.Close()
			continue
		}
		cancel()
		rdc.Close()
		ReportStatus(name, "OK", nil)
	}
}

func startApiMonitoringService() {

	name := "monitoring_api"
	firstRun := true

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := "https://" + utils.Config.Frontend.SiteDomain + "/api/v1/epoch/latest"
	// add apikey (if any) to url but don't log the api key when errors occur
	errFields := map[string]interface{}{
		"url": url}
	url += "?apikey=" + utils.Config.Monitoring.ApiKey

	for {
		if !firstRun {
			time.Sleep(time.Minute)
		}
		firstRun = false

		resp, err := client.Get(url)
		if err != nil {
			utils.LogError(err, "getting client error", 0, errFields)
			ReportStatus(name, strings.ReplaceAll(err.Error(), utils.Config.Monitoring.ApiKey, ""), nil)
			continue
		}

		if resp.StatusCode != 200 {
			errorMsg := fmt.Errorf("error: api epoch / latest endpoint returned a non 200 status: %v", resp.StatusCode)
			utils.LogError(nil, errorMsg, 0, errFields)
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}

		ReportStatus(name, "OK", nil)
	}
}

func startAppMonitoringService() {

	name := "monitoring_app"
	firstRun := true

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := "https://" + utils.Config.Frontend.SiteDomain + "/api/v1/app/dashboard"
	// add apikey (if any) to url but don't log the api key when errors occur
	errFields := map[string]interface{}{
		"url": url}
	url += "?apikey=" + utils.Config.Monitoring.ApiKey

	for {
		if !firstRun {
			time.Sleep(time.Minute)
		}
		firstRun = false

		resp, err := client.Post(url, "application/json", strings.NewReader(`{"indicesOrPubkey": "1,2"}`))
		if err != nil {
			utils.LogError(err, "POST to dashboard URL error", 0, errFields)
			ReportStatus(name, strings.ReplaceAll(err.Error(), utils.Config.Monitoring.ApiKey, ""), nil)
			continue
		}

		if resp.StatusCode != 200 {
			errorMsg := fmt.Errorf("error: api app endpoint returned a non 200 status: %v", resp.StatusCode)
			utils.LogError(nil, errorMsg, 0, errFields)
			ReportStatus(name, errorMsg.Error(), nil)
			continue
		}

		ReportStatus(name, "OK", nil)
	}
}

func startServicesMonitoringService() {

	name := "monitoring_services"
	firstRun := true

	servicesToCheck := map[string]time.Duration{
		"eth1indexer":                 time.Minute * 15,
		"slotVizUpdater":              time.Minute * 15,
		"slotUpdater":                 time.Minute * 15,
		"latestProposedSlotUpdater":   time.Minute * 15,
		"epochUpdater":                time.Minute * 15,
		"rewardsExporter":             time.Minute * 15,
		"mempoolUpdater":              time.Minute * 15,
		"indexPageDataUpdater":        time.Minute * 15,
		"latestBlockUpdater":          time.Minute * 15,
		"headBlockRootHashUpdater":    time.Minute * 15,
		"notification-collector":      time.Minute * 15,
		"relaysUpdater":               time.Minute * 15,
		"ethstoreExporter":            time.Minute * 60,
		"statsUpdater":                time.Minute * 30,
		"poolsUpdater":                time.Minute * 30,
		"slotExporter":                time.Minute * 15,
		"statistics":                  time.Minute * 90,
		"ethStoreStatistics":          time.Minute * 15,
		"lastExportedStatisticDay":    time.Minute * 15,
		"validatorStateCountsUpdater": time.Minute * 90,
		//"notification-sender", //exclude for now as the sender is only running on mainnet
	}

	if utils.Config.Monitoring.ServiceMonitoringConfigurations != nil {
		for _, service := range utils.Config.Monitoring.ServiceMonitoringConfigurations {
			if service.Duration == 0 {
				delete(servicesToCheck, service.Name)
				logger.Infof("Removing %v from monitoring service", service.Name)
			} else {
				servicesToCheck[service.Name] = service.Duration
				logger.Infof("Change timeout for %v to %v", service.Name, service.Duration)
			}
		}
	}

	for {
		if !firstRun {
			time.Sleep(time.Minute)
		}
		firstRun = false

		now := time.Now()
		hasError := false
		for serviceName, maxTimeout := range servicesToCheck {
			var status string
			err := db.WriterDb.Get(&status, `select status from service_status where last_update > $1 and name = $2 ORDER BY last_update DESC LIMIT 1;`, now.Add(maxTimeout*-1), serviceName)

			if err != nil {

				if err == sql.ErrNoRows {
					errorMsg := fmt.Errorf("error: missing status entry for service %v", serviceName)
					utils.LogError(err, errorMsg, 0)
					ReportStatus(name, errorMsg.Error(), nil)
					hasError = true
					break
				} else {
					errorMsg := fmt.Errorf("error: could not retrieve service status from the service_status table: %v", err)
					ReportStatus(name, errorMsg.Error(), nil)
					hasError = true
					break
				}
			}

			if status != "Running" {
				errorMsg := fmt.Errorf("error: service %v has unexpected state %v", serviceName, status)
				ReportStatus(name, errorMsg.Error(), nil)
				hasError = true
				break
			}
		}

		if !hasError {
			ReportStatus(name, "OK", nil)
		}

		_, err := db.WriterDb.Exec("DELETE FROM service_status WHERE last_update < NOW() - INTERVAL '1 WEEK'")

		if err != nil {
			logger.Errorf("error cleaning up service_status table")
		}
	}

}
