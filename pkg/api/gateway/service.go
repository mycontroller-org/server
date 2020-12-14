package gateway

import (
	"context"
	"sync"

	"github.com/jaegertracing/jaeger/pkg/queue"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gw_provider"
)

// Service details
type Service struct {
	GatewayConfig                  *gwml.Config
	provider                       gwpd.Provider
	messageToProviderQueue         *queue.BoundedQueue
	topicMessageToProvider         string
	topicSleepingMessageToProvider string
	sleepingNodeMessageQueue       map[string][]*msgml.Message
	mutex                          sync.RWMutex
	ctx                            context.Context
}
