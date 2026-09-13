package get

import (
	"fmt"
	"sort"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	getCmd.AddCommand(settingsGetCmd)
}

type settingsRow struct {
	Key   string
	Value string
}

var settingsGetCmd = &cobra.Command{
	Use:     "settings <alias> [key-path]",
	Aliases: []string{"setting"},
	Short:   "Print system settings",
	Long: `Print system settings.

With no key path, lists all keys and values.
With a map key, lists all nested keys and values under it.
With a leaf key, prints only that key and value.

  myc get settings home
  myc get settings home geoLocation
  myc get settings home geoLocation.latitude
`,
	Example: `  myc get settings home
  myc get settings home geoLocation
  myc get settings home language
  myc get settings home geoLocation.latitude -o yaml`,
	Args: cobra.RangeArgs(1, 2),
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := rootCmd.MustClient(args[0])
		settings, err := client.GetSystemSettings()
		if err != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
			return
		}
		spec := settings.Spec
		if spec == nil {
			spec = map[string]interface{}{}
		}
		path := ""
		if len(args) == 2 {
			path = strings.TrimSpace(args[1])
		}
		selected, err := lookupSettingsPath(spec, path)
		if err != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
			return
		}
		printSettings(path, selected)
	},
}

func printSettings(path string, selected interface{}) {
	switch rootCmd.OutputFormat {
	case printer.OutputYAML, printer.OutputJSON:
		printer.Print(rootCmd.IOStreams.Out, nil, selected, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
		return
	}

	headers := []printer.Header{
		{Title: "key", ValuePath: "Key"},
		{Title: "value", ValuePath: "Value"},
	}
	if nested, ok := selected.(map[string]interface{}); ok {
		printer.Print(rootCmd.IOStreams.Out, headers, flattenSettings(nested, path), rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
		return
	}
	printer.Print(rootCmd.IOStreams.Out, headers, []interface{}{settingsRow{Key: path, Value: formatSettingValue(selected)}}, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
}

func lookupSettingsPath(spec map[string]interface{}, path string) (interface{}, error) {
	if path == "" {
		return spec, nil
	}
	current := interface{}(spec)
	for _, part := range strings.Split(path, ".") {
		if part == "" {
			return nil, fmt.Errorf("key path %q is not present", path)
		}
		nested, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("key path %q is not present", path)
		}
		next, ok := nested[part]
		if !ok {
			return nil, fmt.Errorf("key path %q is not present", path)
		}
		current = next
	}
	return current, nil
}

func flattenSettings(in map[string]interface{}, prefix string) []interface{} {
	rows := make([]interface{}, 0)
	for _, key := range sortedKeys(in) {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		value := in[key]
		if nested, ok := value.(map[string]interface{}); ok {
			rows = append(rows, flattenSettings(nested, path)...)
			continue
		}
		rows = append(rows, settingsRow{Key: path, Value: formatSettingValue(value)})
	}
	return rows
}

func formatSettingValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if _, ok := value.([]interface{}); ok {
		if raw, err := yaml.Marshal(value); err == nil {
			return string(bytesTrimRightNewline(raw))
		}
	}
	return fmt.Sprint(value)
}

func sortedKeys(in map[string]interface{}) []string {
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func bytesTrimRightNewline(data []byte) []byte {
	for len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == ' ') {
		data = data[:len(data)-1]
	}
	return data
}
