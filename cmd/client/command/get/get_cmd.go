package get

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
	clientTY "github.com/mycontroller-org/server/v2/pkg/types/client"
	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	fwPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"github.com/spf13/cobra"
)

func init() {
	getCmd.AddCommand(gwGetCmd)
	getCmd.AddCommand(nodeGetCmd)
	getCmd.AddCommand(sourceGetCmd)
	getCmd.AddCommand(fieldGetCmd)
	getCmd.AddCommand(firmwareGetCmd)
	getCmd.AddCommand(dataRepositoryGetCmd)
	getCmd.AddCommand(virtualDeviceGetCmd)
	getCmd.AddCommand(virtualAssistantGetCmd)
	getCmd.AddCommand(taskGetCmd)
	getCmd.AddCommand(scheduleGetCmd)
	getCmd.AddCommand(handlerGetCmd)
	getCmd.AddCommand(forwardPayloadGetCmd)
	getCmd.AddCommand(backupGetCmd)
}

var gwGetCmd = &cobra.Command{
	Use:     "gateway",
	Aliases: []string{"gw", "gateways"},
	Short:   "Print the gateway details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id"},
			{Title: "quick id", IsWide: true, ValueFunc: clientTY.GetQuickIDValueFunc(clientTY.ResourceGateway)},
			{Title: "description"},
			{Title: "enabled"},
			{Title: "reconnect delay", ValuePath: "reconnectDelay", IsWide: true},
			{Title: "provider", ValuePath: "provider.type"},
			{Title: "protocol", ValuePath: "provider.protocol.type"},
			{Title: "status", ValuePath: "state.status"},
			{Title: "message", ValuePath: "state.message"},
			{Title: "since", ValuePath: "state.since", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListGateway, gwTY.Config{})
	},
}

var nodeGetCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Short:   "Print the node details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id", IsWide: true},
			{Title: "quick id", IsWide: true, ValueFunc: clientTY.GetQuickIDValueFunc(clientTY.ResourceNode)},
			{Title: "gateway id", ValuePath: "gatewayId"},
			{Title: "node id", ValuePath: "nodeId"},
			{Title: "name"},
			{Title: "version", ValuePath: "labels.version"},
			{Title: "library version", ValuePath: "labels.library_version"},
			{Title: "battery", ValuePath: "others.battery_level"},
			{Title: "status", ValuePath: "state.status"},
			{Title: "since", ValuePath: "state.since", DisplayStyle: printer.DisplayStyleRelativeTime},
			{Title: "last seen", ValuePath: "lastSeen", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListNode, nodeTY.Node{})
	},
}

var sourceGetCmd = &cobra.Command{
	Use:     "source",
	Aliases: []string{"sources"},
	Short:   "Print the source details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id", IsWide: true},
			{Title: "quick id", IsWide: true, ValueFunc: clientTY.GetQuickIDValueFunc(clientTY.ResourceSource)},
			{Title: "gateway id", ValuePath: "gatewayId"},
			{Title: "node id", ValuePath: "nodeId"},
			{Title: "source id", ValuePath: "sourceId"},
			{Title: "name"},
			{Title: "last seen", ValuePath: "lastSeen", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListSource, sourceTY.Source{})
	},
}

var fieldGetCmd = &cobra.Command{
	Use:     "field",
	Aliases: []string{"fields"},
	Short:   "Print the field details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id", IsWide: true},
			{Title: "quick id", IsWide: true, ValueFunc: clientTY.GetQuickIDValueFunc(clientTY.ResourceField)},
			{Title: "gateway id", ValuePath: "gatewayId"},
			{Title: "node id", ValuePath: "nodeId"},
			{Title: "source id", ValuePath: "sourceId"},
			{Title: "field id", ValuePath: "fieldId"},
			{Title: "name"},
			{Title: "metric type", ValuePath: "metricType"},
			{Title: "value", ValuePath: "current.value"},
			{Title: "previous value", ValuePath: "previous.value"},
			{Title: "unit"},
			{Title: "last seen", ValuePath: "lastSeen", DisplayStyle: printer.DisplayStyleRelativeTime},
			{Title: "no change since", ValuePath: "noChangeSince", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListField, fieldTY.Field{})
	},
}

var firmwareGetCmd = &cobra.Command{
	Use:     "firmware",
	Aliases: []string{"firmwares", "fw"},
	Short:   "Print the firmware details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id"},
			{Title: "description"},
			{Title: "filename", ValuePath: "file.name"},
			{Title: "size", ValuePath: "file.size"},
			{Title: "labels"},
			{Title: "modified on", ValuePath: "file.modifiedOn", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListFirmware, firmwareTY.Firmware{})
	},
}

var dataRepositoryGetCmd = &cobra.Command{
	Use:     "data-repository",
	Aliases: []string{"data-repositories", "data-repo"},
	Short:   "Print the data repository details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id"},
			{Title: "description"},
			{Title: "readonly", ValuePath: "readOnly"},
			{Title: "labels"},
			{Title: "modified on", ValuePath: "modifiedOn", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListDataRepository, dataRepoTY.Config{})
	},
}

var virtualDeviceGetCmd = &cobra.Command{
	Use:     "virtual-device",
	Aliases: []string{"virtual-devices", "vd"},
	Short:   "Print the virtual device details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id", IsWide: true},
			{Title: "enabled"},
			{Title: "name"},
			{Title: "description"},
			{Title: "location"},
			{Title: "device type", ValuePath: "deviceType"},
			{Title: "labels"},
			{Title: "modified on", ValuePath: "modifiedOn", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListVirtualDevice, vdTY.VirtualDevice{})
	},
}

var virtualAssistantGetCmd = &cobra.Command{
	Use:     "virtual-assistant",
	Aliases: []string{"virtual-assistants", "va"},
	Short:   "Print the virtual assistant details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id"},
			{Title: "enabled"},
			{Title: "description"},
			{Title: "provider type", ValuePath: "providerType"},
			{Title: "device filter", ValuePath: "deviceFilter"},
			{Title: "status", ValuePath: "state.status"},
			{Title: "message", ValuePath: "state.message"},
			{Title: "since", ValuePath: "state.since", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListVirtualAssistant, vaTY.Config{})
	},
}

var taskGetCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Print the task details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id"},
			{Title: "description"},
			{Title: "enabled"},
			{Title: "ignore duplicate", ValuePath: "ignoreDuplicate", IsWide: true},
			{Title: "auto disable", ValuePath: "autoDisable", IsWide: true},
			{Title: "trigger on event", ValuePath: "triggerOnEvent"},
			// {Title: "message", ValuePath: "state.message"},
			{Title: "last duration", ValuePath: "state.lastDuration"},
			{Title: "last evaluation", ValuePath: "state.lastEvaluation", DisplayStyle: printer.DisplayStyleRelativeTime},
			{Title: "last success", ValuePath: "state.lastSuccess", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListTask, taskTY.Config{})
	},
}

var scheduleGetCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"schedules"},
	Short:   "Print the schedule details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id"},
			{Title: "description"},
			{Title: "enabled"},
			{Title: "type"},
			{Title: "executed count", ValuePath: "state.executedCount"},
			{Title: "message", ValuePath: "state.message"},
			{Title: "last run", ValuePath: "state.lastRun", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListSchedule, scheduleTY.Config{})
	},
}

var handlerGetCmd = &cobra.Command{
	Use:     "handler",
	Aliases: []string{"handlers"},
	Short:   "Print the handler details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id"},
			{Title: "description"},
			{Title: "enabled"},
			{Title: "type"},
			{Title: "status", ValuePath: "state.status"},
			{Title: "message", ValuePath: "state.message"},
			{Title: "since", ValuePath: "state.since", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListHandler, handlerTY.Config{})
	},
}

var forwardPayloadGetCmd = &cobra.Command{
	Use:     "forward-payload",
	Aliases: []string{"forward-payloads"},
	Short:   "Print the forward payload details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "id"},
			{Title: "description"},
			{Title: "enabled"},
			{Title: "source", ValuePath: "srcFieldId"},
			{Title: "destination", ValuePath: "dstFieldId"},
			{Title: "labels"},
		}
		executeGetCmd(headers, client.ListForwardPayload, fwPayloadTY.Config{})
	},
}

var backupGetCmd = &cobra.Command{
	Use:     "backup",
	Aliases: []string{"backups"},
	Short:   "Print the backup details",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.GetClient()

		headers := []printer.Header{
			{Title: "filename", ValuePath: "id"},
			{Title: "location name", ValuePath: "locationName"},
			{Title: "size", ValuePath: "fileSize"},
			{Title: "modified on", ValuePath: "modifiedOn", DisplayStyle: printer.DisplayStyleRelativeTime},
		}
		executeGetCmd(headers, client.ListBackup, backupTY.BackupFile{})
	},
}
