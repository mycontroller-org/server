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
	mutex = &sync.RWMutex{}
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

func addToStore(cmd *Command) {
	mutex.Lock()
	defer mutex.Unlock()
	store[cmd.id] = cmd
}

func removeFromStore(id string) {
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

func (c *Command) setRunning(status bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isRunning = status
}

// Start triggers the command
func (c *Command) Start() error {
	if c.isRunning {
		return errors.New("start of the command already triggered")
	}
	go c.startFn()
	return nil
}

// StartAndWait triggers the command and waits till it completes
func (c *Command) StartAndWait() error {
	if c.isRunning {
		return errors.New("start of the command already triggered")
	}
	c.startFn()
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
	for id, command := range store {
		err := command.Stop()
		if err != nil {
			zap.L().Error("Error on stopping a command", zap.String("id", id), zap.String("Name", command.Name), zap.String("cmd", command.Command), zap.Any("args", command.Args), zap.Error(err))
		}
		if command.ExitFn != nil {
			st := command.cmd.Status()
			command.ExitFn(ExitTypeStop, st)
		}
	}
}

func (c *Command) startFn() {
	// generate an id and assign
	id, err := uuid.NewUUID()
	if err != nil {
		zap.L().Error("failed to generate UUID", zap.Error(err))
		return
	}
	c.id = id.String()
	c.setRunning(true)
	addToStore(c)

	c.cmd = cmd.NewCmd(c.Command, c.Args...)
	if len(c.Env) > 0 {
		c.cmd.Env = c.Env
	}
	zap.L().Debug("env", zap.Any("env", c.cmd.Env))
	statusChan := c.cmd.Start()
	zap.L().Debug("command execution started", zap.String("command", c.Command), zap.Any("args", c.Args))

	statusUpdate := c.StatusUpdate
	if c.StatusUpdate.Seconds() < 1 {
		statusUpdate = 1 * time.Second
	}
	timeout := c.Timeout
	if c.Timeout.Seconds() < 1 {
		timeout = 1 * time.Second
	}
	doneCh := make(chan bool)
	defer close(doneCh)
	// update on exit
	onExitFn := func(exitType string, status cmd.Status) {
		// terminate status update goroutine
		if c.StatusFn != nil {
			doneCh <- true
		}
		if c.ExitFn != nil {
			c.ExitFn(exitType, status)
		}
		// update it the running status
		c.setRunning(false)
		// remove it from the store
		removeFromStore(c.id)
	}

	// status update function
	if c.StatusFn != nil {
		statusTicker := time.NewTicker(statusUpdate)
		defer statusTicker.Stop()

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
			zap.L().Error("error on timeout", zap.String("command", c.Command), zap.Any("args", c.Args))
		}
		st := c.cmd.Status()
		onExitFn(ExitTypeStop, st)

	case <-time.After(timeout): // timeout
		zap.L().Info("timeout hits", zap.String("command", c.Command), zap.Any("args", c.Args))
		err := c.cmd.Stop()
		if err != nil {
			zap.L().Error("error on timeout", zap.String("command", c.Command), zap.Any("args", c.Args))
		}
		st := c.cmd.Status()
		onExitFn(ExitTypeTimeout, st)

	case status := <-statusChan: // command execution completed
		zap.L().Debug("command execution completed", zap.String("command", c.Command), zap.Any("args", c.Args))
		onExitFn(ExitTypeNormal, status)
	}
}
