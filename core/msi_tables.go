// core/msi_tables.go
package core

import (
	"fmt"
	"sort"
	"strings"
)

// Remove unused import "encoding/json" if not required.

// TableRow represents a single row from an MSI table.
type TableRow struct {
	Columns []string
}

// discoveredTable holds a table name along with the method/source
// it was discovered from.
type discoveredTable struct {
	Name   string
	Source string
}

// ListTables discovers and prints table names from an MSI file.
func ListTables(msiPath string) error {
	return SafeExecute("ListTables", func() error {
		session, err := OpenMsiSession(msiPath, 0)
		if err != nil {
			return fmt.Errorf("failed to open MSI session: %v", err)
		}
		defer session.Close()

		results, err := discoverTables(session)
		fmt.Println("📦 Tables in", msiPath)

		if err != nil || len(results) == 0 {
			fmt.Println("   ⚠ No tables found — MSI may be empty, encrypted, or restricted.")
			if DebugMode && err != nil {
				logWarn(fmt.Sprintf("discoverTables error: %v", err))
			}
			return nil
		}

		// Build a map for unique table names and count how many came from each method.
		summary := map[string]int{}
		tableMap := map[string]string{}
		for _, t := range results {
			tableMap[t.Name] = t.Source
			summary[t.Source]++
		}

		var deduped []string
		for table := range tableMap {
			deduped = append(deduped, table)
		}
		sort.Strings(deduped)

		for _, table := range deduped {
			fmt.Printf("   └─ %-30s [via %s]\n", table, tableMap[table])
		}

		if DebugMode {
			fmt.Println("\n🔍 Discovery Summary:")
			for source, count := range summary {
				fmt.Printf("   %-20s → %d tables\n", source, count)
			}
		}
		return nil
	})
}

// ReadTableRows reads all rows from a specified MSI table.
// This function is used by other components (e.g. record editing).
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

// FormatRows neatly formats table rows into a readable string.
func FormatRows(rows []TableRow) string {
	var sb strings.Builder
	for idx, row := range rows {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", idx+1, strings.Join(row.Columns, " | ")))
	}
	return sb.String()
}

// discoverTables tries multiple methods to locate table names.
func discoverTables(session *MsiSession) ([]discoveredTable, error) {
	methods := []struct {
		Name string
		Exec func(*MsiSession) ([]string, error)
	}{
		{"_Tables", tryListSystemTables},
		{"_Columns", tryListColumnsDistinct},
		{"BruteForce", tryListBruteForce},
	}

	var results []discoveredTable
	var errors []string
	for _, method := range methods {
		var names []string
		err := SafeExecute("DiscoverVia::"+method.Name, func() error {
			var e error
			names, e = method.Exec(session)
			return e
		})

		if err == nil && len(names) > 0 {
			for _, name := range names {
				results = append(results, discoveredTable{Name: name, Source: method.Name})
			}
			if DebugMode {
				logInfo(fmt.Sprintf("✅ Tables found using '%s': %d", method.Name, len(names)))
			}
		} else {
			msg := fmt.Sprintf("[%s] %v", method.Name, err)
			if DebugMode {
				logWarn(msg)
			}
			errors = append(errors, msg)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("table discovery failed:\n%s", strings.Join(errors, "\n"))
	}
	return results, nil
}

// tryListSystemTables queries the _Tables table.
func tryListSystemTables(session *MsiSession) ([]string, error) {
	rows, err := session.ExecuteQuery("SELECT * FROM `_Tables`")
	if err != nil {
		return nil, fmt.Errorf("failed to query _Tables: %v", err)
	}
	return extractFirstColumn(rows, "_Tables")
}

// tryListColumnsDistinct queries distinct table names from _Columns.
func tryListColumnsDistinct(session *MsiSession) ([]string, error) {
	rows, err := session.ExecuteQuery("SELECT DISTINCT `Table` FROM `_Columns`")
	if err != nil {
		return nil, fmt.Errorf("failed to query _Columns: %v", err)
	}
	return extractFirstColumn(rows, "_Columns")
}

// tryListBruteForce checks common tables directly.
func tryListBruteForce(session *MsiSession) ([]string, error) {
	common := []string{
		"Property", "Directory", "Feature", "Component",
		"File", "Binary", "Media", "Registry",
		"Shortcut", "CustomAction", "InstallExecuteSequence",
	}
	var found []string
	for _, t := range common {
		rows, err := session.ExecuteQuery(fmt.Sprintf("SELECT * FROM `%s`", t))
		if err == nil && len(rows) > 0 {
			found = append(found, t)
			if DebugMode {
				logInfo(fmt.Sprintf("BruteForce → found '%s'", t))
			}
		} else if DebugMode {
			logWarn(fmt.Sprintf("BruteForce → skipped '%s': %v", t, err))
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("no common tables found")
	}
	return found, nil
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

// extractFirstColumn returns the first column values from rows,
// optionally filtering out entries that start with '_' or match known dummy tables.
func extractFirstColumn(rows []TableRow, source string) ([]string, error) {
	var out []string
	for _, r := range rows {
		if len(r.Columns) > 0 {
			name := strings.TrimSpace(r.Columns[0])
			if name != "" && !strings.HasPrefix(name, "_") && name != "MsiDigitalCertificate" {
				out = append(out, name)
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no tables found in %s", source)
	}
	return out, nil
}
