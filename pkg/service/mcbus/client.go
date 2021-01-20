package mcbus

import "github.com/mycontroller-org/backend/v2/pkg/model/event"

// Close func
func Close() error {
	if busClient != nil {
		return busClient.Close()
	}
	return nil
}

// Publish a data to a topic
func Publish(topic string, data interface{}) error {
	return busClient.Publish(topic, data)
}

// Subscribe a topic
func Subscribe(topic string, handler func(event *event.Event)) (int64, error) {
	return busClient.Subscribe(topic, handler)
}

// Unsubscribe a topic
func Unsubscribe(topic string, subscriptionID int64) error {
	return busClient.Unsubscribe(topic, subscriptionID)
}
