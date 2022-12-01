package reboot

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Cmd.AddCommand(rebootCmd)
}

var rebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboots the requested resources",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}
