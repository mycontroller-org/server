package alexa

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	alexaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/alexa/types"
	"go.uber.org/zap"
)

func (a *Assistant) reportState(directive alexaTY.DirectiveOrEvent) (interface{}, error) {
	endpointID := directive.Endpoint.EndpointID
	// get devices
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: endpointID}}
	vDevices, err := a.deviceAPI.ListDevices(filters, 1, 0, a.cfg.DeviceFilter) // TODO: add an api to get a device
	if err != nil {
		return nil, err
	}

	// update resource state
	err = a.deviceAPI.UpdateDeviceState(vDevices)
	if err != nil {
		return nil, err
	}

	if len(vDevices) == 0 {
		return nil, fmt.Errorf("unable to find requested device. id:%s", endpointID)
	}

	vDevice := vDevices[0]

	properties := make([]alexaTY.Property, 0)

	for _, vResource := range vDevice.Traits {
		// get interface
		aInterface, found := alexaTY.TraitControllerMap[vResource.TraitType]
		if !found {
			a.logger.Warn("trait not implemented", zap.String("deviceId", vDevice.ID), zap.String("deviceName", vDevice.Name), zap.String("trait", vResource.TraitType))
			continue
		}
		// get property name
		propertyName, found := alexaTY.InterfacePropertyNameMap[aInterface]
		if !found {
			a.logger.Warn("interface name not implemented", zap.String("deviceId", vDevice.ID), zap.String("deviceName", vDevice.Name), zap.String("trait", vResource.TraitType), zap.String("interface", aInterface))
			continue
		}

		properties = append(properties, alexaTY.Property{
			Namespace:    aInterface,
			Name:         propertyName,
			Value:        alexaTY.GetValue(propertyName, vResource.Value),
			TimeOfSample: vResource.ValueTimestamp.Format(time.RFC3339),
		})
	}

	response := alexaTY.Response{
		Event: alexaTY.DirectiveOrEvent{
			Header:   directive.Header,
			Endpoint: directive.Endpoint,
		},
		Context: &alexaTY.Context{Properties: properties},
	}

	// update header name
	response.Event.Header.Name = alexaTY.ResponseStateReport

	// update message id
	response.Event.Header.MessageID = utils.RandUUID()

	return &response, nil
}
