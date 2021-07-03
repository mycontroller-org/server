package philipshue

import (
	"errors"
	"fmt"
	"strings"

	gwType "github.com/mycontroller-org/server/v2/plugin/gateway/type"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/server/v2/pkg/model/node"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"go.uber.org/zap"
)

// Post func
func (p *Provider) Post(msg *msgML.Message) error {
	if len(msg.Payloads) == 0 {
		return errors.New("there is no payload details on the message")
	}

	payload := msg.Payloads[0]

	if msg.Type == msgML.TypeAction {
		switch payload.Key {
		case nodeML.ActionRefreshNodeInfo:
			p.actionRefreshNodeInfo(msg.NodeID)

		case gwType.ActionDiscoverNodes:
			p.actionDiscover()
		}
	} else if msg.Type == msgML.TypeSet && strings.HasPrefix(msg.SourceID, "state") {
		p.updateState(msg.NodeID, &payload)
	}
	return nil
}

// ProcessReceived implementation
func (p *Provider) ProcessReceived(rawMsg *msgML.RawMessage) ([]*msgML.Message, error) {
	// not using the queue
	return nil, nil
}

func getID(nodeID string) (int, error) {
	idSlice := strings.Split(nodeID, "_")
	if len(idSlice) != 2 {
		return -1, fmt.Errorf("invalid node id. %s", nodeID)
	}

	intId := convertor.ToInteger(idSlice[1])
	return int(intId), nil
}

func (p *Provider) updateState(nodeID string, data *msgML.Payload) {
	lightID, err := getID(nodeID)
	if err != nil {
		zap.L().Error("error on parsing light id", zap.Error(err))
		return
	}

	light, err := p.bridge.GetLight(lightID)
	if err != nil {
		zap.L().Error("error on getting light", zap.String("nodeId", nodeID), zap.Error(err))
		return
	}

	switch data.Key {
	case FieldPower:
		powerON := convertor.ToBool(data.Value)
		if powerON {
			err = light.On()
		} else {
			err = light.Off()
		}

	case FieldBrightness:
		brightness := uint8(convertor.ToInteger(data.Value))
		err = light.Bri(brightness)

	case FieldHue:
		hue := uint16(convertor.ToInteger(data.Value))
		err = light.Hue(hue)

	case FieldSaturation:
		saturation := uint8(convertor.ToInteger(data.Value))
		err = light.Sat(saturation)

	case FieldColorTemperature:
		ct := uint16(convertor.ToInteger(data.Value))
		err = light.Ct(ct)

	case FieldAlert:
		alert := convertor.ToString(data.Value)
		err = light.Alert(alert)

	case FieldEffect:
		effect := convertor.ToString(data.Value)
		err = light.Effect(effect)

	default:
		zap.L().Error("unsupported field", zap.String("nodeId", nodeID), zap.String("fieldId", data.Key))
		return
	}
	if err != nil {
		zap.L().Error("error on updating field", zap.String("nodeId", nodeID), zap.String("fieldId", data.Key), zap.Any("value", data.Value))
		return
	}

	// get and update light new status
	light, err = p.bridge.GetLight(lightID)
	if err != nil {
		zap.L().Error("error on getting light", zap.String("nodeId", nodeID), zap.Error(err))
	}
	p.updateLight(light)

}
