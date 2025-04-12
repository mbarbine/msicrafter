// core/msi_transform.go
package core

import (
	"fmt"
	"os"
	"strings"
)

// GenerateTransform analyzes differences between the original and modified MSI
// on a per-table basis, then produces a naive .mst transform file capturing changes.
func GenerateTransform(originalMSI, modifiedMSI, outputTransform string) error {
	// Step 0: Validate existence of input files.
	if _, err := os.Stat(originalMSI); os.IsNotExist(err) {
		return fmt.Errorf("original MSI not found: %s", originalMSI)
	}
	if _, err := os.Stat(modifiedMSI); os.IsNotExist(err) {
		return fmt.Errorf("modified MSI not found: %s", modifiedMSI)
	}

	// Step 1: Gather table names from both MSIs.
	origTables, err := getTables(originalMSI)
	if err != nil {
		return fmt.Errorf("failed to list tables in original: %v", err)
	}
	modTables, err := getTables(modifiedMSI)
	if err != nil {
		return fmt.Errorf("failed to list tables in modified: %v", err)
	}

	// Union of both sets for comparison.
	allTablesMap := map[string]bool{}
	for _, t := range origTables {
		allTablesMap[t] = true
	}
	for _, t := range modTables {
		allTablesMap[t] = true
	}
	var allTables []string
	for t := range allTablesMap {
		allTables = append(allTables, t)
	}

	// Step 2: Compare rows per table.
	var differences []string
	for _, table := range allTables {
		origRows, err1 := ReadTableRows(originalMSI, table)
		modRows, err2 := ReadTableRows(modifiedMSI, table)

		if err1 != nil && err2 != nil {
			// Skip table if unreadable in both
			continue
		}

		rowDiff := compareTableRows(table, origRows, modRows)
		if rowDiff != "" {
			differences = append(differences, rowDiff)
		}
	}

	// Step 3: Save transform (mocked as diff file).
	if err := writeMSTStub(differences, outputTransform); err != nil {
		return fmt.Errorf("failed to write transform file: %v", err)
	}

	return nil
}

// compareTableRows returns a textual diff for the given table, or "" if no differences.
func compareTableRows(table string, orig, mod []TableRow) string {
	var sb strings.Builder

	// Naively join column values to compare rows.
	origMap := make(map[string]bool)
	for _, row := range orig {
		key := strings.Join(row.Columns, "|")
		origMap[key] = true
	}

	modMap := make(map[string]bool)
	for _, row := range mod {
		key := strings.Join(row.Columns, "|")
		modMap[key] = true
	}

	// Find additions.
	for key := range modMap {
		if !origMap[key] {
			sb.WriteString(fmt.Sprintf("+ %s => %s\n", table, key))
		}
	}
	// Find deletions.
	for key := range origMap {
		if !modMap[key] {
			sb.WriteString(fmt.Sprintf("- %s => %s\n", table, key))
		}
	}
	return sb.String()
}

// writeMSTStub writes the textual diff to a .mst file as a demonstration.
// A real MST would use the Windows Installer APIs to generate binary output.
func writeMSTStub(differences []string, mstPath string) error {
	f, err := os.Create(mstPath)
	if err != nil {
		return fmt.Errorf("failed to create MST file: %v", err)
	}
	defer f.Close()

	for _, diffLine := range differences {
		if _, err := f.WriteString(diffLine); err != nil {
			return err
		}
	}
	return nil
}

// getTables retrieves a list of table names from an MSI.
func getTables(msiPath string) ([]string, error) {
	tables, err := ListAllTables(msiPath)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

// ListAllTables reads the internal _Tables table and returns table names.
func ListAllTables(msiPath string) ([]string, error) {
	var tableNames []string

	if DebugMode {
		fmt.Printf("[DEBUG] Attempting to read '_Tables' from %s...\n", msiPath)
	}

	rows, err := ReadTableRows(msiPath, "_Tables")
	if err != nil {
		return nil, fmt.Errorf("failed to read _Tables: %v", err)
	}

	if DebugMode {
		fmt.Printf("[DEBUG] _Tables returned %d rows\n", len(rows))
	}

	for _, row := range rows {
		if len(row.Columns) > 0 && row.Columns[0] != "" {
			tableNames = append(tableNames, row.Columns[0])
			if DebugMode {
				fmt.Printf("[DEBUG] Found table: %s\n", row.Columns[0])
			}
		}
	}
	return tableNames, nil
}
