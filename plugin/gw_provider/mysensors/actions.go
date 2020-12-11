package mysensors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/model/node"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	"go.uber.org/zap"
)

// This function is like route for globally defined features for the request like reboot, discover, etc.,
// And this should have addition request implementation defined in "internalValidRequests" map on constants.go file
func handleActions(gwCfg *gwml.Config, fn string, msg *msgml.Message, msMsg *message) error {
	switch fn {

	case nml.ActionDiscover:
		msMsg.Type = typeInternalDiscoverRequest
		msMsg.Payload = payloadEmpty
		msMsg.NodeID = idBroadcast

	case nml.ActionHeartbeatRequest:
		msMsg.Type = typeInternalHeartBeatRequest
		msMsg.Payload = payloadEmpty

	case nml.ActionReboot:
		msMsg.Type = typeInternalReboot
		msMsg.Payload = payloadEmpty

	case nml.ActionRefreshNodeInfo:
		msMsg.Type = typeInternalPresentation
		msMsg.Payload = payloadEmpty

	case nml.ActionReset:
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
		msMsg.Type = typeInternalReboot
		msMsg.Payload = payloadEmpty

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
		msMsg.Payload = getTimestamp(gwCfg)
		msMsg.Type = typeInternalTime

	case nml.ActionFirmwareUpdate, "ST_FIRMWARE_CONFIG_REQUEST":
		pl, err := executeFirmwareConfigRequest(msg)
		if err != nil {
			return err
		}
		msMsg.Command = cmdStream
		msMsg.Type = typeStreamFirmwareConfigResponse
		msMsg.Payload = strings.ToUpper(pl)

	case "ST_FIRMWARE_REQUEST":
		pl, err := executeFirmwareRequest(msg)
		if err != nil {
			return err
		}
		msMsg.Command = cmdStream
		msMsg.Type = typeStreamFirmwareResponse
		msMsg.Payload = strings.ToUpper(pl)

	default:
		return fmt.Errorf("This function is not implemented: %s", fn)
	}
	return nil
}

// geTimestamp returns timestamp in seconds from 1970
// adds zone offset to the actual timestamp
// user can specify different timezone as a gateway label
// if non set, take system timezone
func getTimestamp(gwCfg *gwml.Config) string {
	var loc *time.Location
	// get user defined timezone from gateway label
	tz := gwCfg.Labels.Get(ml.LabelTimezone)
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
func getNodeID(gwCfg *gwml.Config) string {
	f := []pml.Filter{{Key: "gatewayID", Operator: "eq", Value: gwCfg.ID}}
	response, err := nodeAPI.List(f, nil)
	if err != nil {
		zap.L().Error("Failed to find list of nodes", zap.String("gateway", gwCfg.Name), zap.Error(err))
		return ""
	}

	reservedIDs := make([]int, 0)
	if response.Data != nil {
		if nodes, ok := response.Data.([]node.Node); ok {
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
		zap.L().Error("No space left on this network. Reached maximum node counts.", zap.String("gateway", gwCfg.Name))
		return ""
	}
	return strconv.Itoa(electedID)
}

func updateResetFlag(msg *msgml.Message) error {
	// get the node details
	node, err := nodeAPI.GetByIDs(msg.GatewayID, msg.NodeID)
	if err != nil {
		return err
	}
	node.Labels.Set(LabelEraseEEPROM, "true")
	return nodeAPI.Save(node)
}
