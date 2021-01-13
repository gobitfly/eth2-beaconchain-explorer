package db

import (
	"database/sql"
	"encoding/hex"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// AddSubscription adds a new subscription to the database.
func AddSubscription(userID uint64, eventName types.EventName, eventFilter string) error {
	now := time.Now()
	nowTs := now.Unix()
	nowEpoch := utils.TimeToEpoch(now)
	_, err := DB.Exec("INSERT INTO users_subscriptions (user_id, event_name, event_filter, created_ts, created_epoch) VALUES ($1, $2, $3, TO_TIMESTAMP($4), $5) ON CONFLICT DO NOTHING", userID, eventName, eventFilter, nowTs, nowEpoch)
	return err
}

// DeleteSubscription removes a subscription from the database.
func DeleteSubscription(userID uint64, eventName types.EventName, eventFilter string) error {
	_, err := DB.Exec("DELETE FROM users_subscriptions WHERE user_id = $1 and event_name = $2 and event_filter = $3", userID, eventName, eventFilter)
	return err
}

type WatchlistEntry struct {
	UserId              uint64
	Validator_publickey string
}

func AddToWatchlist(watchlist []WatchlistEntry) error {
	qry := ""
	args := make([]interface{}, 0)
	qry += "INSERT INTO users_validators_tags (user_id, validator_publickey, tag) VALUES "

	for _, entry := range watchlist {
		key, err := hex.DecodeString(entry.Validator_publickey)
		if err != nil {
			return err
		}
		// Values
		qry += "("
		args = append(args, entry.UserId)
		qry += fmt.Sprintf("$%v,", len(args))
		args = append(args, key)
		qry += fmt.Sprintf("$%v,", len(args))
		args = append(args, string(types.ValidatorTagsWatchlist))
		qry += fmt.Sprintf("$%v", len(args))
		qry += "),"
	}

	qry = qry[:len(qry)-1] + " ON CONFLICT DO NOTHING;"

	_, err := DB.Exec(qry, args...)
	return err
}

// RemoveFromWatchlist removes a validator for a given user from the users_validators_tag table
// It also deletes any subscriptions for that bookmarked validator
func RemoveFromWatchlist(userId uint64, validator_publickey string) error {
	key, err := hex.DecodeString(validator_publickey)
	if err != nil {
		return err
	}
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM users_subscriptions WHERE user_id = $1 and event_filter = $2", userId, validator_publickey)
	if err != nil {
		return fmt.Errorf("error deleting subscriptions for validator: %v", err)
	}

	_, err = tx.Exec("DELETE FROM users_validators_tags WHERE user_id = $1 and validator_publickey = $2 and tag = $3", userId, key, types.ValidatorTagsWatchlist)
	if err != nil {
		return fmt.Errorf("error deleting validator from watchlist: %v", err)
	}

	err = tx.Commit()

	return err
}

type WatchlistFilter struct {
	Tag            types.Tag
	UserId         uint64
	Validators     *pq.ByteaArray
	JoinValidators bool
}

// GetTaggedValidators returns validaters that were tagged by a user
func GetTaggedValidators(filter WatchlistFilter) ([]*types.TaggedValidators, error) {
	list := []*types.TaggedValidators{}
	args := make([]interface{}, 0)

	// var userId uint64
	// SELECT users_validators_tags.user_id, users_validators_tags.validator_publickey, event_name
	// FROM users_validators_tags inner join users_subscriptions
	// ON users_validators_tags.user_id = users_subscriptions.user_id and ENCODE(users_validators_tags.validator_publickey::bytea, 'hex') = users_subscriptions.event_filter;

	args = append(args, filter.Tag)
	args = append(args, filter.UserId)
	qry := `
		SELECT user_id`

	if filter.JoinValidators {
		qry += ", balance, pubkey, validatorindex"
	}

	qry += `
	FROM users_validators_tags`

	if filter.JoinValidators {
		qry += `
			INNER JOIN validators 
			ON users_validators_tags.validator_publickey = validators.pubkey`
	}

	qry += `
		WHERE tag = $1 AND user_id = $2`
	// select * from users_validators_tags inner join validators on users_validators_tags.validator_publickey = validators.pubkey

	// , ARRAY_AGG(
	// 	SELECT event_name FROM users_subscriptions
	// 	WHERE user_id = $2 and ENCODE(users_validators_tags.validator_publickey::bytea, 'hex') = users_subscriptions.event_filter
	// ) as events

	if filter.Validators != nil {
		args = append(args, *filter.Validators)
		qry += " AND "
		qry += fmt.Sprintf("validator_publickey = ANY($%d)", len(args))
	}

	err := DB.Select(&list, qry, args...)
	return list, err
}

// GetSubscriptionsFilter can be passed to GetSubscriptions() to filter subscriptions.
type GetSubscriptionsFilter struct {
	EventNames    *[]types.EventName
	UserIDs       *[]uint64
	EventFilters  *[]string
	Search        string
	Limit         uint64
	Offset        uint64
	JoinValidator bool
}

// GetSubscriptions returns the subscriptions filtered by the provided filter.
func GetSubscriptions(filter GetSubscriptionsFilter) ([]*types.Subscription, error) {
	subs := []*types.Subscription{}
	qry := "SELECT * FROM users_subscriptions"

	if filter.JoinValidator {
		qry = "SELECT id, user_id, event_name, event_filter, last_sent_ts, created_ts, validators.balance as balance FROM users_subscriptions INNER JOIN validators ON users_subscriptions.event_filter = ENCODE(validators.pubkey::bytea, 'hex')"
	}

	if filter.EventNames == nil && filter.UserIDs == nil && filter.EventFilters == nil {
		err := DB.Select(&subs, qry)
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
	err := DB.Select(&subs, qry, args...)
	return subs, err
}

// UpdateSubscriptionsLastSent upates `last_sent_ts` column of the `users_subscriptions` table.
func UpdateSubscriptionsLastSent(subscriptionIDs []uint64, sent time.Time, epoch uint64) error {
	_, err := DB.Exec(`
		UPDATE users_subscriptions
		SET last_sent_ts = TO_TIMESTAMP($1), last_sent_epoch = $2
		WHERE id = ANY($3)`, sent.Unix(), epoch, pq.Array(subscriptionIDs))
	return err
}

// CountSentMail increases the count of sent mails in the table `mails_sent` for this day.
func CountSentMail(email string) error {
	day := time.Now().Truncate(time.Hour * 24).Unix()
	_, err := FrontendDB.Exec(`
		INSERT INTO mails_sent (email, ts, cnt) VALUES ($1, TO_TIMESTAMP($2), 1)
		ON CONFLICT (email, ts) DO UPDATE SET cnt = mails_sent.cnt+1`, email, day)
	return err
}

// GetMailsSentCount returns the number of sent mails for the day of the passed time.
func GetMailsSentCount(email string, t time.Time) (int, error) {
	day := t.Truncate(time.Hour * 24).Unix()
	count := 0
	err := FrontendDB.Get(&count, "SELECT cnt FROM mails_sent WHERE email = $1 AND ts = TO_TIMESTAMP($2)", email, day)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}
