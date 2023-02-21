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
