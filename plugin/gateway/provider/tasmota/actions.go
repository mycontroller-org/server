package tasmota

import (
	"fmt"

	gwML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
)

// This function is like route for globally defined features for the request like reboot, discover, etc.,
// And this should have addition request implementation defined in "internalValidRequests" map on constants.go file
func handleActions(gwCfg *gwML.Config, action string, msg *msgML.Message, tmMsg *message) error {
	switch action {

	case gwML.ActionDiscoverNodes:
		return fmt.Errorf("discover feature not implemented or not supported")

	case nodeML.ActionHeartbeatRequest:
		tmMsg.Command = cmdStatus
		// 1 = show device parameters information
		tmMsg.Payload = "1"

	case nodeML.ActionReboot:
		tmMsg.Command = cmdRestart
		// 1 = restart device with configuration saved to flash
		tmMsg.Payload = "1"

	case nodeML.ActionRefreshNodeInfo:
		tmMsg.Command = cmdStatus
		// 0 = show all status information (1 - 11)
		tmMsg.Payload = "0"

	case nodeML.ActionReset:
		tmMsg.Command = cmdReset
		// 6 = erase all flash and reset parameters to firmware defaults but keep Wi-Fi and MQTT settings and restart
		tmMsg.Payload = "6"

	//case nml.ActionFirmwareUpdate:

	default:
		return fmt.Errorf("this action is not implemented: %s", action)
	}
	return nil
}
