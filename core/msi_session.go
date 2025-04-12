// core/msi_session.go
package core

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// MsiSession manages a single database handle for multiple operations.
type MsiSession struct {
	dbDispatch *ole.IDispatch
	installer  *ole.IDispatch
}

// OpenMsiSession opens the MSI in the specified mode (0 read, 1 direct).
func OpenMsiSession(msiPath string, mode int) (*MsiSession, error) {
	if err := ole.CoInitialize(0); err != nil {
		return nil, fmt.Errorf("failed to initialize COM: %v", err)
	}

	obj, err := oleutil.CreateObject("WindowsInstaller.Installer")
	if err != nil {
		ole.CoUninitialize()
		return nil, fmt.Errorf("CreateObject error: %v", err)
	}
	inst, err := obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		obj.Release()
		ole.CoUninitialize()
		return nil, fmt.Errorf("QueryInterface error: %v", err)
	}

	dbRaw, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, mode)
	if err != nil {
		inst.Release()
		ole.CoUninitialize()
		return nil, fmt.Errorf("OpenDatabase error: %v", err)
	}
	db := dbRaw.ToIDispatch()

	return &MsiSession{
		dbDispatch: db,
		installer:  inst,
	}, nil
}

// Close closes the database and uninitializes COM.
func (s *MsiSession) Close() {
	if s.dbDispatch != nil {
		s.dbDispatch.Release()
	}
	if s.installer != nil {
		s.installer.Release()
	}
	ole.CoUninitialize()
}

// OpenView runs OpenView on the existing session.
func (s *MsiSession) OpenView(sql string) (*ole.IDispatch, error) {
	viewRaw, err := oleutil.CallMethod(s.dbDispatch, "OpenView", sql)
	if err != nil {
		return nil, fmt.Errorf("OpenView error: %v", err)
	}
	view := viewRaw.ToIDispatch()
	return view, nil
}

// Commit commits changes if the DB is in direct mode.
func (s *MsiSession) Commit() error {
	_, err := oleutil.CallMethod(s.dbDispatch, "Commit")
	return err
}
