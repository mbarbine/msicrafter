// core/msi_record_edit.go
package core

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"msicrafter/retro"
)

// EditRecord updates a single record in the specified table based on its row number.
// The setClause is expected in the format "field1=value1,field2=value2,..."
// This function uses the first column of the selected row as a stand-in for the primary key.
// dryRun simulates the operation without committing the changes;
// interactive mode prompts the user for confirmation before executing the query.
func EditRecord(msiPath, table string, recordNumber int, setClause string, dryRun bool, interactive bool) error {
	// Read all rows for the specified table.
	rows, err := ReadTableRows(msiPath, table)
	if err != nil {
		return fmt.Errorf("failed to read table rows: %v", err)
	}

	if recordNumber < 1 || recordNumber > len(rows) {
		return fmt.Errorf("record number %d is out of range; table '%s' has %d records", recordNumber, table, len(rows))
	}

	// Get the target record (using 1-based indexing).
	targetRow := rows[recordNumber-1]
	if len(targetRow.Columns) == 0 {
		return fmt.Errorf("the selected record has no columns")
	}
	primaryKey := targetRow.Columns[0] // Assumed primary key (for demo purposes).

	// Parse the setClause into a map.
	fields := map[string]string{}
	assignments := strings.Split(setClause, ",")
	for _, assignment := range assignments {
		parts := strings.SplitN(assignment, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid set clause format; expected field=value")
		}
		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		fields[field] = value
	}

	// Validate using our generic validation logic.
	if err := ValidateEdit(table, fields); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	// Build the SET clause for the update.
	var setParts []string
	for field, value := range fields {
		// Escape any single quotes for SQL compatibility.
		escapedValue := strings.ReplaceAll(value, "'", "''")
		setParts = append(setParts, fmt.Sprintf("%s='%s'", field, escapedValue))
	}
	setClauseSQL := strings.Join(setParts, ", ")

	// For this demo we assume the primary key is in the first column.
	// We use a placeholder column name "COL1" for the primary key column.
	primaryKeyEscaped := strings.ReplaceAll(primaryKey, "'", "''")
	whereClause := fmt.Sprintf("COL1='%s'", primaryKeyEscaped)

	// Construct the UPDATE SQL statement.
	updateSQL := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", table, setClauseSQL, whereClause)
	log.Printf("Constructed update SQL: %s", updateSQL)

	// If interactive mode is enabled, show the SQL and ask for confirmation.
	if interactive {
		fmt.Println(retro.Blue + "The following update will be executed:" + retro.Reset)
		fmt.Println(retro.Yellow + updateSQL + retro.Reset)
		fmt.Print("Apply this change? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %v", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("update aborted by user")
		}
	}

	// Initialize COM and open the database in editable mode.
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

	dbRaw, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, 1)
	if err != nil {
		return fmt.Errorf("OpenDatabase error: %v", err)
	}
	db := dbRaw.ToIDispatch()
	defer db.Release()

	viewDisp, err := oleutil.CallMethod(db, "OpenView", updateSQL)
	if err != nil {
		return fmt.Errorf("OpenView error: %v", err)
	}
	view := viewDisp.ToIDispatch()
	defer view.Release()

	// Show a retro spinner during execution.
	done := make(chan bool)
	go retro.ShowSpinner("Editing record...", done)

	_, err = oleutil.CallMethod(view, "Execute", nil)
	close(done)
	if err != nil {
		return fmt.Errorf("Execute error: %v", err)
	}

	// If dry-run mode is specified, do not commit the change.
	if dryRun {
		log.Println("Dry-run enabled: Changes simulated and not committed.")
		return nil
	}

	_, err = oleutil.CallMethod(db, "Commit")
	if err != nil {
		return fmt.Errorf("Commit error: %v", err)
	}

	log.Printf("Record %d in table '%s' updated successfully in %s", recordNumber, table, msiPath)
	return nil
}
