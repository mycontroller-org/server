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
	deleteCmd.AddCommand(userDeleteCmd)
	deleteCmd.AddCommand(policyDeleteCmd)
	deleteCmd.AddCommand(serviceAccountDeleteCmd)
}

var gwDeleteCmd = &cobra.Command{
	Use:     "gateway <alias> <id> [<id>...]",
	Aliases: []string{"gw", "gateways"},
	Short:   "Deletes the given gateways",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteGateway(ids...)
		printStatus(err)
	},
}

var nodeDeleteCmd = &cobra.Command{
	Use:     "node <alias> <id> [<id>...]",
	Aliases: []string{"nodes"},
	Short:   "Deletes the given nodes",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteNode(ids...)
		printStatus(err)
	},
}

var sourceDeleteCmd = &cobra.Command{
	Use:     "source <alias> <id> [<id>...]",
	Aliases: []string{"sources"},
	Short:   "Deletes the given sources",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteSource(ids...)
		printStatus(err)
	},
}

var fieldDeleteCmd = &cobra.Command{
	Use:     "field <alias> <id> [<id>...]",
	Aliases: []string{"fields"},
	Short:   "Deletes the given fields",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteField(ids...)
		printStatus(err)
	},
}

var firmwareDeleteCmd = &cobra.Command{
	Use:     "firmware <alias> <id> [<id>...]",
	Aliases: []string{"firmwares", "fw"},
	Short:   "Deletes the given firmwares",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteFirmware(ids...)
		printStatus(err)
	},
}

var dataRepositoryDeleteCmd = &cobra.Command{
	Use:     "data-repository <alias> <id> [<id>...]",
	Aliases: []string{"data-repositories", "data-repo"},
	Short:   "Deletes the given data repositories",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteDataRepository(ids...)
		printStatus(err)
	},
}

var virtualDeviceDeleteCmd = &cobra.Command{
	Use:     "virtual-device <alias> <id> [<id>...]",
	Aliases: []string{"virtual-devices", "vd"},
	Short:   "Deletes the given virtual devices",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteVirtualDevice(ids...)
		printStatus(err)
	},
}

var virtualAssistantDeleteCmd = &cobra.Command{
	Use:     "virtual-assistant <alias> <id> [<id>...]",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Deletes the given virtual assistants",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteVirtualAssistant(ids...)
		printStatus(err)
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:     "task <alias> <id> [<id>...]",
	Aliases: []string{"tasks"},
	Short:   "Deletes the given tasks",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteTask(ids...)
		printStatus(err)
	},
}

var scheduleDeleteCmd = &cobra.Command{
	Use:     "schedule <alias> <id> [<id>...]",
	Aliases: []string{"schedules"},
	Short:   "Deletes the given schedules",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteSchedule(ids...)
		printStatus(err)
	},
}

var handlerDeleteCmd = &cobra.Command{
	Use:     "handler <alias> <id> [<id>...]",
	Aliases: []string{"handlers"},
	Short:   "Deletes the given handlers",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteHandler(ids...)
		printStatus(err)
	},
}

var forwardPayloadDeleteCmd = &cobra.Command{
	Use:     "forward-payload <alias> <id> [<id>...]",
	Aliases: []string{"forward-payloads"},
	Short:   "Deletes the given forward payloads",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteForwardPayload(ids...)
		printStatus(err)
	},
}

var backupDeleteCmd = &cobra.Command{
	Use:     "backup <alias> <id> [<id>...]",
	Aliases: []string{"backups"},
	Short:   "Deletes the given backups",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		err := client.DeleteBackup(ids...)
		printStatus(err)
	},
}

var userDeleteCmd = &cobra.Command{
	Use:     "user <alias> <username-or-id> [<username-or-id>...]",
	Aliases: []string{"users"},
	Short:   "Deletes the given users",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, selectors := rootCmd.TakeAlias(args)
		ids, err := client.ResolveUserIDs(selectors)
		if err != nil {
			printStatus(err)
			return
		}
		printStatus(client.DeleteUser(ids...))
	},
}

var policyDeleteCmd = &cobra.Command{
	Use:     "policy <alias> <id> [<id>...]",
	Aliases: []string{"policies"},
	Short:   "Deletes the given policies",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, ids := rootCmd.TakeAlias(args)
		printStatus(client.DeletePolicy(ids...))
	},
}

var serviceAccountDeleteCmd = &cobra.Command{
	Use:     "service-account <alias> <name-or-id> [<name-or-id>...]",
	Aliases: []string{"service-accounts", "sa"},
	Short:   "Deletes the given service accounts",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, selectors := rootCmd.TakeAlias(args)
		ids, err := client.ResolveServiceAccountIDs(selectors)
		if err != nil {
			printStatus(err)
			return
		}
		printStatus(client.DeleteServiceAccount(ids...))
	},
}
