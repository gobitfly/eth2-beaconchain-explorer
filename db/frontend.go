package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// FrontendWriterDB is a pointer to the auth-database
var FrontendReaderDB *sqlx.DB
var FrontendWriterDB *sqlx.DB

func MustInitFrontendDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig) {
	FrontendWriterDB, FrontendReaderDB = mustInitDB(writer, reader)
}

// GetUserEmailById returns the email of a user.
func GetUserEmailById(id uint64) (string, error) {
	var mail string = ""
	err := FrontendWriterDB.Get(&mail, "SELECT email FROM users WHERE id = $1", id)
	return mail, err
}

// GetUserEmailsByIds returns the emails of users.
func GetUserEmailsByIds(ids []uint64) (map[uint64]string, error) {
	mailsByID := map[uint64]string{}
	if len(ids) == 0 {
		return mailsByID, nil
	}
	var rows []struct {
		ID    uint64 `db:"id"`
		Email string `db:"email"`
	}
	//
	err := FrontendWriterDB.Select(&rows, "SELECT id, email FROM users WHERE id = ANY($1) AND id NOT IN (SELECT user_id from users_notification_channels WHERE active = false and channel = $2)", pq.Array(ids), types.EmailNotificationChannel)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		mailsByID[r.ID] = r.Email
	}
	return mailsByID, nil
}

// DeleteUserByEmail deletes a user.
func DeleteUserByEmail(email string) error {
	_, err := FrontendWriterDB.Exec("DELETE FROM users WHERE email = $1", email)
	return err
}

func GetUserApiKeyById(id uint64) (string, error) {
	var apiKey string = ""
	err := FrontendWriterDB.Get(&apiKey, "SELECT api_key FROM users WHERE id = $1", id)
	return apiKey, err
}

func GetUserIdByApiKey(apiKey string) (*types.UserWithPremium, error) {
	data := &types.UserWithPremium{}
	row := FrontendWriterDB.QueryRow("SELECT id, (SELECT product_id from users_app_subscriptions WHERE user_id = users.id AND active = true order by id desc limit 1) FROM users WHERE api_key = $1", apiKey)
	err := row.Scan(&data.ID, &data.Product)
	return data, err
}

// DeleteUserById deletes a user.
func DeleteUserById(id uint64) error {
	_, err := FrontendWriterDB.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// UpdatePassword updates the password of a user.
func UpdatePassword(userId uint64, hash []byte) error {
	_, err := FrontendWriterDB.Exec("UPDATE users SET password = $1 WHERE id = $2", hash, userId)
	return err
}

// AddAuthorizeCode registers a code that can be used in exchange for an access token
func AddAuthorizeCode(userId uint64, code, clientId string, appId uint64) error {
	var dbClientID = clientId
	if len(dbClientID) <= 5 { // remain backwards compatible
		dbClientID = code
	}
	now := time.Now()
	nowTs := now.Unix()
	_, err := FrontendWriterDB.Exec("INSERT INTO oauth_codes (user_id, code, app_id, created_ts, client_id) VALUES($1, $2, $3, TO_TIMESTAMP($4), $5) ON CONFLICT (user_id, app_id, client_id) DO UPDATE SET code = $2, created_ts = TO_TIMESTAMP($4), consumed = false", userId, code, appId, nowTs, dbClientID)
	return err
}

// GetAppNameFromRedirectUri receives an oauth redirect_url and returns the registered app name, if exists
func GetAppDataFromRedirectUri(callback string) (*types.OAuthAppData, error) {
	data := []*types.OAuthAppData{}
	err := FrontendWriterDB.Select(&data, "SELECT id, app_name, redirect_uri, active, owner_id FROM oauth_apps WHERE active = true AND redirect_uri = $1", callback)
	if err != nil {
		return nil, err
	}

	if len(data) > 0 {
		return data[0], nil
	}
	return nil, errors.New("no rows found")
}

// CreateAPIKey creates an API key for the user and saves it to the database
func CreateAPIKey(userID uint64) error {
	type user struct {
		Password   string
		RegisterTs time.Time
		Email      string
	}

	tx, err := FrontendWriterDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	u := user{}

	row := tx.QueryRow("SELECT register_ts, password, email FROM users where id = $1", userID)
	err = row.Scan(&u.RegisterTs, &u.Password, &u.Email)
	if err != nil {
		return err
	}

	key, err := utils.GenerateRandomAPIKey()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE users SET api_key = $1 where id = $2", key, userID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetUserAuthDataByAuthorizationCode checks an oauth code for validity, consumes the code and returns the userId on success
func GetUserAuthDataByAuthorizationCode(code string) (*types.OAuthCodeData, error) {
	var rows []*types.OAuthCodeData
	err := FrontendWriterDB.Select(&rows, "UPDATE oauth_codes SET consumed = true WHERE code = $1 AND "+
		"consumed = false AND created_ts + INTERVAL '35 minutes' > NOW() "+
		"RETURNING user_id, app_id;", code)

	if err != nil {
		return nil, err
	}

	for _, r := range rows {
		if r.UserID > 0 {
			return r, nil
		}
	}

	return nil, errors.New("no rows found")
}

// GetByRefreshToken basically used to confirm the claimed user id with the refresh token. Returns the userId if successful
func GetByRefreshToken(claimUserID, claimAppID, claimDeviceID uint64, hashedRefreshToken string) (uint64, error) {
	var userID uint64
	err := FrontendWriterDB.Get(&userID,
		"SELECT user_id FROM users_devices WHERE user_id = $1 AND "+
			"refresh_token = $2 AND app_id = $3 AND id = $4 AND active = true", claimUserID, hashedRefreshToken, claimAppID, claimDeviceID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func GetUserMonitorSharingSetting(userID uint64) (bool, error) {
	var share bool
	err := FrontendWriterDB.Get(&share,
		"SELECT share FROM stats_sharing WHERE user_id = $1 ORDER BY id desc limit 1", userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return share, nil
}

func SetUserMonitorSharingSetting(userID uint64, share bool) error {
	_, err := FrontendWriterDB.Exec("INSERT INTO stats_sharing (user_id, share, ts) VALUES($1, $2, 'NOW()')",
		userID, share,
	)

	return err
}

func GetUserDevicesByUserID(userID uint64) ([]types.PairedDevice, error) {
	data := []types.PairedDevice{}

	rows, err := FrontendWriterDB.Query(
		"SELECT users_devices.id, oauth_apps.app_name, users_devices.device_name, users_devices.active, "+
			"users_devices.notify_enabled, users_devices.created_ts FROM users_devices "+
			"left join oauth_apps on users_devices.app_id = oauth_apps.id WHERE users_devices.user_id = $1 order by created_ts desc", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		pairedDevice := types.PairedDevice{}
		if err := rows.Scan(&pairedDevice.ID, &pairedDevice.AppName, &pairedDevice.DeviceName, &pairedDevice.Active, &pairedDevice.NotifyEnabled, &pairedDevice.CreatedAt); err != nil {
			return nil, err
		}

		data = append(data, pairedDevice)
	}

	if len(data) > 0 {
		return data, nil
	}
	return nil, nil
}

// InsertUserDevice Insert user device and return device id
func InsertUserDevice(userID uint64, hashedRefreshToken string, name string, appID uint64) (uint64, error) {
	var deviceID uint64
	err := FrontendWriterDB.Get(&deviceID, "INSERT INTO users_devices (user_id, refresh_token, device_name, app_id, created_ts) VALUES($1, $2, $3, $4, 'NOW()') RETURNING id",
		userID, hashedRefreshToken, name, appID,
	)

	if err != nil {
		return 0, err
	}

	return deviceID, nil
}

func MobileNotificatonTokenUpdate(userID, deviceID uint64, notifyToken string) error {
	_, err := FrontendWriterDB.Exec("UPDATE users_devices SET notification_token = $1 WHERE user_id = $2 AND id = $3;",
		notifyToken, userID, deviceID,
	)
	return err
}

// AddSubscription adds a new subscription to the database.
func AddSubscription(userID uint64, network string, eventName types.EventName, eventFilter string, eventThreshold float64) error {
	now := time.Now()
	nowTs := now.Unix()
	nowEpoch := utils.TimeToEpoch(now)

	var onConflictDo string = "NOTHING"
	if strings.HasPrefix(string(eventName), "monitoring_") || eventName == types.RocketpoolColleteralMaxReached || eventName == types.RocketpoolColleteralMinReached || eventName == types.ValidatorIsOfflineEventName {
		onConflictDo = "UPDATE SET event_threshold = $6"
	}

	name := string(eventName)
	if network != "" {
		name = strings.ToLower(network) + ":" + string(eventName)
	}
	// channels := pq.StringArray{"email", "push", "webhook"}
	// _, err := FrontendWriterDB.Exec("INSERT INTO users_subscriptions (user_id, event_name, event_filter, created_ts, created_epoch, event_threshold, channels) VALUES ($1, $2, $3, TO_TIMESTAMP($4), $5, $6, $7) ON CONFLICT (user_id, event_name, event_filter) DO "+onConflictDo, userID, name, eventFilter, nowTs, nowEpoch, eventThreshold, channels)
	_, err := FrontendWriterDB.Exec("INSERT INTO users_subscriptions (user_id, event_name, event_filter, created_ts, created_epoch, event_threshold) VALUES ($1, $2, $3, TO_TIMESTAMP($4), $5, $6) ON CONFLICT (user_id, event_name, event_filter) DO "+onConflictDo, userID, name, eventFilter, nowTs, nowEpoch, eventThreshold)
	return err
}

// DeleteSubscription removes a subscription from the database.
func DeleteSubscription(userID uint64, network string, eventName types.EventName, eventFilter string) error {
	name := string(eventName)
	if network != "" && !types.IsUserIndexed(eventName) {
		name = strings.ToLower(network) + ":" + string(eventName)
	}

	_, err := FrontendWriterDB.Exec("DELETE FROM users_subscriptions WHERE user_id = $1 and event_name = $2 and event_filter = $3", userID, name, eventFilter)
	return err
}

func DeleteAllSubscription(userID uint64, network string, eventName types.EventName) error {
	name := string(eventName)
	if network != "" && !types.IsUserIndexed(eventName) {
		name = strings.ToLower(network) + ":" + string(eventName)
	}

	_, err := FrontendWriterDB.Exec("DELETE FROM users_subscriptions WHERE user_id = $1 and event_name = $2", userID, name)
	return err
}

func InsertMobileSubscription(tx *sql.Tx, userID uint64, paymentDetails types.MobileSubscription, store, receipt string, expiration int64, rejectReson string, extSubscriptionId string) error {
	now := time.Now()
	nowTs := now.Unix()
	receiptHash := utils.HashAndEncode(receipt)
	var err error
	if tx == nil {
		_, err = FrontendWriterDB.Exec("INSERT INTO users_app_subscriptions (user_id, product_id, price_micros, currency, created_at, updated_at, validate_remotely, active, store, receipt, expires_at, reject_reason, receipt_hash, subscription_id) VALUES("+
			"$1, $2, $3, $4, TO_TIMESTAMP($5), TO_TIMESTAMP($6), $7, $8, $9, $10, TO_TIMESTAMP($11), $12, $13, $14);",
			userID, paymentDetails.ProductID, paymentDetails.PriceMicros, paymentDetails.Currency, nowTs, nowTs, paymentDetails.Valid, paymentDetails.Valid, store, receipt, expiration, rejectReson, receiptHash, extSubscriptionId,
		)
	} else {
		_, err = tx.Exec("INSERT INTO users_app_subscriptions (user_id, product_id, price_micros, currency, created_at, updated_at, validate_remotely, active, store, receipt, expires_at, reject_reason, receipt_hash, subscription_id) VALUES("+
			"$1, $2, $3, $4, TO_TIMESTAMP($5), TO_TIMESTAMP($6), $7, $8, $9, $10, TO_TIMESTAMP($11), $12, $13, $14);",
			userID, paymentDetails.ProductID, paymentDetails.PriceMicros, paymentDetails.Currency, nowTs, nowTs, paymentDetails.Valid, paymentDetails.Valid, store, receipt, expiration, rejectReson, receiptHash, extSubscriptionId,
		)
	}

	return err
}

func ChangeProductIDFromStripe(tx *sql.Tx, stripeSubscriptionID string, productID string) error {
	now := time.Now()
	nowTs := now.Unix()

	_, err := tx.Exec("UPDATE users_app_subscriptions SET product_id = $2, updated_at = TO_TIMESTAMP($3) where subscription_id = $1 AND store = 'stripe'", stripeSubscriptionID, productID, nowTs)
	if err != nil {
		return err
	}
	return err
}

func GetAppSubscriptionCount(userID uint64) (int64, error) {
	var count int64
	row := FrontendWriterDB.QueryRow(
		"SELECT count(receipt) as count FROM users_app_subscriptions WHERE user_id = $1",
		userID,
	)
	err := row.Scan(&count)
	return count, err
}

type PremiumResult struct {
	Package string `db:"product_id"`
	Store   string `db:"store"`
}

func GetUserPremiumPackage(userID uint64) (PremiumResult, error) {
	var pkg PremiumResult
	err := FrontendWriterDB.Get(&pkg,
		"SELECT COALESCE(product_id, '') as product_id, COALESCE(store, '') as store from users_app_subscriptions WHERE user_id = $1 AND active = true order by id desc",
		userID,
	)
	return pkg, err
}

func GetUserPremiumSubscription(id uint64) (types.UserPremiumSubscription, error) {
	userSub := types.UserPremiumSubscription{}
	err := FrontendWriterDB.Get(&userSub, "SELECT user_id, store, active, COALESCE(product_id, '') as product_id, COALESCE(reject_reason, '') as reject_reason FROM users_app_subscriptions WHERE user_id = $1 ORDER BY active desc, id desc LIMIT 1", id)
	return userSub, err
}

func GetAllAppSubscriptions() ([]*types.PremiumData, error) {
	data := []*types.PremiumData{}

	err := FrontendWriterDB.Select(&data,
		"SELECT id, receipt, store, active, expires_at, product_id from users_app_subscriptions WHERE validate_remotely = true order by id desc",
	)

	return data, err
}

func DisableAllSubscriptionsFromStripeUser(stripeCustomerID string) error {
	userID, err := StripeGetCustomerUserId(stripeCustomerID)
	if err != nil {
		return err
	}

	now := time.Now()
	nowTs := now.Unix()
	_, err = FrontendWriterDB.Exec("UPDATE users_app_subscriptions SET active = $1, updated_at = TO_TIMESTAMP($2), expires_at = TO_TIMESTAMP($3), reject_reason = $4 WHERE user_id = $5 AND store = 'stripe';",
		false, nowTs, nowTs, "stripe_user_deleted", userID,
	)
	return err
}

func GetUserSubscriptionIDByStripe(stripeSubscriptionID string) (uint64, error) {
	var subscriptionID uint64
	row := FrontendWriterDB.QueryRow(
		"SELECT id from users_app_subscriptions WHERE subscription_id = $1",
		stripeSubscriptionID,
	)
	err := row.Scan(&subscriptionID)
	return subscriptionID, err
}

func UpdateUserSubscription(tx *sql.Tx, id uint64, valid bool, expiration int64, rejectReason string) error {
	now := time.Now()
	nowTs := now.Unix()
	var err error
	if tx == nil {
		_, err = FrontendWriterDB.Exec("UPDATE users_app_subscriptions SET active = $1, updated_at = TO_TIMESTAMP($2), expires_at = TO_TIMESTAMP($3), reject_reason = $4 WHERE id = $5;",
			valid, nowTs, expiration, rejectReason, id,
		)
	} else {
		_, err = tx.Exec("UPDATE users_app_subscriptions SET active = $1, updated_at = TO_TIMESTAMP($2), expires_at = TO_TIMESTAMP($3), reject_reason = $4 WHERE id = $5;",
			valid, nowTs, expiration, rejectReason, id,
		)
	}

	return err
}

func SetSubscriptionToExpired(tx *sql.Tx, id uint64) error {
	var err error
	query := "UPDATE users_app_subscriptions SET validate_remotely = false, reject_reason = 'expired' WHERE id = $1;"
	if tx == nil {
		_, err = FrontendWriterDB.Exec(query,
			id,
		)
	} else {
		_, err = tx.Exec(query,
			id,
		)
	}

	return err
}

func GetUserPushTokenByIds(ids []uint64) (map[uint64][]string, error) {
	pushByID := map[uint64][]string{}
	if len(ids) == 0 {
		return pushByID, nil
	}
	var rows []struct {
		ID    uint64 `db:"user_id"`
		Token string `db:"notification_token"`
	}

	err := FrontendWriterDB.Select(&rows, "SELECT DISTINCT ON (user_id, notification_token) user_id, notification_token FROM users_devices WHERE (user_id = ANY($1) AND user_id NOT IN (SELECT user_id from users_notification_channels WHERE active = false and channel = $2)) AND notify_enabled = true AND active = true AND notification_token IS NOT NULL AND LENGTH(notification_token) > 20 ORDER BY user_id, notification_token, id DESC", pq.Array(ids), types.PushNotificationChannel)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		val, ok := pushByID[r.ID]
		if ok {
			pushByID[r.ID] = append(val, r.Token)
		} else {
			pushByID[r.ID] = []string{r.Token}
		}
	}

	return pushByID, nil
}

func MobileDeviceSettingsUpdate(userID, deviceID uint64, notifyEnabled, active string) (*sql.Rows, error) {
	var query = ""
	var args []interface{}

	args = append(args, userID)
	args = append(args, deviceID)

	if notifyEnabled != "" {
		args = append(args, notifyEnabled == "true")
		query = addParamToQuery(query, fmt.Sprintf("notify_enabled = $%d", len(args)))
	}

	if active != "" {
		args = append(args, active == "true")
		query = addParamToQuery(query, fmt.Sprintf("active = $%d", len(args)))
	}

	if query == "" {
		return nil, errors.New("no params for change provided")
	}

	rows, err := FrontendWriterDB.Query("UPDATE users_devices SET "+query+" WHERE user_id = $1 AND id = $2 RETURNING notify_enabled;",
		args...,
	)
	return rows, err
}

func MobileDeviceDelete(userID, deviceID uint64) error {
	_, err := FrontendWriterDB.Exec("DELETE FROM users_devices WHERE user_id = $1 AND id = $2 AND id != 2;", userID, deviceID)
	return err
}

func addParamToQuery(query, param string) string {
	var result = query
	if result != "" {
		result += ","
	}
	result += param
	return result
}

func MobileDeviceSettingsSelect(userID, deviceID uint64) (*sql.Rows, error) {
	rows, err := FrontendWriterDB.Query("SELECT notify_enabled FROM users_devices WHERE user_id = $1 AND id = $2;",
		userID, deviceID,
	)
	return rows, err
}

func NewTransaction() (*sql.Tx, error) {
	return FrontendWriterDB.Begin()
}

func getMachineStatsGap(resultCount uint64) int {
	if resultCount > 20160 { // more than 14 (31)
		return 8
	}
	if resultCount > 10080 { // more than 7 (14)
		return 7
	}
	if resultCount > 2880 { // more than 2 (7)
		return 5
	}
	if resultCount > 1440 { // more than 1 (2)
		return 4
	}
	if resultCount > 770 { // more than 12h
		return 2
	}
	return 1
}

func GetHistoricPrice(currency string, day uint64) (float64, error) {
	if currency == "ETH" {
		currency = "USD"
	}
	currency = strings.ToLower(currency)

	if currency != "eur" && currency != "usd" && currency != "rub" && currency != "cny" && currency != "cad" && currency != "jpy" && currency != "gbp" && currency != "aud" {
		return 0.0, fmt.Errorf("currency %v not supported", currency)
	}

	// Convert day to ts
	genesisTime := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)
	dayStartGenesisTime := time.Date(genesisTime.Year(), genesisTime.Month(), genesisTime.Day(), 0, 0, 0, 0, time.UTC)
	ts := dayStartGenesisTime.Add(utils.Day * time.Duration(day))

	var value float64
	err := ReaderDb.Get(&value, fmt.Sprintf("SELECT %s FROM price WHERE ts = $1", currency), ts)
	if err != nil {
		return 0.0, err
	}
	return value, nil
}

func GetUserAPIKeyStatistics(apikey *string) (*types.ApiStatistics, error) {
	stats := &types.ApiStatistics{}

	query := `
	SELECT (
		SELECT 
			COALESCE(SUM(count), 0) as daily 
		FROM 
			api_statistics 
		WHERE 
			ts > NOW() - INTERVAL '1 day' AND apikey = $1
	), (
		SELECT 
			COALESCE(SUM(count),0) as monthly 
		FROM 
			api_statistics 
		WHERE 
			ts > NOW() - INTERVAL '1 month' AND apikey = $1
	)`

	err := FrontendWriterDB.Get(stats, query, apikey)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func GetSubsForEventFilter(eventName types.EventName) ([][]byte, map[string][]types.Subscription, error) {
	var subs []types.Subscription
	subQuery := `
		SELECT id, user_id, event_filter, last_sent_epoch, created_epoch, event_threshold, ENCODE(unsubscribe_hash, 'hex') as unsubscribe_hash, internal_state from users_subscriptions where event_name = $1
		`

	subMap := make(map[string][]types.Subscription, 0)
	err := FrontendWriterDB.Select(&subs, subQuery, utils.GetNetwork()+":"+string(eventName))
	if err != nil {
		return nil, nil, err
	}

	filtersEncode := make([][]byte, 0, len(subs))
	for _, sub := range subs {
		if _, ok := subMap[sub.EventFilter]; !ok {
			subMap[sub.EventFilter] = make([]types.Subscription, 0)
		}
		subMap[sub.EventFilter] = append(subMap[sub.EventFilter], types.Subscription{
			UserID:         sub.UserID,
			ID:             sub.ID,
			LastEpoch:      sub.LastEpoch,
			EventFilter:    sub.EventFilter,
			CreatedEpoch:   sub.CreatedEpoch,
			EventThreshold: sub.EventThreshold,
			State:          sub.State,
		})

		b, _ := hex.DecodeString(sub.EventFilter)
		filtersEncode = append(filtersEncode, b)
	}
	return filtersEncode, subMap, nil
}

// SaveDataTableState saves the state of the current datatable state update
func SaveDataTableState(user uint64, key string, state types.DataTableSaveState) error {
	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	// check how many table states are stored
	count := 0
	err := FrontendReaderDB.GetContext(ctx, &count, `
		SELECT count(*)
		FROM users_datatable
		WHERE user_id = $1
	`, user)
	if err != nil {
		return err
	}

	// only store the most recent 100 table states across all networks
	if count > 100 {
		_, err := FrontendWriterDB.ExecContext(ctx, `
			DELETE FROM users_datatable 
			WHERE user_id = $1 
			ORDER by updated_at asc 
			LIMIT 10
		`)
		if err != nil {
			return err
		}
	}
	// append network prefix
	key = utils.GetNetwork() + ":" + key

	_, err = FrontendWriterDB.ExecContext(ctx, `
		INSERT INTO 
			users_datatable (user_id, key, state) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (user_id, key) DO UPDATE SET state = $3, updated_at = now()
	`, user, key, state)

	return err
}

// GetDataTablesState retrieves the state for a given user and table
func GetDataTablesState(user uint64, key string) (*types.DataTableSaveState, error) {
	var state types.DataTableSaveState

	// append network prefix
	key = utils.GetNetwork() + ":" + key

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	err := FrontendReaderDB.GetContext(ctx, &state, `
		SELECT state 
		FROM users_datatable
		WHERE user_id = $1 and key = $2
	`, user, key)

	return &state, err
}
