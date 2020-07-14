package utils

import (
	"github.com/gorilla/sessions"
)

// SessionStore is a securecookie-based session-store.
var SessionStore *sessions.CookieStore

// InitSessionStore initializes SessionStore with the given secret.
func InitSessionStore(secret string) {
	SessionStore = sessions.NewCookieStore([]byte(secret))
}
