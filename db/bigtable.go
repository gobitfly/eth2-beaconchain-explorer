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
	"google.golang.org/api/option"
)

var BigtableClient *Bigtable

const (
	DEFAULT_FAMILY            = "f"
	VALIDATOR_BALANCES_FAMILY = "vb"
	ATTESTATIONS_FAMILY       = "at"
	PROPOSALS_FAMILY          = "pr"
	SYNC_COMMITTEES_FAMILY    = "sc"

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

	chainId string
}

func InitBigtable(project, instance, chainId string) (*Bigtable, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	poolSize := 50
	btClient, err := gcp_bigtable.NewClient(ctx, project, instance, option.WithGRPCConnectionPool(poolSize))
	// btClient, err := gcp_bigtable.NewClient(context.Background(), project, instance)

	if err != nil {
		return nil, err
	}

	bt := &Bigtable{
		client:               btClient,
		tableData:            btClient.Open("data"),
		tableBlocks:          btClient.Open("blocks"),
		tableMetadataUpdates: btClient.Open("metadata_updates"),
		tableMetadata:        btClient.Open("metadata"),
		tableBeaconchain:     btClient.Open("beaconchain"),
		chainId:              chainId,
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
			validatorsPerSlot[attesterslot] = make([]uint64, 0, len(assignments)/32)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
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

					if utils.BitAtVector(b.SyncAggregate.SyncCommitteeBits, i) {

						if dutiesBySlot[b.Slot] == nil {
							dutiesBySlot[b.Slot] = make(map[uint64]bool)
						}
						dutiesBySlot[b.Slot][valIndex] = true
					}
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
		for validator := range validators {
			mut.Set(SYNC_COMMITTEES_FAMILY, fmt.Sprintf("%d", validator), gcp_bigtable.Timestamp((max_block_number-slot)*1000), []byte{})
		}
		err := bigtable.tableBeaconchain.Apply(ctx, fmt.Sprintf("%s:e:%s:s:%s", bigtable.chainId, reversedPaddedEpoch(slot/utils.Config.Chain.Config.SlotsPerEpoch), reversedPaddedSlot(slot)), mut)

		if err != nil {
			return err
		}
	}
	logger.Infof("exported sync committee duties to bigtable in %v", time.Since(start))
	return nil
}

func (bigtable *Bigtable) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, limit int64) (map[uint64][]*types.ValidatorBalance, error) {

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	rangeStart := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(startEpoch))
	rangeEnd := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(startEpoch-uint64(limit)))
	res := make(map[uint64][]*types.ValidatorBalance, len(validators))

	if len(validators) == 0 {
		return res, nil
	}

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

	err := bigtable.tableBeaconchain.ReadRows(ctx, gcp_bigtable.NewRange(rangeStart, rangeEnd), func(r gcp_bigtable.Row) bool {
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
	valLen := len(validators)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch))
	rangeEnd := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch-uint64(limit)))
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
	err := bigtable.tableBeaconchain.ReadRows(ctx, gcp_bigtable.NewRange(rangeStart, rangeEnd), func(r gcp_bigtable.Row) bool {
		for _, ri := range r[ATTESTATIONS_FAMILY] {
			keySplit := strings.Split(r.Key(), ":")

			attesterSlot, err := strconv.ParseUint(keySplit[4], 10, 64)
			if err != nil {
				logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
				return false
			}
			attesterSlot = max_block_number - attesterSlot
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
				res[validator] = make([]*types.ValidatorAttestation, 0, limit)
			}

			if len(res[validator]) > 1 && res[validator][len(res[validator])-1].AttesterSlot == attesterSlot {
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

func (bigtable *Bigtable) GetValidatorSyncDutiesHistory(validators []uint64, startEpoch uint64, limit int64) (map[uint64][]*types.ValidatorSyncParticipation, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch))
	rangeEnd := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch-uint64(limit)))
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

	err := bigtable.tableBeaconchain.ReadRows(ctx, gcp_bigtable.NewRange(rangeStart, rangeEnd), func(r gcp_bigtable.Row) bool {
		for _, ri := range r[SYNC_COMMITTEES_FAMILY] {
			keySplit := strings.Split(r.Key(), ":")

			slot, err := strconv.ParseUint(keySplit[4], 10, 64)
			if err != nil {
				logger.Errorf("error parsing slot from row key %v: %v", r.Key(), err)
				return false
			}
			slot = max_block_number - slot
			inclusionSlot := max_block_number - uint64(r[SYNC_COMMITTEES_FAMILY][0].Timestamp)/1000

			status := uint64(1)
			if inclusionSlot == max_block_number {
				inclusionSlot = 0
				status = 0
			}

			validator, err := strconv.ParseUint(strings.TrimPrefix(ri.Column, SYNC_COMMITTEES_FAMILY+":"), 10, 64)
			if err != nil {
				logger.Errorf("error parsing validator from column key %v: %v", ri.Column, err)
				return false
			}

			if res[validator] == nil {
				res[validator] = make([]*types.ValidatorSyncParticipation, 0, limit)
			}

			if len(res[validator]) > 1 && res[validator][len(res[validator])-1].Slot == slot {
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

func (bigtable *Bigtable) GetValidatorMissedAttestationsCount(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*types.ValidatorMissedAttestationsStatistic, error) {

	res := make(map[uint64]*types.ValidatorMissedAttestationsStatistic)

	for i := startEpoch; i >= startEpoch-endEpoch; i-- {
		data, err := bigtable.GetValidatorAttestationHistory(validators, i, 1)

		if err != nil {
			return nil, err
		}

		logger.Infof("retrieved attestation history for epoch %v", i)

		for validator, attestations := range data {
			for _, attestation := range attestations {
				if attestation.Status == 0 {
					if res[validator] == nil {
						res[validator] = &types.ValidatorMissedAttestationsStatistic{
							Index: validator,
						}
					}
					res[validator].MissedAttestations++
				}
			}
		}
	}

	return res, nil
}

func (bigtable *Bigtable) GetValidatorSyncDutiesStatistics(validators []uint64, startEpoch uint64, limit int64) (map[uint64]*types.ValidatorSyncDutiesStatistic, error) {
	data, err := bigtable.GetValidatorSyncDutiesHistory(validators, startEpoch, limit)

	if err != nil {
		return nil, err
	}

	res := make(map[uint64]*types.ValidatorSyncDutiesStatistic)

	for validator, duties := range data {
		for _, duty := range duties {
			if res[validator] == nil {
				res[validator] = &types.ValidatorSyncDutiesStatistic{
					Index: validator,
				}
			}

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
	data, err := bigtable.GetValidatorAttestationHistory(validators, epoch, 100)

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

	// logger.Info(startEpoch, endEpoch)
	rangeStart := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(endEpoch)) // Reverse as keys are sorted in descending order
	rangeEnd := fmt.Sprintf("%s:e:b:%s", bigtable.chainId, reversedPaddedEpoch(startEpoch-1))

	res := make(map[uint64]*types.ValidatorBalanceStatistic)
	err := bigtable.tableBeaconchain.ReadRows(ctx, gcp_bigtable.NewRange(rangeStart, rangeEnd), func(r gcp_bigtable.Row) bool {
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

func (bigtable *Bigtable) GetValidatorProposalHistory(validators []uint64, startEpoch uint64, limit int64) (map[uint64][]*types.ValidatorProposal, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	rangeStart := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch))
	rangeEnd := fmt.Sprintf("%s:e:%s:s:", bigtable.chainId, reversedPaddedEpoch(startEpoch-uint64(limit)))
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

	err := bigtable.tableBeaconchain.ReadRows(ctx, gcp_bigtable.NewRange(rangeStart, rangeEnd), func(r gcp_bigtable.Row) bool {
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
				res[validator] = make([]*types.ValidatorProposal, 0, limit)
			}

			if len(res[validator]) > 1 && res[validator][len(res[validator])-1].Slot == proposalSlot {
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
