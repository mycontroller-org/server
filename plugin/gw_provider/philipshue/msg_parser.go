package philipshue

import (
	"fmt"
	"strings"

	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	gwptcl "github.com/mycontroller-org/backend/v2/plugin/gw_protocol"
	"github.com/mycontroller-org/backend/v2/plugin/gw_protocol/http"
)

// implement message parser

// ToRawMessage func implementation
func (p *Provider) ToRawMessage(msg *msgml.Message) (*msgml.RawMessage, error) {
	return nil, nil
}

// ToMessage implementation
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {
	messages := make([]*msgml.Message, 0)

	// get response config
	respRaw := rawMsg.Others.Get(gwptcl.KeyHTTPResponseConf)
	if respRaw == nil {
		return nil, fmt.Errorf("There is no http response found on the raw message")
	}
	responseCfg, ok := respRaw.(http.ResponseConfig)
	if !ok {
		return nil, fmt.Errorf("Failed to convert response to target format")
	}

	data := make(map[string]interface{})
	err := ut.ToStruct(rawMsg.Data, &data)
	if err != nil {
		return nil, err
	}

	entity, err := getPhEntity(responseCfg.Path)
	if err != nil {
		return nil, err
	}

	switch entity.apiType {
	case "lights":
		if entity.subType != "" {

		} else if entity.id != "" {

		} else {
		}
	case "sensors":
		if entity.subType != "" {

		} else if entity.id != "" {

		} else {
		}
	default:
		//not supported
	}

	return messages, nil
}

type phEntity struct {
	apiType string
	id      string
	subType string
}

func getPhEntity(path string) (*phEntity, error) {
	paths := strings.Split(path, "/")
	entity := &phEntity{}
	switch len(paths) {
	case 3:
		entity.subType = paths[2]
		fallthrough
	case 2:
		entity.id = paths[1]
		fallthrough
	case 1:
		entity.apiType = paths[0]
	default:
		return nil, fmt.Errorf("invalid or not supported path, %s", path)
	}
	return entity, nil
}
