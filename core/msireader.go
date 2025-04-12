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
		return fmt.Errorf("COM object error: %v", err)
	}
	inst, err := obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("QueryInterface error: %v", err)
	}
	defer inst.Release()

	db, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, 0)
	if err != nil {
		return fmt.Errorf("OpenDatabase error: %v", err)
	}
	dbDispatch := db.ToIDispatch()
	defer dbDispatch.Release()

	viewDisp, err := oleutil.CallMethod(dbDispatch, "OpenView", "SELECT * FROM `_Tables`")
	if err != nil {
		return fmt.Errorf("OpenView error: %v", err)
	}
	view := viewDisp.ToIDispatch()
	defer view.Release()

	_, _ = oleutil.CallMethod(view, "Execute", nil)

	fmt.Println("📦 Tables in", msiPath)
	for {
		recordDisp, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordDisp.Value() == nil {
			break
		}
		record := recordDisp.ToIDispatch()
		tableName, _ := oleutil.CallMethod(record, "StringData", 1)
		fmt.Println("   └─", tableName.ToString())
		record.Release()
	}
	return nil
}
