package main

import (
	handler "github.com/mycontroller-org/server/v2/cmd/server/app/handler/start"
	server "github.com/mycontroller-org/server/v2/pkg/start/server"
)

func main() {
	server.Start(handler.StartWebHandler)
}
