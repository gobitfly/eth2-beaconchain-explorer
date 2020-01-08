package db

import (
	"bytes"
	"database/sql"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus"
)

// DB is a pointer to the database
var DB *sqlx.DB
var logger = logrus.New().WithField("module", "db")

// GetLatestEpoch will return the latest epoch from the database
func GetLatestEpoch() (uint64, error) {
	var epoch uint64

	err := DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

	if err != nil {
		return 0, fmt.Errorf("error retrieving latest epoch from DB: %v", err)
	}

	return epoch, nil
}

// GetAllEpochs will return a collection of all of the epochs from the database
func GetAllEpochs() ([]uint64, error) {
	var epochs []uint64
	err := DB.Select(&epochs, "SELECT epoch FROM epochs ORDER BY epoch")

	if err != nil {
		return nil, fmt.Errorf("error retrieving all epochs from DB: %v", err)
	}

	return epochs, nil
}

// GetLastPendingAndProposedBlocks will return all proposed and pending blocks (ignores missed slots) from the database
func GetLastPendingAndProposedBlocks(startEpoch, endEpoch uint64) ([]*types.MinimalBlock, error) {
	var blocks []*types.MinimalBlock

	err := DB.Select(&blocks, "SELECT epoch, slot, blockroot FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND blockroot != '\x01' ORDER BY slot DESC", startEpoch, endEpoch)

	if err != nil {
		return nil, fmt.Errorf("error retrieving last blocks from DB: %v", err)
	}

	return blocks, nil
}

// GetBlocks will return all blocks for a range of epochs from the database
func GetBlocks(startEpoch, endEpoch uint64) ([]*types.MinimalBlock, error) {
	var blocks []*types.MinimalBlock

	err := DB.Select(&blocks, "SELECT epoch, slot, blockroot, parentroot FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND length(blockroot) = 32 ORDER BY slot DESC", startEpoch, endEpoch)

	if err != nil {
		return nil, fmt.Errorf("error retrieving blocks for epoch %v to %v from DB: %v", startEpoch, endEpoch, err)
	}

	return blocks, nil
}

// GetValidatorPublicKey will return the public key for a specific validator from the database
func GetValidatorPublicKey(index uint64) ([]byte, error) {
	var publicKey []byte
	err := DB.Get(&publicKey, "SELECT pubkey FROM validators WHERE validatorindex = $1", index)

	return publicKey, err
}

// GetValidatorIndex will return all of the validators for a public key from the database
func GetValidatorIndex(publicKey []byte) (uint64, error) {
	var index uint64
	err := DB.Get(&index, "SELECT validatorindex FROM validators WHERE pubkey = $1", publicKey)

	return index, err
}

// UpdateCanonicalBlocks will update the blocks for an epoch range in the database
func UpdateCanonicalBlocks(startEpoch, endEpoch uint64, orphanedBlocks [][]byte) error {
	if len(orphanedBlocks) == 0 {
		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE blocks SET status = 1 WHERE epoch >= $1 AND epoch <= $2 AND (status = '1' OR status = '3')", startEpoch, endEpoch)
	if err != nil {
		return err
	}

	for _, orphanedBlock := range orphanedBlocks {
		logger.Printf("Marking block %x as orphaned", orphanedBlock)
		_, err = tx.Exec("UPDATE blocks SET status = '3' WHERE blockroot = $1", orphanedBlock)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SaveAttestationPool will save the attestation pool into the database
func SaveAttestationPool(attestations []*types.Attestation) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("TRUNCATE attestationpool")
	if err != nil {
		return fmt.Errorf("error truncating attestationpool table: %v", err)
	}

	stmtAttestationPool, err := tx.Prepare(`INSERT INTO attestationpool (aggregationbits, custodybits, signature, slot, index, beaconblockroot, source_epoch, source_root, target_epoch, target_root)
 													VALUES    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (slot, index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtAttestationPool.Close()

	for _, attestation := range attestations {
		_, err := stmtAttestationPool.Exec(bitfield.Bitlist(attestation.AggregationBits).Bytes(), bitfield.Bitlist(attestation.CustodyBits).Bytes(), attestation.Signature, attestation.Data.Slot, attestation.Data.CommitteeIndex, attestation.Data.BeaconBlockRoot, attestation.Data.Source.Epoch, attestation.Data.Source.Root, attestation.Data.Target.Epoch, attestation.Data.Target.Root)
		if err != nil {
			return fmt.Errorf("error executing stmtAttestationPool: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db transaction: %v", err)
	}
	return nil
}

// SaveValidatorQueue will save the validator queue into the database
func SaveValidatorQueue(validators *types.ValidatorQueue, validatorIndices map[string]uint64) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("TRUNCATE validatorqueue_activation")
	if err != nil {
		return fmt.Errorf("error truncating validatorqueue_activation table: %v", err)
	}
	_, err = tx.Exec("TRUNCATE validatorqueue_exit")
	if err != nil {
		return fmt.Errorf("error truncating validatorqueue_exit table: %v", err)
	}

	stmtValidatorQueueActivation, err := tx.Prepare(`INSERT INTO validatorqueue_activation (index, publickey)
 													VALUES    ($1, $2) ON CONFLICT (index, publickey) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtValidatorQueueActivation.Close()

	stmtValidatorQueueExit, err := tx.Prepare(`INSERT INTO validatorqueue_exit (index, publickey)
 													VALUES    ($1, $2) ON CONFLICT (index, publickey) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtValidatorQueueExit.Close()

	for _, publickey := range validators.ActivationPublicKeys {
		_, err := stmtValidatorQueueActivation.Exec(validatorIndices[utils.FormatPublicKey(publickey)], publickey)
		if err != nil {
			return fmt.Errorf("error executing stmtValidatorQueueActivation: %v", err)
		}
	}
	for _, publickey := range validators.ExitPublicKeys {
		_, err := stmtValidatorQueueExit.Exec(validatorIndices[utils.FormatPublicKey(publickey)], publickey)
		if err != nil {
			return fmt.Errorf("error executing stmtValidatorQueueExit: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db transaction: %v", err)
	}
	return nil
}

// SaveEpoch will stave the epoch data into the database
func SaveEpoch(data *types.EpochData) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	err = saveBlocks(data.Epoch, data.Blocks, tx)
	if err != nil {
		return fmt.Errorf("error saving blocks to db: %v", err)
	}

	err = saveValidatorSet(data.Epoch, data.Validators, data.ValidatorIndices, tx)
	if err != nil {
		return fmt.Errorf("error saving validator set to db: %v", err)
	}

	err = saveValidatorProposalAssignments(data.Epoch, data.ValidatorAssignmentes.ProposerAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator assignments to db: %v", err)
	}

	err = saveValidatorAttestationAssignments(data.Epoch, data.ValidatorAssignmentes.AttestorAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator assignments to db: %v", err)
	}

	err = saveBeaconCommittees(data.Epoch, data.BeaconCommittees, tx)
	if err != nil {
		return fmt.Errorf("error saving beacon committees to db: %v", err)
	}

	err = saveValidatorBalances(data.Epoch, data.ValidatorBalances, tx)
	if err != nil {
		return fmt.Errorf("error saving validator balances to db: %v", err)
	}

	proposerSlashingsCount := 0
	attesterSlashingsCount := 0
	attestationsCount := 0
	depositCount := 0
	voluntaryExitCount := 0

	for _, slot := range data.Blocks {
		for _, b := range slot {
			proposerSlashingsCount += len(b.ProposerSlashings)
			attesterSlashingsCount += len(b.AttesterSlashings)
			attestationsCount += len(b.Attestations)
			depositCount += len(b.Deposits)
			voluntaryExitCount += len(b.VoluntaryExits)
		}
	}

	validatorBalanceSum := new(big.Int)
	for _, b := range data.ValidatorBalances {
		validatorBalanceSum = new(big.Int).Add(validatorBalanceSum, new(big.Int).SetUint64(b.Balance))
	}

	validatorBalanceAverage := new(big.Int).Div(validatorBalanceSum, new(big.Int).SetInt64(int64(len(data.Validators)))).Uint64()

	validatorsCount := 0
	for _, v := range data.Validators {
		if v.ExitEpoch > data.Epoch {
			validatorsCount++
		}
	}

	_, err = tx.Exec(`INSERT INTO epochs (
												epoch, 
												blockscount, 
												proposerslashingscount, 
												attesterslashingscount, 
												attestationscount, 
												depositscount, 
												voluntaryexitscount, 
												validatorscount, 
												averagevalidatorbalance, 
												finalized, 
                    							eligibleether, 
												globalparticipationrate, 
												votedether
												)
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT (epoch) DO UPDATE SET 
								  blockscount = excluded.blockscount, 
								  proposerslashingscount = excluded.proposerslashingscount,
								  attesterslashingscount = excluded.attesterslashingscount,
								  attestationscount = excluded.attestationscount,
								  depositscount = excluded.depositscount,
								  voluntaryexitscount = excluded.voluntaryexitscount,
								  validatorscount = excluded.validatorscount,
								  averagevalidatorbalance = excluded.averagevalidatorbalance,
								  finalized = excluded.finalized,
								  eligibleether = excluded.eligibleether,
								  globalparticipationrate = excluded.globalparticipationrate,
								  votedether = excluded.votedether`,
		data.Epoch,
		len(data.Blocks),
		proposerSlashingsCount,
		attesterSlashingsCount,
		attestationsCount,
		depositCount,
		voluntaryExitCount,
		validatorsCount,
		validatorBalanceAverage,
		data.EpochParticipationStats.Finalized,
		data.EpochParticipationStats.EligibleEther,
		data.EpochParticipationStats.GlobalParticipationRate,
		data.EpochParticipationStats.VotedEther)

	if err != nil {
		return fmt.Errorf("error executing save epoch statement: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db transaction: %v", err)
	}
	return nil
}

func saveValidatorSet(epoch uint64, validators []*types.Validator, validatorIndices map[string]uint64, tx *sql.Tx) error {

	stmtValidatorSet, err := tx.Prepare(`INSERT INTO validator_set (epoch, validatorindex, withdrawableepoch, withdrawalcredentials, effectivebalance, slashed, activationeligibilityepoch, activationepoch, exitepoch)
 													VALUES    ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (epoch, validatorindex) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtValidatorSet.Close()

	stmtValidators, err := tx.Prepare(`INSERT INTO validators (validatorindex, pubkey) VALUES ($1, $2) ON CONFLICT (validatorindex) DO UPDATE SET pubkey = excluded.pubkey`)
	if err != nil {
		return err
	}
	defer stmtValidators.Close()

	for _, v := range validators {
		if v.WithdrawableEpoch == 18446744073709551615 {
			v.WithdrawableEpoch = 9223372036854775807
		}
		if v.ExitEpoch == 18446744073709551615 {
			v.ExitEpoch = 9223372036854775807
		}
		if v.ActivationEligibilityEpoch == 18446744073709551615 {
			v.ActivationEligibilityEpoch = 9223372036854775807
		}
		if v.ActivationEpoch == 18446744073709551615 {
			v.ActivationEpoch = 9223372036854775807
		}
		_, err := stmtValidatorSet.Exec(epoch, validatorIndices[fmt.Sprintf("%x", v.PublicKey)], v.WithdrawableEpoch, v.WithdrawalCredentials, v.EffectiveBalance, v.Slashed, v.ActivationEligibilityEpoch, v.ActivationEpoch, v.ExitEpoch)
		if err != nil {
			return fmt.Errorf("error executing save validator set statement: %v", err)
		}
		_, err = stmtValidators.Exec(validatorIndices[fmt.Sprintf("%x", v.PublicKey)], v.PublicKey)
		if err != nil {
			return fmt.Errorf("error executing save validator statement: %v", err)
		}
	}

	return nil
}

func saveValidatorProposalAssignments(epoch uint64, assignments map[uint64]uint64, tx *sql.Tx) error {

	stmt, err := tx.Prepare(`INSERT INTO proposal_assignments (epoch, validatorindex, proposerslot, status)
 													VALUES    ($1, $2, $3, $4) ON CONFLICT (epoch, validatorindex, proposerslot) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for slot, validator := range assignments {
		_, err := stmt.Exec(epoch, validator, slot, 0)
		if err != nil {
			return fmt.Errorf("error executing save validator proposal assignment statement: %v", err)
		}
	}

	return nil
}

func saveValidatorAttestationAssignments(epoch uint64, assignments map[string]uint64, tx *sql.Tx) error {

	stmtAttestationAssignments, err := tx.Prepare(`INSERT INTO attestation_assignments (epoch, validatorindex, attesterslot, committeeindex, status)
 													VALUES    ($1, $2, $3, $4, $5) ON CONFLICT (epoch, validatorindex, attesterslot, committeeindex) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtAttestationAssignments.Close()

	for key, validator := range assignments {
		keySplit := strings.Split(key, "-")

		_, err := stmtAttestationAssignments.Exec(epoch, validator, keySplit[0], keySplit[1], 0)
		if err != nil {
			return fmt.Errorf("error executing save validator attestation assignment statement: %v", err)
		}
	}

	return nil
}

func saveBeaconCommittees(epoch uint64, committeesMap map[uint64][]*types.BeaconCommitteItem, tx *sql.Tx) error {

	stmt, err := tx.Prepare(`INSERT INTO beacon_committees (epoch, slot, slotindex, indices)
 													VALUES    ($1, $2, $3, $4) ON CONFLICT (epoch, slot, slotindex) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for slot, comittees := range committeesMap {
		for index, committee := range comittees {
			_, err := stmt.Exec(epoch, slot, index, pq.Array(committee.ValidatorIndices))
			if err != nil {
				return fmt.Errorf("error executing save beacon committee statement: %v", err)
			}
		}
	}

	return nil
}

func saveValidatorBalances(epoch uint64, balances []*types.ValidatorBalance, tx *sql.Tx) error {

	stmt, err := tx.Prepare(`INSERT INTO validator_balances (epoch, validatorindex, balance)
 													VALUES    ($1, $2, $3) ON CONFLICT (epoch, validatorindex) DO UPDATE SET balance = excluded.balance`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range balances {
		_, err := stmt.Exec(epoch, b.Index, b.Balance)
		if err != nil {
			return fmt.Errorf("error executing save validator balance statement: %v", err)
		}
	}

	return nil
}

func saveBlocks(epoch uint64, blocks map[uint64]map[string]*types.Block, tx *sql.Tx) error {

	stmtBlock, err := tx.Prepare(`INSERT INTO blocks (
                    epoch, 
                    slot, 
                    blockroot, 
                    parentroot, 
                    stateroot, 
                    signature, 
                    randaoreveal,
                    graffiti, 
                    eth1data_depositroot, 
                    eth1data_depositcount, 
                    eth1data_blockhash, 
                    proposerslashingscount, 
                    attesterslashingscount, 
                    attestationscount, 
                    depositscount, 
                    voluntaryexitscount, 
                    proposer,
                    status)
 					VALUES    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) ON CONFLICT (slot, blockroot) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBlock.Close()

	stmtProposerSlashing, err := tx.Prepare(`INSERT INTO blocks_proposerslashings (block_slot, block_index, proposerindex, header1_slot, header1_parentroot, header1_stateroot, header1_bodyroot, header1_signature, header2_slot, header2_parentroot, header2_stateroot, header2_bodyroot, header2_signature)
 													VALUES    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtProposerSlashing.Close()

	stmtAttesterSlashing, err := tx.Prepare(`INSERT INTO blocks_attesterslashings (block_slot, block_index, attestation1_custodybit_0indices, attestation1_custodybit_1indices, attestation1_signature, attestation1_slot, attestation1_index, attestation1_beaconblockroot, attestation1_source_epoch, attestation1_source_root, attestation1_target_epoch, attestation1_target_root, attestation2_custodybit_0indices, attestation2_custodybit_1indices, attestation2_signature, attestation2_slot, attestation2_index, attestation2_beaconblockroot, attestation2_source_epoch, attestation2_source_root, attestation2_target_epoch, attestation2_target_root)
 													VALUES    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtAttesterSlashing.Close()

	stmtAttestations, err := tx.Prepare(`INSERT INTO blocks_attestations (block_slot, block_index, aggregationbits, validators, custodybits, signature, slot, committeeindex, beaconblockroot, source_epoch, source_root, target_epoch, target_root)
 													VALUES    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtAttestations.Close()

	stmtDeposits, err := tx.Prepare(`INSERT INTO blocks_deposits (block_slot, block_index, proof, publickey, withdrawalcredentials, amount, signature)
 													VALUES    ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtDeposits.Close()

	stmtVoluntaryExits, err := tx.Prepare(`INSERT INTO blocks_voluntaryexits (block_slot, block_index, epoch, validatorindex, signature)
 													VALUES    ($1, $2, $3, $4, $5) ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtVoluntaryExits.Close()

	stmtProposalAssignments, err := tx.Prepare(`INSERT INTO proposal_assignments (epoch, validatorindex, proposerslot, status)
 													VALUES    ($1, $2, $3, $4) ON CONFLICT (epoch, validatorindex, proposerslot) DO UPDATE SET status = excluded.status`)
	if err != nil {
		return err
	}
	defer stmtProposalAssignments.Close()

	stmtAttestationAssignments, err := tx.Prepare(`INSERT INTO attestation_assignments (epoch, validatorindex, attesterslot, committeeindex, status)
 													VALUES    ($1, $2, $3, $4, $5) ON CONFLICT (epoch, validatorindex, attesterslot, committeeindex) DO UPDATE SET status = excluded.status`)
	if err != nil {
		return err
	}
	defer stmtAttestationAssignments.Close()

	for _, slot := range blocks {
		for _, b := range slot {
			var dbBlockRootHash []byte
			err := DB.Get(&dbBlockRootHash, "SELECT blockroot FROM blocks WHERE slot = $1 and blockroot = $2", b.Slot, b.BlockRoot)

			if err == nil && bytes.Compare(dbBlockRootHash, b.BlockRoot) == 0 {
				logger.Printf("Skipping export of block %x at slot %v as it is already present in the db", b.BlockRoot, b.Slot)
				continue
			}

			_, err = tx.Exec("DELETE FROM blocks WHERE slot = $1 AND length(blockroot) = 1", b.Slot) // Delete placeholder block
			if err != nil {
				return fmt.Errorf("error deleting placeholder block: %v", err)
			}

			_, err = stmtBlock.Exec(epoch,
				b.Slot,
				b.BlockRoot,
				b.ParentRoot,
				b.StateRoot,
				b.Signature,
				b.RandaoReveal,
				b.Graffiti,
				b.Eth1Data.DepositRoot,
				b.Eth1Data.DepositCount,
				b.Eth1Data.BlockHash,
				len(b.ProposerSlashings),
				len(b.AttesterSlashings),
				len(b.Attestations),
				len(b.Deposits),
				len(b.VoluntaryExits),
				b.Proposer,
				b.Status)
			if err != nil {
				return fmt.Errorf("error executing stmtBlocks for block %v: %v", b.Slot, err)
			}

			for i, ps := range b.ProposerSlashings {
				_, err := stmtProposerSlashing.Exec(b.Slot, i, ps.ProposerIndex, ps.Header1.Slot, ps.Header1.ParentRoot, ps.Header1.StateRoot, ps.Header1.BodyRoot, ps.Header1.Signature, ps.Header2.Slot, ps.Header2.ParentRoot, ps.Header2.StateRoot, ps.Header2.BodyRoot, ps.Header2.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtProposerSlashing for block %v: %v", b.Slot, err)
				}
			}

			for i, as := range b.AttesterSlashings {
				_, err := stmtAttesterSlashing.Exec(b.Slot, i, pq.Array(as.Attestation1.Custodybit0indices), pq.Array(as.Attestation1.Custodybit1indices), as.Attestation1.Signature, as.Attestation1.Data.Slot, as.Attestation1.Data.CommitteeIndex, as.Attestation1.Data.BeaconBlockRoot, as.Attestation1.Data.Source.Epoch, as.Attestation1.Data.Source.Root, as.Attestation1.Data.Target.Epoch, as.Attestation1.Data.Target.Root, pq.Array(as.Attestation2.Custodybit0indices), pq.Array(as.Attestation2.Custodybit1indices), as.Attestation2.Signature, as.Attestation2.Data.Slot, as.Attestation2.Data.CommitteeIndex, as.Attestation2.Data.BeaconBlockRoot, as.Attestation2.Data.Source.Epoch, as.Attestation2.Data.Source.Root, as.Attestation2.Data.Target.Epoch, as.Attestation2.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttesterSlashing for block %v: %v", b.Slot, err)
				}
			}

			for i, a := range b.Attestations {

				for _, validator := range a.Attesters {
					_, err = stmtAttestationAssignments.Exec(a.Data.Slot/utils.Config.Chain.SlotsPerEpoch, validator, a.Data.Slot, a.Data.CommitteeIndex, 1)

					if err != nil {
						return fmt.Errorf("error executing stmtAttestationAssignments for block %v: %v", b.Slot, err)
					}
				}

				_, err = stmtAttestations.Exec(b.Slot, i, bitfield.Bitlist(a.AggregationBits).Bytes(), pq.Array(a.Attesters), bitfield.Bitlist(a.CustodyBits).Bytes(), a.Signature, a.Data.Slot, a.Data.CommitteeIndex, a.Data.BeaconBlockRoot, a.Data.Source.Epoch, a.Data.Source.Root, a.Data.Target.Epoch, a.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttestations for block %v: %v", b.Slot, err)
				}
			}

			for i, d := range b.Deposits {
				_, err := stmtDeposits.Exec(b.Slot, i, nil, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtDeposits for block %v: %v", b.Slot, err)
				}
			}

			for i, ve := range b.VoluntaryExits {
				_, err := stmtVoluntaryExits.Exec(b.Slot, i, ve.Epoch, ve.ValidatorIndex, ve.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtVoluntaryExits for block %v: %v", b.Slot, err)
				}
			}

			_, err = stmtProposalAssignments.Exec(epoch, b.Proposer, b.Slot, b.Status)
			if err != nil {
				return fmt.Errorf("error executing stmtProposalAssignments for block %v: %v", b.Slot, err)
			}
		}
	}

	return nil
}

// UpdateEpochStatus will update the epoch status in the database
func UpdateEpochStatus(stats *types.ValidatorParticipation) error {
	_, err := DB.Exec(`UPDATE epochs SET 
                  finalized = $1, 
                  eligibleether = $2, 
                  globalparticipationrate = $3, 
                  votedether = $4
			WHERE epoch = $5`, stats.Finalized, stats.EligibleEther, stats.GlobalParticipationRate, stats.VotedEther, stats.Epoch)

	return err
}
