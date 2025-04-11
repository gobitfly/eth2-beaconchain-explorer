// Copyright (C) 2025 Bitfly GmbH
//
// This file is part of Beaconchain Dashboard.
//
// Beaconchain Dashboard is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Beaconchain Dashboard is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Beaconchain Dashboard.  If not, see <https://www.gnu.org/licenses/>.

package exporter

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gobitfly/eth2-beaconchain-explorer/version"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type PendingQueueIndexer struct {
	running   bool
	runningMu *sync.Mutex
	lc        rpc.Client
	db        *sqlx.DB
}

func NewPendingQueueIndexer(client rpc.Client) *PendingQueueIndexer {
	indexer := &PendingQueueIndexer{
		running:   false,
		lc:        client,
		db:        db.WriterDb,
		runningMu: &sync.Mutex{},
	}
	return indexer
}

func (qi *PendingQueueIndexer) Start() {
	qi.runningMu.Lock()
	if qi.running {
		qi.runningMu.Unlock()
		return
	}
	qi.running = true
	qi.runningMu.Unlock()

	logrus.WithFields(logrus.Fields{"version": version.Version}).Infof("starting pending queue indexer")
	for {
		err := qi.Index()
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Errorf("failed indexing pending queue")
		}
		logrus.Infof("pending queue indexer finished indexing, sleeping for 10 minutes")
		time.Sleep(time.Minute * 10) // interval MUST be longer than one epoch
		// Background: A freshly exported validator will have an eligible epoch of max uint64, by keeping the pending deposits
		// a bit longer in the db, we can rely on the pending deposits table to still get us an estimate for eligibility
	}
}

func (qi *PendingQueueIndexer) Index() error {
	head, err := qi.lc.GetChainHead()
	if err != nil {
		return errors.Wrap(err, "failed to get chain head")
	}
	epoch := head.HeadEpoch

	deposits, err := qi.lc.GetPendingDeposits()
	if err != nil {
		return errors.Wrap(err, "failed to get pending deposits")
	}

	validators, err := qi.lc.GetValidatorState(epoch)
	if err != nil {
		return errors.Wrap(err, "failed to get validator state")
	}

	type MiniState struct {
		Index             uint64
		ExitEpoch         uint64
		WithdrawableEpoch uint64
	}

	totalActiveEffectiveBalance := uint64(0)
	pubkeyToIndexMap := make(map[string]*MiniState)

	for _, v := range validators.Data {
		pubkeyToIndexMap[v.Validator.Pubkey] = &MiniState{
			Index:             uint64(v.Index),
			ExitEpoch:         uint64(v.Validator.ExitEpoch),
			WithdrawableEpoch: uint64(v.Validator.WithdrawableEpoch),
		}
		if epoch >= uint64(v.Validator.ActivationEpoch) && epoch < uint64(v.Validator.ExitEpoch) {
			totalActiveEffectiveBalance += uint64(v.Validator.EffectiveBalance)
		}
	}

	etherChurnByEpoch := utils.GetActivationExitChurnLimit(totalActiveEffectiveBalance)
	count := 0
	balanceAhead := uint64(0)
	clearEpoch := head.HeadEpoch + 1

	// transition period
	// pre electra system will keep going for follow distance until every deposit of the last system is converted to the new system
	// before the new system starts
	electraQueueDelay := uint64(utils.Config.ClConfig.Eth1FollowDistance/utils.Config.ClConfig.SlotsPerEpoch + utils.Config.ClConfig.EpochsPerEth1VotingPeriod)
	if clearEpoch < utils.Config.ClConfig.ElectraForkEpoch+electraQueueDelay {
		clearEpoch = utils.Config.ClConfig.ElectraForkEpoch + electraQueueDelay
	}

	depositsList := make([]types.PendingDeposit, 0)

	// spec vars (in snake_case)
	next_deposit_index := uint64(0)
	max_pending_deposits_per_epoch := utils.Config.ClConfig.MaxPendingDepositsPerEpoch
	if max_pending_deposits_per_epoch == 0 { // eth mainnet spec default
		max_pending_deposits_per_epoch = uint64(16)
	}
	processed_amount := uint64(0)
	state_deposit_balance_to_consume := uint64(0)

	for _, deposit := range deposits.Data {
		// potential improvement: utils.GetActivationExitChurnLimit(totalActiveEffectiveBalance + balanceAhead - withdrawalsAhead)

		// emulate spec based on current view in time (approx estimation)
		// https://github.com/ethereum/consensus-specs/blob/dev/specs/electra/beacon-chain.md#new-process_pending_deposits
		for {
			if deposit.Slot > (clearEpoch+1)*utils.Config.ClConfig.SlotsPerEpoch { // first slot of next epoch is finalized
				clearEpoch++
			} else {
				break
			}
		}

		if next_deposit_index >= max_pending_deposits_per_epoch {
			clearEpoch++
			next_deposit_index = 0
		}

		for {
			available_for_processing := state_deposit_balance_to_consume + etherChurnByEpoch
			is_churn_limit_reached := processed_amount+deposit.Amount > available_for_processing
			if is_churn_limit_reached {
				state_deposit_balance_to_consume = available_for_processing - processed_amount
				next_deposit_index = 0
				processed_amount = 0
				clearEpoch++
			} else {
				break
			}
		}

		miniState, found := pubkeyToIndexMap[deposit.Pubkey.String()]
		var is_validator_exited bool
		var is_validator_withdrawn bool

		if found {
			next_epoch := clearEpoch + 1
			is_validator_exited = miniState.ExitEpoch < 100_000_000_000
			is_validator_withdrawn = miniState.WithdrawableEpoch < next_epoch
		}

		// do not consume is_validator_withdrawn, assume withdrawn by then for is_validator_exited
		if !is_validator_withdrawn && !is_validator_exited {
			processed_amount += deposit.Amount
		}

		pendingDeposit := types.PendingDeposit{
			ID:                    count,
			Pubkey:                deposit.Pubkey,
			WithdrawalCredentials: deposit.WithdrawalCredentials,
			Amount:                deposit.Amount,
			Signature:             deposit.Signature,
			Slot:                  deposit.Slot,
			ValidatorIndex:        sql.NullInt64{},
			QueuedBalanceAhead:    balanceAhead,
			EstClearEpoch:         clearEpoch,
		}

		if found {
			pendingDeposit.ValidatorIndex = sql.NullInt64{
				Int64: int64(miniState.Index),
				Valid: true,
			}
		}
		depositsList = append(depositsList, pendingDeposit)

		next_deposit_index++
		balanceAhead += deposit.Amount
		count++
	}

	return qi.save(depositsList)
}

func (qi *PendingQueueIndexer) save(pendingDeposits []types.PendingDeposit) error {
	tx, err := qi.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to start db transaction")
	}

	defer tx.Rollback()

	// prepare data for bulk insert
	dat := make([][]interface{}, len(pendingDeposits))
	for i, r := range pendingDeposits {
		dat[i] = []interface{}{r.ID, r.ValidatorIndex, encodeToHex(r.Pubkey), encodeToHex(r.WithdrawalCredentials), r.Amount, encodeToHex(r.Signature), r.Slot, r.QueuedBalanceAhead, r.EstClearEpoch}
	}

	err = ClearAndCopyToTable(qi.db, "pending_deposits_queue", []string{"id", "validator_index", "pubkey", "withdrawal_credentials", "amount", "signature", "slot", "queued_balance_ahead", "est_clear_epoch"}, dat)
	if err != nil {
		return fmt.Errorf("error copying data to pending_deposits_queue table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func encodeToHex(data []byte) string {
	return fmt.Sprintf("\\x%x", data)
}

func ClearAndCopyToTable[T []any](db *sqlx.DB, tableName string, columns []string, data []T) error {
	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer conn.Close()
	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()

		pgxdecimal.Register(conn.TypeMap())
		tx, err := conn.Begin(context.Background())

		if err != nil {
			return err
		}
		defer func() {
			err := tx.Rollback(context.Background())
			if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				logrus.Error(err, "error rolling back transaction", 0)
			}
		}()

		// clear
		_, err = tx.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s", tableName))
		if err != nil {
			return errors.Wrap(err, "failed to truncate table")
		}

		// copy
		_, err = tx.CopyFrom(context.Background(), pgx.Identifier{tableName}, columns,
			pgx.CopyFromSlice(len(data), func(i int) ([]interface{}, error) {
				return data[i], nil
			}))

		if err != nil {
			return err
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error copying data to %s: %w", tableName, err)
	}
	return nil
}
