package noop

import (
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"go.uber.org/zap"
)

// This is noop provider, means it will not do anything
// This is here just to avoid custom provider type error
// custom provider should be handled from externally

// Provider struct
type Provider struct {
	GatewayID string
}

// Start func
func (p *Provider) Start(messageReceiveFunc func(rawMsg *msgML.RawMessage) error) error {
	zap.L().Info("custom providers should be managed with proper lables, custom providers should be executed externally with custom code base", zap.String("gatewayId", p.GatewayID))
	return nil
}

// Close func
func (p *Provider) Close() error { return nil }

// Post func
func (p *Provider) Post(message *msgML.Message) error { return nil }

// ProcessReceived func
func (p *Provider) ProcessReceived(rawMesage *msgML.RawMessage) ([]*msgML.Message, error) {
	return nil, nil
}
