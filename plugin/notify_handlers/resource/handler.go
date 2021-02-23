package resource

import (
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	quickIDUtils "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	"go.uber.org/zap"
)

// variables should have this prefixResource
var prefixResource = []string{"resource_", "rs_"}

const (
	keyResourceType     = "--type"
	keyTypeLables       = "--labels"
	keyTypeQID          = "--qid"
	keyResourcePayload  = "--payload"
	keyResourcePredelay = "--predelay"
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

		// example parameters:
		// --labels=true,--type=sf,--payload=increment|1|100|0,test=123,hello=hi
		// --quick_id=gw:mysensor,--payload=reload
		// --quick_id=sf:mysensor.1.21.V_STATUS,--payload=on,--predelay=1s
		data := getAsMap(stringValue)
		zap.L().Debug("about to perform as action", zap.String("rawData", stringValue), zap.Any("parsedMap", data))

		if _, ok := data[keyTypeLables]; ok {
			actionOnLabels(data)
		}

		if _, ok := data[keyTypeQID]; ok {
			actionOnQuickID(data)
		}
	}
	return nil
}

func actionOnLabels(labels map[string]string) {
	resourceType, ok := labels[keyResourceType]
	if !ok {
		return
	}
	payload, ok := labels[keyResourcePayload]
	if !ok {
		return
	}

	// delete unwanted labels
	delete(labels, keyTypeLables)
	delete(labels, keyResourceType)
	delete(labels, keyResourcePayload)

	data := rsML.ResourceLabels{ResourceType: resourceType, Labels: labels, Payload: payload}
	busUtils.PostToResourceService("label_based", &data, rsML.TypeResourceByLabels, rsML.CommandSet)
}

func actionOnQuickID(data map[string]string) {
	quickID, ok := data[keyTypeQID]
	if !ok {
		return
	}

	payload, ok := data[keyResourcePayload]
	if !ok {
		return
	}

	if !quickIDUtils.IsValidQuickID(quickID) {
		return
	}

	dataMap := map[string]string{
		model.KeyID:      quickID,
		model.KeyPayload: payload,
	}
	busUtils.PostToResourceService(quickID, &dataMap, rsML.TypeResourceByQuickID, rsML.CommandSet)
}

func getAsMap(labelString string) map[string]string {
	// replace ignored comma with different char
	newLabelString := strings.ReplaceAll(labelString, `\,`, "@#$@")
	labels := strings.Split(newLabelString, ",")

	data := make(map[string]string)
	for _, tmpLabel := range labels {
		label := strings.ReplaceAll(tmpLabel, "@#$@", ",")
		keyValue := strings.SplitN(label, "=", 2)
		key := strings.ToLower(strings.TrimSpace(keyValue[0]))
		value := ""
		if key == "" {
			continue
		}
		if len(keyValue) == 2 {
			value = strings.TrimSpace(keyValue[1])
		}
		data[key] = value
	}
	return data
}
