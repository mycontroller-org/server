package main

import (
	cmd "github.com/mycontroller-org/server/v2/cmd"
	handler "github.com/mycontroller-org/server/v2/cmd/server/app/handler/start"
	server "github.com/mycontroller-org/server/v2/pkg/start/server"
)

func main() {
	cmd.PrintVersion()
	server.Start(handler.StartWebHandler)
}
