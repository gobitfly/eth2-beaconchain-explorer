package types

import (
	"time"

	"github.com/pkg/errors"
)

type EventName string

const (
	ValidatorBalanceDecreasedEventName EventName = "validator_balance_decreased"
)

func EventFromString(event string) (EventName, error) {
	switch event {
	case "validator_balance_decreased":
		return ValidatorBalanceDecreasedEventName, nil
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
	Sent        *time.Time `db:"sent_ts"`
}
