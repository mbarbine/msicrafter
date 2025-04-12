// core/error_handler.go
package core

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// DebugMode toggles verbose logging. Set in main() via a flag.
var DebugMode bool = false

// transientErrors defines known retryable COM-related errors.
// These should be updated based on real-world testing.
var transientErrors = []string{
	"RPC_E_SERVERFAULT",
	"RPC_E_DISCONNECTED",
	"RPC_S_CALL_FAILED",
	"RPC_E_CALL_REJECTED",
}

// SafeExecute wraps a function call with panic recovery and logging.
// It logs the operation and any error returned or panic recovered.
func SafeExecute(operation string, f func() error) (errRet error) {
	defer func() {
		if r := recover(); r != nil {
			errRet = fmt.Errorf("[PANIC] %s: %v", operation, r)
			log.Printf("[PANIC] %s: %v", operation, r)
		}
	}()
	if DebugMode {
		log.Printf("[DEBUG] Starting operation: %s", operation)
	}
	if err := f(); err != nil {
		log.Printf("[ERROR] %s: %v", operation, err)
		return err
	}
	if DebugMode {
		log.Printf("[DEBUG] Completed operation: %s", operation)
	}
	return nil
}

// SafeExecuteWithRetry executes a function with panic protection and transient error retry.
// It retries transient errors (based on substring match) up to `maxRetries` times.
func SafeExecuteWithRetry(operation string, maxRetries int, f func() error) (errRet error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := SafeExecute(operation, f)
		if err == nil {
			return nil
		}
		if isTransientError(err) && attempt < maxRetries {
			log.Printf("[WARN] %s: Transient error, retrying (%d/%d)...", operation, attempt, maxRetries)
			time.Sleep(2 * time.Second) // exponential backoff could go here
			continue
		}
		return err
	}
	return errors.New("exceeded max retries for operation: " + operation)
}

// isTransientError checks if the error string contains a known transient substring.
func isTransientError(err error) bool {
	msg := strings.ToUpper(err.Error())
	for _, token := range transientErrors {
		if strings.Contains(msg, token) {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase returns true if substr is found in str (case-insensitive).
func ContainsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}
