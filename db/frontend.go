package db

import (
	"eth2-exporter/types"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// FrontendDB is a pointer to the auth-database
var FrontendDB *sqlx.DB

func MustInitFrontendDB(username, password, host, port, name, sessionSecret string) {
	FrontendDB = mustInitDB(username, password, host, port, name)
}

// GetUserEmailById returns the email of a user.
func GetUserEmailById(id uint64) (string, error) {
	var mail string = ""
	err := FrontendDB.Get(&mail, "SELECT email FROM users WHERE id = $1", id)
	return mail, err
}

// DeleteUserByEmail deletes a user.
func DeleteUserByEmail(email string) error {
	_, err := FrontendDB.Exec("DELETE FROM users WHERE email = $1", email)
	return err
}

// DeleteUserById deletes a user.
func DeleteUserById(id uint64) error {
	_, err := FrontendDB.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// UpdatePassword updates the password of a user.
func UpdatePassword(userId uint64, hash []byte) error {
	_, err := FrontendDB.Exec("UPDATE users SET password = $1 WHERE id = $2", hash, userId)
	return err
}

// AddSubscription adds a new subscription to the database.
func AddSubscription(userID uint64, eventName types.EventName, eventFilter string) error {
	_, err := FrontendDB.Exec("INSERT INTO users_subscriptions (user_id, event_name, event_filter) VALUES ($1, $2, $3)", userID, eventName, eventFilter)
	return err
}

// DeleteSubscription removes a subscription from the database.
func DeleteSubscription(userID uint64, eventName types.EventName, eventFilter string) error {
	_, err := FrontendDB.Exec("DELETE FROM notifications_subscriptions WHERE user_id = $1 and event_name = $2 and event_filter = $3", userID, eventName, eventFilter)
	return err
}

// GetSubscriptionsFilter can be passed to GetSubscriptions() to filter subscriptions.
type GetSubscriptionsFilter struct {
	EventNames   *[]types.EventName
	UserIDs      *[]uint64
	EventFilters *[]string
}

// GetSubscriptions returns the subscriptions filtered by the provided filter.
func GetSubscriptions(filter GetSubscriptionsFilter) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}

	qry := "SELECT id, user_id, event_name, event_filter, last_notification_ts FROM notifications_subscriptions"
	if filter.EventNames == nil || filter.UserIDs == nil || filter.EventFilters == nil {
		err := FrontendDB.Select(&subs, qry)
		return subs, err
	}

	filters := []string{}
	args := []interface{}{}

	if filter.EventNames != nil {
		filters = append(filters, fmt.Sprintf("event_name = ANY($%d)", len(filters)+1))
		args = append(args, pq.Array(*filter.EventNames))
	}

	if filter.UserIDs != nil {
		filters = append(filters, fmt.Sprintf("user_id = ANY($%d)", len(filters)+1))
		args = append(args, pq.Array(*filter.UserIDs))
	}

	if filter.EventFilters != nil {
		filters = append(filters, fmt.Sprintf("event_filter = ANY($%d)", len(filters)+1))
		args = append(args, pq.Array(*filter.EventFilters))
	}

	qry += " WHERE " + strings.Join(filters, " AND ")

	err := FrontendDB.Select(&subs, qry, args...)
	return subs, err
}

// UpdateSubscriptionsSent upates `sent_ts` column of the `users_subscriptions` table.
func UpdateSubscriptionsSent(subscriptionIDs []uint64, sent time.Time) error {
	_, err := FrontendDB.Exec("UPDATE users_subscriptions SET sent_ts = TO_TIMESTAMP($1) WHERE id = ANY($2)", sent.Unix(), pq.Array(subscriptionIDs))
	return err
}
