// core/msi_record.go
package core

import (
	"fmt"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// TableRow represents a single record from an MSI table.
type TableRow struct {
	Columns []string
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
	inst, err := obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("QueryInterface error: %v", err)
	}
	defer inst.Release()

	dbRaw, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, 0)
	if err != nil {
		return nil, fmt.Errorf("OpenDatabase error: %v", err)
	}
	db := dbRaw.ToIDispatch()
	defer db.Release()

	// Get the number of columns in the table. We reuse getColumnCount from msi_table_reader.go.
	colCount, err := getColumnCount(db, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get column count: %v", err)
	}

	sql := fmt.Sprintf("SELECT * FROM `%s`", tableName)
	viewRaw, err := oleutil.CallMethod(db, "OpenView", sql)
	if err != nil {
		return nil, fmt.Errorf("OpenView error: %v", err)
	}
	view := viewRaw.ToIDispatch()
	defer view.Release()

	if _, err := oleutil.CallMethod(view, "Execute", nil); err != nil {
		return nil, fmt.Errorf("Execute error: %v", err)
	}

	var rows []TableRow
	for {
		recordRaw, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordRaw.Value() == nil {
			break
		}
		record := recordRaw.ToIDispatch()
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
