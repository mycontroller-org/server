package main

import (
	"go.uber.org/zap"

	"github.com/mycontroller-org/backend/v2/cmd/core/app/handler"
	"github.com/mycontroller-org/backend/v2/pkg/init/core"
)

func main() {
	core.Init(startHandler)
}

func startHandler() {
	err := handler.StartHandler()
	if err != nil {
		zap.L().Fatal("Error on starting http handler", zap.Error(err))
	}
}
