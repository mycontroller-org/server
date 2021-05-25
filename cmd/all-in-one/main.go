package main

import (
	handler "github.com/mycontroller-org/backend/v2/cmd/core/start_handler"
	allinone "github.com/mycontroller-org/backend/v2/pkg/init/all-in-one"
)

func main() {
	allinone.Init(handler.StartHandler)
}
