package utils

import (
	"fmt"
	"net/http"
)

// SessionStore is a securecookie-based session-store.

type CustomSessionStore struct {
	// TODO: Implement
}

func (css *CustomSessionStore) Get(r *http.Request, name string) (*CustomSession, error) {
	// TODO: Implement
	return &CustomSession{}, nil
}

type CustomSession struct {
	// TODO: Implement
}

func (cs *CustomSession) AddFlash(value string) {
	// TODO: Implement

}

func (cs *CustomSession) Save(r *http.Request, w http.ResponseWriter) {
	// TODO: Implement
}

func (cs *CustomSession) SetValue(key string, value interface{}) {
	// TODO: Implement
}

func (cs *CustomSession) Flashes() []interface{} {
	// TODO: Implement
	return []interface{}{}
}

func (cs *CustomSession) Values() map[interface{}]interface{} {
	// TODO: Implement
	return map[interface{}]interface{}{}
}

var SessionStore *CustomSessionStore

// InitSessionStore initializes SessionStore with the given secret.
func InitSessionStore(secret string) {
	SessionStore = &CustomSessionStore{}
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
