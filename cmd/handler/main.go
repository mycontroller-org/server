package main

import (
	cmd "github.com/mycontroller-org/server/v2/cmd/commands"
	"github.com/mycontroller-org/server/v2/pkg/start/handler"
)

func main() {
	cmd.ExecuteCommand(cmd.CallerHandler)
	handler.Init() // init handler service
}
