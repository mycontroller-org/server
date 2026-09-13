package reboot

import (
	"fmt"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/spf13/cobra"
)

func init() {
	rebootCmd.AddCommand(nodeRebootCmd)
}

var nodeRebootCmd = &cobra.Command{
	Use:     "node <alias> <gateway.node> [<gateway.node>...]",
	Aliases: []string{"nodes"},
	Short:   "Reboot one or more nodes",
	Example: `  myc reboot node <alias> mysensor.1 mysensor.2`,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args:          cobra.MinimumNArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rest := rootCmd.TakeAlias(args)
		ids, resolveErr := client.ResolveNodeIDs(rest)
		if len(ids) > 0 {
			if err := client.ExecuteNodeAction(nodeTY.ActionReboot, ids); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "sent reboot to %d node(s)\n", len(ids))
		}
		return resolveErr
	},
}
