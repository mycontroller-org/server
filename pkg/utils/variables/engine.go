package variables

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
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
		quickIdUtils.QuickIdField:          api.DataRepository(),
		quickIdUtils.QuickIdFirmware:       api.DataRepository(),
		quickIdUtils.QuickIdForwardPayload: api.DataRepository(),
		quickIdUtils.QuickIdGateway:        api.DataRepository(),
		quickIdUtils.QuickIdHandler:        api.DataRepository(),
		quickIdUtils.QuickIdNode:           api.DataRepository(),
		quickIdUtils.QuickIdSchedule:       api.DataRepository(),
		quickIdUtils.QuickIdSource:         api.DataRepository(),
		quickIdUtils.QuickIdTask:           api.DataRepository(),
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

// LoadVariables loads all the defined variables
func (v *VariableSpec) Load(variablesPreMap map[string]string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})
	for name, stringValue := range variablesPreMap {
		value := v.getEntity(name, stringValue)
		if value == nil {
			return nil, fmt.Errorf("failed to load a variable. name: %s, keyPath:%s", name, stringValue)
		}
		variables[name] = value
	}

	// clone variables
	clonedVariables := cloneUtil.Clone(variables)

	backToVariables, ok := clonedVariables.(map[string]interface{})
	if ok {
		// descrypt the secrets, tokens
		err := v.enc.DecryptSecrets(backToVariables)
		if err != nil {
			return nil, err
		}

		return backToVariables, nil
	}

	v.logger.Error("error on clone, returning variables as is")
	return variables, nil
}

func (v *VariableSpec) getEntity(name, stringValue string) interface{} {
	genericData := handlerTY.GenericData{}
	err := json.Unmarshal([]byte(stringValue), &genericData)
	if err != nil {
		// if error happens, this could be a normal string or templated string
		// try it as template
		formattedString, err := v.templateEngine.Execute(stringValue, nil)
		if err != nil {
			return fmt.Sprintf("error on executing template. name:%s, template:%s, error:%s", name, stringValue, err.Error())
		}
		return formattedString
	}

	// process only for resource type data
	if !strings.HasPrefix(genericData.Type, handlerTY.DataTypeResource) &&
		genericData.Type != handlerTY.DataTypeWebhook {
		return stringValue
	}

	// calls webhook and loads the response as is
	if genericData.Type == handlerTY.DataTypeWebhook {
		webhookCfg := handlerTY.WebhookData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &webhookCfg)
		if err != nil {
			v.logger.Error("error on loading webhook data", zap.Error(err), zap.String("name", name), zap.String("input", stringValue))
			return err.Error()
		}
		return GetWebhookData(v.logger, name, &webhookCfg)

	} else {
		rsData := handlerTY.ResourceData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &rsData)
		if err != nil {
			v.logger.Error("error on loading resource data", zap.Error(err), zap.String("name", name), zap.String("input", stringValue))
			return err.Error()
		}

		if rsData.QuickID != "" {
			return v.getByQuickID(name, &rsData)
		} else if rsData.ResourceType != "" && len(rsData.Labels) > 0 {
			return v.getByLabels(name, &rsData)
		}
	}

	return stringValue
}

func (v *VariableSpec) getByLabels(name string, rsData *handlerTY.ResourceData) interface{} {
	apiImpl, found := v.genericApiMap[rsData.ResourceType]
	if !found {
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
		return nil
	}

	//get the first element
	if reflect.TypeOf(result.Data).Kind() == reflect.Slice {
		s := reflect.ValueOf(result.Data)
		if s.Len() == 0 {
			return nil
		}
		entity := s.Index(0)
		if rsData.KeyPath != "" {
			_, value, err := helper.GetValueByKeyPath(entity, rsData.KeyPath)
			if err != nil {
				v.logger.Warn("error on getting data from a given keyPath", zap.Error(err), zap.String("keyPath", rsData.KeyPath), zap.Any("entity", entity))
				return nil
			}
			return value
		}
		return entity
	}

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

	if rsData.KeyPath != "" {
		_, value, err := helper.GetValueByKeyPath(entity, rsData.KeyPath)
		if err != nil {
			v.logger.Error("error on getting data from a given keyPath", zap.Error(err), zap.String("keyPath", rsData.KeyPath), zap.Any("entity", entity))
			return nil
		}
		return value
	}
	return entity
}
