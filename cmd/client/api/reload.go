package api

import "net/http"

func (c *Client) ReloadGateway(items ...string) error {
	_, err := c.executeJson(API_GATEWAY_RELOAD, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) ReloadVirtualAssistant(items ...string) error {
	_, err := c.executeJson(API_VIRTUAL_ASSISTANT_RELOAD, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}
