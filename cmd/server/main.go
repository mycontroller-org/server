package main

import (
	handler "github.com/mycontroller-org/backend/v2/cmd/server/app/handler/start"
	server "github.com/mycontroller-org/backend/v2/pkg/start/server"
)

func main() {
	server.Start(handler.StartWebHandler)
}
