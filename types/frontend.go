package types

import (
	"time"

	"github.com/pkg/errors"
)

type EventName string

const (
	ValidatorBalanceDecreasedEventName  EventName = "validator_balance_decreased"
	ValidatorProposalMissedEventName    EventName = "validator_proposal_missed"
	ValidatorAttestationMissedEventName EventName = "validator_attestation_missed"
	ValidatorGotSlashedEventName        EventName = "validator_got_slashed"
	ValidatorDidSlashEventName          EventName = "validator_did_slash"
	ValidatorStateChangedEventName      EventName = "validator_state_changed"
)

func EventFromString(event string) (EventName, error) {
	switch event {
	case "validator_balance_decreased":
		return ValidatorBalanceDecreasedEventName, nil
	case "validator_proposal_missed":
		return ValidatorProposalMissedEventName, nil
	case "validator_attestation_missed":
		return ValidatorAttestationMissedEventName, nil
	case "validator_got_slashed":
		return ValidatorGotSlashedEventName, nil
	case "validator_did_slash":
		return ValidatorDidSlashEventName, nil
	case "validator_state_changed":
		return ValidatorStateChangedEventName, nil
	default:
		return "", errors.Errorf("Could not convert event to string. %v is not a known event type", event)
	}
}

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
