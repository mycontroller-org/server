package alexa

import (
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/version"
	alexaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/alexa/types"
	botAPI "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/api"
	"go.uber.org/zap"
)

func executeDiscover(directive alexaTY.DirectiveOrEvent) (interface{}, error) {
	// get virtual devices
	vDevices, err := botAPI.ListDevices(nil, 300, 0) // limits to 300 devices
	if err != nil {
		return nil, err
	}

	endpoints := make([]alexaTY.Endpoint, 0)

	ver := version.Get()
	for _, vDevice := range vDevices {
		capabilities := make([]alexaTY.Capability, 0)
		for trait := range vDevice.Traits {
			if aInterface, found := alexaTY.TraitControllerMap[trait]; found {
				properties := alexaTY.GetInterfaceProperties(aInterface)
				capabilities = append(capabilities, alexaTY.Capability{
					Type:       "AlexaInterface",
					Interface:  aInterface,
					Version:    "3",
					Properties: &properties,
				})

			} else {
				zap.L().Info("trait not found in the defined map", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("trait", trait))
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
