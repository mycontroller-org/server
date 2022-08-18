package generic

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	jsUtils "github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	"go.uber.org/zap"
)

// Posts a message to endpoint
func (p *Provider) Post(msg *msgTY.Message) error {
	return p.Protocol.Post(msg)
}

// Process received messages
func (p *Provider) ConvertToMessages(rawMsg *msgTY.RawMessage) ([]*msgTY.Message, error) {
	// check the provider type
	messages := make([]*msgTY.Message, 0)
	// execute onReceive script
	if p.Config.Script.OnReceive != "" {
		msgs, err := p.executeScript(p.Config.Script.OnReceive, rawMsg, nil)
		if err != nil {
			return nil, err
		}
		messages = msgs
	} else {
		// convert the rawMessage data to []*msgTY.Message
		err := json.ToStruct(rawMsg.Data, &messages)
		if err != nil {
			zap.L().Error("error on converting raw message data to []*Messages", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
			return nil, err
		}
	}

	// update gateway Id and timestamp
	currentTime := time.Now()
	for index := range messages {
		msg := messages[index]
		msg.GatewayID = p.GatewayConfig.ID
		if msg.Timestamp.IsZero() {
			msg.Timestamp = currentTime
		}
	}

	// send it to gateway message listener
	return messages, nil
}

// execute script and report back the response
func (p *Provider) executeScript(script string, rawMessage *msgTY.RawMessage, variables cmap.CustomMap) ([]*msgTY.Message, error) {
	if variables == nil {
		variables = cmap.CustomMap{}
	}

	// convert payload to json
	jsonPayload := ""
	if strPayload, ok := rawMessage.Data.(string); ok {
		jsonPayload = strPayload
	} else {
		strPayload, err := json.MarshalToString(rawMessage.Data)
		if err != nil {
			zap.L().Error("unable to convert payload to json string", zap.Any("rawMessage", rawMessage), zap.Error(err))
			return nil, err
		}
		jsonPayload = strPayload
	}

	variables[ScriptKeyDataIn] = jsonPayload
	response, err := jsUtils.Execute(script, variables)
	if err != nil {
		return nil, err
	}
	mapResponse, ok := response.(map[string]interface{})
	if !ok {
		zap.L().Warn("script response is not a map[string]interface", zap.String("gatewayId", p.GatewayConfig.ID))
		return nil, nil
	}

	messagesRaw, ok := mapResponse[ScriptKeyDataOut]
	if !ok {
		return nil, nil
	}

	messages, err := toMessages(messagesRaw)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
