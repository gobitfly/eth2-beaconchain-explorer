package utils

import (
	"testing"
)

func TestIsValidUrl(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"http://foo.com", true},
		{"https://foo.com", true},
		{"https://foo.com/a/b/c", true},
		{"https://foo.com:3333/a/b/c", true},
		{"https://foo.com?hello=a", true},
		{`https://foo.com"`, false},
		{"https://https://https://google.com/", false},
		{"foo.com", false},
		{"asdf qwer", false},
	}
	for _, tt := range tests {
		v := IsValidUrl(tt.url)
		if v != tt.valid {
			t.Errorf("wrong url validation for url %v", tt.url)
		}
	}
}

func TestIsValidWithdrawalCredentials(t *testing.T) {
	tests := []struct {
		cred  string
		valid bool
	}{
		// real world examples (sepolia)
		{"0x020000000000000000000000332e43696a505ef45b9319973785f837ce5267b9", true}, // prefixed
		{"020000000000000000000000332e43696a505ef45b9319973785f837ce5267b9", true},
		{"0x020000000000000000000000388ea662ef2c223ec0b047d41bf3c0f362142ad5", true}, // prefixed
		{"020000000000000000000000388ea662ef2c223ec0b047d41bf3c0f362142ad5", true},
		{"0x01000000000000000000000025c4a76e7d118705e7ea2e9b7d8c59930d8acd3b", true}, // prefixed
		{"01000000000000000000000025c4a76e7d118705e7ea2e9b7d8c59930d8acd3b", true},

		// valid but not real world examples
		{"0x000000000000000000000000332e43696a505ef45b9319973785f837ce5267b9", true},
		{"0x010000000000000000000000332e43696a505ef45b9319973785f837ce5267b9", true},

		// invalid examples
		{"0x030000000000000000000000332e43696a505ef45b9319973785f837ce5267b9", false}, // wrong version
		{"0x010000000000000000000004332e43696a505ef45b9319973785f837ce5267b9", false}, // not enough 0 padding
		{"0x010000000000000000000000332e43696a505ef45b9319973785f837ce5267b", false},  // not enough bytes
		{"0x010000000000000000000000332e43696a505ef45b9319973785f83HALLO0000", false}, // invalid characters
		{"0x332e43696a505ef45b9319973785f837ce5267b96", false},                        // just an address (with prefix)
		{"332e43696a505ef45b9319973785f837ce5267b96", false},                          // just an address (without prefix)
		{"0000000000000000000000332e43696a505ef45b9319973785f837ce5267b9", false},     // just padding and address, no versioning at all
		{"dsasxfafsafass", false}, // random string
		{"", false},               // empty string
	}

	for _, tt := range tests {
		v := IsValidWithdrawalCredentials(tt.cred)
		if v != tt.valid {
			t.Errorf("wrong withdrawal credentials validation for %v", tt.cred)
		}
	}
}
