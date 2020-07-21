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
	now := time.Now().Unix()
	_, err := FrontendDB.Exec("INSERT INTO users_subscriptions (user_id, event_name, event_filter, created_ts) VALUES ($1, $2, $3, TO_TIMESTAMP($4))", userID, eventName, eventFilter, now)
	return err
}

// DeleteSubscription removes a subscription from the database.
func DeleteSubscription(userID uint64, eventName types.EventName, eventFilter string) error {
	_, err := FrontendDB.Exec("DELETE FROM users_subscriptions WHERE user_id = $1 and event_name = $2 and event_filter = $3", userID, eventName, eventFilter)
	return err
}

// GetSubscriptionsFilter can be passed to GetSubscriptions() to filter subscriptions.
type GetSubscriptionsFilter struct {
	EventNames   *[]types.EventName
	UserIDs      *[]uint64
	EventFilters *[]string
	Search       string
	Limit        uint64
	Offset       uint64
}

// GetSubscriptions returns the subscriptions filtered by the provided filter.
func GetSubscriptions(filter GetSubscriptionsFilter) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}

	qry := "SELECT * FROM users_subscriptions"
	if filter.EventNames == nil && filter.UserIDs == nil && filter.EventFilters == nil {
		err := FrontendDB.Select(&subs, qry)
		return subs, err
	}

	filters := []string{}
	args := []interface{}{}

	if filter.EventNames != nil {
		args = append(args, pq.Array(*filter.EventNames))
		filters = append(filters, fmt.Sprintf("event_name = ANY($%d)", len(args)))
	}

	if filter.UserIDs != nil {
		args = append(args, pq.Array(*filter.UserIDs))
		filters = append(filters, fmt.Sprintf("user_id = ANY($%d)", len(args)))
	}

	if filter.EventFilters != nil {
		args = append(args, pq.Array(*filter.EventFilters))
		filters = append(filters, fmt.Sprintf("event_filter = ANY($%d)", len(args)))
	}
	qry += " WHERE " + strings.Join(filters, " AND ")

	if filter.Search != "" {
		args = append(args, filter.Search+"%")
		qry += fmt.Sprintf(" AND event_filter LIKE LOWER($%d)", len(args))
	}

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		qry += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	args = append(args, filter.Offset)
	qry += fmt.Sprintf(" OFFSET $%d", len(args))
	err := FrontendDB.Select(&subs, qry, args...)
	return subs, err
}

// UpdateSubscriptionsLastSent upates `last_sent_ts` column of the `users_subscriptions` table.
func UpdateSubscriptionsLastSent(subscriptionIDs []uint64, sent time.Time) error {
	_, err := FrontendDB.Exec("UPDATE users_subscriptions SET last_sent_ts = TO_TIMESTAMP($1) WHERE id = ANY($2)", sent.Unix(), pq.Array(subscriptionIDs))
	return err
}

// CountSentMail increases the count of sent mails in the table `mails_sent` for this day.
func CountSentMail(email string) error {
	day := time.Now().Truncate(time.Hour * 24).Unix()
	_, err := FrontendDB.Exec(`
		INSERT INTO mails_sent (email, ts, cnt) VALUES ($1, $2, 1)
		ON CONFLICT (email, ts) DO UPDATE SET cnt = cnt+1`, email, day)
	return err
}

// GetMailsSentCount returns the number of sent mails for the day of the passed time.
func GetMailsSentCount(email string, t time.Time) (int, error) {
	day := t.Truncate(time.Hour * 24).Unix()
	count := 0
	err := FrontendDB.Get(&count, "SELECT cnt FROM mails_sent WHERE email = $1 AND ts = $2", email, day)
	return count, err
}
