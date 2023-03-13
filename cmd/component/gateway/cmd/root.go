package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mycontroller-org/server/v2/cmd/helper"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
)

var (
	configFile string
	logger     *zap.Logger
)

func init() {
	root.PersistentFlags().StringVarP(&configFile, "config", "c", "./gateway.yaml", "MyController gateway service configuration file")
	root.AddCommand(helper.VersionCmd)

	// logger used only here
	logger = loggerUtils.GetLogger(loggerUtils.ModeRecordAll, "info", "console", false, 0, false)
}

var root = &cobra.Command{
	Use:   "mycontroller-gateway",
	Short: "mycontroller-gateway",
	Long: `MyController gateway service
  Starts MyController gateway service with the given configuration file.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		gateway := helper.Gateway{}
		err := gateway.Start(ctx, configFile)
		if err != nil {
			logger.Fatal("error on starting gateway", zap.Error(err))
		}
	},
}

func Execute() {
	if err := root.Execute(); err != nil {
		logger.Fatal("error", zap.Error(err))
		os.Exit(1)
	}
}
