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

func TestParseUint64Ranges(t *testing.T) {
	tests := []struct {
		in  string
		out []uint64
		err bool
	}{
		{"", []uint64{}, false},
		{"1", []uint64{1}, false},
		{"1,2", []uint64{1, 2}, false},
		{"1-2,4", []uint64{1, 2, 4}, false},
		{"1-4,3,6", []uint64{1, 2, 3, 4, 6}, false},
		{"3,3", []uint64{3}, false},
		{"x", []uint64{}, true},
		{"4-1", []uint64{}, true},
		{"1-2-3", []uint64{}, true},
	}
	for _, tt := range tests {
		out, err := ParseUint64Ranges(tt.in)
		if len(out) != len(tt.out) {
			t.Errorf("wrong output length for input %v: %v != %v", tt.in, out, tt.out)
		}
		for i, v := range out {
			if v != tt.out[i] {
				t.Errorf("wrong output for input %v: %v", tt.in, tt.out)
			}
		}
		if (err != nil) != tt.err {
			t.Errorf("wrong error for input %v: %v", tt.in, tt.err)
		}
	}
}
