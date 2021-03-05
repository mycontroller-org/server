package concurrency

import (
	"errors"
	"fmt"
	"time"
)

// Runner struct
type Runner struct {
	stopper      *Channel
	customFunc   func()
	interval     time.Duration
	isOnetimeJob bool
	isRunning    *SafeBool
}

// Start triggers the execution
func (r *Runner) Start() error {
	if r.isRunning.IsSet() {
		return errors.New("this function should be called only once")
	}
	r.isRunning.Set()

	if r.interval <= 0 {
		return fmt.Errorf("interval should be greater than 0, interval:%v", r.interval)
	}

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	// now enter into "repeatedly at regular intervals"
	for {
		select {
		case <-r.stopper.CH:
			return nil
		case <-ticker.C:
			r.customFunc()
			if r.isOnetimeJob {
				return nil
			}
		}
	}
}

// Close func
func (r *Runner) Close() {
	r.stopper.SafeSend(true)
	r.stopper.SafeClose()
}

// GetAsyncRunner func
// executes the given func on the specified interval
func GetAsyncRunner(customFunc func(), execInterval time.Duration, isOnetimeJob bool) *Runner {
	return &Runner{
		stopper:      NewChannel(0),
		isOnetimeJob: isOnetimeJob,
		customFunc:   customFunc,
		interval:     execInterval,
		isRunning:    &SafeBool{},
	}
}
