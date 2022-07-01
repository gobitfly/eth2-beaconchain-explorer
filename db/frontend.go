package db

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// FrontendWriterDB is a pointer to the auth-database
var FrontendReaderDB *sqlx.DB
var FrontendWriterDB *sqlx.DB

func MustInitFrontendDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig, sessionSecret string) {
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

	key, err := utils.GenerateAPIKey(u.Password, u.Email, fmt.Sprint(u.RegisterTs.Unix()))
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

// GetByRefreshToken basically used to confirm the claimed user id with the refresh token. Returns the userId if successfull
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
	return nil, errors.New("no rows found")
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
	if strings.HasPrefix(string(eventName), "monitoring_") || eventName == types.RocketpoolColleteralMaxReached || eventName == types.RocketpoolColleteralMinReached {
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
		return nil, errors.New("No params for change provided")
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

func CleanupOldMachineStats() error {
	const deleteLIMIT uint64 = 60000 // 200 users make 36000 new inserts per hour

	now := time.Now()
	nowTs := now.Unix()
	var today int = int(nowTs / 86400)

	dayRange := 32
	day := int(today - dayRange)

	deleteCondition := "SELECT COALESCE(min(id), 0) from stats_meta_p where day <= $1"
	deleteConditionGeneral := "SELECT COALESCE(min(id), 0) from stats_process where meta_id <= $1"

	var metaID uint64
	row := FrontendWriterDB.QueryRow(deleteCondition, day)
	err := row.Scan(&metaID)
	if err != nil {
		return err
	}

	var generalID uint64
	row = FrontendWriterDB.QueryRow(deleteConditionGeneral, metaID)
	err = row.Scan(&generalID)
	if err != nil {
		return err
	}
	metaID += deleteLIMIT
	generalID += deleteLIMIT

	tx, err := FrontendWriterDB.Begin()

	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM stats_system WHERE id IN (SELECT id from stats_system where meta_id <= $1 ORDER BY meta_id asc)", metaID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM stats_add_beaconnode WHERE id IN (SELECT id from stats_add_beaconnode WHERE general_id <= $1 ORDER BY general_id asc)", generalID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM stats_add_validator WHERE id IN (SELECT id from stats_add_validator WHERE general_id <= $1 ORDER BY general_id asc)", generalID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM stats_process WHERE id IN (SELECT id FROM stats_process WHERE id <= $1 ORDER BY id asc)", generalID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM stats_meta_p WHERE day < $2 AND id IN (SELECT id from stats_meta_p where day < $2 AND id <= $1 ORDER BY id asc)", metaID, day)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE IF EXISTS stats_meta_" + strconv.Itoa(day-2))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func GetStatsMachineCount(userID uint64) (uint64, error) {
	now := time.Now()
	nowTs := now.Unix()
	var day int = int(nowTs / 86400)

	var count uint64
	row := FrontendWriterDB.QueryRow(
		"SELECT COUNT(DISTINCT sub.machine) as count FROM (SELECT machine from stats_meta_p WHERE day = $2 AND user_id = $1 AND created_trunc + '15 minutes'::INTERVAL > 'now' LIMIT 15) sub",
		userID, day,
	)
	err := row.Scan(&count)
	return count, err
}

func GetStatsMachine(userID uint64) ([]string, error) {
	now := time.Now()
	nowTs := now.Unix()
	var day int = int(nowTs / 86400)
	// log.Println("current day: ", day)
	// for testing
	// day := 18893
	// log.Println("getting machine for day: ", day)

	var machines []string
	err := FrontendWriterDB.Select(&machines,
		"SELECT DISTINCT machine from stats_meta_p WHERE day = $2 AND user_id = $1 LIMIT 300",
		userID, day,
	)
	return machines, err
}

func InsertStatsMeta(tx *sql.Tx, userID uint64, data *types.StatsMeta) (uint64, error) {
	now := time.Now()
	nowTs := now.Unix()
	var day int = int(nowTs / 86400)

	var id uint64
	row := tx.QueryRow(
		"INSERT INTO stats_meta_p (user_id, machine, ts, version, process, created_trunc, exporter_version, day) VALUES($1, $2, TO_TIMESTAMP($3), $4, $5, date_trunc('minute', TO_TIMESTAMP($6)), $7, $8) RETURNING id",
		userID, data.Machine, data.Timestamp, data.Version, data.Process, nowTs, data.ExporterVersion, day,
	)
	err := row.Scan(&id)

	return id, err
}

func CreateNewStatsMetaPartition() error {

	now := time.Now()
	nowTs := now.Unix()
	var day int = int(nowTs / 86400)

	partitionName := "stats_meta_" + strconv.Itoa(day)
	logger.Info("creating new partition table " + partitionName)

	_, err := FrontendWriterDB.Exec("CREATE TABLE " + partitionName + " PARTITION OF stats_meta_p FOR VALUES IN (" + strconv.Itoa(day) + ")")
	if err != nil {
		logger.Errorf("error creating partition %v", err)
		return err
	}
	_, err = FrontendWriterDB.Exec("CREATE UNIQUE INDEX " + partitionName + "_user_id_created_trunc_process_machine_key ON public." + partitionName + " USING btree (user_id, created_trunc, process, machine)")
	if err != nil {
		logger.Errorf("error creating index %v", err)
		return err
	}

	return err
}

func InsertStatsSystem(tx *sql.Tx, meta_id uint64, data *types.StatsSystem) (uint64, error) {
	var id uint64
	row := tx.QueryRow(
		"INSERT INTO stats_system (meta_id, cpu_cores, cpu_threads, cpu_node_system_seconds_total, "+
			"cpu_node_user_seconds_total, cpu_node_iowait_seconds_total, cpu_node_idle_seconds_total,"+
			"memory_node_bytes_total, memory_node_bytes_free, memory_node_bytes_cached, memory_node_bytes_buffers,"+
			"disk_node_bytes_total, disk_node_bytes_free, disk_node_io_seconds, disk_node_reads_total, disk_node_writes_total,"+
			"network_node_bytes_total_receive, network_node_bytes_total_transmit, misc_node_boot_ts_seconds, misc_os"+
			") "+
			"VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20) RETURNING id",
		meta_id, data.CPUCores, data.CPUThreads, data.CPUNodeSystemSecondsTotal, data.CPUNodeUserSecondsTotal,
		data.CPUNodeIowaitSecondsTotal, data.CPUNodeIdleSecondsTotal, data.MemoryNodeBytesTotal, data.MemoryNodeBytesFree,
		data.MemoryNodeBytesCached, data.MemoryNodeBytesBuffers, data.DiskNodeBytesTotal, data.DiskNodeBytesFree,
		data.DiskNodeIoSeconds, data.DiskNodeReadsTotal, data.DiskNodeWritesTotal, data.NetworkNodeBytesTotalReceive,
		data.NetworkNodeBytesTotalTransmit, data.MiscNodeBootTsSeconds, data.MiscOS,
	)
	err := row.Scan(&id)
	return id, err
}

func InsertStatsProcessGeneral(tx *sql.Tx, meta_id uint64, data *types.StatsProcess) (uint64, error) {
	var id uint64
	row := tx.QueryRow(
		"INSERT INTO stats_process (meta_id, cpu_process_seconds_total, memory_process_bytes, client_name, client_version,"+
			"client_build, sync_eth2_fallback_configured,"+
			"sync_eth2_fallback_connected"+
			") "+
			"VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		meta_id, data.CPUProcessSecondsTotal, data.MemoryProcessBytes, data.ClientName, data.ClientVersion, data.ClientBuild,
		data.SyncEth2FallbackConfigured, data.SyncEth2FallbackConnected,
	)
	err := row.Scan(&id)
	return id, err
}

func InsertStatsValidator(tx *sql.Tx, general_id uint64, data *types.StatsAdditionalsValidator) (uint64, error) {
	var id uint64
	_, err := tx.Exec(
		"INSERT INTO stats_add_validator (general_id, validator_total, validator_active) "+
			"VALUES($1, $2, $3)",
		general_id, data.ValidatorTotal, data.ValidatorActive,
	)

	return id, err
}

func InsertStatsBeaconnode(tx *sql.Tx, general_id uint64, data *types.StatsAdditionalsBeaconnode) (uint64, error) {
	var id uint64
	_, err := tx.Exec(
		"INSERT INTO stats_add_beaconnode (general_id, disk_beaconchain_bytes_total, network_libp2p_bytes_total_receive,"+
			"network_libp2p_bytes_total_transmit, network_peers_connected, sync_eth1_connected, sync_eth2_synced,"+
			"sync_beacon_head_slot, sync_eth1_fallback_configured, sync_eth1_fallback_connected"+
			") "+
			"VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		general_id, data.DiskBeaconchainBytesTotal, data.NetworkLibp2pBytesTotalReceive, data.NetworkLibp2pBytesTotalTransmit,
		data.NetworkPeersConnected, data.SyncEth1Connected, data.SyncEth2Synced, data.SyncBeaconHeadSlot, data.SyncEth1FallbackConfigured, data.SyncEth1FallbackConnected,
	)
	return id, err
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

func getMaxDay(limit uint64) int {
	now := time.Now()
	nowTs := now.Unix()
	var day int = int(nowTs / 86400)

	dayRange := int(limit/1440) + 1
	return day - dayRange
}

func GetStatsValidator(userID, limit, offset uint64) (*sql.Rows, error) {
	gapSize := getMachineStatsGap(limit)
	maxDay := getMaxDay(limit)
	row, err := FrontendWriterDB.Query(
		"SELECT t.* FROM (SELECT client_name, client_version, cpu_process_seconds_total, machine, memory_process_bytes, sync_eth2_fallback_configured, sync_eth2_fallback_connected, ts as timestamp, validator_active, validator_total, row_number() OVER(ORDER BY stats_meta_p.id desc) as row FROM stats_add_validator LEFT JOIN stats_process ON stats_add_validator.general_id = stats_process.id "+
			" LEFT JOIN stats_meta_p on stats_process.meta_id = stats_meta_p.id "+
			"WHERE stats_meta_p.day >= $5 AND user_id = $1 AND process = 'validator' ORDER BY stats_meta_p.id desc LIMIT $2 OFFSET $3) t where t.row % $4 = 0",
		userID, limit, offset, gapSize, maxDay,
	)
	return row, err
}

func GetStatsNode(userID, limit, offset uint64) (*sql.Rows, error) {
	gapSize := getMachineStatsGap(limit)
	maxDay := getMaxDay(limit)
	row, err := FrontendWriterDB.Query(
		"SELECT t.* FROM (SELECT client_name, client_version, cpu_process_seconds_total, machine, memory_process_bytes, sync_eth1_fallback_configured, sync_eth1_fallback_connected, sync_eth2_fallback_configured, sync_eth2_fallback_connected, ts as timestamp, disk_beaconchain_bytes_total, network_libp2p_bytes_total_receive, network_libp2p_bytes_total_transmit, network_peers_connected, sync_eth1_connected, sync_eth2_synced, sync_beacon_head_slot, row_number() OVER(ORDER BY stats_meta_p.id desc) as row FROM stats_add_beaconnode left join stats_process on stats_process.id = stats_add_beaconnode.general_id "+
			" LEFT JOIN stats_meta_p on stats_process.meta_id = stats_meta_p.id "+
			"WHERE stats_meta_p.day >= $5 AND user_id = $1 AND process = 'beaconnode' ORDER BY stats_meta_p.id desc LIMIT $2 OFFSET $3) t where t.row % $4 = 0",
		userID, limit, offset, gapSize, maxDay,
	)
	return row, err
}

func GetStatsSystem(userID, limit, offset uint64) (*sql.Rows, error) {
	gapSize := getMachineStatsGap(limit)
	maxDay := getMaxDay(limit)
	row, err := FrontendWriterDB.Query(
		"SELECT t.* FROM (SELECT cpu_cores, cpu_threads, cpu_node_system_seconds_total, cpu_node_user_seconds_total, cpu_node_iowait_seconds_total, cpu_node_idle_seconds_total, memory_node_bytes_total, memory_node_bytes_free, memory_node_bytes_cached, memory_node_bytes_buffers, disk_node_bytes_total, disk_node_bytes_free, disk_node_io_seconds, disk_node_reads_total, disk_node_writes_total, network_node_bytes_total_receive, network_node_bytes_total_transmit, misc_os, misc_node_boot_ts_seconds, ts as timestamp, machine, row_number() OVER(ORDER BY stats_meta_p.id desc) as row from stats_system"+
			" LEFT JOIN stats_meta_p on stats_system.meta_id = stats_meta_p.id "+
			"WHERE stats_meta_p.day >= $5 AND user_id = $1 AND process = 'system' ORDER BY stats_meta_p.id desc LIMIT $2 OFFSET $3) t where t.row % $4 = 0",
		userID, limit, offset, gapSize, maxDay,
	)
	return row, err
}

func GetHistoricPrices(currency string) (map[uint64]float64, error) {
	data := []struct {
		Ts       time.Time
		Currency float64
	}{}

	if currency != "eur" && currency != "usd" && currency != "rub" && currency != "cny" && currency != "cad" && currency != "gbp" {
		return nil, fmt.Errorf("currency %v not supported", currency)
	}

	err := ReaderDb.Select(&data, fmt.Sprintf("SELECT ts, %s AS currency FROM price", currency))
	if err != nil {
		return nil, err
	}

	dataMap := make(map[uint64]float64)
	genesisTime := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)

	for _, d := range data {
		day := uint64(d.Ts.Sub(genesisTime).Hours()) / 24
		dataMap[day] = d.Currency
	}

	return dataMap, nil
}

func GetUserAPIKeyStatistics(apikey *string) (*types.ApiStatistics, error) {
	stats := &types.ApiStatistics{}

	query := fmt.Sprintf(`
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
	)`)

	err := FrontendWriterDB.Get(stats, query, apikey)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func GetSubsForEventFilter(eventName types.EventName) ([][]byte, map[string][]types.Subscription, error) {
	var subs []types.Subscription
	subQuery := `
		SELECT id, user_id, event_filter, last_sent_epoch, created_epoch, event_threshold, ENCODE(unsubscribe_hash, 'hex') as unsubscribe_hash from users_subscriptions where event_name = $1
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
		})

		b, _ := hex.DecodeString(sub.EventFilter)
		filtersEncode = append(filtersEncode, b)
	}
	return filtersEncode, subMap, nil
}
