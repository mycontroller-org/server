package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	commonCmd "github.com/mycontroller-org/server/v2/cmd/common_cmd"
	"github.com/mycontroller-org/server/v2/pkg/start/handler"
)

var (
	configFile string
)

func init() {
	root.PersistentFlags().StringVarP(&configFile, "config", "c", "./handler.yaml", "MyController handler service configuration file")
	root.AddCommand(commonCmd.VersionCmd)
}

var root = &cobra.Command{
	Use:   "mycontroller-handler",
	Short: "mycontroller-handler",
	Long: `MyController handler service
  Starts MyController handler service with the given configuration file.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		handler.Init(configFile) // init handler service
	},
}

func Execute() {
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
