// core/backup.go
package core

import (
	"fmt"
	"io"
	"os"
	"time"
)

// BackupMSI creates a backup copy of the given MSI file,
// naming it with the original filename and a timestamp.
func BackupMSI(msiPath string) (string, error) {
	backupPath := fmt.Sprintf("%s.bak.%s", msiPath, time.Now().Format("20060102_150405"))
	srcFile, err := os.Open(msiPath)
	if err != nil {
		return "", fmt.Errorf("failed to open MSI for backup: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy MSI to backup: %v", err)
	}
	return backupPath, nil
}
