// core/msi_query.go
package core

import (
	"fmt"
	"log"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// QueryMSI opens the MSI database, executes a given SQL query,
// and prints each fetched record in a tab-delimited format.
func QueryMSI(msiPath string, sqlQuery string) error {
	// Initialize COM.
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

	viewDisp, err := oleutil.CallMethod(dbDispatch, "OpenView", sqlQuery)
	if err != nil {
		return fmt.Errorf("OpenView error: %v", err)
	}
	view := viewDisp.ToIDispatch()
	defer view.Release()

	_, err = oleutil.CallMethod(view, "Execute", nil)
	if err != nil {
		return fmt.Errorf("Execute error: %v", err)
	}

	log.Println("🏁 Query Results:")
	for {
		recordDisp, err := oleutil.CallMethod(view, "Fetch")
		if err != nil {
			log.Printf("Fetch error: %v", err)
			break
		}
		if recordDisp.Value() == nil {
			break
		}
		record := recordDisp.ToIDispatch()

		// Try to retrieve FieldCount; if not available, assume 10 fields max.
		fieldCountVar, err := oleutil.GetProperty(record, "FieldCount")
		var fieldCount int
		if err == nil {
			fieldCount = int(fieldCountVar.Val)
		} else {
			fieldCount = 10
		}

		// Print each field separated by tabs.
		for i := 1; i <= fieldCount; i++ {
			field, err := oleutil.CallMethod(record, "StringData", i)
			if err != nil {
				break
			}
			val := field.ToString()
			if val == "" {
				break
			}
			fmt.Printf("[%d] %s\t", i, val)
		}
		fmt.Println()
		record.Release()
	}

	return nil
}
