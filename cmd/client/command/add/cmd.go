package add

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Cmd.AddCommand(addCmd)
	addCmd.AddCommand(serviceAccountAddCmd)
	addCmd.AddCommand(userAddCmd)
}

var addCmd = &cobra.Command{
	Use:           "add",
	Aliases:       []string{"create"},
	Short:         "Adds resources",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}
