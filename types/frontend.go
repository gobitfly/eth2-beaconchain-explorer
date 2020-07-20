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
	ID          uint64     `db:"id"`
	UserID      uint64     `db:"user_id"`
	EventName   EventName  `db:"event_name"`
	EventFilter string     `db:"event_filter"`
	Sent        *time.Time `db:"sent_ts"`
}
