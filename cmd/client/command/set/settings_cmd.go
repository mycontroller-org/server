package set

import (
	"fmt"
	"os"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	setCmd.AddCommand(settingsSetCmd)
}

var settingsSetCmd = &cobra.Command{
	Use:     "settings <alias> [<key-path> [value]]",
	Aliases: []string{"setting"},
	Short:   "Update system settings",
	Long: `Update system settings on a server.

A key path is relative to the settings spec, for example language or geoLocation.latitude.
Use --file to read a single value or, without a key path, to merge a YAML/JSON object.

  myc set settings <alias> language en
  myc set settings <alias> geoLocation.autoUpdate true
  myc set settings <alias> login.message --file message.txt
  myc set settings <alias> --file settings.yaml
`,
	Example: `  myc set settings <alias> language en
  myc set settings <alias> geoLocation.latitude 12.97
  myc set settings <alias> --file settings.yaml`,
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client := rootCmd.MustClient(args[0])
		rest := args[1:]
		if setFile != "" && len(rest) == 0 && setPath == "" {
			data, err := os.ReadFile(setFile)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", setFile, err)
			}
			var overlay map[string]interface{}
			if err := yaml.Unmarshal(data, &overlay); err != nil {
				return fmt.Errorf("failed to parse %s: %w", setFile, err)
			}
			if err := client.MergeSystemSettings(overlay); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "set settings: merged %s\n", setFile)
			return nil
		}
		keyPath, value, err := parseSettingsArgs(rest, setFile, setPath)
		if err != nil {
			return err
		}
		if err := client.SetSystemSettingsPath(keyPath, value, setFile != ""); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "set settings: %s\n", keyPath)
		return nil
	},
}

func parseSettingsArgs(args []string, file, path string) (string, string, error) {
	if path != "" {
		if file != "" {
			value, err := readSetValue(file, "")
			return path, value, err
		}
		if len(args) < 1 {
			return "", "", fmt.Errorf("value or --file is required for key path %s", path)
		}
		return path, args[0], nil
	}
	if file != "" {
		if len(args) < 1 {
			return "", "", fmt.Errorf("key path is required")
		}
		value, err := readSetValue(file, "")
		return args[0], value, err
	}
	if len(args) == 1 {
		return "", "", fmt.Errorf("value or --file is required for key path %s", args[0])
	}
	if len(args) != 2 {
		return "", "", fmt.Errorf("key path and value are required (or use --file)")
	}
	return args[0], args[1], nil
}
