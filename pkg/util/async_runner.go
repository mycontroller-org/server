package util

import (
	"time"
)

// AsyncRunner func
// executes given func on the specified interval
func AsyncRunner(customFunc func(), execInterval time.Duration, stop chan bool) {
	ticker := time.NewTicker(execInterval)
	defer ticker.Stop()
	// now enter into "repeatedly at regular intervals"
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			customFunc()
		}
	}
}
