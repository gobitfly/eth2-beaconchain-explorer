package apikey

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIKeyFromCredentials_MissingCredential(t *testing.T) {
	name := "test-key"
	shortKey := "short"
	cred := HashedKeyCredential{} // Empty credential

	_, err := NewAPIKeyFromCredentials(name, shortKey, cred)
	if !errors.Is(err, ErrMissingCredential) {
		t.Errorf("expected ErrMissingCredential, got %v", err)
	}
}

// == Raw Key

func TestNewRawAPIKey_UniqueAndLength(t *testing.T) {
	key1, err := NewRawAPIKey()
	if err != nil {
		t.Fatalf("NewRawAPIKey() error: %v", err)
	}
	key2, err := NewRawAPIKey()
	if err != nil {
		t.Fatalf("NewRawAPIKey() error: %v", err)
	}
	if bytes.Equal(key1, key2) {
		t.Error("NewRawAPIKey() should generate unique keys")
	}
	if len(key1.Bytes()) != RawKeyEntropyBytes {
		t.Errorf("Expected key length %d, got %d", RawKeyEntropyBytes, len(key1.Bytes()))
	}
}

func TestRawKeyCredential_ToBase62_And_FromBase62(t *testing.T) {
	key, _ := NewRawAPIKey()
	encoded := key.ToBase62()
	if encoded == "" {
		t.Error("ToBase62() returned empty string")
	}
	decoded, err := RawFromBase62(encoded)
	if err != nil {
		t.Fatalf("FromBase62() error: %v", err)
	}
	if !bytes.Equal(decoded, key) {
		t.Error("Decoded key does not match original")
	}
}

func TestFromBase62_InvalidInput(t *testing.T) {
	_, err := FromBase62("!!!invalidbase62!!!")
	if err == nil {
		t.Error("Expected error for invalid base62 input")
	}
}

func TestRawKeyCredential_Obfuscate(t *testing.T) {
	key, _ := NewRawAPIKey()
	obfuscated := key.Obfuscate()
	if !strings.Contains(obfuscated, "...") {
		t.Errorf("Obfuscate() output missing '...': %s", obfuscated)
	}
	if len(obfuscated) < 9 {
		t.Errorf("Obfuscate() output too short: %s", obfuscated)
	}
	if obfuscated[:3] != key.ToBase62()[:3] {
		t.Errorf("Obfuscate() first three chars do not match: got %s, want %s", obfuscated[:3], key.ToBase62()[:3])
	}
	if obfuscated[len(obfuscated)-3:] != key.ToBase62()[len(key.ToBase62())-3:] {
		t.Errorf("Obfuscate() last three chars do not match: got %s, want %s", obfuscated[len(obfuscated)-3:], key.ToBase62()[len(key.ToBase62())-3:])
	}
}

func TestRawKeyCredential_Bytes(t *testing.T) {
	key, _ := NewRawAPIKey()
	b := key.Bytes()
	if len(b) != RawKeyEntropyBytes {
		t.Errorf("Bytes() length mismatch: got %d, want %d", len(b), RawKeyEntropyBytes)
	}
}

func TestRawKeyCredential_Hash_And_GetAPIKeyCredential(t *testing.T) {
	key, _ := NewRawAPIKey()
	if len(key.Hash()) != RawKeyEntropyBytes {
		t.Errorf("Hash() length mismatch: got %d, want %d", len(key.Hash()), RawKeyEntropyBytes)
	}
	cred := key.GetAPIKeyCredential()
	_ = cred // Just ensure it doesn't panic or error
}

func TestNewRawAPIKey_NotEmpty(t *testing.T) {
	key, err := NewRawAPIKey()
	if err != nil {
		t.Fatalf("NewRawAPIKey() error: %v", err)
	}
	zeroKey := make([]byte, RawKeyEntropyBytes)
	if string(key.Bytes()) == string(zeroKey) {
		t.Error("NewRawAPIKey() generated an all-zero key")
	}
}

// == Hashed Key

func TestNewHashedKeyCredential(t *testing.T) {
	raw := []byte{}
	for i := range raw {
		raw[i] = byte(i)
	}
	cred := NewHashedKeyCredential(raw)
	assert.Equal(t, raw, cred.Bytes())
}

func TestHashedKeyCredential_IsEmpty(t *testing.T) {
	var empty HashedKeyCredential
	assert.True(t, empty.IsEmpty())

	raw := make([]byte, RawKeyEntropyBytes)
	raw[0] = 1
	cred := NewHashedKeyCredential(raw)
	assert.False(t, cred.IsEmpty())
}

func TestHashedKeyCredential_Equal(t *testing.T) {
	raw := make([]byte, RawKeyEntropyBytes)
	for i := range raw {
		raw[i] = byte(i)
	}
	cred1 := NewHashedKeyCredential(raw)
	cred2 := NewHashedKeyCredential(raw)
	assert.True(t, cred1.Equal(cred2))

	raw2 := make([]byte, RawKeyEntropyBytes)
	copy(raw2, raw)
	raw2[0]++
	cred3 := NewHashedKeyCredential(raw2)
	assert.False(t, cred1.Equal(cred3))
}

func TestFromBase62_ValidAndInvalid(t *testing.T) {
	// Valid case: encode a raw key, then decode and hash
	raw, err := NewRawAPIKey()
	assert.NoError(t, err)
	encoded := raw.ToBase62()
	hashed, err := FromBase62(encoded)
	assert.NoError(t, err)
	expected := raw.GetAPIKeyCredential()
	assert.True(t, hashed.Equal(expected), "Hashed credential should match expected")

	// Invalid case: input not base62
	_, err = FromBase62("not_base62!!")
	assert.Error(t, err)

	// Invalid case: input wrong length
	short := "abc"
	_, err = FromBase62(short)
	assert.Error(t, err)
}

func TestLogFormat(t *testing.T) {
	raw := "TDbUML7MRn7PbYBbfLRJebfFWOtYad2CcEt6UIBc4km"
	hashed := "Cvc4Sb5ijkEAWdae2m5gnnZhllQSqm18hdBu1exiJbx"
	key, err := FromBase62(raw)
	assert.NoError(t, err)

	formattedErr := fmt.Errorf("example error with key: %s", key)
	assert.Contains(t, formattedErr.Error(), hashed, "Formatted error should contain the hashed key")
}

// == V1 Compatibility

func TestV1ApiKeyCompatibility(t *testing.T) {
	// this is the base62 encoded raw api key
	apiKey := "XYABABAWJHlORma1bz1" //nolint:gosec

	rawKey, err := RawFromBase62(apiKey)
	require.NoError(t, err)
	require.Equal(t, apiKey, rawKey.ToBase62())

	shortKey := rawKey.Obfuscate()
	newKey, err := NewAPIKeyFromCredentials("test", shortKey, rawKey.GetAPIKeyCredential())
	require.NoError(t, err)

	require.Equal(t, rawKey.Hash(), newKey.Value)

}

func TestMinKeyLength(t *testing.T) {
	apiKey := "XYABABAWJHlORma1bz" //nolint:gosec

	_, err := RawFromBase62(apiKey) // should fail as api key is too short
	require.Error(t, err)
}
