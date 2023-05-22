package db

import (
	"context"
	"encoding/binary"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/go-redis/redis/v8"
	itypes "github.com/gobitfly/eth-rewards/types"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

var BigtableClient *Bigtable

const (
	DEFAULT_FAMILY                = "f"
	VALIDATOR_BALANCES_FAMILY     = "vb"
	ATTESTATIONS_FAMILY           = "at"
	PROPOSALS_FAMILY              = "pr"
	SYNC_COMMITTEES_FAMILY        = "sc"
	INCOME_DETAILS_COLUMN_FAMILY  = "id"
	STATS_COLUMN_FAMILY           = "stats"
	MACHINE_METRICS_COLUMN_FAMILY = "mm"
	SERIES_FAMILY                 = "series"

	SUM_COLUMN = "sum"

	max_block_number = 1000000000
	max_epoch        = 1000000000
)

type Bigtable struct {
	client *gcp_bigtable.Client

	tableBeaconchain *gcp_bigtable.Table

	tableData            *gcp_bigtable.Table
	tableBlocks          *gcp_bigtable.Table
	tableMetadataUpdates *gcp_bigtable.Table
	tableMetadata        *gcp_bigtable.Table
	tableMachineMetrics  *gcp_bigtable.Table

	redisCache *redis.Client

	chainId string
}

func InitBigtable(project, instance, chainId, redisAddress string) (*Bigtable, error) {
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
		client:               btClient,
		tableData:            btClient.Open("data"),
		tableBlocks:          btClient.Open("blocks"),
		tableMetadataUpdates: btClient.Open("metadata_updates"),
		tableMetadata:        btClient.Open("metadata"),
		tableBeaconchain:     btClient.Open("beaconchain"),
		tableMachineMetrics:  btClient.Open("machine_metrics"),
		chainId:              chainId,
		redisCache:           rdc,
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

	rowKeyData := fmt.Sprintf("u:%s:p:%s:m:%v", reversePaddedUserID(userID), process, machine)

	ts := gcp_bigtable.Now()
	rateLimitKey := fmt.Sprintf("%s:%d", rowKeyData, ts.Time().Minute())
	keySet, err := bigtable.redisCache.SetNX(ctx, rateLimitKey, "1", time.Minute).Result()
	if err != nil {
		return err
	}
	if !keySet {
		return fmt.Errorf("rate limit, last metric insert was less than 1 min ago")
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

	rangePrefix := fmt.Sprintf("u:%s:p:", reversePaddedUserID(userID))

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
	names, err := bigtable.getMachineMetricNamesMap(userID, 15)
	if err != nil {
		return 0, err
	}

	return uint64(len(names)), nil
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

	rangePrefix := fmt.Sprintf("u:%s:p:%s:m:", reversePaddedUserID(userID), process)
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

func GetMachineRowKey(userID uint64, process string, machine string) string {
	return fmt.Sprintf("u:%s:p:%s:m:%s", reversePaddedUserID(userID), process, machine)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()
	ts := gcp_bigtable.Timestamp(0)

	mut := gcp_bigtable.NewMutation()

	for i, validator := range validators {
		balanceEncoded := make([]byte, 8)
		binary.LittleEndian.PutUint64(balanceEncoded, validator.Balance)

		effectiveBalanceEncoded := make([]byte, 8)
		binary.LittleEndian.PutUint64(effectiveBalanceEncoded, validator.EffectiveBalance)

		combined := append(balanceEncoded, effectiveBalanceEncoded...)
		mut.Set(VALIDATOR_BALANCES_FAMILY, fmt.Sprintf("%d", validator.Index), ts, combined)

		if i%100000 == 0 {
			err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(epoch)), mut)

			if err != nil {
				return err
			}
			mut = gcp_bigtable.NewMutation()
		}
	}
	err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(epoch)), mut)

	if err != nil {
		return err
	}

	logger.Infof("exported validator balances to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveAttestationAssignments(epoch uint64, assignments map[string]uint64) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
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

	for slot, validators := range validatorsPerSlot {
		mut := gcp_bigtable.NewMutation()
		for _, validator := range validators {
			mut.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", validator), ts, []byte{})
		}
		err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(epoch), reversedPaddedSlot(slot)), mut)

		if err != nil {
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

	for slot, validator := range assignments {
		mut := gcp_bigtable.NewMutation()
		mut.Set(PROPOSALS_FAMILY, fmt.Sprintf("%d", validator), ts, []byte{})
		err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(epoch), reversedPaddedSlot(slot)), mut)

		if err != nil {
			return err
		}
	}

	logger.Infof("exported proposal assignments to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveSyncCommitteesAssignments(startSlot, endSlot uint64, validators []uint64) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	start := time.Now()
	ts := gcp_bigtable.Timestamp(0)

	var muts []*gcp_bigtable.Mutation
	var keys []string

	for i := startSlot; i <= endSlot; i++ {
		mut := gcp_bigtable.NewMutation()
		for _, validator := range validators {
			mut.Set(SYNC_COMMITTEES_FAMILY, fmt.Sprintf("%d", validator), ts, []byte{})
		}

		muts = append(muts, mut)
		keys = append(keys, fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(i/utils.Config.Chain.Config.SlotsPerEpoch), reversedPaddedSlot(i)))
	}

	logger.Infof("saving %v mutations for sync duties", len(muts))

	errs, err := bigtable.tableBeaconchain.ApplyBulk(ctx, keys, muts)

	if err != nil {
		return err
	}

	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	logger.Infof("exported sync committee assignments to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveAttestations(blocks map[uint64]map[string]*types.Block) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
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
			logger.Infof("processing slot %v", slot)
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

	for attestedSlot, inclusions := range attestationsBySlot {
		mut := gcp_bigtable.NewMutation()
		for validator, inclusionSlot := range inclusions {
			mut.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", validator), gcp_bigtable.Timestamp((max_block_number-inclusionSlot)*1000), []byte{})
		}
		err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(attestedSlot/utils.Config.Chain.Config.SlotsPerEpoch), reversedPaddedSlot(attestedSlot)), mut)

		if err != nil {
			return err
		}
	}
	logger.Infof("exported attestations to bigtable in %v", time.Since(start))
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

	for _, slot := range slots {
		for _, b := range blocks[slot] {

			if len(b.BlockRoot) != 32 { // skip dummy blocks
				continue
			}
			mut := gcp_bigtable.NewMutation()
			mut.Set(PROPOSALS_FAMILY, fmt.Sprintf("%d", b.Proposer), gcp_bigtable.Timestamp((max_block_number-b.Slot)*1000), []byte{})
			err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(b.Slot/utils.Config.Chain.Config.SlotsPerEpoch), reversedPaddedSlot(b.Slot)), mut)
			if err != nil {
				return err
			}
		}
	}
	logger.Infof("exported proposals to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveSyncComitteeDuties(blocks map[uint64]map[string]*types.Block) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
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
	for slot, validators := range dutiesBySlot {
		mut := gcp_bigtable.NewMutation()
		for validator, participated := range validators {
			if participated {
				mut.Set(SYNC_COMMITTEES_FAMILY, fmt.Sprintf("%d", validator), gcp_bigtable.Timestamp((max_block_number-slot)*1000), []byte{})
			} else {
				mut.Set(SYNC_COMMITTEES_FAMILY, fmt.Sprintf("%d", validator), gcp_bigtable.Timestamp(0), []byte{})
			}
		}
		err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(slot/utils.Config.Chain.Config.SlotsPerEpoch), reversedPaddedSlot(slot)), mut)

		if err != nil {
			return err
		}
	}
	logger.Infof("exported sync committee duties to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {

	valLen := len(validators)
	getAllThreshold := 1000
	validatorMap := make(map[uint64]bool, valLen)
	for _, validatorIndex := range validators {
		validatorMap[validatorIndex] = true
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	ranges := bigtable.getEpochRanges(startEpoch, endEpoch)
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
				Epoch:            max_epoch - epoch,
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

func (bigtable *Bigtable) GetValidatorAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error) {
	valLen := len(validators)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	ranges := bigtable.getSlotRanges(startEpoch, endEpoch)
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
		gcp_bigtable.LatestNFilter(1),
	)

	if len(columnFilters) == 1 { // special case to retrieve data for one validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
			columnFilters[0],
			gcp_bigtable.LatestNFilter(1),
		)
	}
	if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
			gcp_bigtable.LatestNFilter(1),
		)
	}
	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		keySplit := strings.Split(r.Key(), ":")

		attesterSlot, err := strconv.ParseUint(keySplit[4], 10, 64)
		if err != nil {
			logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
			return false
		}
		attesterSlot = max_block_number - attesterSlot
		for _, ri := range r[ATTESTATIONS_FAMILY] {
			inclusionSlot := max_block_number - uint64(ri.Timestamp)/1000

			status := uint64(1)
			if inclusionSlot == max_block_number {
				inclusionSlot = 0
				status = 0
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, ATTESTATIONS_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			if res[validator] == nil {
				res[validator] = make([]*types.ValidatorAttestation, 0)
			}

			if len(res[validator]) > 0 && res[validator][len(res[validator])-1].AttesterSlot == attesterSlot {
				res[validator][len(res[validator])-1].InclusionSlot = inclusionSlot
				res[validator][len(res[validator])-1].Status = status
				res[validator][len(res[validator])-1].Delay = int64(inclusionSlot - attesterSlot)
			} else {
				res[validator] = append(res[validator], &types.ValidatorAttestation{
					Index:          validator,
					Epoch:          attesterSlot / utils.Config.Chain.Config.SlotsPerEpoch,
					AttesterSlot:   attesterSlot,
					CommitteeIndex: 0,
					Status:         status,
					InclusionSlot:  inclusionSlot,
					Delay:          int64(inclusionSlot) - int64(attesterSlot) - 1,
				})
			}

		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorMissedAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]bool, error) {
	valLen := len(validators)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*20))
	defer cancel()

	ranges := bigtable.getSlotRanges(startEpoch, endEpoch)

	res := make(map[uint64]map[uint64]bool)

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
	}
	if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(ATTESTATIONS_FAMILY),
			gcp_bigtable.LatestNFilter(1),
		)
	}

	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		keySplit := strings.Split(r.Key(), ":")

		attesterSlot, err := strconv.ParseUint(keySplit[4], 10, 64)
		if err != nil {
			logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
			return false
		}
		attesterSlot = max_block_number - attesterSlot

		for _, ri := range r[ATTESTATIONS_FAMILY] {
			inclusionSlot := max_block_number - uint64(ri.Timestamp)/1000

			status := uint64(1)
			if inclusionSlot == max_block_number {
				inclusionSlot = 0
				status = 0
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, ATTESTATIONS_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			if status == 0 {
				if res[validator] == nil {
					res[validator] = make(map[uint64]bool, 0)
				}
				res[validator][attesterSlot] = true
			} else if res[validator] != nil && res[validator][attesterSlot] {
				delete(res[validator], attesterSlot)
			}
		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorSyncDutiesHistoryOrdered(validatorIndex uint64, startEpoch uint64, endEpoch uint64, reverseOrdering bool) ([]*types.ValidatorSyncParticipation, error) {
	res, err := bigtable.GetValidatorSyncDutiesHistory([]uint64{validatorIndex}, startEpoch, endEpoch)
	if err != nil {
		return nil, err
	}
	if reverseOrdering {
		utils.ReverseSlice(res[validatorIndex])
	}
	return res[validatorIndex], nil
}

func (bigtable *Bigtable) GetValidatorSyncDutiesHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorSyncParticipation, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	ranges := bigtable.getSlotRanges(startEpoch, endEpoch)
	res := make(map[uint64][]*types.ValidatorSyncParticipation, len(validators))

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
			slot = max_block_number - slot
			inclusionSlot := max_block_number - uint64(ri.Timestamp)/1000

			status := uint64(1) // 1: participated
			if inclusionSlot == max_block_number {
				inclusionSlot = 0
				status = 0 // 0: missed
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, SYNC_COMMITTEES_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

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

		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
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
		missed := len(attestations)
		if missed > 0 {
			res[validator] = &types.ValidatorMissedAttestationsStatistic{
				Index:              validator,
				MissedAttestations: uint64(missed),
			}
		}
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorSyncDutiesStatistics(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*types.ValidatorSyncDutiesStatistic, error) {
	data, err := bigtable.GetValidatorSyncDutiesHistory(validators, startEpoch, endEpoch)

	if err != nil {
		return nil, err
	}

	res := make(map[uint64]*types.ValidatorSyncDutiesStatistic)

	for validator, duties := range data {
		if res[validator] == nil && len(duties) > 0 {
			res[validator] = &types.ValidatorSyncDutiesStatistic{
				Index: validator,
			}
		}

		for _, duty := range duties {
			if duty.Status == 0 {
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
	data, err := bigtable.GetValidatorAttestationHistory(validators, epoch-100, epoch)

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

func (bigtable *Bigtable) GetValidatorBalanceStatistics(startEpoch, endEpoch uint64) (map[uint64]*types.ValidatorBalanceStatistic, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()

	ranges := bigtable.getEpochRanges(startEpoch, endEpoch)
	res := make(map[uint64]*types.ValidatorBalanceStatistic)

	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		keySplit := strings.Split(r.Key(), ":")

		epoch, err := strconv.ParseUint(keySplit[3], 10, 64)
		if err != nil {
			logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
			return false
		}
		epoch = max_epoch - epoch
		logger.Infof("retrieved %v balances entries for epoch %v", len(r[VALIDATOR_BALANCES_FAMILY]), epoch)

		for _, ri := range r[VALIDATOR_BALANCES_FAMILY] {
			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, VALIDATOR_BALANCES_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}
			balances := ri.Value

			balanceBytes := balances[0:8]
			effectiveBalanceBytes := balances[8:16]
			balance := binary.LittleEndian.Uint64(balanceBytes)
			effectiveBalance := binary.LittleEndian.Uint64(effectiveBalanceBytes)

			if res[validator] == nil {
				res[validator] = &types.ValidatorBalanceStatistic{
					Index:                 validator,
					MinEffectiveBalance:   effectiveBalance,
					MaxEffectiveBalance:   0,
					MinBalance:            balance,
					MaxBalance:            0,
					StartEffectiveBalance: 0,
					EndEffectiveBalance:   0,
					StartBalance:          0,
					EndBalance:            0,
				}
			}

			// logger.Info(epoch, startEpoch)
			if epoch == startEpoch {
				res[validator].StartBalance = balance
				res[validator].StartEffectiveBalance = effectiveBalance
			}

			if epoch == endEpoch {
				res[validator].EndBalance = balance
				res[validator].EndEffectiveBalance = effectiveBalance
			}

			if balance > res[validator].MaxBalance {
				res[validator].MaxBalance = balance
			}
			if balance < res[validator].MinBalance {
				res[validator].MinBalance = balance
			}

			if balance > res[validator].MaxEffectiveBalance {
				res[validator].MaxEffectiveBalance = balance
			}
			if balance < res[validator].MinEffectiveBalance {
				res[validator].MinEffectiveBalance = balance
			}
		}

		return true
	}, gcp_bigtable.RowFilter(gcp_bigtable.FamilyFilter(VALIDATOR_BALANCES_FAMILY)))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorProposalHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorProposal, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	ranges := bigtable.getSlotRanges(startEpoch, endEpoch)
	res := make(map[uint64][]*types.ValidatorProposal, len(validators))

	columnFilters := make([]gcp_bigtable.Filter, 0, len(validators))
	for _, validator := range validators {
		columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
	}

	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.FamilyFilter(PROPOSALS_FAMILY),
		gcp_bigtable.InterleaveFilters(columnFilters...),
		gcp_bigtable.LatestNFilter(1),
	)

	if len(columnFilters) == 1 { // special case to retrieve data for one validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(PROPOSALS_FAMILY),
			columnFilters[0],
			gcp_bigtable.LatestNFilter(1),
		)
	}
	if len(columnFilters) == 0 { // special case to retrieve data for all validators
		filter = gcp_bigtable.ChainFilters(
			gcp_bigtable.FamilyFilter(PROPOSALS_FAMILY),
			gcp_bigtable.LatestNFilter(1),
		)
	}

	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		for _, ri := range r[PROPOSALS_FAMILY] {
			keySplit := strings.Split(r.Key(), ":")

			proposalSlot, err := strconv.ParseUint(keySplit[4], 10, 64)
			if err != nil {
				logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
				return false
			}
			proposalSlot = max_block_number - proposalSlot
			inclusionSlot := max_block_number - uint64(r[PROPOSALS_FAMILY][0].Timestamp)/1000

			status := uint64(1)
			if inclusionSlot == max_block_number {
				inclusionSlot = 0
				status = 2
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, PROPOSALS_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

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

		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) SaveValidatorIncomeDetails(epoch uint64, rewards map[uint64]*itypes.ValidatorEpochIncome) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()
	ts := gcp_bigtable.Timestamp(utils.EpochToTime(epoch).UnixMicro())

	total := &itypes.ValidatorEpochIncome{}

	mut := gcp_bigtable.NewMutation()

	muts := 0
	for i, rewardDetails := range rewards {
		muts++

		data, err := proto.Marshal(rewardDetails)

		if err != nil {
			return err
		}

		mut.Set(INCOME_DETAILS_COLUMN_FAMILY, fmt.Sprintf("%d", i), ts, data)

		if muts%100000 == 0 {
			err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(epoch)), mut)

			if err != nil {
				return err
			}
			mut = gcp_bigtable.NewMutation()
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

	sum, err := proto.Marshal(total)
	if err != nil {
		return err
	}

	mut.Set(STATS_COLUMN_FAMILY, SUM_COLUMN, ts, sum)

	err = bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(epoch)), mut)
	if err != nil {
		return err
	}

	logger.Infof("exported validator income details for epoch %v to bigtable in %v", epoch, time.Since(start))
	return nil
}

func (bigtable *Bigtable) GetEpochIncomeHistoryDescending(startEpoch uint64, endEpoch uint64) (*itypes.ValidatorEpochIncome, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	ranges := bigtable.getEpochRanges(startEpoch, endEpoch)
	family := gcp_bigtable.FamilyFilter(STATS_COLUMN_FAMILY)
	columnFilter := gcp_bigtable.ColumnFilter(SUM_COLUMN)
	filter := gcp_bigtable.RowFilter(gcp_bigtable.ChainFilters(family, columnFilter))

	res := itypes.ValidatorEpochIncome{}

	err := bigtable.tableBeaconchain.ReadRows(ctx, ranges, func(r gcp_bigtable.Row) bool {
		if len(r[STATS_COLUMN_FAMILY]) == 0 {
			return false
		}
		err := proto.Unmarshal(r[STATS_COLUMN_FAMILY][0].Value, &res)
		if err != nil {
			logger.Errorf("error decoding income data for row %v: %v", r.Key(), err)
			return false
		}
		return true
	}, filter)

	if err != nil {
		return nil, fmt.Errorf("error reading income statistics from bigtable for epoch: %v err: %w", startEpoch, err)
	}

	return &res, nil
}

func (bigtable *Bigtable) GetEpochIncomeHistory(epoch uint64) (*itypes.ValidatorEpochIncome, error) {

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	key := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(epoch))

	family := gcp_bigtable.FamilyFilter(STATS_COLUMN_FAMILY)
	columnFilter := gcp_bigtable.ColumnFilter(SUM_COLUMN)
	filter := gcp_bigtable.RowFilter(gcp_bigtable.ChainFilters(family, columnFilter))

	row, err := bigtable.tableBeaconchain.ReadRow(ctx, key, filter)
	if err != nil {
		return nil, fmt.Errorf("error reading income statistics from bigtable for epoch: %v err: %w", epoch, err)
	}

	if row != nil {
		res := itypes.ValidatorEpochIncome{}
		err := proto.Unmarshal(row[STATS_COLUMN_FAMILY][0].Value, &res)
		if err != nil {
			return nil, fmt.Errorf("error decoding income data for row %v: %w", row.Key(), err)
		}
		return &res, nil
	}

	// if there is no result we have to calculate the sum
	income, err := bigtable.GetValidatorIncomeDetailsHistory([]uint64{}, epoch, 1)
	if err != nil {
		logger.WithError(err).Error("error getting validator income history")
	}

	total := &itypes.ValidatorEpochIncome{}

	for _, epochs := range income {
		for _, details := range epochs {
			total.AttestationHeadReward += details.AttestationHeadReward
			total.AttestationSourceReward += details.AttestationSourceReward
			total.AttestationSourcePenalty += details.AttestationSourcePenalty
			total.AttestationTargetReward += details.AttestationTargetReward
			total.AttestationTargetPenalty += details.AttestationTargetPenalty
			total.FinalityDelayPenalty += details.FinalityDelayPenalty
			total.ProposerSlashingInclusionReward += details.ProposerSlashingInclusionReward
			total.ProposerAttestationInclusionReward += details.ProposerAttestationInclusionReward
			total.ProposerSyncInclusionReward += details.ProposerSyncInclusionReward
			total.SyncCommitteeReward += details.SyncCommitteeReward
			total.SyncCommitteePenalty += details.SyncCommitteePenalty
			total.SlashingReward += details.SlashingReward
			total.SlashingPenalty += details.SlashingPenalty
			total.TxFeeRewardWei = utils.AddBigInts(total.TxFeeRewardWei, details.TxFeeRewardWei)
		}
	}

	return total, nil
}

// GetValidatorIncomeDetailsHistory returns the validator income details
// startEpoch & endEpoch are inclusive
func (bigtable *Bigtable) GetValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*itypes.ValidatorEpochIncome, error) {
	if startEpoch > endEpoch {
		startEpoch = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*180)
	defer cancel()

	ranges := bigtable.getEpochRanges(startEpoch, endEpoch)
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

			res[validator][max_epoch-epoch] = incomeDetails
		}
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetValidatorIncomeDetailsHistory returns the validator income details
// startEpoch & endEpoch are inclusive
func (bigtable *Bigtable) GetAggregatedValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*itypes.ValidatorEpochIncome, error) {
	if startEpoch > endEpoch {
		startEpoch = 0
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()

	ranges := bigtable.getEpochRanges(startEpoch, endEpoch)

	// logger.Infof("range: %v to %v", rangeStart, rangeEnd)
	incomeStats := make(map[uint64]*itypes.ValidatorEpochIncome, len(validators))

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

	if len(columnFilters) == 1 { // special case to retrieve data for one validators
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
		epoch = max_epoch - epoch
		start := time.Now()

		for _, ri := range r[INCOME_DETAILS_COLUMN_FAMILY] {
			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, INCOME_DETAILS_COLUMN_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			rewardDetails := &itypes.ValidatorEpochIncome{}
			err = proto.Unmarshal(ri.Value, rewardDetails)
			if err != nil {
				logger.Errorf("error decoding validator income data for row %v: %v", r.Key(), err)
				return false
			}

			if incomeStats[validator] == nil {
				incomeStats[validator] = &itypes.ValidatorEpochIncome{}
			}

			incomeStats[validator].AttestationHeadReward += rewardDetails.AttestationHeadReward
			incomeStats[validator].AttestationSourceReward += rewardDetails.AttestationSourceReward
			incomeStats[validator].AttestationSourcePenalty += rewardDetails.AttestationSourcePenalty
			incomeStats[validator].AttestationTargetReward += rewardDetails.AttestationTargetReward
			incomeStats[validator].AttestationTargetPenalty += rewardDetails.AttestationTargetPenalty
			incomeStats[validator].FinalityDelayPenalty += rewardDetails.FinalityDelayPenalty
			incomeStats[validator].ProposerSlashingInclusionReward += rewardDetails.ProposerSlashingInclusionReward
			incomeStats[validator].ProposerAttestationInclusionReward += rewardDetails.ProposerAttestationInclusionReward
			incomeStats[validator].ProposerSyncInclusionReward += rewardDetails.ProposerSyncInclusionReward
			incomeStats[validator].SyncCommitteeReward += rewardDetails.SyncCommitteeReward
			incomeStats[validator].SyncCommitteePenalty += rewardDetails.SyncCommitteePenalty
			incomeStats[validator].SlashingReward += rewardDetails.SlashingReward
			incomeStats[validator].SlashingPenalty += rewardDetails.SlashingPenalty
			incomeStats[validator].TxFeeRewardWei = utils.AddBigInts(incomeStats[validator].TxFeeRewardWei, rewardDetails.TxFeeRewardWei)
		}

		logger.Infof("processed income data for epoch %v in %v", epoch, time.Since(start))
		return true
	}, gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return incomeStats, nil
}

// Deletes all block data from bigtable
func (bigtable *Bigtable) DeleteEpoch(epoch uint64) error {

	// First receive all keys that were written by this block (entities & indices)
	keys := make([]string, 0, 33)
	startSlot := epoch * utils.Config.Chain.Config.SlotsPerEpoch
	endSlot := (epoch+1)*utils.Config.Chain.Config.SlotsPerEpoch - 1

	logger.Infof("deleting epoch %v (slot %v to %v)", epoch, startSlot, endSlot)
	for slot := startSlot; slot <= endSlot; slot++ {
		keys = append(keys, fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(slot/utils.Config.Chain.Config.SlotsPerEpoch), reversedPaddedSlot(slot)))
	}
	keys = append(keys, fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(epoch)))

	// for _, k := range keys {
	// 	logger.Info(k)
	// }

	// Delete all of those keys
	mutsDelete := &types.BulkMutations{
		Keys: make([]string, 0, len(keys)),
		Muts: make([]*gcp_bigtable.Mutation, 0, len(keys)),
	}
	for _, key := range keys {
		mutDelete := gcp_bigtable.NewMutation()
		mutDelete.DeleteRow()
		mutsDelete.Keys = append(mutsDelete.Keys, key)
		mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
	}

	err := bigtable.WriteBulk(mutsDelete, bigtable.tableBeaconchain)
	if err != nil {
		return err
	}

	return nil
}

func (bigtable *Bigtable) getSlotRanges(startEpoch uint64, endEpoch uint64) gcp_bigtable.RowRangeList {

	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	ranges := gcp_bigtable.RowRangeList{}
	if startEpoch == 0 { // special case when the 0 epoch is included
		rangeEnd := fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(0), "\x00")
		rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(0))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))

		// epochs are sorted descending, so start with the larges epoch and end with the smallest
		// add \x00 to make the range inclusive
		if startEpoch < endEpoch {
			rangeEnd = fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(startEpoch+1), "\x00")
			rangeStart = fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(endEpoch))
			ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
		}
	} else if startEpoch == endEpoch { // special case, only retrieve data for one epoch
		rangeEnd := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch-1))
		rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
	} else {
		// epochs are sorted descending, so start with the larges epoch and end with the smallest
		// add \x00 to make the range inclusive
		rangeEnd := fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(startEpoch), "\x00")
		rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(endEpoch))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
	}
	return ranges
}

func (bigtable *Bigtable) getEpochRanges(startEpoch uint64, endEpoch uint64) gcp_bigtable.RowRangeList {

	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	ranges := gcp_bigtable.RowRangeList{}
	if startEpoch == 0 { // special case when the 0 epoch is included
		rangeEnd := fmt.Sprintf("%s:e:b:%s%s", bigtable.chainId, reversedPaddedEpoch(0), "\x00")
		rangeStart := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(0))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))

		// epochs are sorted descending, so start with the largest epoch and end with the smallest
		// add \x00 to make the range inclusive
		if startEpoch < endEpoch {
			rangeEnd = fmt.Sprintf("%s:e:b:%s%s", bigtable.chainId, reversedPaddedEpoch(startEpoch+1), "\x00")
			rangeStart = fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(endEpoch))
			ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
		}
	} else {
		// epochs are sorted descending, so start with the largest epoch and end with the smallest
		// add \x00 to make the range inclusive
		rangeEnd := fmt.Sprintf("%s:e:b:%s%s", bigtable.chainId, reversedPaddedEpoch(startEpoch), "\x00")
		rangeStart := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(endEpoch))
		ranges = append(ranges, gcp_bigtable.NewRange(rangeStart, rangeEnd))
	}
	return ranges
}

func GetCurrentDayClIncome(validator_indices []uint64) (map[uint64]int64, map[uint64]int64, error) {
	dayIncome := make(map[uint64]int64)
	dayProposerIncome := make(map[uint64]int64)
	lastDay, err := GetLastExportedStatisticDay()
	if err != nil {
		return dayIncome, dayProposerIncome, err
	}

	currentDay := uint64(lastDay + 1)
	startEpoch := currentDay * utils.EpochsPerDay()
	endEpoch := startEpoch + utils.EpochsPerDay() - 1
	income, err := BigtableClient.GetValidatorIncomeDetailsHistory(validator_indices, startEpoch, endEpoch)
	if err != nil {
		return dayIncome, dayProposerIncome, err
	}

	// agregate all epoch income data to total day income for each validator
	for validatorIndex, validatorIncome := range income {
		if len(validatorIncome) == 0 {
			continue
		}
		for _, validatorEpochIncome := range validatorIncome {
			dayIncome[validatorIndex] += validatorEpochIncome.TotalClRewards()
			dayProposerIncome[validatorIndex] += int64(validatorEpochIncome.ProposerAttestationInclusionReward) + int64(validatorEpochIncome.ProposerSlashingInclusionReward) + int64(validatorEpochIncome.ProposerSyncInclusionReward)
		}
	}

	return dayIncome, dayProposerIncome, nil
}

func GetCurrentDayProposerIncomeTotal(validator_indices []uint64) (int64, error) {
	_, proposerIncome, err := GetCurrentDayClIncome(validator_indices)

	if err != nil {
		return 0, err
	}

	proposerTotal := int64(0)

	for _, i := range proposerIncome {
		proposerTotal += i
	}

	return proposerTotal, nil
}

func reversePaddedUserID(userID uint64) string {
	return fmt.Sprintf("%09d", ^uint64(0)-userID)
}

func reversedPaddedEpoch(epoch uint64) string {
	return fmt.Sprintf("%09d", max_block_number-epoch)
}

func reversedPaddedSlot(slot uint64) string {
	return fmt.Sprintf("%09d", max_block_number-slot)
}
