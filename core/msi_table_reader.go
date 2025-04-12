// core/msi_table_reader.go
package core

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// TableRow represents a single record from an MSI table,
// storing data as a slice of strings for each column.
type TableRow struct {
	Columns []string
}

// ReadTable reads all rows from the specified table in an MSI database
// and returns them as a slice of TableRow.
func ReadTable(msiPath, tableName string) ([]TableRow, error) {
	if err := ole.CoInitialize(0); err != nil {
		return nil, fmt.Errorf("failed to initialize COM: %v", err)
	}
	defer ole.CoUninitialize()

	obj, err := oleutil.CreateObject("WindowsInstaller.Installer")
	if err != nil {
		return nil, fmt.Errorf("CreateObject error: %v", err)
	}
	inst, err := obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("QueryInterface error: %v", err)
	}
	defer inst.Release()

	// Open the MSI in read-only mode.
	dbRaw, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, 0)
	if err != nil {
		return nil, fmt.Errorf("OpenDatabase error: %v", err)
	}
	db := dbRaw.ToIDispatch()
	defer db.Release()

	// We need to see how many columns the table has. We'll query the `_Columns` table.
	colCount, err := getColumnCount(db, tableName)
	if err != nil {
		return nil, err
	}

	// Now, read all rows from the target table.
	sql := fmt.Sprintf("SELECT * FROM `%s`", tableName)
	viewRaw, err := oleutil.CallMethod(db, "OpenView", sql)
	if err != nil {
		return nil, fmt.Errorf("OpenView error (ReadTable): %v", err)
	}
	view := viewRaw.ToIDispatch()
	defer view.Release()

	if _, err := oleutil.CallMethod(view, "Execute", nil); err != nil {
		return nil, fmt.Errorf("Execute error (ReadTable): %v", err)
	}

	var rows []TableRow
	for {
		recordRaw, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordRaw.Value() == nil {
			break // done
		}
		record := recordRaw.ToIDispatch()

		var columns []string
		for i := 1; i <= colCount; i++ {
			dataRaw, _ := oleutil.CallMethod(record, "StringData", i)
			val := dataRaw.ToString()
			columns = append(columns, val)
		}
		record.Release()
		rows = append(rows, TableRow{Columns: columns})
	}

	return rows, nil
}

// getColumnCount uses the _Columns system table to count columns for a given table.
func getColumnCount(db *ole.IDispatch, tableName string) (int, error) {
	query := fmt.Sprintf("SELECT * FROM `_Columns` WHERE `Table`='%s'", tableName)
	viewRaw, err := oleutil.CallMethod(db, "OpenView", query)
	if err != nil {
		return 0, err
	}
	view := viewRaw.ToIDispatch()
	defer view.Release()

	if _, err := oleutil.CallMethod(view, "Execute", nil); err != nil {
		return 0, err
	}

	var count int
	for {
		recordRaw, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordRaw.Value() == nil {
			break
		}
		count++
		record := recordRaw.ToIDispatch()
		record.Release()
	}
	if count == 0 {
		// Could happen if table has no columns or doesn't exist
		return 0, fmt.Errorf("no columns found for table '%s'", tableName)
	}
	return count, nil
}
