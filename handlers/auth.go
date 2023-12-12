package handlers

import (
	"database/sql"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strings"
	"time"

	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var authSessionName = "auth"
var authResetEmailRateLimit = time.Minute * 2
var authConfirmEmailRateLimit = time.Minute * 2
var authEmailExpireTime = time.Minute * 30
var authInternalServerErrorFlashMsg = "Error: Something went wrong :( Please retry later"

// Register handler renders a template that allows for the creation of a new user.
func Register(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "register.html")
	var registerTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "register", "/register", "Register new account", templateFiles)
	data.Data = types.AuthData{Flashes: utils.GetFlashes(w, r, authSessionName), CsrfField: csrf.TemplateField(r)}
	data.Meta.NoTrack = true

	if handleTemplateError(w, r, "auth.go", "Register", "", registerTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// RegisterPost handles the register-formular to register a new user.
func RegisterPost(w http.ResponseWriter, r *http.Request) {
	logger := logger.WithField("route", r.URL.String())
	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
	}

	err = r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	email = strings.ToLower(email)
	pwd := r.FormValue("password")

	if !utils.IsValidEmail(email) {
		session.AddFlash("Error: Invalid email!")
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		logger.Errorf("error creating db-tx for registering user: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	defer tx.Rollback()

	var existingEmails int
	err = tx.Get(&existingEmails, "SELECT COUNT(*) FROM users WHERE LOWER(email) = $1", email)
	if existingEmails > 0 {
		session.AddFlash("Error: Email already exists!")
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	if err != nil {
		logger.Errorf("error retrieving existing emails: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	pHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)
	if err != nil {
		logger.Errorf("error generating hash for password: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	apiKey, err := utils.GenerateRandomAPIKey()
	if err != nil {
		logger.Errorf("error generating hash for api_key: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	_, err = tx.Exec(`
      INSERT INTO users (password, email, register_ts, api_key)
      VALUES ($1, $2, TO_TIMESTAMP($3), $4)`,
		string(pHash), email, time.Now().Unix(), apiKey,
	)

	if err != nil {
		logger.Errorf("error saving new user into db: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.Errorf("error committing db-tx when registering user: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	err = sendConfirmationEmail(email)
	if err != nil {
		logger.Errorf("error sending confirmation-email: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
	} else {
		session.AddFlash("Your account has been created! Please verify your email by clicking the link in the email we just sent you.")
	}

	session.Save(r, w)

	http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
}

// Login handler renders a template that allows a user to login.
func Login(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "login.html")
	var loginTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	q := r.URL.Query()

	data := InitPageData(w, r, "login", "/login", "Login", templateFiles)

	authData := types.AuthData{
		Flashes:      utils.GetFlashes(w, r, authSessionName),
		CsrfField:    csrf.TemplateField(r),
		RecaptchaKey: utils.Config.Frontend.RecaptchaSiteKey,
	}

	redirectData := struct {
		Redirect_uri string
		State        string
	}{
		Redirect_uri: q.Get("redirect_uri"),
		State:        q.Get("state"),
	}

	data.Data = struct {
		AuthData     types.AuthData
		RedirectData interface{}
	}{
		AuthData:     authData,
		RedirectData: redirectData}
	data.Meta.NoTrack = true

	if handleTemplateError(w, r, "auth.go", "Login", "", loginTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// LoginPost handles authenticating the user.
func LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := utils.HandleRecaptcha(w, r, "/login"); err != nil {
		return
	}

	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("Error retrieving session for login route: %v", err)
	}

	err = session.SCS.RenewToken(r.Context())
	if err != nil {
		logger.Errorf("error renewing session token: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	email = strings.ToLower(email)
	pwd := r.FormValue("password")

	user := struct {
		ID        uint64 `db:"id"`
		Email     string `db:"email"`
		Password  string `db:"password"`
		Confirmed bool   `db:"email_confirmed"`
		ProductID string `db:"product_id"`
		Active    bool   `db:"active"`
		UserGroup string `db:"user_group"`
	}{}

	redirectParam := ""
	redirectURI := r.FormValue("oauth_redirect_uri")
	if redirectURI != "" {
		redirectParam = "?redirect_uri=" + redirectURI

		state := r.FormValue("state")
		if state != "" {
			redirectParam += "&state=" + state
		}
	}

	err = db.FrontendWriterDB.Get(&user, `
		WITH
			latest_and_greatest_sub AS (
				SELECT user_id, product_id, active, created_at FROM users_app_subscriptions 
				left join users on users.id = user_id 
				WHERE users.email = $1 AND active = true
				ORDER BY CASE product_id
					WHEN 'whale' THEN 1
					WHEN 'goldfish' THEN 2
					WHEN 'plankton' THEN 3
					ELSE 4  -- For any other product_id values
				END, users_app_subscriptions.created_at DESC LIMIT 1
			)
		SELECT users.id, email, password, email_confirmed, COALESCE(product_id, '') as product_id, COALESCE(active, false) as active, COALESCE(user_group, '') AS user_group 
		FROM users 
		left join latest_and_greatest_sub on latest_and_greatest_sub.user_id = users.id  
		WHERE email = $1`, email)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Errorf("error retrieving password for user %v: %v", email, err)
		}
		session.AddFlash("Error: Invalid email or password!")
		session.Save(r, w)
		http.Redirect(w, r, "/login"+redirectParam, http.StatusSeeOther)
		return
	}

	if !user.Confirmed {
		session.AddFlash("Error: Email has not been confirmed, please click the link in the email we sent you or <a href='/resend'>resend link</a>!")
		session.Save(r, w)
		http.Redirect(w, r, "/login"+redirectParam, http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pwd))
	if err != nil {
		session.AddFlash("Error: Invalid email or password!")
		session.Save(r, w)
		http.Redirect(w, r, "/login"+redirectParam, http.StatusSeeOther)
		return
	}

	if !user.Active {
		user.ProductID = ""
	}

	session.SetValue("authenticated", true)
	session.SetValue("user_id", user.ID)
	session.SetValue("subscription", user.ProductID)
	session.SetValue("user_group", user.UserGroup)

	// save datatable state settings from anon session
	dataTableStatePrefix := "table:state:" + utils.GetNetwork() + ":"

	for k, state := range session.Values() {
		k, ok := k.(string)
		if ok && strings.HasPrefix(k, dataTableStatePrefix) {
			state, ok := state.(types.DataTableSaveState)
			if ok {
				trimK := strings.TrimPrefix(k, dataTableStatePrefix)
				if len(trimK) > 0 {
					err := db.SaveDataTableState(user.ID, trimK, state)
					if err != nil {
						logger.WithError(err).Error("error saving datatable state from session")
					}
				}
			} else {
				logger.Errorf("error could not parse datatable state from session, state: %+v", state)
			}
			session.DeleteValue(k)
		}
	}
	// session.AddFlash("Successfully logged in")

	session.Save(r, w)

	logger.WithFields(
		logrus.Fields{
			"authenticated": session.GetValue("authenticated"),
			"user_id":       session.GetValue("user_id"),
			"subscription":  session.GetValue("subscription"),
			"user_group":    session.GetValue("user_group"),
		},
	).Info("login succeeded")

	if redirectParam != "" {
		http.Redirect(w, r, "/user/authorize"+redirectParam, http.StatusSeeOther)
		return
	}

	// Index(w, r)
	http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
}

// Logout handles ending the user session.
func Logout(w http.ResponseWriter, r *http.Request) {

	_, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session.SetValue("subscription", "")
	session.SetValue("authenticated", false)
	session.DeleteValue("user_id")

	err = session.SCS.Destroy(r.Context())
	if err != nil {
		logger.Errorf("error destroying session tokent user: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ResetPassword renders a template that lets the user reset his password.
// This only works if the hash in the url is correct.
// Will also confirm email of the user if it has not been confirmed yet.
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "resetPassword.html")
	var resetPasswordTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	vars := mux.Vars(r)
	hash := vars["hash"]

	errFields := map[string]interface{}{
		"route": r.URL.String(),
		"hash":  hash,
	}

	// fetch user from db by hash
	user := struct {
		ID        uint64    `db:"id"`
		Confirmed bool      `db:"email_confirmed"`
		ResetTs   time.Time `db:"password_reset_ts"`
	}{}
	err := db.FrontendWriterDB.Get(&user, `
		SELECT users.id, email_confirmed, password_reset_ts
		FROM users 
		WHERE password_reset_hash = $1`, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.SetFlash(w, r, authSessionName, "Error: Invalid or outdated reset link, please retry.")
		} else {
			utils.LogError(err, "error resetting password", 0, errFields)
			utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		}
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	errFields["user_id"] = user.ID

	// check expired ts
	if user.ResetTs.Add(authEmailExpireTime).Before(time.Now()) {

		utils.SetFlash(w, r, authSessionName, "Error: Invalid or outdated reset link, please retry.")
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	// if email is not confirmed, confirm since they clicked a link emailed to them
	if !user.Confirmed {
		_, err = db.FrontendWriterDB.Exec("UPDATE users SET email_confirmed = 'TRUE' WHERE id = $1", user.ID)
		if err != nil {
			utils.LogError(err, "error updating email-confirmation when password reset", 0, errFields)
			utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
			http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
			return
		}
	}

	data := InitPageData(w, r, "requestReset", "/requestReset", "Reset Password", templateFiles)
	data.Data = types.AuthData{Flashes: utils.GetFlashes(w, r, authSessionName), State: hash, CsrfField: csrf.TemplateField(r)}
	data.Meta.NoTrack = true

	if handleTemplateError(w, r, "auth.go", "ResetPassword", "", resetPasswordTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// ResetPasswordPost resets the password to the value provided in the form, given that the hash is correct.
func ResetPasswordPost(w http.ResponseWriter, r *http.Request) {
	hash := r.FormValue("hash")
	errFields := map[string]interface{}{
		"route": r.URL.String(),
		"hash":  hash,
	}

	// fetch user from db by hash
	user := struct {
		ID      uint64    `db:"id"`
		ResetTs time.Time `db:"password_reset_ts"`
	}{}
	err := db.FrontendWriterDB.Get(&user, `
			SELECT users.id, password_reset_ts
			FROM users 
			WHERE password_reset_hash = $1`, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.SetFlash(w, r, authSessionName, "Error: Invalid or outdated reset link, please retry.")
		} else {
			utils.LogError(err, "error resetting password", 0, errFields)
			utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		}
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	errFields["user_id"] = user.ID

	// check expired ts
	if user.ResetTs.Add(authEmailExpireTime).Before(time.Now()) {
		utils.SetFlash(w, r, authSessionName, "Error: Invalid or outdated reset link, please retry.")
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	// update password
	pwd := r.FormValue("password")
	err = db.UpdatePassword(user.ID, pwd)
	if err != nil {
		utils.LogError(err, "error updating password", 0, errFields)
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong updating your password. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>")
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	// purge all sessions for user
	err = purgeAllSessionsForUser(r.Context(), user.ID)
	if err != nil {
		utils.LogError(err, "error purging sessions for user", 0, errFields)
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	session, _ := utils.SessionStore.Get(r, authSessionName)
	err = session.SCS.RenewToken(r.Context())
	if err != nil {
		utils.LogError(err, "error renewing session token", 0, errFields)
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, authSessionName, "Your password has been updated successfully, you can now log in with your new password.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// RequestResetPassword renders a template that lets the user enter his email and request a reset link.
func RequestResetPassword(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "requestResetPassword.html")
	var requestResetPaswordTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "register", "/register", "Reset Password", templateFiles)
	data.Data = types.AuthData{Flashes: utils.GetFlashes(w, r, authSessionName), CsrfField: csrf.TemplateField(r)}
	data.Meta.NoTrack = true

	if handleTemplateError(w, r, "auth.go", "RequestResetPassword", "", requestResetPaswordTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// RequestResetPasswordPost sends a password-reset-link to the provided (via form) email.
func RequestResetPasswordPost(w http.ResponseWriter, r *http.Request) {
	logger := logger.WithField("route", r.URL.String())

	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	email = strings.ToLower(email)

	if !utils.IsValidEmail(email) {
		utils.SetFlash(w, r, authSessionName, "Error: Invalid email address.")
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	var exists int
	err = db.FrontendWriterDB.Get(&exists, "SELECT COUNT(*) FROM users WHERE email = $1", email)
	if err != nil {
		logger.Errorf("error retrieving user-count: %v", err)
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	if exists == 0 {
		utils.SetFlash(w, r, authSessionName, "Error: Email does not exist.")
		http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
		return
	}

	var rateLimitError *types.RateLimitError
	err = sendPasswordResetEmail(email)
	if err != nil && !errors.As(err, &rateLimitError) {
		logger.Errorf("error sending reset-email: %v", err)
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
	} else if err != nil && errors.As(err, &rateLimitError) {
		utils.SetFlash(w, r, authSessionName, fmt.Sprintf("Error: The ratelimit for sending emails has been exceeded, please try again in %v.", err.(*types.RateLimitError).TimeLeft.Round(time.Second)))
	} else {
		utils.SetFlash(w, r, authSessionName, "An email has been sent which contains a link to reset your password.")
	}

	http.Redirect(w, r, "/requestReset", http.StatusSeeOther)
}

// ResendConfirmation handler sends a template for the user to request another confirmation link via email.
func ResendConfirmation(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "resendConfirmation.html")
	var resendConfirmationTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "resendConfirmation", "/resendConfirmation", "Resend Password Reset", templateFiles)
	data.Data = types.AuthData{Flashes: utils.GetFlashes(w, r, authSessionName), CsrfField: csrf.TemplateField(r)}

	if handleTemplateError(w, r, "auth.go", "ResendConfirmation", "", resendConfirmationTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// ResendConfirmationPost handles sending another confirmation email to the user.
func ResendConfirmationPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		http.Redirect(w, r, "/resend", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	email = strings.ToLower(email)

	if !utils.IsValidEmail(email) {
		utils.SetFlash(w, r, authSessionName, "Error: Invalid email!")
		http.Redirect(w, r, "/resend", http.StatusSeeOther)
		return
	}

	var exists int
	err = db.FrontendWriterDB.Get(&exists, "SELECT COUNT(*) FROM users WHERE email = $1", email)
	if err != nil {
		logger.Errorf("error checking if user exists for email-confirmation: %v", err)
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong :( Please retry later")
		http.Redirect(w, r, "/resend", http.StatusSeeOther)
		return
	}

	if exists == 0 {
		utils.SetFlash(w, r, authSessionName, "Error: Email does not exist!")
		http.Redirect(w, r, "/resend", http.StatusSeeOther)
		return
	}

	var rateLimitError *types.RateLimitError
	err = sendConfirmationEmail(email)
	if err != nil && !errors.As(err, &rateLimitError) {
		logger.Errorf("error sending confirmation-email: %v", err)
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
	} else if err != nil && errors.As(err, &rateLimitError) {
		utils.SetFlash(w, r, authSessionName, fmt.Sprintf("Error: The ratelimit for sending emails has been exceeded, please try again in %v.", err.(*types.RateLimitError).TimeLeft.Round(time.Second)))
	} else {
		utils.SetFlash(w, r, authSessionName, "Email has been sent!")
	}

	http.Redirect(w, r, "/resend", http.StatusSeeOther)
}

// ConfirmEmail confirms the email-address of a user.
func ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	var isConfirmed = false
	err := db.FrontendWriterDB.Get(&isConfirmed, `
	SELECT email_confirmed 
	FROM users 
	WHERE email_confirmation_hash = $1
	`, hash)
	if err != nil {
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if isConfirmed {
		utils.SetFlash(w, r, authSessionName, "Error: Email has already been confirmed!")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	res, err := db.FrontendWriterDB.Exec("UPDATE users SET email_confirmed = 'TRUE' WHERE email_confirmation_hash = $1", hash)
	if err != nil {
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	n, err := res.RowsAffected()
	if err != nil {
		utils.SetFlash(w, r, authSessionName, authInternalServerErrorFlashMsg)
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if n == 0 {
		utils.SetFlash(w, r, authSessionName, "Error: Invalid confirmation-link, please retry.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
	}

	utils.SetFlash(w, r, authSessionName, "Your email has been confirmed! You can log in now.")
	http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
}

func sendConfirmationEmail(email string) error {
	now := time.Now()
	emailConfirmationHash := utils.RandomString(40)

	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lastTs *time.Time
	err = tx.Get(&lastTs, "SELECT email_confirmation_ts FROM users WHERE email = $1", email)
	if err != nil {
		return fmt.Errorf("error getting confirmation-ts: %w", err)
	}
	if lastTs != nil && (*lastTs).Add(authConfirmEmailRateLimit).After(now) {
		return &types.RateLimitError{TimeLeft: (*lastTs).Add(authConfirmEmailRateLimit).Sub(now)}
	}

	_, err = tx.Exec("UPDATE users SET email_confirmation_hash = $1 WHERE email = $2", emailConfirmationHash, email)
	if err != nil {
		return fmt.Errorf("error updating confirmation-hash: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db-tx: %w", err)
	}

	subject := fmt.Sprintf("%s: Verify your email-address", utils.Config.Frontend.SiteDomain)
	msg := fmt.Sprintf(`Please verify your email on %[1]s by clicking this link:

https://%[1]s/confirm/%[2]s

Best regards,

%[1]s
`, utils.Config.Frontend.SiteDomain, emailConfirmationHash)
	err = mail.SendTextMail(email, subject, msg, []types.EmailAttachment{})
	if err != nil {
		return err
	}

	_, err = db.FrontendWriterDB.Exec("UPDATE users SET email_confirmation_ts = TO_TIMESTAMP($1) WHERE email = $2", time.Now().Unix(), email)
	if err != nil {
		return fmt.Errorf("error updating confirmation-ts: %w", err)
	}

	return nil
}

func sendPasswordResetEmail(email string) error {
	now := time.Now()
	resetHash := utils.RandomString(40)

	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lastTs *time.Time
	err = tx.Get(&lastTs, "SELECT password_reset_ts FROM users WHERE email = $1", email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error getting reset-ts: %w", err)
	}
	if lastTs != nil && (*lastTs).Add(authResetEmailRateLimit).After(now) {
		return &types.RateLimitError{TimeLeft: (*lastTs).Add(authResetEmailRateLimit).Sub(now)}
	}

	_, err = tx.Exec("UPDATE users SET password_reset_hash = $1, password_reset_ts = TO_TIMESTAMP($2) WHERE email = $3", resetHash, time.Now().Unix(), email)
	if err != nil {
		return fmt.Errorf("error updating reset-hash: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db-tx: %w", err)
	}

	subject := fmt.Sprintf("%s: Reset your password", utils.Config.Frontend.SiteDomain)
	msg := fmt.Sprintf(`To update your password on %[1]s, please click this link:

https://%[1]s/reset/%[2]s
	
This link will expire in 30 minutes.
	
Best regards,

%[1]s
`, utils.Config.Frontend.SiteDomain, resetHash)
	err = mail.SendTextMail(email, subject, msg, []types.EmailAttachment{})
	if err != nil {
		return err
	}

	return nil
}
