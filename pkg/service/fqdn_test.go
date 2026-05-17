package service

import (
	"io"
	"log/slog"
	"testing"
)

func TestIsFQDNAllowed(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		host    string
		want    bool
	}{
		{"empty allowlist allows everything", nil, "anywhere.example.com", true},
		{"exact match", []string{"example.com"}, "example.com", true},
		{"exact with port", []string{"example.com"}, "example.com:443", true},
		{"exact no match", []string{"example.com"}, "other.com", false},
		{"wildcard one label", []string{"*.example.com"}, "api.example.com", true},
		{"wildcard deep label", []string{"*.example.com"}, "v1.api.example.com", true},
		{"wildcard does not match apex", []string{"*.example.com"}, "example.com", false},
		{"wildcard does not match sibling", []string{"*.example.com"}, "example.org", false},
		{"wildcard does not match suffix-only", []string{"*.example.com"}, "evilexample.com", false},
		{"wildcard with port", []string{"*.example.com"}, "api.example.com:8443", true},
		{"case insensitive", []string{"Example.com"}, "EXAMPLE.com", true},
		{"case insensitive wildcard", []string{"*.Example.com"}, "API.example.COM", true},
		{"multiple patterns, second matches", []string{"foo.com", "*.example.com"}, "x.example.com", true},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &CNProxyHandler{Logger: logger, AllowedFQDNs: tc.allowed}
			if got := h.isFQDNAllowed(tc.host); got != tc.want {
				t.Errorf("isFQDNAllowed(%q) with allowed=%v: got %v, want %v",
					tc.host, tc.allowed, got, tc.want)
			}
		})
	}
}
