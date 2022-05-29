package systemmonitoring

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
)

func (p *Provider) getData(name string, value interface{}, metricType, unit string, isReadOnly bool) msgTY.Payload {
	data := msgTY.NewPayload()
	data.Key = name
	data.SetValue(fmt.Sprintf("%v", value))
	data.MetricType = metricType
	if isReadOnly {
		data.Labels.Set(types.LabelReadOnly, "true")
	}
	return data
}

func (p *Provider) getMsg(sourceID string) msgTY.Message {
	msg := msgTY.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = p.NodeID
	msg.SourceID = sourceID
	msg.Type = msgTY.TypeSet
	msg.Timestamp = time.Now()

	return msg
}

func (p *Provider) getSourcePresentationMsg(sourceID, sourceName string) msgTY.Message {
	msg := msgTY.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = p.NodeID
	msg.SourceID = sourceID
	msg.Type = msgTY.TypePresentation
	msg.Timestamp = time.Now()

	if sourceName != "" {
		data := msgTY.NewPayload()
		data.Key = "name"
		data.SetValue(sourceName)
		msg.Payloads = append(msg.Payloads, data)
	}

	return msg
}

func (p *Provider) postMsg(msg *msgTY.Message) error {
	topic := mcbus.GetTopicPostMessageToProcessor()
	return mcbus.Publish(topic, msg)
}
