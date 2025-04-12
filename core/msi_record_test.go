// core/msi_record_test.go
package core

import (
	"strings"
	"testing"
)

func TestFormatRows(t *testing.T) {
	rows := []TableRow{
		{Columns: []string{"Col1A", "Col2A", "Col3A"}},
		{Columns: []string{"Col1B", "Col2B", "Col3B"}},
	}

	formatted := FormatRows(rows)
	// Expect each row to be formatted with a row number and columns separated by ' | '
	if !strings.Contains(formatted, "[1] Col1A | Col2A | Col3A") {
		t.Errorf("Formatted rows did not contain expected string for row 1. Got: %s", formatted)
	}
	if !strings.Contains(formatted, "[2] Col1B | Col2B | Col3B") {
		t.Errorf("Formatted rows did not contain expected string for row 2. Got: %s", formatted)
	}
}
