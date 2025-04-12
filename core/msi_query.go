// core/msi_query.go
package core

import (
	"fmt"
	"strings"
)

// QueryMSI executes a SQL query on an MSI database and returns the results.
func QueryMSI(msiPath, sqlQuery string) error {
	return SafeExecuteWithRetry("QueryMSI", 3, func() error {
		session, err := OpenMsiSession(msiPath, 0)
		if err != nil {
			return fmt.Errorf("failed to open MSI session: %v", err)
		}
		defer session.Close()

		rows, err := session.ExecuteQuery(sqlQuery)
		if err != nil {
			return fmt.Errorf("query failed: %v", err)
		}

		if len(rows) == 0 {
			fmt.Println("No records found.")
			return nil
		}

		// Print column names if available
		tableName := extractTableName(sqlQuery)
		if tableName != "" {
			if cols, err := session.GetColumnNames(tableName); err == nil {
				fmt.Printf("Columns: %s\n", strings.Join(cols, ", "))
			}
		}

		fmt.Printf("🏁 Query Results (%d rows):\n%s", len(rows), FormatRows(rows))
		return nil
	})
}