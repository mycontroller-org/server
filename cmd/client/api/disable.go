package api

import "net/http"

func (c *Client) DisableGateway(items ...string) error {
	_, err := c.executeJson(API_GATEWAY_DISABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DisableVirtualDevice(items ...string) error {
	_, err := c.executeJson(API_VIRTUAL_DEVICE_DISABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DisableVirtualAssistant(items ...string) error {
	_, err := c.executeJson(API_VIRTUAL_ASSISTANT_DISABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DisableTask(items ...string) error {
	_, err := c.executeJson(API_TASK_DISABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DisableSchedule(items ...string) error {
	_, err := c.executeJson(API_SCHEDULE_DISABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DisableHandler(items ...string) error {
	_, err := c.executeJson(API_HANDLER_DISABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}
