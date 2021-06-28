package main

import (
	cmd "github.com/mycontroller-org/server/v2/cmd"
	"github.com/mycontroller-org/server/v2/pkg/start/handler"
)

func main() {
	cmd.PrintVersion()
	handler.Init() // init handler service
}
