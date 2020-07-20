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
	ID                 uint64     `db:"id"`
	UserID             uint64     `db:"user_id"`
	EventName          string     `db:"event_name"`
	ValidatorPublicKey *string    `db:"validator_publickey"`
	LastNotification   *time.Time `db:"last_notification_ts"`
}
