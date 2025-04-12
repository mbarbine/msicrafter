// core/msi_transform.go
package core

import (
	"fmt"
	"os"
	"strings"
)

// GenerateTransform analyzes differences between the original and modified MSI
// on a per-table basis, then produces a naive .mst transform file capturing changes.
// This is still a simplified approach to demonstrate real diffing.
func GenerateTransform(originalMSI, modifiedMSI, outputTransform string) error {
	// For demonstration, we'll:
	// 1. Enumerate the tables in both MSIs.
	// 2. For each table, read row data from both MSIs.
	// 3. Compare row sets to find added/removed/changed rows.
	// 4. Write a simple MST that attempts to reflect these differences.

	// Step 0: Validate existence of input files.
	if _, err := os.Stat(originalMSI); os.IsNotExist(err) {
		return fmt.Errorf("original MSI not found: %s", originalMSI)
	}
	if _, err := os.Stat(modifiedMSI); os.IsNotExist(err) {
		return fmt.Errorf("modified MSI not found: %s", modifiedMSI)
	}

	// Step 1: Gather table names in each MSI.
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

	// Step 2: For each table, read row data from both MSIs and detect diffs.
	var differences []string
	for _, table := range allTables {
		oRows, _ := ReadTable(originalMSI, table)
		mRows, _ := ReadTable(modifiedMSI, table)

		// We do a naive row-by-row string comparison.
		rowDiff := compareTableRows(table, oRows, mRows)
		if rowDiff != "" {
			differences = append(differences, rowDiff)
		}
	}

	// Step 3: Write out an MST file with these differences.
	// For demonstration, we store the differences in plain text.
	if err := writeMSTStub(differences, outputTransform); err != nil {
		return err
	}

	return nil
}

// compareTableRows returns a textual diff for the given table, or "" if no differences.
func compareTableRows(table string, orig, mod []TableRow) string {
	var sb strings.Builder

	// Convert slices to maps keyed by a joined string of all columns (very naive).
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
	if sb.Len() == 0 {
		return ""
	}
	return sb.String()
}

// writeMSTStub just writes the diff lines to the .mst file for demonstration.
// A real MST has a specific binary structure, typically generated via Windows Installer APIs.
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

// getTables is a helper to enumerate table names in a given MSI.
func getTables(msiPath string) ([]string, error) {
	tables, err := ListAllTables(msiPath)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

// ListAllTables is a variation of ListTables that returns a slice instead of printing to stdout.
func ListAllTables(msiPath string) ([]string, error) {
	tableNames := []string{}

	mTables, err := ReadTable(msiPath, "_Tables")
	if err != nil {
		// If there's an error reading _Tables, we have no fallback
		return tableNames, err
	}
	for _, row := range mTables {
		if len(row.Columns) > 0 && row.Columns[0] != "" {
			tableNames = append(tableNames, row.Columns[0])
		}
	}
	return tableNames, nil
}
