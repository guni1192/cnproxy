package main

import (
	"log"
	"os"

	"github.com/guni1192/cnproxy/pkg/config"
	"github.com/guni1192/cnproxy/pkg/middleware/server"
	"github.com/urfave/cli/v2"
)

func main() {
	var port uint
	var address string
	var enableMetrics bool
	var allowedFQDNs cli.StringSlice
	var configPath string

	app := &cli.App{
		Name:  "cnproxy",
		Usage: "cloud native proxy",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "path to YAML config file",
				Destination: &configPath,
				EnvVars:     []string{"CNPROXY_CONFIG"},
			},
			&cli.UintFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Value:       8080,
				Usage:       "port number",
				Destination: &port,
				EnvVars:     []string{"CNPROXY_PORT"},
			},
			&cli.StringFlag{
				Name:        "address",
				Aliases:     []string{"a"},
				Value:       "0.0.0.0",
				Usage:       "address",
				Destination: &address,
				EnvVars:     []string{"CNPROXY_ADDRESS"},
			},
			&cli.BoolFlag{
				Name:        "enable-metrics",
				Usage:       "enable metrics (OTLP)",
				Value:       false,
				Destination: &enableMetrics,
				EnvVars:     []string{"CNPROXY_ENABLE_METRICS"},
			},
			&cli.StringSliceFlag{
				Name:        "allowed-fqdn",
				Usage:       "allowed FQDNs for proxy connections (can be specified multiple times; '*.example.com' matches any subdomain)",
				Destination: &allowedFQDNs,
				EnvVars:     []string{"CNPROXY_ALLOWED_FQDN"},
			},
		},
		Action: func(c *cli.Context) error {
			cfg := &config.Config{}
			if configPath != "" {
				loaded, err := config.Load(configPath)
				if err != nil {
					return err
				}
				cfg = loaded
			}

			// CLI flags / env vars override config values when explicitly set.
			// IsSet returns false when only the flag's default was used, so
			// untouched defaults don't clobber values from the config file.
			if c.IsSet("port") || cfg.Port == 0 {
				cfg.Port = port
			}
			if c.IsSet("address") || cfg.Address == "" {
				cfg.Address = address
			}
			if c.IsSet("enable-metrics") {
				cfg.EnableMetrics = enableMetrics
			}
			if c.IsSet("allowed-fqdn") {
				cfg.AllowedFQDNs = allowedFQDNs.Value()
			}

			cnproxyServer := &server.CNProxyServer{
				Port:          cfg.Port,
				Address:       cfg.Address,
				EnableMetrics: cfg.EnableMetrics,
				AllowedFQDNs:  cfg.AllowedFQDNs,
				HTTPFilters:   cfg.HTTPFilters,
			}
			return cnproxyServer.Serve()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
