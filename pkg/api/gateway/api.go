package gateway

import (
	"context"
	"errors"
	"fmt"

	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

type GatewayAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	enc     *encryptionAPI.Encryption
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, enc *encryptionAPI.Encryption, bus busTY.Plugin) *GatewayAPI {
	return &GatewayAPI{
		ctx:     ctx,
		logger:  logger.Named("gateway_api"),
		storage: storage,
		enc:     enc,
		bus:     bus,
	}
}

// List by filter and pagination
func (gw *GatewayAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]gwTY.Config, 0)
	return gw.storage.Find(types.EntityGateway, &result, filters, pagination)
}

// Get returns a gateway
func (gw *GatewayAPI) Get(filters []storageTY.Filter) (*gwTY.Config, error) {
	result := &gwTY.Config{}
	err := gw.storage.FindOne(types.EntityGateway, result, filters)
	return result, err
}

// GetByIDs returns a gateway details by id
func (gw *GatewayAPI) GetByIDs(ids []string) ([]gwTY.Config, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids},
	}
	pagination := &storageTY.Pagination{Limit: int64(len(ids))}
	gateways := make([]gwTY.Config, 0)
	_, err := gw.storage.Find(types.EntityGateway, &gateways, filters, pagination)
	return gateways, err
}

// GetByID returns a gateway details
func (gw *GatewayAPI) GetByID(id string) (*gwTY.Config, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := &gwTY.Config{}
	err := gw.storage.FindOne(types.EntityGateway, result, filters)
	return result, err
}

// SaveAndReload gateway
func (gw *GatewayAPI) SaveAndReload(gwCfg *gwTY.Config) error {
	gwCfg.State = &types.State{} //reset state
	err := gw.Save(gwCfg)
	if err != nil {
		return err
	}
	return gw.Reload([]string{gwCfg.ID})
}

// Save gateway config
func (gw *GatewayAPI) Save(gwCfg *gwTY.Config) error {
	eventType := eventTY.TypeUpdated
	if gwCfg.ID == "" {
		gwCfg.ID = utils.RandID()
		eventType = eventTY.TypeCreated
	}

	// encrypt passwords, tokens
	err := gw.enc.EncryptSecrets(gwCfg)
	if err != nil {
		return err
	}

	err = gw.storage.Upsert(types.EntityGateway, gwCfg, nil)
	if err != nil {
		return err
	}
	busUtils.PostEvent(gw.logger, gw.bus, topic.TopicEventGateway, eventType, types.EntityGateway, gwCfg)
	return nil
}

// SetState Updates state data
func (gw *GatewayAPI) SetState(id string, state *types.State) error {
	gwCfg, err := gw.GetByID(id)
	if err != nil {
		return err
	}
	gwCfg.State = state
	return gw.Save(gwCfg)
}

// Delete gateway
func (gw *GatewayAPI) Delete(ids []string) (int64, error) {
	err := gw.Disable(ids)
	if err != nil {
		return 0, err
	}

	// delete one by one and send deletion event
	gateways, err := gw.GetByIDs(ids)
	if err != nil {
		return 0, err
	}
	deleted := int64(0)
	for _, gateway := range gateways {
		filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: gateway.ID}}
		_, err = gw.storage.Delete(types.EntityGateway, filters)
		if err != nil {
			return deleted, err
		}
		deleted++
		// deletion event
		busUtils.PostEvent(gw.logger, gw.bus, topic.TopicEventGateway, eventTY.TypeDeleted, types.EntityGateway, gateway)
	}

	return deleted, nil
}

func (gw *GatewayAPI) Import(data interface{}) error {
	input, ok := data.(gwTY.Config)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return gw.storage.Upsert(types.EntityGateway, &input, filters)
}

func (gw *GatewayAPI) GetEntityInterface() interface{} {
	return gwTY.Config{}
}
