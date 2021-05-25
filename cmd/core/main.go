package main

import (
	handler "github.com/mycontroller-org/backend/v2/cmd/core/start_handler"
	"github.com/mycontroller-org/backend/v2/pkg/init/core"
)

func main() {
	core.Init(handler.StartHandler)
}
