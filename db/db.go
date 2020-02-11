package db

import (
	"bytes"
	"database/sql"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"strings"
	"time"

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
		logger.Printf("marking block %x as orphaned", orphanedBlock)
		_, err = tx.Exec("UPDATE blocks SET status = '3' WHERE blockroot = $1", orphanedBlock)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
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

	stmtValidatorQueueActivation, err := tx.Prepare(`
		INSERT INTO validatorqueue_activation (index, publickey)
		VALUES ($1, $2) ON CONFLICT (index, publickey) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtValidatorQueueActivation.Close()

	stmtValidatorQueueExit, err := tx.Prepare(`
		INSERT INTO validatorqueue_exit (index, publickey)
		VALUES ($1, $2) ON CONFLICT (index, publickey) DO NOTHING`)
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

	logger.Infof("starting export of epoch %v", data.Epoch)
	start := time.Now()

	logger.Infof("exporting block data")
	err = saveBlocks(data.Epoch, data.Blocks, tx)
	if err != nil {
		return fmt.Errorf("error saving blocks to db: %v", err)
	}

	logger.Infof("exporting validators data")
	err = saveValidators(data.Epoch, data.Validators, tx)
	if err != nil {
		return fmt.Errorf("error saving validators to db: %v", err)
	}

	logger.Infof("exporting proposal assignments data")
	err = saveValidatorProposalAssignments(data.Epoch, data.ValidatorAssignmentes.ProposerAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator assignments to db: %v", err)
	}

	logger.Infof("exporting attestation assignments data")
	err = saveValidatorAttestationAssignments(data.Epoch, data.ValidatorAssignmentes.AttestorAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator assignments to db: %v", err)
	}

	logger.Infof("exporting validator balance data")
	err = saveValidatorBalances(data.Epoch, data.Validators, tx)
	if err != nil {
		return fmt.Errorf("error saving validator balances to db: %v", err)
	}

	logger.Infof("exporting epoch statistics data")
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
	for _, v := range data.Validators {
		validatorBalanceSum = new(big.Int).Add(validatorBalanceSum, new(big.Int).SetUint64(v.Balance))
	}

	validatorBalanceAverage := new(big.Int).Div(validatorBalanceSum, new(big.Int).SetInt64(int64(len(data.Validators)))).Uint64()

	validatorsCount := 0
	for _, v := range data.Validators {
		if v.ExitEpoch > data.Epoch {
			validatorsCount++
		}
	}

	_, err = tx.Exec(`
		INSERT INTO epochs (
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
		ON CONFLICT (epoch) DO UPDATE SET 
			blockscount             = excluded.blockscount, 
			proposerslashingscount  = excluded.proposerslashingscount,
			attesterslashingscount  = excluded.attesterslashingscount,
			attestationscount       = excluded.attestationscount,
			depositscount           = excluded.depositscount,
			voluntaryexitscount     = excluded.voluntaryexitscount,
			validatorscount         = excluded.validatorscount,
			averagevalidatorbalance = excluded.averagevalidatorbalance,
			finalized               = excluded.finalized,
			eligibleether           = excluded.eligibleether,
			globalparticipationrate = excluded.globalparticipationrate,
			votedether              = excluded.votedether`,
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

	logger.Infof("export of epoch %v completed, took %v", data.Epoch, time.Since(start))
	return nil
}

func saveValidators(epoch uint64, validators []*types.Validator, tx *sql.Tx) error {
	stmtValidators, err := tx.Prepare(`
		INSERT INTO validators (
			validatorindex,
			pubkey,
			withdrawableepoch,
			withdrawalcredentials,
			balance,
			effectivebalance,
			slashed,
			activationeligibilityepoch,
			activationepoch,
			exitepoch
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		ON CONFLICT (validatorindex) DO UPDATE SET 
			pubkey                     = EXCLUDED.pubkey,
			withdrawableepoch          = EXCLUDED.withdrawableepoch,
			withdrawalcredentials      = EXCLUDED.withdrawalcredentials,
			balance                    = EXCLUDED.balance,
			effectivebalance           = EXCLUDED.effectivebalance,
			slashed                    = EXCLUDED.slashed,
			activationeligibilityepoch = EXCLUDED.activationeligibilityepoch,
			activationepoch            = EXCLUDED.activationepoch,
			exitepoch                  = EXCLUDED.exitepoch`)
	if err != nil {
		return err
	}
	defer stmtValidators.Close()

	var lenActivatedValidators int
	var lastActivatedValidatorIdx uint64
	for _, v := range validators {
		if !(v.ActivationEpoch <= epoch && epoch < v.ExitEpoch) {
			continue
		}
		lenActivatedValidators++
		if v.Index < lastActivatedValidatorIdx {
			continue
		}
		lastActivatedValidatorIdx = v.Index
	}

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
		if v.ActivationEligibilityEpoch < 9223372036854775807 && v.ActivationEpoch == 9223372036854775807 {
			// see: https://github.com/ethereum/eth2.0-specs/blob/master/specs/phase0/beacon-chain.md#get_validator_churn_limit
			// validator_churn_limit = max(MIN_PER_EPOCH_CHURN_LIMIT, len(active_validator_indices) // CHURN_LIMIT_QUOTIENT)
			// validator_churn_limit = max(4, len(active_set) / 2**16)
			// validator.activationepoch = epoch + validator.positioninactivationqueue / validator_churn_limit
			// note: this is only an estimation
			positionInActivationQueue := v.Index - lastActivatedValidatorIdx
			churnLimit := float64(lenActivatedValidators) / 65536
			if churnLimit < 4 {
				churnLimit = 4
			}
			if v.ActivationEligibilityEpoch > epoch {
				v.ActivationEpoch = v.ActivationEligibilityEpoch + uint64(float64(positionInActivationQueue)/churnLimit)
			} else {
				v.ActivationEpoch = epoch + uint64(float64(positionInActivationQueue)/churnLimit)
			}
		}
		_, err = stmtValidators.Exec(v.Index, v.PublicKey, v.WithdrawableEpoch, v.WithdrawalCredentials, v.Balance, v.EffectiveBalance, v.Slashed, v.ActivationEligibilityEpoch, v.ActivationEpoch, v.ExitEpoch)
		if err != nil {
			return fmt.Errorf("error executing save validator statement: %v", err)
		}
	}

	return nil
}

func saveValidatorProposalAssignments(epoch uint64, assignments map[uint64]uint64, tx *sql.Tx) error {
	stmt, err := tx.Prepare(`
		INSERT INTO proposal_assignments (epoch, validatorindex, proposerslot, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (epoch, validatorindex, proposerslot) DO NOTHING`)
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
	stmtAttestationAssignments, err := tx.Prepare(`
		INSERT INTO attestation_assignments (epoch, validatorindex, attesterslot, committeeindex, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (epoch, validatorindex, attesterslot, committeeindex) DO NOTHING`)
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

func saveValidatorBalances(epoch uint64, validators []*types.Validator, tx *sql.Tx) error {
	stmt, err := tx.Prepare(`
		INSERT INTO validator_balances (epoch, validatorindex, balance, effectivebalance)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (epoch, validatorindex) DO UPDATE SET
			balance          = EXCLUDED.balance,
			effectivebalance = EXCLUDED.effectivebalance`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range validators {
		_, err := stmt.Exec(epoch, v.Index, v.Balance, v.EffectiveBalance)
		if err != nil {
			return fmt.Errorf("error executing save validator balance statement: %v", err)
		}
	}

	return nil
}

func saveBlocks(epoch uint64, blocks map[uint64]map[string]*types.Block, tx *sql.Tx) error {
	stmtBlock, err := tx.Prepare(`
		INSERT INTO blocks (epoch, slot, blockroot, parentroot, stateroot, signature, randaoreveal, graffiti, eth1data_depositroot, eth1data_depositcount, eth1data_blockhash, proposerslashingscount, attesterslashingscount, attestationscount, depositscount, voluntaryexitscount, proposer, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (slot, blockroot) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBlock.Close()

	stmtProposerSlashing, err := tx.Prepare(`
		INSERT INTO blocks_proposerslashings (block_slot, block_index, proposerindex, header1_slot, header1_parentroot, header1_stateroot, header1_bodyroot, header1_signature, header2_slot, header2_parentroot, header2_stateroot, header2_bodyroot, header2_signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtProposerSlashing.Close()

	stmtAttesterSlashing, err := tx.Prepare(`
		INSERT INTO blocks_attesterslashings (block_slot, block_index, attestation1_signature, attestation1_slot, attestation1_index, attestation1_beaconblockroot, attestation1_source_epoch, attestation1_source_root, attestation1_target_epoch, attestation1_target_root, attestation2_signature, attestation2_slot, attestation2_index, attestation2_beaconblockroot, attestation2_source_epoch, attestation2_source_root, attestation2_target_epoch, attestation2_target_root)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtAttesterSlashing.Close()

	stmtAttestations, err := tx.Prepare(`
		INSERT INTO blocks_attestations (block_slot, block_index, aggregationbits, validators, signature, slot, committeeindex, beaconblockroot, source_epoch, source_root, target_epoch, target_root)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtAttestations.Close()

	stmtDeposits, err := tx.Prepare(`
		INSERT INTO blocks_deposits (block_slot, block_index, proof, publickey, withdrawalcredentials, amount, signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtDeposits.Close()

	stmtVoluntaryExits, err := tx.Prepare(`
		INSERT INTO blocks_voluntaryexits (block_slot, block_index, epoch, validatorindex, signature)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtVoluntaryExits.Close()

	stmtProposalAssignments, err := tx.Prepare(`
		INSERT INTO proposal_assignments (epoch, validatorindex, proposerslot, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (epoch, validatorindex, proposerslot) DO UPDATE SET status = excluded.status`)
	if err != nil {
		return err
	}
	defer stmtProposalAssignments.Close()

	stmtAttestationAssignments, err := tx.Prepare(`
		INSERT INTO attestation_assignments (epoch, validatorindex, attesterslot, committeeindex, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (epoch, validatorindex, attesterslot, committeeindex) DO UPDATE SET status = excluded.status`)
	if err != nil {
		return err
	}
	defer stmtAttestationAssignments.Close()

	stmtValidatorsLastAttestationSlot, err := tx.Prepare(`UPDATE validators SET lastattestationslot = $1 WHERE validatorindex = $2`)
	if err != nil {
		return err
	}
	defer stmtValidatorsLastAttestationSlot.Close()

	for _, slot := range blocks {
		for _, b := range slot {
			var dbBlockRootHash []byte
			err := DB.Get(&dbBlockRootHash, "SELECT blockroot FROM blocks WHERE slot = $1 and blockroot = $2", b.Slot, b.BlockRoot)

			if err == nil && bytes.Compare(dbBlockRootHash, b.BlockRoot) == 0 {
				logger.Printf("skipping export of block %x at slot %v as it is already present in the db", b.BlockRoot, b.Slot)
				continue
			}
			start := time.Now()
			logger.Infof("exporting block %x at slot %v", b.BlockRoot, b.Slot)

			logger.Infof("deleting placeholder block")
			_, err = tx.Exec("DELETE FROM blocks WHERE slot = $1 AND length(blockroot) = 1", b.Slot) // Delete placeholder block
			if err != nil {
				return fmt.Errorf("error deleting placeholder block: %v", err)
			}

			logger.Infof("writing block data")
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

			logger.Infof("writing proposer slashings data")
			for i, ps := range b.ProposerSlashings {
				_, err := stmtProposerSlashing.Exec(b.Slot, i, ps.ProposerIndex, ps.Header1.Slot, ps.Header1.ParentRoot, ps.Header1.StateRoot, ps.Header1.BodyRoot, ps.Header1.Signature, ps.Header2.Slot, ps.Header2.ParentRoot, ps.Header2.StateRoot, ps.Header2.BodyRoot, ps.Header2.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtProposerSlashing for block %v: %v", b.Slot, err)
				}
			}

			logger.Infof("writing attester slashings data")
			for i, as := range b.AttesterSlashings {
				_, err := stmtAttesterSlashing.Exec(b.Slot, i, as.Attestation1.Signature, as.Attestation1.Data.Slot, as.Attestation1.Data.CommitteeIndex, as.Attestation1.Data.BeaconBlockRoot, as.Attestation1.Data.Source.Epoch, as.Attestation1.Data.Source.Root, as.Attestation1.Data.Target.Epoch, as.Attestation1.Data.Target.Root, as.Attestation2.Signature, as.Attestation2.Data.Slot, as.Attestation2.Data.CommitteeIndex, as.Attestation2.Data.BeaconBlockRoot, as.Attestation2.Data.Source.Epoch, as.Attestation2.Data.Source.Root, as.Attestation2.Data.Target.Epoch, as.Attestation2.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttesterSlashing for block %v: %v", b.Slot, err)
				}
			}

			logger.Infof("writing attestation data")
			for i, a := range b.Attestations {

				for _, validator := range a.Attesters {
					_, err = stmtAttestationAssignments.Exec(a.Data.Slot/utils.Config.Chain.SlotsPerEpoch, validator, a.Data.Slot, a.Data.CommitteeIndex, 1)
					if err != nil {
						return fmt.Errorf("error executing stmtAttestationAssignments for block %v: %v", b.Slot, err)
					}

					_, err = stmtValidatorsLastAttestationSlot.Exec(a.Data.Slot, validator)
					if err != nil {
						return fmt.Errorf("error executing stmtValidatorsLastAttestationSlot for block %v: %v", b.Slot, err)
					}
				}

				_, err = stmtAttestations.Exec(b.Slot, i, bitfield.Bitlist(a.AggregationBits).Bytes(), pq.Array(a.Attesters), a.Signature, a.Data.Slot, a.Data.CommitteeIndex, a.Data.BeaconBlockRoot, a.Data.Source.Epoch, a.Data.Source.Root, a.Data.Target.Epoch, a.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttestations for block %v: %v", b.Slot, err)
				}
			}

			logger.Infof("writing deposits data")
			for i, d := range b.Deposits {
				_, err := stmtDeposits.Exec(b.Slot, i, nil, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtDeposits for block %v: %v", b.Slot, err)
				}
			}

			logger.Infof("writing voluntary exits data")
			for i, ve := range b.VoluntaryExits {
				_, err := stmtVoluntaryExits.Exec(b.Slot, i, ve.Epoch, ve.ValidatorIndex, ve.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtVoluntaryExits for block %v: %v", b.Slot, err)
				}
			}

			logger.Infof("writing proposal assignments data")
			_, err = stmtProposalAssignments.Exec(epoch, b.Proposer, b.Slot, b.Status)
			if err != nil {
				return fmt.Errorf("error executing stmtProposalAssignments for block %v: %v", b.Slot, err)
			}

			logger.Infof("export of block %x at slot %v completed, took %v", b.BlockRoot, b.Slot, time.Since(start))
		}
	}

	return nil
}

// UpdateEpochStatus will update the epoch status in the database
func UpdateEpochStatus(stats *types.ValidatorParticipation) error {
	_, err := DB.Exec(`
		UPDATE epochs SET
			finalized = $1,
			eligibleether = $2,
			globalparticipationrate = $3,
			votedether = $4
		WHERE epoch = $5`,
		stats.Finalized, stats.EligibleEther, stats.GlobalParticipationRate, stats.VotedEther, stats.Epoch)

	return err
}
