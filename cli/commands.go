// cli/commands.go
package cli

import (
	"fmt"
	"msicrafter/core"
	"github.com/urfave/cli/v2"
)

var ListTablesCommand = &cli.Command{
	Name:  "tables",
	Usage: "List all tables in an MSI database",
	Action: func(c *cli.Context) error {
		if c.Args().Len() == 0 {
			return fmt.Errorf("provide path to MSI file")
		}
		return core.ListTables(c.Args().Get(0))
	},
}
