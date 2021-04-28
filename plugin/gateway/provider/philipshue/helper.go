package philipshue

import (
	"fmt"
	"time"

	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
)

func (p *Provider) getPresnMsg(nodeID, sourceID string) *msgML.Message {
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

func (p *Provider) getData(name string, value interface{}, metricType string) msgML.Data {
	data := msgML.NewData()
	data.Name = name
	data.Value = fmt.Sprintf("%v", value)
	data.MetricType = metricType
	return data
}

func (p *Provider) postMsg(msg *msgML.Message) error {
	topic := mcbus.GetTopicPostMessageToCore()
	return mcbus.Publish(topic, msg)
}
