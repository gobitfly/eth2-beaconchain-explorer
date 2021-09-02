package db

import (
	"encoding/json"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"time"
)

// StripeRemoveCustomer removes the stripe customer and sets all subscriptions to inactive
func StripeRemoveCustomer(customerID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	userID, err := StripeGetCustomerUserId(customerID)
	if err == nil {
		now := time.Now()
		nowTs := now.Unix()
		_, err = tx.Exec("UPDATE users_app_subscriptions SET active = $1, updated_at = TO_TIMESTAMP($2), expires_at = TO_TIMESTAMP($3), reject_reason = $4 WHERE user_id = $5 AND store = 'stripe';",
			false, nowTs, nowTs, "stripe_user_deleted", userID,
		)
	} else {
		// logg & continue anyway
		logger.WithError(err).Error("error could not disable stripe mobile subs: " + customerID + "err: ")
	}

	// remove customer id entry from database
	_, err = tx.Exec("UPDATE users SET stripe_customer_id = NULL WHERE stripe_customer_id = $1", customerID)
	if err != nil {
		return err
	}

	// set all subscriptions to inactive for the deleted stripe customer
	_, err = tx.Exec("UPDATE users_stripe_subscriptions SET active = 'f' WHERE stripe_customer_id = $1", customerID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

// StripeCreateSubscription inserts a new subscription
func StripeCreateSubscription(customerID, priceID, subscriptionID string, payload json.RawMessage) error {
	purchaseGroup := utils.GetPurchaseGroup(priceID)
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO users_stripe_subscriptions (subscription_id, customer_id, price_id, active, payload, purchase_group) VALUES ($1, $2, $3, 'f', $4, $5)", subscriptionID, customerID, priceID, payload, purchaseGroup)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

// StripeUpdateSubscription inserts a new subscription
func StripeUpdateSubscription(priceID, subscriptionID string, payload json.RawMessage) error {
	purchaseGroup := utils.GetPurchaseGroup(priceID)
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users_stripe_subscriptions SET price_id = $2, purchase_group = $4, payload = $3 where subscription_id = $1", subscriptionID, priceID, payload, purchaseGroup)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

// StripeUpdateSubscriptionStatus sets the status of a subscription
func StripeUpdateSubscriptionStatus(id string, status bool, payload *json.RawMessage) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if payload == nil {
		_, err = tx.Exec("UPDATE users_stripe_subscriptions SET active = $2 WHERE subscription_id = $1", id, status)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec("UPDATE users_stripe_subscriptions SET active = $2, payload = $3 WHERE subscription_id = $1", id, status, payload)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

// StripeGetUserAPISubscription returns a users current subscription
func StripeGetUserSubscription(id uint64, purchaseGroup string) (types.UserSubscription, error) {
	userSub := types.UserSubscription{}
	err := FrontendDB.Get(&userSub, "SELECT id, email, stripe_customer_id, subscription_id, price_id, active, api_key FROM users LEFT JOIN (SELECT * FROM users_stripe_subscriptions WHERE purchase_group = $2 and (payload->'ended_at')::text = 'null') as us ON users.stripe_customer_id = us.customer_id WHERE users.id = $1 ORDER BY active desc LIMIT 1", id, purchaseGroup)
	return userSub, err
}

// StripeGetSubscription returns a subscription given a subscription_id
func StripeGetSubscription(id string) (*types.StripeSubscription, error) {
	sub := types.StripeSubscription{}
	err := FrontendDB.Get(&sub, "SELECT customer_id, subscription_id, price_id, active FROM users_stripe_subscriptions WHERE subscription_id = $1", id)
	return &sub, err
}

// StripeUpdateCustomerID adds a stripe customer id to a user. It checks if the user already has a stripe customer id.
func StripeUpdateCustomerID(email, customerID string) error {
	tx, err := FrontendDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currID string

	row := tx.QueryRow("SELECT stripe_customer_id FROM users WHERE email = $1", email)
	row.Scan(&currID)

	// customer already exists
	if currID == customerID {
		return nil
	}

	// user already has a customer id
	if currID != "" && customerID != currID {
		return fmt.Errorf("error updating stripe customer id, the user already has an id: %v failed to overwrite with: %v", currID, customerID)
	}

	_, err = tx.Exec("UPDATE users SET stripe_customer_id = $1 WHERE email = $2", customerID, email)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

// StripeGetCustomerEmail returns a customers email given their customerID
func StripeGetCustomerEmail(customerID string) (string, error) {
	email := ""
	err := FrontendDB.Get(&email, "SELECT email FROM users WHERE stripe_customer_id = $1", customerID)
	return email, err
}

func StripeGetCustomerUserId(customerID string) (uint64, error) {
	var id uint64 = 0
	err := FrontendDB.Get(&id, "SELECT id FROM users WHERE stripe_customer_id = $1", customerID)
	return id, err
}
