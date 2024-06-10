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
		},
		Action: func(*cli.Context) error {
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
