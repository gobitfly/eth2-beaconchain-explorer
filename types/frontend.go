package types

import (
	"time"

	"github.com/pkg/errors"
)

type EventName string

const (
	ValidatorBalanceDecreasedEventName              EventName = "validator_balance_decreased"
	ValidatorMissedProposalEventName                EventName = "validator_missed_proposal"
	ValidatorMissedAttestationEventName             EventName = "validator_missed_attestation"
	ValidatorGotSlashedEventName                    EventName = "validator_got_slashed"
	ValidatorDidSlashEventName                      EventName = "validator_did_slash"
	ValidatorStateChangedEventName                  EventName = "validator_state_changed"
	ValidatorReceivedDepositEventName               EventName = "validator_received_deposit"
	NetworkSlashingEventName                        EventName = "network_slashing"
	NetworkValidatorActivationQueueFullEventName    EventName = "network_validator_activation_queue_full"
	NetworkValidatorActivationQueueNotFullEventName EventName = "network_validator_activation_queue_not_full"
	NetworkValidatorExitQueueFullEventName          EventName = "network_validator_exit_queue_full"
	NetworkValidatorExitQueueNotFullEventName       EventName = "network_validator_exit_queue_not_full"
	NetworkLivenessIncreasedEventName               EventName = "network_liveness_increased"
)

var EventNames = []EventName{
	ValidatorBalanceDecreasedEventName,
	ValidatorMissedProposalEventName,
	ValidatorMissedAttestationEventName,
	ValidatorGotSlashedEventName,
	ValidatorDidSlashEventName,
	ValidatorStateChangedEventName,
	ValidatorReceivedDepositEventName,
	NetworkSlashingEventName,
	NetworkValidatorActivationQueueFullEventName,
	NetworkValidatorActivationQueueNotFullEventName,
	NetworkValidatorExitQueueFullEventName,
	NetworkValidatorExitQueueNotFullEventName,
	NetworkLivenessIncreasedEventName,
}

func EventNameFromString(event string) (EventName, error) {
	for _, en := range EventNames {
		if string(en) == event {
			return en, nil
		}
	}
	return "", errors.Errorf("Could not convert event to string. %v is not a known event type", event)
}

type Tag string

const (
	ValidatorTagsWatchlist Tag = "watchlist"
)

type Notification interface {
	GetSubscriptionID() uint64
	GetEventName() EventName
	GetInfo() string
}

type Subscription struct {
	ID          uint64     `db:"id"`
	UserID      uint64     `db:"user_id"`
	EventName   EventName  `db:"event_name"`
	EventFilter string     `db:"event_filter"`
	LastSent    *time.Time `db:"last_sent_ts"`
	Created     time.Time  `db:"created_ts"`
}

type TaggedValidators struct {
	UserID uint64 `db:"user_id"`
	Tag    string `db:"tag"`
	Validator
	Events []EventName `db:"events"`
}
