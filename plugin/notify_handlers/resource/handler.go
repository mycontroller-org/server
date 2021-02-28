package resource

import (
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	variableUtils "github.com/mycontroller-org/backend/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

// Client struct
type Client struct {
	HandlerCfg *handlerML.Config
}

// Start handler implementation
func (c *Client) Start() error { return nil }

// Close handler implementation
func (c *Client) Close() error { return nil }

// State implementation
func (c *Client) State() *model.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &model.State{}
		}
		return c.HandlerCfg.State
	}
	return &model.State{}
}

// Post handler implementation
func (c *Client) Post(data map[string]interface{}) error {
	for name, value := range data {
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerML.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(genericData.Type, handlerML.DataTypeResource) {
			continue
		}

		rsData := handlerML.ResourceData{}
		err = variableUtils.UnmarshalBase64Yaml(genericData.Data, &rsData)
		if err != nil {
			zap.L().Error("error on loading resource data", zap.Error(err), zap.String("name", name), zap.String("input", stringValue))
			continue
		}

		zap.L().Debug("about to perform an action", zap.String("rawData", stringValue), zap.Any("finalData", rsData))
		busUtils.PostToResourceService("resource_selector", rsData, rsML.TypeResourceActionBySelector, rsML.CommandSet)
	}
	return nil
}
