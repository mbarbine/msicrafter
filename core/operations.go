// core/operations.go
package core

import (
	"fmt"
)

// CompareMSI compares two MSI files and prints differences.
func CompareMSI(msi1, msi2 string) error {
	return SafeExecute("CompareMSI", func() error {
		session1, err := OpenMsiSession(msi1, 0)
		if err != nil {
			return fmt.Errorf("failed to open MSI1 session: %v", err)
		}
		defer session1.Close()

		session2, err := OpenMsiSession(msi2, 0)
		if err != nil {
			return fmt.Errorf("failed to open MSI2 session: %v", err)
		}
		defer session2.Close()

		tables1, err := discoverTables(session1)
		if err != nil {
			return fmt.Errorf("failed to list tables in MSI1: %v", err)
		}
		tables2, err := discoverTables(session2)
		if err != nil {
			return fmt.Errorf("failed to list tables in MSI2: %v", err)
		}

		// Convert discoveredTable slices to slices of names.
		names1 := discoveredTablesToNames(tables1)
		names2 := discoveredTablesToNames(tables2)

		fmt.Println("Table differences:")
		for _, t := range tables1 {
			if !contains(names2, t.Name) {
				fmt.Printf("Table '%s' in MSI1 but not MSI2\n", t.Name)
			}
		}
		for _, t := range tables2 {
			if !contains(names1, t.Name) {
				fmt.Printf("Table '%s' in MSI2 but not MSI1\n", t.Name)
			}
		}
		return nil
	})
}

// GenerateTransform creates an MST file from two MSI files.
func GenerateTransform(originalMSI, modifiedMSI, outputMST string) error {
	return SafeExecute("GenerateTransform", func() error {
		fmt.Printf("Would generate transform from '%s' and '%s' to '%s'\n", originalMSI, modifiedMSI, outputMST)
		return fmt.Errorf("transform generation not implemented")
	})
}

// contains returns true if the given slice contains the specified item.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// discoveredTablesToNames converts a slice of discoveredTable to a slice of table name strings.
func discoveredTablesToNames(tables []discoveredTable) []string {
	names := make([]string, 0, len(tables))
	for _, dt := range tables {
		names = append(names, dt.Name)
	}
	return names
}
