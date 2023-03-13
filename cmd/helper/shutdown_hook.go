package helper

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

type ShutdownHook struct {
	logger       *zap.Logger
	callbackFunc func()
	handleEvent  bool
	bus          busTY.Plugin
}

func NewShutdownHook(logger *zap.Logger, callbackFunc func(), bus busTY.Plugin, handleEvent bool) *ShutdownHook {
	return &ShutdownHook{
		logger:       logger.Named("shutdown_hook"),
		callbackFunc: callbackFunc,
		handleEvent:  handleEvent,
		bus:          bus,
	}
}

func (sh *ShutdownHook) Start() {
	if sh.handleEvent {
		sh.handleShutdownEvent()
	}
	sh.handleShutdownSignal()
}

// handel process shutdown signal
func (sh *ShutdownHook) handleShutdownSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// waiting for signal
	sig := <-sigs
	close(sigs)

	sh.logger.Info("shutdown initiated..", zap.Any("signal", sig))
	sh.triggerShutdown()
}

// handle shutdown events
func (sh *ShutdownHook) handleShutdownEvent() {
	shutdownFunc := func(data *busTY.BusData) {
		sh.logger.Info("shutdown initiated..", zap.String("signal", "internal event"))
		sh.triggerShutdown()
	}
	_, err := sh.bus.Subscribe(topic.TopicInternalShutdown, shutdownFunc)
	if err != nil {
		sh.logger.Fatal("error on subscribing shutdown event", zap.Error(err))
		return
	}

}

func (sh *ShutdownHook) triggerShutdown() {
	start := time.Now()

	// trigger callback function
	if sh.callbackFunc != nil {
		sh.callbackFunc()
	}

	sh.logger.Info("closing services completed", zap.String("timeTaken", time.Since(start).String()))
	sh.logger.Debug("bye, see you soon :)")

	// stop web/api service
	os.Exit(0)
}
