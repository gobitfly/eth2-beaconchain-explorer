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
	SCS *scs.SessionManager
}

func (css *CustomSessionStore) Get(r *http.Request, name string) (*CustomSession, error) {
	return &CustomSession{
		SCS: css.SCS,
	}, nil
}

type CustomSession struct {
	SCS *scs.SessionManager
}

func (cs *CustomSession) AddFlash(value string) {
	cs.SCS.Put(context.Background(), "_flash", value)
}

func (cs *CustomSession) Save(r *http.Request, w http.ResponseWriter) error {
	// Not required as sessions are saved on the fly via middleware
	return nil
}

func (cs *CustomSession) SetValue(key string, value interface{}) {
	cs.SCS.Put(context.Background(), key, value)
}

func (cs *CustomSession) GetValue(key string) interface{} {
	return cs.SCS.Get(context.Background(), key)
}

func (cs *CustomSession) DeleteValue(key string) {
	cs.SCS.Remove(context.Background(), key)
}

func (cs *CustomSession) Flashes(vars ...string) []interface{} {
	// TODO: Implement
	key := "_flash"
	if len(vars) > 0 {
		key = vars[0]
	}

	return []interface{}{cs.SCS.Pop(context.Background(), key)}
}

func (cs *CustomSession) Values() map[interface{}]interface{} {
	r := make(map[interface{}]interface{})

	for _, key := range cs.SCS.Keys(context.Background()) {
		v := cs.SCS.Get(context.Background(), key)

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
	sessionManager.Lifetime = 3 * time.Hour
	sessionManager.IdleTimeout = 20 * time.Minute
	sessionManager.Cookie.Name = "session_id"
	// sessionManager.Cookie.Domain = "example.com"
	sessionManager.Cookie.HttpOnly = true
	// sessionManager.Cookie.Path = "/example/"
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	sessionManager.Cookie.Secure = true

	sessionManager.Store = redisstore.NewWithPrefix(pool, fmt.Sprintf("%d:", Config.Chain.Config.DepositChainID))

	SessionStore = &CustomSessionStore{
		SCS: sessionManager,
	}
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
