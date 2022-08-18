package types

import (
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	convertorUtil "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
)

const (
	NamespaceDiscovery            = "Alexa.Discovery"
	NamespacePowerController      = "Alexa.PowerController"
	NamespaceBrightnessController = "Alexa.BrightnessController"
	NamespaceColorController      = "Alexa.ColorController"
	NamespacePercentageController = "Alexa.PercentageController"

	NameDiscoverResponse = "Discover.Response"

	PropertyNamePowerState      = "powerState"
	PropertyNameBrightness      = "brightness"
	PropertyNameBrightnessDelta = "brightnessDelta"
	PropertyNameColor           = "color"
	PropertyNamePercentage      = "percentage"

	DirectiveTurnOn        = "TurnOn"
	DirectiveTurnOff       = "TurnOff"
	DirectiveSetBrightness = "SetBrightness"
)

var (
	TraitControllerMap = map[string]string{
		vdTY.DeviceTraitOnOff:        NamespacePowerController,
		vdTY.DeviceTraitBrightness:   NamespaceBrightnessController,
		vdTY.DeviceTraitColorSetting: NamespaceColorController,
	}

	InterfacePropertyNameMap = map[string]string{
		NamespacePowerController:      PropertyNamePowerState,
		NamespaceBrightnessController: PropertyNameBrightness,
		NamespaceColorController:      PropertyNameColor,
	}
)

func GetValue(propertyName string, value interface{}) interface{} {
	switch propertyName {
	case PropertyNamePowerState:
		if convertorUtil.ToBool(value) {
			return "ON"
		}
		return "OFF"

	case PropertyNameBrightness:
		return convertorUtil.ToInteger(value)

	case PropertyNameColor:
		return map[string]float32{
			"hue":        0.0,
			"saturation": 0.0,
			"brightness": 0.0,
		}

	default:
		return convertorUtil.ToString(value)
	}
}

func GetInterfaceProperties(aInterface string) Properties {
	properties := Properties{
		Supported:           make([]cmap.CustomStringMap, 0),
		ProactivelyReported: true,
		Retrievable:         true,
	}
	switch aInterface {
	case NamespacePowerController:
		properties.Supported = []cmap.CustomStringMap{
			{"name": PropertyNamePowerState},
		}

	case NamespaceBrightnessController:
		properties.Supported = []cmap.CustomStringMap{
			{"name": PropertyNameBrightness},
		}

	case NamespaceColorController:
		properties.Supported = []cmap.CustomStringMap{
			{"name": PropertyNameColor},
		}

	case NamespacePercentageController:
		properties.Supported = []cmap.CustomStringMap{
			{"name": PropertyNamePercentage},
		}

	default:
		properties.ProactivelyReported = false
		properties.Retrievable = false
	}

	return properties
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery.html#display-categories
var (
	DisplayCategory = map[string]string{
		vdTY.DeviceTypeLight:          "LIGHT",
		vdTY.DeviceTypeSwitch:         "SWITCH",
		vdTY.DeviceTypeCamera:         "CAMERA",
		vdTY.DeviceTypeWindow:         "EXTERIOR_BLIND",
		vdTY.DeviceTypeDoor:           "DOOR",
		vdTY.DeviceTypeDoorBell:       "DOORBELL",
		vdTY.DeviceTypeFan:            "FAN",
		vdTY.DeviceTypeGarageDoor:     "GARAGE_DOOR",
		vdTY.DeviceTypeTelevision:     "TV",
		vdTY.DeviceTypeThermostat:     "THERMOSTAT",
		vdTY.DeviceTypeRouter:         "ROUTER",
		vdTY.DeviceTypeOutlet:         "SMARTPLUG",
		vdTY.DeviceTypeAirConditioner: "AIR_CONDITIONER",
		vdTY.DeviceTypeAirCooler:      "AIR_CONDITIONER",
	}
)

func GetDisplayCategory(deviceType string) []string {
	category, found := DisplayCategory[deviceType]
	if !found {
		category = "OTHERS"
	}
	return []string{category}
}
