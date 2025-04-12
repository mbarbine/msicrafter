// core/apply_transform.go
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

// ApplyTransform reads the MST transform file at mstPath and applies its changes to the target MSI.
// If dryRun is true, no changes are committed. If interactive is true, the user will be asked for confirmation before each query.
func ApplyTransform(msiPath string, mstPath string, dryRun bool, interactive bool) error {
	// Open and read the MST file.
	file, err := os.Open(mstPath)
	if err != nil {
		return fmt.Errorf("failed to open MST file: %v", err)
	}
	defer file.Close()

	// Open the target MSI in editable mode.
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

	// Prepare to read the diff lines.
	scanner := bufio.NewScanner(file)
	var queries []string
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines.
		if strings.TrimSpace(line) == "" {
			continue
		}
		op, table, values, err := parseDiffLine(line)
		if err != nil {
			log.Printf("Skipping invalid diff line: %v", err)
			continue
		}
		var query string
		switch op {
		case "+":
			// Build an INSERT query.
			// Assume all values are already in proper order.
			escapedValues := make([]string, len(values))
			for i, v := range values {
				escapedValues[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
			}
			query = fmt.Sprintf("INSERT INTO `%s` VALUES (%s)", table, strings.Join(escapedValues, ", "))
		case "-":
			// Build a DELETE query.
			// Since we don’t have actual column names, we assume them as COL1, COL2, … in order.
			conds := []string{}
			for i, v := range values {
				conds = append(conds, fmt.Sprintf("COL%d='%s'", i+1, strings.ReplaceAll(v, "'", "''")))
			}
			query = fmt.Sprintf("DELETE FROM `%s` WHERE %s", table, strings.Join(conds, " AND "))
		default:
			continue
		}
		queries = append(queries, query)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading MST file: %v", err)
	}

	// Show spinner while processing queries.
	done := make(chan bool)
	go retro.ShowSpinner("Applying transform...", done)

	// Process each query.
	for _, query := range queries {
		// If interactive, ask for confirmation.
		if interactive {
			fmt.Println(retro.Blue + "The following query will be executed:" + retro.Reset)
			fmt.Println(retro.Yellow + query + retro.Reset)
			fmt.Print("Apply this change? (y/n): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				close(done)
				return fmt.Errorf("failed to read input: %v", err)
			}
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				log.Printf("Skipping query: %s", query)
				continue
			}
		}
		// If dry-run, print the query and continue.
		if dryRun {
			log.Printf("[Dry-run] Would execute query: %s", query)
			continue
		}
		// Execute the query.
		viewDisp, err := oleutil.CallMethod(db, "OpenView", query)
		if err != nil {
			close(done)
			return fmt.Errorf("OpenView error for query [%s]: %v", query, err)
		}
		view := viewDisp.ToIDispatch()
		_, err = oleutil.CallMethod(view, "Execute", nil)
		view.Release()
		if err != nil {
			close(done)
			return fmt.Errorf("Execute error for query [%s]: %v", query, err)
		}
	}
	close(done)

	// If not in dry-run mode, commit the changes.
	if !dryRun {
		_, err = oleutil.CallMethod(db, "Commit")
		if err != nil {
			return fmt.Errorf("Commit error: %v", err)
		}
		log.Println("Transform applied and changes committed successfully.")
	} else {
		log.Println("Dry-run enabled: No changes were committed.")
	}
	return nil
}

// parseDiffLine parses a diff line from the MST file.
// Expected format: "+ TableName => value1|value2|value3"
// or                   "- TableName => value1|value2|value3"
func parseDiffLine(line string) (operation string, table string, values []string, err error) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		err = fmt.Errorf("empty line")
		return
	}
	operation = string(line[0])
	if operation != "+" && operation != "-" {
		err = fmt.Errorf("invalid operation: %s", operation)
		return
	}
	parts := strings.SplitN(line[1:], "=>", 2)
	if len(parts) < 2 {
		err = fmt.Errorf("invalid diff line format")
		return
	}
	table = strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])
	if valueStr != "" {
		values = strings.Split(valueStr, "|")
		for i, v := range values {
			values[i] = strings.TrimSpace(v)
		}
	} else {
		values = []string{}
	}
	return
}
