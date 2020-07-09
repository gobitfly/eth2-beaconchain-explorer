package utils

import (
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
)

var flashStore *sessions.CookieStore

func InitFlash(secret string) {
	flashStore = sessions.NewCookieStore([]byte(secret))
}

func SetFlash(w http.ResponseWriter, r *http.Request, name string, value string) {
	session, err := flashStore.Get(r, name)
	if err != nil {
		return
	}
	session.AddFlash(value)
	session.Save(r, w)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	session, err := flashStore.Get(r, name)
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
