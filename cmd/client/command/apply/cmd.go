package apply

import (
	"errors"
	"fmt"
	"io"
	"os"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/spf13/cobra"
)

var (
	filenameSlice []string
	dryRun        bool
	replace       bool
)

func init() {
	rootCmd.Cmd.AddCommand(applyCmd)
	applyCmd.Flags().StringSliceVarP(&filenameSlice, "filename", "f", []string{}, "YAML or JSON file with resources. Use '-' to read from stdin. Repeat for multiple files")
	applyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "validate and report actions without changing resources")
	applyCmd.Flags().BoolVar(&replace, "replace", false, "if a resource already exists on add, delete it and recreate it with the same id")
	_ = applyCmd.MarkFlagRequired("filename")
}

var applyCmd = &cobra.Command{
	Use:           "apply <alias>",
	Short:         "Add, merge, or delete resources from a YAML or JSON file",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `Apply gateways, nodes, sources, fields, firmware, data repositories, users, policies, and service accounts from a YAML or JSON file.

Each resource must include kind (gateway, node, source, field, firmware, data-repository, user, policy, service-account) and operation (add, merge, delete).
Firmware binary files are not part of apply; upload them with myc upload firmware.
Adding a service account prints the token once; save it immediately.
Add fails when the resource already exists, unless --replace is set or the
resource has replace: true.
With replace, the existing resource is deleted and recreated with the same id
so references to that id stay valid.
Add, merge, and replace fail when the parent resource is not present
(node needs gateway, source needs node, field needs source).
A parent added earlier in the same file counts as present.
Delete of a missing resource is reported as not available and the remaining resources are still applied.

YAML example:

  kind: gateway
  operation: add
  id: mysensor
  description: MySensors USB
  enabled: true
  ---
  kind: node
  operation: add
  gatewayId: mysensor
  nodeId: "1"
  name: Living Room
  ---
  kind: source
  operation: merge
  gatewayId: mysensor
  nodeId: "1"
  sourceId: dht
  name: DHT Sensor
  ---
  kind: field
  operation: delete
  gatewayId: mysensor
  nodeId: "1"
  sourceId: dht
  fieldId: temperature

  # shared kind/operation/replace with an items list
  kind: field
  operation: add
  replace: true
  items:
    - gatewayId: mysensor
      nodeId: "1"
      sourceId: dht
      fieldId: temperature
      name: Temperature
      metricType: gauge
      unit: °C

If an item includes fieldId, it is applied as a field even when kind is source.
A JSON array of the same objects is also supported.
`,
	Example: `  myc apply <alias> -f resources.yaml
  myc apply <alias> -f resources.yaml --dry-run
  myc apply <alias> -f nodes.yaml -f sources.yaml --replace
  myc apply <alias> -f - --dry-run < resources.json`,
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client := rootCmd.MustClient(args[0])
		err := runApply(newAPIResourceClient(client), filenameSlice, replace, dryRun, rootCmd.IOStreams.In, rootCmd.IOStreams.Out, rootCmd.IOStreams.ErrOut)
		if errors.Is(err, ErrApplyFailed) {
			os.Exit(1)
		}
		return err
	},
}

func runApply(client ResourceClient, filenames []string, replace, dryRun bool, in io.Reader, out, errOut io.Writer) error {
	if len(filenames) == 0 {
		return fmt.Errorf("must specify at least one --filename")
	}

	resources := make([]Resource, 0)
	for _, filename := range filenames {
		data, source, err := readInput(filename, in)
		if err != nil {
			return err
		}
		parsed, err := ParseResources(data, source)
		if err != nil {
			return err
		}
		resources = append(resources, parsed...)
	}

	return Apply(client, resources, replace, dryRun, out, errOut)
}

func readInput(filename string, in io.Reader) ([]byte, string, error) {
	if filename == "-" {
		data, err := io.ReadAll(in)
		if err != nil {
			return nil, "stdin", fmt.Errorf("failed to read stdin: %w", err)
		}
		return data, "stdin", nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, filename, fmt.Errorf("failed to read %s: %w", filename, err)
	}
	return data, filename, nil
}
