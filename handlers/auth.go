package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"

	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var loginTemplate = template.Must(template.New("login").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/login.html"))

var sessionName = "beaconchain"

func Signup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("Error parsing form data for signup route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	u := r.FormValue("username")
	p := r.FormValue("password")

	pHash, err := bcrypt.GenerateFromPassword([]byte(p), 10)
	if err != nil {
		logger.Errorf("error generating hash for password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = db.FrontendDB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", u, string(pHash))
	if err != nil {
		logger.Errorf("error saving new user into db: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	session, err := db.SessionStore.Get(r, sessionName)
	if err != nil {
		logger.Errorf("error retrieving session for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	flashes := session.Flashes()
	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/login",
		},
		Active:                "login",
		Data:                  flashes,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

	err = sessions.Save(r, w)
	if err != nil {
		logger.Errorf("error saving session data for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = loginTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("Error parsing form data for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session, err := db.SessionStore.Get(r, sessionName)
	if err != nil {
		logger.Errorf("Error retrieving session for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	u := r.FormValue("username")
	p := r.FormValue("password")

	up := struct {
		Username string
		Password string
	}{}
	err = db.FrontendDB.Get(&up, "SELECT username, password FROM users WHERE username = $1", u)
	if err != nil {
		logger.Errorf("error retrieving password for user %v: %v", u, err)
		session.AddFlash("Invalid username or password!")

		err = sessions.Save(r, w)
		if err != nil {
			logger.Errorf("error saving session data for login route: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(up.Password), []byte(p))

	if err != nil {
		logger.Errorf("error verifying password for user %v: %v", up.Username, err)
		session.AddFlash("Invalid username or password!")

		err = sessions.Save(r, w)
		if err != nil {
			logger.Errorf("error saving session data for login route: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session.Values["authenticated"] = true

	err = sessions.Save(r, w)
	if err != nil {
		logger.Errorf("error saving session data for login route: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := db.SessionStore.Get(r, sessionName)
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
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
