package api

import (
	"fmt"
	"net/http"

	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
)

// ExecuteNodeAction sends an action to one or more nodes.
func (c *Client) ExecuteNodeAction(action string, ids []string, payload ...string) error {
	var pl string
	if len(payload) > 0 {
		pl = payload[0]
	}
	return c.executeIDAction(API_ACTION_NODE, action, pl, ids)
}

func (c *Client) ExecuteGatewayAction(action string, ids []string) error {
	return c.executeIDAction(API_ACTION_GATEWAY, action, "", ids)
}

func (c *Client) executeIDAction(api, action, payload string, ids []string) error {
	if action == "" {
		return fmt.Errorf("action is required")
	}
	if len(ids) == 0 {
		return fmt.Errorf("at least one id is required")
	}
	query := map[string]interface{}{
		"action": action,
		"id":     ids,
	}
	if payload != "" {
		query["payload"] = payload
	}
	_, err := c.executeJson(api, http.MethodGet, nil, query, nil, http.StatusOK)
	return err
}

func (c *Client) ExecuteAction(actions []webHandlerTY.ActionConfig) error {
	_, err := c.executeJson(API_ACTION, http.MethodPost, nil, nil, actions, http.StatusOK)
	return err
}
