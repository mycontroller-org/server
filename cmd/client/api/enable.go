package api

import "net/http"

func (c *Client) EnableGateway(items ...string) error {
	_, err := c.executeJson(API_GATEWAY_ENABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) EnableVirtualDevice(items ...string) error {
	_, err := c.executeJson(API_VIRTUAL_DEVICE_ENABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) EnableVirtualAssistant(items ...string) error {
	_, err := c.executeJson(API_VIRTUAL_ASSISTANT_ENABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) EnableTask(items ...string) error {
	_, err := c.executeJson(API_TASK_ENABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) EnableSchedule(items ...string) error {
	_, err := c.executeJson(API_SCHEDULE_ENABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) EnableHandler(items ...string) error {
	_, err := c.executeJson(API_HANDLER_ENABLE, http.MethodPost, nil, nil, items, http.StatusOK)
	return err
}
