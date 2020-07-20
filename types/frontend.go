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
	EventName() EventName
	Info() string
}

type Subscription struct {
	ID                 uint64     `db:"id"`
	UserID             uint64     `db:"user_id"`
	EventName          string     `db:"event_name"`
	ValidatorPublicKey *string    `db:"validator_publickey"`
	LastNotification   *time.Time `db:"last_notification_ts"`
}
