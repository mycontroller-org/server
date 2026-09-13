package add

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Cmd.AddCommand(addCmd)
	addCmd.AddCommand(serviceAccountAddCmd)
}

var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"create"},
	Short:   "Adds resources",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}
