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
			err := core.GenerateTransform(orig, mod, output)
			if err == nil {
				fmt.Printf("Transform (MST) created successfully at: %s\n", output)
			}
			return err
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

// ApplyTransformCommand applies a transform file to an MSI database.
var ApplyTransformCommand = &cli.Command{
	Name:  "apply",
	Usage: "Apply an MST transform file to an MSI database",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Simulate applying the transform without committing changes",
		},
		&cli.BoolFlag{
			Name:  "interactive",
			Usage: "Ask for confirmation before each operation",
		},
	},
	Action: func(c *cli.Context) error {
		return core.SafeExecute("ApplyTransform", func() error {
			if c.Args().Len() < 2 {
				return fmt.Errorf("provide path to MST file and target MSI file")
			}
			mstPath := c.Args().Get(0)
			msiPath := c.Args().Get(1)
			dryRun := c.Bool("dry-run")
			interactive := c.Bool("interactive")
			return core.ApplyTransform(msiPath, mstPath, dryRun, interactive)
		})
	},
}
// ListRecordsCommand lists the records of a specified table in an MSI database.
var ListRecordsCommand = &cli.Command{
	Name:  "list-records",
	Usage: "List all records of a given table in an MSI database",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "table",
			Usage:    "Table name to list records from",
			Required: true,
		},
	},

	Action: func(c *cli.Context) error {
		return SafeExecute("ListRecords", func() error {
			if c.Args().Len() == 0 {
				return fmt.Errorf("provide path to MSI file")
			}
			msiPath := c.Args().Get(0)
			tableName := c.String("table")
			rows, err := ReadTableRows(msiPath, tableName)
			if err != nil {
				return err
			}
			fmt.Println("Records in table", tableName)
			fmt.Println(FormatRows(rows))
			return nil
		})
	},
}

// EditRecordCommand edits a specific record in a given table by its row number.
var EditRecordCommand = &cli.Command{
	Name:  "edit-record",
	Usage: "Edit a specific record from a table in an MSI database by record number",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "table",
			Usage:    "Table name in which to edit a record",
			Required: true,
		},
		&cli.IntFlag{
			Name:     "record",
			Usage:    "Record number (row number, starting at 1) to edit",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "set",
			Usage:    "Set clause in the format field=value[,field2=value2...]",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Simulate the edit without committing changes",
		},
		&cli.BoolFlag{
			Name:  "interactive",
			Usage: "Ask for confirmation before applying the changes",
		},
	},
	Action: func(c *cli.Context) error {
		return core.SafeExecute("EditRecord", func() error {
			if c.Args().Len() == 0 {
				return fmt.Errorf("provide path to MSI file")
			}
			msiPath := c.Args().Get(0)
			tableName := c.String("table")
			recordNum := c.Int("record")
			setClause := c.String("set")
			dryRun := c.Bool("dry-run")
			interactive := c.Bool("interactive")
			return core.EditRecord(msiPath, tableName, recordNum, setClause, dryRun, interactive)
		})
	},
}

// Append the new command to the commands list.
var Commands = []*cli.Command{
	ListTablesCommand,
	QueryCommand,
	EditCommand,
	TransformCommand,
	DiffCommand,
	ExportCommand,
	BackupCommand,
	ApplyTransformCommand,
	ListRecordsCommand,
	EditRecordCommand, // <-- New record-level editing command.
}