package db

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

// FrontendDB is a pointer to the auth-database
var FrontendDB *sqlx.DB

func MustInitFrontendDB(username, password, host, port, name, sessionSecret string) {
	FrontendDB = mustInitDB(username, password, host, port, name)
}

func GetUserEmailById(id int64) (string, error) {
	var mail string = ""

	err := FrontendDB.Get(&mail, `
	SELECT 
		email
	FROM 
		users
	WHERE id = $1`, id)
	if err != nil {
		logger.Errorf("error GetUserEmailById: %v %v", id, err)
		return "", errors.New("Error: Something went wrong.")
	}
	return mail, nil
}

func DeleteUserByEmail(email string) error {
	_, err := FrontendDB.Exec(`
	DELETE 
	FROM 
		users
	WHERE email = $1`, email)
	if err != nil {
		logger.Errorf("error deleting user by email for user: %v %v", email, err)
		return errors.New("Error: Something went wrong.")
	}
	return nil
}

func DeleteUserById(id int64) error {
	_, err := FrontendDB.Exec(`
	DELETE 
	FROM 
		users
	WHERE id = $1`, id)
	if err != nil {
		logger.Errorf("error deleting user by id for user: %v %v", id, err)
		return errors.New("Error: Something went wrong.")
	}
	return nil
}

func UpdatePassword(userId int64, hash []byte) error {
	var GenericUpdatePasswordError string = "Error: Something went wrong updating your password 😕. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>"

	_, err := FrontendDB.Exec("UPDATE users SET password = $1 WHERE id = $2", hash, userId)
	if err != nil {
		logger.Errorf("error updating password for user: %v", err)
		return errors.New(GenericUpdatePasswordError)
	}
	return nil
}

func UpdateEmail(userId int64, email string) error {
	var GenericUpdateEmailError string = "Error: Something went wrong updating your email 😕. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>"

	tx, err := FrontendDB.Beginx()
	if err != nil {
		logger.Errorf("error creating db-tx for registering user: %v", err)
		return errors.New(GenericUpdateEmailError)
	}
	defer tx.Rollback()
	var existingEmails struct {
		emailCount int
		userEmail  string
	}
	err = tx.Get(&existingEmails, "SELECT COUNT(*), email FROM users WHERE email = $1", email)

	if existingEmails.userEmail == email {
		return nil
	} else if existingEmails.emailCount > 0 {
		return errors.New("Error: Email already exists please choose a unique email")
	}

	_, err = tx.Exec(`UPDATE users SET email = $1 WHERE id = $2`, email, userId)
	if err != nil {
		logger.Errorf("error: updating email for user: %v", err)
		return errors.New(GenericUpdateEmailError)
	}
	_, err = tx.Exec(`UPDATE users SET email_confirmed = false WHERE id = $2`, email, userId)
	if err != nil {
		logger.Errorf("error: updating email for user: %v", err)
		return errors.New(GenericUpdateEmailError)
	}
	return nil
}
