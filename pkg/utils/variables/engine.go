package variables

import (
	"context"
	"fmt"
	"reflect"
	"time"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	webhookTimeout = time.Second * 10
)

type genericAPI interface {
	List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error)
}

type VariableSpec struct {
	logger         *zap.Logger
	api            *entitiesAPI.API
	genericApiMap  map[string]genericAPI
	templateEngine types.TemplateEngine
	enc            *encryptionAPI.Encryption
}

func New(ctx context.Context, templateEngine types.TemplateEngine) (types.VariablesEngine, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	api, err := entitiesAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	enc, err := encryptionAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	apiMap := map[string]genericAPI{
		quickIdUtils.QuickIdDataRepository: api.DataRepository(),
		quickIdUtils.QuickIdField:          api.Field(),
		quickIdUtils.QuickIdFirmware:       api.Firmware(),
		quickIdUtils.QuickIdForwardPayload: api.ForwardPayload(),
		quickIdUtils.QuickIdGateway:        api.Gateway(),
		quickIdUtils.QuickIdHandler:        api.Handler(),
		quickIdUtils.QuickIdNode:           api.Node(),
		quickIdUtils.QuickIdSchedule:       api.Schedule(),
		quickIdUtils.QuickIdSource:         api.Source(),
		quickIdUtils.QuickIdTask:           api.Task(),
	}

	return &VariableSpec{
		logger:         logger,
		api:            api,
		genericApiMap:  apiMap,
		templateEngine: templateEngine,
		enc:            enc,
	}, nil
}

func (v *VariableSpec) TemplateEngine() types.TemplateEngine {
	return v.templateEngine
}

// loads all the defined variables
func (v *VariableSpec) Load(variablesPreMap map[string]interface{}) (map[string]interface{}, error) {
	variables := make(map[string]interface{})
	for name, variable := range variablesPreMap {
		variableData, ok := variable.(map[string]interface{})
		if !ok {
			v.logger.Error("error on converting variable data to cmap", zap.String("name", name), zap.String("actualType", fmt.Sprintf("%T", variable)))
			continue
		}
		value := v.getEntity(name, variableData)
		if value == nil {
			return nil, fmt.Errorf("failed to load a variable. name: %s, config:%+v", name, variable)
		}
		variables[name] = value
	}

	// clone variables
	clonedVariables := cloneUtil.Clone(variables)

	backToVariables, ok := clonedVariables.(map[string]interface{})
	if ok {
		// decrypt the secrets, tokens
		err := v.enc.DecryptSecrets(backToVariables)
		if err != nil {
			return nil, err
		}

		return backToVariables, nil
	}

	v.logger.Error("error on clone, returning variables as is", zap.String("actualType", fmt.Sprintf("%T", variables)))
	return variables, nil
}

func (v *VariableSpec) getEntity(name string, variable cmap.CustomMap) interface{} {
	switch variable.GetString(types.KeyType) {
	case types.VariableTypeString:
		stringValue := variable.GetString(types.KeyValue)
		formattedString, err := v.templateEngine.Execute(stringValue, nil)
		if err != nil {
			return fmt.Sprintf("error on executing template. name:%s, template:%s, error:%s", name, stringValue, err.Error())
		}
		return formattedString

	case types.VariableTypeWebhook:
		webhookCfg := handlerTY.WebhookData{}
		err := utils.MapToStruct(utils.TagNameNone, variable, &webhookCfg)
		if err != nil {
			v.logger.Error("error on converting into webhook data", zap.Error(err), zap.String("name", name), zap.Any("input", variable))
			return err.Error()
		}
		return GetWebhookData(v.logger, name, &webhookCfg)

	case types.VariableTypeResourceByQuickID, types.VariableTypeResourceByLabels:
		rsData := handlerTY.ResourceData{}
		err := utils.MapToStruct(utils.TagNameNone, variable, &rsData)
		if err != nil {
			v.logger.Error("error on converting into resource data", zap.Error(err), zap.String("name", name), zap.Any("input", variable))
			return err.Error()
		}
		if rsData.QuickID != "" {
			return v.getByQuickID(name, &rsData)
		} else if rsData.ResourceType != "" && len(rsData.Labels) > 0 {
			return v.getByLabels(name, &rsData)
		}

	default:
		return variable
	}

	return fmt.Sprintf("variable type not supported:%s, name:%s", variable.GetString(types.KeyType), name)
}

func (v *VariableSpec) getByLabels(name string, rsData *handlerTY.ResourceData) interface{} {
	apiImpl, found := v.genericApiMap[rsData.ResourceType]
	if !found {
		v.logger.Error("api not available to get a resource", zap.Any("data", rsData))
		return fmt.Sprintf("error on getting resourceType:%s", rsData.ResourceType)
	}

	filters := make([]storageTY.Filter, 0)
	for key, value := range rsData.Labels {
		filter := storageTY.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: storageTY.OperatorEqual, Value: value}
		filters = append(filters, filter)
	}
	pagination := &storageTY.Pagination{Limit: 1} // limit to one element
	result, err := apiImpl.List(filters, pagination)
	if err != nil {
		v.logger.Warn("error on getting label based entity", zap.Error(err), zap.String("name", name), zap.Any("rsData", rsData))
		return err.Error()
	}
	if result.Count == 0 {
		v.logger.Info("no records found", zap.String("variableName", name), zap.Any("data", rsData))
		return nil
	}

	// get the first element
	if reflect.TypeOf(result.Data).Elem().Kind() == reflect.Slice {
		s := reflect.ValueOf(result.Data).Elem()
		if s.Len() == 0 {
			return nil
		}
		// get the entity as interface
		entity := s.Index(0).Interface()

		if rsData.KeyPath == "." || rsData.KeyPath == "" { // return the object as is
			return entity
		} else {
			_, value, err := helper.GetValueByKeyPath(entity, rsData.KeyPath)
			if err != nil {
				v.logger.Warn("error on getting data from a given keyPath", zap.Error(err), zap.String("keyPath", rsData.KeyPath), zap.Any("entity", entity))
				return nil
			}
			return value
		}
	}

	v.logger.Warn("record type not as expected", zap.String("variableName", name), zap.Any("config", rsData), zap.String("resultType", fmt.Sprintf("%T", result.Data)))
	return nil
}

func (v *VariableSpec) getByQuickID(name string, rsData *handlerTY.ResourceData) interface{} {
	quickID := fmt.Sprintf("%s:%s", rsData.ResourceType, rsData.QuickID)
	resourceType, keys, err := quickIdUtils.EntityKeyValueMap(quickID)
	if err != nil {
		v.logger.Warn("failed to parse variable", zap.Any("name", name), zap.Any("data", rsData), zap.Error(err))
		return nil
	}
	var entity interface{}

	switch resourceType {
	case quickIdUtils.QuickIdGateway:
		item, err := v.api.Gateway().GetByID(keys[types.KeyGatewayID])
		entity = item
		if err != nil {
			v.logger.Warn("gateway not available", zap.Any("keys", keys))
			return nil
		}

	case quickIdUtils.QuickIdNode:
		item, err := v.api.Node().GetByGatewayAndNodeID(keys[types.KeyGatewayID], keys[types.KeyNodeID])
		entity = item
		if err != nil {
			v.logger.Warn("node not available", zap.Any("keys", keys))
			return nil
		}

	case quickIdUtils.QuickIdSource:
		item, err := v.api.Source().GetByIDs(keys[types.KeyGatewayID], keys[types.KeyNodeID], keys[types.KeySourceID])
		entity = item
		if err != nil {
			v.logger.Warn("source not available", zap.Any("keys", keys))
			return nil
		}

	case quickIdUtils.QuickIdField:
		item, err := v.api.Field().GetByIDs(keys[types.KeyGatewayID], keys[types.KeyNodeID], keys[types.KeySourceID], keys[types.KeyFieldID])
		entity = item
		if err != nil {
			v.logger.Warn("field not available", zap.Any("keys", keys), zap.Error(err))
			return nil
		}

	case quickIdUtils.QuickIdTask:
		item, err := v.api.Task().GetByID(keys[types.KeyID])
		entity = item
		if err != nil {
			v.logger.Warn("task not available", zap.Any("keys", keys))
			return nil
		}

	case quickIdUtils.QuickIdSchedule:
		item, err := v.api.Schedule().GetByID(keys[types.KeyID])
		entity = item
		if err != nil {
			v.logger.Warn("schedule not available", zap.Any("keys", keys))
			return nil
		}

	case quickIdUtils.QuickIdHandler:
		item, err := v.api.Handler().GetByID(keys[types.KeyID])
		entity = item
		if err != nil {
			v.logger.Warn("handler not available", zap.Any("keys", keys))
			return nil
		}

	case quickIdUtils.QuickIdDataRepository:
		item, err := v.api.DataRepository().GetByID(keys[types.KeyID])
		entity = item
		if err != nil {
			v.logger.Warn("data not available in data repository", zap.Any("keys", keys))
			return nil
		}

	default:
		data, err := v.templateEngine.Execute(keys[types.KeyTemplate], nil)
		if err != nil {
			v.logger.Warn("failed to parse template", zap.Any("keys", keys), zap.Error(err))
			return nil
		}
		entity = data
	}

	if entity == nil {
		return nil
	}

	if rsData.KeyPath == "." || rsData.KeyPath == "" { // return the object as is
		return entity
	} else {
		_, value, err := helper.GetValueByKeyPath(entity, rsData.KeyPath)
		if err != nil {
			v.logger.Error("error on getting data from a given keyPath", zap.Error(err), zap.String("keyPath", rsData.KeyPath), zap.Any("entity", entity))
			return nil
		}
		return value
	}
}
