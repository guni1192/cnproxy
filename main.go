package main

import (
	"context"
	"log"
	"os"

	"github.com/guni1192/cnproxy/pkg/middleware/logger"
	"github.com/guni1192/cnproxy/pkg/middleware/server"
	"github.com/urfave/cli/v2"
)

func main() {
	l := logger.New()

	app := &cli.App{
		Name:  "cnproxy",
		Usage: "cloud native proxy",
		Action: func(*cli.Context) error {
			cnproxyServer := &server.CNProxyServer{
				Context: context.Background(),
				Logger:  l,
			}
			return cnproxyServer.Serve()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
