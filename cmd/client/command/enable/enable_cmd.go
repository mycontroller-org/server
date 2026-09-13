package enable

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/spf13/cobra"
)

func init() {
	enableCmd.AddCommand(gatewayEnableCmd)
	enableCmd.AddCommand(virtualDeviceEnableCmd)
	enableCmd.AddCommand(virtualAssistantEnableCmd)
	enableCmd.AddCommand(taskEnableCmd)
	enableCmd.AddCommand(scheduleEnableCmd)
	enableCmd.AddCommand(handlerEnableCmd)
}

var gatewayEnableCmd = &cobra.Command{
	Use:     "gateway <alias> <id> [<id>...]",
	Aliases: []string{"gw", "gateways"},
	Short:   "Enables the given gateways",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.EnableGateway(ids...)
		printStatus(err)
	},
}

var virtualDeviceEnableCmd = &cobra.Command{
	Use:     "virtual-device <alias> <id> [<id>...]",
	Aliases: []string{"virtual-devices", "vd"},
	Short:   "Enables the given virtual devices",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.EnableVirtualDevice(ids...)
		printStatus(err)
	},
}

var virtualAssistantEnableCmd = &cobra.Command{
	Use:     "virtual-assistant <alias> <id> [<id>...]",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Enables the given virtual assistants",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.EnableVirtualAssistant(ids...)
		printStatus(err)
	},
}

var taskEnableCmd = &cobra.Command{
	Use:     "task <alias> <id> [<id>...]",
	Aliases: []string{"tasks"},
	Short:   "Enables the given tasks",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.EnableTask(ids...)
		printStatus(err)
	},
}

var scheduleEnableCmd = &cobra.Command{
	Use:     "schedule <alias> <id> [<id>...]",
	Aliases: []string{"schedules"},
	Short:   "Enables the given schedules",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.EnableSchedule(ids...)
		printStatus(err)
	},
}

var handlerEnableCmd = &cobra.Command{
	Use:     "handler <alias> <id> [<id>...]",
	Aliases: []string{"handlers"},
	Short:   "Enables the given handlers",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.EnableHandler(ids...)
		printStatus(err)
	},
}
