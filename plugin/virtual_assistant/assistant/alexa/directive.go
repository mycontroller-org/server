package alexa

import (
	"fmt"
	"time"

	actionAPI "github.com/mycontroller-org/server/v2/pkg/api/action"
	vdAPI "github.com/mycontroller-org/server/v2/pkg/api/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	converterUtil "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/types"
	alexaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/alexa/types"
	botAPI "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/device_api"
	"go.uber.org/zap"
)

func executiveDirective(directive alexaTY.DirectiveOrEvent) *alexaTY.Response {
	switch directive.Header.Namespace {
	case alexaTY.NamespacePowerController:
		return executeDirectivePowerController(directive.Endpoint.EndpointID, directive.Header.Name)

	case alexaTY.NamespaceBrightnessController:
		return executeDirectiveBrightnessController(directive.Endpoint.EndpointID, directive.Header.Name, directive.Payload)

	default:
		zap.L().Warn("namespace not implemented", zap.String("namespace", directive.Header.Namespace), zap.String("name", directive.Header.Name))
	}

	return getErrorResponse(directive.Endpoint.EndpointID, alexaTY.ErrorTypeInternalError, "this namespace not implemented")
}

// PowerController
func executeDirectivePowerController(endpointID, directive string) *alexaTY.Response {
	payload := false
	if directive == alexaTY.DirectiveTurnOn {
		payload = true
	} else if directive == alexaTY.DirectiveTurnOff {
		payload = false
	} else {
		return getErrorResponse(endpointID, alexaTY.ErrorTypeInvalidDirective, fmt.Sprintf("%s directive not supported for %s", directive, alexaTY.NamespacePowerController))
	}
	return executeResourceAction(endpointID, alexaTY.NamespacePowerController, directive, vdTY.DeviceTraitOnOff, payload)
}

// BrightnessController
func executeDirectiveBrightnessController(endpointID, directive string, payload cmap.CustomMap) *alexaTY.Response {
	if directive == alexaTY.DirectiveSetBrightness {
		if payload.Get(alexaTY.PropertyNameBrightness) != nil {
			payload := payload.GetInt64(alexaTY.PropertyNameBrightness)
			return executeResourceAction(endpointID, alexaTY.NamespaceBrightnessController, directive, vdTY.DeviceTraitBrightness, payload)
		}
		// else if payload.Get(alexaTY.PropertyNameBrightness) != nil {
		// 	// TODO: implement for brightnessDelta
		// }
	}
	return getErrorResponse(endpointID, alexaTY.ErrorTypeInvalidDirective, fmt.Sprintf("%s directive not supported for %s", directive, alexaTY.NamespaceBrightnessController))
}

func executeResourceAction(endpointID, namespace, name string, trait string, payload interface{}) *alexaTY.Response {

	vDevice, err := vdAPI.GetByID(endpointID)
	if err != nil {
		zap.L().Error("error on getting virtual device", zap.String("endpointId", endpointID), zap.Error(err))
		return getErrorResponse(endpointID, alexaTY.ErrorTypeNoSuchEndpoint, "there is no virtual device with this id")
	}

	resource, found := vDevice.Traits[trait]
	if !found {
		zap.L().Error("error on getting virtual device trait", zap.String("endpointId", endpointID), zap.String("deviceName", vDevice.Name), zap.String("trait", vdTY.DeviceTraitOnOff))
		return getErrorResponse(endpointID, alexaTY.ErrorTypeNoSuchEndpoint, "trait not configured for this directive")
	}

	if resource.Type != vdTY.ResourceByQuickID {
		zap.L().Error("label based trait not supported for OnOff", zap.String("deviceId", vDevice.ID), zap.String("deviceName", vDevice.Name))
		return getErrorResponse(endpointID, alexaTY.ErrorTypeNoSuchEndpoint, "label based trait not supported for OnOff")
	}

	err = actionAPI.ExecuteActionOnResourceByQuickID(&handlerType.ResourceData{
		ResourceType: resource.ResourceType,
		QuickID:      fmt.Sprintf("%s:%s", resource.ResourceType, resource.QuickID),
		Payload:      converterUtil.ToString(payload),
		PreDelay:     "0s",
	})
	if err != nil {
		zap.L().Error("error on executing an action", zap.String("deviceId", vDevice.ID), zap.String("deviceName", vDevice.Name), zap.Error(err))
		return getErrorResponse(vDevice.ID, alexaTY.ErrorTypeInternalError, "error on executing an action")
	}

	// wait few seconds to get actual status
	// TODO: implement wait for response event and return immediately, if executed
	time.Sleep(time.Millisecond * 800) // wait for 800 milliseconds
	_value, _timestamp, err := botAPI.GetResourceState(vDevice, trait, &resource)
	if err != nil {
		return getErrorResponse(vDevice.ID, alexaTY.ErrorTypeInternalError, "error on getting resource state")
	}

	propertyName, found := alexaTY.InterfacePropertyNameMap[namespace]
	if !found {
		return getErrorResponse(vDevice.ID, alexaTY.ErrorTypeInternalError, fmt.Sprintf("error on getting property name for '%s'", namespace))
	}

	return &alexaTY.Response{
		Event: alexaTY.DirectiveOrEvent{
			Header: alexaTY.Header{
				Namespace:      "Alexa",
				Name:           "Response",
				MessageID:      utils.RandUUID(),
				PayloadVersion: "3",
			},
			Endpoint: &alexaTY.DirectiveEndpoint{
				EndpointID: vDevice.ID,
			},
		},
		Context: &alexaTY.Context{
			Properties: []alexaTY.Property{
				{
					Namespace:    namespace,
					Name:         name,
					Value:        alexaTY.GetValue(propertyName, _value),
					TimeOfSample: _timestamp.Format(time.RFC3339),
				},
			},
		},
	}
}
