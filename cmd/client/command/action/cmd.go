package action

import (
	"fmt"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Cmd.AddCommand(actionCmd)
	actionCmd.AddCommand(nodeActionCmd)
	actionCmd.AddCommand(gatewayActionCmd)
}

var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "Send an action to a node or gateway",
	Long: `Send a server action to a node or gateway.

Node actions: reboot, reset, firmware-update, heartbeat, refresh-node-info
Gateway actions: discover-nodes

Node ids are quick ids: gatewayId.nodeId (for example mysensor.1).
Gateway ids are the gateway id (for example mysensor).
Reload a gateway with myc reload gateway; there is no gateway restart action.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

var nodeActionCmd = &cobra.Command{
	Use:     "node <action> <gateway.node> [<gateway.node>...]",
	Aliases: []string{"nodes"},
	Short:   "Send an action to one or more nodes",
	Example: `  myc action node reboot mysensor.1 mysensor.2
  myc action node reset mysensor.1
  myc action node firmware-update mysensor.1
  myc action node heartbeat mysensor.1
  myc action node refresh-node-info mysensor.1`,
	Args:          cobra.MinimumNArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		action, err := normalizeNodeAction(args[0])
		if err != nil {
			return err
		}
		client := rootCmd.GetClient()
		ids, resolveErr := client.ResolveNodeIDs(args[1:])
		if len(ids) > 0 {
			if err := client.ExecuteNodeAction(action, ids); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "sent %s to %d node(s)\n", action, len(ids))
		}
		return resolveErr
	},
}

var gatewayActionCmd = &cobra.Command{
	Use:           "gateway <action> <id> [<id>...]",
	Aliases:       []string{"gw", "gateways"},
	Short:         "Send an action to one or more gateways",
	Example:       `  myc action gateway discover-nodes mysensor gw2`,
	Args:          cobra.MinimumNArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		action, err := normalizeGatewayAction(args[0])
		if err != nil {
			return err
		}
		client := rootCmd.GetClient()
		ids, resolveErr := client.ResolveGatewayIDs(args[1:], true)
		if len(ids) > 0 {
			if err := client.ExecuteGatewayAction(action, ids); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "sent %s to %d gateway(s)\n", action, len(ids))
		}
		return resolveErr
	},
}

func normalizeNodeAction(action string) (string, error) {
	switch strings.ToLower(strings.ReplaceAll(action, "_", "-")) {
	case "reboot":
		return nodeTY.ActionReboot, nil
	case "reset":
		return nodeTY.ActionReset, nil
	case "firmware-update":
		return nodeTY.ActionFirmwareUpdate, nil
	case "heartbeat", "heartbeat-request":
		return nodeTY.ActionHeartbeatRequest, nil
	case "refresh-node-info", "refresh":
		return nodeTY.ActionRefreshNodeInfo, nil
	default:
		return "", fmt.Errorf("unsupported node action %q (supported: reboot, reset, firmware-update, heartbeat, refresh-node-info)", action)
	}
}

func normalizeGatewayAction(action string) (string, error) {
	switch strings.ToLower(strings.ReplaceAll(action, "_", "-")) {
	case "discover-nodes", "discover":
		return gatewayTY.ActionDiscoverNodes, nil
	default:
		return "", fmt.Errorf("unsupported gateway action %q (supported: discover-nodes)", action)
	}
}
