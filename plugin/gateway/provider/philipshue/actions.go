package philipshue

import (
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"go.uber.org/zap"
)

// action refresh node
func (p *Provider) actionRefreshNodeInfo(nodeID string) {
	if !strings.HasPrefix(nodeID, "state") {
		return
	}

	lightID, err := getID(nodeID)
	if err != nil {
		zap.L().Error("error on parsing light id", zap.Error(err))
		return
	}

	light, err := p.bridge.GetLight(lightID)
	if err != nil {
		zap.L().Error("error on fetching light info", zap.String("nodeId", nodeID), zap.Error(err))
		return
	}

	p.updateLight(light)
}

func (p *Provider) actionDiscover() {
	newLight, err := p.bridge.GetNewLights()
	if err != nil {
		zap.L().Error("error on getting new lights")
		return
	}

	for stringID := range newLight.Lights {
		lightID := int(convertor.ToInteger(stringID))
		light, err := p.bridge.GetLight(lightID)
		if err != nil {
			zap.L().Error("error on getting light details", zap.Int("id", lightID), zap.Error(err))
			continue
		}
		p.updateLight(light)
	}

	newSensors, err := p.bridge.GetNewSensors()
	if err != nil {
		zap.L().Error("error on getting new sensors")
		return
	}

	for stringID := range newSensors.Sensors {
		sensorID := int(convertor.ToInteger(stringID))
		sensor, err := p.bridge.GetSensor(sensorID)
		if err != nil {
			zap.L().Error("error on getting sensor details", zap.Int("id", sensorID), zap.Error(err))
			continue
		}
		p.updateSensor(sensor)
	}

}
