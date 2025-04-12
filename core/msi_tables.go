// core/msi_tables.go
package core

import (
	"fmt"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// TableRow represents a single record (row) from an MSI table.
type TableRow struct {
	Columns []string
}

// ListTables opens the MSI database at msiPath and prints the names of all tables.
func ListTables(msiPath string) error {
	// Initialize COM.
	if err := ole.CoInitialize(0); err != nil {
		return fmt.Errorf("failed to initialize COM: %v", err)
	}
	defer ole.CoUninitialize()

	// Create the Windows Installer object.
	obj, err := oleutil.CreateObject("WindowsInstaller.Installer")
	if err != nil {
		return fmt.Errorf("COM CreateObject error: %v", err)
	}
	defer obj.Release()

	inst, err := obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("QueryInterface error: %v", err)
	}
	defer inst.Release()

	// Open the MSI in read-only mode.
	dbResult, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, 0)
	if err != nil {
		return fmt.Errorf("OpenDatabase error: %v", err)
	}
	db := dbResult.ToIDispatch()
	if db == nil {
		return fmt.Errorf("OpenDatabase returned nil dispatch")
	}
	defer db.Release()

	// Open a view to query the system table that holds all table names.
	viewResult, err := oleutil.CallMethod(db, "OpenView", "SELECT * FROM `_Tables`")
	if err != nil {
		return fmt.Errorf("OpenView error (missing _Tables): %v", err)
	}
	view := viewResult.ToIDispatch()
	if view == nil {
		return fmt.Errorf("OpenView returned nil dispatch")
	}
	defer view.Release()

	// Execute the query.
	if _, err := oleutil.CallMethod(view, "Execute"); err != nil {
		return fmt.Errorf("Execute view error: %v", err)
	}

	fmt.Println("📦 Tables in", msiPath)

	foundAny := false
	// Loop through records.
	for {
		recordResult, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordResult.Value() == nil {
			break // End of records
		}
		record := recordResult.ToIDispatch()
		if record == nil {
			break
		}

		// Fetch the table name from the first column.
		tableNameVariant, err := oleutil.CallMethod(record, "StringData", 1)
		if err == nil && tableNameVariant != nil {
			name := tableNameVariant.ToString()
			fmt.Println("   └─", name)
			foundAny = true
		}
		record.Release()
	}

	if !foundAny {
		fmt.Println("   ⚠ No tables found — MSI may be empty, encrypted, or invalid.")
	}

	return nil
}

// ReadTableRows reads all rows from the specified table in the MSI database.
func ReadTableRows(msiPath, tableName string) ([]TableRow, error) {
	if err := ole.CoInitialize(0); err != nil {
		return nil, fmt.Errorf("failed to initialize COM: %v", err)
	}
	defer ole.CoUninitialize()

	obj, err := oleutil.CreateObject("WindowsInstaller.Installer")
	if err != nil {
		return nil, fmt.Errorf("CreateObject error: %v", err)
	}
	defer obj.Release()

	inst, err := obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("QueryInterface error: %v", err)
	}
	defer inst.Release()

	dbResult, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, 0)
	if err != nil {
		return nil, fmt.Errorf("OpenDatabase error: %v", err)
	}
	db := dbResult.ToIDispatch()
	if db == nil {
		return nil, fmt.Errorf("OpenDatabase returned nil dispatch")
	}
	defer db.Release()

	// Get number of columns using the helper.
	colCount, err := getColumnCount(db, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get column count: %v", err)
	}

	sql := fmt.Sprintf("SELECT * FROM `%s`", tableName)
	viewResult, err := oleutil.CallMethod(db, "OpenView", sql)
	if err != nil {
		return nil, fmt.Errorf("OpenView error (ReadTable): %v", err)
	}
	view := viewResult.ToIDispatch()
	if view == nil {
		return nil, fmt.Errorf("OpenView returned nil dispatch")
	}
	defer view.Release()

	if _, err := oleutil.CallMethod(view, "Execute", nil); err != nil {
		return nil, fmt.Errorf("Execute error (ReadTable): %v", err)
	}

	var rows []TableRow
	for {
		recordResult, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordResult.Value() == nil {
			break
		}
		record := recordResult.ToIDispatch()
		if record == nil {
			break
		}
		var cols []string
		for i := 1; i <= colCount; i++ {
			dataRaw, _ := oleutil.CallMethod(record, "StringData", i)
			val := dataRaw.ToString()
			cols = append(cols, val)
		}
		record.Release()
		rows = append(rows, TableRow{Columns: cols})
	}
	return rows, nil
}

// FormatRows returns a formatted string representing the rows in a tabular layout.
func FormatRows(rows []TableRow) string {
	var sb strings.Builder
	for i, row := range rows {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, strings.Join(row.Columns, " | ")))
	}
	return sb.String()
}

// getColumnCount uses the _Columns system table to count columns for a given table.
func getColumnCount(db *ole.IDispatch, tableName string) (int, error) {
	query := fmt.Sprintf("SELECT * FROM `_Columns` WHERE `Table`='%s'", tableName)
	viewResult, err := oleutil.CallMethod(db, "OpenView", query)
	if err != nil {
		return 0, fmt.Errorf("OpenView error for _Columns: %v", err)
	}
	view := viewResult.ToIDispatch()
	if view == nil {
		return 0, fmt.Errorf("OpenView returned nil dispatch for _Columns")
	}
	defer view.Release()

	if _, err := oleutil.CallMethod(view, "Execute"); err != nil {
		return 0, fmt.Errorf("Execute error for _Columns: %v", err)
	}

	var count int
	for {
		recordResult, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordResult.Value() == nil {
			break
		}
		record := recordResult.ToIDispatch()
		if record != nil {
			record.Release()
		}
		count++
	}
	if count == 0 {
		return 0, fmt.Errorf("no columns found for table '%s'", tableName)
	}
	return count, nil
}
