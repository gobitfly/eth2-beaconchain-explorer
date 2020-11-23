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

func GetUserIdByStripeId(customerID string) (types.User, error) {
	user := types.User{}
	err := FrontendDB.Get(&user, "SELECT id as user_id FROM users WHERE stripe_customerID = $1", customerID)
	return user, err
}

func GetUserSubscription(id uint64) (types.UserSubscription, error) {
	userSub := types.UserSubscription{}
	err := FrontendDB.Get(&userSub, "SELECT email, stripe_customerID, stripe_subscriptionID, stripe_productID, stripe_active FROM users WHERE id = $1", id)
	return userSub, err
}

func UpdateStripeCustomer(email, customerID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET stripe_customerID = $1 WHERE email = $2", customerID, email)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func UpdateRemoveStripeCustomer(customerID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET stripe_customerID = NULL, stripe_subscriptionID = NULL, stripe_productID = NULL, stripe_active = 'f' WHERE stripe_customerID = $1", customerID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func UpdateAddSubscription(customerID, productID, subscriptionID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET stripe_subscriptionID = $1, stripe_productID = $2 WHERE stripe_customerID = $3", subscriptionID, productID, customerID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func UpdateFulfillOrder(customerID string) (*string, error) {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET stripe_active = 't' WHERE stripe_customerID = $1", customerID)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRow("SELECT stripe_productID FROM users WHERE stripe_customerID = $1", customerID)
	var productID string
	row.Scan(&productID)

	err = tx.Commit()
	return &productID, err
}

func UpdateRemoveSubscription(customerID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET stripe_subscriptionID = NULL, stripe_productID = NULL, stripe_active = 'f' WHERE stripe_customerID = $1", customerID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

// UpdateStripeCustomerID sets the stripe customer id
// func UpdateStripeCustomer(userID uint64, customerID, productID *string) error {
// 	tx, err := FrontendDB.Begin()
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	var subscriptionID *string
// 	row := tx.QueryRow("SELECT stripe_subscriptionID from users WHERE id = $1", userID)

// 	row.Scan(&subscriptionID)

// 	if subscriptionID != nil {
// 		tx.Rollback()
// 		return fmt.Errorf("error customer already has a subscription %v", subscriptionID)
// 	}

// 	_, err = tx.Exec("UPDATE users SET stripe_customerID = $1, stripe_productID = $2 WHERE id = $3", customerID, productID, userID)
// 	if err != nil {
// 		return err
// 	}

// 	err = tx.Commit()
// 	return err
// }

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
