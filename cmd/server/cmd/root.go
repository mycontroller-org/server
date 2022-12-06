package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	commonCmd "github.com/mycontroller-org/server/v2/cmd/common_cmd"
	handler "github.com/mycontroller-org/server/v2/cmd/server/app/handler/start"
	server "github.com/mycontroller-org/server/v2/pkg/start/server"
)

var (
	configFile string
)

func init() {
	root.PersistentFlags().StringVarP(&configFile, "config", "c", "./mycontroller.yaml", "MyController server configuration file")
	root.AddCommand(commonCmd.VersionCmd)
}

var root = &cobra.Command{
	Use:   "mycontroller-server",
	Short: "mycontroller-server",
	Long: `MyController Server
  Starts MyController server with the given configuration file.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		server.Start(configFile, handler.StartWebHandler)
	},
}

func Execute() {
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
