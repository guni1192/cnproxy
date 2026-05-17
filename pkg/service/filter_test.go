package service

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/guni1192/cnproxy/pkg/config"
)

func TestIsHTTPRequestAllowed(t *testing.T) {
	filters := []config.HTTPFilter{
		{
			Host:    "api.example.com",
			Methods: []string{"GET", "POST"},
			Paths:   []string{"/v1/users", "/v1/items/*"},
		},
		{
			// Wildcard host, any method, exact path.
			Host:  "*.public.example.com",
			Paths: []string{"/health"},
		},
		{
			// Host-only rule: any method, any path.
			Host: "internal.example.com",
		},
	}

	tests := []struct {
		name   string
		host   string
		method string
		path   string
		want   bool
	}{
		{"method+path match", "api.example.com", "GET", "/v1/users", true},
		{"prefix match", "api.example.com", "POST", "/v1/items/42", true},
		{"prefix matches base without trailing slash", "api.example.com", "GET", "/v1/items", true},
		{"method not in list", "api.example.com", "DELETE", "/v1/users", false},
		{"method case-insensitive", "api.example.com", "get", "/v1/users", true},
		{"path not in list", "api.example.com", "GET", "/v2/users", false},
		{"wildcard host with port", "metrics.public.example.com:8080", "GET", "/health", true},
		{"wildcard host wrong path", "metrics.public.example.com", "GET", "/admin", false},
		{"host-only rule allows anything", "internal.example.com", "DELETE", "/danger", true},
		{"no rule matches host", "unknown.example.com", "GET", "/", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &CNProxyHandler{
				Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
				HTTPFilters: filters,
			}
			r := httptest.NewRequest(tc.method, "http://"+tc.host+tc.path, nil)
			r.Host = tc.host
			if got := h.isHTTPRequestAllowed(r); got != tc.want {
				t.Errorf("isHTTPRequestAllowed(host=%q method=%q path=%q): got %v, want %v",
					tc.host, tc.method, tc.path, got, tc.want)
			}
		})
	}
}

func TestIsHTTPRequestAllowedEmptyFilters(t *testing.T) {
	h := &CNProxyHandler{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	r := httptest.NewRequest("GET", "http://anything.example.com/anywhere", nil)
	if !h.isHTTPRequestAllowed(r) {
		t.Fatal("with no filters configured, any request must be allowed")
	}
}

// TestHandleProxyEnforcesHTTPFilter exercises the full HandleProxy path: an
// allowed FQDN that fails the http_filters rules must get 403 before any
// upstream connection happens.
func TestHandleProxyEnforcesHTTPFilter(t *testing.T) {
	h := &CNProxyHandler{
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedFQDNs: []string{"api.example.com"},
		HTTPFilters: []config.HTTPFilter{
			{Host: "api.example.com", Methods: []string{"GET"}, Paths: []string{"/v1/users"}},
		},
	}

	cases := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"matched rule passes filter (502 from no upstream)", "GET", "/v1/users", http.StatusBadGateway},
		{"method blocked", "DELETE", "/v1/users", http.StatusForbidden},
		{"path blocked", "GET", "/v1/admin", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://api.example.com"+tc.path, nil)
			r.Host = "api.example.com"
			w := httptest.NewRecorder()
			h.HandleProxy(w, r)
			if w.Code != tc.status {
				t.Errorf("status: got %d, want %d (body=%q)", w.Code, tc.status, strings.TrimSpace(w.Body.String()))
			}
		})
	}
}

// TestServeHTTPRoutesAbsoluteFormToProxy guards the routing fix: an HTTP
// proxy request uses absolute-form (RFC 9112 §3.2.2), so r.URL.Path is "/"
// not "". The earlier switch sent it to 404. Now any absolute-form request
// or CONNECT must reach HandleProxy.
func TestServeHTTPRoutesAbsoluteFormToProxy(t *testing.T) {
	h := &CNProxyHandler{
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedFQDNs: []string{"blocked.example.com"},
	}
	// Use a host that's NOT allowed so HandleProxy short-circuits with 403
	// before attempting any upstream dial. If we got 404, the request never
	// reached HandleProxy.
	r := httptest.NewRequest("GET", "http://denied.example.com/some/path", nil)
	r.Host = "denied.example.com"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("absolute-form HTTP proxy request should reach HandleProxy (expected 403, got %d body=%q)",
			w.Code, strings.TrimSpace(w.Body.String()))
	}
}

func TestServeHTTPHealthcheckStillWorks(t *testing.T) {
	h := &CNProxyHandler{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("/health: got %d, want 200", w.Code)
	}
}
