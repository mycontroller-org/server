package update

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Cmd.AddCommand(updateCmd)
	updateCmd.AddCommand(userUpdateCmd)
	updateCmd.AddCommand(serviceAccountUpdateCmd)
}

var updateCmd = &cobra.Command{
	Use:           "update",
	Short:         "Updates users and service accounts",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}
