package client

import (
	"strings"

	"github.com/mycontroller-org/server/v2/cmd/client/printer"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
)

const (
	ResourceGateway = "gateway"
	ResourceNode    = "node"
	ResourceSource  = "source"
	ResourceField   = "field"
)

func GetQuickIDValueFunc(resourceType string) printer.ValueFunc {
	var keys []string
	switch resourceType {
	case ResourceGateway:
		keys = []string{types.KeyID}

	case ResourceNode:
		keys = []string{types.KeyGatewayID, types.KeyNodeID}

	case ResourceSource:
		keys = []string{types.KeyGatewayID, types.KeyNodeID, types.KeySourceID}

	case ResourceField:
		keys = []string{types.KeyGatewayID, types.KeyNodeID, types.KeySourceID, types.KeyFieldID}

	}
	return func(data interface{}) string {
		qids := []string{}
		for _, key := range keys {
			strValue := ""
			_, value, err := filterUtils.GetValueByKeyPath(data, key)
			if err != nil {
				strValue = err.Error()
				if strings.HasPrefix(err.Error(), "key_not_found") {
					strValue = ""
				}
			} else {
				strValue = convertor.ToString(value)
			}
			qids = append(qids, strValue)
		}
		return strings.Join(qids, ".")
	}
}
