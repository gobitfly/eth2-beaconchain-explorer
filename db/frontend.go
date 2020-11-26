package db

import (
	"database/sql"
	"errors"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"time"

	"github.com/jmoiron/sqlx"
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
	_, err := FrontendDB.Exec("INSERT INTO oauth_codes (user_id, code, app_id, created_ts) VALUES($1, $2, $3, 'now')", userId, code, appId)
	return err
}

// GetAppNameFromRedirectUri receives an oauth redirect_url and returns the registered app name, if exists
func GetAppDataFromRedirectUri(callback string) (*types.OAuthAppData, error) {
	data := []*types.OAuthAppData{}
	FrontendDB.Select(&data, "SELECT id, app_name, redirect_uri, active, owner_id FROM oauth_apps WHERE active = true AND redirect_uri = $1", callback)
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

	key, err := utils.GenerateAPIKey(u.Password, u.Email, string(u.RegisterTs.Unix()))
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
	data := types.OAuthCodeData{
		UserID: 0,
		AppID:  0,
	}
	rows, err := FrontendDB.Query("UPDATE oauth_codes SET consumed = true WHERE code = $1 AND "+
		"consumed = false AND created_ts + INTERVAL '5 minutes' > NOW() "+
		"RETURNING user_id, app_id;", code)

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err := rows.Scan(&data.UserID, &data.AppID)
		if err != nil {
			return nil, err
		}
	}

	if data.UserID > 0 {
		return &data, nil
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

// InsertUserDevice Insert user device and return device id
func InsertUserDevice(userID uint64, hashedRefreshToken string, name string, appID uint64) (uint64, error) {
	var deviceID uint64
	err := FrontendDB.Get(&deviceID, "INSERT INTO users_devices (user_id, refresh_token, device_name, app_id, created_ts) VALUES($1, $2, $3, $4, 'now') RETURNING id",
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

func MobileDeviceSettingsUpdate(userID, deviceID uint64, notifyEnabled bool) (*sql.Rows, error) {
	rows, err := FrontendDB.Query("UPDATE users_devices SET notify_enabled = $1 WHERE user_id = $2 AND id = $3 RETURNING notify_enabled;",
		notifyEnabled, userID, deviceID,
	)
	return rows, err
}

func MobileDeviceSettingsSelect(userID, deviceID uint64) (*sql.Rows, error) {
	rows, err := FrontendDB.Query("SELECT notify_enabled FROM users_devices WHERE user_id = $1 AND id = $2;",
		userID, deviceID,
	)
	return rows, err
}
