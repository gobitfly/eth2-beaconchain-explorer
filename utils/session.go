package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
)

// SessionStore is a securecookie-based session-store.

type CustomSessionStore struct {
	// TODO: Implement
	SCS *scs.SessionManager
}

func (css *CustomSessionStore) Get(r *http.Request, name string) (*CustomSession, error) {
	// TODO: Implement
	return &CustomSession{
		SCS:       css.SCS,
		ContextFn: r.Context,
	}, nil
}

type CustomSession struct {
	SCS       *scs.SessionManager
	ContextFn func() context.Context
	// TODO: Implement
}

func (cs *CustomSession) Save(r *http.Request, w http.ResponseWriter) error {
	// Not required as sessions are saved on the fly via middleware
	return nil
}

func (cs *CustomSession) SetValue(key string, value interface{}) {
	cs.SCS.Put(cs.ContextFn(), key, value)
}

func (cs *CustomSession) GetValue(key string) interface{} {
	return cs.SCS.Get(cs.ContextFn(), key)
}

func (cs *CustomSession) DeleteValue(key string) {
	cs.SCS.Remove(cs.ContextFn(), key)
}

func (cs *CustomSession) Values() map[interface{}]interface{} {
	r := make(map[interface{}]interface{})

	for _, key := range cs.SCS.Keys(cs.ContextFn()) {
		v := cs.SCS.Get(cs.ContextFn(), key)

		if v != nil {
			r[key] = v
		}
	}
	return r
}

var SessionStore *CustomSessionStore

// InitSessionStore initializes SessionStore with the given secret.
func InitSessionStore(secret string) {

	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", Config.RedisCacheEndpoint)
		},
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = Week
	sessionManager.Cookie.Name = "session_id"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = true

	sessionManager.Store = redisstore.New(pool)

	SessionStore = &CustomSessionStore{
		SCS: sessionManager,
	}
}

func SetFlash(w http.ResponseWriter, r *http.Request, name string, value string) {
	cookie := http.Cookie{Name: "flash", Value: value, Expires: time.Now().Add(5 * time.Minute)}
	http.SetCookie(w, &cookie)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie("flash")
	if err != nil {
		return "", nil
	}
	return cookie.Value, nil
}

func GetFlashes(w http.ResponseWriter, r *http.Request, name string) []interface{} {
	cookie, err := r.Cookie("flash")
	if err != nil || cookie.Value == "" {
		return nil
	}

	defer func() {
		// remove cookie
		cookie = &http.Cookie{Name: "flash", Value: "", Expires: time.Now().Add(-1 * time.Minute)}
		http.SetCookie(w, cookie)
	}()

	return []interface{}{cookie.Value}
}

func HandleRecaptcha(w http.ResponseWriter, r *http.Request, errorRoute string) error {
	if len(Config.Frontend.RecaptchaSecretKey) > 0 && len(Config.Frontend.RecaptchaSiteKey) > 0 {
		recaptchaResponse := r.FormValue("g-recaptcha-response")
		if len(recaptchaResponse) == 0 {
			SetFlash(w, r, "pricing_flash", "Error: Failed to create request")
			logger.Warnf("error no recaptcha response present for route: %v", r.URL.String())
			http.Redirect(w, r, errorRoute, http.StatusSeeOther)
			return fmt.Errorf("no recaptcha")
		}

		valid, err := ValidateReCAPTCHA(recaptchaResponse)
		if err != nil || !valid {
			SetFlash(w, r, "pricing_flash", "Error: Failed to create request")
			logger.Warnf("error validating recaptcha %v route: %v -> %v", recaptchaResponse, r.URL.String(), err)
			http.Redirect(w, r, errorRoute, http.StatusSeeOther)
			return fmt.Errorf("invalid recaptcha")
		}
	}
	return nil
}
