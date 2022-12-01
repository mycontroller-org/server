package reload

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"

	"github.com/spf13/cobra"
)

func init() {
	reloadCmd.AddCommand(gwReloadCmd)
	reloadCmd.AddCommand(virtualAssistantReloadCmd)
}

var gwReloadCmd = &cobra.Command{
	Use:     "gateway",
	Aliases: []string{"gw", "gateways"},
	Short:   "Reloads the given gateways",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.ReloadGateway(args...)
		printStatus(err)
	},
}

var virtualAssistantReloadCmd = &cobra.Command{
	Use:     "virtual-assistant",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Reloads the given virtual assistants",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.ReloadVirtualAssistant(args...)
		printStatus(err)
	},
}
