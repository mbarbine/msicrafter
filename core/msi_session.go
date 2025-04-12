package core

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// comState tracks global COM initialization.
var (
	comMutex       sync.Mutex
	comInitCount   int
	comInitialized bool
)

// InitCOM initializes the COM library for the application.
func InitCOM() error {
	comMutex.Lock()
	defer comMutex.Unlock()

	if comInitCount == 0 {
		if err := ole.CoInitialize(0); err != nil {
			return fmt.Errorf("failed to initialize COM: %v", err)
		}
		comInitialized = true
		if DebugMode {
			logInfo("COM initialized globally")
		}
	}
	comInitCount++
	return nil
}

// CleanupCOM releases COM resources.
func CleanupCOM() error {
	comMutex.Lock()
	defer comMutex.Unlock()

	if comInitCount == 0 {
		return nil // Already cleaned up or never initialized
	}

	comInitCount--
	if comInitCount == 0 && comInitialized {
		ole.CoUninitialize()
		comInitialized = false
		if DebugMode {
			logInfo("COM cleaned up globally")
		}
	}
	return nil
}

// MsiSession manages a single MSI database handle.
type MsiSession struct {
	dbDispatch *ole.IDispatch
	installer  *ole.IDispatch
	msiPath    string
	mode       int
	closed     bool
	localCOM   bool // Tracks if this session initialized COM
}

// OpenMsiSession opens an MSI database in the specified mode (0=read-only, 1=read-write).
func OpenMsiSession(msiPath string, mode int) (*MsiSession, error) {
	var session *MsiSession
	err := SafeExecuteWithRetry("OpenMsiSession", 3, func() error {
		if mode != 0 && mode != 1 {
			return fmt.Errorf("invalid mode %d: must be 0 (read-only) or 1 (read-write)", mode)
		}

		// Check if COM is already initialized globally
		comMutex.Lock()
		localCOM := !comInitialized
		comMutex.Unlock()

		if localCOM {
			if err := ole.CoInitialize(0); err != nil {
				return fmt.Errorf("failed to initialize COM: %v", err)
			}
		}

		obj, err := oleutil.CreateObject("WindowsInstaller.Installer")
		if err != nil {
			if localCOM {
				ole.CoUninitialize()
			}
			return fmt.Errorf("failed to create WindowsInstaller: %v", err)
		}
		defer func() {
			if err != nil {
				obj.Release()
				if localCOM {
					ole.CoUninitialize()
				}
			}
		}()

		inst, err := obj.QueryInterface(ole.IID_IDispatch)
		if err != nil {
			return fmt.Errorf("failed to query interface: %v", err)
		}

		dbRaw, err := oleutil.CallMethod(inst, "OpenDatabase", msiPath, mode)
		if err != nil {
			inst.Release()
			return fmt.Errorf("failed to open database '%s': %v", msiPath, err)
		}
		db := dbRaw.ToIDispatch()
		if db == nil {
			inst.Release()
			return fmt.Errorf("open database '%s' returned nil", msiPath)
		}

		session = &MsiSession{
			dbDispatch: db,
			installer:  inst,
			msiPath:    msiPath,
			mode:       mode,
			localCOM:   localCOM,
		}
		if DebugMode {
			logInfo(fmt.Sprintf("Opened MSI session for '%s' (mode=%d, localCOM=%v)", msiPath, mode, localCOM))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Close releases COM resources for this session.
func (s *MsiSession) Close() error {
	if s.closed {
		return nil
	}
	return SafeExecute("CloseMsiSession", func() error {
		if s.dbDispatch != nil {
			s.dbDispatch.Release()
			s.dbDispatch = nil
		}
		if s.installer != nil {
			s.installer.Release()
			s.installer = nil
		}
		if s.localCOM {
			ole.CoUninitialize()
			if DebugMode {
				logInfo(fmt.Sprintf("Closed local COM for '%s'", s.msiPath))
			}
		}
		s.closed = true
		if DebugMode {
			logInfo(fmt.Sprintf("Closed MSI session for '%s'", s.msiPath))
		}
		return nil
	})
}

// ExecuteQuery runs a SQL query and returns the results.
func (s *MsiSession) ExecuteQuery(sql string) ([]TableRow, error) {
	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}

	view, err := s.openView(sql)
	if err != nil {
		return nil, err
	}
	defer s.closeView(view)

	colCount, err := s.getColumnCount(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get column count for '%s': %v", sql, err)
	}
	if DebugMode {
		logInfo(fmt.Sprintf("Query '%s' has %d columns", sql, colCount))
	}

	if _, err := oleutil.CallMethod(view, "Execute"); err != nil {
		return nil, fmt.Errorf("failed to execute query '%s': %v", sql, err)
	}

	var rows []TableRow
	for {
		recRaw, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recRaw.Value() == nil {
			if err != nil && DebugMode {
				logWarn(fmt.Sprintf("Fetch error for '%s': %v", sql, err))
			}
			break
		}
		rec := recRaw.ToIDispatch()
		if rec == nil {
			if DebugMode {
				logWarn(fmt.Sprintf("Fetch returned nil dispatch for '%s'", sql))
			}
			continue
		}

		var cols []string
		for i := 1; i <= colCount; i++ {
			valRaw, err := oleutil.CallMethod(rec, "StringData", i)
			if err != nil || valRaw == nil {
				if DebugMode && err != nil {
					logWarn(fmt.Sprintf("StringData(%d) error for '%s': %v", i, sql, err))
				}
				cols = append(cols, "")
				continue
			}
			cols = append(cols, valRaw.ToString())
		}
		rec.Release()
		rows = append(rows, TableRow{Columns: cols})
	}
	if DebugMode && len(rows) > 100 {
		logInfo(fmt.Sprintf("Fetched %d rows for '%s'", len(rows), sql))
	}
	return rows, nil
}

// openView creates a new view for a SQL query.
func (s *MsiSession) openView(sql string) (*ole.IDispatch, error) {
	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}
	var view *ole.IDispatch
	err := SafeExecute("OpenView", func() error {
		viewRaw, err := oleutil.CallMethod(s.dbDispatch, "OpenView", sql)
		if err != nil {
			return fmt.Errorf("failed to open view for '%s': %v", sql, err)
		}
		view = viewRaw.ToIDispatch()
		if view == nil {
			return fmt.Errorf("open view for '%s' returned nil", sql)
		}
		if DebugMode {
			logInfo(fmt.Sprintf("Opened view for query '%s' on '%s'", sql, s.msiPath))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

// closeView closes and releases a view.
func (s *MsiSession) closeView(view *ole.IDispatch) error {
	if view == nil {
		return nil
	}
	return SafeExecute("CloseView", func() error {
		if _, err := oleutil.CallMethod(view, "Close"); err != nil && DebugMode {
			logWarn(fmt.Sprintf("Failed to close view for '%s': %v", s.msiPath, err))
		}
		view.Release()
		return nil
	})
}

// Commit saves changes to the database.
func (s *MsiSession) Commit() error {
	if s.closed {
		return fmt.Errorf("session is closed")
	}
	if s.mode != 1 {
		return fmt.Errorf("commit not allowed in read-only mode")
	}
	return SafeExecute("CommitMsiSession", func() error {
		_, err := oleutil.CallMethod(s.dbDispatch, "Commit")
		if err != nil {
			return fmt.Errorf("failed to commit changes for '%s': %v", s.msiPath, err)
		}
		if DebugMode {
			logInfo(fmt.Sprintf("Committed changes for '%s'", s.msiPath))
		}
		return nil
	})
}

// GetColumnNames retrieves column names for a table.
func (s *MsiSession) GetColumnNames(tableName string) ([]string, error) {
	rows, err := s.ExecuteQuery(fmt.Sprintf("SELECT `Column` FROM `_Columns` WHERE `Table`='%s'", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for '%s': %v", tableName, err)
	}
	cols := make([]string, 0, len(rows))
	for _, row := range rows {
		if len(row.Columns) > 0 && row.Columns[0] != "" {
			cols = append(cols, row.Columns[0])
		}
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("no columns found for '%s'", tableName)
	}
	if DebugMode {
		logInfo(fmt.Sprintf("Retrieved %d columns for '%s'", len(cols), tableName))
	}
	return cols, nil
}

// getColumnCount determines the number of columns for a query.
func (s *MsiSession) getColumnCount(sql string) (int, error) {
	var colCount int
	err := SafeExecute("GetColumnCount", func() error {
		tableName := extractTableName(sql)
		if tableName != "" && s.mode == 0 {
			rows, err := s.ExecuteQuery(fmt.Sprintf("SELECT COUNT(*) FROM `_Columns` WHERE `Table`='%s'", tableName))
			if err == nil && len(rows) > 0 && len(rows[0].Columns) > 0 {
				if count, err := strconv.Atoi(rows[0].Columns[0]); err == nil && count >= 0 {
					colCount = count
					if DebugMode {
						logInfo(fmt.Sprintf("Column count for '%s' via _Columns: %d", tableName, colCount))
					}
					return nil
				}
			}
			if DebugMode && err != nil {
				logWarn(fmt.Sprintf("Failed to count columns via _Columns for '%s': %v", tableName, err))
			}
		}

		view, err := s.openView(sql)
		if err != nil {
			return err
		}
		defer s.closeView(view)

		if _, err := oleutil.CallMethod(view, "Execute"); err != nil {
			return fmt.Errorf("execute view for column count failed: %v", err)
		}
		recRaw, err := oleutil.CallMethod(view, "Fetch")
		if err != nil || recRaw.Value() == nil {
			colCount = 0
			if DebugMode {
				logInfo(fmt.Sprintf("Assuming 0 columns for query '%s'", sql))
			}
			return nil
		}
		rec := recRaw.ToIDispatch()
		if rec == nil {
			return fmt.Errorf("fetch returned nil dispatch")
		}
		defer rec.Release()

		fieldCount, err := oleutil.GetProperty(rec, "FieldCount")
		if err != nil {
			return fmt.Errorf("get FieldCount failed: %v", err)
		}
		colCount = int(fieldCount.Val)
		if DebugMode {
			logInfo(fmt.Sprintf("Column count for '%s' via FieldCount: %d", sql, colCount))
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return colCount, nil
}

// EditTable updates rows in a table based on a set clause and optional where clause.
func (s *MsiSession) EditTable(tableName, setClause, whereClause string, dryRun, interactive bool) error {
	if s.closed {
		return fmt.Errorf("session is closed")
	}
	if s.mode != 1 {
		return fmt.Errorf("edit not allowed in read-only mode")
	}
	return SafeExecute("EditTable", func() error {
		setPairs := strings.Split(setClause, ",")
		var setFields []string
		for _, pair := range setPairs {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid set clause: %s", pair)
			}
			setFields = append(setFields, fmt.Sprintf("`%s`='%s'", parts[0], parts[1]))
		}
		sql := fmt.Sprintf("UPDATE `%s` SET %s", tableName, strings.Join(setFields, ", "))
		if whereClause != "" {
			sql += fmt.Sprintf(" WHERE %s", whereClause)
		}

		if dryRun || interactive {
			previewSQL := fmt.Sprintf("SELECT * FROM `%s`", tableName)
			if whereClause != "" {
				previewSQL += fmt.Sprintf(" WHERE %s", whereClause)
			}
			rows, err := s.ExecuteQuery(previewSQL)
			if err != nil {
				return fmt.Errorf("failed to preview changes: %v", err)
			}
			fmt.Printf("Preview changes for '%s':\n%s\n", tableName, FormatRows(rows))
		}

		if interactive {
			fmt.Print("Apply changes? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				return fmt.Errorf("update cancelled by user")
			}
		}

		if !dryRun {
			view, err := s.openView(sql)
			if err != nil {
				return fmt.Errorf("failed to prepare update: %v", err)
			}
			defer s.closeView(view)
			if _, err := oleutil.CallMethod(view, "Execute"); err != nil {
				return fmt.Errorf("failed to execute update: %v", err)
			}
			return s.Commit()
		}
		return nil
	})
}

// EditTable is a convenience function to edit a table without manually managing a session.
func EditTable(msiPath, tableName, setClause, whereClause string, dryRun, interactive bool) error {
	return SafeExecute("EditTable", func() error {
		session, err := OpenMsiSession(msiPath, 1)
		if err != nil {
			return fmt.Errorf("failed to open MSI session: %v", err)
		}
		defer session.Close()

		err = session.EditTable(tableName, setClause, whereClause, dryRun, interactive)
		if err != nil {
			return err
		}
		return nil
	})
}

// EditRecord updates a specific row in a table.
func (s *MsiSession) EditRecord(tableName string, rowNum int, setClause string, dryRun, interactive bool) error {
	if s.closed {
		return fmt.Errorf("session is closed")
	}
	if s.mode != 1 {
		return fmt.Errorf("edit not allowed in read-only mode")
	}
	return SafeExecute("EditRecord", func() error {
		rows, err := s.ExecuteQuery(fmt.Sprintf("SELECT * FROM `%s`", tableName))
		if err != nil {
			return fmt.Errorf("failed to fetch table '%s': %v", tableName, err)
		}
		if rowNum < 1 || rowNum > len(rows) {
			return fmt.Errorf("invalid row number %d; table has %d rows", rowNum, len(rows))
		}

		cols, err := s.GetColumnNames(tableName)
		if err != nil {
			return fmt.Errorf("failed to get columns for '%s': %v", tableName, err)
		}
		if len(cols) == 0 {
			return fmt.Errorf("no columns found for '%s'", tableName)
		}
		pkColumn := cols[0]
		pkValue := rows[rowNum-1].Columns[0]
		if pkValue == "" {
			return fmt.Errorf("primary key value is empty for row %d", rowNum)
		}

		setPairs := strings.Split(setClause, ",")
		var setFields []string
		for _, pair := range setPairs {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid set clause: %s", pair)
			}
			setFields = append(setFields, fmt.Sprintf("`%s`='%s'", parts[0], parts[1]))
		}
		sql := fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`='%s'", tableName, strings.Join(setFields, ", "), pkColumn, pkValue)

		if dryRun || interactive {
			fmt.Printf("Preview: Would update row %d in '%s':\n%s\n", rowNum, tableName, FormatRows([]TableRow{rows[rowNum-1]}))
		}

		if interactive {
			fmt.Print("Apply changes? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				return fmt.Errorf("update cancelled by user")
			}
		}

		if !dryRun {
			view, err := s.openView(sql)
			if err != nil {
				return fmt.Errorf("failed to prepare update: %v", err)
			}
			defer s.closeView(view)
			if _, err := oleutil.CallMethod(view, "Execute"); err != nil {
				return fmt.Errorf("failed to execute update: %v", err)
			}
			return s.Commit()
		}
		return nil
	})
}

// extractTableName parses the table name from a SQL query.
func extractTableName(sql string) string {
	sql = strings.ToUpper(strings.TrimSpace(sql))
	if strings.HasPrefix(sql, "SELECT") || strings.HasPrefix(sql, "UPDATE") {
		fromIdx := strings.Index(sql, "FROM")
		if fromIdx >= 0 {
			rest := strings.TrimSpace(sql[fromIdx+4:])
			if strings.HasPrefix(rest, "`") {
				endIdx := strings.Index(rest[1:], "`")
				if endIdx >= 0 {
					return rest[1 : endIdx+1]
				}
			} else {
				endIdx := strings.Index(rest, " ")
				if endIdx >= 0 {
					return rest[:endIdx]
				}
				return rest
			}
		}
	}
	return ""
}