package mysensors

import (
	"strings"
	"time"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	"go.uber.org/zap"
)

func timeHandler(gw *gwml.Config, ms message) *message {
	ms.Type = "1" // I_TIME
	ms.Ack = "0"
	ms.Payload = string(time.Now().Local().Unix())
	return &ms
}

func idRequestHandler(gw *gwml.Config, ms message) *message {
	ms.Type = "4" // I_ID_RESPONSE
	ms.Ack = "0"
	f := []pml.Filter{{Key: "gatewayID", Operator: "eq", Value: gw.ID}}
	nodes, err := nodeAPI.ListNodes(f, pml.Pagination{})
	if err != nil {
		zap.L().Error("Failed to find list of nodes", zap.Error(err), zap.Any("message", ms))
		return nil
	}
	ids := make([]int, 0)
	for _, n := range nodes {
		if id, ok := n.Others[OthersKeyNodeID]; ok {
			ids = append(ids, id.(int))
		}
	}
	// find first available id
	electedID := 1
	for id := 1; id <= 255; id++ {
		found := false
		for _, rid := range ids {
			if rid == id {
				found = true
				break
			}
		}
		if !found {
			electedID = id
			break
		}
	}

	if electedID == 255 {
		zap.L().Error("No space available on this network. Reached maximum node counts.", zap.String("gateway", gw.Name))
		return nil
	}
	ms.Payload = string(electedID)
	return &ms
}

func configHandler(gw *gwml.Config, ms message) *message {
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
