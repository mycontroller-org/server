package disable

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/spf13/cobra"
)

func init() {
	disableCmd.AddCommand(gatewayDisableCmd)
	disableCmd.AddCommand(virtualDeviceDisableCmd)
	disableCmd.AddCommand(virtualAssistantDisableCmd)
	disableCmd.AddCommand(taskDisableCmd)
	disableCmd.AddCommand(scheduleDisableCmd)
	disableCmd.AddCommand(handlerDisableCmd)
}

var gatewayDisableCmd = &cobra.Command{
	Use:     "gateway",
	Aliases: []string{"gw", "gateways"},
	Short:   "Disables the given gateways",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DisableGateway(args...)
		printStatus(err)
	},
}

var virtualDeviceDisableCmd = &cobra.Command{
	Use:     "virtual-device",
	Aliases: []string{"virtual-devices", "vd"},
	Short:   "Disables the given virtual devices",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DisableVirtualDevice(args...)
		printStatus(err)
	},
}

var virtualAssistantDisableCmd = &cobra.Command{
	Use:     "virtual-assistant",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Disables the given virtual assistants",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DisableVirtualAssistant(args...)
		printStatus(err)
	},
}

var taskDisableCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Disables the given tasks",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DisableTask(args...)
		printStatus(err)
	},
}

var scheduleDisableCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"schedules"},
	Short:   "Disables the given schedules",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DisableSchedule(args...)
		printStatus(err)
	},
}

var handlerDisableCmd = &cobra.Command{
	Use:     "handler",
	Aliases: []string{"handlers"},
	Short:   "Disables the given handlers",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DisableHandler(args...)
		printStatus(err)
	},
}
