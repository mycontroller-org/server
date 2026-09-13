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
	Use:     "gateway <alias> <id> [<id>...]",
	Aliases: []string{"gw", "gateways"},
	Short:   "Disables the given gateways",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DisableGateway(ids...)
		printStatus(err)
	},
}

var virtualDeviceDisableCmd = &cobra.Command{
	Use:     "virtual-device <alias> <id> [<id>...]",
	Aliases: []string{"virtual-devices", "vd"},
	Short:   "Disables the given virtual devices",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DisableVirtualDevice(ids...)
		printStatus(err)
	},
}

var virtualAssistantDisableCmd = &cobra.Command{
	Use:     "virtual-assistant <alias> <id> [<id>...]",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Disables the given virtual assistants",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DisableVirtualAssistant(ids...)
		printStatus(err)
	},
}

var taskDisableCmd = &cobra.Command{
	Use:     "task <alias> <id> [<id>...]",
	Aliases: []string{"tasks"},
	Short:   "Disables the given tasks",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DisableTask(ids...)
		printStatus(err)
	},
}

var scheduleDisableCmd = &cobra.Command{
	Use:     "schedule <alias> <id> [<id>...]",
	Aliases: []string{"schedules"},
	Short:   "Disables the given schedules",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DisableSchedule(ids...)
		printStatus(err)
	},
}

var handlerDisableCmd = &cobra.Command{
	Use:     "handler <alias> <id> [<id>...]",
	Aliases: []string{"handlers"},
	Short:   "Disables the given handlers",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DisableHandler(ids...)
		printStatus(err)
	},
}
