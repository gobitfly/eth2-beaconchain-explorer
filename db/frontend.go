package db

import (
	"database/sql"
	"errors"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
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
	err := FrontendDB.Select(&rows, "SELECT id, email FROM users WHERE id = ANY($1)", pq.Array(ids))
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
	_, err := FrontendDB.Exec("DELETE FROM users WHERE email = $1", email)
	return err
}

func GetUserApiKeyById(id uint64) (string, error) {
	var apiKey string = ""
	err := FrontendDB.Get(&apiKey, "SELECT api_key FROM users WHERE id = $1", id)
	return apiKey, err
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

// AddAuthorizeCode registers a code that can be used in exchange for an access token
func AddAuthorizeCode(userId uint64, code string, appId uint64) error {
	now := time.Now()
	nowTs := now.Unix()
	_, err := FrontendDB.Exec("INSERT INTO oauth_codes (user_id, code, app_id, created_ts) VALUES($1, $2, $3, TO_TIMESTAMP($4))", userId, code, appId, nowTs)
	return err
}

// GetAppNameFromRedirectUri receives an oauth redirect_url and returns the registered app name, if exists
func GetAppDataFromRedirectUri(callback string) (*types.OAuthAppData, error) {
	data := []*types.OAuthAppData{}
	err := FrontendDB.Select(&data, "SELECT id, app_name, redirect_uri, active, owner_id FROM oauth_apps WHERE active = true AND redirect_uri = $1", callback)
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

	tx, err := FrontendDB.Begin()
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
	err := FrontendDB.Select(&rows, "UPDATE oauth_codes SET consumed = true WHERE code = $1 AND "+
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
	err := FrontendDB.Get(&userID,
		"SELECT user_id FROM users_devices WHERE user_id = $1 AND "+
			"refresh_token = $2 AND app_id = $3 AND id = $4 AND active = true", claimUserID, hashedRefreshToken, claimAppID, claimDeviceID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func GetUserDevicesByUserID(userID uint64) ([]types.PairedDevice, error) {
	data := []types.PairedDevice{}

	rows, err := FrontendDB.Query(
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
	err := FrontendDB.Get(&deviceID, "INSERT INTO users_devices (user_id, refresh_token, device_name, app_id, created_ts) VALUES($1, $2, $3, $4, 'NOW()') RETURNING id",
		userID, hashedRefreshToken, name, appID,
	)

	if err != nil {
		return 0, err
	}

	return deviceID, nil
}

func MobileNotificatonTokenUpdate(userID, deviceID uint64, notifyToken string) error {
	_, err := FrontendDB.Exec("UPDATE users_devices SET notification_token = $1 WHERE user_id = $2 AND id = $3;",
		notifyToken, userID, deviceID,
	)
	return err
}

func InsertMobileSubscription(userID uint64, paymentDetails types.MobileSubscription, store, receipt string, expiration int64, rejectReson string) error {
	now := time.Now()
	nowTs := now.Unix()
	_, err := FrontendDB.Exec("INSERT INTO users_app_subscriptions (user_id, product_id, price_micros, currency, created_at, updated_at, validate_remotely, active, store, receipt, expires_at, reject_reason) VALUES("+
		"$1, $2, $3, $4, TO_TIMESTAMP($5), TO_TIMESTAMP($6), $7, $8, $9, $10, TO_TIMESTAMP($11), $12);",
		userID, paymentDetails.ProductID, paymentDetails.PriceMicros, paymentDetails.Currency, nowTs, nowTs, paymentDetails.Valid, paymentDetails.Valid, store, receipt, expiration, rejectReson,
	)
	return err
}

func GetAppSubscriptionCount(userID uint64) (int64, error) {
	var count int64
	row := FrontendDB.QueryRow(
		"SELECT count(receipt) as count FROM users_app_subscriptions WHERE user_id = $1",
		userID,
	)
	err := row.Scan(&count)
	return count, err
}

func GetUserPremiumPackage(userID uint64) (string, error) {
	var pkg string
	row := FrontendDB.QueryRow(
		"SELECT product_id from users_app_subscriptions WHERE user_id = $1 AND active = true order by id desc",
		userID,
	)
	err := row.Scan(&pkg)
	return pkg, err
}

func GetAllAppSubscriptions() ([]*types.PremiumData, error) {
	data := []*types.PremiumData{}

	err := FrontendDB.Select(&data,
		"SELECT id, receipt, store, active from users_app_subscriptions WHERE validate_remotely = true order by id desc",
	)

	return data, err
}

func UpdateUserSubscription(id uint64, valid bool, expiration int64, rejectReason string) error {
	now := time.Now()
	nowTs := now.Unix()
	_, err := FrontendDB.Exec("UPDATE users_app_subscriptions SET active = $1, updated_at = TO_TIMESTAMP($2), expires_at = TO_TIMESTAMP($3), reject_reason = $4 WHERE id = $5;",
		valid, nowTs, expiration, rejectReason, id,
	)
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
	err := FrontendDB.Select(&rows, "SELECT user_id, notification_token FROM users_devices WHERE user_id = ANY($1) AND notify_enabled = true AND active = true AND notification_token IS NOT NULL GROUP BY user_id, notification_token ", pq.Array(ids))
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

	rows, err := FrontendDB.Query("UPDATE users_devices SET "+query+" WHERE user_id = $1 AND id = $2 RETURNING notify_enabled;",
		args...,
	)
	return rows, err
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
	rows, err := FrontendDB.Query("SELECT notify_enabled FROM users_devices WHERE user_id = $1 AND id = $2;",
		userID, deviceID,
	)
	return rows, err
}

func UserClientEntry(userID uint64, clientName string, clientVersion int64, notifyEnabled bool) error {
	var updateClientVersion = ""
	if clientVersion != 0 {
		updateClientVersion = ", client_version = $3"
	}

	_, err := FrontendDB.Exec(
		"INSERT INTO users_clients (user_id, client, client_version, notify_enabled, created_ts) VALUES($1, $2, $3, $4, 'NOW()')"+
			"ON CONFLICT (user_id, client) "+
			"DO UPDATE SET notify_enabled = $4"+updateClientVersion+";",
		userID, clientName, clientVersion, notifyEnabled,
	)

	return err
}

func UserClientDelete(userID uint64, clientName string) error {
	_, err := FrontendDB.Exec("DELETE FROM users_clients WHERE user_id = $1 AND client = $2 ", userID, clientName)
	return err
}
