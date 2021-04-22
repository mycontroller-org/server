package mysensors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	gwML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// This function is like route for globally defined features for the request like reboot, discover, etc.,
// And this should have addition request implementation defined in "internalValidRequests" map on constants.go file
func handleActions(gwCfg *gwML.Config, fn string, msg *msgML.Message, msMsg *message) error {
	switch fn {

	case nodeML.ActionDiscover:
		msMsg.Type = actionDiscoverRequest
		msMsg.Payload = payloadEmpty
		msMsg.NodeID = idBroadcast

	case nodeML.ActionHeartbeatRequest:
		msMsg.Type = actionHeartBeatRequest
		msMsg.Payload = payloadEmpty

	case nodeML.ActionReboot:
		msMsg.Type = actionReboot
		msMsg.Payload = payloadEmpty

	case nodeML.ActionRefreshNodeInfo:
		msMsg.Type = actionRequestPresentation
		msMsg.Payload = payloadEmpty

	case nodeML.ActionReset:
		// NOTE: This feature supports only for MySensorsBootloaderRF24
		// set reset flag on the node labels
		// reboot the node
		// on a node reboot, bootloader asks the firmware details
		// we pass erase EEPROM command.
		// erase EEPROM possible only via bootloader
		err := updateResetFlag(msg)
		if err != nil {
			return err
		}
		msMsg.Type = actionReboot
		msMsg.Payload = payloadEmpty

	case "I_CONFIG":
		msMsg.Type = actionConfig
		isImperial := gwCfg.Labels.GetBool(LabelImperialSystem)
		if isImperial {
			msMsg.Payload = "I"
		} else {
			msMsg.Payload = "M"
		}

	case "I_ID_REQUEST":
		msMsg.Type = actionIDResponse
		msMsg.Payload = getNodeID(gwCfg)
		if msMsg.Payload == "" {
			return errors.New("error on getting node ID")
		}

	case "I_TIME":
		msMsg.Payload = getTimestamp(gwCfg)
		msMsg.Type = actionTime

	case nodeML.ActionFirmwareUpdate, "ST_FIRMWARE_CONFIG_REQUEST":
		pl, err := executeFirmwareConfigRequest(msg)
		if err != nil {
			return err
		}
		msMsg.Command = cmdStream
		msMsg.Type = actionFirmwareConfigResponse
		msMsg.Payload = strings.ToUpper(pl)

	case "ST_FIRMWARE_REQUEST":
		pl, err := executeFirmwareRequest(msg)
		if err != nil {
			return err
		}
		msMsg.Command = cmdStream
		msMsg.Type = actionFirmwareResponse
		msMsg.Payload = strings.ToUpper(pl)

	default:
		return fmt.Errorf("this function is not implemented: %s", fn)
	}
	return nil
}

// geTimestamp returns timestamp in seconds from 1970
// adds zone offset to the actual timestamp
// user can specify different timezone as a gateway label
// if non set, take system timezone
func getTimestamp(gwCfg *gwML.Config) string {
	var loc *time.Location
	// get user defined timezone from gateway label
	tz := gwCfg.Labels.Get(model.LabelTimezone)
	if tz != "" {
		_loc, err := time.LoadLocation(tz)
		if err != nil {
			zap.L().Error("Failed to parse used defined timezone, fallback to system time zone", zap.String("userDefinedTimezone", tz))
			_loc = time.Now().Location()
		}
		loc = _loc
	}

	// set system location, if non set
	if loc == nil {
		loc = time.Now().Location()
	}

	// get zone offset and include it on the unix timestamp
	_, offset := time.Now().In(loc).Zone()
	timestamp := time.Now().Unix() + int64(offset)
	return strconv.FormatInt(timestamp, 10)
}

// get node id
func getNodeID(gwCfg *gwML.Config) string {
	filters := []stgML.Filter{{Key: "gatewayID", Operator: "eq", Value: gwCfg.ID}}
	response, err := nodeAPI.List(filters, nil)
	if err != nil {
		zap.L().Error("error on finding list of nodes", zap.String("gateway", gwCfg.ID), zap.Error(err))
		return ""
	}

	reservedIDs := make([]int, 0)
	if response.Data != nil {
		if nodes, ok := response.Data.([]nodeML.Node); ok {
			for _, n := range nodes {
				if n.Labels.Get(LabelNodeID) != "" {
					id := n.Labels.GetInt(LabelNodeID)
					reservedIDs = append(reservedIDs, id)
				}
			}
		}
	}

	// find first available id
	electedID := 1
	for id := 1; id <= 255; id++ {
		electedID = id
		found := false
		for _, rid := range reservedIDs {
			if rid == id {
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	if electedID == 255 {
		zap.L().Error("No space left on this network. Reached maximum node counts.", zap.String("gateway", gwCfg.ID))
		return ""
	}
	return strconv.Itoa(electedID)
}

func updateResetFlag(msg *msgML.Message) error {
	// get the node details
	node, err := nodeAPI.GetByGatewayAndNodeID(msg.GatewayID, msg.NodeID)
	if err != nil {
		return err
	}
	node.Labels.Set(LabelEraseEEPROM, "true")
	return nodeAPI.Save(node)
}
