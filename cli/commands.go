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
		return core.SafeExecute("ListTables", func() error {
			if c.Args().Len() == 0 {
				return fmt.Errorf("provide path to MSI file")
			}
			return core.ListTables(c.Args().Get(0))
		})
	},
}

// QueryCommand executes an arbitrary SQL query against an MSI database.
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
		return core.SafeExecute("Query", func() error {
			if c.Args().Len() == 0 {
				return fmt.Errorf("provide path to MSI file")
			}
			msiPath := c.Args().Get(0)
			sqlQuery := c.String("q")
			return core.QueryMSI(msiPath, sqlQuery)
		})
	},
}

// EditCommand updates a table in an MSI database using a set clause.
// Additional flags allow simulation (dry-run) and interactive confirmation.
var EditCommand = &cli.Command{
	Name:  "edit",
	Usage: "Edit a table in an MSI database",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "table",
			Usage:    "Table name to edit",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "set",
			Usage:    "Set clause in format field=value[,field2=value2...]",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Simulate the edit without committing the changes",
		},
		&cli.BoolFlag{
			Name:  "interactive",
			Usage: "Ask for confirmation before applying the changes",
		},
	},
	Action: func(c *cli.Context) error {
		return core.SafeExecute("EditTable", func() error {
			if c.Args().Len() == 0 {
				return fmt.Errorf("provide path to MSI file")
			}
			msiPath := c.Args().Get(0)
			tableName := c.String("table")
			setClause := c.String("set")
			dryRun := c.Bool("dry-run")
			interactive := c.Bool("interactive")
			return core.EditTable(msiPath, tableName, setClause, dryRun, interactive)
		})
	},
}

// TransformCommand generates a transform file (MST) from original and modified MSI files.
var TransformCommand = &cli.Command{
	Name:  "transform",
	Usage: "Generate a transform file from original and modified MSI files",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "original",
			Usage:    "Path to the original MSI file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "modified",
			Usage:    "Path to the modified MSI file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "output",
			Usage:    "Path for output transform (.mst) file",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		return core.SafeExecute("GenerateTransform", func() error {
			orig := c.String("original")
			mod := c.String("modified")
			output := c.String("output")
			return core.GenerateTransform(orig, mod, output)
		})
	},
}

// DiffCommand compares two MSI files and prints a simple diff summary.
var DiffCommand = &cli.Command{
	Name:  "diff",
	Usage: "Compare two MSI files for patch differences",
	Action: func(c *cli.Context) error {
		return core.SafeExecute("CompareMSI", func() error {
			if c.Args().Len() < 2 {
				return fmt.Errorf("provide paths to two MSI files")
			}
			msi1 := c.Args().Get(0)
			msi2 := c.Args().Get(1)
			return core.CompareMSI(msi1, msi2)
		})
	},
}

// ExportCommand exports MSI tables to CSV or JSON and compresses them into a zip file.
var ExportCommand = &cli.Command{
	Name:  "export",
	Usage: "Export MSI tables to CSV or JSON and compress them into a zip file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "format",
			Usage:    "Export format: 'csv' or 'json'",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "output",
			Usage:    "Output zip file path",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		return core.SafeExecute("ExportMSI", func() error {
			if c.Args().Len() == 0 {
				return fmt.Errorf("provide path to MSI file")
			}
			msiPath := c.Args().Get(0)
			format := c.String("format")
			output := c.String("output")
			return core.ExportMSI(msiPath, format, output)
		})
	},
}

// BackupCommand creates a backup copy of an MSI file.
var BackupCommand = &cli.Command{
	Name:  "backup",
	Usage: "Create a backup of the MSI file",
	Action: func(c *cli.Context) error {
		return core.SafeExecute("BackupMSI", func() error {
			if c.Args().Len() == 0 {
				return fmt.Errorf("provide path to MSI file")
			}
			msiPath := c.Args().Get(0)
			backupPath, err := core.BackupMSI(msiPath)
			if err != nil {
				return err
			}
			fmt.Printf("Backup created: %s\n", backupPath)
			return nil
		})
	},
}

// Commands is the consolidated slice of all CLI commands.
var Commands = []*cli.Command{
	ListTablesCommand,
	QueryCommand,
	EditCommand,
	TransformCommand,
	DiffCommand,
	ExportCommand,
	BackupCommand,
}
