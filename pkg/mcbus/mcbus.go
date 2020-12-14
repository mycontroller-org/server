package mcbus

import (
	"context"
	"strings"
	"sync"

	"github.com/mustafaturan/bus"
	"github.com/mustafaturan/monoton"
	"github.com/mustafaturan/monoton/sequencer"
	"go.uber.org/zap"
)

// common vars
var (
	ctx       = context.TODO()
	localBus  *bus.Bus
	isRunning = false
	mutex     sync.RWMutex
)

// Start func
func Start() {
	mutex.Lock()
	defer mutex.Unlock()
	if isRunning {
		zap.L().Warn("Bus service already running")
		return
	}
	node := uint64(1)
	initialTime := uint64(1577865600000) // set 2020-01-01 PST as initial time
	m, err := monoton.New(sequencer.NewMillisecond(), node, initialTime)
	if err != nil {
		zap.L().Fatal("Error on creating bus", zap.Error(err))
	}
	// init an id generator
	var idGenerator bus.Next = (*m).Next
	// create a new bus instance
	b, err := bus.NewBus(idGenerator)
	if err != nil {
		zap.L().Fatal("Error on creating bus", zap.Error(err))
	}
	localBus = b
	isRunning = true
}

// Close func
func Close() {
	if localBus != nil {
		// deregister handlers
		for _, hk := range localBus.HandlerKeys() {
			localBus.DeregisterHandler(hk)
		}
		// deregister topics
		for _, t := range localBus.Topics() {
			localBus.DeregisterTopics(t)
		}
	}
}

// Publish a data to a topic
func Publish(topicName string, data interface{}) (*bus.Event, error) {
	ev, err := localBus.Emit(ctx, topicName, data)
	if err != nil && strings.Contains(err.Error(), "not found") {
		zap.L().Debug("Topic not registered. Registering now", zap.String("topic", topicName), zap.Any("data", data))
		// register a topic
		localBus.RegisterTopics(topicName)
		return localBus.Emit(ctx, topicName, data)
	}
	return ev, err
}

// Subscribe a topic
func Subscribe(key string, handler *bus.Handler) {
	localBus.RegisterHandler(key, handler)
}

// Unsubscribe a topic
func Unsubscribe(key string) {
	localBus.DeregisterHandler(key)
}
