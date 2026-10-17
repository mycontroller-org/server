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

Node actions: reboot, reset, firmware-update, heartbeat, refresh-node-info, set-sleep-duration:SENSOR_ID:PERCENT, set-sleep-duration:PERCENT
Gateway actions: discover-nodes

  myc action node <alias> reboot mysensor.1
  myc action node <alias> set-sleep-duration:206:32 mysensor.1
  myc action node <alias> set-sleep-duration:32 mysensor.1
  myc action gateway <alias> discover-nodes mysensor
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
	Example: `  myc action node <alias> reboot mysensor.1 mysensor.2
  myc action node <alias> reset mysensor.1
  myc action node <alias> firmware-update mysensor.1
  myc action node <alias> heartbeat mysensor.1
  myc action node <alias> refresh-node-info mysensor.1
  myc action node <alias> set-sleep-duration:206:32 mysensor.1
  myc action node <alias> set-sleep-duration:32 mysensor.1`,
	Args:          cobra.MinimumNArgs(3),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rest := rootCmd.TakeAlias(args)

		// Support "action" and "action:payload"
		raw := rest[0]
		actionStr := raw
		payload := ""
		if idx := strings.Index(raw, ":"); idx >= 0 {
			actionStr = raw[:idx]
			payload = raw[idx+1:]
		}

		action, err := normalizeNodeAction(actionStr)
		if err != nil {
			return err
		}

		// Payload is required for some actions
		if action == nodeTY.ActionSetSleepDuration && payload == "" {
			return fmt.Errorf("action %q requires a payload (e.g. set-sleep-duration:32 or set-sleep-duration:206:32)", actionStr)
		}

		ids, resolveErr := client.ResolveNodeIDs(rest[1:])
		if len(ids) > 0 {
			if payload != "" {
				if err := client.ExecuteNodeAction(action, ids, payload); err != nil {
					return err
				}
			} else {
				if err := client.ExecuteNodeAction(action, ids); err != nil {
					return err
				}
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
	Example:       `  myc action gateway <alias> discover-nodes mysensor gw2`,
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
	case "set-sleep-duration":
		return nodeTY.ActionSetSleepDuration, nil
	default:
		return "", fmt.Errorf("unsupported node action %q (supported: reboot, reset, firmware-update, heartbeat, refresh-node-info, set-sleep-duration)", action)
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
