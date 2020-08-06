package command

import (
	"errors"
	"sync"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	store = make(map[string]*Command)
	mutex = sync.RWMutex{}
)

// command exit types
const (
	ExitTypeNormal  = "normal"
	ExitTypeTimeout = "timeout"
	ExitTypeStop    = "stop"
)

// Command deails
type Command struct {
	id           string
	Name         string
	Command      string
	Args         []string
	Env          []string
	Timeout      time.Duration
	StatusUpdate time.Duration
	StatusFn     func(cmd.Status)
	ExitFn       func(string, cmd.Status)
	isRunning    bool
	mutex        sync.RWMutex
	cmd          *cmd.Cmd
	stopCh       chan bool
}

func add(cmd *Command) {
	mutex.Lock()
	defer mutex.Unlock()
	store[cmd.id] = cmd
}

func remove(id string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(store, id)
}

// IsRunning returns the current status of the command
func (c *Command) IsRunning() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isRunning
}

// Start triggers the command
func (c *Command) Start() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.isRunning {
		return errors.New("start of the command already triggered")
	}
	// generate an id and assign
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	c.id = id.String()
	c.isRunning = true
	add(c)
	go c.startFn()
	return nil
}

// Stop terminates the command
func (c *Command) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.isRunning {
		return errors.New("This command is not started")
	}
	c.stopCh <- true
	return nil
}

// StopAll terminates all the running commands
func StopAll() {
	for id, c := range store {
		err := c.Stop()
		if err != nil {
			zap.L().Error("Error on stopping a command", zap.String("id", id), zap.String("Name", c.Name), zap.String("cmd", c.Command), zap.Any("args", c.Args), zap.Error(err))
		}
		if c.ExitFn != nil {
			st := c.cmd.Status()
			c.ExitFn(ExitTypeStop, st)
		}
	}
}

func (c *Command) startFn() {
	c.cmd = cmd.NewCmd(c.Command, c.Args...)
	if len(c.Env) > 0 {
		c.cmd.Env = c.Env
	}
	zap.L().Debug("env", zap.Any("env", c.cmd.Env))
	statusChan := c.cmd.Start()
	zap.L().Debug("command execution started", zap.String("command", c.Command), zap.Any("args", c.Args))

	statusUpdate := c.StatusUpdate
	if c.StatusUpdate.Seconds() < 2 {
		statusUpdate = 5 * time.Second
	}
	timeout := c.Timeout
	if c.Timeout.Seconds() < 2 {
		timeout = 10 * time.Second
	}
	statusTicker := time.NewTicker(statusUpdate)
	doneCh := make(chan bool)
	// update on exit
	onExitFn := func(exitType string, status cmd.Status) {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		// terminate status update goroutine
		if c.StatusFn != nil {
			doneCh <- true
		}
		statusTicker.Stop()
		if c.ExitFn != nil {
			c.ExitFn(exitType, status)
		}
		// update it the running status
		c.isRunning = false
		// remove it from the store
		remove(c.id)
	}

	// status update function
	if c.StatusFn != nil {
		go func(done <-chan bool) {
			for {
				select {
				case <-done:
					return
				case <-statusTicker.C:
					c.StatusFn(c.cmd.Status())
				}
			}
		}(doneCh)
	}

	select {
	case <-c.stopCh: // stop triggered
		zap.L().Debug("stop triggered", zap.String("command", c.Command), zap.Any("args", c.Args))
		err := c.cmd.Stop()
		if err != nil {
			zap.L().Error("Error on timout", zap.String("command", c.Command), zap.Any("args", c.Args))
		}
		st := c.cmd.Status()
		onExitFn(ExitTypeStop, st)

	case <-time.After(timeout): // timeout
		zap.L().Info("timout", zap.String("command", c.Command), zap.Any("args", c.Args))
		err := c.cmd.Stop()
		if err != nil {
			zap.L().Error("Error on timout", zap.String("command", c.Command), zap.Any("args", c.Args))
		}
		st := c.cmd.Status()
		onExitFn(ExitTypeTimeout, st)

	case status := <-statusChan: // command execution completed
		zap.L().Debug("command execution completed", zap.String("command", c.Command), zap.Any("args", c.Args))
		onExitFn(ExitTypeNormal, status)
	}
}
