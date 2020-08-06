package mcbus

import (
	"context"
	"strings"

	"github.com/mustafaturan/bus"
	svc "github.com/mycontroller-org/mycontroller-v2/pkg/service"
	"go.uber.org/zap"
)

var ctx = context.TODO()

// Publish a data to a topic
func Publish(topicName string, data interface{}) (*bus.Event, error) {
	ev, err := svc.BUS.Emit(ctx, topicName, data)
	if err != nil && strings.Contains(err.Error(), "not found") {
		zap.L().Debug("Topic not registered. Registering now", zap.String("topic", topicName), zap.Any("data", data))
		// register a topic
		svc.BUS.RegisterTopics(topicName)
		return svc.BUS.Emit(ctx, topicName, data)
	}
	return ev, err
}

// Subscribe a topic
func Subscribe(key string, handler *bus.Handler) {
	svc.BUS.RegisterHandler(key, handler)
}

// Unsubscribe a topic
func Unsubscribe(key string) {
	svc.BUS.DeregisterHandler(key)
}
