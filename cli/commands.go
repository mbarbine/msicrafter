package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"os"
	"github.com/urfave/cli/v2"
	"msicrafter/core"
)

// Commands is the consolidated slice of all CLI commands.
var Commands = []*cli.Command{
	listTablesCommand(),
	queryCommand(),
	editCommand(),
	transformCommand(),
	diffCommand(),
	exportCommand(),
	backupCommand(),
	applyTransformCommand(),
	listRecordsCommand(),
	editRecordCommand(),
	editTableCommand(),
}


func editTableCommand() *cli.Command {
    return &cli.Command{
        Name:      "edit",
        Aliases:   []string{"update"},
        Usage:     "Edit a table in an MSI database",
        ArgsUsage: "<msi_file>",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:     "table",
                Aliases:  []string{"t"},
                Usage:    "Table name to edit",
                Required: true,
            },
            &cli.StringFlag{
                Name:     "set",
                Aliases:  []string{"s"},
                Usage:    "Set clause (e.g., Property='NewValue',Value='Test')",
                Required: true,
            },
            &cli.StringFlag{
                Name:    "where",
                Aliases: []string{"w"},
                Usage:   "Where clause (e.g., Property='Key')",
            },
            &cli.BoolFlag{
                Name:    "dry-run",
                Aliases: []string{"n"},
                Usage:   "Simulate edit without committing",
            },
            &cli.BoolFlag{
                Name:    "interactive",
                Aliases: []string{"i"},
                Usage:   "Prompt for confirmation before editing",
            },
        },
        Action: func(c *cli.Context) error {
            return core.SafeExecute("EditTable", func() error {
                if c.Args().Len() < 1 {
                    return fmt.Errorf("MSI file path is required")
                }
                msiPath := c.Args().Get(0)
                if err := validateFileExists(msiPath, "MSI"); err != nil {
                    return err
                }
                tableName := c.String("table")
                setClause := c.String("set")
                whereClause := c.String("where")
                dryRun := c.Bool("dry-run")
                interactive := c.Bool("interactive")

                session, err := core.OpenMsiSession(msiPath, 1) // Read-write
                if err != nil {
                    return fmt.Errorf("failed to open MSI session: %v", err)
                }
                defer session.Close()

                err = session.EditTable(tableName, setClause, whereClause, dryRun, interactive)
                if err == nil && !dryRun {
                    fmt.Printf("Table '%s' updated in: %s\n", tableName, msiPath)
                }
                return err
            })
        },
    }
}
// listTablesCommand shows all tables in a given MSI database.
func listTablesCommand() *cli.Command {
	return &cli.Command{
		Name:      "tables",
		Aliases:   []string{"ls"},
		Usage:     "List all tables in an MSI database",
		ArgsUsage: "<msi_file>",
		Action: func(c *cli.Context) error {
			return core.SafeExecute("ListTables", func() error {
				msiPath, err := validateMSIPath(c)
				if err != nil {
					return err
				}
				return core.ListTables(msiPath)
			})
		},
	}
}

// queryCommand executes an arbitrary SQL query against an MSI database.
func queryCommand() *cli.Command {
	return &cli.Command{
		Name:      "query",
		Aliases:   []string{"sql"},
		Usage:     "Execute a SQL query against an MSI database",
		ArgsUsage: "<msi_file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "query",
				Aliases:  []string{"q"},
				Usage:    "SQL query to execute (e.g., 'SELECT * FROM Property')",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			return core.SafeExecute("Query", func() error {
				msiPath, err := validateMSIPath(c)
				if err != nil {
					return err
				}
				sqlQuery := c.String("query")
				if strings.TrimSpace(sqlQuery) == "" {
					return fmt.Errorf("query cannot be empty")
				}
				return core.QueryMSI(msiPath, sqlQuery)
			})
		},
	}
}

// editCommand updates a table in an MSI database using a set clause.
func editCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Aliases:   []string{"update"},
		Usage:     "Edit a table in an MSI database",
		ArgsUsage: "<msi_file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "table",
				Aliases:  []string{"t"},
				Usage:    "Table name to edit",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "set",
				Aliases:  []string{"s"},
				Usage:    "Set clause (e.g., 'field=value,field2=value2')",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "where",
				Usage: "Optional WHERE clause to filter rows",
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "Simulate the edit without committing changes",
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Prompt for confirmation before applying changes",
			},
		},
		Action: func(c *cli.Context) error {
			return core.SafeExecute("EditTable", func() error {
				msiPath, err := validateMSIPath(c)
				if err != nil {
					return err
				}
				tableName := c.String("table")
				setClause := c.String("set")
				whereClause := c.String("where")
				dryRun := c.Bool("dry-run")
				interactive := c.Bool("interactive")
				return core.EditTable(msiPath, tableName, setClause, whereClause, dryRun, interactive)
			})
		},
	}
}

// transformCommand generates a transform file (MST) from original and modified MSI files.
func transformCommand() *cli.Command {
	return &cli.Command{
		Name:    "transform",
		Aliases: []string{"mst"},
		Usage:   "Generate a transform file from original and modified MSI files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "original",
				Aliases:  []string{"o"},
				Usage:    "Path to the original MSI file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "modified",
				Aliases:  []string{"m"},
				Usage:    "Path to the modified MSI file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"out"},
				Usage:    "Path for output transform (.mst) file",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			return core.SafeExecute("GenerateTransform", func() error {
				orig := c.String("original")
				mod := c.String("modified")
				output := c.String("output")
				if err := validateFileExists(orig, "original MSI"); err != nil {
					return err
				}
				if err := validateFileExists(mod, "modified MSI"); err != nil {
					return err
				}
				if err := validateOutputPath(output, ".mst"); err != nil {
					return err
				}
				err := core.GenerateTransform(orig, mod, output)
				if err == nil {
					fmt.Printf("Transform created: %s\n", output)
				}
				return err
			})
		},
	}
}

// diffCommand compares two MSI files and prints a diff summary.
func diffCommand() *cli.Command {
	return &cli.Command{
		Name:      "diff",
		Aliases:   []string{"compare"},
		Usage:     "Compare two MSI files for differences",
		ArgsUsage: "<msi_file1> <msi_file2>",
		Action: func(c *cli.Context) error {
			return core.SafeExecute("CompareMSI", func() error {
				if c.Args().Len() < 2 {
					return fmt.Errorf("two MSI file paths are required")
				}
				msi1 := c.Args().Get(0)
				msi2 := c.Args().Get(1)
				if err := validateFileExists(msi1, "first MSI"); err != nil {
					return err
				}
				if err := validateFileExists(msi2, "second MSI"); err != nil {
					return err
				}
				return core.CompareMSI(msi1, msi2)
			})
		},
	}
}

// exportCommand exports MSI tables to CSV or JSON and compresses them into a zip file.
func exportCommand() *cli.Command {
	return &cli.Command{
		Name:      "export",
		Aliases:   []string{"dump"},
		Usage:     "Export MSI tables to CSV or JSON and compress into a zip file",
		ArgsUsage: "<msi_file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "format",
				Aliases:  []string{"f"},
				Usage:    "Export format: 'csv' or 'json'",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "Output zip file path",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			return core.SafeExecute("ExportMSI", func() error {
				msiPath, err := validateMSIPath(c)
				if err != nil {
					return err
				}
				format := strings.ToLower(c.String("format"))
				output := c.String("output")
				if format != "csv" && format != "json" {
					return fmt.Errorf("format must be 'csv' or 'json', got '%s'", format)
				}
				if err := validateOutputPath(output, ".zip"); err != nil {
					return err
				}
				err = core.ExportMSI(msiPath, format, output)
				if err == nil {
					fmt.Printf("Exported tables to: %s\n", output)
				}
				return err
			})
		},
	}
}

// backupCommand creates a backup copy of an MSI file.
func backupCommand() *cli.Command {
	return &cli.Command{
		Name:      "backup",
		Aliases:   []string{"bak"},
		Usage:     "Create a backup of an MSI file",
		ArgsUsage: "<msi_file>",
		Action: func(c *cli.Context) error {
			return core.SafeExecute("BackupMSI", func() error {
				msiPath, err := validateMSIPath(c)
				if err != nil {
					return err
				}
				backupPath, err := core.BackupMSI(msiPath)
				if err != nil {
					return err
				}
				fmt.Printf("Backup created: %s\n", backupPath)
				return nil
			})
		},
	}
}

// applyTransformCommand applies a transform file to an MSI database.
func applyTransformCommand() *cli.Command {
	return &cli.Command{
		Name:      "apply",
		Aliases:   []string{"patch"},
		Usage:     "Apply an MST transform file to an MSI database",
		ArgsUsage: "<mst_file> <msi_file>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "Simulate applying the transform without committing changes",
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Prompt for confirmation before applying changes",
			},
		},
		Action: func(c *cli.Context) error {
			return core.SafeExecute("ApplyTransform", func() error {
				if c.Args().Len() < 2 {
					return fmt.Errorf("MST and MSI file paths are required")
				}
				mstPath := c.Args().Get(0)
				msiPath := c.Args().Get(1)
				if err := validateFileExists(mstPath, "MST"); err != nil {
					return err
				}
				if err := validateFileExists(msiPath, "MSI"); err != nil {
					return err
				}
				dryRun := c.Bool("dry-run")
				interactive := c.Bool("interactive")
				err := core.ApplyTransform(msiPath, mstPath, dryRun, interactive)
				if err == nil && !dryRun {
					fmt.Printf("Transform applied to: %s\n", msiPath)
				}
				return err
			})
		},
	}
}

// listRecordsCommand lists the records of a specified table in an MSI database.
func listRecordsCommand() *cli.Command {
	return &cli.Command{
		Name:      "records",
		Aliases:   []string{"list-records", "rows"},
		Usage:     "List all records of a table in an MSI database",
		ArgsUsage: "<msi_file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "table",
				Aliases:  []string{"t"},
				Usage:    "Table name to list records from",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Include column names in output",
			},
		},
		Action: func(c *cli.Context) error {
			return core.SafeExecute("ListRecords", func() error {
				msiPath, err := validateMSIPath(c)
				if err != nil {
					return err
				}
				tableName := c.String("table")
				verbose := c.Bool("verbose")
				rows, err := core.ReadTableRows(msiPath, tableName)
				if err != nil {
					return err
				}
				if len(rows) == 0 {
					fmt.Printf("No records found in table '%s'\n", tableName)
					return nil
				}
				if verbose {
					cols, err := core.GetColumnNames(msiPath, tableName)
					if err == nil {
						fmt.Printf("Table '%s' columns: %s\n", tableName, strings.Join(cols, ", "))
					}
				}
				fmt.Printf("Records in table '%s' (%d rows):\n", tableName, len(rows))
				fmt.Println(core.FormatRows(rows))
				return nil
			})
		},
	}
}

func editRecordCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit-record",
		Aliases:   []string{"update-record"},
		Usage:     "Edit a specific record in a table by row number",
		ArgsUsage: "<msi_file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "table",
				Aliases:  []string{"t"},
				Usage:    "Table name to edit",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "row",
				Aliases:  []string{"r"},
				Usage:    "Row number to edit (starting at 1)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "set",
				Aliases:  []string{"s"},
				Usage:    "Set clause (e.g., 'field=value,field2=value2')",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "Simulate the edit without committing changes",
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Prompt for confirmation before applying changes",
			},
		},
		Action: func(c *cli.Context) error {
			return core.SafeExecute("EditRecord", func() error {
				msiPath, err := validateMSIPath(c)
				if err != nil {
					return err
				}
				tableName := c.String("table")
				rowNum := c.Int("row")
				setClause := c.String("set")
				dryRun := c.Bool("dry-run")
				interactive := c.Bool("interactive")
				if rowNum < 1 {
					return fmt.Errorf("row number must be positive, got %d", rowNum)
				}
				return core.EditRecord(msiPath, tableName, rowNum, setClause, dryRun, interactive)
			})
		},
	}
}

// validateMSIPath ensures a single MSI file path is provided and exists.
func validateMSIPath(c *cli.Context) (string, error) {
	if c.Args().Len() == 0 {
		return "", fmt.Errorf("MSI file path is required")
	}
	if c.Args().Len() > 1 {
		return "", fmt.Errorf("only one MSI file path is allowed, got %d", c.Args().Len())
	}
	msiPath := c.Args().Get(0)
	return msiPath, validateFileExists(msiPath, "MSI")
}

// validateFileExists checks if a file exists and has the expected extension.
func validateFileExists(path, fileType string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("%s path cannot be empty", fileType)
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s file does not exist: %s", fileType, path)
	}
	if err != nil {
		return fmt.Errorf("failed to access %s file '%s': %v", fileType, path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s path is a directory, not a file: %s", fileType, path)
	}
	return nil
}

// validateOutputPath ensures the output path is valid and has the expected extension.
func validateOutputPath(path, expectedExt string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	if !strings.HasSuffix(strings.ToLower(path), expectedExt) {
		return fmt.Errorf("output file must have %s extension, got '%s'", expectedExt, path)
	}
	dir := filepath.Dir(path)
	if dir != "." {
		if err := validateDirExists(dir); err != nil {
			return fmt.Errorf("output directory invalid: %v", err)
		}
	}
	return nil
}

// validateDirExists checks if the parent directory for an output file exists.
func validateDirExists(dir string) error {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	if err != nil {
		return fmt.Errorf("failed to access directory '%s': %v", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dir)
	}
	return nil
}