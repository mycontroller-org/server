package handler

import (
	"github.com/mycontroller-org/server/v2/pkg/types"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"
)

const PluginNOOP = "noop"

func init() {
	Register(PluginNOOP, NewNoopPlugin)
}

// NoopClient struct
type NoopClient struct {
	HandlerCfg *handlerType.Config
}

func NewNoopPlugin(config *handlerType.Config) (handlerType.Plugin, error) {
	return &NoopClient{}, nil
}

func (p *NoopClient) Name() string {
	return PluginNOOP
}

// Start handler implementation
func (c *NoopClient) Start() error { return nil }

// Close handler implementation
func (c *NoopClient) Close() error { return nil }

// Post handler implementation
func (c *NoopClient) Post(variables map[string]interface{}) error { return nil }

// State implementation
func (c *NoopClient) State() *types.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &types.State{}
		}
		return c.HandlerCfg.State
	}
	return &types.State{}
}
