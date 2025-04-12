// core/apply_transform_test.go
package core

import (
	"reflect"
	"testing"
)

func TestParseDiffLine_Insert(t *testing.T) {
	// Example: "+ Property => ProductVersion|9.9.9|Author:RetroWizard"
	line := "+ Property => ProductVersion|9.9.9|Author:RetroWizard"
	op, table, values, err := parseDiffLine(line)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if op != "+" {
		t.Errorf("Expected operation '+', got: %s", op)
	}
	if table != "Property" {
		t.Errorf("Expected table 'Property', got: %s", table)
	}
	expectedValues := []string{"ProductVersion", "9.9.9", "Author:RetroWizard"}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("Expected values %v, got: %v", expectedValues, values)
	}
}

func TestParseDiffLine_Delete(t *testing.T) {
	// Example: "- CustomAction => CA1|SomeAction"
	line := "- CustomAction => CA1|SomeAction"
	op, table, values, err := parseDiffLine(line)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if op != "-" {
		t.Errorf("Expected operation '-', got: %s", op)
	}
	if table != "CustomAction" {
		t.Errorf("Expected table 'CustomAction', got: %s", table)
	}
	expectedValues := []string{"CA1", "SomeAction"}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("Expected values %v, got: %v", expectedValues, values)
	}
}

func TestParseDiffLine_Invalid(t *testing.T) {
	// Test an invalid diff line that should produce an error.
	line := "X InvalidLine"
	_, _, _, err := parseDiffLine(line)
	if err == nil {
		t.Errorf("Expected error for invalid line, got nil")
	}
}

// ----------------------------------------------
// New Tests for More Exhaustive Coverage
// ----------------------------------------------

func TestParseDiffLine_EmptyLine(t *testing.T) {
	line := ""
	_, _, _, err := parseDiffLine(line)
	if err == nil {
		t.Errorf("Expected error for empty line, got nil")
	}
}

func TestParseDiffLine_WhitespaceLine(t *testing.T) {
	line := "    "
	_, _, _, err := parseDiffLine(line)
	if err == nil {
		t.Errorf("Expected error for whitespace line, got nil")
	}
}

func TestParseDiffLine_MissingDelimiter(t *testing.T) {
	// No "=>" part
	line := "+ Property  ProductVersion|9.9.9"
	_, _, _, err := parseDiffLine(line)
	if err == nil {
		t.Errorf("Expected error for missing '=>' delimiter, got nil")
	}
}

func TestParseDiffLine_NoTableName(t *testing.T) {
	line := "+  => ProductVersion|9.9.9"
	_, _, values, err := parseDiffLine(line)
	if err == nil {
		t.Errorf("Expected error for missing table name, got nil")
	}
	if len(values) > 0 {
		t.Errorf("Expected no values returned on error, got: %v", values)
	}
}

func TestParseDiffLine_NoValues(t *testing.T) {
	line := "+ Property => "
	op, table, values, err := parseDiffLine(line)
	if err != nil {
		t.Fatalf("Did not expect a parse error, got: %v", err)
	}
	if op != "+" {
		t.Errorf("Expected '+', got: %s", op)
	}
	if table != "Property" {
		t.Errorf("Expected table 'Property', got: %s", table)
	}
	// No values means an empty slice
	if len(values) != 0 {
		t.Errorf("Expected 0 values, got: %v", values)
	}
}

func TestParseDiffLine_InvalidOperation(t *testing.T) {
	line := "* Property => SomeVal"
	_, _, _, err := parseDiffLine(line)
	if err == nil {
		t.Errorf("Expected error for invalid operation '*', got nil")
	}
}

func TestParseDiffLine_LeadingTrailingPipes(t *testing.T) {
	line := "+ Property => |LeadingVal|TrailingVal|"
	op, table, vals, err := parseDiffLine(line)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if op != "+" {
		t.Errorf("Expected '+', got: %s", op)
	}
	if table != "Property" {
		t.Errorf("Expected 'Property', got: %s", table)
	}
	expected := []string{"", "LeadingVal", "TrailingVal", ""}
	if !reflect.DeepEqual(vals, expected) {
		t.Errorf("Expected %v, got %v", expected, vals)
	}
}

