// retro/colors.go
package retro

import (
	"fmt"
)

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Yellow = "\033[33m"
	Reset  = "\033[0m"
)

// ShowSuccess prints a success message in green.
func ShowSuccess(message string) {
	fmt.Printf("%s[SUCCESS] %s%s\n", Green, message, Reset)
}

// ShowError prints an error message in red.
func ShowError(message string) {
	fmt.Printf("%s[ERROR] %s%s\n", Red, message, Reset)
}

// ShowInfo prints an informational message in blue.
func ShowInfo(message string) {
	fmt.Printf("%s[INFO] %s%s\n", Blue, message, Reset)
}

// ShowWarning prints a warning message in yellow.
func ShowWarning(message string) {
	fmt.Printf("%s[WARNING] %s%s\n", Yellow, message, Reset)
}
