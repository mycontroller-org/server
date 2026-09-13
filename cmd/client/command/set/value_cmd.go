package set

import (
	"fmt"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/spf13/cobra"
)

var valueSetCmd = &cobra.Command{
	Use:   "value",
	Short: "Set a live field value",
	Long:  `Set a live sensor/actuator value. This sends an action; it does not change stored resource metadata.`,
}

func init() {
	valueSetCmd.AddCommand(valueFieldCmd)
}

var valueFieldCmd = &cobra.Command{
	Use:     "field <alias> <quick-id> [quick-id...] <payload>",
	Aliases: []string{"fields"},
	Short:   "Set a live field value",
	Example: `  myc set value field home gw1.1.1.V_CUSTOM 23.5
  myc set value field home mysensor.1.dht.temperature 21.0`,
	Args:          cobra.MinimumNArgs(3),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if setFile != "" || setPath != "" {
			return fmt.Errorf("myc set value field does not use --file or --path; use myc set field to update stored properties")
		}
		client, rest := rootCmd.TakeAlias(args)
		payload := rest[len(rest)-1]
		return executeSetFieldValue(client, rest[:len(rest)-1], payload)
	},
}
