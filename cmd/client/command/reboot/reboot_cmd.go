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
	Use:     "node",
	Aliases: []string{"nodes"},
	Short:   "Reboots the given nodes",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.ActionNode(nodeTY.ActionReboot, args)
		if err != nil {
			fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
			return
		}
		fmt.Fprintln(rootCmd.IOStreams.Out, "Nodes reboot command supplied")
	},
}
