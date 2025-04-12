package core

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"
)

// DebugMode toggles verbose logging for the core package.
// Set via core.DebugMode = true in main() or environment overrides.
var DebugMode bool = false

// transientErrors defines known retryable COM-related or RPC errors.
var transientErrors = []string{
	"RPC_E_SERVERFAULT",
	"RPC_E_DISCONNECTED",
	"RPC_S_CALL_FAILED",
	"RPC_E_CALL_REJECTED",
	"CO_E_SERVER_EXEC_FAILURE",
}

// randSource provides a thread-safe random number generator.
var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
var randMu sync.Mutex

// SafeExecute wraps a function call with panic recovery, timing, and logging.
func SafeExecute(operation string, f func() error) (err error) {
	start := time.Now()
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%s: panic: %v", operation, rec)
			logError(operation, err, true)
		}
		if DebugMode && err == nil {
			duration := time.Since(start)
			logInfo(fmt.Sprintf("%s completed in %v", operation, duration))
		}
	}()

	if DebugMode {
		logInfo(fmt.Sprintf("starting %s", operation))
	}

	err = f()
	if err != nil {
		err = fmt.Errorf("%s: %w", operation, err)
		logError(operation, err, false)
	}
	return err
}

// SafeExecuteWithRetry retries f up to maxRetries times for transient errors.
func SafeExecuteWithRetry(operation string, maxRetries int, f func() error) error {
	if maxRetries < 1 {
		return fmt.Errorf("%s: invalid maxRetries: %d", operation, maxRetries)
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := SafeExecute(operation, f)
		if err == nil {
			return nil
		}
		lastErr = err
		if isTransientError(err) && attempt < maxRetries {
			backoff := backoffDuration(attempt)
			logWarn(fmt.Sprintf("%s: transient error (%v), retrying in %v (attempt %d/%d)",
				operation, err, backoff, attempt, maxRetries))
			time.Sleep(backoff)
			continue
		}
		break
	}
	return fmt.Errorf("%s: failed after %d attempts: %w", operation, maxRetries, lastErr)
}

// isTransientError checks if err contains known transient error substrings.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToUpper(err.Error())
	for _, token := range transientErrors {
		if strings.Contains(msg, token) {
			return true
		}
	}
	return false
}

// backoffDuration returns an exponential backoff with random jitter.
func backoffDuration(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	randMu.Lock()
	jitter := time.Duration(randSource.Intn(200)) * time.Millisecond
	randMu.Unlock()
	base := time.Second << (attempt - 1) // 1s, 2s, 4s, ...
	return base + jitter
}

// logError logs an error with optional stack info in debug mode.
// isPanic indicates if it was a panic event.
func logError(operation string, err error, isPanic bool) {
	if DebugMode {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			prefix := "ERROR"
			if isPanic {
				prefix = "PANIC"
			}
			msg := fmt.Sprintf("%s:%d %s: %v", file, line, operation, err)
			log.Printf("[%s] %s", prefix, msg)
			structuredLog(prefix, operation, msg)
			return
		}
	}
	// Fallback
	prefix := "ERROR"
	if isPanic {
		prefix = "PANIC"
	}
	msg := fmt.Sprintf("%s: %v", operation, err)
	log.Printf("[%s] %s", prefix, msg)
	structuredLog(prefix, operation, msg)
}

// logInfo prints an info message in debug mode.
func logInfo(msg string) {
	if !DebugMode {
		return
	}
	log.Printf("[DEBUG] %s", msg)
	structuredLog("DEBUG", "", msg)
}

// logWarn prints a warning message.
func logWarn(msg string) {
	log.Printf("[WARN] %s", msg)
	structuredLog("WARN", "", msg)
}

// structuredLog outputs JSON logs for external systems.
func structuredLog(level, operation, message string) {
	if !DebugMode && level != "WARN" {
		return
	}
	entry := map[string]string{
		"level":     level,
		"operation": operation,
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"message":   message,
	}
	// Use json.MarshalIndent for readability in debug mode
	if DebugMode {
		raw, err := json.MarshalIndent(entry, "", "  ")
		if err == nil {
			log.Printf("[JSON] %s", raw)
			return
		}
		if DebugMode {
			log.Printf("[DEBUG] JSON marshal failed: %v", err)
		}
	}
	// Fallback to compact JSON
	raw, _ := json.Marshal(entry)
	log.Printf("[JSON] %s", raw)
}

// ContainsIgnoreCase checks if substr is found in str, ignoring case.
func ContainsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}