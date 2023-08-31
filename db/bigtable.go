package db

import (
	"context"
	"encoding/binary"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/go-redis/redis/v8"
	itypes "github.com/gobitfly/eth-rewards/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

var BigtableClient *Bigtable

const (
	DEFAULT_FAMILY                       = "f"
	VALIDATOR_BALANCES_FAMILY            = "vb"
	ATTESTATIONS_FAMILY                  = "at"
	PROPOSALS_FAMILY                     = "pr"
	SYNC_COMMITTEES_FAMILY               = "sc"
	SYNC_COMMITTEES_PARTICIPATION_FAMILY = "sp"
	INCOME_DETAILS_COLUMN_FAMILY         = "id"
	STATS_COLUMN_FAMILY                  = "stats"
	MACHINE_METRICS_COLUMN_FAMILY        = "mm"
	SERIES_FAMILY                        = "series"

	SUM_COLUMN = "sum"

	MAX_BLOCK_NUMBER = 1000000000 - 1
	MAX_EPOCH        = 1000000000 - 1

	MAX_BATCH_MUTATIONS = 100000
)

type Bigtable struct {
	client *gcp_bigtable.Client

	tableBeaconchain             *gcp_bigtable.Table
	tableValidators              *gcp_bigtable.Table
	tableValidatorBalances       *gcp_bigtable.Table
	tableValidatorAttestations   *gcp_bigtable.Table
	tableValidatorProposals      *gcp_bigtable.Table
	tableValidatorSyncCommittees *gcp_bigtable.Table
	tableValidatorIncomeDetails  *gcp_bigtable.Table

	tableData            *gcp_bigtable.Table
	tableBlocks          *gcp_bigtable.Table
	tableMetadataUpdates *gcp_bigtable.Table
	tableMetadata        *gcp_bigtable.Table

	tableMachineMetrics *gcp_bigtable.Table

	redisCache *redis.Client

	lastAttestationCache    map[uint64]uint64
	lastAttestationCacheMux *sync.Mutex

	chainId string
}

func InitBigtable(project, instance, chainId, redisAddress string) (*Bigtable, error) {

	if utils.Config.Bigtable.Emulator {
		logger.Infof("using emulated local bigtable environment, setting BIGTABLE_EMULATOR_HOST env variable to 127.0.0.1:%d", utils.Config.Bigtable.EmulatorPort)
		err := os.Setenv("BIGTABLE_EMULATOR_HOST", fmt.Sprintf("127.0.0.1:%d", utils.Config.Bigtable.EmulatorPort))

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
		client:                       btClient,
		tableData:                    btClient.Open("data"),
		tableBlocks:                  btClient.Open("blocks"),
		tableMetadataUpdates:         btClient.Open("metadata_updates"),
		tableMetadata:                btClient.Open("metadata"),
		tableBeaconchain:             btClient.Open("beaconchain"),
		tableMachineMetrics:          btClient.Open("machine_metrics"),
		tableValidators:              btClient.Open("beaconchain_validators"),
		tableValidatorBalances:       btClient.Open("beaconchain_validator_balances"),
		tableValidatorAttestations:   btClient.Open("beaconchain_validator_attestations"),
		tableValidatorProposals:      btClient.Open("beaconchain_validator_proposals"),
		tableValidatorSyncCommittees: btClient.Open("beaconchain_validator_sync"),
		tableValidatorIncomeDetails:  btClient.Open("beaconchain_validator_income"),
		chainId:                      chainId,
		redisCache:                   rdc,
		lastAttestationCacheMux:      &sync.Mutex{},
	}

	BigtableClient = bt
	return bt, nil
}

func (bigtable *Bigtable) Close() {
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

	err = bigtable.tableMachineMetrics.Apply(
		ctx,
		rowKeyData,
		dataMut,
	)
	if err != nil {
		return err
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	// start := time.Now()
	ts := gcp_bigtable.Timestamp(0)

	muts := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	keys := make([]string, 0, MAX_BATCH_MUTATIONS)

	epochKey := bigtable.reversedPaddedEpoch(epoch)
	for _, validator := range validators {
		balanceEncoded := make([]byte, 8)
		binary.LittleEndian.PutUint64(balanceEncoded, validator.Balance)

		effectiveBalanceEncoded := make([]byte, 8)
		binary.LittleEndian.PutUint64(effectiveBalanceEncoded, validator.EffectiveBalance)

		combined := append(balanceEncoded, effectiveBalanceEncoded...)
		mut := &gcp_bigtable.Mutation{}
		mut.Set(VALIDATOR_BALANCES_FAMILY, "b", ts, combined)
		key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator.Index), epochKey)

		muts = append(muts, mut)
		keys = append(keys, key)

		if len(muts) == MAX_BATCH_MUTATIONS {
			errs, err := bigtable.tableValidatorBalances.ApplyBulk(ctx, keys, muts)

			if err != nil {
				return err
			}

			for _, err := range errs {
				return err
			}
			muts = make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
			keys = make([]string, 0, MAX_BATCH_MUTATIONS)
		}
	}

	if len(muts) > 0 {
		errs, err := bigtable.tableValidatorBalances.ApplyBulk(ctx, keys, muts)

		if err != nil {
			return err
		}

		for _, err := range errs {
			return err
		}
	}

	// logger.Infof("exported validator balances to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveAttestationAssignments(epoch uint64, assignments map[string]uint64) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	start := time.Now()
	ts := gcp_bigtable.Timestamp(0)

	validatorsPerSlot := make(map[uint64][]uint64)
	for key, validator := range assignments {
		keySplit := strings.Split(key, "-")

		attesterslot, err := strconv.ParseUint(keySplit[0], 10, 64)
		if err != nil {
			return err
		}

		if validatorsPerSlot[attesterslot] == nil {
			validatorsPerSlot[attesterslot] = make([]uint64, 0, len(assignments)/int(utils.Config.Chain.Config.SlotsPerEpoch))
		}
		validatorsPerSlot[attesterslot] = append(validatorsPerSlot[attesterslot], validator)
	}

	muts := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	keys := make([]string, 0, MAX_BATCH_MUTATIONS)

	for slot, validators := range validatorsPerSlot {

		for _, validator := range validators {
			mut := gcp_bigtable.NewMutation()
			key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator), bigtable.reversedPaddedEpoch(epoch))
			mut.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", slot), ts, []byte{})

			muts = append(muts, mut)
			keys = append(keys, key)
		}

		if len(muts) == MAX_BATCH_MUTATIONS {
			errs, err := bigtable.tableValidatorAttestations.ApplyBulk(ctx, keys, muts)

			if err != nil {
				return err
			}

			for _, err := range errs {
				return err
			}
			muts = make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
			keys = make([]string, 0, MAX_BATCH_MUTATIONS)
		}
	}

	if len(muts) > 0 {
		errs, err := bigtable.tableValidatorAttestations.ApplyBulk(ctx, keys, muts)

		if err != nil {
			return err
		}
		for _, err := range errs {
			return err
		}
	}

	logger.Infof("exported attestation assignments to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveProposalAssignments(epoch uint64, assignments map[uint64]uint64) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()
	ts := gcp_bigtable.Timestamp(0)

	muts := make([]*gcp_bigtable.Mutation, 0, len(assignments))
	keys := make([]string, 0, len(assignments))

	for slot, validator := range assignments {
		mut := gcp_bigtable.NewMutation()
		mut.Set(PROPOSALS_FAMILY, "p", ts, []byte{})

		key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator), bigtable.reversedPaddedEpoch(epoch), bigtable.reversedPaddedSlot(slot))

		muts = append(muts, mut)
		keys = append(keys, key)
	}

	errs, err := bigtable.tableValidatorProposals.ApplyBulk(ctx, keys, muts)

	if err != nil {
		return err
	}

	for _, err := range errs {
		return err
	}

	logger.Infof("exported proposal assignments to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveSyncCommitteesAssignments(startSlot, endSlot uint64, validators []uint64) error {

	return nil //disabled as not needed

	// ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	// defer cancel()

	// start := time.Now()
	// ts := gcp_bigtable.Timestamp(0)

	// muts := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	// keys := make([]string, 0, MAX_BATCH_MUTATIONS)

	// for i := startSlot; i <= endSlot; i++ {
	// 	for _, validator := range validators {
	// 		mut := gcp_bigtable.NewMutation()
	// 		mut.Set(SYNC_COMMITTEES_FAMILY, "s", ts, []byte{})

	// 		key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator), bigtable.reversedPaddedEpoch(utils.EpochOfSlot(i)), bigtable.reversedPaddedSlot(i))

	// 		muts = append(muts, mut)
	// 		keys = append(keys, key)

	// 		if len(muts) == MAX_BATCH_MUTATIONS {
	// 			logger.Infof("saving %v mutations for sync duties", len(muts))
	// 			errs, err := bigtable.tableValidatorSyncCommittees.ApplyBulk(ctx, keys, muts)

	// 			if err != nil {
	// 				return err
	// 			}

	// 			for _, err := range errs {
	// 				if err != nil {
	// 					return err
	// 				}
	// 			}

	// 			muts = make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	// 			keys = make([]string, 0, MAX_BATCH_MUTATIONS)
	// 		}
	// 	}

	// }

	// if len(muts) > 0 {
	// 	logger.Infof("saving %v mutations for sync duties", len(muts))
	// 	errs, err := bigtable.tableValidatorSyncCommittees.ApplyBulk(ctx, keys, muts)

	// 	if err != nil {
	// 		return err
	// 	}

	// 	for _, err := range errs {
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	// logger.Infof("exported sync committee assignments to bigtable in %v", time.Since(start))
	// return nil
}

func (bigtable *Bigtable) SaveAttestations(blocks map[uint64]map[string]*types.Block) error {

	// Initialize in memory last attestation cache lazily
	bigtable.lastAttestationCacheMux.Lock()
	if bigtable.lastAttestationCache == nil {
		t := time.Now()
		var err error
		bigtable.lastAttestationCache, err = bigtable.GetLastAttestationSlots([]uint64{})

		if err != nil {
			bigtable.lastAttestationCacheMux.Unlock()
			return err
		}
		logger.Infof("initialized in memory last attestation slot cache with %v validators in %v", len(bigtable.lastAttestationCache), time.Since(t))

	}
	bigtable.lastAttestationCacheMux.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	start := time.Now()

	attestationsBySlot := make(map[uint64]map[uint64]uint64) //map[attestedSlot]map[validator]includedSlot

	slots := make([]uint64, 0, len(blocks))
	for slot := range blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	for _, slot := range slots {
		for _, b := range blocks[slot] {
			// logger.Infof("processing slot %v", slot)
			for _, a := range b.Attestations {
				for _, validator := range a.Attesters {
					inclusionSlot := slot
					attestedSlot := a.Data.Slot
					if attestationsBySlot[attestedSlot] == nil {
						attestationsBySlot[attestedSlot] = make(map[uint64]uint64)
					}

					if attestationsBySlot[attestedSlot][validator] == 0 || inclusionSlot < attestationsBySlot[attestedSlot][validator] {
						attestationsBySlot[attestedSlot][validator] = inclusionSlot
					}
				}
			}
		}
	}

	mutsInclusionSlot := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	keysInclusionSlot := make([]string, 0, MAX_BATCH_MUTATIONS)

	for attestedSlot, inclusions := range attestationsBySlot {
		mutLastAttestationSlot := gcp_bigtable.NewMutation()
		mutLastAttestationSlotSet := false

		epoch := utils.EpochOfSlot(attestedSlot)
		bigtable.lastAttestationCacheMux.Lock()
		for validator, inclusionSlot := range inclusions {

			mutInclusionSlot := gcp_bigtable.NewMutation()
			mutInclusionSlot.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", attestedSlot), gcp_bigtable.Timestamp((MAX_BLOCK_NUMBER-inclusionSlot)*1000), []byte{})
			key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator), bigtable.reversedPaddedEpoch(epoch))

			mutsInclusionSlot = append(mutsInclusionSlot, mutInclusionSlot)
			keysInclusionSlot = append(keysInclusionSlot, key)
			if attestedSlot > bigtable.lastAttestationCache[validator] {
				mutLastAttestationSlot.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", validator), gcp_bigtable.Timestamp((attestedSlot)*1000), []byte{})
				bigtable.lastAttestationCache[validator] = attestedSlot
				mutLastAttestationSlotSet = true
			}

		}
		bigtable.lastAttestationCacheMux.Unlock()

		attstart := time.Now()

		if len(mutsInclusionSlot) == MAX_BATCH_MUTATIONS {
			errs, err := bigtable.tableValidatorAttestations.ApplyBulk(ctx, keysInclusionSlot, mutsInclusionSlot)
			if err != nil {
				return err
			}
			for _, err := range errs {
				return err
			}
			logger.Infof("applied attestation mutations in %v", time.Since(attstart))
			mutsInclusionSlot = make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
			keysInclusionSlot = make([]string, 0, MAX_BATCH_MUTATIONS)
		}
		if mutLastAttestationSlotSet {
			err := bigtable.tableValidators.Apply(ctx, fmt.Sprintf("%s:lastAttestationSlot", bigtable.chainId), mutLastAttestationSlot)
			if err != nil {
				return err
			}
		}
	}

	if len(mutsInclusionSlot) > 0 {
		attstart := time.Now()
		errs, err := bigtable.tableValidatorAttestations.ApplyBulk(ctx, keysInclusionSlot, mutsInclusionSlot)
		if err != nil {
			return err
		}
		for _, err := range errs {
			return err
		}
		logger.Infof("applied attestation mutations in %v", time.Since(attstart))
	}

	logger.Infof("exported attestations (new) to bigtable in %v", time.Since(start))
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

func (bigtable *Bigtable) SaveProposals(blocks map[uint64]map[string]*types.Block) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()

	slots := make([]uint64, 0, len(blocks))
	for slot := range blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	muts := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	keys := make([]string, 0, MAX_BATCH_MUTATIONS)

	for _, slot := range slots {
		for _, b := range blocks[slot] {

			if len(b.BlockRoot) != 32 { // skip dummy blocks
				continue
			}
			mut := gcp_bigtable.NewMutation()
			mut.Set(PROPOSALS_FAMILY, "b", gcp_bigtable.Timestamp((MAX_BLOCK_NUMBER-b.Slot)*1000), []byte{})
			key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(b.Proposer), bigtable.reversedPaddedEpoch(utils.EpochOfSlot(b.Slot)), bigtable.reversedPaddedSlot(b.Slot))

			muts = append(muts, mut)
			keys = append(keys, key)
		}
	}
	errs, err := bigtable.tableValidatorProposals.ApplyBulk(ctx, keys, muts)

	if err != nil {
		return err
	}

	for _, err := range errs {
		return err
	}
	logger.Infof("exported proposals to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveSyncComitteeDuties(blocks map[uint64]map[string]*types.Block) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	start := time.Now()

	dutiesBySlot := make(map[uint64]map[uint64]bool) //map[dutiesSlot]map[validator]bool

	slots := make([]uint64, 0, len(blocks))
	for slot := range blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	for _, slot := range slots {
		for _, b := range blocks[slot] {
			if b.Status == 2 {
				continue
			} else if b.SyncAggregate != nil && len(b.SyncAggregate.SyncCommitteeValidators) > 0 {
				bitLen := len(b.SyncAggregate.SyncCommitteeBits) * 8
				valLen := len(b.SyncAggregate.SyncCommitteeValidators)
				if bitLen < valLen {
					return fmt.Errorf("error getting sync_committee participants: bitLen != valLen: %v != %v", bitLen, valLen)
				}
				for i, valIndex := range b.SyncAggregate.SyncCommitteeValidators {
					if dutiesBySlot[b.Slot] == nil {
						dutiesBySlot[b.Slot] = make(map[uint64]bool)
					}
					dutiesBySlot[b.Slot][valIndex] = utils.BitAtVector(b.SyncAggregate.SyncCommitteeBits, i)
				}
			}
		}
	}

	if len(dutiesBySlot) == 0 {
		logger.Infof("no sync duties to export")
		return nil
	}

	muts := make([]*gcp_bigtable.Mutation, 0, utils.Config.Chain.Config.SlotsPerEpoch*utils.Config.Chain.Config.SyncCommitteeSize+1)
	keys := make([]string, 0, utils.Config.Chain.Config.SlotsPerEpoch*utils.Config.Chain.Config.SyncCommitteeSize+1)

	for slot, validators := range dutiesBySlot {
		participation := uint64(0)
		for validator, participated := range validators {
			mut := gcp_bigtable.NewMutation()
			if participated {
				mut.Set(SYNC_COMMITTEES_FAMILY, "s", gcp_bigtable.Timestamp((MAX_BLOCK_NUMBER-slot)*1000), []byte{})
				participation++
			} else {
				mut.Set(SYNC_COMMITTEES_FAMILY, "s", gcp_bigtable.Timestamp(0), []byte{})
			}
			key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator), bigtable.reversedPaddedEpoch(utils.EpochOfSlot(slot)), bigtable.reversedPaddedSlot(slot))
			muts = append(muts, mut)
			keys = append(keys, key)
		}
		mut := gcp_bigtable.NewMutation()
		key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, SYNC_COMMITTEES_PARTICIPATION_FAMILY, bigtable.reversedPaddedSlot(slot))
		participationEncoded := make([]byte, 8)
		binary.LittleEndian.PutUint64(participationEncoded, uint64(participation))
		mut.Set(SYNC_COMMITTEES_PARTICIPATION_FAMILY, "s", gcp_bigtable.Timestamp(0), participationEncoded)
		muts = append(muts, mut)
		keys = append(keys, key)
	}

	errs, err := bigtable.tableValidatorSyncCommittees.ApplyBulk(ctx, keys, muts)

	if err != nil {
		return err
	}

	for _, err := range errs {
		return err
	}

	logger.Infof("exported sync committee duties to bigtable in %v", time.Since(start))
	return nil
}

// GetMaxValidatorindexForEpoch returns the higest validatorindex with a balance at that epoch
func (bigtable *Bigtable) GetMaxValidatorindexForEpoch(epoch uint64) (uint64, error) {

	// TODO: Implement

	return 0, nil
}

func (bigtable *Bigtable) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
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
			ranges := bigtable.getValidatorsEpochRanges(vals, startEpoch, endEpoch)
			ro := gcp_bigtable.LimitRows(int64(endEpoch-startEpoch+1) * int64(len(vals)))

			handleRow := func(r gcp_bigtable.Row) bool {
				keySplit := strings.Split(r.Key(), ":")

				epoch, err := strconv.ParseUint(keySplit[2], 10, 64)
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
					effectiveBalanceBytes := balances[8:16]
					balance := binary.LittleEndian.Uint64(balanceBytes)
					effectiveBalance := binary.LittleEndian.Uint64(effectiveBalanceBytes)

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

			err := bigtable.tableValidatorBalances.ReadRows(gCtx, ranges, handleRow, ro)
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

func (bigtable *Bigtable) GetValidatorAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error) {

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	slots := []uint64{}

	for slot := startEpoch * utils.Config.Chain.Config.SlotsPerEpoch; slot < (endEpoch+1)*utils.Config.Chain.Config.SlotsPerEpoch; slot++ {
		slots = append(slots, slot)
	}
	orphanedSlotsMap, err := GetOrphanedSlotsMap(slots)
	if err != nil {
		return nil, err
	}

	res := make(map[uint64][]*types.ValidatorAttestation, len(validators))
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
			ranges := bigtable.getValidatorsEpochRanges(vals, startEpoch, endEpoch)
			err = bigtable.tableValidatorAttestations.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
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
					inclusionSlot := MAX_BLOCK_NUMBER - uint64(ri.Timestamp)/1000

					status := uint64(1)
					if inclusionSlot == MAX_BLOCK_NUMBER {
						inclusionSlot = 0
						status = 0
					} else if orphanedSlotsMap[inclusionSlot] {
						status = 0
					}

					resMux.Lock()
					if res[validator] == nil {
						res[validator] = make([]*types.ValidatorAttestation, 0)
					}

					if len(res[validator]) > 0 && res[validator][len(res[validator])-1].AttesterSlot == attesterSlot {
						// don't override successful attestion, that was included in a different slot
						if status == 1 || res[validator][len(res[validator])-1].Status != 1 {
							res[validator][len(res[validator])-1].InclusionSlot = inclusionSlot
							res[validator][len(res[validator])-1].Status = status
							res[validator][len(res[validator])-1].Delay = int64(inclusionSlot - attesterSlot)
						}
					} else {
						res[validator] = append(res[validator], &types.ValidatorAttestation{
							Index:          validator,
							Epoch:          utils.EpochOfSlot(attesterSlot),
							AttesterSlot:   attesterSlot,
							CommitteeIndex: 0,
							Status:         status,
							InclusionSlot:  inclusionSlot,
							Delay:          int64(inclusionSlot) - int64(attesterSlot) - 1,
						})
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

func (bigtable *Bigtable) GetLastAttestationSlots(validators []uint64) (map[uint64]uint64, error) {
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

func (bigtable *Bigtable) GetSyncParticipationBySlot(slot uint64) (uint64, error) {

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, SYNC_COMMITTEES_PARTICIPATION_FAMILY, bigtable.reversedPaddedSlot(slot))

	row, err := bigtable.tableValidatorSyncCommittees.ReadRow(ctx, key)
	if err != nil {
		return 0, err
	}

	for _, ri := range row[SYNC_COMMITTEES_PARTICIPATION_FAMILY] {
		return binary.LittleEndian.Uint64(ri.Value), nil
	}

	return 0, nil
}

func (bigtable *Bigtable) GetValidatorMissedAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]bool, error) {

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()

	slots := []uint64{}

	for slot := startEpoch * utils.Config.Chain.Config.SlotsPerEpoch; slot < (endEpoch+1)*utils.Config.Chain.Config.SlotsPerEpoch; slot++ {
		slots = append(slots, slot)
	}
	orphanedSlotsMap, err := GetOrphanedSlotsMap(slots)
	if err != nil {
		return nil, err
	}

	res := make(map[uint64]map[uint64]bool)
	foundValid := make(map[uint64]map[uint64]bool)

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
			ranges := bigtable.getValidatorsEpochRanges(vals, startEpoch, endEpoch)
			err = bigtable.tableValidatorAttestations.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
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

					inclusionSlot := MAX_BLOCK_NUMBER - uint64(ri.Timestamp)/1000

					status := uint64(1)
					if inclusionSlot == MAX_BLOCK_NUMBER {
						status = 0
					}

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
			}, gcp_bigtable.RowFilter(filter))

			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorSyncDutiesHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorSyncParticipation, error) {

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	res := make(map[uint64][]*types.ValidatorSyncParticipation, len(validators))
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
			ranges := bigtable.getValidatorsEpochRanges(vals, startEpoch, endEpoch)
			err := bigtable.tableValidatorSyncCommittees.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {

				for _, ri := range r[SYNC_COMMITTEES_FAMILY] {
					keySplit := strings.Split(r.Key(), ":")

					validator, err := bigtable.validatorKeyToIndex(keySplit[1])
					if err != nil {
						logger.Errorf("error parsing validator from row key %v: %v", r.Key(), err)
						return false
					}

					slot, err := strconv.ParseUint(keySplit[3], 10, 64)
					if err != nil {
						logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
						return false
					}
					slot = MAX_BLOCK_NUMBER - slot
					inclusionSlot := MAX_BLOCK_NUMBER - uint64(ri.Timestamp)/1000

					status := uint64(1) // 1: participated
					if inclusionSlot == MAX_BLOCK_NUMBER {
						inclusionSlot = 0
						status = 0 // 0: missed
					}

					resMux.Lock()
					if res[validator] == nil {
						res[validator] = make([]*types.ValidatorSyncParticipation, 0)
					}

					if len(res[validator]) > 0 && res[validator][len(res[validator])-1].Slot == slot {
						res[validator][len(res[validator])-1].Status = status
					} else {
						res[validator] = append(res[validator], &types.ValidatorSyncParticipation{
							Slot:   slot,
							Status: status,
						})
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

func (bigtable *Bigtable) GetValidatorMissedAttestationsCount(validators []uint64, firstEpoch uint64, lastEpoch uint64) (map[uint64]*types.ValidatorMissedAttestationsStatistic, error) {
	if firstEpoch > lastEpoch {
		return nil, fmt.Errorf("GetValidatorMissedAttestationsCount received an invalid firstEpoch (%d) and lastEpoch (%d) combination", firstEpoch, lastEpoch)
	}

	res := make(map[uint64]*types.ValidatorMissedAttestationsStatistic)

	data, err := bigtable.GetValidatorMissedAttestationHistory(validators, firstEpoch, lastEpoch)

	if err != nil {
		return nil, err
	}

	logger.Infof("retrieved missed attestation history for epochs %v - %v", firstEpoch, lastEpoch)

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
	data, err := bigtable.GetValidatorSyncDutiesHistory(validators, startEpoch, endEpoch)

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
	data, err := bigtable.GetValidatorAttestationHistory(validators, epoch-99, epoch)

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

		logrus.Infof("retrieving validator balance stats for validators %v - %v", vals[0], vals[len(vals)-1])

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

func (bigtable *Bigtable) GetValidatorProposalHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorProposal, error) {
	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	res := make(map[uint64][]*types.ValidatorProposal, len(validators))
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
			ranges := bigtable.getValidatorsEpochRanges(vals, startEpoch, endEpoch)
			err := bigtable.tableValidatorProposals.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
				for _, ri := range r[PROPOSALS_FAMILY] {
					keySplit := strings.Split(r.Key(), ":")

					proposalSlot, err := strconv.ParseUint(keySplit[3], 10, 64)
					if err != nil {
						logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
						return false
					}
					proposalSlot = MAX_BLOCK_NUMBER - proposalSlot
					inclusionSlot := MAX_BLOCK_NUMBER - uint64(r[PROPOSALS_FAMILY][0].Timestamp)/1000

					status := uint64(1)
					if inclusionSlot == MAX_BLOCK_NUMBER {
						inclusionSlot = 0
						status = 2
					}

					validator, err := bigtable.validatorKeyToIndex(keySplit[1])
					if err != nil {
						logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
						return false
					}

					resMux.Lock()
					if res[validator] == nil {
						res[validator] = make([]*types.ValidatorProposal, 0)
					}

					if len(res[validator]) > 0 && res[validator][len(res[validator])-1].Slot == proposalSlot {
						res[validator][len(res[validator])-1].Slot = proposalSlot
						res[validator][len(res[validator])-1].Status = status
					} else {
						res[validator] = append(res[validator], &types.ValidatorProposal{
							Index:  validator,
							Status: status,
							Slot:   proposalSlot,
						})
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

func (bigtable *Bigtable) SaveValidatorIncomeDetails(epoch uint64, rewards map[uint64]*itypes.ValidatorEpochIncome) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	start := time.Now()
	ts := gcp_bigtable.Timestamp(utils.EpochToTime(epoch).UnixMicro())

	total := &itypes.ValidatorEpochIncome{}

	muts := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	keys := make([]string, 0, MAX_BATCH_MUTATIONS)

	mutsCount := 0
	for i, rewardDetails := range rewards {

		data, err := proto.Marshal(rewardDetails)

		if err != nil {
			return err
		}

		mut := &gcp_bigtable.Mutation{}
		mut.Set(INCOME_DETAILS_COLUMN_FAMILY, "i", ts, data)
		key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(i), bigtable.reversedPaddedEpoch(epoch))

		muts = append(muts, mut)
		keys = append(keys, key)
		mutsCount++

		if len(muts) == MAX_BATCH_MUTATIONS {
			errs, err := bigtable.tableValidatorIncomeDetails.ApplyBulk(ctx, keys, muts)

			if err != nil {
				return err
			}
			for _, err := range errs {
				return err
			}
			muts = make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
			keys = make([]string, 0, MAX_BATCH_MUTATIONS)
		}

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

	if len(muts) > 0 {
		errs, err := bigtable.tableValidatorIncomeDetails.ApplyBulk(ctx, keys, muts)

		if err != nil {
			return err
		}
		for _, err := range errs {
			return err
		}
	}

	sum, err := proto.Marshal(total)
	if err != nil {
		return err
	}

	mut := &gcp_bigtable.Mutation{}
	mut.Set(STATS_COLUMN_FAMILY, SUM_COLUMN, ts, sum)

	err = bigtable.tableValidatorIncomeDetails.Apply(ctx, fmt.Sprintf("%s:%s:%s", bigtable.chainId, SUM_COLUMN, bigtable.reversedPaddedEpoch(epoch)), mut)
	if err != nil {
		return err
	}

	logger.Infof("exported validator income details for epoch %v to bigtable in %v", epoch, time.Since(start))
	return nil
}

func (bigtable *Bigtable) MigrateEpochSchemaV1ToV2(epoch uint64) error {
	funcStart := time.Now()

	defer func() {
		logger.Infof("migration of epoch %v completed in %v", epoch, time.Since(funcStart))
	}()

	// start := time.Now()

	type validatorEpochData struct {
		ValidatorIndex           uint64
		Proposals                map[uint64]uint64
		AttestationTargetSlot    uint64
		AttestationInclusionSlot uint64
		SyncParticipation        map[uint64]uint64
		EffectiveBalance         uint64
		Balance                  uint64
		IncomeDetails            *itypes.ValidatorEpochIncome
	}

	epochData := make(map[uint64]*validatorEpochData)
	filter := gcp_bigtable.LatestNFilter(1)
	ctx := context.Background()

	prefixEpochRange := gcp_bigtable.PrefixRange(fmt.Sprintf("%s:e:b:%s", bigtable.chainId, bigtable.reversedPaddedEpoch(epoch)))

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

				if epochData[validator] == nil {
					epochData[validator] = &validatorEpochData{
						ValidatorIndex:    validator,
						Proposals:         make(map[uint64]uint64),
						SyncParticipation: make(map[uint64]uint64),
					}
				}

				if columnFamily == VALIDATOR_BALANCES_FAMILY {
					// logger.Infof("processing balance data for validator %d", validator)
					balances := ri.Value
					balanceBytes := balances[0:8]
					effectiveBalanceBytes := balances[8:16]
					epochData[validator].Balance = binary.LittleEndian.Uint64(balanceBytes)
					epochData[validator].EffectiveBalance = binary.LittleEndian.Uint64(effectiveBalanceBytes)
				} else if columnFamily == INCOME_DETAILS_COLUMN_FAMILY {
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

	// logger.Infof("retrieved epoch data for %d validators in %v", len(epochData), time.Since(start))
	// start = time.Now()

	prefixEpochSlotRange := gcp_bigtable.PrefixRange(fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, bigtable.reversedPaddedEpoch(epoch)))

	err = bigtable.tableBeaconchain.ReadRows(ctx, prefixEpochSlotRange, func(r gcp_bigtable.Row) bool {
		// logger.Infof("processing row %v", r.Key())

		keySplit := strings.Split(r.Key(), ":")

		rowKeyEpoch, err := strconv.ParseUint(keySplit[2], 10, 64)
		if err != nil {
			logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
			return false
		}

		rowKeyEpoch = MAX_EPOCH - rowKeyEpoch

		if epoch != rowKeyEpoch {
			logger.Errorf("retrieved different epoch than requested, requested: %d, retrieved: %d", epoch, rowKeyEpoch)
		}

		slot, err := strconv.ParseUint(keySplit[4], 10, 64)
		if err != nil {
			logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
			return false
		}
		slot = MAX_BLOCK_NUMBER - slot

		// logger.Infof("epoch is %d, slot is %d", rowKeyEpoch, slot)

		for columnFamily, readItems := range r {

			for _, ri := range readItems {

				validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, columnFamily+":"), 10, 64)
				if err != nil {
					logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
					return false
				}

				inclusionSlot := uint64(0)

				if ri.Timestamp > 0 {
					inclusionSlot = MAX_BLOCK_NUMBER - uint64(ri.Timestamp)/1000
				}

				// logger.Infof("retrieved field %s from column family %s for validator %d", ri.Column, columnFamily, validator)

				if epochData[validator] == nil {
					epochData[validator] = &validatorEpochData{
						ValidatorIndex:    validator,
						Proposals:         make(map[uint64]uint64),
						SyncParticipation: make(map[uint64]uint64),
					}
				}

				if columnFamily == ATTESTATIONS_FAMILY {
					// logger.Infof("processing balance data for validator %d", validator)
					epochData[validator].AttestationTargetSlot = slot
					epochData[validator].AttestationInclusionSlot = inclusionSlot
					// logger.Infof("processing attestation data for validator %d, target slot %d, inclusion slot %d", validator, slot, inclusionSlot)
				} else if columnFamily == PROPOSALS_FAMILY {
					epochData[validator].Proposals[slot] = inclusionSlot
					// logger.Infof("processing proposer data for validator %d, proposal slot %d, inclusion slot %d", validator, slot, inclusionSlot)
				} else if columnFamily == SYNC_COMMITTEES_FAMILY {
					epochData[validator].SyncParticipation[slot] = inclusionSlot
					//logger.Infof("processing sync data for validator %d, proposal slot %d, inclusion slot %d", validator, slot, inclusionSlot)
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

	// logger.Infof("retrieved slot data for %d validators in %v", len(epochData), time.Since(start))
	// start = time.Now()

	// save validator balance data
	validators := make([]*types.Validator, 0, len(epochData))

	for _, validator := range epochData {
		validators = append(validators, &types.Validator{
			Index:            validator.ValidatorIndex,
			EffectiveBalance: validator.EffectiveBalance,
			Balance:          validator.Balance,
		})
	}

	err = bigtable.SaveValidatorBalances(epoch, validators)
	if err != nil {
		return err
	}
	// logger.Infof("migrated balance data in %v", time.Since(start))
	// start = time.Now()

	i := 0
	mutsInclusionSlot := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	keysInclusionSlot := make([]string, 0, MAX_BATCH_MUTATIONS)

	for _, validator := range epochData {
		mutInclusionSlot := gcp_bigtable.NewMutation()
		mutInclusionSlot.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", validator.AttestationTargetSlot), gcp_bigtable.Timestamp((MAX_BLOCK_NUMBER-validator.AttestationInclusionSlot)*1000), []byte{})
		key := fmt.Sprintf("%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator.ValidatorIndex), bigtable.reversedPaddedEpoch(epoch))

		mutsInclusionSlot = append(mutsInclusionSlot, mutInclusionSlot)
		keysInclusionSlot = append(keysInclusionSlot, key)

		if len(mutsInclusionSlot) == MAX_BATCH_MUTATIONS {
			errs, err := bigtable.tableValidatorAttestations.ApplyBulk(ctx, keysInclusionSlot, mutsInclusionSlot)
			if err != nil {
				return err
			}
			for _, err := range errs {
				return err
			}
			mutsInclusionSlot = make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
			keysInclusionSlot = make([]string, 0, MAX_BATCH_MUTATIONS)
		}
		i++
	}

	if len(mutsInclusionSlot) > 0 {
		errs, err := bigtable.tableValidatorAttestations.ApplyBulk(ctx, keysInclusionSlot, mutsInclusionSlot)

		if err != nil {
			return err
		}

		for _, err := range errs {
			return err
		}
	}
	// logger.Infof("migrated attestation data in %v", time.Since(start))
	// start = time.Now()

	mutsProposals := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	keysProposals := make([]string, 0, MAX_BATCH_MUTATIONS)

	for _, validator := range epochData {
		if len(validator.Proposals) == 0 {
			continue
		}
		for slot, inclusionSlot := range validator.Proposals {
			mut := gcp_bigtable.NewMutation()
			mut.Set(PROPOSALS_FAMILY, "b", gcp_bigtable.Timestamp((MAX_BLOCK_NUMBER-inclusionSlot)*1000), []byte{})
			key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator.ValidatorIndex), bigtable.reversedPaddedEpoch(epoch), bigtable.reversedPaddedSlot(slot))

			mutsProposals = append(mutsProposals, mut)
			keysProposals = append(keysProposals, key)
		}
	}
	errs, err := bigtable.tableValidatorProposals.ApplyBulk(ctx, keysProposals, mutsProposals)

	if err != nil {
		return err
	}

	for _, err := range errs {
		return err
	}
	// logger.Infof("migrated proposal data in %v", time.Since(start))
	// start = time.Now()

	mutsSync := make([]*gcp_bigtable.Mutation, 0, MAX_BATCH_MUTATIONS)
	keysSync := make([]string, 0, MAX_BATCH_MUTATIONS)

	for _, validator := range epochData {

		if len(validator.SyncParticipation) == 0 {
			continue
		}
		for slot, inclusionSlot := range validator.SyncParticipation {
			mut := gcp_bigtable.NewMutation()
			mut.Set(SYNC_COMMITTEES_FAMILY, "s", gcp_bigtable.Timestamp((MAX_BLOCK_NUMBER-inclusionSlot)*1000), []byte{})

			key := fmt.Sprintf("%s:%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validator.ValidatorIndex), bigtable.reversedPaddedEpoch(epoch), bigtable.reversedPaddedSlot(slot))
			mutsSync = append(mutsSync, mut)
			keysSync = append(keysSync, key)
		}
	}

	errs, err = bigtable.tableValidatorSyncCommittees.ApplyBulk(ctx, keysSync, mutsSync)

	if err != nil {
		return err
	}

	for _, err := range errs {
		return err
	}
	// logger.Infof("migrated sync data in %v", time.Since(start))
	// start = time.Now()

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

	// logger.Infof("migrated income data in %v", time.Since(start))
	// start = time.Now()

	return nil

}

// GetValidatorIncomeDetailsHistory returns the validator income details
// startEpoch & endEpoch are inclusive
func (bigtable *Bigtable) GetValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*itypes.ValidatorEpochIncome, error) {

	if len(validators) == 0 {
		return nil, fmt.Errorf("passing empty validator array is unsupported")
	}

	batchSize := 1000
	concurrency := 10

	if startEpoch > endEpoch {
		startEpoch = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*180)
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
			ranges := bigtable.getValidatorsEpochRanges(vals, startEpoch, endEpoch)
			err := bigtable.tableValidatorIncomeDetails.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
				keySplit := strings.Split(r.Key(), ":")

				validator, err := bigtable.validatorKeyToIndex(keySplit[1])
				if err != nil {
					logger.Errorf("error parsing validator from row key %v: %v", r.Key(), err)
					return false
				}

				epoch, err := strconv.ParseUint(keySplit[2], 10, 64)
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

// Deletes all block data from bigtable
func (bigtable *Bigtable) DeleteEpoch(epoch uint64) error {
	// TOTO: Implement
	return fmt.Errorf("NOT IMPLEMENTED")
}

func (bigtable *Bigtable) getValidatorsEpochRanges(validatorIndices []uint64, startEpoch uint64, endEpoch uint64) gcp_bigtable.RowRangeList {

	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	ranges := make(gcp_bigtable.RowRangeList, 0, int((endEpoch-startEpoch))*len(validatorIndices))

	for _, validatorIndex := range validatorIndices {
		validatorKey := bigtable.validatorIndexToKey(validatorIndex)

		// epochs are sorted descending, so start with the largest epoch and end with the smallest
		// add \x00 to make the range inclusive
		rangeEnd := fmt.Sprintf("%s:%s:%s%s", bigtable.chainId, validatorKey, bigtable.reversedPaddedEpoch(startEpoch), "\x00")
		rangeStart := fmt.Sprintf("%s:%s:%s", bigtable.chainId, bigtable.validatorIndexToKey(validatorIndex), bigtable.reversedPaddedEpoch(endEpoch))
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

func (bigtable *Bigtable) ClearByPrefix(family, prefix string, dryRun bool) ([]string, error) {
	if family == "" || prefix == "" {
		return []string{}, fmt.Errorf("please provide family [%v] and prefix [%v]", family, prefix)
	}

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	rowRange := gcp_bigtable.PrefixRange(prefix)
	deleteKeys := []string{}

	err := bigtable.tableData.ReadRows(ctx, rowRange, func(row gcp_bigtable.Row) bool {
		row_ := row[family][0]
		deleteKeys = append(deleteKeys, row_.Row)
		return true
	})
	if err != nil {
		return deleteKeys, err
	}

	if len(deleteKeys) == 0 {
		return deleteKeys, fmt.Errorf("no keys found")
	}

	if dryRun {
		return deleteKeys, nil
	}

	mutsDelete := &types.BulkMutations{
		Keys: make([]string, 0, len(deleteKeys)),
		Muts: make([]*gcp_bigtable.Mutation, 0, len(deleteKeys)),
	}
	for _, key := range deleteKeys {
		mutDelete := gcp_bigtable.NewMutation()
		mutDelete.DeleteRow()
		mutsDelete.Keys = append(mutsDelete.Keys, key)
		mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
	}
	err = bigtable.WriteBulk(mutsDelete, bigtable.tableData)
	return deleteKeys, err
}

func GetCurrentDayClIncome(validator_indices []uint64) (map[uint64]int64, error) {
	dayIncome := make(map[uint64]int64)
	lastDay, err := GetLastExportedStatisticDay()
	if err != nil {
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
	return fmt.Sprintf("%09d", MAX_BLOCK_NUMBER-epoch)
}

func (bigtable *Bigtable) reversedPaddedSlot(slot uint64) string {
	return fmt.Sprintf("%09d", MAX_BLOCK_NUMBER-slot)
}
