package upload

import (
	"fmt"
	"io"
	"time"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	"github.com/nleeper/goment"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Cmd.AddCommand(uploadCmd)
	uploadCmd.AddCommand(firmwareUploadCmd)
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload files to the server",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

var firmwareUploadCmd = &cobra.Command{
	Use:     "firmware <alias> <id> <file>",
	Aliases: []string{"fw"},
	Short:   "Upload a firmware binary to an existing firmware resource",
	Long: `Upload a firmware binary to an existing firmware resource.

Create the firmware metadata first with myc apply, then upload the file:

  myc apply home -f firmware.yaml
  myc upload firmware home stm32-app ./app.signed.bin
`,
	Example: `  myc upload firmware home stm32-app ./app.signed.bin
  myc upload fw home stm32-app ./app.bin`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(3),
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client := rootCmd.MustClient(args[0])
		id := args[1]
		filename := args[2]
		existing, err := client.FindFirmware(id)
		if err != nil {
			return fmt.Errorf("failed to look up firmware %s: %w", id, err)
		}
		if existing == nil {
			return fmt.Errorf("firmware %s is not present", id)
		}
		if err := client.UploadFirmware(id, filename); err != nil {
			return fmt.Errorf("failed to upload %s: %w", filename, err)
		}
		firmware, err := client.FindFirmware(id)
		if err != nil {
			return fmt.Errorf("uploaded, but failed to load firmware details: %w", err)
		}
		if firmware == nil {
			return fmt.Errorf("uploaded, but firmware %s is not present", id)
		}
		printFirmwareUpload(rootCmd.IOStreams.Out, firmware)
		return nil
	},
}

func printFirmwareUpload(out io.Writer, firmware *firmwareTY.Firmware) {
	file := firmware.File
	rows := [][2]string{
		{"firmware", firmware.ID},
		{"name", file.Name},
		{"internal name", file.InternalName},
		{"size", formatFileSize(file.Size)},
		{"checksum", file.Checksum},
		{"modified", formatRelativeTime(file.ModifiedOn)},
	}
	width := 0
	for _, row := range rows {
		if len(row[0]) > width {
			width = len(row[0])
		}
	}
	_, _ = fmt.Fprintln(out, "uploaded")
	for _, row := range rows {
		if row[1] == "" {
			continue
		}
		_, _ = fmt.Fprintf(out, "  %-*s  %s\n", width, row[0], row[1])
	}
}

func formatFileSize(size int) string {
	if size <= 0 {
		return "0 B"
	}
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB"}
	value := float64(size)
	unit := "B"
	for _, next := range units {
		if value < 1024 {
			break
		}
		value /= 1024
		unit = next
	}
	return fmt.Sprintf("%.2f %s", value, unit)
}

func formatRelativeTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	g, err := goment.New(value)
	if err != nil {
		return value.Format(time.RFC3339)
	}
	return g.FromNow()
}
