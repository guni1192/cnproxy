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
package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port          uint     `yaml:"port"`
	Address       string   `yaml:"address"`
	EnableMetrics bool     `yaml:"enable_metrics"`
	AllowedFQDNs  []string `yaml:"allowed_fqdns"`
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
