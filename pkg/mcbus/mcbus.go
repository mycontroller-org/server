package mcbus

import (
	"context"
	"strings"

	"github.com/mustafaturan/bus"
	srv "github.com/mycontroller-org/mycontroller-v2/pkg/service"
	"go.uber.org/zap"
)

var ctx = context.TODO()

// Publish a data to a topic
func Publish(topicName string, data interface{}) (*bus.Event, error) {
	ev, err := srv.BUS.Emit(ctx, topicName, data)
	if strings.Contains(err.Error(), "not found") {
		zap.L().Info("Topic not registered. Registering now", zap.String("topic", topicName))
		// register a topic
		srv.BUS.RegisterTopics(topicName)
		return srv.BUS.Emit(ctx, topicName, data)
	}
	return ev, err
}

// Subscribe a topic
func Subscribe(key string, handler *bus.Handler) {
	srv.BUS.RegisterHandler(key, handler)
}

// Unsubscribe a topic
func Unsubscribe(key string) {
	srv.BUS.DeregisterHandler(key)
}
