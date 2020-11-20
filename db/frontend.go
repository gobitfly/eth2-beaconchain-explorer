package db

import (
	"errors"
	"eth2-exporter/types"

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

func GetUserApiKeyById(id uint64) (string, error) {
	var apiKey string = ""
	err := FrontendDB.Get(&apiKey, "SELECT api_key FROM users WHERE id = $1", id)
	return apiKey, err
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
