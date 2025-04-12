// cli/commands.go
package cli

import (
	"fmt"
	"msicrafter/core"

	"github.com/urfave/cli/v2"
)

// ListTablesCommand shows all tables in a given MSI database.
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

// QueryCommand executes a SQL query against an MSI database.
var QueryCommand = &cli.Command{
	Name:  "query",
	Usage: "Execute a SQL query against an MSI database",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "q",
			Aliases:  []string{"query"},
			Usage:    "SQL query to execute",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		if c.Args().Len() == 0 {
			return fmt.Errorf("provide path to MSI file")
		}
		msiPath := c.Args().Get(0)
		sqlQuery := c.String("q")
		return core.QueryMSI(msiPath, sqlQuery)
	},
}
