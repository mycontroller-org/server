package tasmota

import (
	"fmt"

	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
)

// This function is like route for globally defined features for the request like reboot, discover, etc.,
// And this should have addition request implementation defined in "internalValidRequests" map on constants.go file
func handleActions(gwCfg *gwml.Config, fn string, msg *msgml.Message, tmMsg *message) error {
	switch fn {

	case nml.ActionDiscover:
		return fmt.Errorf("Discover feature not implemented or not supported")

	case nml.ActionHeartbeatRequest:
		tmMsg.Command = cmdStatus
		// 1 = show device parameters information
		tmMsg.Payload = "1"

	case nml.ActionReboot:
		tmMsg.Command = cmdRestart
		// 1 = restart device with configuration saved to flash
		tmMsg.Payload = "1"

	case nml.ActionRefreshNodeInfo:
		tmMsg.Command = cmdStatus
		// 0 = show all status information (1 - 11)
		tmMsg.Payload = "0"

	case nml.ActionReset:
		tmMsg.Command = cmdReset
		// 6 = erase all flash and reset parameters to firmware defaults but keep Wi-Fi and MQTT settings and restart
		tmMsg.Payload = "6"

	//case nml.ActionFirmwareUpdate:

	default:
		return fmt.Errorf("This function is not implemented: %s", fn)
	}
	return nil
}
