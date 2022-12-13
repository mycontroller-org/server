package root

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	"github.com/mycontroller-org/server/v2/pkg/version"

	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(versionCmd)
}

type VersionMap struct {
	Spec map[string]interface{} `json:"spec"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the client and server version information",
	PreRun: func(cmd *cobra.Command, args []string) {
		UpdateStreams(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		headers := []printer.Header{
			{Title: "component", ValuePath: "spec.type"},
			{Title: "version", ValuePath: "spec.version"},
			{Title: "build date", ValuePath: "spec.buildDate"},
			{Title: "git commit", ValuePath: "spec.gitCommit"},
			{Title: "golang", ValuePath: "spec.goLang"},
			{Title: "platform", ValuePath: "spec.platform"},
			{Title: "arch", ValuePath: "spec.arch"},
			{Title: "host id", ValuePath: "spec.hostId"},
		}

		rows := make([]interface{}, 0)

		// client version details
		clientVersion := version.Get()
		clientRow := map[string]interface{}{
			"type":      "client",
			"version":   clientVersion.Version,
			"buildDate": clientVersion.BuildDate,
			"gitCommit": clientVersion.GitCommit,
			"goLang":    clientVersion.GoLang,
			"platform":  clientVersion.Platform,
			"arch":      clientVersion.Arch,
			"hostId":    clientVersion.HostID,
		}
		rows = append(rows, VersionMap{Spec: clientRow})

		serverRow := map[string]interface{}{"type": "server"}
		if CONFIG.URL == "" {
			serverRow["version"] = "not logged in"
		} else {
			client := GetClient()
			serverVersion, err := client.GetServerVersion()
			if err != nil {
				serverRow["version"] = fmt.Sprintf("error:%s", err)
			} else {
				serverRow["version"] = serverVersion.Version
				serverRow["buildDate"] = serverVersion.BuildDate
				serverRow["gitCommit"] = serverVersion.GitCommit
				serverRow["goLang"] = serverVersion.GoLang
				serverRow["platform"] = serverVersion.Platform
				serverRow["arch"] = serverVersion.Arch
				serverRow["hostId"] = serverVersion.HostID
			}
		}
		rows = append(rows, VersionMap{Spec: serverRow})

		printer.Print(IOStreams.Out, headers, rows, HideHeader, OutputFormat, Pretty)
	},
}
