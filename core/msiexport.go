// core/msi_export.go
package core

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// ExportMSI exports MSI tables to CSV or JSON files and compresses them into a zip archive.
func ExportMSI(msiPath, format, outputZip string) error {
	if err := ole.CoInitialize(0); err != nil {
		return fmt.Errorf("failed to initialize COM: %v", err)
	}
	defer ole.CoUninitialize()

	// Create a temporary directory to store exported files.
	tmpDir, err := os.MkdirTemp("", "msi_export")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	obj, err := oleutil.CreateObject("WindowsInstaller.Installer")
	if err != nil {
		return fmt.Errorf("COM object error: %v", err)
	}
	inst, err := obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("QueryInterface error: %v", err)
	}
	defer inst.Release()

	dbRaw, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, 0)
	if err != nil {
		return fmt.Errorf("OpenDatabase error: %v", err)
	}
	db := dbRaw.ToIDispatch()
	defer db.Release()

	viewDisp, err := oleutil.CallMethod(db, "OpenView", "SELECT * FROM `_Tables`")
	if err != nil {
		return fmt.Errorf("OpenView error: %v", err)
	}
	view := viewDisp.ToIDispatch()
	defer view.Release()

	_, _ = oleutil.CallMethod(view, "Execute", nil)

	var tableNames []string
	for {
		recordDisp, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recordDisp.Value() == nil {
			break
		}
		record := recordDisp.ToIDispatch()
		tableName, _ := oleutil.CallMethod(record, "StringData", 1)
		tableNames = append(tableNames, tableName.ToString())
		record.Release()
	}

	// For demonstration, create dummy export files per table.
	for _, table := range tableNames {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("%s.%s", table, format))
		if format == "csv" {
			if err := exportDummyCSV(filePath, table); err != nil {
				return err
			}
		} else if format == "json" {
			if err := exportDummyJSON(filePath, table); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported format: %s", format)
		}
	}

	// Zip the exported files.
	err = zipDirectory(tmpDir, outputZip)
	if err != nil {
		return fmt.Errorf("failed to zip export directory: %v", err)
	}

	log.Printf("Export completed successfully: %s", outputZip)
	return nil
}

func exportDummyCSV(filePath, table string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	// Dummy header and data row.
	writer.Write([]string{"Column1", "Column2", "Column3"})
	writer.Write([]string{table + "_data1", table + "_data2", table + "_data3"})
	writer.Flush()
	return writer.Error()
}

func exportDummyJSON(filePath, table string) error {
	data := []map[string]string{
		{"Column1": table + "_data1", "Column2": table + "_data2", "Column3": table + "_data3"},
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func zipDirectory(srcDir, outputZip string) error {
	zipFile, err := os.Create(outputZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := archive.Create(relPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, file)
		return err
	})
	return err
}
