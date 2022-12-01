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
	Use:     "gateway",
	Aliases: []string{"gw", "gateways"},
	Short:   "Enables the given gateways",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.EnableGateway(args...)
		printStatus(err)
	},
}

var virtualDeviceEnableCmd = &cobra.Command{
	Use:     "virtual-device",
	Aliases: []string{"virtual-devices", "vd"},
	Short:   "Enables the given virtual devices",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.EnableVirtualDevice(args...)
		printStatus(err)
	},
}

var virtualAssistantEnableCmd = &cobra.Command{
	Use:     "virtual-assistant",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Enables the given virtual assistants",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.EnableVirtualAssistant(args...)
		printStatus(err)
	},
}

var taskEnableCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Enables the given tasks",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.EnableTask(args...)
		printStatus(err)
	},
}

var scheduleEnableCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"schedules"},
	Short:   "Enables the given schedules",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.EnableSchedule(args...)
		printStatus(err)
	},
}

var handlerEnableCmd = &cobra.Command{
	Use:     "handler",
	Aliases: []string{"handlers"},
	Short:   "Enables the given handlers",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.EnableHandler(args...)
		printStatus(err)
	},
}
