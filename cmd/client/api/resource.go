package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	httpUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
)

func (c *Client) SaveGateway(gateway *gwTY.Config) error {
	return c.saveResource(API_GATEWAY_LIST, gateway)
}

func (c *Client) SaveFirmware(firmware *firmwareTY.Firmware) error {
	return c.saveResource(API_FIRMWARE_LIST, firmware)
}

func (c *Client) SaveDataRepository(item *dataRepoTY.Config) error {
	return c.saveResource(API_DATA_REPOSITORY_LIST, item)
}

func (c *Client) UploadFirmware(id, filename string) error {
	client := httpUtils.New(c.Insecure, "10m")
	url := fmt.Sprintf("%s%s/%s", c.ServerAddress, API_FIRMWARE_UPLOAD, id)
	_, err := client.ExecuteMultipart(url, http.MethodPost, c.getHeaders(nil), "file", filename, http.StatusOK)
	return err
}

func (c *Client) SaveNode(node *nodeTY.Node) error {
	return c.saveResource(API_NODE_LIST, node)
}

func (c *Client) SaveSource(source *sourceTY.Source) error {
	return c.saveResource(API_SOURCE_LIST, source)
}

func (c *Client) SaveField(field *fieldTY.Field) error {
	return c.saveResource(API_FIELD_LIST, field)
}

func (c *Client) GetNode(id string) (*nodeTY.Node, error) {
	item := &nodeTY.Node{}
	found, err := c.getByID(API_NODE_LIST, id, item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) GetSource(id string) (*sourceTY.Source, error) {
	item := &sourceTY.Source{}
	found, err := c.getByID(API_SOURCE_LIST, id, item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) GetField(id string) (*fieldTY.Field, error) {
	item := &fieldTY.Field{}
	found, err := c.getByID(API_FIELD_LIST, id, item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) FindGateway(id string) (*gwTY.Config, error) {
	if id == "" {
		return nil, nil
	}
	item := &gwTY.Config{}
	found, err := c.findResource(API_GATEWAY_LIST, idFilters(id), item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) FindFirmware(id string) (*firmwareTY.Firmware, error) {
	if id == "" {
		return nil, nil
	}
	item := &firmwareTY.Firmware{}
	found, err := c.findResource(API_FIRMWARE_LIST, idFilters(id), item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) FindDataRepository(id string) (*dataRepoTY.Config, error) {
	if id == "" {
		return nil, nil
	}
	item := &dataRepoTY.Config{}
	found, err := c.findResource(API_DATA_REPOSITORY_LIST, idFilters(id), item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) FindNode(id, gatewayID, nodeID string) (*nodeTY.Node, error) {
	item := &nodeTY.Node{}
	found, err := c.findResource(API_NODE_LIST, idFilters(id), item)
	if err != nil {
		return nil, err
	}
	if found {
		return item, nil
	}
	if gatewayID == "" || nodeID == "" {
		return nil, nil
	}
	item = &nodeTY.Node{}
	found, err = c.findResource(API_NODE_LIST, []storageTY.Filter{
		equalFilter(types.KeyGatewayID, gatewayID),
		equalFilter(types.KeyNodeID, nodeID),
	}, item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) FindSource(id, gatewayID, nodeID, sourceID string) (*sourceTY.Source, error) {
	item := &sourceTY.Source{}
	found, err := c.findResource(API_SOURCE_LIST, idFilters(id), item)
	if err != nil {
		return nil, err
	}
	if found {
		return item, nil
	}
	if gatewayID == "" || nodeID == "" || sourceID == "" {
		return nil, nil
	}
	item = &sourceTY.Source{}
	found, err = c.findResource(API_SOURCE_LIST, []storageTY.Filter{
		equalFilter(types.KeyGatewayID, gatewayID),
		equalFilter(types.KeyNodeID, nodeID),
		equalFilter(types.KeySourceID, sourceID),
	}, item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) FindField(id, gatewayID, nodeID, sourceID, fieldID string) (*fieldTY.Field, error) {
	item := &fieldTY.Field{}
	found, err := c.findResource(API_FIELD_LIST, idFilters(id), item)
	if err != nil {
		return nil, err
	}
	if found {
		return item, nil
	}
	if gatewayID == "" || nodeID == "" || sourceID == "" || fieldID == "" {
		return nil, nil
	}
	item = &fieldTY.Field{}
	found, err = c.findResource(API_FIELD_LIST, []storageTY.Filter{
		equalFilter(types.KeyGatewayID, gatewayID),
		equalFilter(types.KeyNodeID, nodeID),
		equalFilter(types.KeySourceID, sourceID),
		equalFilter(types.KeyFieldID, fieldID),
	}, item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) saveResource(api string, body interface{}) error {
	_, err := c.executeJson(api, http.MethodPost, nil, nil, body, http.StatusOK)
	return err
}

func (c *Client) getByID(api, id string, dest interface{}) (bool, error) {
	if id == "" {
		return false, nil
	}
	res, err := c.executeJson(fmt.Sprintf("%s/%s", api, id), http.MethodGet, nil, nil, nil, 0)
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if res.StatusCode != http.StatusOK {
		// FindOne currently returns 500 when no document matches
		if res.StatusCode == http.StatusInternalServerError && strings.Contains(res.StringBody(), "no documents") {
			return false, nil
		}
		return false, fmt.Errorf("failed with status code. [statusCode: %v, body: %s]", res.StatusCode, res.StringBody())
	}
	if len(res.Body) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(res.Body, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) findResource(api string, filters []storageTY.Filter, dest interface{}) (bool, error) {
	items, err := c.findResources(api, filters, 1)
	if err != nil || len(items) == 0 {
		return false, err
	}
	if err := utils.MapToStruct(utils.TagNameJSON, items[0], dest); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) findResources(api string, filters []storageTY.Filter, limit uint64) ([]map[string]interface{}, error) {
	if len(filters) == 0 {
		return nil, nil
	}
	queryParams, err := listQueryParams(filters, limit)
	if err != nil {
		return nil, err
	}
	result, err := c.listResource(api, queryParams)
	if err != nil {
		return nil, err
	}
	return decodeItems(result)
}

func listQueryParams(filters []storageTY.Filter, limit uint64) (map[string]interface{}, error) {
	filtersBytes, err := json.Marshal(filters)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"limit":  limit,
		"offset": uint64(0),
		"filter": string(filtersBytes),
	}, nil
}

func decodeItems(result *storageTY.Result) ([]map[string]interface{}, error) {
	if result == nil || result.Data == nil {
		return nil, nil
	}
	raw, ok := result.Data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response type:%T", result.Data)
	}
	items := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		data, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid item type:%T", item)
		}
		items = append(items, data)
	}
	return items, nil
}

func idFilters(id string) []storageTY.Filter {
	if id == "" {
		return nil
	}
	return []storageTY.Filter{equalFilter(types.KeyID, id)}
}

func equalFilter(key, value string) storageTY.Filter {
	return storageTY.Filter{
		Key:      key,
		Value:    value,
		Operator: storageTY.OperatorEqual,
	}
}
