package main

import (
	cmd "github.com/mycontroller-org/server/v2/cmd/commands"
	"github.com/mycontroller-org/server/v2/pkg/start/gateway"
)

func main() {
	cmd.ExecuteCommand(cmd.CallerGateway)
	gateway.Init() // init gateway service
}
