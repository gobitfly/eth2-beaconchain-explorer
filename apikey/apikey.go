package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/deatil/go-encoding/base62"
	"github.com/google/uuid"
)

/* APIKey holds the hashed key and metadata for an API key. */

var ErrMissingCredential = errors.New("api key credential is nil")

func NewAPIKeyFromCredentials(name, shortKey string, value HashedKeyCredential) (APIKey, error) {
	if value.IsEmpty() {
		return APIKey{}, ErrMissingCredential
	}

	return APIKey{
		ID:         nil,
		Value:      value[:],
		Name:       name,
		ShortKey:   shortKey,
		CreatedAt:  nil,
		DisabledAt: nil,
		LastUsedAt: nil,
	}, nil
}

func NewAPIKey(name string) (APIKey, RawKeyCredential, error) {
	rawKey, err := NewRawAPIKey()
	if err != nil {
		return APIKey{}, RawKeyCredential{}, err
	}
	shortKey := rawKey.Obfuscate()
	key, err := NewAPIKeyFromCredentials(name, shortKey, rawKey.GetAPIKeyCredential())
	return key, rawKey, err
}

type APIKey struct {
	ID         *uuid.UUID `db:"api_key_id"`
	UserID     uint64     `db:"user_id"`
	Value      []byte     `db:"api_key"`
	Name       string     `db:"name"`
	ShortKey   string     `db:"short_key"`
	CreatedAt  *time.Time `db:"created_at"`
	DisabledAt *time.Time `db:"disabled_at"`
	LastUsedAt *time.Time `db:"last_used_at"`
}

// == Raw Key ==
/*
Raw key is the primitive of an unhashed API key as they passed to the user on creation and
provided on every request. Be mindful of the security implications of dealing with this type.
*/

type RawKeyCredential []byte

const RawKeyEntropyBytes = 32
const MinRawKeyLength = 15

func NewRawAPIKey() (RawKeyCredential, error) {
	b := make([]byte, RawKeyEntropyBytes)
	_, err := rand.Read(b)
	if err != nil {
		return RawKeyCredential{}, err
	}
	return b, nil
}

func RawFromBase62(encoded string) (RawKeyCredential, error) {
	decoded, err := base62.StdEncoding.DecodeString(encoded)
	if err != nil {
		return RawKeyCredential{}, err
	}
	if len(decoded) < MinRawKeyLength { // TODO: once we drop v1 support, enforce a key length of 32 bytes
		return RawKeyCredential{}, fmt.Errorf("invalid raw key length: expected > %d, got %d", MinRawKeyLength, len(decoded))
	}
	return base62.StdEncoding.DecodeString(encoded)
}

// ToBase62 returns the raw API key in Base62 encoding.
// Be mindful of the security implications of exposing raw API keys.
func (r RawKeyCredential) ToBase62() string {
	return base62.StdEncoding.EncodeToString(r.Bytes())
}

func (r RawKeyCredential) Obfuscate() string {
	if len(r) < 6 {
		return "[invalid]"
	}
	base62Encoded := r.ToBase62()
	return fmt.Sprintf("%s...%s", base62Encoded[:3], base62Encoded[len(base62Encoded)-3:])
}

func (r RawKeyCredential) Hash() []byte {
	return HashSHA256(r.Bytes())
}

func (r RawKeyCredential) GetAPIKeyCredential() HashedKeyCredential {
	return NewHashedKeyCredential(r.Hash())
}

// Bytes returns the raw API key bytes.
// Be mindful of the security implications of exposing raw API keys.
func (r RawKeyCredential) Bytes() []byte {
	return r[:]
}

// == Hashed Key ==
/*
Hashed key is the primitive of a hashed version of the raw API key, used for storage and comparison.
*/

type HashedKeyCredential []byte

func NewHashedKeyCredential(raw []byte) HashedKeyCredential {
	return raw
}

func FromBase62(encoded string) (HashedKeyCredential, error) {
	raw, err := RawFromBase62(encoded)
	if err != nil {
		return HashedKeyCredential{}, err
	}
	return raw.GetAPIKeyCredential(), nil
}

func (k HashedKeyCredential) IsEmpty() bool {
	return k.Equal(HashedKeyCredential{})
}

func (k HashedKeyCredential) Equal(other HashedKeyCredential) bool {
	return subtle.ConstantTimeCompare(k.Bytes(), other.Bytes()) == 1
}

func (k HashedKeyCredential) Bytes() []byte {
	return k[:]
}

func (k HashedKeyCredential) String() string {
	return base62.StdEncoding.EncodeToString(k.Bytes())
}

func HashSHA256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
