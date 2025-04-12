// core/msi_windows_api.go
// +build windows

package core

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	msiDLL                 = windows.NewLazySystemDLL("msi.dll")
	procMsiOpenDatabaseW   = msiDLL.NewProc("MsiOpenDatabaseW")
	procMsiDatabaseOpenViewW = msiDLL.NewProc("MsiDatabaseOpenViewW")
	procMsiViewExecute     = msiDLL.NewProc("MsiViewExecute")
	procMsiViewFetch       = msiDLL.NewProc("MsiViewFetch")
	procMsiRecordGetStringW = msiDLL.NewProc("MsiRecordGetStringW")
	procMsiCloseHandle     = msiDLL.NewProc("MsiCloseHandle")
)

const (
	MSIDBOPEN_READONLY = 0
)

type MsiHandle uintptr

// NativeMsiQueryTables uses the Windows API directly to list MSI tables
func NativeMsiQueryTables(msiPath string) ([]string, error) {
	pathPtr, err := windows.UTF16PtrFromString(msiPath)
	if err != nil {
		return nil, fmt.Errorf("UTF16 conversion failed: %w", err)
	}

	var dbHandle MsiHandle
	r, _, err := procMsiOpenDatabaseW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("0"))),
		uintptr(unsafe.Pointer(&dbHandle)),
	)
	if r != 0 {
		return nil, fmt.Errorf("MsiOpenDatabaseW failed: %v", err)
	}
	defer procMsiCloseHandle.Call(uintptr(dbHandle))

	query := syscall.StringToUTF16Ptr("SELECT `Name` FROM `_Tables`")
	var viewHandle MsiHandle
	r, _, err = procMsiDatabaseOpenViewW.Call(uintptr(dbHandle), uintptr(unsafe.Pointer(query)), uintptr(unsafe.Pointer(&viewHandle)))
	if r != 0 {
		return nil, fmt.Errorf("MsiDatabaseOpenViewW failed: %v", err)
	}
	defer procMsiCloseHandle.Call(uintptr(viewHandle))

	r, _, err = procMsiViewExecute.Call(uintptr(viewHandle), 0)
	if r != 0 {
		return nil, fmt.Errorf("MsiViewExecute failed: %v", err)
	}

	var tableNames []string
	for {
		var recHandle MsiHandle
		r, _, _ = procMsiViewFetch.Call(uintptr(viewHandle), uintptr(unsafe.Pointer(&recHandle)))
		if r != 0 {
			break // No more items or error
		}
		if recHandle == 0 {
			break
		}
		defer procMsiCloseHandle.Call(uintptr(recHandle))

		buf := make([]uint16, 256)
		bufLen := uint32(len(buf))
		r, _, _ = procMsiRecordGetStringW.Call(uintptr(recHandle), 1, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&bufLen)))
		if r == 0 {
			tableNames = append(tableNames, syscall.UTF16ToString(buf))
		}
	}

	return tableNames, nil
}
