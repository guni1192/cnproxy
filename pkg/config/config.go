// Package config loads cnproxy's YAML configuration file.
//
// The schema mirrors the CLI flags so a single config file can replace them:
//
//	port: 8080
//	address: 0.0.0.0
//	enable_metrics: false
//	allowed_fqdns:
//	  - example.com
//	  - "*.example.com"
//	http_filters:
//	  - host: api.example.com
//	    methods: [GET, POST]
//	    paths:
//	      - /v1/users
//	      - /v1/items/*
//
// CONNECT (HTTPS) tunnels carry encrypted payloads, so http_filters apply
// only to plain HTTP requests. CONNECT is gated solely by allowed_fqdns.
package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port          uint         `yaml:"port"`
	Address       string       `yaml:"address"`
	EnableMetrics bool         `yaml:"enable_metrics"`
	AllowedFQDNs  []string     `yaml:"allowed_fqdns"`
	HTTPFilters   []HTTPFilter `yaml:"http_filters"`
}

// HTTPFilter restricts plain HTTP traffic to a host.
//
// An empty Methods means "any method"; an empty Paths means "any path".
// Host supports the same wildcard form as allowed_fqdns ("*.example.com").
// A path ending in "/*" is a prefix match against the portion before "/*";
// any other path must match exactly.
type HTTPFilter struct {
	Host    string   `yaml:"host"`
	Methods []string `yaml:"methods"`
	Paths   []string `yaml:"paths"`
}

// Load reads and parses a YAML config file from path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	cfg := &Config{}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}
