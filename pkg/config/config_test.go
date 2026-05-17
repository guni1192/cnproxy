package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cnproxy.yaml")
	body := `
port: 9090
address: 127.0.0.1
enable_metrics: true
allowed_fqdns:
  - example.com
  - "*.example.com"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	want := &Config{
		Port:          9090,
		Address:       "127.0.0.1",
		EnableMetrics: true,
		AllowedFQDNs:  []string{"example.com", "*.example.com"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cnproxy.yaml")
	if err := os.WriteFile(path, []byte("typo_field: 1\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load: want error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "typo_field") {
		t.Fatalf("Load: error should mention unknown field, got %v", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatal("Load: want error for missing file, got nil")
	}
}
