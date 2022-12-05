package api

import (
	"net/http"

	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
)

func (c *Client) ExecuteNodeAction(action string, nodeIDs []string) error {
	_, err := c.executeJson(API_ACTION_NODE, http.MethodGet, nil, nil, nodeIDs, http.StatusOK)
	return err
}

func (c *Client) ExecuteAction(actions []webHandlerTY.ActionConfig) error {
	_, err := c.executeJson(API_ACTION, http.MethodPost, nil, nil, actions, http.StatusOK)
	return err
}
