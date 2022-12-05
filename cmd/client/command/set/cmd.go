package set

import (
	"fmt"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"

	"github.com/spf13/cobra"
)

const (
	QuickIDPrefixGateway = "gateway"
	QuickIDPrefixNode    = "node"
	QuickIDPrefixField   = "field"
)

// var (
// 	labelSlice []string
// )

func init() {
	rootCmd.Cmd.AddCommand(setCmd)
	// setCmd.PersistentFlags().StringSliceVarP(&labelSlice, "label", "l", []string{}, "filter the resource by label. comma separated or repeated label=value")
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets the value to the given resource(s)",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

func executeSetCmd(quickIdPrefix, keyPath string, resources []string, payload string) {
	client := rootCmd.GetClient()
	actions := []webHandlerTY.ActionConfig{}
	for _, resource := range resources {
		action := webHandlerTY.ActionConfig{
			Resource: fmt.Sprintf("%s:%s", quickIdPrefix, resource),
			KayPath:  keyPath,
			Payload:  payload,
		}
		actions = append(actions, action)
	}
	err := client.ExecuteAction(actions)
	if err != nil {
		fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s", err.Error())
	}
}
