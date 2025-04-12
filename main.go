// main.go
package main

import (
    "log"
    "os"

    urfavecli "github.com/urfave/cli/v2"
    mcli "msicrafter/cli"
    "msicrafter/core"
    "msicrafter/retro"
)

var (
    version   = "dev"
    buildDate = "4112025"
)

func main() {
    retro.ShowSplash()
    log.Printf("msicrafter version: %s", version)

    if err := core.InitCOM(); err != nil {
        log.Fatalf("[FATAL] COM initialization failed: %v", err)
    }
    defer core.CleanupCOM()

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
        Commands: mcli.Commands,
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatalf("[FATAL] %v", err)
    }
}