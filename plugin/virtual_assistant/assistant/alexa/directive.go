package alexa

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	alexaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/alexa/types"
	"go.uber.org/zap"
)

func (a *Assistant) executiveDirective(directive alexaTY.DirectiveOrEvent) *alexaTY.Response {
	switch directive.Header.Namespace {
	case alexaTY.NamespacePowerController:
		return a.executeDirectivePowerController(directive.Endpoint.EndpointID, directive.Header.Name)

	case alexaTY.NamespaceBrightnessController:
		return a.executeDirectiveBrightnessController(directive.Endpoint.EndpointID, directive.Header.Name, directive.Payload)

	default:
		a.logger.Warn("namespace not implemented", zap.String("namespace", directive.Header.Namespace), zap.String("name", directive.Header.Name))
	}

	return a.getErrorResponse(directive.Endpoint.EndpointID, alexaTY.ErrorTypeInternalError, "this namespace not implemented")
}

// PowerController
func (a *Assistant) executeDirectivePowerController(endpointID, directive string) *alexaTY.Response {
	payload := false
	if directive == alexaTY.DirectiveTurnOn {
		payload = true
	} else if directive == alexaTY.DirectiveTurnOff {
		payload = false
	} else {
		return a.getErrorResponse(endpointID, alexaTY.ErrorTypeInvalidDirective, fmt.Sprintf("%s directive not supported for %s", directive, alexaTY.NamespacePowerController))
	}
	return a.executeResourceAction(endpointID, alexaTY.NamespacePowerController, directive, vdTY.DeviceTraitOnOff, payload)
}

// BrightnessController
func (a *Assistant) executeDirectiveBrightnessController(endpointID, directive string, payload cmap.CustomMap) *alexaTY.Response {
	if directive == alexaTY.DirectiveSetBrightness {
		if payload.Get(alexaTY.PropertyNameBrightness) != nil {
			payload := payload.GetInt64(alexaTY.PropertyNameBrightness)
			return a.executeResourceAction(endpointID, alexaTY.NamespaceBrightnessController, directive, vdTY.DeviceTraitBrightness, payload)
		}
		// else if payload.Get(alexaTY.PropertyNameBrightness) != nil {
		// 	// TODO: implement for brightnessDelta
		// }
	}
	return a.getErrorResponse(endpointID, alexaTY.ErrorTypeInvalidDirective, fmt.Sprintf("%s directive not supported for %s", directive, alexaTY.NamespaceBrightnessController))
}

func (a *Assistant) executeResourceAction(endpointID, namespace, name string, trait string, payload interface{}) *alexaTY.Response {

	vDevice, err := a.deviceAPI.GetByID(endpointID)
	if err != nil {
		a.logger.Error("error on getting virtual device", zap.String("endpointId", endpointID), zap.Error(err))
		return a.getErrorResponse(endpointID, alexaTY.ErrorTypeNoSuchEndpoint, "there is no virtual device with this id")
	}

	var resource vdTY.Resource
	found := false

	for _, vResource := range vDevice.Traits {
		if vResource.TraitType == trait {
			resource = vResource
			found = true
			break
		}
	}

	if !found {
		a.logger.Error("error on getting virtual device trait", zap.String("endpointId", endpointID), zap.String("deviceName", vDevice.Name), zap.String("trait", vdTY.DeviceTraitOnOff))
		return a.getErrorResponse(endpointID, alexaTY.ErrorTypeNoSuchEndpoint, "trait not configured for this directive")
	}

	// post data to the actual resource
	quickId := fmt.Sprintf("%s:%s", resource.ResourceType, resource.QuickID)
	err = a.deviceAPI.PostActionOnResourceByQuickID(resource.ResourceType, quickId, payload)

	if err != nil {
		a.logger.Error("error on executing an action", zap.String("deviceId", vDevice.ID), zap.String("deviceName", vDevice.Name), zap.Error(err))
		return a.getErrorResponse(vDevice.ID, alexaTY.ErrorTypeInternalError, "error on executing an action")
	}

	// wait few seconds to get actual status
	// TODO: implement wait for response event and return immediately, if executed
	time.Sleep(time.Millisecond * 800) // wait for 800 milliseconds
	_value, _timestamp, err := a.deviceAPI.GetResourceState(vDevice, trait, &resource)
	if err != nil {
		return a.getErrorResponse(vDevice.ID, alexaTY.ErrorTypeInternalError, "error on getting resource state")
	}

	propertyName, found := alexaTY.InterfacePropertyNameMap[namespace]
	if !found {
		return a.getErrorResponse(vDevice.ID, alexaTY.ErrorTypeInternalError, fmt.Sprintf("error on getting property name for '%s'", namespace))
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
