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

package types

import (
	"database/sql"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type PendingDeposit struct {
	ID                    int           `db:"id"`
	ValidatorIndex        sql.NullInt64 `db:"validator_index"`
	Pubkey                hexutil.Bytes `db:"pubkey"`
	WithdrawalCredentials hexutil.Bytes `db:"withdrawal_credentials"`
	Amount                uint64        `db:"amount"`
	Signature             hexutil.Bytes `db:"signature"`
	Slot                  uint64        `db:"slot"`
	QueuedBalanceAhead    uint64        `db:"queued_balance_ahead"`
	EstClearEpoch         uint64        `db:"est_clear_epoch"` // approx epoch where validator deposit will be credited on beaconchain and validator getting assigned an index (happens in transition from est_clear_epoch-1 to est_clear_epoch)
	// eligible = est_clear_epoch + 1
	// activation = eligible + 2 + MAX_SEED_LOOKAHEAD

	// More background:
	// in transition from est_clear_epoch-1 => est_clear_epoch: creation of validator
	// in transition from est_clear_epoch => est_clear_epoch+1: activation eligibility will be set to est_clear_epoch+1 [1 epoch delay]
	// in transition from est_clear_epoch+2 => est_clear_epoch+3: checkpoint est_clear_epoch+1 finalized, set activation epoch to (est_clear_epoch+1)+1+4  [1 epoch delay, 4 = MAX_SEED_LOOKAHEAD]
}
