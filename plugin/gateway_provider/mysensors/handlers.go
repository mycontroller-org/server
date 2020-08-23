package mysensors

import (
	"strings"
	"time"

	nodeAPI "github.com/mycontroller-org/backend/pkg/api/node"
	ml "github.com/mycontroller-org/backend/pkg/model"
	gwml "github.com/mycontroller-org/backend/pkg/model/gateway"
	"go.uber.org/zap"
)

func timeHandler(gw *gwml.Config, ms myMessage) *myMessage {
	ms.Type = "1" // I_TIME
	ms.Ack = "0"
	ms.Payload = string(time.Now().Local().Unix())
	return &ms
}

func idRequestHandler(gw *gwml.Config, ms myMessage) *myMessage {
	ms.Type = "4" // I_ID_RESPONSE
	ms.Ack = "0"
	f := []ml.Filter{{Key: "gatewayID", Operator: "eq", Value: gw.ID}}
	nodes, err := nodeAPI.ListNodes(f, ml.Pagination{})
	if err != nil {
		zap.L().Error("Failed to find list of nodes", zap.Error(err), zap.Any("message", ms))
		return nil
	}
	id := 1
	for _, n := range nodes {
		if n.ShortID == string(id) {
			id++
			continue
		}
	}
	if id == 255 {
		zap.L().Error("In this network reached maximum node counts.", zap.Any("message", ms))
		return nil
	}
	ms.Payload = string(id)
	return &ms
}

func configHandler(gw *gwml.Config, ms myMessage) *myMessage {
	ms.Type = "6" // I_CONFIG
	ms.Ack = "0"
	if uc, ok := gw.Provider.Config[KeyProviderUnitsConfig]; ok {
		if strings.ToLower(uc.(string)) == "metric" {
			ms.Payload = "M"
		} else {
			ms.Payload = "I"
		}
	} else {
		ms.Payload = "M"
	}
	return &ms
}
