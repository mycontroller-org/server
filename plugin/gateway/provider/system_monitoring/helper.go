package systemmonitoring

import (
	"fmt"
	"time"

	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
)

func (p *Provider) getData(name string, value interface{}, metricType string) msgML.Payload {
	data := msgML.NewPayload()
	data.Key = name
	data.Value = fmt.Sprintf("%v", value)
	data.MetricType = metricType
	return data
}

func (p *Provider) getMsg(sourceID string) msgML.Message {
	msg := msgML.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = p.NodeID
	msg.SourceID = sourceID
	msg.Type = msgML.TypeSet
	msg.Timestamp = time.Now()

	return msg
}

func (p *Provider) getSourcePresentationMsg(sourceID, sourceName string) msgML.Message {
	msg := msgML.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = p.NodeID
	msg.SourceID = sourceID
	msg.Type = msgML.TypePresentation
	msg.Timestamp = time.Now()

	if sourceName != "" {
		data := msgML.NewPayload()
		data.Key = "name"
		data.Value = sourceName
		msg.Payloads = append(msg.Payloads, data)
	}

	return msg
}

func (p *Provider) postMsg(msg *msgML.Message) error {
	topic := mcbus.GetTopicPostMessageToCore()
	return mcbus.Publish(topic, msg)
}
