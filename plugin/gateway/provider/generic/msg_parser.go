package generic

import (
	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	jsUtils "github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	"go.uber.org/zap"
)

// Post func
func (p *Provider) Post(msg *msgTY.Message) error {
	if p.ProtocolType == ProtocolTypeHttpGeneric {
		return p.postHTTP(msg)
	}
	// TODO: add support for mqtt
	return nil
}

// Process received messages
func (p *Provider) ProcessReceived(rawMesage *msgTY.RawMessage) ([]*msgTY.Message, error) {
	// check the provider type

	messages := make([]*msgTY.Message, 0)
	// execute onReceive script
	if p.Config.Script.OnReceive != "" {
		msgs, err := executeScript(p.Config.Script.OnReceive, rawMesage, nil)
		if err != nil {
			return nil, err
		}
		messages = msgs
	}

	// send it to gateway message listener
	return messages, nil
}

// execute script and report back the response
func executeScript(script string, rawMesage *msgTY.RawMessage, variables cmap.CustomMap) ([]*msgTY.Message, error) {
	if variables == nil {
		variables = cmap.CustomMap{}
	}
	// convert payload to json
	jsonPayload := ""
	if strPayload, ok := rawMesage.Data.(string); ok {
		jsonPayload = strPayload
	} else {
		strPayload, err := json.MarshalToString(rawMesage.Data)
		if err != nil {
			zap.L().Error("unable to convert payload to json", zap.Any("rawMessage", rawMesage), zap.Error(err))
			return nil, err
		}
		jsonPayload = strPayload
	}

	variables[KeyReceivedMessages] = jsonPayload
	response, err := jsUtils.Execute(script, variables)
	if err != nil {
		return nil, err
	}
	mapResponse, ok := response.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	messagesRaw, ok := mapResponse[KeyReceivedMessages]
	if !ok {
		return nil, nil
	}

	messages, err := toMessages(messagesRaw)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
