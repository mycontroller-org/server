package main

import (
	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	clientTY "github.com/mycontroller-org/server/v2/pkg/types/client"

	_ "github.com/mycontroller-org/server/v2/cmd/client/command/delete"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/disable"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/enable"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/get"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/reload"
	_ "github.com/mycontroller-org/server/v2/cmd/client/command/set"
)

func main() {
	streams := clientTY.NewStdStreams()
	rootCmd.Execute(streams)
}
