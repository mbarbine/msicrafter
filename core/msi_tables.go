// core/msi_tables.go
package core

import (
	"fmt"
	"strings"
)

// TableRow represents a single row from an MSI table.
type TableRow struct {
	Columns []string
}

// ListTables discovers and prints table names from an MSI file.
func ListTables(msiPath string) error {
	return SafeExecute("ListTables", func() error {
		session, err := OpenMsiSession(msiPath, 0)
		if err != nil {
			return fmt.Errorf("failed to open MSI session: %v", err)
		}
		defer session.Close()

		tables, err := discoverTables(session)
		fmt.Println("📦 Tables in", msiPath)
		if err != nil || len(tables) == 0 {
			fmt.Println("   ⚠ No tables found — MSI may be empty, encrypted, or restricted.")
			if DebugMode && err != nil {
				logWarn(fmt.Sprintf("discoverTables error: %v", err))
			}
			return nil
		}

		for _, table := range tables {
			fmt.Println("   └─", table)
		}
		return nil
	})
}

// ReadTableRows reads all rows from a specified MSI table.
func ReadTableRows(msiPath, tableName string) ([]TableRow, error) {
	var rows []TableRow
	err := SafeExecuteWithRetry("ReadTableRows", 3, func() error {
		session, err := OpenMsiSession(msiPath, 0)
		if err != nil {
			return fmt.Errorf("failed to open MSI session: %v", err)
		}
		defer session.Close()

		sql := fmt.Sprintf("SELECT * FROM `%s`", tableName)
		rows, err = session.ExecuteQuery(sql)
		if err != nil {
			return fmt.Errorf("failed to read table '%s': %v", tableName, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GetColumnNames retrieves column names for a table.
func GetColumnNames(msiPath, tableName string) ([]string, error) {
	session, err := OpenMsiSession(msiPath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open MSI session: %v", err)
	}
	defer session.Close()
	return session.GetColumnNames(tableName)
}

// FormatRows neatly formats table rows into a readable string.
func FormatRows(rows []TableRow) string {
	var sb strings.Builder
	for idx, row := range rows {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", idx+1, strings.Join(row.Columns, " | ")))
	}
	return sb.String()
}

// discoverTables tries multiple methods to locate table names.
func discoverTables(session *MsiSession) ([]string, error) {
	methods := []struct {
		name string
		fn   func(*MsiSession) ([]string, error)
	}{
		{"_Tables System Table", tryListSystemTables},
		{"_Columns Distinct", tryListColumnsDistinct},
		{"Brute Force", tryListBruteForce},
	}

	var combinedErrors []string
	for _, method := range methods {
		var tables []string
		err := SafeExecute(fmt.Sprintf("DiscoverTables(%s)", method.name), func() error {
			var err error
			tables, err = method.fn(session)
			if err != nil {
				return err
			}
			return nil
		})
		if err == nil && len(tables) > 0 {
			if DebugMode {
				logInfo(fmt.Sprintf("Discovered tables via '%s'", method.name))
			}
			return tables, nil
		}
		if err != nil && DebugMode {
			logWarn(fmt.Sprintf("Discovery method '%s' failed: %v", method.name, err))
		}
		combinedErrors = append(combinedErrors, fmt.Sprintf("%s: %v", method.name, err))
	}

	return nil, fmt.Errorf("table discovery failed:\n%s", strings.Join(combinedErrors, "\n"))
}

// tryListSystemTables queries the _Tables table.
func tryListSystemTables(session *MsiSession) ([]string, error) {
	rows, err := session.ExecuteQuery("SELECT * FROM `_Tables`")
	if err != nil {
		return nil, fmt.Errorf("failed to query _Tables: %v", err)
	}
	var tables []string
	for _, row := range rows {
		if len(row.Columns) > 0 && row.Columns[0] != "" {
			tables = append(tables, row.Columns[0])
		}
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("_Tables is empty")
	}
	return tables, nil
}

// tryListColumnsDistinct queries distinct table names from _Columns.
func tryListColumnsDistinct(session *MsiSession) ([]string, error) {
	rows, err := session.ExecuteQuery("SELECT DISTINCT `Table` FROM `_Columns`")
	if err != nil {
		return nil, fmt.Errorf("failed to query _Columns: %v", err)
	}
	var tables []string
	for _, row := range rows {
		if len(row.Columns) > 0 && row.Columns[0] != "" {
			tables = append(tables, row.Columns[0])
		}
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("_Columns has no valid tables")
	}
	return tables, nil
}

// tryListBruteForce checks common tables directly.
func tryListBruteForce(session *MsiSession) ([]string, error) {
	commonTables := []string{
		"Property", "Directory", "Feature", "Component",
		"File", "Binary", "Media", "Registry",
	}
	var found []string
	for _, table := range commonTables {
		rows, err := session.ExecuteQuery(fmt.Sprintf("SELECT * FROM `%s`", table))
		if err == nil && len(rows) > 0 {
			found = append(found, table)
			if DebugMode {
				logInfo(fmt.Sprintf("BruteForce: Found table '%s'", table))
			}
		} else if DebugMode {
			logWarn(fmt.Sprintf("BruteForce: Skipped table '%s': %v", table, err))
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("no common tables found")
	}
	return found, nil
}