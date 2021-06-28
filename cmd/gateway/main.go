package main

import (
	cmd "github.com/mycontroller-org/server/v2/cmd"
	"github.com/mycontroller-org/server/v2/pkg/start/gateway"
)

func main() {
	cmd.PrintVersion()
	gateway.Init() // init gateway service
}
