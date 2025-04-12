// core/msi_reader.go
package core

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// ListTables opens the MSI database at msiPath and prints the names of all tables.
func ListTables(msiPath string) error {
	if err := ole.CoInitialize(0); err != nil {
		return fmt.Errorf("failed to initialize COM: %v", err)
	}
	defer ole.CoUninitialize()

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

	dbResult, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, 0)
	if err != nil {
		return fmt.Errorf("OpenDatabase error: %v", err)
	}
	db := dbResult.ToIDispatch()
	if db == nil {
		return fmt.Errorf("OpenDatabase returned nil dispatch")
	}
	defer db.Release()

	viewResult, err := oleutil.CallMethod(db, "OpenView", "SELECT * FROM `_Tables`")
	if err != nil {
		return fmt.Errorf("OpenView error (missing _Tables): %v", err)
	}
	view := viewResult.ToIDispatch()
	if view == nil {
		return fmt.Errorf("OpenView returned nil dispatch")
	}
	defer view.Release()

	if _, err := oleutil.CallMethod(view, "Execute"); err != nil {
		return fmt.Errorf("Execute view error: %v", err)
	}

	fmt.Println("📦 Tables in", msiPath)

	for {
		recordResult, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordResult.Value() == nil {
			break
		}

		record := recordResult.ToIDispatch()
		if record == nil {
			break
		}

		// Attempt to read StringData safely
		tableNameVariant, err := oleutil.CallMethod(record, "StringData", 1)
		if err == nil && tableNameVariant != nil {
			fmt.Println("   └─", tableNameVariant.ToString())
		}
		record.Release()
	}

	return nil
}
