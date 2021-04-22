package concurrency

import (
	"context"
	"fmt"
	"time"
)

// FuncCallWithTimeout calls a given func and exists if the operation not completed in time
func FuncCallWithTimeout(callBackFunc func() <-chan bool, timeoutFunc func(), timeoutDuration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	select {
	case <-callBackFunc():
		// completed in time
		return nil
	case <-ctx.Done():
		if timeoutFunc != nil {
			timeoutFunc()
		}
		return fmt.Errorf("reached timeout: %s", timeoutDuration.String())
	}
}
