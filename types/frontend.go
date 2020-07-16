package types

import "time"

type EventName string

const (
	ValidatorBalanceDecreasedEventName EventName = "validator_balance_decreased"
)

type Notification interface {
	EventName() EventName
	Info() string
}

type Subscription struct {
	ID               uint64     `db:"id"`
	UserID           int64      `db:"user_id"`
	EventName        string     `db:"event_name"`
	ValidatorIndex   *uint64    `db:"validatorindex"`
	LastNotification *time.Time `db:"last_notification_ts"`
}
