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

  myc action node home reboot mysensor.1
  myc action gateway home discover-nodes mysensor
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

var nodeActionCmd = &cobra.Command{
	Use:     "node <alias> <action> <gateway.node> [<gateway.node>...]",
	Aliases: []string{"nodes"},
	Short:   "Send an action to one or more nodes",
	Example: `  myc action node home reboot mysensor.1 mysensor.2
  myc action node home reset mysensor.1
  myc action node home firmware-update mysensor.1
  myc action node home heartbeat mysensor.1
  myc action node home refresh-node-info mysensor.1`,
	Args:          cobra.MinimumNArgs(3),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rest := rootCmd.TakeAlias(args)
		action, err := normalizeNodeAction(rest[0])
		if err != nil {
			return err
		}
		ids, resolveErr := client.ResolveNodeIDs(rest[1:])
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
	Use:           "gateway <alias> <action> <id> [<id>...]",
	Aliases:       []string{"gw", "gateways"},
	Short:         "Send an action to one or more gateways",
	Example:       `  myc action gateway home discover-nodes mysensor gw2`,
	Args:          cobra.MinimumNArgs(3),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rest := rootCmd.TakeAlias(args)
		action, err := normalizeGatewayAction(rest[0])
		if err != nil {
			return err
		}
		ids, resolveErr := client.ResolveGatewayIDs(rest[1:], true)
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
