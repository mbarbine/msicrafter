// main.go
package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"msicrafter/cli"
	"msicrafter/retro"
)

func main() {
	// Display a retro splash screen
	retro.ShowSplash()

	// Setup CLI app with commands.
	app := &cli.App{
		Name:  "msicrafter",
		Usage: "Retro-powered MSI table editor & transform tool",
		Commands: []*cli.Command{
			cli.ListTablesCommand,
			cli.QueryCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
