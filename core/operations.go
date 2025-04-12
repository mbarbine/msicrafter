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

		fmt.Println("Table differences:")
		for _, t := range tables1 {
			if !contains(tables2, t) {
				fmt.Printf("Table '%s' in MSI1 but not MSI2\n", t)
			}
		}
		for _, t := range tables2 {
			if !contains(tables1, t) {
				fmt.Printf("Table '%s' in MSI2 but not MSI1\n", t)
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}