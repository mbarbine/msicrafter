// core/msi_edit.go
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

// EditTable updates records in the specified table based on the setClause.
// The function now supports two new modes:
// - dryRun: Simulate the update without committing changes.
// - interactive: Display the update SQL and ask for confirmation before applying.
func EditTable(msiPath, tableName, setClause string, dryRun bool, interactive bool) error {
	// Parse the setClause into a field=>value map.
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

	// Validate the edit using our validation routine.
	if err := ValidateEdit(tableName, fields); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	// Build an UPDATE SQL statement.
	setParts := []string{}
	for field, value := range fields {
		setParts = append(setParts, fmt.Sprintf("%s='%s'", field, value))
	}
	updateSQL := fmt.Sprintf("UPDATE `%s` SET %s", tableName, strings.Join(setParts, ", "))
	log.Printf("Constructed update SQL: %s", updateSQL)

	// If interactive mode is enabled, display the SQL and ask for confirmation.
	if interactive {
		fmt.Println(retro.Blue + "The following update will be executed:" + retro.Reset)
		fmt.Println(retro.Yellow + updateSQL + retro.Reset)
		fmt.Print("Confirm changes? (y/n): ")
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

	// Open database in direct (editable) mode.
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

	// Start a spinner to indicate processing.
	done := make(chan bool)
	go retro.ShowSpinner("Updating table...", done)

	_, err = oleutil.CallMethod(view, "Execute", nil)
	if err != nil {
		close(done)
		return fmt.Errorf("Execute error: %v", err)
	}

	// If dry-run, do not commit.
	if dryRun {
		close(done)
		log.Println("Dry-run enabled: Changes simulated and not committed.")
		return nil
	}

	// Commit the changes.
	_, err = oleutil.CallMethod(db, "Commit")
	close(done)
	if err != nil {
		return fmt.Errorf("Commit error: %v", err)
	}

	log.Printf("Table '%s' updated successfully in %s", tableName, msiPath)
	return nil
}
