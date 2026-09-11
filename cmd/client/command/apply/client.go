package apply

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/cmd/client/api"
	"github.com/mycontroller-org/server/v2/pkg/json"
)

type apiResourceClient struct {
	client *api.Client
}

func newAPIResourceClient(client *api.Client) ResourceClient {
	return &apiResourceClient{client: client}
}

func (c *apiResourceClient) FindGateway(id string) (string, error) {
	item, err := c.client.FindGateway(id)
	if err != nil || item == nil {
		return "", err
	}
	return item.ID, nil
}

func (c *apiResourceClient) FindNode(id, gatewayID, nodeID string) (string, error) {
	item, err := c.client.FindNode(id, gatewayID, nodeID)
	if err != nil || item == nil {
		return "", err
	}
	return item.ID, nil
}

func (c *apiResourceClient) FindSource(id, gatewayID, nodeID, sourceID string) (string, error) {
	item, err := c.client.FindSource(id, gatewayID, nodeID, sourceID)
	if err != nil || item == nil {
		return "", err
	}
	return item.ID, nil
}

func (c *apiResourceClient) FindField(id, gatewayID, nodeID, sourceID, fieldID string) (string, error) {
	item, err := c.client.FindField(id, gatewayID, nodeID, sourceID, fieldID)
	if err != nil || item == nil {
		return "", err
	}
	return item.ID, nil
}

func (c *apiResourceClient) FindFirmware(id string) (string, error) {
	item, err := c.client.FindFirmware(id)
	if err != nil || item == nil {
		return "", err
	}
	return item.ID, nil
}

func (c *apiResourceClient) FindDataRepository(id string) (string, error) {
	item, err := c.client.FindDataRepository(id)
	if err != nil || item == nil {
		return "", err
	}
	return item.ID, nil
}

func (c *apiResourceClient) SaveGateway(resource Resource) error {
	if resource.Gateway == nil {
		return nil
	}
	return c.client.SaveGateway(resource.Gateway)
}

func (c *apiResourceClient) SaveNode(resource Resource) error {
	if resource.Node == nil {
		return nil
	}
	return c.client.SaveNode(resource.Node)
}

func (c *apiResourceClient) SaveSource(resource Resource) error {
	if resource.Src == nil {
		return nil
	}
	return c.client.SaveSource(resource.Src)
}

func (c *apiResourceClient) SaveField(resource Resource) error {
	if resource.Field == nil {
		return nil
	}
	return c.client.SaveField(resource.Field)
}

func (c *apiResourceClient) SaveFirmware(resource Resource) error {
	if resource.Firmware == nil {
		return nil
	}
	return c.client.SaveFirmware(resource.Firmware)
}

func (c *apiResourceClient) SaveDataRepository(resource Resource) error {
	if resource.DataRepository == nil {
		return nil
	}
	return c.client.SaveDataRepository(resource.DataRepository)
}

func (c *apiResourceClient) DeleteGateway(ids ...string) error {
	return c.client.DeleteGateway(ids...)
}

func (c *apiResourceClient) DeleteNode(ids ...string) error {
	return c.client.DeleteNode(ids...)
}

func (c *apiResourceClient) DeleteSource(ids ...string) error {
	return c.client.DeleteSource(ids...)
}

func (c *apiResourceClient) DeleteField(ids ...string) error {
	return c.client.DeleteField(ids...)
}

func (c *apiResourceClient) DeleteFirmware(ids ...string) error {
	return c.client.DeleteFirmware(ids...)
}

func (c *apiResourceClient) DeleteDataRepository(ids ...string) error {
	return c.client.DeleteDataRepository(ids...)
}

func (c *apiResourceClient) GetExisting(resource Resource) ([]byte, error) {
	var item interface{}
	var err error
	switch resource.Kind {
	case KindGateway:
		item, err = c.client.FindGateway(resource.ID())
	case KindNode:
		gatewayID, nodeID, _, _ := resource.NaturalKeys()
		item, err = c.client.FindNode(resource.ID(), gatewayID, nodeID)
	case KindSource:
		gatewayID, nodeID, sourceID, _ := resource.NaturalKeys()
		item, err = c.client.FindSource(resource.ID(), gatewayID, nodeID, sourceID)
	case KindField:
		gatewayID, nodeID, sourceID, fieldID := resource.NaturalKeys()
		item, err = c.client.FindField(resource.ID(), gatewayID, nodeID, sourceID, fieldID)
	case KindFirmware:
		item, err = c.client.FindFirmware(resource.ID())
	case KindDataRepository:
		item, err = c.client.FindDataRepository(resource.ID())
	default:
		return nil, fmt.Errorf("unsupported kind %q", resource.Kind)
	}
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("resource is not present")
	}
	return json.Marshal(item)
}
