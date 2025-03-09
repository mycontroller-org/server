package alexa

import (
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/version"
	alexaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/alexa/types"
	"go.uber.org/zap"
)

const (
	DefaultDeviceLimit = int64(300) // limits to 300 devices
)

func (a *Assistant) executeDiscover(directive alexaTY.DirectiveOrEvent) (interface{}, error) {
	// get virtual devices
	vDevices, err := a.deviceAPI.ListDevices(nil, DefaultDeviceLimit, 0, a.cfg.DeviceFilter)
	if err != nil {
		return nil, err
	}

	endpoints := make([]alexaTY.Endpoint, 0)

	ver := version.Get()
	for _, vDevice := range vDevices {
		capabilities := make([]alexaTY.Capability, 0)
		for _, vResource := range vDevice.Traits {
			if aInterface, found := alexaTY.TraitControllerMap[vResource.TraitType]; found {
				properties := alexaTY.GetInterfaceProperties(aInterface)
				capabilities = append(capabilities, alexaTY.Capability{
					Type:       "AlexaInterface",
					Interface:  aInterface,
					Version:    "3",
					Properties: &properties,
				})

			} else {
				a.logger.Info("trait not found in the defined map", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("trait", vResource.TraitType))
			}
		}

		// add alex capability
		capabilities = append(capabilities, alexaTY.Capability{
			Type:      "AlexaInterface",
			Interface: "Alexa",
			Version:   "3",
		})

		endpoints = append(endpoints, alexaTY.Endpoint{
			EndpointID:        vDevice.ID,
			ManufacturerName:  "MyController",
			Description:       vDevice.Description,
			FriendlyName:      vDevice.Name,
			DisplayCategories: alexaTY.GetDisplayCategory(vDevice.DeviceType),
			AdditionalAttributes: &alexaTY.AdditionalAttributes{
				Manufacturer:    "MyController",
				SoftwareVersion: ver.Version,
			},
			Capabilities: capabilities,
			Cookie:       cmap.CustomStringMap{},
		})
	}

	response := alexaTY.Response{
		Event: alexaTY.DirectiveOrEvent{
			Header: alexaTY.Header{
				Namespace:      alexaTY.NamespaceDiscovery,
				Name:           alexaTY.NameDiscoverResponse,
				MessageID:      directive.Header.MessageID,
				PayloadVersion: "3",
			},
			Payload: map[string]interface{}{
				"endpoints": endpoints,
			},
		},
	}

	return &response, nil
}
