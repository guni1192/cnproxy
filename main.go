package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/guni1192/cnproxy/pkg/middleware/opentelemetry"
	"github.com/guni1192/cnproxy/pkg/middleware/server"
	"github.com/urfave/cli/v2"

	"go.opentelemetry.io/otel"
)

func main() {
	var port uint
	var address string
	var enableOtel bool

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
			},
			&cli.StringFlag{
				Name:        "address",
				Aliases:     []string{"a"},
				Value:       "0.0.0.0",
				Usage:       "address",
				Destination: &address,
			},
			&cli.BoolFlag{
				Name:        "enable-otel",
				Usage:       "enable opentelemetry",
				Value:       false,
				Destination: &enableOtel,
			},
		},
		Action: func(*cli.Context) error {
			ctx := context.Background()

			if enableOtel {
				conn, err := opentelemetry.Connect()
				if err != nil {
					return fmt.Errorf("failed to connect otlp server: %v", err)
				}
				shutdownMetricsProvider, err := opentelemetry.SetupMetricsProvider(ctx, nil, conn)
				if err != nil {
					return fmt.Errorf("failed to setup metrics provider: %v", err)
				}
				defer shutdownMetricsProvider(ctx)

				meter := otel.Meter("cnproxy")
				requestCount, err := meter.Int64Counter("request_count")
				if err != nil {
					return fmt.Errorf("failed to create counter: %v", err)
				}
				requestCount.Add(context.Background(), 1)
			}

			cnproxyServer := &server.CNProxyServer{
				Port:    port,
				Address: address,
			}
			return cnproxyServer.Serve()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
