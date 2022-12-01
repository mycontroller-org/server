package api

import (
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

func (c *Client) listResource(api string, queryParams map[string]interface{}) (*storageTY.Result, error) {
	res, err := c.executeJson(api, http.MethodGet, nil, queryParams, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}

	result := &storageTY.Result{}
	err = json.Unmarshal(res.Body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ListGateway(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_GATEWAY_LIST, queryParams)
}

func (c *Client) ListField(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_FIELD_LIST, queryParams)
}

func (c *Client) ListNode(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_NODE_LIST, queryParams)
}

func (c *Client) ListSource(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_SOURCE_LIST, queryParams)
}

func (c *Client) ListFirmware(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_FIRMWARE_LIST, queryParams)
}

func (c *Client) ListDataRepository(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_DATA_REPOSITORY_LIST, queryParams)
}

func (c *Client) ListVirtualDevice(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_VIRTUAL_DEVICE_LIST, queryParams)
}

func (c *Client) ListVirtualAssistant(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_VIRTUAL_ASSISTANT_LIST, queryParams)
}

func (c *Client) ListTask(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_TASK_LIST, queryParams)
}

func (c *Client) ListSchedule(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_SCHEDULE_LIST, queryParams)
}

func (c *Client) ListHandler(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_HANDLER_LIST, queryParams)
}

func (c *Client) ListForwardPayload(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_FORWARD_PAYLOAD_LIST, queryParams)
}

func (c *Client) ListBackup(queryParams map[string]interface{}) (*storageTY.Result, error) {
	return c.listResource(API_BACKUP_LIST, queryParams)
}
