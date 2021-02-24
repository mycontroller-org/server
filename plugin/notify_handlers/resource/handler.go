package resource

import (
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	variablesUtils "github.com/mycontroller-org/backend/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

// variables should have this prefixResource
var prefixResource = []string{"resource_", "rs_"}

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
		match := false
		modifiedName := strings.ToLower(name)
		for _, identity := range prefixResource {
			if strings.HasPrefix(modifiedName, identity) {
				match = true
				break
			}
		}
		if !match {
			continue
		}

		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		data, err := variablesUtils.GetResourceSelector(stringValue)
		if err != nil {
			return err
		}
		zap.L().Debug("about to perform an action", zap.String("rawData", stringValue), zap.Any("finalData", data))
		busUtils.PostToResourceService("resource_selector", data, rsML.TypeResourceActionBySelector, rsML.CommandSet)
	}
	return nil
}
