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

func GetUserEmailById(id uint64) (string, error) {
	var mail string = ""
	err := FrontendDB.Get(&mail, `
	SELECT 
		email
	FROM 
		users
	WHERE id = $1`, id)
	return mail, err
}

func DeleteUserByEmail(email string) error {
	_, err := FrontendDB.Exec(`
	DELETE 
	FROM 
		users
	WHERE email = $1`, email)
	return err
}

func DeleteUserById(id uint64) error {
	_, err := FrontendDB.Exec(`
	DELETE 
	FROM 
		users
	WHERE id = $1`, id)
	return err
}

func UpdatePassword(userId uint64, hash []byte) error {
	_, err := FrontendDB.Exec("UPDATE users SET password = $1 WHERE id = $2", hash, userId)
	return err
}

func UpdateSubscriptionTime(subscriptionID uint64, t time.Time) error {
	_, err := FrontendDB.Exec("UPDATE users_subscriptions SET last_notification_ts = TO_TIMESTAMP($1) WHERE id = $2", t.Unix(), subscriptionID)
	return err
}

func AddSubscription(userID uint64, eventName types.EventName, eventFilter string) error {
	_, err := FrontendDB.Exec("INSERT INTO users_subscriptions (user_id, event_name, event_filter) VALUES ($1, $2, $3)", userID, eventName, eventFilter)
	return err
}

type GetSubscriptionsFilter struct {
	EventNames   *[]types.EventName
	UserIDs      *[]uint64
	EventFilters *[]string
}

func GetSubscriptions(filter GetSubscriptionsFilter) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}
	qry := "SELECT * FROM users_subscriptions"
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

func UpdateSubscriptionsSent(subscriptionIDs []uint64, sent time.Time) error {
	_, err := FrontendDB.Exec("UPDATE users_subscriptions SET sent_ts = TO_TIMESTAMP($1) WHERE id = ANY($2)", sent.Unix(), pq.Array(subscriptionIDs))
	return err
}
