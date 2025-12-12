package main

import (
	"log"
	"os"

	"github.com/guni1192/cnproxy/pkg/middleware/server"
	"github.com/urfave/cli/v2"
)

func main() {
	var port uint
	var address string
	var enableMetrics bool
	var allowedFQDNs cli.StringSlice

	app := &cli.App{
		Name:  "cnproxy",
		Usage: "cloud native proxy",
		Flags: []cli.Flag{
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
				Usage:       "allowed FQDNs for proxy connections (can be specified multiple times)",
				Destination: &allowedFQDNs,
				EnvVars:     []string{"CNPROXY_ALLOWED_FQDN"},
			},
		},
		Action: func(*cli.Context) error {
			cnproxyServer := &server.CNProxyServer{
				Port:          port,
				Address:       address,
				EnableMetrics: enableMetrics,
				AllowedFQDNs:  allowedFQDNs.Value(),
			}
			return cnproxyServer.Serve()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
