package utils

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

// SessionStore is a securecookie-based session-store.
var SessionStore *sessions.CookieStore

// InitSessionStore initializes SessionStore with the given secret.
func InitSessionStore(secret string) {
	SessionStore = sessions.NewCookieStore([]byte(secret))
	SessionStore.Options.HttpOnly = true
}

func SetFlash(w http.ResponseWriter, r *http.Request, name string, value string) {
	session, err := SessionStore.Get(r, name)
	if err != nil {
		return
	}
	session.AddFlash(value)
	session.Save(r, w)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	session, err := SessionStore.Get(r, name)
	if err != nil {
		return "", nil
	}
	fm := session.Flashes()
	if fm == nil {
		return "", nil
	}
	session.Save(r, w)
	return fmt.Sprintf("%v", fm[0]), nil
}

func GetFlashes(w http.ResponseWriter, r *http.Request, name string) []interface{} {
	session, err := SessionStore.Get(r, name)
	if err != nil {
		return []interface{}{}
	}
	flashes := session.Flashes()
	session.Save(r, w)
	return flashes
}
