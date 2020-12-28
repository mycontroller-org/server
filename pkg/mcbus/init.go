package mcbus

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	busml "github.com/mycontroller-org/backend/v2/plugin/bus"
	embedBus "github.com/mycontroller-org/backend/v2/plugin/bus/embedded"
	"github.com/mycontroller-org/backend/v2/plugin/bus/natsio"
)

var busClient busml.Client

// InitBus function
func InitBus(config cmap.CustomMap) error {
	client, err := initBusImpl(config)
	if err != nil {
		return err
	}
	busClient = client
	return nil
}

func initBusImpl(config cmap.CustomMap) (busml.Client, error) {
	busType := config.GetString(model.NameType)

	if busType == "" { // if non defined, update default
		busType = busml.TypeEmbedded
	}

	switch busType {
	case busml.TypeEmbedded:
		return embedBus.Init()
	case busml.TypeNatsIO:
		return natsio.Init(config)
	default:
		return nil, fmt.Errorf("Specified bus type not implemented. %s", busType)
	}
}
