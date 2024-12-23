package db

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/lib/pq"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/go-redis/redis/v8"
	itypes "github.com/gobitfly/eth-rewards/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

var BigtableClient *Bigtable

const (
	DEFAULT_FAMILY                        = "f"
	VALIDATOR_BALANCES_FAMILY             = "vb"
	VALIDATOR_HIGHEST_ACTIVE_INDEX_FAMILY = "ha"
	ATTESTATIONS_FAMILY                   = "at"
	SYNC_COMMITTEES_FAMILY                = "sc"
	INCOME_DETAILS_COLUMN_FAMILY          = "id"
	STATS_COLUMN_FAMILY                   = "stats"
	MACHINE_METRICS_COLUMN_FAMILY         = "mm"
	SERIES_FAMILY                         = "series"

	SUM_COLUMN = "sum"

	MAX_CL_BLOCK_NUMBER = 1000000000 - 1
	MAX_EL_BLOCK_NUMBER = 1000000000
	MAX_EPOCH           = 1000000000 - 1

	max_block_number_v1 = 1000000000
	max_epoch_v1        = 1000000000

	MAX_BATCH_MUTATIONS   = 100000
	DEFAULT_BATCH_INSERTS = 10000

	REPORT_TIMEOUT = time.Second * 10
)

type Bigtable struct {
	client *gcp_bigtable.Client

	tableBeaconchain       *gcp_bigtable.Table
	tableValidators        *gcp_bigtable.Table
	tableValidatorsHistory *gcp_bigtable.Table

	tableData            *gcp_bigtable.Table
	tableBlocks          *gcp_bigtable.Table
	tableMetadataUpdates *gcp_bigtable.Table
	tableMetadata        *gcp_bigtable.Table

	tableMachineMetrics *gcp_bigtable.Table

	redisCache *redis.Client

	LastAttestationCache    map[uint64]uint64
	LastAttestationCacheMux *sync.Mutex

	chainId string

	v2SchemaCutOffEpoch uint64

	machineMetricsQueuedWritesChan chan (types.BulkMutation)
}

func InitBigtable(project, instance, chainId, redisAddress string) (*Bigtable, error) {

	if utils.Config.Bigtable.Emulator {

		if utils.Config.Bigtable.EmulatorHost == "" {
			utils.Config.Bigtable.EmulatorHost = "127.0.0.1"
		}
		logger.Infof("using emulated local bigtable environment, setting BIGTABLE_EMULATOR_HOST env variable to %s:%d", utils.Config.Bigtable.EmulatorHost, utils.Config.Bigtable.EmulatorPort)
		err := os.Setenv("BIGTABLE_EMULATOR_HOST", fmt.Sprintf("%s:%d", utils.Config.Bigtable.EmulatorHost, utils.Config.Bigtable.EmulatorPort))

		if err != nil {
			logger.Fatalf("unable to set bigtable emulator environment variable: %v", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	poolSize := 50
	btClient, err := gcp_bigtable.NewClient(ctx, project, instance, option.WithGRPCConnectionPool(poolSize))
	// btClient, err := gcp_bigtable.NewClient(context.Background(), project, instance)
	if err != nil {
		return nil, err
	}

	rdc := redis.NewClient(&redis.Options{
		Addr:        redisAddress,
		ReadTimeout: time.Second * 20,
	})

	if err := rdc.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	bt := &Bigtable{
		client:                         btClient,
		tableData:                      btClient.Open("data"),
		tableBlocks:                    btClient.Open("blocks"),
		tableMetadataUpdates:           btClient.Open("metadata_updates"),
		tableMetadata:                  btClient.Open("metadata"),
		tableBeaconchain:               btClient.Open("beaconchain"),
		tableMachineMetrics:            btClient.Open("machine_metrics"),
		tableValidators:                btClient.Open("beaconchain_validators"),
		tableValidatorsHistory:         btClient.Open("beaconchain_validators_history"),
		chainId:                        chainId,
		redisCache:                     rdc,
		LastAttestationCacheMux:        &sync.Mutex{},
		v2SchemaCutOffEpoch:            utils.Config.Bigtable.V2SchemaCutOffEpoch,
		machineMetricsQueuedWritesChan: make(chan types.BulkMutation, MAX_BATCH_MUTATIONS),
	}

	if utils.Config.Frontend.Enabled { // Only activate machine metrics inserts on frontend / api instances
		go bt.commitQueuedMachineMetricWrites()
	}

	BigtableClient = bt
	return bt, nil
}

func (bigtable *Bigtable) commitQueuedMachineMetricWrites() {

	// copy the pending mutations over and commit them

	batchSize := 10000

	muts := types.NewBulkMutations(batchSize)

	tmr := time.NewTicker(time.Second * 10)
	for {
		select {
		case mut, ok := <-bigtable.machineMetricsQueuedWritesChan:

			if ok {
				muts.Keys = append(muts.Keys, mut.Key)
				muts.Muts = append(muts.Muts, mut.Mut)
			}

			if len(muts.Keys) >= batchSize || !ok && len(muts.Keys) > 0 { // commit when batch size is reached or on channel close
				logger.Infof("committing %v queued machine metric inserts (trigger=batchSize, ok=%v)", len(muts.Keys), ok)
				err := bigtable.WriteBulk(muts, bigtable.tableMachineMetrics, batchSize)

				if err == nil {
					muts = types.NewBulkMutations(batchSize)
				} else {
					logger.Errorf("error writing queued machine metrics to bigtable: %v", err)
				}
			}

			if !ok { // insert chan is closed, stop the timer and exit
				tmr.Stop()
				logger.Infof("stopping batched machine metrics insert")
				return
			}

		case <-tmr.C:
			if len(muts.Keys) > 0 {
				logger.Infof("committing %v queued machine metric inserts (trigger=timeout)", len(muts.Keys))
				err := bigtable.WriteBulk(muts, bigtable.tableMachineMetrics, DEFAULT_BATCH_INSERTS)

				if err == nil {
					muts = types.NewBulkMutations(batchSize)
				} else {
					logger.Errorf("error writing queued machine metrics to bigtable: %v", err)
				}
			}
		}
	}

}

func (bigtable *Bigtable) Close() {
	close(bigtable.machineMetricsQueuedWritesChan)
	time.Sleep(time.Second * 5)
	bigtable.client.Close()
}

func (bigtable *Bigtable) GetClient() *gcp_bigtable.Client {
	return bigtable.client
}

func (bigtable *Bigtable) SaveMachineMetric(process string, userID uint64, machine string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	rowKeyData := fmt.Sprintf("u:%s:p:%s:m:%v", bigtable.reversePaddedUserID(userID), process, machine)

	ts := gcp_bigtable.Now()
	rateLimitKey := fmt.Sprintf("%s:%d", rowKeyData, ts.Time().Minute())
	keySet, err := bigtable.redisCache.SetNX(ctx, rateLimitKey, "1", time.Minute).Result()
	if err != nil {
		return err
	}
	if !keySet {
		return fmt.Errorf("rate limit, last metric insert was less than 1 min ago")
	}

	// for limiting machines per user, add the machine field to a redis set
	// bucket period is 15mins
	machineLimitKey := fmt.Sprintf("%s:%d", bigtable.reversePaddedUserID(userID), ts.Time().Minute()%15)
	pipe := bigtable.redisCache.Pipeline()
	pipe.SAdd(ctx, machineLimitKey, machine)
	pipe.Expire(ctx, machineLimitKey, time.Minute*15)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	dataMut := gcp_bigtable.NewMutation()
	dataMut.Set(MACHINE_METRICS_COLUMN_FAMILY, "v1", ts, data)

	bulkMut := types.BulkMutation{ // schedule the mutation for writing
		Key: rowKeyData,
		Mut: dataMut,
	}

	bigtable.machineMetricsQueuedWritesChan <- bulkMut

	return nil
}

func (bigtable Bigtable) getMachineMetricNamesMap(userID uint64, searchDepth int) (map[string]bool, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	rangePrefix := fmt.Sprintf("u:%s:p:", bigtable.reversePaddedUserID(userID))

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(MACHINE_METRICS_COLUMN_FAMILY),
		gcp_bigtable.LatestNFilter(searchDepth),
		gcp_bigtable.TimestampRangeFilter(time.Now().Add(time.Duration(searchDepth*-1)*time.Minute), time.Now()),
		gcp_bigtable.StripValueFilter(),
	)

	machineNames := make(map[string]bool)

	err := bigtable.tableMachineMetrics.ReadRows(ctx, gcp_bigtable.PrefixRange(rangePrefix), func(r gcp_bigtable.Row) bool {
		success, _, machine, _ := machineMetricRowParts(r.Key())
		if !success {
			return false
		}
		machineNames[machine] = true

		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return machineNames, err
	}

	return machineNames, nil
}

func (bigtable Bigtable) GetMachineMetricsMachineNames(userID uint64) ([]string, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"userId": userID,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	names, err := bigtable.getMachineMetricNamesMap(userID, 300)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for key := range names {
		result = append(result, key)
	}

	return result, nil
}

func (bigtable Bigtable) GetMachineMetricsMachineCount(userID uint64) (uint64, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"userId": userID,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	machineLimitKey := fmt.Sprintf("%s:%d", bigtable.reversePaddedUserID(userID), time.Now().Minute()%15)

	card, err := bigtable.redisCache.SCard(ctx, machineLimitKey).Result()
	if err != nil {
		return 0, err
	}
	return uint64(card), nil
}

func (bigtable Bigtable) GetMachineMetricsNode(userID uint64, limit, offset int) ([]*types.MachineMetricNode, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"userId": userID,
			"limit":  limit,
			"offset": offset,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	return getMachineMetrics(bigtable, "beaconnode", userID, limit, offset,
		func(data []byte, machine string) *types.MachineMetricNode {
			obj := &types.MachineMetricNode{}
			err := proto.Unmarshal(data, obj)
			if err != nil {
				return nil
			}
			obj.Machine = &machine
			return obj
		},
	)
}

func (bigtable Bigtable) GetMachineMetricsValidator(userID uint64, limit, offset int) ([]*types.MachineMetricValidator, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"userId": userID,
			"limit":  limit,
			"offset": offset,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	return getMachineMetrics(bigtable, "validator", userID, limit, offset,
		func(data []byte, machine string) *types.MachineMetricValidator {
			obj := &types.MachineMetricValidator{}
			err := proto.Unmarshal(data, obj)
			if err != nil {
				return nil
			}
			obj.Machine = &machine
			return obj
		},
	)
}

func (bigtable Bigtable) GetMachineMetricsSystem(userID uint64, limit, offset int) ([]*types.MachineMetricSystem, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"userId": userID,
			"limit":  limit,
			"offset": offset,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	return getMachineMetrics(bigtable, "system", userID, limit, offset,
		func(data []byte, machine string) *types.MachineMetricSystem {
			obj := &types.MachineMetricSystem{}
			err := proto.Unmarshal(data, obj)
			if err != nil {
				return nil
			}
			obj.Machine = &machine
			return obj
		},
	)
}

func getMachineMetrics[T types.MachineMetricSystem | types.MachineMetricNode | types.MachineMetricValidator](bigtable Bigtable, process string, userID uint64, limit, offset int, marshler func(data []byte, machine string) *T) ([]*T, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	rangePrefix := fmt.Sprintf("u:%s:p:%s:m:", bigtable.reversePaddedUserID(userID), process)
	res := make([]*T, 0)
	if offset <= 0 {
		offset = 1
	}

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(MACHINE_METRICS_COLUMN_FAMILY),
		gcp_bigtable.LatestNFilter(limit),
		gcp_bigtable.CellsPerRowOffsetFilter(offset),
	)
	gapSize := getMachineStatsGap(uint64(limit))
	err := bigtable.tableMachineMetrics.ReadRows(ctx, gcp_bigtable.PrefixRange(rangePrefix), func(r gcp_bigtable.Row) bool {
		success, _, machine, _ := machineMetricRowParts(r.Key())
		if !success {
			return false
		}
		var count = -1
		for _, ri := range r[MACHINE_METRICS_COLUMN_FAMILY] {
			count++
			if count%gapSize != 0 {
				continue
			}

			obj := marshler(ri.Value, machine)
			if obj == nil {
				return false
			}

			res = append(res, obj)
		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable Bigtable) GetMachineRowKey(userID uint64, process string, machine string) string {
	return fmt.Sprintf("u:%s:p:%s:m:%s", bigtable.reversePaddedUserID(userID), process, machine)
}

// Returns a map[userID]map[machineName]machineData
// machineData contains the latest machine data in CurrentData
// and 5 minute old data in fiveMinuteOldData (defined in limit)
// as well as the insert timestamps of both
func (bigtable Bigtable) GetMachineMetricsForNotifications(rowKeys gcp_bigtable.RowList) (map[uint64]map[string]*types.MachineMetricSystemUser, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"rowKeys": rowKeys,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*200))
	defer cancel()

	res := make(map[uint64]map[string]*types.MachineMetricSystemUser) // userID -> machine -> data

	limit := 5

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(MACHINE_METRICS_COLUMN_FAMILY),
		gcp_bigtable.LatestNFilter(limit),
	)

	err := bigtable.tableMachineMetrics.ReadRows(ctx, rowKeys, func(r gcp_bigtable.Row) bool {
		success, userID, machine, _ := machineMetricRowParts(r.Key())
		if !success {
			return false
		}

		count := 0
		for _, ri := range r[MACHINE_METRICS_COLUMN_FAMILY] {

			obj := &types.MachineMetricSystem{}
			err := proto.Unmarshal(ri.Value, obj)
			if err != nil {
				return false
			}

			if _, found := res[userID]; !found {
				res[userID] = make(map[string]*types.MachineMetricSystemUser)
			}

			last, found := res[userID][machine]

			if found && count == limit-1 {
				res[userID][machine] = &types.MachineMetricSystemUser{
					UserID:                    userID,
					Machine:                   machine,
					CurrentData:               last.CurrentData,
					FiveMinuteOldData:         obj,
					CurrentDataInsertTs:       last.CurrentDataInsertTs,
					FiveMinuteOldDataInsertTs: ri.Timestamp.Time().Unix(),
				}
			} else {
				res[userID][machine] = &types.MachineMetricSystemUser{
					UserID:                    userID,
					Machine:                   machine,
					CurrentData:               obj,
					FiveMinuteOldData:         nil,
					CurrentDataInsertTs:       ri.Timestamp.Time().Unix(),
					FiveMinuteOldDataInsertTs: 0,
				}
			}
			count++

		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func machineMetricRowParts(r string) (bool, uint64, string, string) {
	keySplit := strings.Split(r, ":")

	userID, err := strconv.ParseUint(keySplit[1], 10, 64)
	if err != nil {
		logger.Errorf("error parsing slot from row key %v: %v", r, err)
		return false, 0, "", ""
	}
	userID = ^uint64(0) - userID

	machine := ""
	if len(keySplit) >= 6 {
		machine = keySplit[5]
	}

	process := keySplit[3]

	return true, userID, machine, process
}

func (bigtable *Bigtable) SaveValidatorBalances(epoch uint64, validators []*types.Validator) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	// start := time.Now()
	ts := gcp_bigtable.Time(utils.EpochToTime(epoch))

	muts := types.NewBulkMutations(len(validators))

	highestActiveIndex := uint64(0)
	epochKey := bigtable.reversedPaddedEpoch(epoch)

	for _, validator := range validators {

		if validator.Balance > 0 && validator.Index > highestActiveIndex {
			highestActiveIndex = validator.Index
		}

		balanceEncoded := make([]byte, 8)
		binary.LittleEndian.PutUint64(balanceEncoded, validator.Balance)
		effectiveBalanceEncoded := uint8(validator.EffectiveBalance / 1e9) // we can encode the effective balance in 1 byte as it is capped at 32ETH and only decrements in 1 ETH steps

		combined := append(balanceEncoded, effectiveBalanceEncoded)
		mut := &gcp_bigtable.Mutation{}
		mut.Set(VALIDATOR_BALANCES_FAMILY, "b", ts, combined)
		key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator.Index), VALIDATOR_BALANCES_FAMILY, epochKey)

		muts.Add(key, mut)
	}

	err := bigtable.WriteBulk(muts, bigtable.tableValidatorsHistory, MAX_BATCH_MUTATIONS)
	if err != nil {
		return err
	}

	// store the highes active validator index for that epoch
	highestActiveIndexEncoded := make([]byte, 8)
	binary.LittleEndian.PutUint64(highestActiveIndexEncoded, highestActiveIndex)

	mut := &gcp_bigtable.Mutation{}
	mut.Set(VALIDATOR_HIGHEST_ACTIVE_INDEX_FAMILY, VALIDATOR_HIGHEST_ACTIVE_INDEX_FAMILY, ts, highestActiveIndexEncoded)
	key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, VALIDATOR_HIGHEST_ACTIVE_INDEX_FAMILY, epochKey)
	err = bigtable.tableValidatorsHistory.Apply(ctx, key, mut)
	if err != nil {
		return err
	}
	return nil
}

func (bigtable *Bigtable) SaveAttestationDuties(duties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error {

	// Initialize in memory last attestation cache lazily
	bigtable.LastAttestationCacheMux.Lock()
	if bigtable.LastAttestationCache == nil {
		t := time.Now()
		var err error
		bigtable.LastAttestationCache, err = bigtable.GetLastAttestationSlots([]uint64{})

		if err != nil {
			bigtable.LastAttestationCacheMux.Unlock()
			return err
		}
		logger.Infof("initialized in memory last attestation slot cache with %v validators in %v", len(bigtable.LastAttestationCache), time.Since(t))

	}
	bigtable.LastAttestationCacheMux.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	start := time.Now()

	mutsInclusionSlot := types.NewBulkMutations(MAX_BATCH_MUTATIONS)

	mutLastAttestationSlot := gcp_bigtable.NewMutation()
	mutLastAttestationSlotCount := 0

	for attestedSlot, validators := range duties {
		for validator, inclusions := range validators {
			epoch := utils.EpochOfSlot(uint64(attestedSlot))
			bigtable.LastAttestationCacheMux.Lock()
			if len(inclusions) == 0 { // for missed attestations we write the the attested slot as inclusion slot (attested == included means no attestation duty was performed)
				inclusions = append(inclusions, attestedSlot)
			}
			for _, inclusionSlot := range inclusions {
				key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(uint64(validator)), ATTESTATIONS_FAMILY, bigtable.reversedPaddedEpoch(epoch))

				mutInclusionSlot := gcp_bigtable.NewMutation()
				ts := gcp_bigtable.Time(utils.SlotToTime(uint64(inclusionSlot)))
				mutInclusionSlot.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", attestedSlot), ts, []byte{})

				mutsInclusionSlot.Add(key, mutInclusionSlot)

				if inclusionSlot != MAX_CL_BLOCK_NUMBER && uint64(attestedSlot) > bigtable.LastAttestationCache[uint64(validator)] {
					mutLastAttestationSlot.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", validator), gcp_bigtable.Timestamp((attestedSlot)*1000), []byte{})
					bigtable.LastAttestationCache[uint64(validator)] = uint64(attestedSlot)
					mutLastAttestationSlotCount++

					if mutLastAttestationSlotCount == MAX_BATCH_MUTATIONS {
						mutStart := time.Now()
						err := bigtable.tableValidators.Apply(ctx, fmt.Sprintf("%s:lastAttestationSlot", bigtable.chainId), mutLastAttestationSlot)
						if err != nil {
							bigtable.LastAttestationCacheMux.Unlock()
							return fmt.Errorf("error applying last attestation slot mutations: %v", err)
						}
						mutLastAttestationSlot = gcp_bigtable.NewMutation()
						mutLastAttestationSlotCount = 0
						logger.Infof("applied last attestation slot mutations in %v", time.Since(mutStart))
					}
				}
			}
			bigtable.LastAttestationCacheMux.Unlock()
		}
	}

	err := bigtable.WriteBulk(mutsInclusionSlot, bigtable.tableValidatorsHistory, MAX_BATCH_MUTATIONS)

	if err != nil {
		return fmt.Errorf("error writing attestation inclusion slot mutations: %v", err)
	}

	if mutLastAttestationSlotCount > 0 {
		err := bigtable.tableValidators.Apply(ctx, fmt.Sprintf("%s:lastAttestationSlot", bigtable.chainId), mutLastAttestationSlot)
		if err != nil {
			return fmt.Errorf("error applying last attestation slot mutations: %v", err)
		}
	}

	logger.Infof("exported %v attestations to bigtable in %v", mutsInclusionSlot.Len(), time.Since(start))
	return nil
}

// This method is only to be used for migrating the last attestation slot to bigtable and should not be used for any other purpose
func (bigtable *Bigtable) SetLastAttestationSlot(validator uint64, lastAttestationSlot uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	mutLastAttestationSlot := gcp_bigtable.NewMutation()
	mutLastAttestationSlot.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", validator), gcp_bigtable.Timestamp(lastAttestationSlot*1000), []byte{})
	err := bigtable.tableValidators.Apply(ctx, fmt.Sprintf("%s:lastAttestationSlot", bigtable.chainId), mutLastAttestationSlot)
	if err != nil {
		return err
	}

	return nil
}

func (bigtable *Bigtable) SaveSyncComitteeDuties(duties map[types.Slot]map[types.ValidatorIndex]bool) error {
	start := time.Now()

	if len(duties) == 0 {
		logger.Infof("no sync duties to export")
		return nil
	}

	muts := types.NewBulkMutations(int(utils.Config.Chain.ClConfig.SlotsPerEpoch*utils.Config.Chain.ClConfig.SyncCommitteeSize + 1))

	for slot, validators := range duties {
		for validator, participated := range validators {
			mut := gcp_bigtable.NewMutation()
			if participated {
				ts := gcp_bigtable.Time(utils.SlotToTime(uint64(slot)).Add(time.Second)) // add 1 second to avoid collisions with duties
				mut.Set(SYNC_COMMITTEES_FAMILY, "s", ts, []byte{})
			} else {
				ts := gcp_bigtable.Time(utils.SlotToTime(uint64(slot)))
				mut.Set(SYNC_COMMITTEES_FAMILY, "s", ts, []byte{})
			}
			key := fmt.Sprintf("%s:%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(uint64(validator)), SYNC_COMMITTEES_FAMILY, bigtable.reversedPaddedEpoch(utils.EpochOfSlot(uint64(slot))), bigtable.reversedPaddedSlot(uint64(slot)))

			muts.Add(key, mut)
		}
	}

	err := bigtable.WriteBulk(muts, bigtable.tableValidatorsHistory, MAX_BATCH_MUTATIONS)
	if err != nil {
		return err
	}

	logger.Infof("exported %v sync committee duties to bigtable in %v", muts.Len(), time.Since(start))
	return nil
}

// GetMaxValidatorindexForEpoch returns the higest validatorindex with a balance at that epoch
// Clickhouse Port: Not required
func (bigtable *Bigtable) GetMaxValidatorindexForEpoch(epoch uint64) (uint64, error) {
	return bigtable.getMaxValidatorindexForEpochV2(epoch)
}

func (bigtable *Bigtable) getMaxValidatorindexForEpochV2(epoch uint64) (uint64, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"epoch": epoch,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, VALIDATOR_HIGHEST_ACTIVE_INDEX_FAMILY, bigtable.reversedPaddedEpoch(epoch))

	row, err := bigtable.tableValidatorsHistory.ReadRow(ctx, key)
	if err != nil {
		return 0, err
	}

	for _, ri := range row[VALIDATOR_HIGHEST_ACTIVE_INDEX_FAMILY] {
		return binary.LittleEndian.Uint64(ri.Value), nil
	}

	return 0, nil
}

// Clickhouse port: Done
func (bigtable *Bigtable) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {
	// Disable balance fetching from clickhouse until we have a compatible data set available
	// endEpochTs := utils.EpochToTime(endEpoch)
	// if utils.Config.ClickHouseEnabled && time.Since(endEpochTs) > utils.Config.ClickhouseDelay { // fetch data from clickhouse instead
	// 	logger.Infof("fetching validator balance history from clickhouse for validators %v, epochs %v - %v", validators, startEpoch, endEpoch)
	// 	return bigtable.getValidatorBalanceHistoryClickhouse(validators, startEpoch, endEpoch)
	// } else
	if endEpoch < bigtable.v2SchemaCutOffEpoch {
		return bigtable.getValidatorBalanceHistoryV1(validators, startEpoch, endEpoch)
	} else {
		return bigtable.getValidatorBalanceHistoryV2(validators, startEpoch, endEpoch)
	}
}

//lint:ignore U1000 will be used later on
func (bigtable *Bigtable) getValidatorBalanceHistoryClickhouse(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {
	startEpochTs := utils.EpochToTime(startEpoch)
	endEpochTs := utils.EpochToTime(endEpoch)

	type row struct {
		ValidatorIndex        uint64 `db:"validator_index"`
		Epoch                 uint64 `db:"epoch"`
		BalanceStart          int64  `db:"balance_start"`
		EffectiveBalanceStart int64  `db:"balance_effective_start"`
		BalanceEnd            int64  `db:"balance_end"`
		EffectiveBalanceEnd   int64  `db:"balance_effective_end"`
	}
	rows := []*row{}

	query := `
			SELECT 
				validator_index, 
				epoch AS epoch, 
				balance_start, 
				balance_effective_start,
				balance_end, 
				balance_effective_end
			FROM validator_dashboard_data_epoch FINAL WHERE epoch_timestamp >= ? AND epoch_timestamp <= ? AND validator_index IN (?) ORDER BY epoch ASC`

	err := ClickhouseReaderDb.Select(&rows, query, startEpochTs, endEpochTs, validators)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	res := make(map[uint64][]*types.ValidatorBalance, len(validators))
	for _, r := range rows {
		if res[r.ValidatorIndex] == nil {
			res[r.ValidatorIndex] = make([]*types.ValidatorBalance, 0)
		}
		balance := &types.ValidatorBalance{
			Epoch:            r.Epoch,
			Balance:          uint64(r.BalanceEnd),
			EffectiveBalance: uint64(r.EffectiveBalanceEnd),
			Index:            r.ValidatorIndex,
			PublicKey:        []byte{},
		}

		res[r.ValidatorIndex] = append(res[r.ValidatorIndex], balance)
	}

	for validator, att := range res {
		sort.Slice(att, func(i, j int) bool {
			return att[i].Epoch > att[j].Epoch
		})
		res[validator] = att
	}

	return res, nil
}

func (bigtable *Bigtable) getValidatorBalanceHistoryV2(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"validators_count": len(validators),
			"startEpoch":       startEpoch,
			"endEpoch":         endEpoch,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	res := make(map[uint64][]*types.ValidatorBalance, len(validators))
	resMux := &sync.Mutex{}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i := 0; i < len(validators); i += batchSize {

		upperBound := i + batchSize
		if len(validators) < upperBound {
			upperBound = len(validators)
		}
		vals := validators[i:upperBound]

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			ranges := bigtable.getValidatorsEpochRanges(vals, VALIDATOR_BALANCES_FAMILY, startEpoch, endEpoch)
			ro := gcp_bigtable.LimitRows(int64(endEpoch-startEpoch+1) * int64(len(vals)))

			handleRow := func(r gcp_bigtable.Row) bool {
				// logger.Info(r.Key())
				keySplit := strings.Split(r.Key(), ":")

				epoch, err := strconv.ParseUint(keySplit[3], 10, 64)
				if err != nil {
					logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
					return false
				}

				validator, err := bigtable.validatorKeyToIndex(keySplit[1])
				if err != nil {
					logger.Errorf("error parsing validator index from row key %v: %v", r.Key(), err)
					return false
				}
				resMux.Lock()
				if res[validator] == nil {
					res[validator] = make([]*types.ValidatorBalance, 0)
				}
				resMux.Unlock()

				for _, ri := range r[VALIDATOR_BALANCES_FAMILY] {
					balances := ri.Value

					balanceBytes := balances[0:8]
					balance := binary.LittleEndian.Uint64(balanceBytes)
					var effectiveBalance uint64
					if len(balances) == 9 { // in new schema the effective balance is encoded in 1 byte
						effectiveBalance = uint64(balances[8]) * 1e9
					} else {
						effectiveBalanceBytes := balances[8:16]
						effectiveBalance = binary.LittleEndian.Uint64(effectiveBalanceBytes)
					}

					resMux.Lock()
					res[validator] = append(res[validator], &types.ValidatorBalance{
						Epoch:            MAX_EPOCH - epoch,
						Balance:          balance,
						EffectiveBalance: effectiveBalance,
						Index:            validator,
						PublicKey:        []byte{},
					})
					resMux.Unlock()
				}
				return true
			}

			err := bigtable.tableValidatorsHistory.ReadRows(gCtx, ranges, handleRow, gcp_bigtable.RowFilter(gcp_bigtable.LatestNFilter(1)), ro)
			if err != nil {
				return err
			}

			// logrus.Infof("retrieved data for validators %v - %v", vals[0], vals[len(vals)-1])
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) getValidatorBalanceHistoryV1(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {

	valLen := len(validators)
	getAllThreshold := 1000
	validatorMap := make(map[uint64]bool, valLen)
	for _, validatorIndex := range validators {
		validatorMap[validatorIndex] = true
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	ranges := bigtable.getEpochRangesV1(startEpoch, endEpoch)
	res := make(map[uint64][]*types.ValidatorBalance, valLen)

	columnFilters := []gcp_bigtable.Filter{}
	if valLen < getAllThreshold {
		columnFilters = make([]gcp_bigtable.Filter, 0, valLen)
		for _, validator := range validators {
			columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
		}
	}

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(VALIDATOR_BALANCES_FAMILY),
		gcp_bigtable.InterleaveFilters(columnFilters...),
	)

	if len(columnFilters) == 1 { // special case to retrieve data for one validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(VALIDATOR_BALANCES_FAMILY),
			columnFilters[0],
		)
	}
	if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.FamilyFilter(VALIDATOR_BALANCES_FAMILY)
	}

	handleRow := func(r gcp_bigtable.Row) bool {
		keySplit := strings.Split(r.Key(), ":")

		epoch, err := strconv.ParseUint(keySplit[3], 10, 64)
		if err != nil {
			logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
			return false
		}

		for _, ri := range r[VALIDATOR_BALANCES_FAMILY] {
			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, VALIDATOR_BALANCES_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			// If we requested more than getAllThreshold validators we will
			// get data for all validators and need to filter out all
			// unwanted ones
			if valLen >= getAllThreshold && !validatorMap[validator] {
				continue
			}

			balances := ri.Value

			balanceBytes := balances[0:8]
			effectiveBalanceBytes := balances[8:16]
			balance := binary.LittleEndian.Uint64(balanceBytes)
			effectiveBalance := binary.LittleEndian.Uint64(effectiveBalanceBytes)

			if res[validator] == nil {
				res[validator] = make([]*types.ValidatorBalance, 0)
			}

			res[validator] = append(res[validator], &types.ValidatorBalance{
				Epoch:            max_epoch_v1 - epoch,
				Balance:          balance,
				EffectiveBalance: effectiveBalance,
				Index:            validator,
				PublicKey:        []byte{},
			})
		}
		return true
	}

	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, handleRow, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Clickhouse port: Done
func (bigtable *Bigtable) GetValidatorAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error) {
	endEpochTs := utils.EpochToTime(endEpoch)
	if utils.Config.ClickHouseEnabled && time.Since(endEpochTs) > utils.Config.ClickhouseDelay { // fetch data from clickhouse instead
		logger.Infof("fetching validator attestation history from clickhouse for validators %v, epochs %v - %v", validators, startEpoch, endEpoch)
		return bigtable.getValidatorAttestationHistoryClickhouse(validators, startEpoch, endEpoch)
	} else if endEpoch < bigtable.v2SchemaCutOffEpoch {
		return bigtable.getValidatorAttestationHistoryV1(validators, startEpoch, endEpoch)
	} else {
		return bigtable.getValidatorAttestationHistoryV2(validators, startEpoch, endEpoch)
	}
}

func (bigtable *Bigtable) getValidatorAttestationHistoryClickhouse(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error) {
	startEpochTs := utils.EpochToTime(startEpoch)
	endEpochTs := utils.EpochToTime(endEpoch)

	type row struct {
		ValidatorIndex        uint64 `db:"validator_index"`
		Epoch                 uint64 `db:"epoch"`
		Slot                  uint64 `db:"slot"`
		InclusionDelay        uint64 `db:"inclusion_delay_sum"`
		OptimalInclusionDelay uint64 `db:"optimal_inclusion_delay_sum"`
		AttestationsScheduled uint64 `db:"attestations_scheduled"`
		AttestationsObserved  uint64 `db:"attestations_observed"`
	}

	rows := []*row{}

	query := `
		SELECT
			validator_dashboard_data_epoch.validator_index,
			validator_dashboard_data_epoch.epoch,
			validator_attestation_assignments_slot.slot,
			validator_dashboard_data_epoch.inclusion_delay_sum,
			validator_dashboard_data_epoch.optimal_inclusion_delay_sum,
			validator_dashboard_data_epoch.attestations_scheduled,
			validator_dashboard_data_epoch.attestations_observed
		FROM validator_dashboard_data_epoch FINAL 
		LEFT JOIN validator_attestation_assignments_slot FINAL ON 
			validator_dashboard_data_epoch.validator_index = validator_attestation_assignments_slot.validator_index AND
			validator_dashboard_data_epoch.epoch_timestamp = validator_attestation_assignments_slot.epoch_timestamp AND
			validator_dashboard_data_epoch.epoch = validator_attestation_assignments_slot.epoch
		WHERE 
			validator_dashboard_data_epoch.epoch_timestamp >= ? AND 
			validator_dashboard_data_epoch.epoch_timestamp <= ? AND 
			validator_dashboard_data_epoch.validator_index IN (?) 
		ORDER BY epoch ASC, slot ASC`
	err := ClickhouseReaderDb.Select(&rows, query, startEpochTs, endEpochTs, validators)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	res := make(map[uint64][]*types.ValidatorAttestation, len(validators))

	for _, r := range rows {
		if r.AttestationsScheduled == 0 {
			continue
		}
		if res[r.ValidatorIndex] == nil {
			res[r.ValidatorIndex] = make([]*types.ValidatorAttestation, 0)
		}

		attestation := &types.ValidatorAttestation{
			Index:        r.ValidatorIndex,
			Epoch:        r.Epoch,
			AttesterSlot: r.Slot,
		}

		if r.AttestationsScheduled == r.AttestationsObserved {
			attestation.Status = 1
			attestation.InclusionSlot = r.Slot + r.InclusionDelay + 1
			attestation.Delay = int64(r.OptimalInclusionDelay)
		} else {
			attestation.Delay = 0 - int64(r.Slot) - 1 // bug for bug compatibility *sigh*
		}

		res[r.ValidatorIndex] = append(res[r.ValidatorIndex], attestation)
	}

	// Sort the result by attesterSlot desc
	for validator, att := range res {
		sort.Slice(att, func(i, j int) bool {
			return att[i].AttesterSlot > att[j].AttesterSlot
		})
		res[validator] = att
	}

	return res, nil
}

func (bigtable *Bigtable) getValidatorAttestationHistoryV2(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error) {
	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"validatorsCount": len(validators),
			"startEpoch":      startEpoch,
			"endEpoch":        endEpoch,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	res := make(map[uint64][]*types.ValidatorAttestation, len(validators))
	resMux := &sync.Mutex{}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	attestationsMap := make(map[types.ValidatorIndex]map[types.Slot][]*types.ValidatorAttestation)

	for i := 0; i < len(validators); i += batchSize {
		upperBound := i + batchSize
		if len(validators) < upperBound {
			upperBound = len(validators)
		}
		vals := validators[i:upperBound]

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			ranges := bigtable.getValidatorsEpochRanges(vals, ATTESTATIONS_FAMILY, startEpoch, endEpoch)
			filter := gcp_bigtable.LimitRows(int64(endEpoch-startEpoch+1) * int64(len(vals))) // max is one row per epoch
			err := bigtable.tableValidatorsHistory.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
				keySplit := strings.Split(r.Key(), ":")

				validator, err := bigtable.validatorKeyToIndex(keySplit[1])
				if err != nil {
					logger.Errorf("error parsing validator from row key %v: %v", r.Key(), err)
					return false
				}

				for _, ri := range r[ATTESTATIONS_FAMILY] {
					attesterSlotString := strings.Replace(ri.Column, ATTESTATIONS_FAMILY+":", "", 1)
					attesterSlot, err := strconv.ParseUint(attesterSlotString, 10, 64)
					if err != nil {
						logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
						return false
					}

					status, inclusionSlot := bigtable.getAttestationStatusAndInclusionSlot(ri.Timestamp, attesterSlot)
					resMux.Lock()
					if attestationsMap[types.ValidatorIndex(validator)] == nil {
						attestationsMap[types.ValidatorIndex(validator)] = make(map[types.Slot][]*types.ValidatorAttestation)
					}

					if attestationsMap[types.ValidatorIndex(validator)][types.Slot(attesterSlot)] == nil {
						attestationsMap[types.ValidatorIndex(validator)][types.Slot(attesterSlot)] = make([]*types.ValidatorAttestation, 0)
					}

					attestationsMap[types.ValidatorIndex(validator)][types.Slot(attesterSlot)] = append(attestationsMap[types.ValidatorIndex(validator)][types.Slot(attesterSlot)], &types.ValidatorAttestation{
						InclusionSlot: inclusionSlot,
						Status:        status,
					})
					resMux.Unlock()
				}
				return true
			}, filter)

			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Find all missed and orphaned slots
	slots := []uint64{}
	maxSlot := ((endEpoch + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1
	for slot := startEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= maxSlot; slot++ {
		slots = append(slots, slot)
	}

	var missedSlotsMap map[uint64]bool
	var orphanedSlotsMap map[uint64]bool

	g = new(errgroup.Group)

	g.Go(func() error {
		var err error
		missedSlotsMap, err = GetMissedSlotsMap(slots)
		return err
	})

	g.Go(func() error {
		var err error
		orphanedSlotsMap, err = GetOrphanedSlotsMap(slots)
		return err
	})
	err := g.Wait()
	if err != nil {
		return nil, err
	}

	// Convert the attestationsMap info to the return format
	// Set the delay of the inclusionSlot
	for validator, attestations := range attestationsMap {
		if res[uint64(validator)] == nil {
			res[uint64(validator)] = make([]*types.ValidatorAttestation, 0)
		}
		for attesterSlot, att := range attestations {
			sort.Slice(att, func(i, j int) bool {
				return att[i].InclusionSlot < att[j].InclusionSlot
			})
			currentAttInfo := att[0]
			for _, attInfo := range att {
				if orphanedSlotsMap[attInfo.InclusionSlot] {
					attInfo.Status = 0
				}

				if currentAttInfo.Status != 1 && attInfo.Status == 1 {
					currentAttInfo.Status = attInfo.Status
					currentAttInfo.InclusionSlot = attInfo.InclusionSlot
				}
			}

			missedSlotsCount := uint64(0)
			for slot := uint64(attesterSlot) + 1; slot < currentAttInfo.InclusionSlot; slot++ {
				if missedSlotsMap[slot] || orphanedSlotsMap[slot] {
					missedSlotsCount++
				}
			}
			currentAttInfo.Index = uint64(validator)
			currentAttInfo.Epoch = uint64(attesterSlot) / utils.Config.Chain.ClConfig.SlotsPerEpoch
			currentAttInfo.CommitteeIndex = 0
			currentAttInfo.AttesterSlot = uint64(attesterSlot)
			currentAttInfo.Delay = int64(currentAttInfo.InclusionSlot - uint64(attesterSlot) - missedSlotsCount - 1)

			res[uint64(validator)] = append(res[uint64(validator)], currentAttInfo)
		}
	}

	// Sort the result by attesterSlot desc
	for validator, att := range res {
		sort.Slice(att, func(i, j int) bool {
			return att[i].AttesterSlot > att[j].AttesterSlot
		})
		res[validator] = att
	}

	return res, nil
}

func (bigtable *Bigtable) getValidatorAttestationHistoryV1(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error) {
	valLen := len(validators)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	ranges := bigtable.getSlotRangesForEpochV1(startEpoch, endEpoch)
	res := make(map[uint64][]*types.ValidatorAttestation, len(validators))

	columnFilters := []gcp_bigtable.Filter{}
	if valLen < 1000 {
		columnFilters = make([]gcp_bigtable.Filter, 0, len(validators))
		for _, validator := range validators {
			columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
		}
	}

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
		gcp_bigtable.InterleaveFilters(columnFilters...),
	)

	if len(columnFilters) == 1 { // special case to retrieve data for one validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
			columnFilters[0],
		)
	}
	if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY)
	}

	maxSlot := (endEpoch + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch
	// map with structure attestationsMap[validator][attesterSlot]
	attestationsMap := make(map[uint64]map[uint64][]*types.ValidatorAttestation)

	// Save info for all inclusionSlot for attestations in attestationsMap
	// Set the maxSlot to the highest inclusionSlot
	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		keySplit := strings.Split(r.Key(), ":")

		attesterSlot, err := strconv.ParseUint(keySplit[4], 10, 64)
		if err != nil {
			logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
			return false
		}
		attesterSlot = max_block_number_v1 - attesterSlot
		for _, ri := range r[ATTESTATIONS_FAMILY] {
			inclusionSlot := max_block_number_v1 - uint64(ri.Timestamp)/1000

			status := uint64(1)
			if inclusionSlot == max_block_number_v1 {
				inclusionSlot = 0
				status = 0
			}

			if inclusionSlot > maxSlot {
				maxSlot = inclusionSlot
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, ATTESTATIONS_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			if attestationsMap[validator] == nil {
				attestationsMap[validator] = make(map[uint64][]*types.ValidatorAttestation)
			}

			if attestationsMap[validator][attesterSlot] == nil {
				attestationsMap[validator][attesterSlot] = make([]*types.ValidatorAttestation, 0)
			}

			attestationsMap[validator][attesterSlot] = append(attestationsMap[validator][attesterSlot], &types.ValidatorAttestation{
				InclusionSlot: inclusionSlot,
				Status:        status,
			})
		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	// Find all missed and orphaned slots
	slots := []uint64{}
	for slot := startEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= maxSlot; slot++ {
		slots = append(slots, slot)
	}

	var missedSlotsMap map[uint64]bool
	var orphanedSlotsMap map[uint64]bool

	g := new(errgroup.Group)

	g.Go(func() error {
		missedSlotsMap, err = GetMissedSlotsMap(slots)
		return err
	})

	g.Go(func() error {
		orphanedSlotsMap, err = GetOrphanedSlotsMap(slots)
		return err
	})
	err = g.Wait()
	if err != nil {
		return nil, err
	}

	// Convert the attestationsMap info to the return format
	// Set the delay of the inclusionSlot
	for validator, attestations := range attestationsMap {
		if res[validator] == nil {
			res[validator] = make([]*types.ValidatorAttestation, 0)
		}
		for attesterSlot, att := range attestations {
			currentAttInfo := att[0]
			for _, attInfo := range att {
				if orphanedSlotsMap[attInfo.InclusionSlot] {
					attInfo.Status = 0
				}

				if currentAttInfo.Status != 1 && attInfo.Status == 1 {
					currentAttInfo.Status = attInfo.Status
					currentAttInfo.InclusionSlot = attInfo.InclusionSlot
				}
			}

			missedSlotsCount := uint64(0)
			for slot := attesterSlot + 1; slot < currentAttInfo.InclusionSlot; slot++ {
				if missedSlotsMap[slot] || orphanedSlotsMap[slot] {
					missedSlotsCount++
				}
			}
			currentAttInfo.Index = validator
			currentAttInfo.Epoch = attesterSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch
			currentAttInfo.CommitteeIndex = 0
			currentAttInfo.AttesterSlot = attesterSlot
			currentAttInfo.Delay = int64(currentAttInfo.InclusionSlot - attesterSlot - missedSlotsCount - 1)

			res[validator] = append(res[validator], currentAttInfo)
		}
	}

	// Sort the result by attesterSlot desc
	for validator, att := range res {
		sort.Slice(att, func(i, j int) bool {
			return att[i].AttesterSlot > att[j].AttesterSlot
		})
		res[validator] = att
	}

	return res, nil
}

func (bigtable *Bigtable) GetLastAttestationSlots(validators []uint64) (map[uint64]uint64, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"validatorsCount": len(validators),
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	valLen := len(validators)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	res := make(map[uint64]uint64, len(validators))

	columnFilters := []gcp_bigtable.Filter{}
	if valLen < 1000 {
		columnFilters = make([]gcp_bigtable.Filter, 0, len(validators))
		for _, validator := range validators {
			columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
		}
	}

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
		gcp_bigtable.InterleaveFilters(columnFilters...),
		gcp_bigtable.LatestNFilter(1),
	)

	if len(columnFilters) == 1 { // special case to retrieve data for one validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
			columnFilters[0],
			gcp_bigtable.LatestNFilter(1),
		)
	} else if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
			gcp_bigtable.LatestNFilter(1),
		)
	}

	key := fmt.Sprintf("%s:lastAttestationSlot", bigtable.chainId)

	row, err := bigtable.tableValidators.ReadRow(ctx, key, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	for _, ri := range row[ATTESTATIONS_FAMILY] {
		attestedSlot := uint64(ri.Timestamp) / 1000

		validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, ATTESTATIONS_FAMILY+":"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
		}

		res[validator] = attestedSlot
	}

	return res, nil
}

// Clickhouse port: Done
func (bigtable *Bigtable) GetValidatorMissedAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]bool, error) {
	if utils.Config.ClickHouseEnabled && time.Since(utils.EpochToTime(endEpoch)) > utils.Config.ClickhouseDelay { // fetch data from clickhouse instead
		logger.Infof("fetching validator missed attestation history from clickhouse for validators %v, epochs %v - %v", validators, startEpoch, endEpoch)
		return bigtable.getValidatorMissedAttestationHistoryClickhouse(validators, startEpoch, endEpoch)
	} else if endEpoch < bigtable.v2SchemaCutOffEpoch {
		return bigtable.getValidatorMissedAttestationHistoryV1(validators, startEpoch, endEpoch)
	} else {
		return bigtable.getValidatorMissedAttestationHistoryV2(validators, startEpoch, endEpoch)
	}
}

func (bigtable *Bigtable) getValidatorMissedAttestationHistoryClickhouse(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]bool, error) {
	startEpochTs := utils.EpochToTime(startEpoch)
	endEpochTs := utils.EpochToTime(endEpoch)

	type row struct {
		ValidatorIndex        uint64 `db:"validator_index"`
		Epoch                 uint64 `db:"epoch"`
		Slot                  uint64 `db:"slot"`
		AttestationsScheduled uint64 `db:"attestations_scheduled"`
		AttestationsObserved  uint64 `db:"attestations_observed"`
	}

	rows := []*row{}

	query := `
		SELECT
			validator_dashboard_data_epoch.validator_index,
			validator_dashboard_data_epoch.epoch,
			validator_attestation_assignments_slot.slot,
			validator_dashboard_data_epoch.attestations_scheduled,
			validator_dashboard_data_epoch.attestations_observed
		FROM validator_dashboard_data_epoch FINAL 
		LEFT JOIN validator_attestation_assignments_slot FINAL ON 
			validator_dashboard_data_epoch.validator_index = validator_attestation_assignments_slot.validator_index AND
			validator_dashboard_data_epoch.epoch_timestamp = validator_attestation_assignments_slot.epoch_timestamp AND
			validator_dashboard_data_epoch.epoch = validator_attestation_assignments_slot.epoch
		WHERE 
			validator_dashboard_data_epoch.epoch_timestamp >= ? AND 
			validator_dashboard_data_epoch.epoch_timestamp <= ? AND 
			validator_dashboard_data_epoch.validator_index IN (?) 
		ORDER BY epoch ASC, slot ASC`
	err := ClickhouseReaderDb.Select(&rows, query, startEpochTs, endEpochTs, validators)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	res := make(map[uint64]map[uint64]bool, len(validators))

	for _, r := range rows {
		if r.AttestationsScheduled == 0 {
			continue
		}

		if r.AttestationsScheduled > 0 && r.AttestationsObserved == 0 {
			if res[r.ValidatorIndex] == nil {
				res[r.ValidatorIndex] = make(map[uint64]bool, 0)
			}
			res[r.ValidatorIndex][r.Slot] = true
		}
	}

	return res, nil
}

func (bigtable *Bigtable) getValidatorMissedAttestationHistoryV2(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]bool, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"validatorsCount": len(validators),
			"startEpoch":      startEpoch,
			"endEpoch":        endEpoch,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()

	slots := []uint64{}

	for slot := startEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot < (endEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch; slot++ {
		slots = append(slots, slot)
	}
	orphanedSlotsMap, err := GetOrphanedSlotsMap(slots)
	if err != nil {
		return nil, err
	}

	res := make(map[uint64]map[uint64]bool)
	foundValid := make(map[uint64]map[uint64]bool)

	resMux := &sync.Mutex{}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i := 0; i < len(validators); i += batchSize {
		upperBound := i + batchSize
		if len(validators) < upperBound {
			upperBound = len(validators)
		}
		vals := validators[i:upperBound]

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			ranges := bigtable.getValidatorsEpochRanges(vals, ATTESTATIONS_FAMILY, startEpoch, endEpoch)

			filter := gcp_bigtable.LimitRows(int64(endEpoch-startEpoch+1) * int64(len(vals))) // max is one row per epoch

			err := bigtable.tableValidatorsHistory.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
				keySplit := strings.Split(r.Key(), ":")

				validator, err := bigtable.validatorKeyToIndex(keySplit[1])
				if err != nil {
					logger.Errorf("error parsing validator from row key %v: %v", r.Key(), err)
					return false
				}

				for _, ri := range r[ATTESTATIONS_FAMILY] {
					attesterSlotString := strings.Replace(ri.Column, ATTESTATIONS_FAMILY+":", "", 1)
					attesterSlot, err := strconv.ParseUint(attesterSlotString, 10, 64)
					if err != nil {
						logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
						return false
					}

					status, inclusionSlot := bigtable.getAttestationStatusAndInclusionSlot(ri.Timestamp, attesterSlot)

					resMux.Lock()
					// only if the attestation was not included in another slot we count it as missed
					if (status == 0 || orphanedSlotsMap[inclusionSlot]) && (foundValid[validator] == nil || !foundValid[validator][attesterSlot]) {
						if res[validator] == nil {
							res[validator] = make(map[uint64]bool, 0)
						}
						res[validator][attesterSlot] = true
					} else {
						if res[validator] != nil && res[validator][attesterSlot] {
							delete(res[validator], attesterSlot)
						}
						if foundValid[validator] == nil {
							foundValid[validator] = make(map[uint64]bool, 0)
						}
						foundValid[validator][attesterSlot] = true
					}
					resMux.Unlock()
				}
				return true
			}, filter)

			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) getAttestationStatusAndInclusionSlot(ts gcp_bigtable.Timestamp, attesterSlot uint64) (status uint64, inclusionSlot uint64) {
	if ts.Time().Before(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)) {
		inclusionSlot = MAX_CL_BLOCK_NUMBER - uint64(ts)/1000
		status = uint64(1)
		if inclusionSlot == MAX_CL_BLOCK_NUMBER {
			inclusionSlot = 0
			status = 0
		}
		return status, inclusionSlot
	} else {
		inclusionSlotTs := ts.Time().Unix()
		inclusionSlot = utils.TimeToSlot(uint64(inclusionSlotTs))
		if inclusionSlot == attesterSlot {
			inclusionSlot = 0
			status = 0
		} else {
			status = 1
		}
		return status, inclusionSlot
	}
}

func (bigtable *Bigtable) getValidatorMissedAttestationHistoryV1(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]bool, error) {
	valLen := len(validators)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()

	slots := []uint64{}

	for slot := startEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot < (endEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch; slot++ {
		slots = append(slots, slot)
	}
	orphanedSlotsMap, err := GetOrphanedSlotsMap(slots)
	if err != nil {
		return nil, err
	}

	ranges := bigtable.getSlotRangesForEpochV1(startEpoch, endEpoch)

	res := make(map[uint64]map[uint64]bool)
	foundValid := make(map[uint64]map[uint64]bool)

	columnFilters := []gcp_bigtable.Filter{}
	if valLen < 1000 {
		columnFilters = make([]gcp_bigtable.Filter, 0, len(validators))
		for _, validator := range validators {
			columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
		}
	}

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
		gcp_bigtable.InterleaveFilters(columnFilters...),
	)

	if len(columnFilters) == 1 { // special case to retrieve data for one validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
			columnFilters[0],
		)
	}
	if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY)
	}

	err = bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		keySplit := strings.Split(r.Key(), ":")

		attesterSlot, err := strconv.ParseUint(keySplit[4], 10, 64)
		if err != nil {
			logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
			return false
		}
		attesterSlot = max_block_number_v1 - attesterSlot

		for _, ri := range r[ATTESTATIONS_FAMILY] {
			inclusionSlot := max_block_number_v1 - uint64(ri.Timestamp)/1000

			status := uint64(1)
			if inclusionSlot == max_block_number_v1 {
				status = 0
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, ATTESTATIONS_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			// only if the attestation was not included in another slot we count it as missed
			if (status == 0 || orphanedSlotsMap[inclusionSlot]) && (foundValid[validator] == nil || !foundValid[validator][attesterSlot]) {
				if res[validator] == nil {
					res[validator] = make(map[uint64]bool, 0)
				}
				res[validator][attesterSlot] = true
			} else {
				if res[validator] != nil && res[validator][attesterSlot] {
					delete(res[validator], attesterSlot)
				}
				if foundValid[validator] == nil {
					foundValid[validator] = make(map[uint64]bool, 0)
				}
				foundValid[validator][attesterSlot] = true
			}
		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetValidatorSyncDutiesHistory returns the sync participation status for the given validators ranging from startSlot to endSlot (both inclusive)
//
// The returned map uses the following keys: [validatorIndex][slot]
//
// The function is able to handle both V1 and V2 schema based on the configured v2SchemaCutOffEpoch
// Clickhouse port: Done
func (bigtable *Bigtable) GetValidatorSyncDutiesHistory(validators []uint64, startSlot uint64, endSlot uint64) (map[uint64]map[uint64]*types.ValidatorSyncParticipation, error) {
	if utils.Config.ClickHouseEnabled && time.Since(utils.SlotToTime(endSlot)) > utils.Config.ClickhouseDelay { // fetch data from clickhouse instead
		logger.Infof("fetching validator sync duties history from clickhouse for validators %v, slots %v - %v", validators, startSlot, endSlot)
		return bigtable.getValidatorSyncDutiesHistoryClickhouse(validators, startSlot, endSlot)
	} else if endSlot/utils.Config.Chain.ClConfig.SlotsPerEpoch < bigtable.v2SchemaCutOffEpoch {
		if startSlot/utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 {
			return nil, fmt.Errorf("getValidatorSyncDutiesHistoryV1 is not supported for epoch 0")
		}
		return bigtable.getValidatorSyncDutiesHistoryV1(validators, startSlot, endSlot)
	} else {
		return bigtable.getValidatorSyncDutiesHistoryV2(validators, startSlot, endSlot)
	}
}

func (bigtable *Bigtable) getValidatorSyncDutiesHistoryClickhouse(validators []uint64, startSlot uint64, endSlot uint64) (map[uint64]map[uint64]*types.ValidatorSyncParticipation, error) {
	startEpoch := utils.EpochOfSlot(startSlot)
	endEpoch := utils.EpochOfSlot(endSlot)
	startEpochTs := utils.EpochToTime(startEpoch)
	endEpochTs := utils.EpochToTime(endEpoch)

	type row struct {
		ValidatorIndex uint64 `db:"validator_index"`
		Epoch          uint64 `db:"epoch"`
		Slot           uint64 `db:"slot"`
		Executed       bool   `db:"executed"`
	}

	rows := []*row{}

	query := `
		SELECT
			validator_sync_committee_votes_slot.validator_index,
			validator_sync_committee_votes_slot.epoch,
			validator_sync_committee_votes_slot.slot,
			validator_sync_committee_votes_slot.executed
		FROM validator_sync_committee_votes_slot FINAL
		WHERE
			validator_sync_committee_votes_slot.epoch_timestamp >= ? AND
			validator_sync_committee_votes_slot.epoch_timestamp <= ? AND
			validator_sync_committee_votes_slot.validator_index IN (?) AND
			validator_sync_committee_votes_slot.slot >= ? AND
			validator_sync_committee_votes_slot.slot <= ?
		ORDER BY epoch ASC, slot ASC`

	err := ClickhouseReaderDb.Select(&rows, query, startEpochTs, endEpochTs, validators, startSlot, endSlot)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	res := make(map[uint64]map[uint64]*types.ValidatorSyncParticipation, len(validators))

	for _, r := range rows {
		if res[r.ValidatorIndex] == nil {
			res[r.ValidatorIndex] = make(map[uint64]*types.ValidatorSyncParticipation, 0)
		}
		sp := &types.ValidatorSyncParticipation{
			Slot:   r.Slot,
			Period: 0, //utils.SyncPeriodOfEpoch(utils.EpochOfSlot(r.Slot)), //*sigh*
		}
		if r.Executed {
			sp.Status = 1
		} else {
			sp.Status = 0
		}
		res[r.ValidatorIndex][sp.Slot] = sp
	}

	return res, nil
}

func (bigtable *Bigtable) getValidatorSyncDutiesHistoryV2(validators []uint64, startSlot uint64, endSlot uint64) (map[uint64]map[uint64]*types.ValidatorSyncParticipation, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"validatorsCount": len(validators),
			"startSlot":       startSlot,
			"endSlot":         endSlot,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()

	res := make(map[uint64]map[uint64]*types.ValidatorSyncParticipation, len(validators))
	resMux := &sync.Mutex{}

	filter := gcp_bigtable.LatestNFilter(1)

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i := 0; i < len(validators); i += batchSize {
		i := i
		upperBound := i + batchSize
		if len(validators) < upperBound {
			upperBound = len(validators)
		}
		vals := validators[i:upperBound]

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			ranges := bigtable.getValidatorSlotRanges(vals, SYNC_COMMITTEES_FAMILY, startSlot, endSlot)

			err := bigtable.tableValidatorsHistory.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
				keySplit := strings.Split(r.Key(), ":")

				validator, err := bigtable.validatorKeyToIndex(keySplit[1])
				if err != nil {
					logger.Errorf("error parsing validator from row key %v: %v", r.Key(), err)
					return false
				}
				slot, err := strconv.ParseUint(keySplit[4], 10, 64)
				if err != nil {
					logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
					return false
				}
				slot = MAX_CL_BLOCK_NUMBER - slot

				for _, ri := range r[SYNC_COMMITTEES_FAMILY] {
					status := bigtable.getSyncStatus(ri.Timestamp, slot)

					resMux.Lock()
					if res[validator] == nil {
						res[validator] = make(map[uint64]*types.ValidatorSyncParticipation, 0)
					}

					if len(res[validator]) > 0 && res[validator][slot] != nil {
						res[validator][slot].Status = status
					} else {
						res[validator][slot] = &types.ValidatorSyncParticipation{
							Slot:   slot,
							Status: status,
						}
					}
					resMux.Unlock()
				}
				return true
			}, gcp_bigtable.RowFilter(filter))
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) getSyncStatus(ts gcp_bigtable.Timestamp, slot uint64) uint64 {
	status := uint64(0)
	if ts.Time().Before(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)) {
		// for old schema data read the inclusion slot directly from the timestamp field
		inclusionSlot := MAX_CL_BLOCK_NUMBER - uint64(ts)/1000

		status = uint64(1) // 1: participated
		if inclusionSlot == MAX_CL_BLOCK_NUMBER {
			status = 0 // 0: missed
		}
		return status
	} else {
		// for new schema data calculate the inclusion slot based on the slot timestamp
		slotTs := utils.SlotToTime(slot)
		inclusionSlotTs := ts.Time()

		if slotTs.Equal(inclusionSlotTs) {
			status = 0 // 0: missed
		} else if inclusionSlotTs.Equal(slotTs.Add(time.Second)) { // participated is encoded as slot timestamp + 1 sec
			status = 1 // 1: participated
		} else {
			logger.Errorf("unexpected inclusion slot timestamp, slot %d, slotTs: %v, inclusionSlotTs: %v", slot, slotTs, inclusionSlotTs)
		}
		return status
	}
}

func (bigtable *Bigtable) getValidatorSyncDutiesHistoryV1(validators []uint64, startSlot uint64, endSlot uint64) (map[uint64]map[uint64]*types.ValidatorSyncParticipation, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	ranges := bigtable.getSlotRangesV1(startSlot, endSlot)
	res := make(map[uint64]map[uint64]*types.ValidatorSyncParticipation, len(validators))

	columnFilters := make([]gcp_bigtable.Filter, 0, len(validators))
	for _, validator := range validators {
		columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
	}

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(SYNC_COMMITTEES_FAMILY),
		gcp_bigtable.InterleaveFilters(columnFilters...),
		gcp_bigtable.LatestNFilter(1),
	)

	if len(columnFilters) == 1 { // special case to retrieve data for one validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(SYNC_COMMITTEES_FAMILY),
			columnFilters[0],
			gcp_bigtable.LatestNFilter(1),
		)
	}
	if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(SYNC_COMMITTEES_FAMILY),
			gcp_bigtable.LatestNFilter(1),
		)
	}

	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		for _, ri := range r[SYNC_COMMITTEES_FAMILY] {
			keySplit := strings.Split(r.Key(), ":")

			slot, err := strconv.ParseUint(keySplit[4], 10, 64)
			if err != nil {
				logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
				return false
			}
			slot = max_block_number_v1 - slot
			inclusionSlot := max_block_number_v1 - uint64(ri.Timestamp)/1000

			status := uint64(1) // 1: participated
			if inclusionSlot == max_block_number_v1 {
				inclusionSlot = 0
				status = 0 // 0: missed
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, SYNC_COMMITTEES_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			if res[validator] == nil {
				res[validator] = make(map[uint64]*types.ValidatorSyncParticipation)
			}

			if len(res[validator]) > 0 && res[validator][slot] != nil {
				res[validator][slot].Status = status
			} else {
				res[validator][slot] = &types.ValidatorSyncParticipation{
					Slot:   slot,
					Status: status,
				}
			}
		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorMissedAttestationsCount(validators []uint64, firstEpoch uint64, lastEpoch uint64) (map[uint64]*types.ValidatorMissedAttestationsStatistic, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"validatorsCount": len(validators),
			"startEpoch":      firstEpoch,
			"endEpoch":        lastEpoch,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	if firstEpoch > lastEpoch {
		return nil, fmt.Errorf("GetValidatorMissedAttestationsCount received an invalid firstEpoch (%d) and lastEpoch (%d) combination", firstEpoch, lastEpoch)
	}

	res := make(map[uint64]*types.ValidatorMissedAttestationsStatistic)

	data, err := bigtable.GetValidatorMissedAttestationHistory(validators, firstEpoch, lastEpoch)

	if err != nil {
		return nil, err
	}

	// logger.Infof("retrieved missed attestation history for epochs %v - %v", firstEpoch, lastEpoch)

	for validator, attestations := range data {
		if len(attestations) == 0 {
			continue
		}
		res[validator] = &types.ValidatorMissedAttestationsStatistic{
			Index:              validator,
			MissedAttestations: uint64(len(attestations)),
		}
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorSyncDutiesStatistics(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*types.ValidatorSyncDutiesStatistic, error) {
	data, err := bigtable.GetValidatorSyncDutiesHistory(validators, startEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, ((endEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch)-1)

	if err != nil {
		return nil, err
	}

	slotsMap := make(map[uint64]bool)
	for _, duties := range data {
		for _, duty := range duties {
			slotsMap[duty.Slot] = true
		}
	}
	slots := []uint64{}
	for slot := range slotsMap {
		slots = append(slots, slot)
	}

	orphanedSlots, err := GetOrphanedSlots(slots)
	if err != nil {
		return nil, err
	}

	orphanedSlotsMap := make(map[uint64]bool)
	for _, slot := range orphanedSlots {
		orphanedSlotsMap[slot] = true
	}

	res := make(map[uint64]*types.ValidatorSyncDutiesStatistic)

	for validator, duties := range data {
		if res[validator] == nil && len(duties) > 0 {
			res[validator] = &types.ValidatorSyncDutiesStatistic{
				Index: validator,
			}
		}

		for _, duty := range duties {
			if orphanedSlotsMap[duty.Slot] {
				res[validator].OrphanedSync++
			} else if duty.Status == 0 {
				res[validator].MissedSync++
			} else {
				res[validator].ParticipatedSync++
			}
		}
	}

	return res, nil
}

// returns the validator attestation effectiveness in %
func (bigtable *Bigtable) GetValidatorEffectiveness(validators []uint64, epoch uint64) ([]*types.ValidatorEffectiveness, error) {
	end := epoch
	start := uint64(0)
	lookback := uint64(99)
	if end > lookback {
		start = end - lookback
	}
	data, err := bigtable.GetValidatorAttestationHistory(validators, start, end)

	if err != nil {
		return nil, err
	}

	res := make([]*types.ValidatorEffectiveness, 0, len(validators))
	type readings struct {
		Count uint64
		Sum   float64
	}

	aggEffectiveness := make(map[uint64]*readings)

	for validator, history := range data {
		for _, attestation := range history {
			if aggEffectiveness[validator] == nil {
				aggEffectiveness[validator] = &readings{}
			}
			if attestation.InclusionSlot > 0 {
				// logger.Infof("adding %v for epoch %v %.2f%%", attestation.InclusionSlot, attestation.AttesterSlot, 1.0/float64(attestation.InclusionSlot-attestation.AttesterSlot)*100)
				aggEffectiveness[validator].Sum += 1.0 / float64(attestation.InclusionSlot-attestation.AttesterSlot)
				aggEffectiveness[validator].Count++
			} else {
				aggEffectiveness[validator].Sum += 0 // missed attestations get a penalty of 32 slots
				aggEffectiveness[validator].Count++
			}
		}
	}
	for validator, reading := range aggEffectiveness {
		res = append(res, &types.ValidatorEffectiveness{
			Validatorindex:        validator,
			AttestationEfficiency: float64(reading.Sum) / float64(reading.Count) * 100,
		})
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorBalanceStatistics(validators []uint64, startEpoch, endEpoch uint64) (map[uint64]*types.ValidatorBalanceStatistic, error) {
	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"validatorsCount": len(validators),
			"startEpoch":      startEpoch,
			"endEpoch":        endEpoch,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	type ResultContainer struct {
		mu  sync.Mutex
		res map[uint64]*types.ValidatorBalanceStatistic
	}
	resultContainer := ResultContainer{}
	resultContainer.res = make(map[uint64]*types.ValidatorBalanceStatistic)

	// g, gCtx := errgroup.WithContext(ctx)
	batchSize := 10000
	// g.SetLimit(1)
	for i := 0; i < len(validators); i += batchSize {
		upperBound := i + batchSize
		if len(validators) < upperBound {
			upperBound = len(validators)
		}
		vals := validators[i:upperBound]

		// logrus.Infof("retrieving validator balance stats for validators %v - %v", vals[0], vals[len(vals)-1])

		res, err := bigtable.GetValidatorBalanceHistory(vals, startEpoch, endEpoch)
		if err != nil {
			return nil, err
		}
		resultContainer.mu.Lock()
		for validator, balances := range res {
			for _, balance := range balances {
				if resultContainer.res[validator] == nil {
					resultContainer.res[validator] = &types.ValidatorBalanceStatistic{
						Index:                 validator,
						MinEffectiveBalance:   balance.EffectiveBalance,
						MaxEffectiveBalance:   0,
						MinBalance:            balance.Balance,
						MaxBalance:            0,
						StartEffectiveBalance: 0,
						EndEffectiveBalance:   0,
						StartBalance:          0,
						EndBalance:            0,
					}
				}

				if balance.Epoch == startEpoch {
					resultContainer.res[validator].StartBalance = balance.Balance
					resultContainer.res[validator].StartEffectiveBalance = balance.EffectiveBalance
				}

				if balance.Epoch == endEpoch {
					resultContainer.res[validator].EndBalance = balance.Balance
					resultContainer.res[validator].EndEffectiveBalance = balance.EffectiveBalance
				}

				if balance.Balance > resultContainer.res[validator].MaxBalance {
					resultContainer.res[validator].MaxBalance = balance.Balance
				}
				if balance.Balance < resultContainer.res[validator].MinBalance {
					resultContainer.res[validator].MinBalance = balance.Balance
				}

				if balance.EffectiveBalance > resultContainer.res[validator].MaxEffectiveBalance {
					resultContainer.res[validator].MaxEffectiveBalance = balance.EffectiveBalance
				}
				if balance.EffectiveBalance < resultContainer.res[validator].MinEffectiveBalance {
					resultContainer.res[validator].MinEffectiveBalance = balance.EffectiveBalance
				}
			}
		}

		resultContainer.mu.Unlock()

	}

	return resultContainer.res, nil
}

func (bigtable *Bigtable) SaveValidatorIncomeDetails(epoch uint64, rewards map[uint64]*itypes.ValidatorEpochIncome) error {
	start := time.Now()
	ts := gcp_bigtable.Time(utils.EpochToTime(epoch))

	total := &itypes.ValidatorEpochIncome{}

	muts := types.NewBulkMutations(len(rewards))

	for i, rewardDetails := range rewards {
		data, err := proto.Marshal(rewardDetails)

		if err != nil {
			return err
		}

		mut := &gcp_bigtable.Mutation{}
		mut.Set(INCOME_DETAILS_COLUMN_FAMILY, "i", ts, data)
		key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(i), INCOME_DETAILS_COLUMN_FAMILY, bigtable.reversedPaddedEpoch(epoch))

		muts.Add(key, mut)

		total.AttestationHeadReward += rewardDetails.AttestationHeadReward
		total.AttestationSourceReward += rewardDetails.AttestationSourceReward
		total.AttestationSourcePenalty += rewardDetails.AttestationSourcePenalty
		total.AttestationTargetReward += rewardDetails.AttestationTargetReward
		total.AttestationTargetPenalty += rewardDetails.AttestationTargetPenalty
		total.FinalityDelayPenalty += rewardDetails.FinalityDelayPenalty
		total.ProposerSlashingInclusionReward += rewardDetails.ProposerSlashingInclusionReward
		total.ProposerAttestationInclusionReward += rewardDetails.ProposerAttestationInclusionReward
		total.ProposerSyncInclusionReward += rewardDetails.ProposerSyncInclusionReward
		total.SyncCommitteeReward += rewardDetails.SyncCommitteeReward
		total.SyncCommitteePenalty += rewardDetails.SyncCommitteePenalty
		total.SlashingReward += rewardDetails.SlashingReward
		total.SlashingPenalty += rewardDetails.SlashingPenalty
		total.TxFeeRewardWei = utils.AddBigInts(total.TxFeeRewardWei, rewardDetails.TxFeeRewardWei)
	}

	sum, err := proto.Marshal(total)
	if err != nil {
		return err
	}

	mut := &gcp_bigtable.Mutation{}
	mut.Set(STATS_COLUMN_FAMILY, SUM_COLUMN, ts, sum)

	muts.Add(fmt.Sprintf("%s:%s:%s", bigtable.chainId, SUM_COLUMN, bigtable.reversedPaddedEpoch(epoch)), mut)

	err = bigtable.WriteBulk(muts, bigtable.tableValidatorsHistory, MAX_BATCH_MUTATIONS)

	if err != nil {
		return err
	}

	logger.Infof("exported validator income details for epoch %v to bigtable in %v", epoch, time.Since(start))
	return nil
}

// GetValidatorIncomeDetailsHistory returns the validator income details
// startEpoch & endEpoch are inclusive
// return object is a map of validator_index -> epoch -> incomeDetails
// Clickhouse port: Done
func (bigtable *Bigtable) GetValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*itypes.ValidatorEpochIncome, error) {
	endEpochTs := utils.EpochToTime(endEpoch)
	if utils.Config.ClickHouseEnabled && time.Since(endEpochTs) > utils.Config.ClickhouseDelay { // fetch data from clickhouse instead
		logger.Infof("fetching validator income details from clickhouse for validators %v, epochs %v - %v", validators, startEpoch, endEpoch)
		return bigtable.getValidatorIncomeDetailsHistoryClickhouse(validators, startEpoch, endEpoch)
	} else if endEpoch < bigtable.v2SchemaCutOffEpoch {
		return bigtable.getValidatorIncomeDetailsHistoryV1(validators, startEpoch, endEpoch)
	} else {
		return bigtable.getValidatorIncomeDetailsHistoryV2(validators, startEpoch, endEpoch)
	}
}

func (bigtable *Bigtable) getValidatorIncomeDetailsHistoryClickhouse(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*itypes.ValidatorEpochIncome, error) {
	startEpochTs := utils.EpochToTime(startEpoch)
	endEpochTs := utils.EpochToTime(endEpoch)

	type row struct {
		ValidatorIndex               uint64 `db:"validator_index"`
		Epoch                        uint64 `db:"epoch"`
		AttestationsSourceReward     int64  `db:"attestations_source_reward"`
		AttestationsTargetReward     int64  `db:"attestations_target_reward"`
		AttestationsHeadReward       int64  `db:"attestations_head_reward"`
		AttestationsInactivityReward int64  `db:"attestations_inactivity_reward"`
		AttestationsInclusionReward  int64  `db:"attestations_inclusion_reward"`
		SyncRewards                  int64  `db:"sync_reward_rewards_only"`
		SyncPenalties                int64  `db:"sync_reward_penalties_only"`
		SyncScheduled                int64  `db:"sync_scheduled"`
		SyncExecuted                 int64  `db:"sync_executed"`
		BlocksClAttestationsReward   int64  `db:"blocks_cl_attestations_reward"`
		BlocksClSyncAggregateReward  int64  `db:"blocks_cl_sync_aggregate_reward"`
		BlocksClSlasherReward        int64  `db:"blocks_cl_slasher_reward"`
		BlocksScheduled              int64  `db:"blocks_scheduled"`
		BlocksProposed               int64  `db:"blocks_proposed"`
	}
	rows := []*row{}

	query := `
			SELECT 
				validator_index, 
				epoch, 
				attestations_source_reward, 
				attestations_target_reward, 
				attestations_head_reward, 
				attestations_inactivity_reward, 
				attestations_inclusion_reward, 
				sync_reward_rewards_only, 
				sync_reward_penalties_only,
				sync_scheduled,
				sync_executed,
				blocks_cl_attestations_reward, 
				blocks_cl_sync_aggregate_reward, 
				blocks_cl_slasher_reward, 
				blocks_scheduled, 
				blocks_proposed 
			FROM validator_dashboard_data_epoch FINAL WHERE epoch_timestamp >= ? AND epoch_timestamp <= ? AND validator_index IN (?) ORDER BY epoch ASC`

	err := ClickhouseReaderDb.Select(&rows, query, startEpochTs, endEpochTs, validators)
	if err != nil {
		return nil, err
	}

	res := make(map[uint64]map[uint64]*itypes.ValidatorEpochIncome, len(validators))
	proposalEpochs := make(map[uint64]bool)
	proposalValidators := make(map[uint64]bool)
	for _, r := range rows {
		if r.BlocksScheduled == 0 && r.SyncScheduled == 0 && r.AttestationsHeadReward == 0 && r.AttestationsSourceReward == 0 && r.AttestationsTargetReward == 0 && r.AttestationsInactivityReward == 0 && r.BlocksClAttestationsReward == 0 && r.BlocksClSlasherReward == 0 && r.BlocksClSyncAggregateReward == 0 && r.SyncRewards == 0 {
			continue
		}
		if res[r.ValidatorIndex] == nil {
			res[r.ValidatorIndex] = make(map[uint64]*itypes.ValidatorEpochIncome)
		}
		income := &itypes.ValidatorEpochIncome{}
		income.AttestationHeadReward = uint64(r.AttestationsHeadReward)
		if r.AttestationsSourceReward > 0 {
			income.AttestationSourceReward = uint64(r.AttestationsSourceReward)
		} else {
			income.AttestationSourcePenalty = uint64(r.AttestationsSourceReward * -1)
		}
		if r.AttestationsTargetReward > 0 {
			income.AttestationTargetReward = uint64(r.AttestationsTargetReward)
		} else {
			income.AttestationTargetPenalty = uint64(r.AttestationsTargetReward * -1)
		}
		income.FinalityDelayPenalty = uint64(r.AttestationsInactivityReward * -1)
		income.ProposerAttestationInclusionReward = uint64(r.BlocksClAttestationsReward)
		income.ProposerSlashingInclusionReward = uint64(r.BlocksClSlasherReward)
		income.ProposerSyncInclusionReward = uint64(r.BlocksClSyncAggregateReward)
		income.SyncCommitteeReward = uint64(r.SyncRewards)
		income.SyncCommitteePenalty = uint64(r.SyncPenalties * -1)
		income.ProposalsMissed = uint64(r.BlocksScheduled - r.BlocksProposed)

		if r.BlocksProposed > 0 {
			proposalEpochs[r.Epoch] = true
			proposalValidators[r.ValidatorIndex] = true
		}
		res[r.ValidatorIndex][r.Epoch] = income
	}

	if len(proposalEpochs) > 0 {
		// get proposal tx fee reward data
		type row struct {
			Proposer uint64 `db:"proposer"`
			Epoch    uint64 `db:"epoch"`
			TxFee    string `db:"tx_fee"`
		}
		rows := []*row{}
		query := `
			SELECT 
				proposer, 
				epoch, 
				sum(fee_recipient_reward) * 1e18 as tx_fee 
			FROM blocks 
			LEFT JOIN execution_payloads ON blocks.exec_block_hash = execution_payloads.block_hash 
			WHERE blocks.epoch = ANY($1) AND blocks.proposer = ANY($2) AND blocks.status = '1' AND fee_recipient_reward IS NOT NULL GROUP BY proposer, epoch`

		err := ReaderDb.Select(&rows, query, pq.Array(maps.Keys(proposalEpochs)), pq.Array(maps.Keys(proposalValidators)))
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		for _, r := range rows {
			if res[r.Proposer] == nil {
				res[r.Proposer] = make(map[uint64]*itypes.ValidatorEpochIncome)
			}
			if res[r.Proposer][r.Epoch] == nil {
				res[r.Proposer][r.Epoch] = &itypes.ValidatorEpochIncome{}
			}
			reward, ok := new(big.Float).SetString(r.TxFee)
			rewardInt, _ := reward.Int(nil)
			if !ok {
				logger.Errorf("error parsing tx fee reward for validator %v epoch %v: %v", r.Proposer, r.Epoch, r.TxFee)
				return nil, fmt.Errorf("error parsing tx fee reward for validator %v epoch %v: %v", r.Proposer, r.Epoch, r.TxFee)
			}
			res[r.Proposer][r.Epoch].TxFeeRewardWei = rewardInt.Bytes()
		}
	}
	return res, nil
}

func (bigtable *Bigtable) getValidatorIncomeDetailsHistoryV2(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*itypes.ValidatorEpochIncome, error) {

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"validatorsCount": len(validators),
			"startEpoch":      startEpoch,
			"endEpoch":        endEpoch,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	if startEpoch > endEpoch {
		startEpoch = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	res := make(map[uint64]map[uint64]*itypes.ValidatorEpochIncome, len(validators))
	resMux := &sync.Mutex{}

	filter := gcp_bigtable.LatestNFilter(1)

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i := 0; i < len(validators); i += batchSize {

		upperBound := i + batchSize
		if len(validators) < upperBound {
			upperBound = len(validators)
		}
		vals := validators[i:upperBound]

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			ranges := bigtable.getValidatorsEpochRanges(vals, INCOME_DETAILS_COLUMN_FAMILY, startEpoch, endEpoch)
			err := bigtable.tableValidatorsHistory.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
				keySplit := strings.Split(r.Key(), ":")

				validator, err := bigtable.validatorKeyToIndex(keySplit[1])
				if err != nil {
					logger.Errorf("error parsing validator from row key %v: %v", r.Key(), err)
					return false
				}

				epoch, err := strconv.ParseUint(keySplit[3], 10, 64)
				if err != nil {
					logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
					return false
				}

				for _, ri := range r[INCOME_DETAILS_COLUMN_FAMILY] {
					incomeDetails := &itypes.ValidatorEpochIncome{}
					err = proto.Unmarshal(ri.Value, incomeDetails)
					if err != nil {
						logger.Errorf("error decoding validator income data for row %v: %v", r.Key(), err)
						return false
					}

					resMux.Lock()
					if res[validator] == nil {
						res[validator] = make(map[uint64]*itypes.ValidatorEpochIncome)
					}

					res[validator][MAX_EPOCH-epoch] = incomeDetails
					resMux.Unlock()
				}
				return true
			}, gcp_bigtable.RowFilter(filter))

			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) getValidatorIncomeDetailsHistoryV1(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*itypes.ValidatorEpochIncome, error) {
	if startEpoch > endEpoch {
		startEpoch = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	ranges := bigtable.getEpochRangesV1(startEpoch, endEpoch)
	res := make(map[uint64]map[uint64]*itypes.ValidatorEpochIncome, len(validators))

	valLen := len(validators)

	// read entire row if you require more than 1000 validators
	var columnFilters []gcp_bigtable.Filter
	if valLen < 1000 {
		columnFilters = make([]gcp_bigtable.Filter, 0, valLen)
		for _, validator := range validators {
			columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
		}
	}

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(INCOME_DETAILS_COLUMN_FAMILY),
		gcp_bigtable.InterleaveFilters(columnFilters...),
		gcp_bigtable.LatestNFilter(1),
	)

	if len(columnFilters) == 1 { // special case to retrieve data for one validator
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(INCOME_DETAILS_COLUMN_FAMILY),
			columnFilters[0],
			gcp_bigtable.LatestNFilter(1),
		)
	}
	if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(INCOME_DETAILS_COLUMN_FAMILY),
			gcp_bigtable.LatestNFilter(1),
		)
	}

	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		keySplit := strings.Split(r.Key(), ":")

		epoch, err := strconv.ParseUint(keySplit[3], 10, 64)
		if err != nil {
			logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
			return false
		}

		// logger.Info(max_epoch - epoch)
		for _, ri := range r[INCOME_DETAILS_COLUMN_FAMILY] {
			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, INCOME_DETAILS_COLUMN_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			incomeDetails := &itypes.ValidatorEpochIncome{}
			err = proto.Unmarshal(ri.Value, incomeDetails)
			if err != nil {
				logger.Errorf("error decoding validator income data for row %v: %v", r.Key(), err)
				return false
			}

			if res[validator] == nil {
				res[validator] = make(map[uint64]*itypes.ValidatorEpochIncome)
			}

			res[validator][max_epoch_v1-epoch] = incomeDetails
		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetAggregatedValidatorIncomeDetailsHistory returns aggregated validator income details
// startEpoch & endEpoch are inclusive
func (bigtable *Bigtable) GetAggregatedValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*itypes.ValidatorEpochIncome, error) {
	if startEpoch > endEpoch {
		startEpoch = 0
	}

	type ResultContainer struct {
		mu  sync.Mutex
		res map[uint64]*itypes.ValidatorEpochIncome
	}
	resultContainer := ResultContainer{}
	resultContainer.res = make(map[uint64]*itypes.ValidatorEpochIncome, len(validators))

	batchSize := 10000
	for i := 0; i < len(validators); i += batchSize {

		upperBound := i + batchSize
		if len(validators) < upperBound {
			upperBound = len(validators)
		}
		vals := validators[i:upperBound]

		logrus.Infof("retrieving validator income stats for validators %v - %v", vals[0], vals[len(vals)-1])

		res, err := bigtable.GetValidatorIncomeDetailsHistory(vals, startEpoch, endEpoch)

		if err != nil {
			return nil, err
		}
		resultContainer.mu.Lock()
		for validator, epochs := range res {
			for _, rewardDetails := range epochs {

				if resultContainer.res[validator] == nil {
					resultContainer.res[validator] = &itypes.ValidatorEpochIncome{}
				}

				resultContainer.res[validator].AttestationHeadReward += rewardDetails.AttestationHeadReward
				resultContainer.res[validator].AttestationSourceReward += rewardDetails.AttestationSourceReward
				resultContainer.res[validator].AttestationSourcePenalty += rewardDetails.AttestationSourcePenalty
				resultContainer.res[validator].AttestationTargetReward += rewardDetails.AttestationTargetReward
				resultContainer.res[validator].AttestationTargetPenalty += rewardDetails.AttestationTargetPenalty
				resultContainer.res[validator].FinalityDelayPenalty += rewardDetails.FinalityDelayPenalty
				resultContainer.res[validator].ProposerSlashingInclusionReward += rewardDetails.ProposerSlashingInclusionReward
				resultContainer.res[validator].ProposerAttestationInclusionReward += rewardDetails.ProposerAttestationInclusionReward
				resultContainer.res[validator].ProposerSyncInclusionReward += rewardDetails.ProposerSyncInclusionReward
				resultContainer.res[validator].SyncCommitteeReward += rewardDetails.SyncCommitteeReward
				resultContainer.res[validator].SyncCommitteePenalty += rewardDetails.SyncCommitteePenalty
				resultContainer.res[validator].SlashingReward += rewardDetails.SlashingReward
				resultContainer.res[validator].SlashingPenalty += rewardDetails.SlashingPenalty
				resultContainer.res[validator].TxFeeRewardWei = utils.AddBigInts(resultContainer.res[validator].TxFeeRewardWei, rewardDetails.TxFeeRewardWei)
			}
		}
		resultContainer.mu.Unlock()
	}

	return resultContainer.res, nil
}

// GetTotalValidatorIncomeDetailsHistory returns the total validator income for a given range of epochs
// It is considerably faster than fetching the individual income for each validator and aggregating it
// startEpoch & endEpoch are inclusive
// Clickhouse port: Not required, uses only data for the last 10 epochs after head
func (bigtable *Bigtable) GetTotalValidatorIncomeDetailsHistory(startEpoch uint64, endEpoch uint64) (map[uint64]*itypes.ValidatorEpochIncome, error) {
	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"startEpoch": startEpoch,
			"endEpoch":   endEpoch,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	if startEpoch > endEpoch {
		startEpoch = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	res := make(map[uint64]*itypes.ValidatorEpochIncome, endEpoch-startEpoch+1)

	filter := gcp_bigtable.LimitRows(int64(endEpoch - startEpoch + 1))

	rowRange := bigtable.getTotalIncomeEpochRanges(startEpoch, endEpoch)
	err := bigtable.tableValidatorsHistory.ReadRows(ctx, rowRange, func(r gcp_bigtable.Row) bool {
		keySplit := strings.Split(r.Key(), ":")

		epoch, err := strconv.ParseUint(keySplit[2], 10, 64)
		if err != nil {
			logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
			return false
		}

		for _, ri := range r[STATS_COLUMN_FAMILY] {
			incomeDetails := &itypes.ValidatorEpochIncome{}
			err = proto.Unmarshal(ri.Value, incomeDetails)
			if err != nil {
				logger.Errorf("error decoding validator income data for row %v: %v", r.Key(), err)
				return false
			}

			res[MAX_EPOCH-epoch] = incomeDetails
		}
		return true
	}, filter)

	if err != nil {
		return nil, err
	}
	return res, nil
}

// Deletes all block data from bigtable
func (bigtable *Bigtable) DeleteEpoch(epoch uint64) error {
	// TOTO: Implement
	return fmt.Errorf("NOT IMPLEMENTED")
}

func (bigtable *Bigtable) getValidatorsEpochRanges(validatorIndices []uint64, prefix string, startEpoch uint64, endEpoch uint64) gcp_bigtable.RowRangeList {
	if endEpoch > math.MaxInt64 {
		endEpoch = 0
	}
	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	ranges := make(gcp_bigtable.RowRangeList, 0, int((endEpoch-startEpoch+1))*len(validatorIndices))

	for _, validatorIndex := range validatorIndices {
		validatorKey := bigtable.validatorIndexToKey(validatorIndex)

		// epochs are sorted descending, so start with the largest epoch and end with the smallest
		// add \x00 to make the range inclusive
		rangeEnd := fmt.Sprintf("%s:%s:%s:%s%s", bigtable.chainId, validatorKey, prefix, bigtable.reversedPaddedEpoch(startEpoch), "\x00")
		rangeStart := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, validatorKey, prefix, bigtable.reversedPaddedEpoch(endEpoch))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
	}
	return ranges
}

func (bigtable *Bigtable) getTotalIncomeEpochRanges(startEpoch uint64, endEpoch uint64) gcp_bigtable.RowRange {
	if endEpoch > math.MaxInt64 {
		endEpoch = 0
	}
	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	rangeEnd := fmt.Sprintf("%s:%s:%s%s", bigtable.chainId, SUM_COLUMN, bigtable.reversedPaddedEpoch(startEpoch), "\x00")
	rangeStart := fmt.Sprintf("%s:%s:%s", bigtable.chainId, SUM_COLUMN, bigtable.reversedPaddedEpoch(endEpoch))

	return gcp_bigtable.NewRange(rangeStart, rangeEnd)
}

func (bigtable *Bigtable) getValidatorSlotRanges(validatorIndices []uint64, prefix string, startSlot uint64, endSlot uint64) gcp_bigtable.RowRangeList {
	if endSlot > math.MaxInt64 {
		endSlot = 0
	}
	if endSlot < startSlot { // handle overflows
		startSlot = 0
	}

	startEpoch := utils.EpochOfSlot(startSlot)
	endEpoch := utils.EpochOfSlot(endSlot)

	ranges := make(gcp_bigtable.RowRangeList, 0, len(validatorIndices))

	for _, validatorIndex := range validatorIndices {
		validatorKey := bigtable.validatorIndexToKey(validatorIndex)

		rangeEnd := fmt.Sprintf("%s:%s:%s:%s:%s%s", bigtable.chainId, validatorKey, prefix, bigtable.reversedPaddedEpoch(startEpoch), bigtable.reversedPaddedSlot(startSlot), "\x00")
		rangeStart := fmt.Sprintf("%s:%s:%s:%s:%s", bigtable.chainId, validatorKey, prefix, bigtable.reversedPaddedEpoch(endEpoch), bigtable.reversedPaddedSlot(endSlot))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
	}
	return ranges
}

func (bigtable *Bigtable) validatorIndexToKey(index uint64) string {
	return utils.ReverseString(fmt.Sprintf("%d", index))
}

func (bigtable *Bigtable) validatorKeyToIndex(key string) (uint64, error) {
	key = utils.ReverseString(key)
	indexKey, err := strconv.ParseUint(key, 10, 64)

	if err != nil {
		return 0, err
	}
	return indexKey, nil
}

func GetCurrentDayClIncome(validator_indices []uint64) (map[uint64]int64, error) {
	dayIncome := make(map[uint64]int64)
	lastDay, err := GetLastExportedStatisticDay()
	if err != nil {
		if err == ErrNoStats {
			return dayIncome, nil
		}
		return dayIncome, err
	}

	currentDay := uint64(lastDay + 1)
	startEpoch := currentDay * utils.EpochsPerDay()
	endEpoch := startEpoch + utils.EpochsPerDay() - 1
	income, err := BigtableClient.GetValidatorIncomeDetailsHistory(validator_indices, startEpoch, endEpoch)
	if err != nil {
		return dayIncome, err
	}

	// agregate all epoch income data to total day income for each validator
	for validatorIndex, validatorIncome := range income {
		if len(validatorIncome) == 0 {
			continue
		}
		for _, validatorEpochIncome := range validatorIncome {
			dayIncome[validatorIndex] += validatorEpochIncome.TotalClRewards()
		}
	}

	return dayIncome, nil
}

func (bigtable *Bigtable) reversePaddedUserID(userID uint64) string {
	return fmt.Sprintf("%09d", ^uint64(0)-userID)
}

func (bigtable *Bigtable) reversedPaddedEpoch(epoch uint64) string {
	return fmt.Sprintf("%09d", MAX_EPOCH-epoch)
}

func (bigtable *Bigtable) reversedPaddedSlot(slot uint64) string {
	return fmt.Sprintf("%09d", MAX_CL_BLOCK_NUMBER-slot)
}

func (bigtable *Bigtable) MigrateIncomeDataV1V2Schema(epoch uint64) error {
	type validatorEpochData struct {
		ValidatorIndex uint64
		IncomeDetails  *itypes.ValidatorEpochIncome
	}

	epochData := make(map[uint64]*validatorEpochData)
	filter := gcp_bigtable.ChainFilters(gcp_bigtable.FamilyFilter(INCOME_DETAILS_COLUMN_FAMILY), gcp_bigtable.LatestNFilter(1))
	ctx := context.Background()

	prefixEpochRange := gcp_bigtable.PrefixRange(fmt.Sprintf("%s:e:b:%s", bigtable.chainId, fmt.Sprintf("%09d", (MAX_EPOCH)-epoch)))

	err := bigtable.tableBeaconchain.ReadRows(ctx, prefixEpochRange, func(r gcp_bigtable.Row) bool {
		// logger.Infof("processing row %v", r.Key())

		keySplit := strings.Split(r.Key(), ":")

		rowKeyEpoch, err := strconv.ParseUint(keySplit[3], 10, 64)
		if err != nil {
			logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
			return false
		}

		rowKeyEpoch = MAX_EPOCH - rowKeyEpoch

		if epoch != rowKeyEpoch {
			logger.Errorf("retrieved different epoch than requested, requested: %d, retrieved: %d", epoch, rowKeyEpoch)
		}

		// logger.Infof("epoch is %d", rowKeyEpoch)

		for columnFamily, readItems := range r {

			for _, ri := range readItems {

				if ri.Column == "stats:sum" { // skip migrating the total epoch income stats
					continue
				}

				validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, columnFamily+":"), 10, 64)
				if err != nil {
					logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
					return false
				}

				// logger.Infof("retrieved field %s from column family %s for validator %d", ri.Column, columnFamily, validator)
				if columnFamily == INCOME_DETAILS_COLUMN_FAMILY {
					if epochData[validator] == nil {
						epochData[validator] = &validatorEpochData{
							ValidatorIndex: validator,
						}
					}
					// logger.Infof("processing income details data for validator %d", validator)
					incomeDetails := &itypes.ValidatorEpochIncome{}
					err = proto.Unmarshal(ri.Value, incomeDetails)
					if err != nil {
						logger.Errorf("error decoding validator income data for row %v: %v", r.Key(), err)
						return false
					}

					epochData[validator].IncomeDetails = incomeDetails
				} else {
					logger.Errorf("retrieved unexpected column family %s", columnFamily)
				}
			}
		}

		return true
	}, gcp_bigtable.RowFilter(filter))

	if err != nil {
		return err
	}

	incomeData := make(map[uint64]*itypes.ValidatorEpochIncome)
	for _, validator := range epochData {
		if validator.IncomeDetails == nil {
			continue
		}
		incomeData[validator.ValidatorIndex] = validator.IncomeDetails
	}

	err = bigtable.SaveValidatorIncomeDetails(epoch, incomeData)
	if err != nil {
		return err
	}

	return nil
}

func (bigtable *Bigtable) getSlotRangesForEpochV1(startEpoch uint64, endEpoch uint64) gcp_bigtable.RowRangeList {

	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	ranges := gcp_bigtable.RowRangeList{}
	if startEpoch == 0 { // special case when the 0 epoch is included
		rangeEnd := fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpochV1(0), ":")
		rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpochV1(0))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))

		// epochs are sorted descending, so start with the larges epoch and end with the smallest
		// add ':', a character lexicographically after digits, to make the range inclusive
		if startEpoch < endEpoch {
			rangeEnd = fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpochV1(startEpoch+1), ":")
			rangeStart = fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpochV1(endEpoch))
			ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
		}
	} else {
		// epochs are sorted descending, so start with the larges epoch and end with the smallest
		// add ':', a character lexicographically after digits, to make the range inclusive
		rangeEnd := fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpochV1(startEpoch), ":")
		rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpochV1(endEpoch))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
	}
	return ranges
}

func (bigtable *Bigtable) getSlotRangesV1(startSlot uint64, endSlot uint64) gcp_bigtable.RowRangeList {

	if endSlot < startSlot { // handle overflows
		startSlot = 0
	}

	ranges := gcp_bigtable.RowRangeList{}
	if startSlot == 0 { // special case when the 0 slot is included
		rangeEnd := fmt.Sprintf("%s:e:%s:s:%s\x00", bigtable.chainId, reversedPaddedEpochV1(0), reversedPaddedSlotV1(0))
		rangeStart := fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpochV1(0), reversedPaddedSlotV1(0))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))

		// epochs are sorted descending, so start with the larges epoch and end with the smallest
		// add ':', a character lexicographically after digits, to make the range inclusive
		if startSlot < endSlot {
			rangeEnd = fmt.Sprintf("%s:e:%s:s:%s\x00", bigtable.chainId, reversedPaddedEpochV1(utils.EpochOfSlot(startSlot)), reversedPaddedSlotV1(startSlot))
			rangeStart = fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpochV1(utils.EpochOfSlot(endSlot)), reversedPaddedSlotV1(endSlot))
			ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
		}
	} else {
		// epochs are sorted descending, so start with the larges epoch and end with the smallest
		// add ':', a character lexicographically after digits, to make the range inclusive
		rangeEnd := fmt.Sprintf("%s:e:%s:s:%s\x00", bigtable.chainId, reversedPaddedEpochV1(utils.EpochOfSlot(startSlot)), reversedPaddedSlotV1(startSlot))
		rangeStart := fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpochV1(utils.EpochOfSlot(endSlot)), reversedPaddedSlotV1(endSlot))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
	}
	return ranges
}

func (bigtable *Bigtable) getEpochRangesV1(startEpoch uint64, endEpoch uint64) gcp_bigtable.RowRangeList {

	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	ranges := gcp_bigtable.RowRangeList{}
	if startEpoch == 0 { // special case when the 0 epoch is included
		rangeEnd := fmt.Sprintf("%s:e:b:%s%s", bigtable.chainId, reversedPaddedEpochV1(0), "\x00")
		rangeStart := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpochV1(0))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))

		// epochs are sorted descending, so start with the largest epoch and end with the smallest
		// add \x00 to make the range inclusive
		if startEpoch < endEpoch {
			rangeEnd = fmt.Sprintf("%s:e:b:%s%s", bigtable.chainId, reversedPaddedEpochV1(startEpoch+1), "\x00")
			rangeStart = fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpochV1(endEpoch))
			ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
		}
	} else {
		// epochs are sorted descending, so start with the largest epoch and end with the smallest
		// add \x00 to make the range inclusive
		rangeEnd := fmt.Sprintf("%s:e:b:%s%s", bigtable.chainId, reversedPaddedEpochV1(startEpoch), "\x00")
		rangeStart := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpochV1(endEpoch))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
	}
	return ranges
}

func reversedPaddedEpochV1(epoch uint64) string {
	return fmt.Sprintf("%09d", max_block_number_v1-epoch)
}

func reversedPaddedSlotV1(slot uint64) string {
	return fmt.Sprintf("%09d", max_block_number_v1-slot)
}
