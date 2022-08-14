package db

import (
	"context"
	"encoding/binary"
	"eth2-exporter/types"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"google.golang.org/api/option"
)

var BigtableClient *Bigtable

const (
	DEFAULT_FAMILY            = "f"
	VALIDATOR_BALANCES_FAMILY = "vb"
	ATTESTATIONS_FAMILY       = "at"
	PROPOSALS_FAMILY          = "pr"

	max_block_number = 1000000000
	max_epoch        = 1000000000
)

type Bigtable struct {
	client           *gcp_bigtable.Client
	tableData        *gcp_bigtable.Table
	tableBlocks      *gcp_bigtable.Table
	tableBeaconchain *gcp_bigtable.Table
	chainId          string
}

func NewBigtable(project, instance, chainId string) (*Bigtable, error) {
	poolSize := 50
	btClient, err := gcp_bigtable.NewClient(context.Background(), project, instance, option.WithGRPCConnectionPool(poolSize))
	// btClient, err := gcp_bigtable.NewClient(context.Background(), project, instance)

	if err != nil {
		return nil, err
	}

	bt := &Bigtable{
		client:           btClient,
		tableData:        btClient.Open("data"),
		tableBlocks:      btClient.Open("blocks"),
		tableBeaconchain: btClient.Open("beaconchain"),
		chainId:          chainId,
	}

	BigtableClient = bt
	return bt, nil
}

func (bigtable *Bigtable) Close() {
	bigtable.client.Close()
}

func (bigtable *Bigtable) SaveValidatorBalances(epoch uint64, validators []*types.Validator) error {
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
			err := bigtable.tableBeaconchain.Apply(context.Background(), fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(epoch)), mut)

			if err != nil {
				return err
			}
			mut = gcp_bigtable.NewMutation()
		}
	}
	err := bigtable.tableBeaconchain.Apply(context.Background(), fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(epoch)), mut)

	if err != nil {
		return err
	}

	logger.Infof("exported validator balances to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveAttestationAssignments(epoch uint64, assignments map[string]uint64) error {
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
			validatorsPerSlot[attesterslot] = make([]uint64, 0, len(assignments)/32)
		}
		validatorsPerSlot[attesterslot] = append(validatorsPerSlot[attesterslot], validator)
	}

	for slot, validators := range validatorsPerSlot {
		mut := gcp_bigtable.NewMutation()
		for _, validator := range validators {
			mut.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", validator), ts, []byte{})
		}
		err := bigtable.tableBeaconchain.Apply(context.Background(), fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(epoch), reversedPaddedSlot(slot)), mut)

		if err != nil {
			return err
		}
	}

	logger.Infof("exported attestation assignments to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveProposalAssignments(epoch uint64, assignments map[uint64]uint64) error {
	start := time.Now()
	ts := gcp_bigtable.Timestamp(0)

	for slot, validator := range assignments {
		mut := gcp_bigtable.NewMutation()
		mut.Set(PROPOSALS_FAMILY, fmt.Sprintf("%d", validator), ts, []byte{})
		err := bigtable.tableBeaconchain.Apply(context.Background(), fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(epoch), reversedPaddedSlot(slot)), mut)

		if err != nil {
			return err
		}
	}

	logger.Infof("exported proposal assignments to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveAttestations(blocks map[uint64]map[string]*types.Block) error {
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
			for _, a := range b.Attestations {
				for _, validator := range a.Attesters {
					inclusionSlot := slot
					attestedSlot := a.Data.Slot
					if attestationsBySlot[attestedSlot] == nil {
						attestationsBySlot[attestedSlot] = make(map[uint64]uint64)
					}
					attestationsBySlot[attestedSlot][validator] = inclusionSlot
				}
			}
		}
	}

	for attestedSlot, inclusions := range attestationsBySlot {
		mut := gcp_bigtable.NewMutation()
		for validator, inclusionSlot := range inclusions {
			mut.Set(ATTESTATIONS_FAMILY, fmt.Sprintf("%d", validator), gcp_bigtable.Timestamp((max_block_number-inclusionSlot)*1000), []byte{})
		}
		err := bigtable.tableBeaconchain.Apply(context.Background(), fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(attestedSlot/32), reversedPaddedSlot(attestedSlot)), mut)

		if err != nil {
			return err
		}
	}
	logger.Infof("exported attestations to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) SaveProposals(blocks map[uint64]map[string]*types.Block) error {
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
			err := bigtable.tableBeaconchain.Apply(context.Background(), fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(b.Slot/32), reversedPaddedSlot(b.Slot)), mut)
			if err != nil {
				return err
			}
		}
	}
	logger.Infof("exported proposals to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, limit int64) (map[uint64][]*types.ValidatorBalance, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	rangeStart := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(startEpoch))
	res := make(map[uint64][]*types.ValidatorBalance, len(validators))

	columnFilters := make([]gcp_bigtable.Filter, 0, len(validators))
	for _, validator := range validators {
		columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
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

	err := bigtable.tableBeaconchain.ReadRows(ctx, gcp_bigtable.NewRange(rangeStart, ""), func(r gcp_bigtable.Row) bool {
		for _, ri := range r[VALIDATOR_BALANCES_FAMILY] {
			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, VALIDATOR_BALANCES_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			keySplit := strings.Split(r.Key(), ":")

			epoch, err := strconv.ParseUint(keySplit[3], 10, 64)
			if err != nil {
				logger.Errorf("error parsing epoch from row key %v: %v", r.Key(), err)
				return false
			}

			balances := ri.Value

			balanceBytes := balances[0:8]
			effectiveBalanceBytes := balances[8:16]
			balance := binary.LittleEndian.Uint64(balanceBytes)
			effectiveBalance := binary.LittleEndian.Uint64(effectiveBalanceBytes)

			if res[validator] == nil {
				res[validator] = make([]*types.ValidatorBalance, 0, limit)
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
	}, gcp_bigtable.LimitRows(limit), gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorAttestationHistory(validators []uint64, startEpoch uint64, limit int64) (map[uint64][]*types.ValidatorAttestation, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch))
	res := make(map[uint64][]*types.ValidatorAttestation, len(validators))

	columnFilters := make([]gcp_bigtable.Filter, 0, len(validators))
	for _, validator := range validators {
		columnFilters = append(columnFilters, gcp_bigtable.ColumnFilter(fmt.Sprintf("%d", validator)))
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

	err := bigtable.tableBeaconchain.ReadRows(ctx, gcp_bigtable.NewRange(rangeStart, ""), func(r gcp_bigtable.Row) bool {
		for _, ri := range r[ATTESTATIONS_FAMILY] {
			keySplit := strings.Split(r.Key(), ":")

			attesterSlot, err := strconv.ParseUint(keySplit[4], 10, 64)
			if err != nil {
				logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
				return false
			}
			attesterSlot = max_block_number - attesterSlot
			inclusionSlot := max_block_number - uint64(r[ATTESTATIONS_FAMILY][0].Timestamp)/1000

			if inclusionSlot == max_block_number {
				inclusionSlot = 0
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, ATTESTATIONS_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			if res[validator] == nil {
				res[validator] = make([]*types.ValidatorAttestation, 0, limit)
			}

			if len(res[validator]) > 1 && res[validator][len(res[validator])-1].AttesterSlot == attesterSlot {
				res[validator][len(res[validator])-1].InclusionSlot = inclusionSlot
			} else {

				res[validator] = append(res[validator], &types.ValidatorAttestation{
					Index:          validator,
					Epoch:          attesterSlot / 32,
					AttesterSlot:   attesterSlot,
					CommitteeIndex: 0,
					Status:         0,
					InclusionSlot:  inclusionSlot,
					Delay:          0,
				})
			}

		}
		return true
	}, gcp_bigtable.LimitRows(limit), gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func reversedPaddedEpoch(epoch uint64) string {
	return fmt.Sprintf("%09d", max_block_number-epoch)
}

func reversedPaddedSlot(slot uint64) string {
	return fmt.Sprintf("%09d", max_block_number-slot)
}
