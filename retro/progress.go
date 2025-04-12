// retro/progress.go
package retro

import (
	"fmt"
	"time"
)

// ShowSpinner displays a retro spinner with the given message until the done channel is closed.
func ShowSpinner(message string, done chan bool) {
	spinner := []string{"|", "/", "-", "\\"}
	i := 0
	fmt.Printf("%s ", message)
	for {
		select {
		case <-done:
			fmt.Print("\r")
			return
		default:
			fmt.Printf("\r%s %s", message, spinner[i%len(spinner)])
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}
