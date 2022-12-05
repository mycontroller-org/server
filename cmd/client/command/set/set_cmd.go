package set

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/spf13/cobra"
)

func init() {
	setCmd.AddCommand(fieldGetCmd)
}

var fieldGetCmd = &cobra.Command{
	Use:     "field",
	Aliases: []string{"fields"},
	Short:   "Sets the value to the fields resource",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		payload := args[len(args)-1]
		executeSetCmd(QuickIDPrefixField, "", args[:len(args)-1], payload)
	},
}
