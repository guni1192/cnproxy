package main

import (
	"log"
	"os"

	"github.com/guni1192/cnproxy/pkg/middleware/server"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "cnproxy",
		Usage: "cloud native proxy",
		Action: func(*cli.Context) error {
			return server.Serve()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
