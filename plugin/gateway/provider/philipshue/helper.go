package philipshue

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
)

func (p *Provider) getPresentationMsg(nodeID, sourceID string) *msgML.Message {
	msg := msgML.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = nodeID
	msg.SourceID = sourceID
	msg.Type = msgML.TypePresentation
	msg.Timestamp = time.Now()
	return &msg
}

func (p *Provider) getMsg(nodeID, sourceID string) *msgML.Message {
	msg := msgML.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = nodeID
	msg.SourceID = sourceID
	msg.Type = msgML.TypeSet
	msg.Timestamp = time.Now()
	return &msg
}

func (p *Provider) getPayload(name string, value interface{}, metricType string, isReadOnly bool) msgML.Payload {
	data := msgML.NewPayload()
	data.Key = name
	data.Value = fmt.Sprintf("%v", value)
	data.MetricType = metricType
	if isReadOnly {
		data.Labels.Set(model.LabelReadOnly, "true")
	}
	return data
}

func (p *Provider) postMsg(msg *msgML.Message) error {
	topic := mcbus.GetTopicPostMessageToServer()
	return mcbus.Publish(topic, msg)
}
