package commands

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/version"

	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version details",
	Run: func(cmd *cobra.Command, args []string) {
		ver := version.Get()
		fmt.Printf("%+v\n", ver)
	},
}
