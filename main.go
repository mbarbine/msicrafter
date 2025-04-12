// main.go
package main

import (
	"log"
	"os"

	urfavecli "github.com/urfave/cli/v2" // alias for external CLI package
	mcli "msicrafter/cli"              // alias for your local cli package
	"msicrafter/core"
	"msicrafter/retro"
)

var (
	version   = "dev"    // default version; override with ldflags during build if needed
	buildDate = "4112025"
)

func main() {
	// Display the splash screen. (Note: retro.ShowSplash now takes no arguments)
	retro.ShowSplash()
	log.Printf("msicrafter version: %s", version)

	app := &urfavecli.App{
		Name:    "msicrafter",
		Version: version,
		Usage:   "Retro-powered MSI table editor & transform tool",
		Flags: []urfavecli.Flag{
			&urfavecli.BoolFlag{
				Name:  "debug",
				Usage: "Enable verbose debug logging",
			},
		},
		Before: func(c *urfavecli.Context) error {
			core.DebugMode = c.Bool("debug")
			if core.DebugMode {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Println("[DEBUG] Debug mode enabled.")
			} else {
				log.SetFlags(log.LstdFlags)
			}
			return nil
		},
		Commands: mcli.Commands, // use commands from your local cli package
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
}
