package generic

import (
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/json"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
)

// returns interface to []Message
func toMessages(data interface{}) ([]*msgTY.Message, error) {
	if data == nil {
		return nil, errors.New("empty messages")
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	messages := make([]msgTY.Message, 0)
	err = json.Unmarshal(bytes, &messages)
	if err != nil {
		return nil, err
	}

	finalMsgs := make([]*msgTY.Message, 0)
	for index := range messages {
		msg := messages[index]
		finalMsgs = append(finalMsgs, &msg)
	}

	return finalMsgs, nil
}

func toHttpNode(config interface{}) (*HttpNode, error) {
	return nil, nil
}
