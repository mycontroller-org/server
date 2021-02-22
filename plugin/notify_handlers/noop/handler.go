package noop

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
)

// Client struct
type Client struct {
	HandlerCfg *handlerML.Config
}

// Start handler implementation
func (c *Client) Start() error { return nil }

// Close handler implementation
func (c *Client) Close() error { return nil }

// Post handler implementation
func (c *Client) Post(variables map[string]interface{}) error { return nil }

// State implementation
func (c *Client) State() *model.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &model.State{}
		}
		return c.HandlerCfg.State
	}
	return &model.State{}
}
