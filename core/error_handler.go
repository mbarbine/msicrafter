// core/error_handler.go
package core

import (
	"errors"
	"log"
	"time"
	"fmt"
	"strings"
)

var transientErrors = []string{
	"RPC_E_SERVERFAULT", // example error codes/messages
	"RPC_E_DISCONNECTED",
	// add more as needed
}

// SafeExecuteWithRetry extends SafeExecute to retry up to `maxRetries` times if a known transient error occurs.
func SafeExecuteWithRetry(operation string, maxRetries int, f func() error) (errRet error) {
	var attempt int
	for attempt = 1; attempt <= maxRetries; attempt++ {
		err := SafeExecute(operation, f)
		if err == nil {
			return nil
		}
		// Check if it's transient
		if isTransientError(err) && attempt < maxRetries {
			log.Printf("[WARN] %s: Transient error detected, retrying (attempt %d/%d)...", operation, attempt, maxRetries)
			time.Sleep(2 * time.Second) // backoff
			continue
		}
		return err
	}
	return errors.New("exceeded max retries")
}
func SafeExecute(operation string, f func() error) (errRet error) {
	defer func() {
		if r := recover(); r != nil {
			errRet = fmt.Errorf("[ERROR] %s: Recovered from panic: %v", operation, r)
			log.Printf("[ERROR] %s: Recovered from panic: %v", operation, r)
		}
	}()
	if err := f(); err != nil {
		log.Printf("[ERROR] %s: %v", operation, err)
		return err
	}
	return nil
}

func isTransientError(err error) bool {
	for _, e := range transientErrors {
		if e != "" && ContainsIgnoreCase(err.Error(), e) {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if `substr` is in `str`, ignoring case.
func ContainsIgnoreCase(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 && 
		len(str) >= len(substr) && 
		// naive approach: strings.ToLower() for both
		// or advanced approach: do partial case-insensitive match
		// For demonstration, we do:
		contains(strings.ToLower(str), strings.ToLower(substr))
}

// implement a helper
func contains(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && strings.Contains(s, sub)
}
