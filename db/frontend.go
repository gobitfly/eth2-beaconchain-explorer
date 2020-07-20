package db

import (
	"eth2-exporter/types"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
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
	_, err := FrontendDB.Exec("UPDATE notifications_subscriptions SET last_notification_ts = TO_TIMESTAMP($1) WHERE id = $2", t.Unix(), subscriptionID)
	return err
}

func AddSubscription(userID uint64, eventName types.EventName, validatorPublickey *string) error {

	// check if the subscription already exists
	filter := GetSubscriptionsFilter{
		EventNames: &[]types.EventName{eventName},
		UserIDs:    &[]uint64{userID},
	}
	
	if validatorPublickey != nil {
		
		filter.ValidatorPubkey = []string{validatorPublicKey}
	}

	subs, err := GetSubscriptions(filter)
	if err != nil {
		return err
	}

	if len(subs) != 0 {
		return errors.Errorf("This subscription does not already exist. user: %v, event: %v", userID, eventName)
	}

	if validatorPublickey == nil {
		_, err = FrontendDB.Exec("INSERT INTO notifications_subscriptions (user_id, event_name) VALUES ($1, $2)", userID, eventName)
		return err
	}
	_, err = FrontendDB.Exec("INSERT INTO notifications_subscriptions (user_id, event_name, validator_publickey) VALUES ($1, $2, $3)", userID, eventName, *validatorPublickey)
	return err
}

type GetSubscriptionsFilter struct {
	EventNames *[]types.EventName
	UserIDs    *[]uint64
	validatorPubkey *[]string
}

func GetSubscriptions(filter GetSubscriptionsFilter) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}
	qry := "SELECT id, user_id, event_name, encode(validator_publickey::bytea, 'hex'), last_notification_ts FROM notifications_subscriptions"
	var args []interface{}
	if filter.EventNames != nil && filter.UserIDs != nil {
		qry += " WHERE event_name = ANY($1) AND user_id = ANY($2)"
		args = []interface{}{pq.Array(*filter.EventNames), pq.Array(*filter.UserIDs)}
	} else if filter.EventNames != nil {
		qry += " WHERE event_name = ANY($1)"
		args = []interface{}{pq.Array(*filter.EventNames)}
	} else if filter.UserIDs != nil {
		qry += " WHERE user_id = ANY($1)"
		args = []interface{}{pq.Array(*filter.UserIDs)}
	}
	err := FrontendDB.Select(&subs, qry, args...)
	return subs, err
}

func GetUserValidatorSubscriptions(userId uint64, pubKey []byte) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}
	err := FrontendDB.Select(&subs, "SELECT * FROM notifications_subscriptions WHERE user_id = $1 and validator_publickey = $2", userId, pubKey)
	return subs, err
}

func GetUserSubscriptions(userId uint64) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}
	err := FrontendDB.Select(&subs, "SELECT * FROM notifications_subscriptions WHERE user_id = $1", userId)
	return subs, err
}
