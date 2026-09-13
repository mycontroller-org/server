package main

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	clientTY "github.com/mycontroller-org/server/v2/pkg/types/client"

	_ "github.com/mycontroller-org/server/v2/cmd/client/command/action"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/alias"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/apply"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/delete"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/disable"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/enable"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/get"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/reboot"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/reload"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/server"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/set"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/upload"
)

func main() {
	streams := clientTY.NewStdStreams()
	rootCmd.Execute(streams)
}
