package db

import (
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
	return mail, err
}

func DeleteUserByEmail(email string) error {
	_, err := FrontendDB.Exec(`
	DELETE 
	FROM 
		users
	WHERE email = $1`, email)
	return err
}

func DeleteUserById(id int64) error {
	_, err := FrontendDB.Exec(`
	DELETE 
	FROM 
		users
	WHERE id = $1`, id)
	return err
}

func UpdatePassword(userId int64, hash []byte) error {
	_, err := FrontendDB.Exec("UPDATE users SET password = $1 WHERE id = $2", hash, userId)
	return err
}
