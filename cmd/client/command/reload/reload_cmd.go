package reload

import (
	"fmt"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/spf13/cobra"
)

func init() {
	reloadCmd.AddCommand(gwReloadCmd)
	reloadCmd.AddCommand(virtualAssistantReloadCmd)
}

var gwReloadCmd = &cobra.Command{
	Use:     "gateway <id> [<id>...]",
	Aliases: []string{"gw", "gateways"},
	Short:   "Reload one or more gateways",
	Example: `  myc reload gateway mysensor gw2`,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		ids, resolveErr := client.ResolveGatewayIDs(args, false)
		if len(ids) > 0 {
			err := client.ReloadGateway(ids...)
			printStatus(err, len(ids), "gateway")
			if err != nil {
				return
			}
		}
		if resolveErr != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", resolveErr)
		}
	},
}

var virtualAssistantReloadCmd = &cobra.Command{
	Use:     "virtual-assistant <id> [<id>...]",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Reload one or more virtual assistants",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.GetClient().ReloadVirtualAssistant(args...)
		printStatus(err, len(args), "virtual assistant")
	},
}
