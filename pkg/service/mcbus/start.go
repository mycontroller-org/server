package mcbus

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	busML "github.com/mycontroller-org/server/v2/plugin/bus"
	embedBus "github.com/mycontroller-org/server/v2/plugin/bus/embedded"
	"github.com/mycontroller-org/server/v2/plugin/bus/natsio"
	"go.uber.org/zap"
)

var (
	busClient busML.Client
	pauseSRV  concurrency.SafeBool
)

// Start function
func Start(config cmap.CustomMap) {
	// update topic prefix
	topicPrefix = config.GetString(keyTopicPrefix)

	// update client
	client, err := startBusImpl(config)
	if err != nil {
		zap.L().Fatal("failed to init bus client", zap.Error(err))
	}
	busClient = client
	pauseSRV = concurrency.SafeBool{}
}

func startBusImpl(config cmap.CustomMap) (busML.Client, error) {
	busType := config.GetString(model.NameType)

	if busType == "" { // if non defined, update default
		busType = busML.TypeEmbedded
	}

	switch busType {
	case busML.TypeEmbedded:
		return embedBus.Init()
	case busML.TypeNatsIO:
		return natsio.Init(config)
	default:
		return nil, fmt.Errorf("specified bus type not implemented. %s", busType)
	}
}
