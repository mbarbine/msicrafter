// core/error_handler.go
package core

import (
	"log"
)

// SafeExecute wraps a function call with deferred recovery to handle panics,
// emulating a try/catch mechanism. It logs errors along with the operation name.
func SafeExecute(operation string, f func() error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] %s: Recovered from panic: %v", operation, r)
		}
	}()
	if err := f(); err != nil {
		log.Printf("[ERROR] %s: %v", operation, err)
	}
}
