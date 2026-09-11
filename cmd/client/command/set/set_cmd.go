package set

import (
	"github.com/spf13/cobra"
)

func init() {
	setCmd.AddCommand(newSetResourceCmd("gateway", []string{"gw", "gateways"}, "gateway"))
	setCmd.AddCommand(newSetResourceCmd("node", []string{"nodes"}, "node"))
	setCmd.AddCommand(newSetResourceCmd("source", []string{"sources"}, "source"))
	setCmd.AddCommand(newSetResourceCmd("field", []string{"fields"}, "field"))
	setCmd.AddCommand(newSetResourceCmd("firmware", []string{"firmwares", "fw"}, "firmware"))
	setCmd.AddCommand(newSetResourceCmd("data-repository", []string{"data-repositories", "data-repo", "datarepository"}, "data-repository"))
	setCmd.AddCommand(valueSetCmd)
}

func newSetResourceCmd(use string, aliases []string, kind string) *cobra.Command {
	return &cobra.Command{
		Use:     use + " <id> <key-path> [value]",
		Aliases: aliases,
		Short:   "Set a nested field on " + use + " resource(s)",
		Long: `Update a nested property on one or more resources.
The key path uses dots, for example formatter.onReceive or data.onConfig.

The value can be given as the last argument or read from --file.
To set a live field value, use myc set value field.
`,
		Example:       setExamples(use),
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			selectors, keyPath, value, err := parseSetArgs(args, setFile, setPath)
			if err != nil {
				return err
			}
			return executeSetPath(kind, selectors, keyPath, value)
		},
	}
}

func setExamples(use string) string {
	examples := `  myc set ` + use + ` <id> description "updated from cli"
  myc set ` + use + ` <id> data.onConfig --file onConfig.js`
	if use == "field" {
		examples = `  myc set field mysensor.1.dht.temperature formatter.onReceive --file on_receive.js
  myc set field mysensor.1.dht.temperature formatter.onReceive "return value;"
  myc set field <id> --path formatter.onReceive --file on_receive.js
  myc set field gw1.1.1.V_CUSTOM name "Custom"`
	}
	if use == "gateway" {
		examples = `  myc set gateway mysensor description "USB gateway"
  myc set gateway mysensor provider.protocol.script --file script.js`
	}
	if use == "node" {
		examples = `  myc set node mysensor.1 name "Living Room"
  myc set node <id> others.note --file note.txt`
	}
	if use == "source" {
		examples = `  myc set source mysensor.1.dht name "DHT"
  myc set source <id> others.script --file script.js`
	}
	if use == "firmware" {
		examples = `  myc set firmware stm32-app-slot-a description "slot A"
  myc set firmware stm32-app-slot-a labels.ms_flash_slot A`
	}
	if use == "data-repository" {
		examples = `  myc set data-repository ota_stm32_ab data.onConfig --file onConfig.js
  myc set data-repo ota_stm32_ab data.onBlock --file onBlock.js`
	}
	return examples
}
