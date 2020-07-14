package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"time"

	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var userTemplate = template.Must(template.New("login").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user.html"))

var authSessionName = "beaconchain"

func User(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/user",
		},
		Active:                "user",
		Data:                  nil,
		User:                  getUser(w, r),
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

func Register(w http.ResponseWriter, r *http.Request) {
	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("error retrieving session for register route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Errorf("Error parsing form data for register route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	pwd := r.FormValue("password")

	if !utils.IsValidEmail(email) {
		session.AddFlash("Error: Invalid email!")
		err = session.Save(r, w)
		if err != nil {
			logger.Errorf("error saving session data for register route: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var existingEmails int
	err = db.FrontendDB.Get(&existingEmails, "SELECT COUNT(*) FROM users WHERE email = $1", email)
	if existingEmails > 0 {
		session.AddFlash("Error: Email already exists!")
		err = session.Save(r, w)
		if err != nil {
			logger.Errorf("error saving session data for register route: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)
	if err != nil {
		logger.Errorf("error generating hash for password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	registerTs := time.Now().Unix()
	_, err = db.FrontendDB.Exec(`
		INSERT INTO users (password, email, register_ts)
		VALUES ($1, $2, TO_TIMESTAMP($3))`,
		string(pHash), email, registerTs,
	)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			fmt.Println("pq error:", err.Code.Name())
		}
		logger.Errorf("error saving new user into db: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go func(email string) {
		err = sendConfirmationEmail(email)
		if err != nil {
			logger.Errorf("error sending confirmation-email: %v", err)
			return
		}
		logger.Infof("sent confirmation-email")
	}(email)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("Error parsing form data for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("Error retrieving session for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	pwd := r.FormValue("password")

	up := struct {
		ID       int64
		Email    string
		Password string
	}{}

	err = db.FrontendDB.Get(&up, "SELECT id, email, password FROM users WHERE email = $1", email)
	if err != nil {
		logger.Errorf("error retrieving password for user %v: %v", email, err)
		session.AddFlash("Error: Invalid email or password!")

		err = session.Save(r, w)
		if err != nil {
			logger.Errorf("error saving session data for login route: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(up.Password), []byte(pwd))

	if err != nil {
		logger.Errorf("error verifying password for user %v: %v", up.Email, err)
		session.AddFlash("Error: Invalid email or password!")

		err = session.Save(r, w)
		if err != nil {
			logger.Errorf("error saving session data for login route: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session.Values["authenticated"] = true
	session.Values["user_id"] = up.ID

	err = session.Save(r, w)
	if err != nil {
		logger.Errorf("error saving session data for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("error retrieving session for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session.Values["authenticated"] = false
	session.AddFlash("You have been logged out!")

	err = sessions.Save(r, w)
	if err != nil {
		logger.Errorf("error saving session data for logout route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func SendConfirmationEmail(w http.ResponseWriter, r *http.Request) {
	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("error retrieving session for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userID, exists := session.Values["user_id"]
	if !exists {
		return
	}

	var email string
	err = db.FrontendDB.Get(&email, "SELECT email FROM users WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Bad Request: Email does not exist", http.StatusBadRequest)
		return
	}

	err = sendConfirmationEmail(email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session.AddFlash("Email has been sent")
	fmt.Fprintf(w, "Email has been sent")
}

func ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	var userID int64
	err := db.FrontendDB.Get(&userID, "SELECT id FROM users WHERE email_confirmation_hash = $1", hash)
	if err != nil {
		logger.Errorf("error retrieving user-id for confirm-email route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = db.FrontendDB.Exec("UPDATE users SET email_confirmed = 'TRUE' WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getUser(w http.ResponseWriter, r *http.Request) *types.User {
	u := &types.User{}
	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("error getting session from sessionStore: %v", err)
		return u
	}
	u.Flashes = session.Flashes()
	ok := false
	u.Authenticated, ok = session.Values["authenticated"].(bool)
	if !ok {
		u.Authenticated = false
		return u
	}
	u.UserID, ok = session.Values["user_id"].(int64)
	if !ok {
		u.Authenticated = false
		return u
	}
	session.Save(r, w)
	return u
}

func sendConfirmationEmail(email string) error {
	emailConfirmationHashTs := time.Now().Unix()
	emailConfirmationHash := utils.RandomString(40)
	_, err := db.FrontendDB.Exec(`
		UPDATE users 
		SET (email_confirmation_hash, email_confirmation_hash_ts) = ($1, $2)
		WHERE email = $3`, emailConfirmationHash, emailConfirmationHashTs, email)
	if err != nil {
		return err
	}

	subject := "beaconcha.in: Confirm your email-address"
	msg := fmt.Sprintf(`Please confirm your email on https://beaconcha.in by clicking this link:

https://beaconcha.in/confirm/%s

Best regards,

beaconcha.in
`, emailConfirmationHash)
	return utils.SendMail(email, subject, msg)
}

func sendResetEmail(email string) error {
	resetHashTs := time.Now().Unix()
	resetHash := utils.RandomString(40)
	_, err := db.FrontendDB.Exec(`
		UPDATE users 
		SET (password_reset_hash, password_reset_hash_ts) = ($1, $2)
		WHERE email = $3`, resetHash, resetHashTs, email)
	if err != nil {
		return err
	}

	subject := "beaconcha.in: Reset your password"
	msg := fmt.Sprintf(`You can reset your password on https://beaconcha.in by clicking this link:

https://beaconcha.in/reset/%s

Best regards,

beaconcha.in
`, resetHash)
	return utils.SendMail(email, subject, msg)
}
