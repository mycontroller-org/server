package server

import (
	"fmt"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Cmd.AddCommand(serverCmd)
	serverCmd.AddCommand(infoCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Show server information",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

type infoRow struct {
	Field string
	Value string
}

var infoCmd = &cobra.Command{
	Use:   "info <alias>",
	Short: "Print version details of a MyController server",
	Example: `  myc server info <alias>`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := rootCmd.LookupClient(args[0])
		if err != nil {
			return err
		}
		info, err := client.GetServerVersion()
		if err != nil {
			return fmt.Errorf("failed to get server info: %w", err)
		}
		headers := []printer.Header{
			{Title: "field", ValuePath: "Field"},
			{Title: "value", ValuePath: "Value"},
		}
		rows := []interface{}{
			infoRow{"alias", args[0]},
			infoRow{"version", info.Version},
			infoRow{"build date", info.BuildDate},
			infoRow{"git commit", info.GitCommit},
			infoRow{"golang", info.GoVersion},
			infoRow{"platform", info.Platform},
			infoRow{"arch", info.Arch},
			infoRow{"host id", info.HostID},
		}
		printer.Print(rootCmd.IOStreams.Out, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
		return nil
	},
}
