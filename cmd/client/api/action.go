package api

import (
	"net/http"
)

func (c *Client) ActionNode(action string, nodeIDs []string) error {
	_, err := c.executeJson(API_NODE_ACTION, http.MethodGet, nil, nil, nodeIDs, http.StatusOK)
	return err
}
