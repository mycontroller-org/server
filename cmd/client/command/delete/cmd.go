package delete

import (
	"fmt"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Cmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes the requested resources",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

func printStatus(err error) {
	if err != nil {
		fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
		return
	}
	fmt.Fprintln(rootCmd.IOStreams.Out, "Deleted successfully")
}
