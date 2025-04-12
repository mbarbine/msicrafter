// main.go
package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"msicrafter/cli"
	"msicrafter/retro"
)

var debugEnabled bool

func main() {
	retro.ShowSplash()

	app := &cli.App{
		Name:  "msicrafter",
		Usage: "Retro-powered MSI table editor & transform tool",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable verbose debug logging",
			},
		},
		Before: func(c *cli.Context) error {
			debugEnabled = c.Bool("debug")
			if debugEnabled {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Println("[DEBUG] Debug mode enabled.")
			} else {
				log.SetFlags(0)
			}
			return nil
		},
		Commands: cli.Commands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
