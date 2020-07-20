package db

import (
	"eth2-exporter/types"
	"time"

	"github.com/jmoiron/sqlx"
)

// FrontendDB is a pointer to the auth-database
var FrontendDB *sqlx.DB

func MustInitFrontendDB(username, password, host, port, name, sessionSecret string) {
	FrontendDB = mustInitDB(username, password, host, port, name)
}

func GetUserEmailById(id int64) (string, error) {
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

func DeleteUserById(id int64) error {
	_, err := FrontendDB.Exec(`
	DELETE 
	FROM 
		users
	WHERE id = $1`, id)
	return err
}

func UpdatePassword(userId int64, hash []byte) error {
	_, err := FrontendDB.Exec("UPDATE users SET password = $1 WHERE id = $2", hash, userId)
	return err
}

func UpdateSubscriptionTime(subscriptionID uint64, t time.Time) error {
	_, err := FrontendDB.Exec("UPDATE notifications_subscriptions SET last_notification_ts = TO_TIMESTAMP($1) WHERE id = $2", t.Unix(), subscriptionID)
	return err
}

func AddSubscription(userID int64, eventName string, validatorIndex *uint64) error {
	var err error
	if validatorIndex == nil {
		_, err = FrontendDB.Exec("INSERT INTO notifications_subscriptions (user_id, event_name) VALUES ($1, $2)", userID, eventName)
		return err
	}
	_, err = FrontendDB.Exec("INSERT INTO notifications_subscriptions (user_id, event_name, validatorindex) VALUES ($1, $2, $3)", userID, eventName, *validatorIndex)
	return err
}

func GetSubscriptions(eventName types.EventName) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}
	err := FrontendDB.Select(&subs, "SELECT * FROM notifications_subscriptions WHERE event_name = $1", eventName)
	return subs, err
}

func GetUserSubscription(userId int64, pubKey []byte) (*types.Subscription, error) {
	sub := &types.Subscription{}
	err := FrontendDB.Get(&sub, "SELECT * FROM notifications_subscriptions WHERE user_id = $1 and pubkey = $2", userId, pubKey)
	return sub, err
}

func GetUserSubscriptions(userId int64) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}
	err := FrontendDB.Select(&subs, "SELECT * FROM notifications_subscriptions WHERE user_id = $1", userId)
	return subs, err
}
