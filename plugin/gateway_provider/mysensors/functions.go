package mysensors

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	"go.uber.org/zap"
)

// This function is like route for globally defined features for the request like reboot, discover, etc.,
// And this should have addition request implementation defined in "internalValidRequests" map on constants.go file
func handleRequests(gwCfg *gwml.Config, fn string, msg *msgml.Message, msMsg *message) error {
	switch fn {

	case nml.FuncDiscover:
		msMsg.Type = typeInternalDiscoverRequest
		msMsg.Payload = payloadEmpty
		msMsg.NodeID = idBroadcast

	case nml.FuncFirmwareUpdate, "ST_FIRMWARE_CONFIG_REQUEST":
		pl, err := executeFirmwareConfigRequest(msg)
		if err != nil {
			return err
		}
		msMsg.Command = cmdStream
		msMsg.Type = typeStreamFirmwareConfigResponse
		msMsg.Payload = pl

	case nml.FuncHeartbeat:
		msMsg.Type = typeInternalHeartBeatRequest
		msMsg.Payload = payloadEmpty

	case nml.FuncReboot:
		msMsg.Type = typeInternalReboot
		msMsg.Payload = payloadEmpty

	case nml.FuncRefreshNodeInfo:
		msMsg.Type = typeInternalPresentation
		msMsg.Payload = payloadEmpty

	case nml.FuncReset: // yet to implement
		return fmt.Errorf("This function is not implemented: %s", fn)

	case "I_CONFIG":
		msMsg.Type = typeInternalConfigResponse
		isImperial := gwCfg.Labels.GetBool(LabelImperialSystem)
		if isImperial {
			msMsg.Payload = "I"
		} else {
			msMsg.Payload = "M"
		}

	case "I_ID_REQUEST":
		msMsg.Type = typeInternalIDResponse
		msMsg.Payload = getNodeID(gwCfg)
		if msMsg.Payload == "" {
			return errors.New("Failed to get node ID")
		}

	case "I_TIME":
		msMsg.Type = typeInternalTime
		msMsg.Payload = strconv.FormatInt(time.Now().Local().Unix(), 10)

	default:
		return fmt.Errorf("This function is not implemented: %s", fn)
	}
	return nil
}

// get node id
func getNodeID(gwCfg *gwml.Config) string {
	f := []pml.Filter{{Key: "gatewayID", Operator: "eq", Value: gwCfg.ID}}
	nodes, err := nodeAPI.List(f, pml.Pagination{})
	if err != nil {
		zap.L().Error("Failed to find list of nodes", zap.String("gateway", gwCfg.Name), zap.Error(err))
		return ""
	}
	ids := make([]int, 0)
	for _, n := range nodes {
		if n.Labels.Get(LabelNodeID) != "" {
			id := n.Labels.GetInt(LabelNodeID)
			ids = append(ids, id)
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
