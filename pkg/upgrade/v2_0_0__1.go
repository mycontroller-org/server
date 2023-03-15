package upgrade

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// these upgrades used to get from pre-release of v2.0.0
// converts base64 encoded variable and parameter values into normal fields
// updates backup handler config changes
func upgrade_2_0_0__1(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, api *entitiesAPI.API) error {
	filters := []storageTY.Filter{}
	pagination := &storageTY.Pagination{}

	// update tasks
	recordLimit := int64(20)
	offset := int64(0)
	for {
		pagination.Offset = offset
		pagination.Limit = recordLimit
		result, err := api.Task().List(filters, pagination)
		if err != nil {
			logger.Error("error on getting task", zap.Error(err))
			return err
		}
		data, ok := result.Data.(*[]taskTY.Config)
		if !ok {
			logger.Error("received invalid type", zap.String("actualType", fmt.Sprintf("%T", result.Data)))
			return errors.New("received invalid type")
		}

		for _, task := range *data {
			loggerWithEntity := logger.With(zap.String("entityName", fmt.Sprintf("task:%s", task.ID)))

			// update variables
			updatedVariables, err := upgrade_2_0_0__1_updateVariables(loggerWithEntity, task.Variables)
			if err != nil {
				return err
			}
			task.Variables = updatedVariables

			// update handler parameters
			updatedParameters, err := upgrade_2_0_0__1_updateParameters(loggerWithEntity, task.HandlerParameters)
			if err != nil {
				return err
			}
			task.HandlerParameters = updatedParameters

			// save the task
			// do not use save method, it needs bus
			err = api.Task().Import(task)
			if err != nil {
				logger.Error("task import error", zap.Error(err))
				return err
			}
		}

		offset += recordLimit
		if result.Count < offset {
			break
		}
	}

	// update schedules
	offset = int64(0)
	for {
		pagination.Offset = offset
		pagination.Limit = recordLimit
		result, err := api.Schedule().List(filters, pagination)
		if err != nil {
			logger.Error("error on getting schedule", zap.Error(err))
			return err
		}
		data, ok := result.Data.(*[]scheduleTY.Config)
		if !ok {
			logger.Error("received invalid type", zap.String("actualType", fmt.Sprintf("%T", result.Data)))
			return errors.New("received invalid type")
		}

		for _, schedule := range *data {
			loggerWithEntity := logger.With(zap.String("entityName", fmt.Sprintf("schedule:%s", schedule.ID)))

			// update variables
			updatedVariables, err := upgrade_2_0_0__1_updateVariables(loggerWithEntity, schedule.Variables)
			if err != nil {
				return err
			}
			schedule.Variables = updatedVariables

			// update handler parameters
			updatedParameters, err := upgrade_2_0_0__1_updateParameters(loggerWithEntity, schedule.HandlerParameters)
			if err != nil {
				return err
			}
			schedule.HandlerParameters = updatedParameters

			// save the schedule
			// do not use save method, it needs bus
			err = api.Schedule().Import(schedule)
			if err != nil {
				logger.Error("schedule import error", zap.Error(err))
				return err
			}
		}

		offset += recordLimit
		if result.Count < offset {
			break
		}
	}

	// update handler
	err := upgrade_2_0_0__1_backup_handler_config(logger, api)
	if err != nil {
		return err
	}

	return nil
}

// upgrade variables
func upgrade_2_0_0__1_updateVariables(logger *zap.Logger, variables map[string]interface{}) (map[string]interface{}, error) {
	updatedVariables := map[string]interface{}{}
	for name := range variables {
		variableString := variables[name]
		if _, ok := variableString.(string); !ok {
			logger.Warn("supplied variable is not a string type, already migrated?", zap.String("name", name))
			updatedVariables[name] = variableString
			continue
		}

		logger.Debug("updating variable", zap.String("name", name), zap.String("variable", variableString.(string)))

		// convert the json string into instance
		variable := map[string]interface{}{}
		err := json.Unmarshal([]byte(variableString.(string)), &variable)
		if err != nil {
			// it is string variable
			updatedVariables[name] = map[string]interface{}{
				"type":  types.VariableTypeString,
				"value": variableString.(string),
			}
			continue
		}

		logger.Debug("converted variable", zap.Any("variable", variable))

		varStruct := struct {
			Var map[string]interface{}
		}{
			Var: variable,
		}
		_, varType, err := helper.GetValueByKeyPath(varStruct, "var.type")
		if err != nil {
			logger.Error("error on getting type", zap.String("variableName", name), zap.Error(err))
			return nil, err
		}

		_, varDataString, err := helper.GetValueByKeyPath(varStruct, "var.data")
		if err != nil {
			logger.Error("error on getting data string", zap.String("variableName", name), zap.Error(err))
			return nil, err
		}

		varData := map[string]interface{}{}
		err = yamlUtils.UnmarshalBase64Yaml(varDataString.(string), &varData)
		if err != nil {
			logger.Error("error on converting to struct", zap.String("variableName", name), zap.Error(err))
			return nil, err
		}

		// decode the data
		switch varType.(string) {
		case types.VariableTypeResourceByQuickID, types.VariableTypeResourceByLabels, types.VariableTypeWebhook:
			newStruct := map[string]interface{}{
				"type": varType.(string),
			}
			for key, value := range varData {
				newStruct[key] = value
			}
			updatedVariables[name] = newStruct

		default:
			logger.Warn("unknown variable type received, skipped", zap.String("name", name), zap.String("type", varType.(string)))

		}
	}
	logger.Debug("updated variables", zap.Any("new", updatedVariables))
	return updatedVariables, nil
}

// upgrade handler parameters
func upgrade_2_0_0__1_updateParameters(logger *zap.Logger, parameters map[string]interface{}) (map[string]interface{}, error) {
	updatedParameters := map[string]interface{}{}
	for name := range parameters {
		parameterString := parameters[name]
		if _, ok := parameterString.(string); !ok {
			logger.Warn("supplied parameter is not a string type, already migrated?", zap.String("name", name))
			updatedParameters[name] = parameterString
			continue
		}
		logger.Debug("updating parameter", zap.String("name", name), zap.String("parameter", parameterString.(string)))

		// convert the json string into instance
		parameter := map[string]interface{}{}
		err := json.Unmarshal([]byte(parameterString.(string)), &parameter)
		if err != nil {
			// it is string parameter, not usable
			logger.Warn("seems string type set in parameters, which is not usable. removing it", zap.String("name", name))
			continue
		}

		logger.Debug("converted parameter", zap.Any("parameter", parameter))

		varStruct := struct {
			Para map[string]interface{}
		}{
			Para: parameter,
		}
		_, parameterType, err := helper.GetValueByKeyPath(varStruct, "para.type")
		if err != nil {
			logger.Error("error on getting type", zap.String("parameterName", name), zap.Error(err))
			return nil, err
		}

		_, disabledValue, err := helper.GetValueByKeyPath(varStruct, "para.disabled")
		if err != nil {
			logger.Warn("error on getting type, setting as empty value", zap.String("parameterName", name), zap.Error(err))
			disabledValue = ""
		}

		_, parameterDataString, err := helper.GetValueByKeyPath(varStruct, "para.data")
		if err != nil {
			logger.Error("error on getting data string", zap.String("ParameterName", name), zap.Error(err))
			return nil, err
		}

		parameterData := map[string]interface{}{}
		err = yamlUtils.UnmarshalBase64Yaml(parameterDataString.(string), &parameterData)
		if err != nil {
			logger.Error("error on converting to struct", zap.String("parameterName", name), zap.Error(err))
			return nil, err
		}

		logger.Debug("actual parameter data", zap.String("name", name), zap.Any("data", parameterData))

		// decode the data
		switch parameterType.(string) {
		case types.VariableTypeResourceByLabels, types.VariableTypeResourceByQuickID,
			handlerTY.DataTypeEmail, handlerTY.DataTypeTelegram, handlerTY.DataTypeWebhook:
			newStruct := map[string]interface{}{
				"type":     parameterType.(string),
				"disabled": disabledValue.(string),
			}
			for key, value := range parameterData {
				newStruct[key] = value
			}
			updatedParameters[name] = newStruct

		case handlerTY.DataTypeBackup:
			newStruct := map[string]interface{}{
				"type":         parameterType.(string),
				"disabled":     disabledValue.(string),
				"providerType": parameterData["providerType"],
			}
			// get spec
			spec, ok := parameterData["spec"].(map[string]interface{})
			if !ok {
				logger.Warn("unable to get spec detail, skipped", zap.String("name", name), zap.String("type", parameterType.(string)))
				continue
			}
			for key, value := range spec {
				newStruct[key] = value
			}
			updatedParameters[name] = newStruct

		default:
			logger.Warn("unknown parameter type received, skipped", zap.String("name", name), zap.String("type", parameterType.(string)))

		}
	}
	logger.Debug("updated parameters", zap.Any("new", updatedParameters))
	return updatedParameters, nil
}

// upgrades backup handler config
// moves .spec.spec => .spec
// removes .spec.spec
func upgrade_2_0_0__1_backup_handler_config(logger *zap.Logger, api *entitiesAPI.API) error {
	filters := []storageTY.Filter{}
	pagination := &storageTY.Pagination{}

	// update tasks
	recordLimit := int64(20)
	offset := int64(0)
	for {
		pagination.Offset = offset
		pagination.Limit = recordLimit
		result, err := api.Handler().List(filters, pagination)
		if err != nil {
			logger.Error("error on getting handlers", zap.Error(err))
			return err
		}
		data, ok := result.Data.(*[]handlerTY.Config)
		if !ok {
			logger.Error("received invalid type", zap.String("actualType", fmt.Sprintf("%T", result.Data)))
			return errors.New("received invalid type")
		}

		for _, handler := range *data {
			if handler.Type != handlerTY.DataTypeBackup {
				continue
			}
			logger.Debug("updating handler", zap.String("name", handler.ID), zap.Any("handler", handler))

			// get .spec.spec
			rawSpec := handler.Spec["spec"]
			if rawSpec == nil {
				logger.Warn("seems upgrade completed? 'spec.spec' not found", zap.Any("data", handler))
				continue
			}
			spec, ok := rawSpec.(cmap.CustomMap)
			if !ok {
				logger.Error("error on converting .spec.spec into map[string]interface{}", zap.String("name", handler.ID), zap.Any("data", handler), zap.String("actualType", fmt.Sprintf("%T", rawSpec)))
				return errors.New("error on converting .spec.spec into map[string]interface{}")
			}
			// move .spec.spec => .spec
			for key, value := range spec {
				handler.Spec[key] = value
			}
			// delete .spec.spec
			delete(handler.Spec, "spec")

			logger.Debug("updated handler", zap.Any("data", handler))
			// update to database
			err = api.Handler().Import(handler)
			if err != nil {
				logger.Error("error on saving a handler", zap.String("name", handler.ID), zap.Any("data", handler))
			}
		}

		offset += recordLimit
		if result.Count < offset {
			break
		}
	}
	return nil
}
