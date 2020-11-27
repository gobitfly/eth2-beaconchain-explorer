package db

import (
	"eth2-exporter/types"
	"fmt"
)

func UpdateRemoveStripeCustomer(customerID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET stripe_customerID = NULL, stripe_subscriptionID = NULL, stripe_priceID = NULL, stripe_active = 'f' WHERE stripe_customerID = $1", customerID)
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

	_, err = tx.Exec("UPDATE users SET stripe_subscriptionID = $1, stripe_priceID = $2 WHERE stripe_customerID = $3", subscriptionID, productID, customerID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func UpdateSubscriptionStatus(customerID string, status bool) error {
	_, err := FrontendDB.Exec("UPDATE users SET stripe_active = $1 WHERE stripe_customerID = $2", status, customerID)
	return err
}

func UpdateActivateSubsciption(customerID string) error {
	return UpdateSubscriptionStatus(customerID, true)
}

func UpdateCancelSubscription(customerID string) error {
	return UpdateSubscriptionStatus(customerID, false)
}

func UpdateRemoveSubscription(customerID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET stripe_subscriptionID = NULL, stripe_priceID = NULL, stripe_active = 'f' WHERE stripe_customerID = $1", customerID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func GetUserIdByStripeId(customerID string) (types.User, error) {
	user := types.User{}
	err := FrontendDB.Get(&user, "SELECT id as user_id FROM users WHERE stripe_customerID = $1", customerID)
	return user, err
}

func GetUserSubscription(id uint64) (types.UserSubscription, error) {
	userSub := types.UserSubscription{}
	err := FrontendDB.Get(&userSub, "SELECT email, stripe_customerID, stripe_subscriptionID, stripe_priceID, stripe_active, api_key FROM users WHERE id = $1", id)
	return userSub, err
}

func GetUserPriceID(customerID string) (*string, error) {
	var priceID *string
	err := FrontendDB.Get(&priceID, "SELECT stripe_priceID FROM users WHERE stripe_customerID = $1", customerID)
	return priceID, err
}

func UpdateStripeCustomer(email, customerID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currID string

	row := tx.QueryRow("SELECT stripe_customerID FROM users WHERE email = $1", email)
	row.Scan(&currID)

	if currID != "" && customerID != currID {
		return fmt.Errorf("error updating stripe customer id, the user already has an id: %v failed to overwrite with: %v", currID, customerID)
	}

	_, err = tx.Exec("UPDATE users SET stripe_customerID = $1 WHERE email = $2", customerID, email)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
