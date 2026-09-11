package set

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	"github.com/spf13/cobra"
)

var keyPathPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)+$`)

const (
	QuickIDPrefixField = "field"
)

var (
	setFile string
	setPath string
)

func init() {
	rootCmd.Cmd.AddCommand(setCmd)
	setCmd.PersistentFlags().StringVarP(&setFile, "file", "f", "", "read the value from this file")
	setCmd.PersistentFlags().StringVar(&setPath, "path", "", "nested field path, for example formatter.onReceive")
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a nested property on a resource, or a live field value",
	Long: `Update a nested property (scripts and other text) on a resource.

  myc set field gw1.1.1.V_CUSTOM formatter.onReceive --file on_receive.js
  myc set data-repository ota_stm32_ab data.onConfig --file onConfig.js

Set a live field value with a separate command:

  myc set value field gw1.1.1.V_CUSTOM 23.5
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

func executeSetFieldValue(resources []string, payload string) error {
	client := rootCmd.GetClient()
	actions := make([]webHandlerTY.ActionConfig, 0, len(resources))
	for _, resource := range resources {
		actions = append(actions, webHandlerTY.ActionConfig{
			Resource: fmt.Sprintf("%s:%s", QuickIDPrefixField, resource),
			Payload:  payload,
		})
	}
	if err := client.ExecuteAction(actions); err != nil {
		return fmt.Errorf("error: %s", err)
	}
	for _, resource := range resources {
		_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "set field: %s\n", resource)
	}
	return nil
}

func executeSetPath(kind string, selectors []string, keyPath, value string) error {
	client := rootCmd.GetClient()
	failed := 0
	for _, selector := range selectors {
		if err := client.SetResourcePath(kind, selector, keyPath, value, setFile != ""); err != nil {
			_, _ = fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error: %s %s: %s\n", kind, selector, err)
			failed++
			continue
		}
		_, _ = fmt.Fprintf(rootCmd.IOStreams.Out, "set %s: %s %s\n", kind, selector, keyPath)
	}
	if failed > 0 {
		return fmt.Errorf("failed to set %d resource(s)", failed)
	}
	return nil
}

func readSetValue(filename, inline string) (string, error) {
	if filename != "" {
		data, err := os.ReadFile(filename)
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", filename, err)
		}
		return string(data), nil
	}
	return inline, nil
}

func parseSetArgs(args []string, file, path string) (selectors []string, keyPath, value string, err error) {
	if path != "" {
		if len(args) < 1 {
			return nil, "", "", fmt.Errorf("resource id is required")
		}
		if file != "" {
			value, err = readSetValue(file, "")
			return args, path, value, err
		}
		if len(args) < 2 {
			return nil, "", "", fmt.Errorf("value or --file is required for key path %s", path)
		}
		return args[:len(args)-1], path, args[len(args)-1], nil
	}
	if file != "" {
		if len(args) < 2 {
			return nil, "", "", fmt.Errorf("resource id and key path are required")
		}
		value, err = readSetValue(file, "")
		return args[:len(args)-1], args[len(args)-1], value, err
	}
	if len(args) >= 1 && looksLikeKeyPath(args[len(args)-1]) {
		return nil, "", "", fmt.Errorf("value or --file is required for key path %s", args[len(args)-1])
	}
	if len(args) < 3 {
		return nil, "", "", fmt.Errorf("resource id, key path, and value are required (or use --file)")
	}
	return args[:len(args)-2], args[len(args)-2], args[len(args)-1], nil
}

func looksLikeKeyPath(value string) bool {
	return keyPathPattern.MatchString(strings.TrimSpace(value))
}
