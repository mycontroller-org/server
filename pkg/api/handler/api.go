package handler

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
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

type HandlerAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	enc     *encryptionAPI.Encryption
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, enc *encryptionAPI.Encryption, bus busTY.Plugin) *HandlerAPI {
	return &HandlerAPI{
		ctx:     ctx,
		logger:  logger.Named("handler_api"),
		storage: storage,
		enc:     enc,
		bus:     bus,
	}
}

// List by filter and pagination
func (h *HandlerAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	out := make([]handlerTY.Config, 0)
	return h.storage.Find(types.EntityHandler, &out, filters, pagination)
}

// Get a config
func (h *HandlerAPI) Get(f []storageTY.Filter) (handlerTY.Config, error) {
	out := handlerTY.Config{}
	err := h.storage.FindOne(types.EntityHandler, &out, f)
	return out, err
}

// SaveAndReload handler
func (h *HandlerAPI) SaveAndReload(cfg *handlerTY.Config) error {
	cfg.State = &types.State{} // reset state
	err := h.Save(cfg)
	if err != nil {
		return err
	}
	return h.Reload([]string{cfg.ID})
}

// Save config
func (h *HandlerAPI) Save(cfg *handlerTY.Config) error {
	eventType := eventTY.TypeUpdated
	if cfg.ID == "" {
		cfg.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}

	// encrypt passwords
	err := h.enc.EncryptSecrets(cfg)
	if err != nil {
		return err
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: cfg.ID},
	}
	err = h.storage.Upsert(types.EntityHandler, cfg, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(h.logger, h.bus, topic.TopicEventHandler, eventType, types.EntityHandler, cfg)
	return nil
}

// SetState Updates state data
func (h *HandlerAPI) SetState(id string, state *types.State) error {
	cfg, err := h.GetByID(id)
	if err != nil {
		return err
	}
	cfg.State = state
	return h.Save(cfg)
}

// GetByTypeName returns a handler by type and name
func (h *HandlerAPI) GetByTypeName(handlerPluginType, name string) (*handlerTY.Config, error) {
	f := []storageTY.Filter{
		{Key: types.KeyHandlerType, Value: handlerPluginType},
		{Key: types.KeyHandlerName, Value: name},
	}
	out := &handlerTY.Config{}
	err := h.storage.FindOne(types.EntityHandler, out, f)
	return out, err
}

// GetByID returns a handler by id
func (h *HandlerAPI) GetByID(ID string) (*handlerTY.Config, error) {
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: ID},
	}
	out := &handlerTY.Config{}
	err := h.storage.FindOne(types.EntityHandler, out, f)
	return out, err
}

// Delete Service
func (h *HandlerAPI) Delete(ids []string) (int64, error) {
	err := h.Disable(ids)
	if err != nil {
		return 0, err
	}
	f := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	return h.storage.Delete(types.EntityHandler, f)
}

func (h *HandlerAPI) Import(data interface{}) error {
	input, ok := data.(handlerTY.Config)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return h.storage.Upsert(types.EntityHandler, &input, filters)
}

func (h *HandlerAPI) GetEntityInterface() interface{} {
	return handlerTY.Config{}
}
