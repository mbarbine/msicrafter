// core/msi_validations.go
package core

import (
	"fmt"
	"strings"
)

// ValidateEdit checks if the fields to be edited are valid.
// This stub can be extended with actual validation rules.
func ValidateEdit(table string, fields map[string]string) error {
	if table == "" {
		return fmt.Errorf("table name cannot be empty")
	}
	for field, value := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("field name cannot be empty")
		}
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("value for field %s cannot be empty", field)
		}
	}
	return nil
}
