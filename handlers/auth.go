package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"time"

	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var loginTemplate = template.Must(template.New("login").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/login.html"))

var sessionName = "beaconchain"

func Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("Error parsing form data for signup route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	pwd := r.FormValue("password")

	pHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)
	if err != nil {
		logger.Errorf("error generating hash for password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	registerTimestamp := time.Now().Unix()
	emailConfirmationHash := utils.RandomString(40)
	_, err = db.FrontendDB.Exec(`
		INSERT INTO users (password, email, email_confirmed, email_confirmation_hash, register_ts)
		VALUES ($1, $2, 'FALSE', $3, TO_TIMESTAMP($4))`,
		string(pHash), email, emailConfirmationHash, registerTimestamp,
	)
	if err != nil {
		logger.Errorf("error saving new user into db: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("Error parsing form data for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session, err := utils.SessionStore.Get(r, sessionName)
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
	session, err := utils.SessionStore.Get(r, sessionName)
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
	session, err := utils.SessionStore.Get(r, sessionName)
	if err != nil {
		logger.Errorf("error retrieving session for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	session.Values["authenticated"] = false
}

func getUser(w http.ResponseWriter, r *http.Request) *types.User {
	u := &types.User{}
	session, err := utils.SessionStore.Get(r, sessionName)
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

func sendConfirmationEmail() {

}
