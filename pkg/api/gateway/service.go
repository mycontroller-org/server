package gateway

import (
	"context"
	"sync"

	"github.com/jaegertracing/jaeger/pkg/queue"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
)

// Provider instance
type Provider interface {
	ToRawMessage(message *msgml.Message) (*msgml.RawMessage, error)
	ToMessage(rawMesage *msgml.RawMessage) (*msgml.Message, error)
	Post(rawMessage *msgml.RawMessage) error
	Start(rxMessageFunc func(rawMsg *msgml.RawMessage) error) error
	Close() error
}

// Service details
type Service struct {
	Config                    *gwml.Config
	Provider                  Provider
	OutMsgQueue               *queue.BoundedQueue
	TopicMsg2Provider         string
	TopicSleepingMsg2Provider string
	SleepingNodeMsgQueue      map[string][]*msgml.Message
	mutex                     sync.RWMutex
	ctx                       context.Context
}
