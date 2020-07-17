package handlers

import (
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var userTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/settings.html"))

// UserSettings renders the user-template
func UserSettings(w http.ResponseWriter, r *http.Request) {
	userSettingsData := &types.UserSettingsPageData{}

	// TODO: remove before production
	// userTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/settings.html"))

	w.Header().Set("Content-Type", "text/html")

	user, session, err := getUserSession(w, r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: check if user is authenticated
	if user.Authenticated == false {
		session.Flashes("Error: Please authenticate first")
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email, err := db.GetUserEmailById(user.UserID)
	if err != nil {
		logger.Errorf("Error retrieving the email for user: %v %v", user.UserID, err)
		session.Flashes("Error: Something went wrong.")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

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
	err = userTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func DeleteUserPost(w http.ResponseWriter, r *http.Request) {
	logger = logger.WithField("route", r.URL.String())
	user, session, err := getUserSession(w, r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user.Authenticated == true {
		err := db.DeleteUserById(user.UserID)
		if err != nil {
			logger.Errorf("error deleting user by email for user: %v %v", user.UserID, err)
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			session.Flashes("Error: Could not delete user.")
			session.Save(r, w)
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}

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
		logger.Errorf("error updating password for user: %v", err)
		session.AddFlash("Error: Something went wrong updating your password 😕. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}
	session.AddFlash("Password Updated Successfully ✔️")
	session.Save(r, w)
	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
}

// UpdateEmailPost gets called from the settings page to request a new email update. Only once the update link is pressed does the email actually change.
func UpdateEmailPost(w http.ResponseWriter, r *http.Request) {
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

	var existingEmails struct {
		Count int
		Email string
	}
	err = db.FrontendDB.Get(&existingEmails, "SELECT email FROM users WHERE email = $1", email)

	if existingEmails.Email == email {
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	} else if existingEmails.Email != "" {
		session.AddFlash("Error: Email already exists please choose a unique email")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	var rateLimitError *types.RateLimitError
	err = sendEmailUpdateConfirmation(user.UserID, email)
	if err != nil {
		logger.Errorf("error sending confirmation-email: %v", err)
		if errors.As(err, &rateLimitError) {
			session.AddFlash(fmt.Sprintf("Error: The ratelimit for sending emails has been exceeded, please try again in %v.", err.(*types.RateLimitError).TimeLeft.Round(time.Second)))
		} else {
			session.AddFlash(authInternalServerErrorFlashMsg)
		}
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	session.AddFlash("Verification link sent to your new email " + email)
	session.Save(r, w)
	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
}

// ConfirmUpdateEmail confirms and updates the email address of the user. Given an update link the email in the db is changed.
func ConfirmUpdateEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	_, session, err := getUserSession(w, r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	newEmail, err := url.QueryUnescape(q.Get("email"))
	if err != nil {
		utils.SetFlash(w, r, authSessionName, "Error: Could not update your email please try again.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	user := struct {
		ID        int64     `db:"id"`
		Email     string    `db:"email"`
		ConfirmTs time.Time `db:"email_confirmation_ts"`
		Confirmed bool      `db:"email_confirmed"`
	}{}

	err = db.FrontendDB.Get(&user, "SELECT id, email, email_confirmation_ts, email_confirmed FROM users WHERE email_confirmation_hash = $1", hash)
	if err != nil {
		logger.Errorf("error retreiveing email for confirmation_hash %v %v", hash, err)
		utils.SetFlash(w, r, authSessionName, "Error: Could not Update Email.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if user.Confirmed != true {
		utils.SetFlash(w, r, authSessionName, "Error: Cannot update email for an unconfirmed address.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if user.ConfirmTs.Add(time.Minute * 30).Before(time.Now()) {
		utils.SetFlash(w, r, authSessionName, "Confirmation link has expired.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	var emailExists string
	err = db.FrontendDB.Get(&emailExists, "SELECT email FROM users WHERE email = $1", newEmail)
	if emailExists != "" {
		utils.SetFlash(w, r, authSessionName, "Error: Email already exists. We could not update your email.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	_, err = db.FrontendDB.Exec(`UPDATE users SET email = $1 WHERE id = $2`, newEmail, user.ID)
	if err != nil {
		logger.Errorf("error: updating email for user: %v", err)
		utils.SetFlash(w, r, authSessionName, "Error: Could not Update Email.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	session.Values["authenticated"] = false
	delete(session.Values, "user_id")

	utils.SetFlash(w, r, authSessionName, "Your email has been updated successfully! <br> You can log in with your new email.")
	http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
}

func sendEmailUpdateConfirmation(userId int64, newEmail string) error {
	now := time.Now()
	emailConfirmationHash := utils.RandomString(40)

	tx, err := db.FrontendDB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lastTs *time.Time
	err = tx.Get(&lastTs, "SELECT email_confirmation_ts FROM users WHERE id = $1", userId)
	if err != nil {
		return fmt.Errorf("error getting confirmation-ts: %w", err)
	}
	if lastTs != nil && (*lastTs).Add(authConfirmEmailRateLimit).After(now) {
		return &types.RateLimitError{(*lastTs).Add(authConfirmEmailRateLimit).Sub(now)}
	}

	_, err = tx.Exec("UPDATE users SET email_confirmation_hash = $1 WHERE id = $2", emailConfirmationHash, userId)
	if err != nil {
		return fmt.Errorf("error updating confirmation-hash: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error commiting db-tx: %w", err)
	}

	subject := fmt.Sprintf("%s: Verify your email-address", utils.Config.Frontend.SiteDomain)
	msg := fmt.Sprintf(`To update your email on %[1]s please verify it by clicking this link:

https://%[1]s/updateEmail/%[2]s?email=%[3]s

Best regards,

%[1]s
`, utils.Config.Frontend.SiteDomain, emailConfirmationHash, url.QueryEscape(newEmail))
	err = utils.SendMail(newEmail, subject, msg)
	if err != nil {
		return err
	}

	_, err = db.FrontendDB.Exec("UPDATE users SET email_confirmation_ts = TO_TIMESTAMP($1) WHERE id = $2", time.Now().Unix(), userId)
	if err != nil {
		return fmt.Errorf("error updating confirmation-ts: %w", err)
	}

	return nil
}
