package db

import (
	"encoding/hex"
	"eth2-exporter/types"
	"fmt"
	"strings"
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

// AddSubscription adds a new subscription to the database. It checkes if the subscription already exists before inserting.
func AddSubscription(userID uint64, eventName string, validatorPublickey *string) error {

	event, err := types.EventFromString(eventName)
	if err != nil {
		return err
	}

	filter := GetSubscriptionsFilter{
		EventNames: &[]types.EventName{event},
		UserIDs:    &[]uint64{userID},
	}

	if validatorPublickey != nil {
		filter.ValidatorPublicKeys = &[]string{*validatorPublickey}
	}

	subs, err := GetSubscriptions(filter)
	if err != nil {
		return err
	}

	if len(subs) != 0 {
		return errors.Errorf("This subscription already exist. user: %v, event: %v validator: %v", userID, eventName, *validatorPublickey)
	}

	if validatorPublickey == nil {
		_, err = FrontendDB.Exec("INSERT INTO notifications_subscriptions (user_id, event_name) VALUES ($1, $2)", userID, eventName)
		return err
	}
	pubKey, err := hex.DecodeString(*validatorPublickey)
	if err != nil {
		return err
	}
	_, err = FrontendDB.Exec("INSERT INTO notifications_subscriptions (user_id, event_name, validator_publickey) VALUES ($1, $2, $3)", userID, eventName, pubKey)
	return err
}

// DeleteSubscription removes a subscription from the database.
func DeleteSubscription(userID uint64, eventName types.EventName, validatorPublickey *string) error {
	if validatorPublickey == nil {
		_, err := FrontendDB.Exec("DELETE FROM notifications_subscriptions WHERE user_id = $1 and event_name = $2 and validator_publickey IS NULL", userID, eventName)
		return err
	}
	pubKey, err := hex.DecodeString(*validatorPublickey)
	if err != nil {
		return err
	}

	_, err = FrontendDB.Exec("DELETE FROM notifications_subscriptions WHERE user_id = $1 and event_name = $2 and validator_publickey = $3", userID, eventName, pubKey)
	return err
}

type GetSubscriptionsFilter struct {
	EventNames          *[]types.EventName
	UserIDs             *[]uint64
	ValidatorPublicKeys *[]string
}

func GetSubscriptions(filter GetSubscriptionsFilter) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}
	qry := "SELECT id, user_id, event_name, encode(validator_publickey::bytea, 'hex') as validator_publickey, last_notification_ts FROM notifications_subscriptions"
	if filter.EventNames == nil || filter.UserIDs == nil || filter.ValidatorPublicKeys == nil {
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

	if filter.ValidatorPublicKeys != nil {
		filters = append(filters, fmt.Sprintf("encode(validator_publickey::bytea, 'hex') = ANY($%d)", len(filters)+1))
		args = append(args, pq.Array(*filter.ValidatorPublicKeys))
	}

	qry += " WHERE " + strings.Join(filters, " AND ")

	err := FrontendDB.Select(&subs, qry, args...)
	return subs, err
}
