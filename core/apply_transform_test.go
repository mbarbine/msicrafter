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
