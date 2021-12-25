package mcbus

import (
	busML "github.com/mycontroller-org/server/v2/pkg/model/bus"
	"go.uber.org/zap"
)

// Close func
func Close() error {
	if busClient != nil {
		return busClient.Close()
	}
	return nil
}

// Publish a data to a topic
func Publish(topic string, data interface{}) error {
	if pauseSRV.IsSet() {
		return nil
	}

	return busClient.Publish(topic, data)
}

// Subscribe a topic
func Subscribe(topic string, handler func(data *busML.BusData)) (int64, error) {
	return busClient.Subscribe(topic, handler)
}

// Unsubscribe a topic
func Unsubscribe(topic string, subscriptionID int64) error {
	return busClient.Unsubscribe(topic, subscriptionID)
}

// QueueSubscribe a topic
func QueueSubscribe(topic, queueName string, handler func(data *busML.BusData)) (int64, error) {
	return busClient.QueueSubscribe(topic, queueName, handler)
}

// QueueUnsubscribe a topic
func QueueUnsubscribe(topic, queueName string, subscriptionID int64) error {
	return busClient.QueueUnsubscribe(topic, queueName, subscriptionID)
}

// Pause bus service
func Pause() {
	pauseSRV.Set()
	zap.L().Info("bus service paused")
}

// Resume bus service
func Resume() {
	pauseSRV.Reset()
	zap.L().Info("bus service resumed")
}
