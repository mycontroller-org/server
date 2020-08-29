package mysensors

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	"go.uber.org/zap"
)

func handleInternalFunctions(gwCfg *gwml.Config, fn string, msMsg *message) error {
	switch fn {

	case nml.FuncReboot, "I_REBOOT":
		msMsg.Type = TypeInternalReboot
		msMsg.Payload = payloadEmpty

	case nml.FuncReset: // yet to implement
		return fmt.Errorf("This function not implemented: %s", fn)

	case nml.FuncDiscover, "I_DISCOVER_REQUEST":
		msMsg.Type = TypeInternalDiscoverRequest
		msMsg.Payload = payloadEmpty

	case nml.FuncRefreshNodeInfo, "I_PRESENTATION":
		msMsg.Type = TypeInternalPresentation
		msMsg.Payload = payloadEmpty

	case nml.FuncHeartbeat, "I_HEARTBEAT_REQUEST":
		msMsg.Type = TypeInternalHeartBeatRequest
		msMsg.Payload = payloadEmpty

	case "I_TIME":
		msMsg.Type = TypeInternalTime
		msMsg.Payload = strconv.FormatInt(time.Now().Local().Unix(), 10)

	case "I_ID_REQUEST":
		msMsg.Type = TypeInternalIDResponse
		msMsg.Payload = getNodeID(gwCfg)
		if msMsg.Payload == "" {
			return errors.New("Failed to get node ID")
		}

	case "I_CONFIG":
		msMsg.Type = TypeInternalConfigResponse
		isImperial := gwCfg.Labels.GetBool(KeyIsImperialSystem)
		if isImperial {
			msMsg.Payload = "I"
		} else {
			msMsg.Payload = "M"
		}

	default:
		return fmt.Errorf("This function not implemented: %s", fn)
	}
	return nil
}

// get node id
func getNodeID(gwCfg *gwml.Config) string {
	f := []pml.Filter{{Key: "gatewayID", Operator: "eq", Value: gwCfg.ID}}
	nodes, err := nodeAPI.ListNodes(f, pml.Pagination{})
	if err != nil {
		zap.L().Error("Failed to find list of nodes", zap.String("gateway", gwCfg.Name), zap.Error(err))
		return ""
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
		zap.L().Error("No space available on this network. Reached maximum node counts.", zap.String("gateway", gwCfg.Name))
		return ""
	}
	return string(electedID)
}
