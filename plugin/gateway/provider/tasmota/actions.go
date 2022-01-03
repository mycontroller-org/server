package tasmota

import (
	"fmt"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/type"
)

// This function is like route for globally defined features for the request like reboot, discover, etc.,
// And this should have addition request implementation defined in "internalValidRequests" map on constants.go file
func handleActions(gwCfg *gwTY.Config, action string, msg *msgTY.Message, tmMsg *message) error {
	switch action {

	case gwTY.ActionDiscoverNodes:
		return fmt.Errorf("discover feature not implemented or not supported")

	case nodeTY.ActionHeartbeatRequest:
		tmMsg.Command = cmdStatus
		// 1 = show device parameters information
		tmMsg.Payload = "1"

	case nodeTY.ActionReboot:
		tmMsg.Command = cmdRestart
		// 1 = restart device with configuration saved to flash
		tmMsg.Payload = "1"

	case nodeTY.ActionRefreshNodeInfo:
		tmMsg.Command = cmdStatus
		// 0 = show all status information (1 - 11)
		tmMsg.Payload = "0"

	case nodeTY.ActionReset:
		tmMsg.Command = cmdReset
		// 6 = erase all flash and reset parameters to firmware defaults but keep Wi-Fi and MQTT settings and restart
		tmMsg.Payload = "6"

	//case nml.ActionFirmwareUpdate:

	default:
		return fmt.Errorf("this action is not implemented: %s", action)
	}
	return nil
}
