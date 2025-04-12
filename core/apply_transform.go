// core/apply_transform.go
package core

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"msicrafter/retro"
)

// ApplyTransform reads an MST diff file line by line, then updates the target MSI
// with Insert or Delete queries. If dryRun is true, no commit is performed.
// If interactive is true, we prompt before each query.
func ApplyTransform(msiPath, mstFile string, dryRun, interactive bool) error {
	return SafeExecute("ApplyTransform", func() error {
		// Read lines from MST
		lines, err := loadDiffLines(mstFile)
		if err != nil {
			return err
		}
		if len(lines) == 0 {
			return fmt.Errorf("no diff lines in %s", mstFile)
		}

		// Open MSI session
		session, err := OpenMsiSession(msiPath, 1) // Read-write mode
		if err != nil {
			return fmt.Errorf("failed to open MSI session: %v", err)
		}
		defer session.Close()

		// Build and execute queries
		var queries []string
		for _, line := range lines {
			op, table, vals, e := parseDiffLine(line)
			if e != nil {
				log.Printf("[WARN] skipping invalid diff line '%s': %v", line, e)
				continue
			}
			q := buildSQL(op, table, vals)
			if DebugMode {
				fmt.Printf("[DEBUG] MST line => %s\n -> built query: %s\n", line, q)
			}
			queries = append(queries, q)
		}
		if len(queries) == 0 {
			return fmt.Errorf("no valid queries built from MST")
		}

		done := make(chan bool)
		go retro.ShowSpinner("Applying MST transform...", done)

		for _, q := range queries {
			if interactive && !confirmQuery(q) {
				log.Printf("[INFO] Skipped query: %s", q)
				continue
			}
			if dryRun {
				log.Printf("[DRY-RUN] %s", q)
				continue
			}
			_, err := session.ExecuteQuery(q)
			if err != nil {
				close(done)
				return fmt.Errorf("execute query '%s' failed: %v", q, err)
			}
		}
		close(done)

		if !dryRun {
			if err := session.Commit(); err != nil {
				return fmt.Errorf("commit failed: %v", err)
			}
			log.Println("[INFO] Transform applied and committed.")
		} else {
			log.Println("[INFO] Dry run complete; no changes committed.")
		}
		return nil
	})
}

func loadDiffLines(mstFile string) ([]string, error) {
	f, err := os.Open(mstFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open MST file: %v", err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		txt := strings.TrimSpace(sc.Text())
		if txt != "" {
			lines = append(lines, txt)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("error reading MST: %v", err)
	}
	return lines, nil
}

func parseDiffLine(line string) (op, table string, values []string, err error) {
	line = strings.TrimSpace(line)
	if len(line) < 3 {
		err = fmt.Errorf("invalid line '%s'", line)
		return
	}
	op = string(line[0])
	if op != "+" && op != "-" {
		err = fmt.Errorf("invalid op '%s'; must be + or -", op)
		return
	}
	parts := strings.SplitN(line[1:], "=>", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("missing '=>' in '%s'", line)
		return
	}
	table = strings.TrimSpace(parts[0])
	valStr := strings.TrimSpace(parts[1])
	if table == "" {
		err = fmt.Errorf("no table name found in '%s'", line)
		return
	}
	if valStr == "" {
		values = []string{}
		return
	}
	vals := strings.Split(valStr, "|")
	for i, v := range vals {
		vals[i] = strings.TrimSpace(v)
	}
	values = vals
	return
}

func buildSQL(op, table string, vals []string) string {
	switch op {
	case "+":
		// Insert
		parts := make([]string, len(vals))
		for i, v := range vals {
			parts[i] = fmt.Sprintf("'%s'", escape(v))
		}
		return fmt.Sprintf("INSERT INTO `%s` VALUES (%s)", table, strings.Join(parts, ","))
	case "-":
		// Delete
		// Assume columns are COL1, COL2, etc., for simplicity
		whereParts := make([]string, len(vals))
		for i, v := range vals {
			whereParts[i] = fmt.Sprintf("COL%d='%s'", i+1, escape(v))
		}
		return fmt.Sprintf("DELETE FROM `%s` WHERE %s", table, strings.Join(whereParts, " AND "))
	}
	return ""
}

func confirmQuery(q string) bool {
	fmt.Printf("\nQuery:\n  %s\nApply? (y/n): ", q)
	in := bufio.NewReader(os.Stdin)
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

func escape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}