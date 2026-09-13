package api

import (
	"fmt"
	"net/http"
)

func (c *Client) DeleteGateway(items ...string) error {
	_, err := c.executeJson(API_GATEWAY_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteNode(nodes ...string) error {
	_, err := c.executeJson(API_NODE_DELETE, http.MethodDelete, nil, nil, nodes, http.StatusOK)
	return err
}

func (c *Client) DeleteSource(nodes ...string) error {
	_, err := c.executeJson(API_SOURCE_DELETE, http.MethodDelete, nil, nil, nodes, http.StatusOK)
	return err
}

func (c *Client) DeleteField(nodes ...string) error {
	_, err := c.executeJson(API_FIELD_DELETE, http.MethodDelete, nil, nil, nodes, http.StatusOK)
	return err
}

func (c *Client) DeleteFirmware(nodes ...string) error {
	_, err := c.executeJson(API_FIRMWARE_DELETE, http.MethodDelete, nil, nil, nodes, http.StatusOK)
	return err
}

func (c *Client) DeleteDataRepository(nodes ...string) error {
	_, err := c.executeJson(API_DATA_REPOSITORY_DELETE, http.MethodDelete, nil, nil, nodes, http.StatusOK)
	return err
}

func (c *Client) DeleteVirtualDevice(items ...string) error {
	_, err := c.executeJson(API_VIRTUAL_DEVICE_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteVirtualAssistant(items ...string) error {
	_, err := c.executeJson(API_VIRTUAL_ASSISTANT_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteTask(items ...string) error {
	_, err := c.executeJson(API_TASK_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteSchedule(items ...string) error {
	_, err := c.executeJson(API_SCHEDULE_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteHandler(items ...string) error {
	_, err := c.executeJson(API_HANDLER_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteForwardPayload(items ...string) error {
	_, err := c.executeJson(API_FORWARD_PAYLOAD_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteBackup(items ...string) error {
	_, err := c.executeJson(API_BACKUP_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteUser(items ...string) error {
	profile, err := c.GetProfile()
	if err != nil {
		return err
	}
	for _, id := range items {
		if id == profile.ID {
			return fmt.Errorf("cannot delete the current user %s", profile.Username)
		}
	}
	_, err = c.executeJson(API_USER_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeletePolicy(items ...string) error {
	_, err := c.executeJson(API_POLICY_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}

func (c *Client) DeleteServiceAccount(items ...string) error {
	_, err := c.executeJson(API_SERVICE_ACCOUNT_DELETE, http.MethodDelete, nil, nil, items, http.StatusOK)
	return err
}
