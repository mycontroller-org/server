package resourceaction

import (
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	quickid "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	rsUtils "github.com/mycontroller-org/backend/v2/pkg/utils/resource_service"
	tplUtils "github.com/mycontroller-org/backend/v2/pkg/utils/template"
	"go.uber.org/zap"
)

// variables should have this prefix
var prefix = []string{"resource", "rs"}

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
func (c *Client) Post(variables map[string]interface{}) error {
	for _, value := range variables {
		quickID := getQuickID(value)
		if quickID != "" {
			sendPayload(quickID, variables)
		}
	}
	return nil
}

func getQuickID(value interface{}) string {
	stringValue, ok := value.(string)
	if !ok {
		return ""
	}
	keys := strings.SplitN(stringValue, ":", 2)
	key := strings.ToLower(keys[0])
	if len(keys) == 2 && utils.ContainsString(prefix, key) {
		return keys[1]
	}
	return ""
}

func sendPayload(stringValue string, variables map[string]interface{}) {
	// update with template
	value, err := tplUtils.Execute(stringValue, variables)
	if err != nil {
		zap.L().Error("failed to parse template", zap.Any("data", stringValue), zap.Error(err))
		return
	}
	// split quickID and payload
	data := strings.SplitN(value, "=", 2)
	if len(data) != 2 {
		return
	}
	id := data[0]
	payload := data[1]

	if !quickid.IsValidQuickID(id) {
		return
	}

	dataMap := map[string]string{
		model.KeyID:      id,
		model.KeyPayload: payload,
	}
	rsUtils.PostData(id, &dataMap, rsML.TypeQuickID, rsML.CommandSet)
}
