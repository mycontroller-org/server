package main

import (
	handler "github.com/mycontroller-org/backend/v2/cmd/core/app/handler/init"
	allInOne "github.com/mycontroller-org/backend/v2/pkg/init/all-in-one"
)

func main() {
	allInOne.Init(handler.InitHandler)
}
