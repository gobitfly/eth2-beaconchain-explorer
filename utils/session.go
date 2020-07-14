package utils

import (
	"github.com/gorilla/sessions"
)

var SessionStore *sessions.CookieStore

func InitSession(secret string) {
	SessionStore = sessions.NewCookieStore([]byte(secret))
}

// func SetFlash(w http.ResponseWriter, r *http.Request, name string, value string) {
// 	session, err := SessionStore.Get(r, name)
// 	if err != nil {
// 		return
// 	}
// 	session.AddFlash(value)
// 	session.Save(r, w)
// }
//
// func GetFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
// 	session, err := SessionStore.Get(r, name)
// 	if err != nil {
// 		return "", nil
// 	}
// 	fm := session.Flashes()
// 	if fm == nil {
// 		return "", nil
// 	}
// 	session.Save(r, w)
// 	return fmt.Sprintf("%v", fm[0]), nil
// }
