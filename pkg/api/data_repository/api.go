package datarepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	"github.com/mycontroller-org/server/v2/pkg/json"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	repositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type DataRepositoryAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	enc     *encryptionAPI.Encryption
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, enc *encryptionAPI.Encryption, bus busTY.Plugin) *DataRepositoryAPI {
	return &DataRepositoryAPI{
		ctx:     ctx,
		logger:  logger.Named("data_repository_api"),
		storage: storage,
		enc:     enc,
		bus:     bus,
	}
}

// List by filter and pagination
func (dr *DataRepositoryAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]repositoryTY.Config, 0)
	return dr.storage.Find(types.EntityDataRepository, &result, filters, pagination)
}

// Get returns a item
func (dr *DataRepositoryAPI) Get(filters []storageTY.Filter) (*repositoryTY.Config, error) {
	result := &repositoryTY.Config{}
	err := dr.storage.FindOne(types.EntityDataRepository, result, filters)
	if err == nil {
		updateResult, err := dr.updateResult(result)
		if err != nil {
			return nil, err
		}
		result = updateResult
	}
	return result, err
}

// Save is used to update items from UI
func (dr *DataRepositoryAPI) Save(data *repositoryTY.Config) error {
	if data.ID == "" {
		return errors.New("'id' can not be empty")
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: data.ID},
	}

	data.ModifiedOn = time.Now()

	// encrypt passwords, tokens
	err := dr.enc.EncryptSecrets(data)
	if err != nil {
		return err
	}

	// in mongodb can not save map[interface{}]interface{} type
	// convert it to map[string]interface{} type
	if dr.storage.Name() == storageTY.TypeMongoDB {
		updatedResult, err := dr.updateResult(data)
		if err != nil {
			return err
		}
		data = updatedResult
	}

	err = dr.storage.Upsert(types.EntityDataRepository, data, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(dr.logger, dr.bus, topic.TopicEventDataRepository, eventTY.TypeUpdated, types.EntityDataRepository, data)
	return nil
}

// GetByID returns a item by id
func (dr *DataRepositoryAPI) GetByID(id string) (*repositoryTY.Config, error) {
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	out := &repositoryTY.Config{}
	err := dr.storage.FindOne(types.EntityDataRepository, out, f)
	if err == nil {
		updatedResult, err := dr.updateResult(out)
		if err != nil {
			return nil, err
		}
		out = updatedResult
	}
	return out, err
}

// Delete items
func (dr *DataRepositoryAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return dr.storage.Delete(types.EntityDataRepository, filters)
}

// map[interface{}]interface{} type not working as expected in javascript in task module
// convert it to map[string]interface{}, by calling json Marshal and Unmarshal
func (dr *DataRepositoryAPI) updateResult(data *repositoryTY.Config) (*repositoryTY.Config, error) {
	updateResult := &repositoryTY.Config{}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dataBytes, updateResult)
	if err != nil {
		return nil, err
	}
	return updateResult, nil
}

func (dr *DataRepositoryAPI) Import(data interface{}) error {
	input, ok := data.(repositoryTY.Config)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}

	return dr.storage.Upsert(types.EntityDataRepository, &input, filters)
}

func (dr *DataRepositoryAPI) GetEntityInterface() interface{} {
	return repositoryTY.Config{}
}
