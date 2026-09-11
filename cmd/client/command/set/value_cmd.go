package set

import (
	"fmt"

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
	Use:     "field <quick-id> [quick-id...] <payload>",
	Aliases: []string{"fields"},
	Short:   "Set a live field value",
	Example: `  myc set value field gw1.1.1.V_CUSTOM 23.5
  myc set value field mysensor.1.dht.temperature 21.0`,
	Args:          cobra.MinimumNArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if setFile != "" || setPath != "" {
			return fmt.Errorf("myc set value field does not use --file or --path; use myc set field to update stored properties")
		}
		payload := args[len(args)-1]
		return executeSetFieldValue(args[:len(args)-1], payload)
	},
}
