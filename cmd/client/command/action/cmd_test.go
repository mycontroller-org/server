package action

import (
	"testing"

	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeNodeAction(t *testing.T) {
	cases := map[string]string{
		"reboot":            nodeTY.ActionReboot,
		"reset":             nodeTY.ActionReset,
		"firmware-update":   nodeTY.ActionFirmwareUpdate,
		"firmware_update":   nodeTY.ActionFirmwareUpdate,
		"heartbeat":         nodeTY.ActionHeartbeatRequest,
		"heartbeat-request": nodeTY.ActionHeartbeatRequest,
		"refresh":           nodeTY.ActionRefreshNodeInfo,
		"refresh-node-info": nodeTY.ActionRefreshNodeInfo,
		"refresh_node_info": nodeTY.ActionRefreshNodeInfo,
	}
	for in, want := range cases {
		got, err := normalizeNodeAction(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	_, err := normalizeNodeAction("sleep")
	require.Error(t, err)
}

func TestNormalizeGatewayAction(t *testing.T) {
	got, err := normalizeGatewayAction("discover")
	require.NoError(t, err)
	assert.Equal(t, gatewayTY.ActionDiscoverNodes, got)
	got, err = normalizeGatewayAction("discover_nodes")
	require.NoError(t, err)
	assert.Equal(t, gatewayTY.ActionDiscoverNodes, got)
	_, err = normalizeGatewayAction("reboot")
	require.Error(t, err)
}
