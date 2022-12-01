package delete

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd.AddCommand(gwDeleteCmd)
	deleteCmd.AddCommand(nodeDeleteCmd)
	deleteCmd.AddCommand(sourceDeleteCmd)
	deleteCmd.AddCommand(fieldDeleteCmd)
	deleteCmd.AddCommand(firmwareDeleteCmd)
	deleteCmd.AddCommand(dataRepositoryDeleteCmd)
	deleteCmd.AddCommand(virtualDeviceDeleteCmd)
	deleteCmd.AddCommand(virtualAssistantDeleteCmd)
	deleteCmd.AddCommand(taskDeleteCmd)
	deleteCmd.AddCommand(scheduleDeleteCmd)
	deleteCmd.AddCommand(handlerDeleteCmd)
	deleteCmd.AddCommand(forwardPayloadDeleteCmd)
	deleteCmd.AddCommand(backupDeleteCmd)
}

var gwDeleteCmd = &cobra.Command{
	Use:     "gateway",
	Aliases: []string{"gw", "gateways"},
	Short:   "Deletes the given gateways",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteGateway(args...)
		printStatus(err)
	},
}

var nodeDeleteCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Short:   "Deletes the given nodes",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteNode(args...)
		printStatus(err)
	},
}

var sourceDeleteCmd = &cobra.Command{
	Use:     "source",
	Aliases: []string{"sources"},
	Short:   "Deletes the given sources",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteSource(args...)
		printStatus(err)
	},
}

var fieldDeleteCmd = &cobra.Command{
	Use:     "field",
	Aliases: []string{"fields"},
	Short:   "Deletes the given fields",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteField(args...)
		printStatus(err)
	},
}

var firmwareDeleteCmd = &cobra.Command{
	Use:     "firmware",
	Aliases: []string{"firmwares", "fw"},
	Short:   "Deletes the given firmwares",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteFirmware(args...)
		printStatus(err)
	},
}

var dataRepositoryDeleteCmd = &cobra.Command{
	Use:     "data-repository",
	Aliases: []string{"data-repositories", "data-repo"},
	Short:   "Deletes the given data repositories",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteDataRepository(args...)
		printStatus(err)
	},
}

var virtualDeviceDeleteCmd = &cobra.Command{
	Use:     "virtual-device",
	Aliases: []string{"virtual-devices", "vd"},
	Short:   "Deletes the given virtual devices",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteVirtualDevice(args...)
		printStatus(err)
	},
}

var virtualAssistantDeleteCmd = &cobra.Command{
	Use:     "virtual-assistant",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Deletes the given virtual assistants",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteVirtualAssistant(args...)
		printStatus(err)
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Deletes the given tasks",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteTask(args...)
		printStatus(err)
	},
}

var scheduleDeleteCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"schedules"},
	Short:   "Deletes the given schedules",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteSchedule(args...)
		printStatus(err)
	},
}

var handlerDeleteCmd = &cobra.Command{
	Use:     "handler",
	Aliases: []string{"handlers"},
	Short:   "Deletes the given handlers",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteHandler(args...)
		printStatus(err)
	},
}

var forwardPayloadDeleteCmd = &cobra.Command{
	Use:     "forward-payload",
	Aliases: []string{"forward-payloads"},
	Short:   "Deletes the given forward payloads",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteForwardPayload(args...)
		printStatus(err)
	},
}

var backupDeleteCmd = &cobra.Command{
	Use:     "backup",
	Aliases: []string{"backups"},
	Short:   "Deletes the given backups",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()
		err := client.DeleteBackup(args...)
		printStatus(err)
	},
}
