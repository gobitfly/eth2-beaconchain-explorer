package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

var userTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/settings.html"))

// UserSettings renders the user-template
func UserSettings(w http.ResponseWriter, r *http.Request) {
	userSettingsData := &types.UserSettingsPageData{}

	// TODO: remove before production
	userTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/settings.html"))

	w.Header().Set("Content-Type", "text/html")

	user := getUser(w, r)
	// TODO: check if user is authenticated
	if user.Authenticated == false {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email, _ := db.GetUserEmailById(user.UserID)

	userSettingsData.Email = email
	userSettingsData.Flashes = utils.GetFlashes(w, r, authSessionName)

	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/user",
		},
		Active:                "user",
		Data:                  userSettingsData,
		User:                  user,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}
	err := userTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func DeleteUserPost(w http.ResponseWriter, r *http.Request) {
	logger = logger.WithField("route", r.URL.String())
	user := getUser(w, r)
	if user.Authenticated == true {
		db.DeleteUserById(user.UserID)
		Logout(w, r)
	} else {
		logger.Error("Trying to delete a unauthenticated user")
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}
}

func UpdatePasswordPost(w http.ResponseWriter, r *http.Request) {
	var GenericUpdatePasswordError string = "Error: Something went wrong updating your password 😕. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>"

	user, session, err := getUserSession(w, r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pwdNew := r.FormValue("password")
	pwdOld := r.FormValue("old-password")

	currentUser := struct {
		ID        int64  `db:"id"`
		Email     string `db:"email"`
		Password  string `db:"password"`
		Confirmed bool   `db:"email_confirmed"`
	}{}

	err = db.FrontendDB.Get(&currentUser, "SELECT id, email, password, email_confirmed FROM users WHERE id = $1", user.UserID)
	if err != nil {
		logger.Errorf("error retrieving password for user %v: %v", user.UserID, err)
		session.AddFlash("Error: Invalid password!")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	if !currentUser.Confirmed {
		session.AddFlash("Error: Email has not been comfirmed, please click the link in the email we sent you or <a href='/resend'>resend link</a>!")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(pwdOld))
	if err != nil {
		logger.Errorf("error verifying password for user %v: %v", currentUser.Email, err)
		session.AddFlash("Error: Invalid password!")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	pHash, err := bcrypt.GenerateFromPassword([]byte(pwdNew), 10)
	if err != nil {
		logger.Errorf("error generating hash for password: %v", err)
		session.AddFlash(GenericUpdatePasswordError)
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	err = db.UpdatePassword(user.UserID, pHash)
	if err != nil {
		session.AddFlash(err.Error)
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}
	session.AddFlash("Password Updated Successfully ✔️")
	session.Save(r, w)
	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
}

func UpdateEmailPost(w http.ResponseWriter, r *http.Request) {
	_, session, err := getUserSession(w, r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}
	email := r.FormValue("email")

	if !utils.IsValidEmail(email) {
		session.AddFlash("Error: Invalid email format!")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	// var GenericUpdateEmailError string = "Error: Something went wrong updating your email 😕. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>"

	// tx, err := db.FrontendDB.Beginx()
	// if err != nil {
	// 	logger.Errorf("error creating db-tx for registering user: %v", err)
	// 	session.AddFlash(GenericUpdateEmailError)
	// }
	// defer tx.Rollback()
	var existingEmails struct {
		Count int
		Email string
	}
	err = db.FrontendDB.Get(&existingEmails, "SELECT COUNT(*), email FROM users WHERE email = $1", email)
	// tx.Commit()

	if existingEmails.Email == email {
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	} else if existingEmails.Count > 0 {
		session.AddFlash("Error: Email already exists please choose a unique email")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	session.AddFlash("Verification link sent to your new email <i>" + email + "</i>")
	session.Save(r, w)
	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
}

func ConfirmUpdateEmail(w http.ResponseWriter, r *http.Request) {
	// _, err = tx.Exec(`UPDATE users SET email = $1 WHERE id = $2`, email, user.UserID)
	// if err != nil {
	// 	logger.Errorf("error: updating email for user: %v", err)
	// 	session.AddFlash(GenericUpdateEmailError)
	// 	session.Save(r, w)
	// 	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
	// 	return
	// }
	// _, err = tx.Exec(`UPDATE users SET email_confirmed = false WHERE id = $2`, email, user.UserID)
	// if err != nil {
	// 	logger.Errorf("error: updating email for user: %v", err)
	// 	session.AddFlash(GenericUpdateEmailError)
	// 	session.Save(r, w)
	// 	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
	// 	return
	// }

}
