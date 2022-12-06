package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	commonCmd "github.com/mycontroller-org/server/v2/cmd/common_cmd"
	"github.com/mycontroller-org/server/v2/pkg/start/gateway"
)

var (
	configFile string
)

func init() {
	root.PersistentFlags().StringVarP(&configFile, "config", "c", "./gateway.yaml", "MyController gateway service configuration file")
	root.AddCommand(commonCmd.VersionCmd)
}

var root = &cobra.Command{
	Use:   "mycontroller-gateway",
	Short: "mycontroller-gateway",
	Long: `MyController gateway service
  Starts MyController gateway service with the given configuration file.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		gateway.Init(configFile) // init gateway service
	},
}

func Execute() {
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
