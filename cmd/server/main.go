package main

import (
	cmd "github.com/mycontroller-org/server/v2/cmd/commands"
	handler "github.com/mycontroller-org/server/v2/cmd/server/app/handler/start"
	server "github.com/mycontroller-org/server/v2/pkg/start/server"
)

func main() {
	cmd.ExecuteCommand(cmd.CallerServer)
	server.Start(handler.StartWebHandler)
}
