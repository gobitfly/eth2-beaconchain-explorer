package db

import "eth2-exporter/types"

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

func GetUserProductID(customerID string) (string, error) {
	var productID string
	err := FrontendDB.Get(&productID, "SELECT stripe_productID FROM users WHERE stripe_customerID = $1", customerID)
	return productID, err
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
