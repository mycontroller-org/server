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
	messages := make([]*msgTY.Message, 0)

	err := json.ToStruct(data, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
